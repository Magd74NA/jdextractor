package jdextract

import (
	"embed"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

//go:embed web/index.html
var webFiles embed.FS

// registerRoutes attaches all HTTP handlers to mux.
func (a *App) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("GET /api/config", a.handleGetConfig)
	mux.HandleFunc("PATCH /api/config", a.handleUpdateConfig)
	mux.HandleFunc("GET /api/templates", a.handleGetTemplates)
	mux.HandleFunc("PATCH /api/templates", a.handleSaveTemplates)
	mux.HandleFunc("GET /api/jobs/{id}/files", a.handleGetJobFiles)
	mux.HandleFunc("PATCH /api/jobs/{id}/files", a.handleSaveJobFiles)
	mux.HandleFunc("GET /api/jobs", a.handleListJobs)
	mux.HandleFunc("PATCH /api/jobs/{id}", a.handleUpdateJobStatus)
	mux.HandleFunc("DELETE /api/jobs/{id}", a.handleDeleteJob)
	mux.HandleFunc("POST /api/process", a.handleProcess)
	mux.HandleFunc("POST /api/process/batch", a.handleProcessBatch)
	mux.HandleFunc("POST /api/process/local", a.handleProcessLocal)
}

// handleGetTemplates returns the current resume and cover letter templates.
func (a *App) handleGetTemplates(w http.ResponseWriter, r *http.Request) {
	out := struct {
		Resume string `json:"resume"`
		Cover  string `json:"cover,omitempty"`
	}{}
	if b, err := os.ReadFile(filepath.Join(a.Paths.Templates, "resume.txt")); err == nil {
		out.Resume = string(b)
	}
	if b, err := os.ReadFile(filepath.Join(a.Paths.Templates, "cover.txt")); err == nil {
		out.Cover = string(b)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// handleSaveTemplates writes resume.txt and/or cover.txt to the templates directory.
// Only non-nil fields in the body are written.
func (a *App) handleSaveTemplates(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Resume *string `json:"resume"`
		Cover  *string `json:"cover"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.Resume != nil {
		if err := os.WriteFile(filepath.Join(a.Paths.Templates, "resume.txt"), []byte(*body.Resume), 0644); err != nil {
			http.Error(w, "write resume template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if body.Cover != nil {
		if err := os.WriteFile(filepath.Join(a.Paths.Templates, "cover.txt"), []byte(*body.Cover), 0644); err != nil {
			http.Error(w, "write cover template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleGetJobFiles returns the resume and (optionally) cover letter content
// for a job identified by its exact directory name.
func (a *App) handleGetJobFiles(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" || id == "." || strings.Contains(id, "/") || strings.Contains(id, "\\") {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}
	dir := filepath.Join(a.Paths.Jobs, id)

	resumeBytes, err := os.ReadFile(filepath.Join(dir, "resume.txt"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "job not found", http.StatusNotFound)
		} else {
			http.Error(w, "read error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	out := struct {
		Resume string `json:"resume"`
		Cover  string `json:"cover,omitempty"`
	}{Resume: string(resumeBytes)}

	if coverBytes, err := os.ReadFile(filepath.Join(dir, "cover.txt")); err == nil {
		out.Cover = string(coverBytes)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// handleSaveJobFiles writes resume.txt and/or cover.txt for a job.
// Only non-empty fields in the body are written; omitted fields are left untouched.
func (a *App) handleSaveJobFiles(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" || id == "." || strings.Contains(id, "/") || strings.Contains(id, "\\") {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}
	var body struct {
		Resume *string `json:"resume"`
		Cover  *string `json:"cover"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	dir := filepath.Join(a.Paths.Jobs, id)
	if body.Resume != nil {
		if err := os.WriteFile(filepath.Join(dir, "resume.txt"), []byte(*body.Resume), 0644); err != nil {
			http.Error(w, "write resume: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if body.Cover != nil {
		if err := os.WriteFile(filepath.Join(dir, "cover.txt"), []byte(*body.Cover), 0644); err != nil {
			http.Error(w, "write cover: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

var (
	validDeepSeekModels = []string{"deepseek-chat", "deepseek-reasoner"}
	validBackends       = []string{"deepseek", "kimi"}
)

// handleGetConfig returns the current in-memory Config as JSON.
func (a *App) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a.Config)
}

// handleUpdateConfig applies a partial config update, persists it to disk, and
// updates the in-memory Config. Only non-nil fields in the body are applied.
func (a *App) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DeepSeekApiKey *string `json:"deepseek_api_key"`
		DeepSeekModel  *string `json:"deepseek_model"`
		KimiApiKey     *string `json:"kimi_api_key"`
		KimiModel      *string `json:"kimi_model"`
		Backend        *string `json:"backend"`
		Port           *int    `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.DeepSeekModel != nil && !slices.Contains(validDeepSeekModels, *body.DeepSeekModel) {
		http.Error(w, "invalid model: must be deepseek-chat or deepseek-reasoner", http.StatusBadRequest)
		return
	}
	if body.Backend != nil && !slices.Contains(validBackends, *body.Backend) {
		http.Error(w, "invalid backend: must be deepseek or kimi", http.StatusBadRequest)
		return
	}
	if body.DeepSeekApiKey != nil {
		a.Config.DeepSeekApiKey = *body.DeepSeekApiKey
	}
	if body.DeepSeekModel != nil {
		a.Config.DeepSeekModel = *body.DeepSeekModel
	}
	if body.KimiApiKey != nil {
		a.Config.KimiApiKey = *body.KimiApiKey
	}
	if body.KimiModel != nil {
		a.Config.KimiModel = *body.KimiModel
	}
	if body.Backend != nil {
		a.Config.Backend = *body.Backend
	}
	if body.Port != nil {
		a.Config.Port = *body.Port
	}
	path := filepath.Join(a.Paths.Config, "config.json")
	if err := SaveConfig(path, a.Config); err != nil {
		http.Error(w, "failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
