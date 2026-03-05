package jdextract

import (
	"embed"
	"encoding/json"
	"net/http"
)

//go:embed web/index.html
var webFiles embed.FS

// registerRoutes attaches all HTTP handlers to mux.
func (a *App) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("GET /api/jobs", a.handleListJobs)
}

// handleListJobs returns all processed job applications as a JSON array.
func (a *App) handleListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := ListJobs(a)
	if err != nil {
		http.Error(w, "failed to list jobs", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
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
