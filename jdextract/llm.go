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

const deepseekURL = "https://api.deepseek.com/chat/completions"

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

func InvokeDeepseekApi(ctx context.Context, apiKey string, c *http.Client, backoff int, requestBody json.RawMessage) (string, error) {
	if backoff != 0 {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Duration(backoff) * time.Millisecond):
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deepseekURL, bytes.NewReader(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
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
		return InvokeDeepseekApi(ctx, apiKey, c, backoff, requestBody)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("deepseek api returned status: %d", resp.StatusCode)
	}

	buff, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(buff), nil
}
