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

const responseFormat = `Respond using exactly these XML tags, in this order:
<company>company name</company>
<role>role title</role>
<score>integer 1-10</score>
<resume>
full tailored resume text
</resume>
<cover>
tailored cover letter (include ONLY if a base cover letter was provided)
</cover>`

// LLMInvoker is a function that posts a JSON request body to an LLM endpoint
// and returns the raw response body. Use InvokeDeepseekApi or InvokeKimiApi.
type LLMInvoker func(ctx context.Context, apiKey string, c *http.Client, backoff int, body json.RawMessage) (string, error)

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

// GenerateAll sends the parsed job description and base templates to DeepSeek
// and extracts the structured output from its response.
//
// nodes is the filtered AST from Parse; baseResume is required; baseCover is
// optional — pass nil to skip cover letter generation.
//
// The LLM responds in plain text with XML delimiter tags (<company>, <role>,
// <score>, <resume>, <cover>). GenerateAll extracts each field with compiled
// regexps. It returns an error if any of company, role, or resume are empty,
// which surfaces prompt compliance failures rather than silently writing empty
// files. score defaults to 0 on parse failure (non-fatal). cover is nil when
// baseCover is nil or the model omitted the tag.
func GenerateAll(
	ctx context.Context,
	invoker LLMInvoker,
	streamInvoker StreamingLLMInvoker,
	apiKey string,
	model string,
	c *http.Client,
	nodes []JobDescriptionNode,
	baseResume string,
	baseCover *string,
	promptConfig PromptConfig,
	onDelta func(string),
) (company, role, resume string, cover *string, score, tokensUsed int, err error) {
	jobJSON, err := json.Marshal(nodes)
	if err != nil {
		return "", "", "", nil, 0, 0, fmt.Errorf("json encode: %w", err)
	}

	// Build system prompt from config
	systemPrompt := promptConfig.SystemPrompt + "\n\n" + promptConfig.TaskList + "\n\n" + responseFormat

	var sb strings.Builder
	fmt.Fprintf(&sb, "JOB DESCRIPTION:\n%s\n\nBASE RESUME:\n%s", jobJSON, Sanitize(baseResume))
	if baseCover != nil {
		fmt.Fprintf(&sb, "\n\nBASE COVER LETTER:\n%s", Sanitize(*baseCover))
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
		return "", "", "", nil, 0, 0, fmt.Errorf("marshal request: %w", err)
	}

	var content string
	if useStreaming {
		content, err = streamInvoker(ctx, apiKey, c, json.RawMessage(bodyBytes), onDelta)
		if err != nil {
			return "", "", "", nil, 0, 0, err
		}
	} else {
		raw, err := invoker(ctx, apiKey, c, 0, json.RawMessage(bodyBytes))
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
		content = apiResp.Choices[0].Message.Content
		tokensUsed = apiResp.Usage.TotalTokens
	}

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

	return company, role, resume, cover, score, tokensUsed, nil
}
