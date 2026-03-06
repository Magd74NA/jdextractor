package jdextract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
