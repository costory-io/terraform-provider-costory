package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const contextEndpoint = "/api/v1/terraform/context"

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a lightweight Costory API client used by the provider.
type Client struct {
	baseURL    string
	slug       string
	token      string
	httpClient httpDoer
}

// ContextResponse represents the Costory context payload returned by the API.
type ContextResponse struct {
	ServiceAccount string   `json:"service_account"`
	SubIDs         []string `json:"sub_ids"`
}

// NewClient creates a new Costory API client.
func NewClient(baseURL, slug, token string, httpClient httpDoer) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    baseURL,
		slug:       slug,
		token:      token,
		httpClient: httpClient,
	}
}

// GetContext fetches service-account context for the configured Costory tenant.
func (c *Client) GetContext(ctx context.Context) (*ContextResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(contextEndpoint), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-Costory-Slug", c.slug)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out ContextResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	if out.SubIDs == nil {
		out.SubIDs = []string{}
	}

	return &out, nil
}

func (c *Client) endpoint(path string) string {
	base := strings.TrimRight(c.baseURL, "/")
	return base + "/" + strings.TrimLeft(path, "/")
}
