package jdextract

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const JINA_URL = "https://r.jina.ai/"

func InitiateClient() (*http.Client, error) {
	client := http.DefaultClient
	err := testInitiateClient(client)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func testInitiateClient(c *http.Client) error {
	testRequest, err := c.Get("https://r.jina.ai/https://example.com")
	if err != nil {
		return err
	}

	_, err = io.ReadAll(testRequest.Body)
	if err != nil {
		return err
	}

	return nil
}

func buildJinaUrl(target string) (*url.URL, error) {
	parsedTarget, err := url.Parse(JINA_URL + target)
	if err != nil {
		return nil, err
	}
	return parsedTarget, nil
}

func FetchJobDescription(ctx context.Context, t string, c *http.Client, backoff int) (string, error) {
	if backoff != 0 {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Duration(backoff) * time.Millisecond):
		}
	}
	target, err := buildJinaUrl(t)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return "", err
	}
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
		return FetchJobDescription(ctx, t, c, backoff)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Request returned with code: %d", resp.StatusCode)
	}

	limit := io.LimitReader(resp.Body, 100000)

	buff, err := io.ReadAll(limit)
	if err != nil {
		return "", err
	}

	return string(buff), nil
}
