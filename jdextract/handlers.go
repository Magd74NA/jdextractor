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
	mux.HandleFunc("PATCH /api/jobs/{id}", a.handleUpdateJobStatus)
	mux.HandleFunc("DELETE /api/jobs/{id}", a.handleDeleteJob)
}

// jobResponse is the wire type for GET /api/jobs. It embeds ApplicationMeta
// and adds Dir as a JSON field — Dir is excluded from the stored meta.json
// (json:"-") so it must be lifted here for the frontend to use as an ID.
type jobResponse struct {
	ApplicationMeta
	Dir string `json:"dir"`
}

// handleListJobs returns all processed job applications as a JSON array.
func (a *App) handleListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := ListJobs(a)
	if err != nil {
		http.Error(w, "failed to list jobs", http.StatusInternalServerError)
		return
	}
	out := make([]jobResponse, len(jobs))
	for i, j := range jobs {
		out[i] = jobResponse{ApplicationMeta: j, Dir: j.Dir}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// handleUpdateJobStatus decodes {"status":"..."} and updates the job's meta.json.
func (a *App) handleUpdateJobStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := UpdateJobStatus(a, id, body.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteJob removes a job directory by its exact directory name.
func (a *App) handleDeleteJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := DeleteJob(a, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
