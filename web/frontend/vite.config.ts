import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [svelte(), tailwindcss()],
  build: {
    outDir: '../dist',
    emptyOutDir: true,
  },
  server: {
    // Lets `npm run dev` alone talk to a manually started sidecar, in
    // addition to the Go orchestrator's own proxying in --dev mode.
    proxy: {
      '/api': 'http://127.0.0.1:8765',
    },
  },
})
