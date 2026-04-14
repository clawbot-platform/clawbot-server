package identityclient

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

const (
	defaultHTTPTimeout = 5 * time.Second
)

type Client struct {
	baseURL       string
	defaultTenant string
	httpClient    *http.Client
}

func New(baseURL string, timeout time.Duration, defaultTenant string) *Client {
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}

	return &Client{
		baseURL:       strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		defaultTenant: strings.TrimSpace(defaultTenant),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) Healthz(ctx context.Context) error {
	return c.doJSON(ctx, http.MethodGet, "/healthz", nil, nil, "", "")
}

func (c *Client) Compare(ctx context.Context, req CompareRequest, correlationID string, caseID string) (CompareResponse, error) {
	req.TenantID = c.resolveTenantID(req.TenantID)
	if req.TenantID == "" {
		return CompareResponse{}, fmt.Errorf("compare request requires tenant_id")
	}

	var decoded CompareResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/compare", req, &decoded, correlationID, caseID); err != nil {
		return CompareResponse{}, err
	}

	return decoded, nil
}

func (c *Client) ScreenOFAC(ctx context.Context, req ScreenOFACRequest, correlationID string) (ScreenOFACResponse, error) {
	req.TenantID = c.resolveTenantID(req.TenantID)
	if req.TenantID == "" {
		return ScreenOFACResponse{}, fmt.Errorf("ofac screening request requires tenant_id")
	}

	var decoded ScreenOFACResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/watchlist/ofac/screenings", req, &decoded, correlationID, req.CaseID); err != nil {
		return ScreenOFACResponse{}, err
	}

	return decoded, nil
}

func (c *Client) resolveTenantID(requestTenant string) string {
	if trimmed := strings.TrimSpace(requestTenant); trimmed != "" {
		return trimmed
	}
	return c.defaultTenant
}

func (c *Client) doJSON(
	ctx context.Context,
	method string,
	path string,
	requestBody any,
	responseBody any,
	correlationID string,
	caseID string,
) error {
	if c.baseURL == "" {
		return fmt.Errorf("identity base url is not configured")
	}

	var bodyReader io.Reader
	if requestBody != nil {
		payload, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("marshal %s %s request: %w", method, path, err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build %s %s request: %w", method, path, err)
	}
	req.Header.Set("Accept", "application/json")
	if requestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(correlationID) != "" {
		req.Header.Set("X-Correlation-ID", strings.TrimSpace(correlationID))
	}
	if strings.TrimSpace(caseID) != "" {
		req.Header.Set("X-Case-ID", strings.TrimSpace(caseID))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute %s %s request: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s %s failed: %w", method, path, readStatusError(resp))
	}

	if responseBody == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(responseBody); err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("decode %s %s response: %w", method, path, err)
	}

	return nil
}

func readStatusError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	text := strings.TrimSpace(string(body))
	if text == "" {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return fmt.Errorf("status %d: %s", resp.StatusCode, text)
}
