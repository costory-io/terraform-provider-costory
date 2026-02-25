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
	billingDatasourceTypeGCP = "GCP"
	billingDatasourceTypeAWS = "AWS"
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
	token      string
	httpClient httpDoer
}

// ServiceAccountResponse represents the service-account payload returned by the API.
type ServiceAccountResponse struct {
	ServiceAccount string   `json:"service_account"`
	SubIDs         []string `json:"sub_ids"`
}

type serviceAccountAPIResponse struct {
	ServiceAccount      string   `json:"service_account"`
	ServiceAccountCamel string   `json:"serviceAccount"`
	ServiceAccountEmail string   `json:"serviceAccountEmail"`
	SubIDs              []string `json:"sub_ids"`
	SubIDsCamel         []string `json:"subIds"`
}

// GCPBillingDatasourceRequest is the Terraform input used to create/validate a GCP billing datasource.
type GCPBillingDatasourceRequest struct {
	Name              string
	BQURI             string
	IsDetailedBilling *bool
	StartDate         *string
	EndDate           *string
}

// GCPBillingDatasource is the normalized datasource payload returned by the Costory API.
type GCPBillingDatasource struct {
	ID                string
	Type              string
	Status            *string
	Name              string
	BQURI             string
	IsDetailedBilling *bool
	StartDate         *string
	EndDate           *string
}

// AWSBillingDatasourceRequest is the Terraform input used to create/validate an AWS billing datasource.
type AWSBillingDatasourceRequest struct {
	Name                string
	BucketName          string
	RoleARN             string
	Prefix              string
	EKSSplitDataEnabled *bool
	StartDate           *string
	EndDate             *string
	EKSSplit            *bool
}

// AWSBillingDatasource is the normalized datasource payload returned by the Costory API.
type AWSBillingDatasource struct {
	ID                  string
	Type                string
	Status              *string
	Name                string
	BucketName          string
	RoleARN             string
	Prefix              string
	EKSSplitDataEnabled *bool
	StartDate           *string
	EndDate             *string
	EKSSplit            *bool
}

type gcpBillingDatasourceAPIRequest struct {
	Type              string  `json:"type"`
	Name              string  `json:"name"`
	BQTablePath       string  `json:"bqTablePath"`
	IsDetailedBilling *bool   `json:"isDetailedBilling,omitempty"`
	StartDate         *string `json:"startDate,omitempty"`
	EndDate           *string `json:"endDate,omitempty"`
}

type gcpBillingDatasourceAPIResponse struct {
	ID                string  `json:"id"`
	Type              string  `json:"type"`
	Status            *string `json:"status"`
	Name              string  `json:"name"`
	BQURI             string  `json:"bqUri"`
	IsDetailedBilling *bool   `json:"isDetailedBilling"`
	StartDate         *string `json:"startDate"`
	EndDate           *string `json:"endDate"`
}

type awsBillingDatasourceAPIRequest struct {
	Type                string  `json:"type"`
	Name                string  `json:"name"`
	BucketName          string  `json:"bucketName"`
	RoleARN             string  `json:"roleArn"`
	Prefix              string  `json:"prefix"`
	EKSSplitDataEnabled *bool   `json:"eksSplitDataEnabled,omitempty"`
	StartDate           *string `json:"startDate,omitempty"`
	EndDate             *string `json:"endDate,omitempty"`
	EKSSplit            *bool   `json:"eksSplit,omitempty"`
}

type awsBillingDatasourceAPIResponse struct {
	ID                  string  `json:"id"`
	Type                string  `json:"type"`
	Status              *string `json:"status"`
	Name                string  `json:"name"`
	BucketName          string  `json:"bucketName"`
	RoleARN             string  `json:"roleArn"`
	Prefix              string  `json:"prefix"`
	EKSSplitDataEnabled *bool   `json:"eksSplitDataEnabled"`
	StartDate           *string `json:"startDate"`
	EndDate             *string `json:"endDate"`
	EKSSplit            *bool   `json:"eksSplit"`
}

type apiErrorResponse struct {
	Error  string `json:"error"`
	Reason string `json:"reason"`
}

// NewClient creates a new Costory API client.
func NewClient(baseURL, token string, httpClient httpDoer) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: httpClient,
	}
}

// GetServiceAccount fetches service-account data for the configured Costory tenant.
func (c *Client) GetServiceAccount(ctx context.Context) (*ServiceAccountResponse, error) {
	body, statusCode, err := doEndpoint(ctx, c, endpointGetServiceAccount, noRequest{})
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out serviceAccountAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := &ServiceAccountResponse{
		ServiceAccount: firstNonEmptyString(out.ServiceAccount, out.ServiceAccountCamel, out.ServiceAccountEmail),
		SubIDs:         firstStringSlice(out.SubIDs, out.SubIDsCamel),
	}
	if normalized.SubIDs == nil {
		normalized.SubIDs = []string{}
	}

	return normalized, nil
}

// ValidateGCPBillingDatasource validates a GCP billing datasource before creation.
func (c *Client) ValidateGCPBillingDatasource(ctx context.Context, req GCPBillingDatasourceRequest) error {
	body, statusCode, err := doEndpoint(ctx, c, endpointValidateGCPBillingDatasource, req.toAPIRequest())
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
	body, statusCode, err := doEndpoint(ctx, c, endpointCreateGCPBillingDatasource, req.toAPIRequest())
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
	routeParams := billingDatasourceByIDRouteParams{ID: datasourceID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointGetGCPBillingDatasourceByID, routeParams, noRequest{})
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

// ValidateAWSBillingDatasource validates an AWS billing datasource before creation.
func (c *Client) ValidateAWSBillingDatasource(ctx context.Context, req AWSBillingDatasourceRequest) error {
	body, statusCode, err := doEndpoint(ctx, c, endpointValidateAWSBillingDatasource, req.toAPIRequest())
	if err != nil {
		return err
	}

	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return nil
	}

	return unexpectedStatusError(statusCode, body)
}

// CreateAWSBillingDatasource creates an AWS billing datasource and returns its API representation.
func (c *Client) CreateAWSBillingDatasource(ctx context.Context, req AWSBillingDatasourceRequest) (*AWSBillingDatasource, error) {
	body, statusCode, err := doEndpoint(ctx, c, endpointCreateAWSBillingDatasource, req.toAPIRequest())
	if err != nil {
		return nil, err
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out awsBillingDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toAWSBillingDatasource()
	if normalized.ID == "" {
		return nil, errors.New("create response did not include datasource id")
	}

	return normalized, nil
}

// GetAWSBillingDatasource gets an AWS billing datasource by ID.
func (c *Client) GetAWSBillingDatasource(ctx context.Context, datasourceID string) (*AWSBillingDatasource, error) {
	routeParams := billingDatasourceByIDRouteParams{ID: datasourceID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointGetAWSBillingDatasourceByID, routeParams, noRequest{})
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if statusCode != http.StatusOK {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out awsBillingDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toAWSBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}

	return normalized, nil
}

// DeleteBillingDatasource deletes a billing datasource by ID.
func (c *Client) DeleteBillingDatasource(ctx context.Context, datasourceID string) error {
	routeParams := billingDatasourceByIDRouteParams{ID: datasourceID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointDeleteBillingDatasourceByID, routeParams, noRequest{})
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

func doEndpoint[TReq any, TResp any](
	ctx context.Context,
	c *Client,
	endpoint endpointContract[TReq, TResp],
	request TReq,
) ([]byte, int, error) {
	switch endpoint.RequestTransport {
	case requestTransportNone:
		return c.doJSON(ctx, endpoint.Method, endpoint.Path, nil)
	case requestTransportJSONBody:
		return c.doJSON(ctx, endpoint.Method, endpoint.Path, request)
	default:
		return nil, 0, fmt.Errorf("unsupported request transport for %s %s: %s", endpoint.Method, endpoint.Path, endpoint.RequestTransport)
	}
}

func doEndpointWithRouteParams[TParams any, TReq any, TResp any](
	ctx context.Context,
	c *Client,
	endpoint endpointWithRouteParamsContract[TParams, TReq, TResp],
	params TParams,
	request TReq,
) ([]byte, int, error) {
	if endpoint.ParamsTransport != requestTransportRouteParams {
		return nil, 0, fmt.Errorf("unsupported route params transport for endpoint %s", endpoint.Method)
	}

	path := endpoint.Path(params)
	switch endpoint.RequestBodyTransport {
	case requestTransportNone:
		return c.doJSON(ctx, endpoint.Method, path, nil)
	case requestTransportJSONBody:
		return c.doJSON(ctx, endpoint.Method, path, request)
	default:
		return nil, 0, fmt.Errorf("unsupported request transport for %s %s: %s", endpoint.Method, path, endpoint.RequestBodyTransport)
	}
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
		BQTablePath:       r.BQURI,
		IsDetailedBilling: r.IsDetailedBilling,
		StartDate:         r.StartDate,
		EndDate:           r.EndDate,
	}
}

func (r AWSBillingDatasourceRequest) toAPIRequest() awsBillingDatasourceAPIRequest {
	return awsBillingDatasourceAPIRequest{
		Type:                billingDatasourceTypeAWS,
		Name:                r.Name,
		BucketName:          r.BucketName,
		RoleARN:             r.RoleARN,
		Prefix:              r.Prefix,
		EKSSplitDataEnabled: r.EKSSplitDataEnabled,
		StartDate:           r.StartDate,
		EndDate:             r.EndDate,
		EKSSplit:            r.EKSSplit,
	}
}

func (r gcpBillingDatasourceAPIResponse) toGCPBillingDatasource() *GCPBillingDatasource {
	return &GCPBillingDatasource{
		ID:                r.ID,
		Type:              r.Type,
		Status:            r.Status,
		Name:              r.Name,
		BQURI:             r.BQURI,
		IsDetailedBilling: r.IsDetailedBilling,
		StartDate:         r.StartDate,
		EndDate:           r.EndDate,
	}
}

func (r awsBillingDatasourceAPIResponse) toAWSBillingDatasource() *AWSBillingDatasource {
	return &AWSBillingDatasource{
		ID:                  r.ID,
		Type:                r.Type,
		Status:              r.Status,
		Name:                r.Name,
		BucketName:          r.BucketName,
		RoleARN:             r.RoleARN,
		Prefix:              r.Prefix,
		EKSSplitDataEnabled: r.EKSSplitDataEnabled,
		StartDate:           r.StartDate,
		EndDate:             r.EndDate,
		EKSSplit:            r.EKSSplit,
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstStringSlice(values ...[]string) []string {
	for _, value := range values {
		if value != nil {
			return append([]string(nil), value...)
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
	var apiErr apiErrorResponse
	if err := json.Unmarshal(body, &apiErr); err == nil {
		apiErr.Error = strings.TrimSpace(apiErr.Error)
		apiErr.Reason = strings.TrimSpace(apiErr.Reason)
		if apiErr.Error != "" || apiErr.Reason != "" {
			return fmt.Errorf("unexpected status code %d: error=%s reason=%s", statusCode, apiErr.Error, apiErr.Reason)
		}
	}

	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(statusCode)
	}

	return fmt.Errorf("unexpected status code %d: %s", statusCode, message)
}
