package jdextract

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	deepseekURL = "https://api.deepseek.com/chat/completions"
	kimiURL     = "https://inference.baseten.co/v1/chat/completions"
)

type deepseekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepseekRequest struct {
	Model    string            `json:"model"`
	Messages []deepseekMessage `json:"messages"`
	Stream   bool              `json:"stream"`
}

type deepseekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// invokeAPI posts requestBody to url with the given Authorization header value
// and returns the raw JSON response body. It retries on HTTP 429 with exponential
// backoff (500ms → 2.5s → 12.5s); fails if the next delay would exceed 10 seconds.
func invokeAPI(ctx context.Context, url, authHeader string, c *http.Client, backoff int, requestBody json.RawMessage) (string, error) {
	if backoff != 0 {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Duration(backoff) * time.Millisecond):
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		if backoff == 0 {
			backoff = 500
		}
		backoff *= 5
		if backoff > 10000 {
			return "", fmt.Errorf("rate limited: max retries exceeded")
		}
		return invokeAPI(ctx, url, authHeader, c, backoff, requestBody)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api returned status: %d", resp.StatusCode)
	}

	buff, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(buff), nil
}

// InvokeDeepseekApi posts requestBody to the DeepSeek chat completions endpoint.
// Pass backoff=0 on the first call; retries are handled internally.
func InvokeDeepseekApi(ctx context.Context, apiKey string, c *http.Client, backoff int, requestBody json.RawMessage) (string, error) {
	return invokeAPI(ctx, deepseekURL, "Bearer "+apiKey, c, backoff, requestBody)
}

// InvokeKimiApi posts requestBody to the Kimi K2.5 endpoint on Baseten.
// Pass backoff=0 on the first call; retries are handled internally.
func InvokeKimiApi(ctx context.Context, apiKey string, c *http.Client, backoff int, requestBody json.RawMessage) (string, error) {
	return invokeAPI(ctx, kimiURL, "Api-Key "+apiKey, c, backoff, requestBody)
}

// streamChunk is an OpenAI-compatible streaming chunk.
type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// StreamingLLMInvoker posts a streaming request and calls onDelta for each
// content token. It returns the fully accumulated content string.
type StreamingLLMInvoker func(ctx context.Context, apiKey string, c *http.Client, requestBody json.RawMessage, onDelta func(string)) (string, error)

// invokeAPIStream posts requestBody to url with streaming enabled and calls
// onDelta for each content delta. Returns the full accumulated content.
func invokeAPIStream(ctx context.Context, url, authHeader string, c *http.Client, requestBody json.RawMessage, onDelta func(string)) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("api returned status %d: %s", resp.StatusCode, string(body))
	}

	var sb strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				fmt.Fprintf(&sb, "%s", delta)
				onDelta(delta)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("stream read error: %w", err)
	}
	return sb.String(), nil
}

// InvokeDeepseekApiStream calls the DeepSeek API with streaming enabled.
func InvokeDeepseekApiStream(ctx context.Context, apiKey string, c *http.Client, requestBody json.RawMessage, onDelta func(string)) (string, error) {
	return invokeAPIStream(ctx, deepseekURL, "Bearer "+apiKey, c, requestBody, onDelta)
}

// InvokeKimiApiStream calls the Kimi API with streaming enabled.
func InvokeKimiApiStream(ctx context.Context, apiKey string, c *http.Client, requestBody json.RawMessage, onDelta func(string)) (string, error) {
	return invokeAPIStream(ctx, kimiURL, "Api-Key "+apiKey, c, requestBody, onDelta)
}
