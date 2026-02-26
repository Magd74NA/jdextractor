package jdextract

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	toon "github.com/toon-format/toon-go"
)

type jobDescriptionPayload struct {
	Nodes []JobDescriptionNode `toon:"nodes"`
}

const systemPrompt = `You are a professional resume writer and career coach.
You will receive a job description in TOON format, a base resume, and optionally a base cover letter.

Your tasks:
1. Extract the company name and role title from the job description.
2. Rewrite the resume to align with the job — keep experience truthful, sharpen bullets to mirror the job's language and priorities.
3. If a base cover letter is provided, draft a tailored cover letter for this role.

Respond with a JSON object with exactly these keys:
- "Result": object with "company" (string) and "role" (string)
- "Resume": full tailored resume text (string)
- "Cover": tailored cover letter (string) — include ONLY if a base cover letter was provided`

func GenerateAll(
	ctx context.Context,
	apiKey string,
	model string,
	c *http.Client,
	nodes []JobDescriptionNode,
	baseResume string,
	baseCover *string,
) (company, role, resume string, cover *string, tokensUsed int, err error) {
	jobTOON, err := toon.MarshalString(jobDescriptionPayload{Nodes: nodes})
	if err != nil {
		return "", "", "", nil, 0, fmt.Errorf("toon encode: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("JOB DESCRIPTION:\n")
	sb.WriteString(jobTOON)
	sb.WriteString("\n\nBASE RESUME:\n")
	sb.WriteString(baseResume)
	if baseCover != nil {
		sb.WriteString("\n\nBASE COVER LETTER:\n")
		sb.WriteString(*baseCover)
	}

	reqBody := deepseekRequest{
		Model: model,
		Messages: []deepseekMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: sb.String()},
		},
		ResponseFormat: map[string]string{"type": "json_object"},
		Stream:         false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", "", nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	raw, err := InvokeDeepseekApi(ctx, apiKey, c, 0, json.RawMessage(bodyBytes))
	if err != nil {
		return "", "", "", nil, 0, err
	}

	var apiResp deepseekResponse
	if err := json.Unmarshal([]byte(raw), &apiResp); err != nil {
		return "", "", "", nil, 0, fmt.Errorf("decode api response: %w", err)
	}
	if len(apiResp.Choices) == 0 {
		return "", "", "", nil, 0, fmt.Errorf("api returned no choices")
	}

	var result struct {
		Result struct {
			Company string `json:"company"`
			Role    string `json:"role"`
		} `json:"Result"`
		Resume string  `json:"Resume"`
		Cover  *string `json:"Cover"`
	}
	if err := json.Unmarshal([]byte(apiResp.Choices[0].Message.Content), &result); err != nil {
		return "", "", "", nil, 0, fmt.Errorf("decode llm json output: %w", err)
	}

	company = result.Result.Company
	role = result.Result.Role
	resume = result.Resume
	if baseCover != nil {
		cover = result.Cover
	}

	return company, role, resume, cover, apiResp.Usage.TotalTokens, nil
}
