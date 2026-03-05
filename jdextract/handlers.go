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
	mux.HandleFunc("POST /api/process", a.handleProcess)
	mux.HandleFunc("POST /api/process/batch", a.handleProcessBatch)
	mux.HandleFunc("POST /api/process/local", a.handleProcessLocal)
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

// handleProcess fetches a job description from a URL via jina.ai and runs the
// full generation pipeline, returning {"dir":"..."} on success.
func (a *App) handleProcess(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.URL == "" {
		http.Error(w, "invalid JSON body: url required", http.StatusBadRequest)
		return
	}
	raw, err := FetchJobDescription(r.Context(), body.URL, &a.Client, 0)
	if err != nil {
		http.Error(w, "fetch error: "+err.Error(), http.StatusBadGateway)
		return
	}
	dir, err := a.Process(r.Context(), raw)
	if err != nil {
		http.Error(w, "process error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Dir string `json:"dir"`
	}{Dir: dir})
}

// batchItemResult is the per-URL outcome returned by handleProcessBatch.
type batchItemResult struct {
	URL   string `json:"url"`
	Dir   string `json:"dir,omitempty"`
	Error string `json:"error,omitempty"`
}

// handleProcessBatch accepts {"urls":[...]} and processes all URLs concurrently,
// returning an array of per-URL outcomes. Individual failures do not abort others.
func (a *App) handleProcessBatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URLs []string `json:"urls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.URLs) == 0 {
		http.Error(w, "invalid JSON body: urls required", http.StatusBadRequest)
		return
	}
	var results []batchItemResult
	for br := range a.ProcessBatch(r.Context(), body.URLs) {
		res := batchItemResult{URL: br.URL, Dir: br.Dir}
		if br.Err != nil {
			res.Error = br.Err.Error()
		}
		results = append(results, res)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// handleProcessLocal accepts {"content":"..."} (raw job description text) and
// runs the generation pipeline directly, returning {"dir":"..."} on success.
func (a *App) handleProcessLocal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Content == "" {
		http.Error(w, "invalid JSON body: content required", http.StatusBadRequest)
		return
	}
	dir, err := a.Process(r.Context(), body.Content)
	if err != nil {
		http.Error(w, "process error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Dir string `json:"dir"`
	}{Dir: dir})
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
