package dev.mustangwebui.sidecar;

import com.sun.net.httpserver.Filter;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.List;

/**
 * Entry point for the sidecar process spawned by the Go orchestrator.
 * Binds to 127.0.0.1 only and requires a bearer token (passed via
 * --token, generated fresh per launch) on every request except /healthz.
 */
public final class Main {

    public static void main(String[] args) throws IOException {
        int port = -1;
        String token = null;

        for (int i = 0; i < args.length - 1; i++) {
            switch (args[i]) {
                case "--port" -> port = Integer.parseInt(args[i + 1]);
                case "--token" -> token = args[i + 1];
                default -> { }
            }
        }

        if (port < 0 || token == null) {
            System.err.println("usage: sidecar --port <port> --token <token>");
            System.exit(2);
        }

        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", port), 0);

        server.createContext("/healthz", Main::handleHealthz);

        String expectedAuth = "Bearer " + token;
        server.createContext("/api", Main::routeApi).getFilters().add(bearerAuthFilter(expectedAuth));

        server.setExecutor(null);
        server.start();

        System.out.println("sidecar listening on 127.0.0.1:" + port);

        Runtime.getRuntime().addShutdownHook(new Thread(() -> server.stop(1)));
    }

    private static void handleHealthz(HttpExchange exchange) throws IOException {
        sendJson(exchange, 200, "{\"status\":\"ok\"}");
    }

    private static void routeApi(HttpExchange exchange) throws IOException {
        String path = exchange.getRequestURI().getPath();
        String method = exchange.getRequestMethod();

        if ("/api/inspect".equals(path)) {
            if (!"POST".equalsIgnoreCase(method)) {
                sendJson(exchange, 405, "{\"error\":\"method not allowed\"}");
                return;
            }
            InspectHandler.handle(exchange);
            return;
        }

        sendJson(exchange, 501, "{\"error\":\"not implemented yet\"}");
    }

    private static Filter bearerAuthFilter(String expectedAuth) {
        return new Filter() {
            @Override
            public String description() {
                return "requires a matching Authorization: Bearer <token> header";
            }

            @Override
            public void doFilter(HttpExchange exchange, Chain chain) throws IOException {
                List<String> headers = exchange.getRequestHeaders().get("Authorization");
                String actual = headers == null || headers.isEmpty() ? null : headers.get(0);
                if (!expectedAuth.equals(actual)) {
                    sendJson(exchange, 401, "{\"error\":\"unauthorized\"}");
                    return;
                }
                chain.doFilter(exchange);
            }
        };
    }

    static void sendJson(HttpExchange exchange, int status, String body) throws IOException {
        byte[] bytes = body.getBytes(StandardCharsets.UTF_8);
        exchange.getResponseHeaders().set("Content-Type", "application/json");
        exchange.sendResponseHeaders(status, bytes.length);
        try (OutputStream os = exchange.getResponseBody()) {
            os.write(bytes);
        }
    }

    private Main() { }
}
