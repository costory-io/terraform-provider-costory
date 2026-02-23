package costoryapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	billingDatasourceTypeGCP = "gcp"
	maxRetryAttempts         = 4
	maxResponseBodyBytes     = 1024 * 1024
)

// ErrNotFound is returned when the requested Costory resource does not exist.
var ErrNotFound = errors.New("costory resource not found")

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

// ServiceAccountResponse represents the service-account payload returned by the API.
type ServiceAccountResponse struct {
	ServiceAccount string   `json:"service_account"`
	SubIDs         []string `json:"sub_ids"`
}

// GCPBillingDatasourceRequest is the Terraform input used to create/validate a GCP billing datasource.
type GCPBillingDatasourceRequest struct {
	Name              string
	BQTablePath       string
	IsDetailedBilling *bool
	StartDate         *string
	EndDate           *string
}

// GCPBillingDatasource is the normalized datasource payload returned by the Costory API.
type GCPBillingDatasource struct {
	ID                string
	Name              string
	BQTablePath       string
	IsDetailedBilling *bool
	StartDate         *string
	EndDate           *string
}

type gcpBillingDatasourceAPIRequest struct {
	Type              string  `json:"type"`
	Name              string  `json:"name"`
	BQTablePath       string  `json:"bq_table_path"`
	IsDetailedBilling *bool   `json:"is_detailed_billing,omitempty"`
	StartDate         *string `json:"start_date,omitempty"`
	EndDate           *string `json:"end_date,omitempty"`
}

type gcpBillingDatasourceAPIResponse struct {
	ID                     string  `json:"id"`
	Name                   string  `json:"name"`
	BQTablePath            string  `json:"bq_table_path"`
	BQTablePathCamel       string  `json:"bqTablePath"`
	IsDetailedBilling      *bool   `json:"is_detailed_billing"`
	IsDetailedBillingCamel *bool   `json:"isDetailedBilling"`
	StartDate              *string `json:"start_date"`
	StartDateCamel         *string `json:"startDate"`
	EndDate                *string `json:"end_date"`
	EndDateCamel           *string `json:"endDate"`
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

// GetServiceAccount fetches service-account data for the configured Costory tenant.
func (c *Client) GetServiceAccount(ctx context.Context) (*ServiceAccountResponse, error) {
	body, statusCode, err := c.doJSON(ctx, http.MethodGet, routeServiceAccount, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out ServiceAccountResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	if out.SubIDs == nil {
		out.SubIDs = []string{}
	}

	return &out, nil
}

// ValidateGCPBillingDatasource validates a GCP billing datasource before creation.
func (c *Client) ValidateGCPBillingDatasource(ctx context.Context, req GCPBillingDatasourceRequest) error {
	body, statusCode, err := c.doJSON(ctx, http.MethodPost, routeBillingDatasourceValidate, req.toAPIRequest())
	if err != nil {
		return err
	}

	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return nil
	}

	return unexpectedStatusError(statusCode, body)
}

// CreateGCPBillingDatasource creates a GCP billing datasource and returns its API representation.
func (c *Client) CreateGCPBillingDatasource(ctx context.Context, req GCPBillingDatasourceRequest) (*GCPBillingDatasource, error) {
	body, statusCode, err := c.doJSON(ctx, http.MethodPost, routeBillingDatasourceBase, req.toAPIRequest())
	if err != nil {
		return nil, err
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out gcpBillingDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toGCPBillingDatasource()
	if normalized.ID == "" {
		return nil, errors.New("create response did not include datasource id")
	}

	return normalized, nil
}

// GetGCPBillingDatasource gets a GCP billing datasource by ID.
func (c *Client) GetGCPBillingDatasource(ctx context.Context, datasourceID string) (*GCPBillingDatasource, error) {
	body, statusCode, err := c.doJSON(ctx, http.MethodGet, routeBillingDatasourceByID(datasourceID), nil)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if statusCode != http.StatusOK {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out gcpBillingDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toGCPBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}

	return normalized, nil
}

// DeleteBillingDatasource deletes a billing datasource by ID.
func (c *Client) DeleteBillingDatasource(ctx context.Context, datasourceID string) error {
	body, statusCode, err := c.doJSON(ctx, http.MethodDelete, routeBillingDatasourceByID(datasourceID), nil)
	if err != nil {
		return err
	}

	if statusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if statusCode == http.StatusNoContent || statusCode == http.StatusOK {
		return nil
	}

	return unexpectedStatusError(statusCode, body)
}

func (c *Client) endpoint(path string) string {
	base := strings.TrimRight(c.baseURL, "/")
	return base + "/" + strings.TrimLeft(path, "/")
}

func (c *Client) doJSON(ctx context.Context, method, path string, requestBody any) ([]byte, int, error) {
	var payload []byte
	if requestBody != nil {
		var err error
		payload, err = json.Marshal(requestBody)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request body: %w", err)
		}
	}

	for attempt := range maxRetryAttempts {
		var bodyReader io.Reader
		if payload != nil {
			bodyReader = bytes.NewReader(payload)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.endpoint(path), bodyReader)
		if err != nil {
			return nil, 0, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("X-Costory-Slug", c.slug)
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, 0, fmt.Errorf("execute request: %w", err)
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes))
		closeErr := resp.Body.Close()
		if readErr != nil {
			return nil, 0, fmt.Errorf("read response body: %w", readErr)
		}
		if closeErr != nil {
			return nil, 0, fmt.Errorf("close response body: %w", closeErr)
		}

		if resp.StatusCode >= http.StatusInternalServerError && attempt < maxRetryAttempts-1 {
			if err := waitForRetry(ctx, attempt); err != nil {
				return nil, 0, err
			}
			continue
		}

		return body, resp.StatusCode, nil
	}

	return nil, 0, errors.New("request retries exhausted")
}

func (r GCPBillingDatasourceRequest) toAPIRequest() gcpBillingDatasourceAPIRequest {
	return gcpBillingDatasourceAPIRequest{
		Type:              billingDatasourceTypeGCP,
		Name:              r.Name,
		BQTablePath:       r.BQTablePath,
		IsDetailedBilling: r.IsDetailedBilling,
		StartDate:         r.StartDate,
		EndDate:           r.EndDate,
	}
}

func (r gcpBillingDatasourceAPIResponse) toGCPBillingDatasource() *GCPBillingDatasource {
	bqTablePath := r.BQTablePath
	if bqTablePath == "" {
		bqTablePath = r.BQTablePathCamel
	}

	return &GCPBillingDatasource{
		ID:                r.ID,
		Name:              r.Name,
		BQTablePath:       bqTablePath,
		IsDetailedBilling: firstBoolPtr(r.IsDetailedBilling, r.IsDetailedBillingCamel),
		StartDate:         firstStringPtr(r.StartDate, r.StartDateCamel),
		EndDate:           firstStringPtr(r.EndDate, r.EndDateCamel),
	}
}

func firstBoolPtr(values ...*bool) *bool {
	for _, value := range values {
		if value != nil {
			out := *value
			return &out
		}
	}
	return nil
}

func firstStringPtr(values ...*string) *string {
	for _, value := range values {
		if value != nil {
			out := *value
			return &out
		}
	}
	return nil
}

func waitForRetry(ctx context.Context, attempt int) error {
	backoff := time.Duration(1<<attempt) * 500 * time.Millisecond
	timer := time.NewTimer(backoff)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return fmt.Errorf("retry canceled: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}

func unexpectedStatusError(statusCode int, body []byte) error {
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(statusCode)
	}

	return fmt.Errorf("unexpected status code %d: %s", statusCode, message)
}
