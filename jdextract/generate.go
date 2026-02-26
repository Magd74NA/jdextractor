package jdextract

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const systemPrompt = `You are a professional resume writer and career coach.
You will receive a job description as a JSON array of classified lines, a base resume, and optionally a base cover letter.

Your tasks:
1. Extract the company name and role title from the job description.
2. Rewrite the resume to align with the job — keep experience truthful, sharpen bullets to mirror the job's language and priorities.
3. If a base cover letter is provided, draft a tailored cover letter for this role.
4. Rate how well the base resume matches the job requirements on a scale of 1–10 (1 = poor fit, 10 = perfect fit).

Respond using exactly these XML tags, in this order:
<company>company name</company>
<role>role title</role>
<score>integer 1-10</score>
<resume>
full tailored resume text
</resume>
<cover>
tailored cover letter (include ONLY if a base cover letter was provided)
</cover>`

var (
	companyTagRe = regexp.MustCompile(`(?s)<company>(.*?)</company>`)
	roleTagRe    = regexp.MustCompile(`(?s)<role>(.*?)</role>`)
	scoreTagRe   = regexp.MustCompile(`(?s)<score>(\d+)</score>`)
	resumeTagRe  = regexp.MustCompile(`(?s)<resume>(.*?)</resume>`)
	coverTagRe   = regexp.MustCompile(`(?s)<cover>(.*?)</cover>`)
)

func extractTag(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func GenerateAll(
	ctx context.Context,
	apiKey string,
	model string,
	c *http.Client,
	nodes []JobDescriptionNode,
	baseResume string,
	baseCover *string,
) (company, role, resume string, cover *string, score, tokensUsed int, err error) {
	jobJSON, err := json.Marshal(nodes)
	if err != nil {
		return "", "", "", nil, 0, 0, fmt.Errorf("json encode: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("JOB DESCRIPTION:\n")
	sb.WriteString(string(jobJSON))
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
		Stream: false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", "", nil, 0, 0, fmt.Errorf("marshal request: %w", err)
	}

	raw, err := InvokeDeepseekApi(ctx, apiKey, c, 0, json.RawMessage(bodyBytes))
	if err != nil {
		return "", "", "", nil, 0, 0, err
	}

	var apiResp deepseekResponse
	if err := json.Unmarshal([]byte(raw), &apiResp); err != nil {
		return "", "", "", nil, 0, 0, fmt.Errorf("decode api response: %w", err)
	}
	if len(apiResp.Choices) == 0 {
		return "", "", "", nil, 0, 0, fmt.Errorf("api returned no choices")
	}

	content := apiResp.Choices[0].Message.Content

	company = extractTag(companyTagRe, content)
	role = extractTag(roleTagRe, content)
	resume = extractTag(resumeTagRe, content)

	if scoreStr := extractTag(scoreTagRe, content); scoreStr != "" {
		score, _ = strconv.Atoi(scoreStr)
	}

	if baseCover != nil {
		if c := extractTag(coverTagRe, content); c != "" {
			cover = &c
		}
	}

	if company == "" || role == "" || resume == "" {
		return "", "", "", nil, 0, 0, fmt.Errorf("llm response missing required fields (company=%q role=%q resume_len=%d)", company, role, len(resume))
	}

	return company, role, resume, cover, score, apiResp.Usage.TotalTokens, nil
}
