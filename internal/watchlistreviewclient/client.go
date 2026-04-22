package watchlistreviewclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Review(ctx context.Context, req ReviewRequest, correlationID string) (ReviewResponse, error) {
	if c == nil || c.baseURL == "" {
		return ReviewResponse{}, fmt.Errorf("watchlist review base url is not configured")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/reviews", bytes.NewReader(body))
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if strings.TrimSpace(correlationID) != "" {
		httpReq.Header.Set("X-Correlation-ID", strings.TrimSpace(correlationID))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return ReviewResponse{}, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	var decoded ReviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return ReviewResponse{}, fmt.Errorf("decode response: %w", err)
	}
	return decoded, nil
}
