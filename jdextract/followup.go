package jdextract

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

const networkingResponseFormat = `Respond using exactly these XML tags, in this order:
<subject>email subject line (omit entirely if channel is not email)</subject>
<message>
the full follow-up message text
</message>
<channel>recommended channel: email, linkedin, phone, in-person</channel>
<timing>suggested timing, e.g. "Monday morning" or "within 2 days"</timing>
<notes>brief reasoning for the approach taken</notes>`

// FollowupResult holds the parsed LLM output for a follow-up message.
type FollowupResult struct {
	Subject string `json:"subject,omitempty"`
	Message string `json:"message"`
	Channel string `json:"channel"`
	Timing  string `json:"timing"`
	Notes   string `json:"notes"`
}

var (
	followupSubjectRe = regexp.MustCompile(`(?s)<subject>(.*?)</subject>`)
	followupMessageRe = regexp.MustCompile(`(?s)<message>(.*?)</message>`)
	followupChannelRe = regexp.MustCompile(`(?s)<channel>(.*?)</channel>`)
	followupTimingRe  = regexp.MustCompile(`(?s)<timing>(.*?)</timing>`)
	followupNotesRe   = regexp.MustCompile(`(?s)<notes>(.*?)</notes>`)
)

// SummarizeConversation uses the LLM to generate a concise summary from a conversation's messages.
func SummarizeConversation(
	ctx context.Context,
	invoker LLMInvoker,
	apiKey string,
	model string,
	c *http.Client,
	conv Conversation,
) (string, error) {
	var sb strings.Builder
	sb.WriteString("Summarize the following conversation thread in 1-2 sentences. Be concise and capture the key points.\n\n")
	for _, msg := range conv.Messages {
		sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", msg.Date, msg.Sender, msg.Content))
	}

	reqBody := deepseekRequest{
		Model: model,
		Messages: []deepseekMessage{
			{Role: "system", Content: "You are a concise summarizer. Respond with only the summary, no preamble."},
			{Role: "user", Content: sb.String()},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	raw, err := invoker(ctx, apiKey, c, 0, json.RawMessage(bodyBytes))
	if err != nil {
		return "", err
	}
	var apiResp deepseekResponse
	if err := json.Unmarshal([]byte(raw), &apiResp); err != nil {
		return "", fmt.Errorf("decode api response: %w", err)
	}
	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("api returned no choices")
	}
	return strings.TrimSpace(apiResp.Choices[0].Message.Content), nil
}

// GenerateFollowup builds a prompt from contact context and conversation history,
// calls the LLM, and parses the XML-tagged response. Follows the same pattern as GenerateAll.
func GenerateFollowup(
	ctx context.Context,
	invoker LLMInvoker,
	streamInvoker StreamingLLMInvoker,
	apiKey string,
	model string,
	c *http.Client,
	contact ContactMeta,
	promptConfig NetworkingPromptConfig,
	onDelta func(string),
) (*FollowupResult, error) {
	systemPrompt := promptConfig.SystemPrompt + "\n\n" + promptConfig.TaskList + "\n\n" + networkingResponseFormat

	var sb strings.Builder
	sb.WriteString("CONTACT:\n")
	sb.WriteString(fmt.Sprintf("Name: %s\n", contact.Name))
	if contact.Company != "" {
		sb.WriteString(fmt.Sprintf("Company: %s\n", contact.Company))
	}
	if contact.Role != "" {
		sb.WriteString(fmt.Sprintf("Role: %s\n", contact.Role))
	}
	if contact.Source != "" {
		sb.WriteString(fmt.Sprintf("How we met: %s\n", contact.Source))
	}
	sb.WriteString(fmt.Sprintf("Relationship status: %s\n", contact.Status))
	if contact.Notes != "" {
		sb.WriteString(fmt.Sprintf("Notes: %s\n", contact.Notes))
	}
	if len(contact.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(contact.Tags, ", ")))
	}

	sb.WriteString("\nCONVERSATION HISTORY:\n")
	if len(contact.Conversations) == 0 {
		sb.WriteString("No prior conversations logged.\n")
	} else {
		// Include summary for each conversation thread
		for i, conv := range contact.Conversations {
			channel := conv.Channel
			if channel == "" {
				channel = "unknown"
			}
			sb.WriteString(fmt.Sprintf("Thread %d (%s): %s\n", i+1, channel, conv.Summary))
		}

		// Include last 5 messages from the most recent conversation for full context
		latest := contact.Conversations[len(contact.Conversations)-1]
		if len(latest.Messages) > 0 {
			sb.WriteString("\nRECENT MESSAGES (latest thread):\n")
			start := 0
			if len(latest.Messages) > 5 {
				start = len(latest.Messages) - 5
			}
			for _, msg := range latest.Messages[start:] {
				sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", msg.Date, msg.Sender, msg.Content))
			}
		}
	}

	useStreaming := streamInvoker != nil && onDelta != nil

	reqBody := deepseekRequest{
		Model: model,
		Messages: []deepseekMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: sb.String()},
		},
		Stream: useStreaming,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var content string
	if useStreaming {
		content, err = streamInvoker(ctx, apiKey, c, json.RawMessage(bodyBytes), onDelta)
		if err != nil {
			return nil, err
		}
	} else {
		raw, err := invoker(ctx, apiKey, c, 0, json.RawMessage(bodyBytes))
		if err != nil {
			return nil, err
		}
		var apiResp deepseekResponse
		if err := json.Unmarshal([]byte(raw), &apiResp); err != nil {
			return nil, fmt.Errorf("decode api response: %w", err)
		}
		if len(apiResp.Choices) == 0 {
			return nil, fmt.Errorf("api returned no choices")
		}
		content = apiResp.Choices[0].Message.Content
	}

	result := &FollowupResult{
		Subject: strings.TrimSpace(extractTag(followupSubjectRe, content)),
		Message: strings.TrimSpace(extractTag(followupMessageRe, content)),
		Channel: strings.TrimSpace(extractTag(followupChannelRe, content)),
		Timing:  strings.TrimSpace(extractTag(followupTimingRe, content)),
		Notes:   strings.TrimSpace(extractTag(followupNotesRe, content)),
	}

	if result.Message == "" {
		return nil, fmt.Errorf("llm response missing required <message> field")
	}

	return result, nil
}
