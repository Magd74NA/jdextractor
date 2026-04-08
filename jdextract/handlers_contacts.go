package jdextract

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
)

// contactResponse is the wire type for GET /api/contacts. It embeds ContactMeta
// and adds Dir as a JSON field — Dir is excluded from the stored meta.json
// (json:"-") so it must be lifted here for the frontend to use as an ID.
type contactResponse struct {
	ContactMeta
	Dir string `json:"dir"`
}

// handleListContacts returns all contacts as JSON,
// filtered by any query parameters present in the request.
func (a *App) handleListContacts(w http.ResponseWriter, r *http.Request) {
	contacts, err := ListContacts(a)
	if err != nil {
		http.Error(w, "list contacts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	contacts = applyContactFilters(contacts, r)
	out := make([]contactResponse, len(contacts))
	for i, c := range contacts {
		out[i] = contactResponse{ContactMeta: c, Dir: c.Dir}
	}
	writeJSON(w, out)
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
	writeJSON(w, contactResponse{ContactMeta: *contact, Dir: contact.Dir})
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

// handleAddConversation appends a conversation thread to a contact.
func (a *App) handleAddConversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	var conv Conversation
	if !decodeBody(w, r, &conv) {
		return
	}
	if conv.Summary == "" {
		http.Error(w, "summary is required", http.StatusBadRequest)
		return
	}
	if err := AddConversation(a, id, conv); err != nil {
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

// handleAddMessage appends a message to an existing conversation thread.
func (a *App) handleAddMessage(w http.ResponseWriter, r *http.Request) {
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
	var msg Message
	if !decodeBody(w, r, &msg) {
		return
	}
	if msg.Content == "" || msg.Sender == "" {
		http.Error(w, "sender and content are required", http.StatusBadRequest)
		return
	}
	if err := AddMessage(a, id, index, msg); err != nil {
		http.Error(w, "add message: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteMessage removes a message by index from a conversation thread.
func (a *App) handleDeleteMessage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	convIdx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil || convIdx < 0 {
		http.Error(w, "invalid conversation index", http.StatusBadRequest)
		return
	}
	msgIdx, err := strconv.Atoi(r.PathValue("msgIndex"))
	if err != nil || msgIdx < 0 {
		http.Error(w, "invalid message index", http.StatusBadRequest)
		return
	}
	if err := DeleteMessage(a, id, convIdx, msgIdx); err != nil {
		http.Error(w, "delete message: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleUpdateConversationSummary updates the summary of a conversation.
func (a *App) handleUpdateConversationSummary(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	index, err := strconv.Atoi(r.PathValue("index"))
	if err != nil || index < 0 {
		http.Error(w, "invalid conversation index", http.StatusBadRequest)
		return
	}
	var body struct {
		Summary string `json:"summary"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	if err := UpdateConversationSummary(a, id, index, body.Summary); err != nil {
		http.Error(w, "update summary: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleSummarizeConversation uses the LLM to generate a summary from the conversation messages.
func (a *App) handleSummarizeConversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	index, err := strconv.Atoi(r.PathValue("index"))
	if err != nil || index < 0 {
		http.Error(w, "invalid conversation index", http.StatusBadRequest)
		return
	}
	contact, err := GetContact(a, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if index >= len(contact.Conversations) {
		http.Error(w, "conversation index out of range", http.StatusBadRequest)
		return
	}
	conv := contact.Conversations[index]
	if len(conv.Messages) == 0 {
		http.Error(w, "no messages to summarize", http.StatusBadRequest)
		return
	}

	b := a.Backend()

	summary, err := SummarizeConversation(r.Context(), b.Invoker, b.APIKey, b.Model, &a.Client, conv)
	if err != nil {
		http.Error(w, "summarize: "+err.Error(), http.StatusBadGateway)
		return
	}
	if err := UpdateConversationSummary(a, id, index, summary); err != nil {
		http.Error(w, "save summary: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"summary": summary})
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

	var body struct {
		Guidance string `json:"guidance"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	b := a.Backend()

	result, err := GenerateFollowup(r.Context(), b.Invoker, nil, b.APIKey, b.Model, &a.Client, *contact, a.NetworkingPromptConfig, body.Guidance, nil)
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

	var body struct {
		Guidance string `json:"guidance"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	flusher := initSSE(w)
	if flusher == nil {
		return
	}

	b := a.Backend()

	writeSSE(w, flusher, ProgressEvent{Stage: StageGenerating, Message: "Generating follow-up message\u2026"})

	onDelta := func(delta string) {
		writeSSE(w, flusher, ProgressEvent{Stage: StageContent, Delta: delta})
	}

	result, err := GenerateFollowup(r.Context(), b.Invoker, b.StreamInvoker, b.APIKey, b.Model, &a.Client, *contact, a.NetworkingPromptConfig, body.Guidance, onDelta)
	if err != nil {
		writeSSE(w, flusher, ProgressEvent{Stage: StageError, Message: "generate followup: " + err.Error()})
		return
	}

	// JSON-encode the full result into Message so the frontend can parse all fields.
	resultJSON, _ := json.Marshal(result)
	writeSSE(w, flusher, ProgressEvent{Stage: StageComplete, Message: string(resultJSON)})
}

// handleSendFollowup logs a sent follow-up message and optionally advances the follow-up date.
func (a *App) handleSendFollowup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}
	var body struct {
		Content          string `json:"content"`
		Channel          string `json:"channel"`
		NextFollowUpDate string `json:"next_followup_date"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	if body.Content == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}
	if body.Channel == "" {
		http.Error(w, "channel is required", http.StatusBadRequest)
		return
	}
	updated, err := SendFollowup(a, id, SendFollowupInput{
		Content:          body.Content,
		Channel:          body.Channel,
		NextFollowUpDate: body.NextFollowUpDate,
	})
	if err != nil {
		if updated == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	writeJSON(w, contactResponse{ContactMeta: *updated, Dir: id})
}

func contactsToResponses(contacts []ContactMeta) []contactResponse {
	out := make([]contactResponse, len(contacts))
	for i, c := range contacts {
		out[i] = contactResponse{ContactMeta: c, Dir: c.Dir}
	}
	return out
}

// handleOverdueFollowups returns contacts with overdue follow-up dates.
func (a *App) handleOverdueFollowups(w http.ResponseWriter, r *http.Request) {
	contacts, err := ListOverdueFollowups(a)
	if err != nil {
		http.Error(w, "list overdue: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, contactsToResponses(contacts))
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
	writeJSON(w, contactsToResponses(contacts))
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
	if err := SaveJSON(path, a.NetworkingPromptConfig, 0600); err != nil {
		http.Error(w, "save networking prompt config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
