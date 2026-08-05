// Package web embeds the built frontend (web/frontend, via Vite) into the
// Go binary. `make build-frontend` regenerates web/dist; the checked-in
// copy is a placeholder so `go build` works before the frontend is built.
package web

import "embed"

//go:embed all:dist
var DistFS embed.FS
