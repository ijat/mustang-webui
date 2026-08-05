.PHONY: dev-sidecar dev-frontend build-sidecar build-frontend build test check clean

# Manual three-terminal dev workflow (see CLAUDE.md for the full loop):
#   make build-sidecar && make dev-sidecar   # terminal 1
#   make dev-frontend                        # terminal 2
#   go run ./cmd/mustang-webui --dev         # terminal 3

build-sidecar:
	cd sidecar && mvn -q -DskipTests package

dev-sidecar:
	java -jar sidecar/target/sidecar-*.jar --port 8765 --token dev

dev-frontend:
	npm --prefix web/frontend run dev

build-frontend:
	npm --prefix web/frontend run build

build: build-sidecar build-frontend
	go build -o bin/mustang-webui ./cmd/mustang-webui

test:
	go test ./...
	cd sidecar && mvn -q -DskipTests test
	npm --prefix web/frontend run check

clean:
	rm -rf bin sidecar/target web/frontend/node_modules
