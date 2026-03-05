package jdextract

import (
	"embed"
	"net/http"
)

//go:embed web/index.html
var webFiles embed.FS

// registerRoutes attaches all HTTP handlers to mux.
func (a *App) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", a.handleIndex)
}

// handleIndex serves the embedded web UI.
func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := webFiles.ReadFile("web/index.html")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}
