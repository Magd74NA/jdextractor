package jdextract

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

//go:embed web/dist
var webFiles embed.FS

// registerRoutes attaches all HTTP handlers to mux.
func (a *App) registerRoutes(mux *http.ServeMux) {
	mux.Handle("/", spaHandler())
	mux.HandleFunc("GET /api/config", a.handleGetConfig)
	mux.HandleFunc("PATCH /api/config", a.handleUpdateConfig)
	mux.HandleFunc("GET /api/config/prompt", a.handleGetPromptConfig)
	mux.HandleFunc("PATCH /api/config/prompt", a.handleUpdatePromptConfig)
	mux.HandleFunc("GET /api/templates", a.handleGetTemplates)
	mux.HandleFunc("PATCH /api/templates", a.handleSaveTemplates)
	mux.HandleFunc("GET /api/jobs/{id}/files", a.handleGetJobFiles)
	mux.HandleFunc("PATCH /api/jobs/{id}/files", a.handleSaveJobFiles)
	mux.HandleFunc("GET /api/jobs", a.handleListJobs)
	mux.HandleFunc("PATCH /api/jobs/{id}", a.handleUpdateJobStatus)
	mux.HandleFunc("DELETE /api/jobs/{id}", a.handleDeleteJob)
	mux.HandleFunc("POST /api/process", a.handleProcess)
	mux.HandleFunc("POST /api/process/stream", a.handleProcessStream)
	mux.HandleFunc("POST /api/process/batch", a.handleProcessBatch)
	mux.HandleFunc("POST /api/process/local", a.handleProcessLocal)
	mux.HandleFunc("POST /api/process/local/stream", a.handleProcessLocalStream)

	// Contacts — specific paths before wildcard /{id}
	mux.HandleFunc("GET /api/contacts/overdue", a.handleOverdueFollowups)
	mux.HandleFunc("GET /api/contacts/upcoming", a.handleUpcomingFollowups)
	mux.HandleFunc("GET /api/contacts", a.handleListContacts)
	mux.HandleFunc("POST /api/contacts", a.handleCreateContact)
	mux.HandleFunc("GET /api/contacts/{id}", a.handleGetContact)
	mux.HandleFunc("PATCH /api/contacts/{id}", a.handleUpdateContact)
	mux.HandleFunc("DELETE /api/contacts/{id}", a.handleDeleteContact)
	mux.HandleFunc("POST /api/contacts/{id}/conversations", a.handleAddConversation)
	mux.HandleFunc("DELETE /api/contacts/{id}/conversations/{index}", a.handleDeleteConversation)
	mux.HandleFunc("PATCH /api/contacts/{id}/conversations/{index}", a.handleUpdateConversationSummary)
	mux.HandleFunc("POST /api/contacts/{id}/conversations/{index}/summarize", a.handleSummarizeConversation)
	mux.HandleFunc("POST /api/contacts/{id}/conversations/{index}/messages", a.handleAddMessage)
	mux.HandleFunc("DELETE /api/contacts/{id}/conversations/{index}/messages/{msgIndex}", a.handleDeleteMessage)
	mux.HandleFunc("POST /api/contacts/{id}/followup", a.handleGenerateFollowup)
	mux.HandleFunc("POST /api/contacts/{id}/followup/stream", a.handleGenerateFollowupStream)
	mux.HandleFunc("POST /api/contacts/{id}/followup/send", a.handleSendFollowup)

	// Networking prompt config
	mux.HandleFunc("GET /api/config/networking-prompt", a.handleGetNetworkingPromptConfig)
	mux.HandleFunc("PATCH /api/config/networking-prompt", a.handleUpdateNetworkingPromptConfig)
}

// writeJSON sets Content-Type and encodes v as JSON.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// decodeBody decodes JSON from r.Body into v. Returns false and writes a 400
// on failure so callers can early-return immediately.
func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return false
	}
	return true
}

// validID is an alias kept for local readability; delegates to ValidID in store.go.
func validID(id string) bool {
	return ValidID(id)
}

var (
	validDeepSeekModels = []string{"deepseek-chat", "deepseek-reasoner"}
	validKimiModels     = []string{"moonshotai/Kimi-K2.5"}
	validBackends       = []string{"deepseek", "kimi"}
)

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
	writeJSON(w, out)
}

// handleSaveTemplates writes resume.txt and/or cover.txt to the templates directory.
// Only non-nil fields in the body are written.
func (a *App) handleSaveTemplates(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Resume *string `json:"resume"`
		Cover  *string `json:"cover"`
	}
	if !decodeBody(w, r, &body) {
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
	if !validID(id) {
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
	writeJSON(w, out)
}

// handleSaveJobFiles writes resume.txt and/or cover.txt for a job.
// Only non-nil fields in the body are written; omitted fields are left untouched.
func (a *App) handleSaveJobFiles(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}
	var body struct {
		Resume *string `json:"resume"`
		Cover  *string `json:"cover"`
	}
	if !decodeBody(w, r, &body) {
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

// handleGetConfig returns the current in-memory Config as JSON.
func (a *App) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, a.Config)
}

// handleGetPromptConfig returns the current in-memory PromptConfig as JSON.
func (a *App) handleGetPromptConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, a.PromptConfig)
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
	if !decodeBody(w, r, &body) {
		return
	}
	if body.DeepSeekModel != nil && !slices.Contains(validDeepSeekModels, *body.DeepSeekModel) {
		http.Error(w, "invalid model: must be deepseek-chat or deepseek-reasoner", http.StatusBadRequest)
		return
	}
	if body.KimiModel != nil && !slices.Contains(validKimiModels, *body.KimiModel) {
		http.Error(w, "invalid model: must be moonshotai/Kimi-K2.5", http.StatusBadRequest)
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
	if err := SaveJSON(path, a.Config, 0600); err != nil {
		http.Error(w, "failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleUpdatePromptConfig applies a prompt config update, persists it to disk, and
// updates the in-memory PromptConfig.
func (a *App) handleUpdatePromptConfig(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TaskList     *string `json:"task_list"`
		SystemPrompt *string `json:"system_prompt"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	if body.TaskList != nil {
		a.PromptConfig.TaskList = *body.TaskList
	}
	if body.SystemPrompt != nil {
		a.PromptConfig.SystemPrompt = *body.SystemPrompt
	}
	path := filepath.Join(a.Paths.Config, "prompt.json")
	if err := SaveJSON(path, a.PromptConfig, 0600); err != nil {
		http.Error(w, "failed to save prompt config: "+err.Error(), http.StatusInternalServerError)
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
	writeJSON(w, out)
}

// handleUpdateJobStatus decodes a partial job update and applies it to meta.json.
// Accepts any combination of status, company, role, and date fields.
func (a *App) handleUpdateJobStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status  *string `json:"status"`
		Company *string `json:"company"`
		Role    *string `json:"role"`
		Date    *string `json:"date"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	if body.Status != nil {
		if err := UpdateJobStatus(a, id, *body.Status); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if body.Company != nil || body.Role != nil || body.Date != nil {
		if err := UpdateJobMeta(a, id, body.Company, body.Role, body.Date); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
	if !decodeBody(w, r, &body) || body.URL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
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
	writeJSON(w, struct {
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
	if !decodeBody(w, r, &body) || len(body.URLs) == 0 {
		http.Error(w, "urls required", http.StatusBadRequest)
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
	writeJSON(w, results)
}

// handleProcessLocal accepts {"content":"..."} (raw job description text) and
// runs the generation pipeline directly, returning {"dir":"..."} on success.
func (a *App) handleProcessLocal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	if !decodeBody(w, r, &body) || body.Content == "" {
		http.Error(w, "content required", http.StatusBadRequest)
		return
	}
	dir, err := a.Process(r.Context(), body.Content)
	if err != nil {
		http.Error(w, "process error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, struct {
		Dir string `json:"dir"`
	}{Dir: dir})
}

// writeSSE marshals a ProgressEvent as an SSE data line and flushes.
func writeSSE(w http.ResponseWriter, flusher http.Flusher, event ProgressEvent) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// initSSE sets SSE headers and returns the Flusher. Returns nil if not supported.
func initSSE(w http.ResponseWriter) http.Flusher {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return nil
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	return flusher
}

// handleProcessStream is the SSE variant of handleProcess. It streams progress
// events as the pipeline runs: fetching → parsing → generating → saving → complete.
func (a *App) handleProcessStream(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if !decodeBody(w, r, &body) || body.URL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}
	flusher := initSSE(w)
	if flusher == nil {
		return
	}

	writeSSE(w, flusher, ProgressEvent{Stage: StageFetching, Message: "Fetching job description\u2026"})
	raw, err := FetchJobDescription(r.Context(), body.URL, &a.Client, 0)
	if err != nil {
		writeSSE(w, flusher, ProgressEvent{Stage: StageError, Message: "fetch error: " + err.Error()})
		return
	}

	dir, err := a.ProcessWithProgress(r.Context(), raw, func(e ProgressEvent) {
		writeSSE(w, flusher, e)
	})
	if err != nil {
		writeSSE(w, flusher, ProgressEvent{Stage: StageError, Message: "process error: " + err.Error()})
		return
	}
	writeSSE(w, flusher, ProgressEvent{Stage: StageComplete, Dir: dir})
}

// handleProcessLocalStream is the SSE variant of handleProcessLocal.
// Streams progress events: parsing → generating → saving → complete.
func (a *App) handleProcessLocalStream(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	if !decodeBody(w, r, &body) || body.Content == "" {
		http.Error(w, "content required", http.StatusBadRequest)
		return
	}
	flusher := initSSE(w)
	if flusher == nil {
		return
	}

	dir, err := a.ProcessWithProgress(r.Context(), body.Content, func(e ProgressEvent) {
		writeSSE(w, flusher, e)
	})
	if err != nil {
		writeSSE(w, flusher, ProgressEvent{Stage: StageError, Message: "process error: " + err.Error()})
		return
	}
	writeSSE(w, flusher, ProgressEvent{Stage: StageComplete, Dir: dir})
}

// contentTypes maps file extensions to MIME types for gzip-compressed assets.
var contentTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".js":   "application/javascript",
	".css":  "text/css",
}

// spaHandler returns an http.Handler that serves gzip-compressed embedded
// SPA assets. All assets in dist are .gz files; the handler sets
// Content-Encoding and Content-Type, and streams the pre-compressed bytes.
// Unmatched paths fall back to index.html.gz for client-side routing.
func spaHandler() http.Handler {
	dist, _ := fs.Sub(webFiles, "web/dist")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "index.html"
		} else {
			path = strings.TrimPrefix(path, "/")
		}
		gzPath := path + ".gz"
		if _, err := fs.Stat(dist, gzPath); err != nil {
			// Fall back to index.html for SPA routing.
			gzPath = "index.html.gz"
			path = "index.html"
		}
		ext := filepath.Ext(path)
		if ct, ok := contentTypes[ext]; ok {
			w.Header().Set("Content-Type", ct)
		}
		w.Header().Set("Content-Encoding", "gzip")
		data, _ := fs.ReadFile(dist, gzPath)
		w.Write(data)
	})
}
