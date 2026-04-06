package jdextract

import (
	"net/http"
	"path/filepath"
	"strconv"
)

// handleListContacts returns all contacts as JSON.
func (a *App) handleListContacts(w http.ResponseWriter, r *http.Request) {
	contacts, err := ListContacts(a)
	if err != nil {
		http.Error(w, "list contacts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if contacts == nil {
		contacts = []ContactMeta{}
	}
	writeJSON(w, contacts)
}

// handleCreateContact creates a new contact from the JSON body.
func (a *App) handleCreateContact(w http.ResponseWriter, r *http.Request) {
	var body ContactMeta
	if !decodeBody(w, r, &body) {
		return
	}
	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	dir, err := CreateContact(a, body)
	if err != nil {
		http.Error(w, "create contact: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]string{"dir": dir})
}

// handleGetContact returns a single contact by its directory name.
func (a *App) handleGetContact(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	contact, err := GetContact(a, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, contact)
}

// handleUpdateContact applies partial updates to a contact.
func (a *App) handleUpdateContact(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	var updates ContactUpdate
	if !decodeBody(w, r, &updates) {
		return
	}
	if err := UpdateContact(a, id, updates); err != nil {
		http.Error(w, "update contact: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteContact removes a contact by its directory name.
func (a *App) handleDeleteContact(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	if err := DeleteContact(a, id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleAddConversation appends a conversation entry to a contact.
func (a *App) handleAddConversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	var entry ConversationEntry
	if !decodeBody(w, r, &entry) {
		return
	}
	if entry.Summary == "" {
		http.Error(w, "summary is required", http.StatusBadRequest)
		return
	}
	if err := AddConversation(a, id, entry); err != nil {
		http.Error(w, "add conversation: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteConversation removes a conversation by index.
func (a *App) handleDeleteConversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	indexStr := r.PathValue("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		http.Error(w, "invalid conversation index", http.StatusBadRequest)
		return
	}
	if err := DeleteConversation(a, id, index); err != nil {
		http.Error(w, "delete conversation: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleGenerateFollowup generates a follow-up message (non-streaming).
func (a *App) handleGenerateFollowup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	contact, err := GetContact(a, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	invoker := LLMInvoker(InvokeDeepseekApi)
	var streamInvoker StreamingLLMInvoker
	apiKey := a.Config.DeepSeekApiKey
	model := a.Config.DeepSeekModel
	if a.Config.Backend == "kimi" {
		invoker = InvokeKimiApi
		apiKey = a.Config.KimiApiKey
		model = a.Config.KimiModel
	}

	result, err := GenerateFollowup(r.Context(), invoker, streamInvoker, apiKey, model, &a.Client, *contact, a.NetworkingPromptConfig, nil)
	if err != nil {
		http.Error(w, "generate followup: "+err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, result)
}

// handleGenerateFollowupStream generates a follow-up message with SSE streaming.
func (a *App) handleGenerateFollowupStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	contact, err := GetContact(a, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	flusher := initSSE(w)
	if flusher == nil {
		return
	}

	invoker := LLMInvoker(InvokeDeepseekApi)
	streamInvoker := StreamingLLMInvoker(InvokeDeepseekApiStream)
	apiKey := a.Config.DeepSeekApiKey
	model := a.Config.DeepSeekModel
	if a.Config.Backend == "kimi" {
		invoker = InvokeKimiApi
		streamInvoker = InvokeKimiApiStream
		apiKey = a.Config.KimiApiKey
		model = a.Config.KimiModel
	}

	writeSSE(w, flusher, ProgressEvent{Stage: StageGenerating, Message: "Generating follow-up message\u2026"})

	onDelta := func(delta string) {
		writeSSE(w, flusher, ProgressEvent{Stage: StageContent, Delta: delta})
	}

	result, err := GenerateFollowup(r.Context(), invoker, streamInvoker, apiKey, model, &a.Client, *contact, a.NetworkingPromptConfig, onDelta)
	if err != nil {
		writeSSE(w, flusher, ProgressEvent{Stage: StageError, Message: "generate followup: " + err.Error()})
		return
	}

	// Encode result as JSON and send in the dir field (reusing ProgressEvent.Dir for result payload)
	// We send a final event with the full result encoded in message.
	writeSSE(w, flusher, ProgressEvent{Stage: StageComplete, Message: result.Message})
}

// handleOverdueFollowups returns contacts with overdue follow-up dates.
func (a *App) handleOverdueFollowups(w http.ResponseWriter, r *http.Request) {
	contacts, err := ListOverdueFollowups(a)
	if err != nil {
		http.Error(w, "list overdue: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if contacts == nil {
		contacts = []ContactMeta{}
	}
	writeJSON(w, contacts)
}

// handleUpcomingFollowups returns contacts with follow-up dates in the next N days.
func (a *App) handleUpcomingFollowups(w http.ResponseWriter, r *http.Request) {
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}
	contacts, err := ListUpcomingFollowups(a, days)
	if err != nil {
		http.Error(w, "list upcoming: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if contacts == nil {
		contacts = []ContactMeta{}
	}
	writeJSON(w, contacts)
}

// handleGetNetworkingPromptConfig returns the current networking prompt config.
func (a *App) handleGetNetworkingPromptConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, a.NetworkingPromptConfig)
}

// handleUpdateNetworkingPromptConfig updates and persists the networking prompt config.
func (a *App) handleUpdateNetworkingPromptConfig(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SystemPrompt *string `json:"system_prompt"`
		TaskList     *string `json:"task_list"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	if body.SystemPrompt != nil {
		a.NetworkingPromptConfig.SystemPrompt = *body.SystemPrompt
	}
	if body.TaskList != nil {
		a.NetworkingPromptConfig.TaskList = *body.TaskList
	}
	path := filepath.Join(a.Paths.Config, "networking_prompt.json")
	if err := SaveNetworkingPromptConfig(path, a.NetworkingPromptConfig); err != nil {
		http.Error(w, "save networking prompt config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
