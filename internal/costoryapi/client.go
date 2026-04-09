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
	billingDatasourceTypeGCP          = "GCP"
	billingDatasourceTypeAWS          = "AWS"
	billingDatasourceTypeCursor       = "Cursor"
	billingDatasourceTypeAnthropic    = "Anthropic"
	billingDatasourceTypeElasticCloud = "ElasticCloud"
	billingDatasourceTypeAzure        = "Azure"
	metricsDatasourceTypeS3V2         = "AwsS3V2"
	maxRetryAttempts                  = 4
	maxResponseBodyBytes              = 1024 * 1024
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

// CursorBillingDatasourceRequest is the Terraform input used to create/validate a Cursor billing datasource.
type CursorBillingDatasourceRequest struct {
	Name        string
	AdminAPIKey string
	StartDate   *string
	EndDate     *string
}

// CursorBillingDatasource is the normalized datasource payload returned by the Costory API.
type CursorBillingDatasource struct {
	ID         string
	Type       string
	Status     *string
	Name       string
	BQTableURI string
	StartDate  *string
	EndDate    *string
}

// AnthropicBillingDatasourceRequest is the Terraform input used to create/validate an Anthropic billing datasource.
type AnthropicBillingDatasourceRequest struct {
	Name        string
	AdminAPIKey string
	StartDate   *string
	EndDate     *string
}

// AnthropicBillingDatasource is the normalized datasource payload returned by the Costory API.
type AnthropicBillingDatasource struct {
	ID         string
	Type       string
	Status     *string
	Name       string
	BQTableURI string
	StartDate  *string
	EndDate    *string
}

// ElasticCloudBillingDatasourceRequest is the Terraform input used to create/validate an Elastic Cloud billing datasource.
type ElasticCloudBillingDatasourceRequest struct {
	Name           string
	APIKey         string
	OrganizationID string
	StartDate      *string
	EndDate        *string
}

// ElasticCloudBillingDatasource is the normalized datasource payload returned by the Costory API.
type ElasticCloudBillingDatasource struct {
	ID             string
	Type           string
	Status         *string
	Name           string
	OrganizationID string
	BQTableURI     string
	StartDate      *string
	EndDate        *string
}

// AzureBillingDatasourceRequest is the Terraform input used to create/validate an Azure billing datasource.
type AzureBillingDatasourceRequest struct {
	Name               string
	SASURL             string
	StorageAccountName string
	ContainerName      string
	ActualsPath        string
	AmortizedPath      string
}

// AzureBillingDatasource is the normalized datasource payload returned by the Costory API.
type AzureBillingDatasource struct {
	ID                 string
	Type               string
	Status             *string
	Name               string
	StorageAccountName string
	ContainerName      string
	ActualsPath        string
	AmortizedPath      string
}

// TeamCreateRequest is the Terraform input used to create a team.
type TeamCreateRequest struct {
	Name        string
	Description *string
	Visibility  *string
}

// TeamUpdateRequest is the Terraform input used to update a team.
type TeamUpdateRequest struct {
	Name        *string
	Description *string
	Visibility  *string
}

// Team is the normalized team payload returned by the Costory API.
type Team struct {
	ID          string
	Name        string
	Description string
	Visibility  string
	CreatedAt   string
	UpdatedAt   string
}

// TeamMemberRequest is the Terraform input used to add a team member.
type TeamMemberRequest struct {
	UserID *string
	Email  *string
	Role   *string
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

type externalBillingDatasourceAPIRequest struct {
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	AdminAPIKey string  `json:"adminApiKey"`
	StartDate   *string `json:"startDate,omitempty"`
	EndDate     *string `json:"endDate,omitempty"`
}

type externalBillingDatasourceAPIResponse struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Status     *string `json:"status"`
	Name       string  `json:"name"`
	BQTableURI string  `json:"bqTableUri"`
	StartDate  *string `json:"startDate"`
	EndDate    *string `json:"endDate"`
}

type elasticCloudBillingDatasourceAPIRequest struct {
	Type           string  `json:"type"`
	Name           string  `json:"name"`
	APIKey         string  `json:"apiKey"`
	OrganizationID string  `json:"organizationId"`
	StartDate      *string `json:"startDate,omitempty"`
	EndDate        *string `json:"endDate,omitempty"`
}

type elasticCloudBillingDatasourceAPIResponse struct {
	ID             string  `json:"id"`
	Type           string  `json:"type"`
	Status         *string `json:"status"`
	Name           string  `json:"name"`
	OrganizationID string  `json:"organizationId"`
	BQTableURI     string  `json:"bqTableUri"`
	StartDate      *string `json:"startDate"`
	EndDate        *string `json:"endDate"`
}

type azureBillingDatasourceAPIRequest struct {
	Type               string `json:"type"`
	Name               string `json:"name"`
	SASURL             string `json:"sasUrl"`
	StorageAccountName string `json:"storageAccountName"`
	ContainerName      string `json:"containerName"`
	ActualsPath        string `json:"actualsPath"`
	AmortizedPath      string `json:"amortizedPath"`
}

type azureBillingDatasourceAPIResponse struct {
	ID                 string  `json:"id"`
	Type               string  `json:"type"`
	Status             *string `json:"status"`
	Name               string  `json:"name"`
	StorageAccountName string  `json:"storageAccountName"`
	ContainerName      string  `json:"containerName"`
	ActualsPath        string  `json:"actualsPath"`
	AmortizedPath      string  `json:"amortizedPath"`
}

type teamCreateAPIRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Visibility  *string `json:"visibility,omitempty"`
}

type teamUpdateAPIRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Visibility  *string `json:"visibility,omitempty"`
}

type teamMemberAPIRequest struct {
	UserID *string `json:"userId,omitempty"`
	Email  *string `json:"email,omitempty"`
	Role   *string `json:"role,omitempty"`
}

type teamAPIResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// MetricsDefinition is a single metric definition for an AwsS3V2 metrics datasource.
type MetricsDefinition struct {
	MetricName  string
	GapFilling  string
	Aggregation string
	ValueColumn string
	DateColumn  string
	Dimensions  []string
	Unit        string
}

// MetricsDatasourceRequest is the Terraform input used to create/validate a metrics datasource.
type MetricsDatasourceRequest struct {
	Name               string
	Type               string
	BucketName         string
	Prefix             string
	RoleARN            string
	MetricsDefinitions []MetricsDefinition
}

// MetricsDatasource is the normalized metrics datasource payload returned by the Costory API.
type MetricsDatasource struct {
	ID                string
	Type              string
	Status            *string
	Name              string
	BucketName        string
	Prefix            string
	RoleARN           string
	MetricsDefinition []MetricsDefinition
}

type metricsDefinitionAPI struct {
	MetricName  string   `json:"metricName"`
	GapFilling  string   `json:"gapFilling"`
	Aggregation string   `json:"aggregation"`
	ValueColumn string   `json:"valueColumn"`
	DateColumn  string   `json:"dateColumn"`
	Dimensions  []string `json:"dimensions,omitempty"`
	Unit        string   `json:"unit,omitempty"`
}

type metricsDatasourceAPIRequest struct {
	Type              string                 `json:"type"`
	Name              string                 `json:"name"`
	BucketName        string                 `json:"bucketName"`
	Prefix            string                 `json:"prefix"`
	RoleARN           string                 `json:"roleArn"`
	MetricsDefinition []metricsDefinitionAPI `json:"metricsDefinition"`
}

type metricsDatasourceAPIResponse struct {
	ID                string                 `json:"id"`
	Type              string                 `json:"type"`
	Status            *string                `json:"status"`
	Name              string                 `json:"name"`
	BucketName        string                 `json:"bucketName"`
	Prefix            string                 `json:"prefix"`
	RoleARN           string                 `json:"roleArn"`
	MetricsDefinition []metricsDefinitionAPI `json:"metricsDefinition"`
}

type metricsDatasourcePatchAPIRequest struct {
	MetricsDefinition []metricsDefinitionAPI `json:"metricsDefinition"`
}

type metricsDatasourceValidateAPIResponse struct {
	IsSuccess bool     `json:"isSuccess"`
	Errors    []string `json:"errors"`
}

type successResponse struct {
	Success bool `json:"success"`
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

// ValidateCursorBillingDatasource validates a Cursor billing datasource before creation.
func (c *Client) ValidateCursorBillingDatasource(ctx context.Context, req CursorBillingDatasourceRequest) error {
	body, statusCode, err := doEndpoint(ctx, c, endpointValidateCursorBillingDatasource, req.toAPIRequest())
	if err != nil {
		return err
	}

	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return nil
	}

	return unexpectedStatusError(statusCode, body)
}

// CreateCursorBillingDatasource creates a Cursor billing datasource and returns its API representation.
func (c *Client) CreateCursorBillingDatasource(ctx context.Context, req CursorBillingDatasourceRequest) (*CursorBillingDatasource, error) {
	body, statusCode, err := doEndpoint(ctx, c, endpointCreateCursorBillingDatasource, req.toAPIRequest())
	if err != nil {
		return nil, err
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out externalBillingDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toCursorBillingDatasource()
	if normalized.ID == "" {
		return nil, errors.New("create response did not include datasource id")
	}

	return normalized, nil
}

// GetCursorBillingDatasource gets a Cursor billing datasource by ID.
func (c *Client) GetCursorBillingDatasource(ctx context.Context, datasourceID string) (*CursorBillingDatasource, error) {
	routeParams := billingDatasourceByIDRouteParams{ID: datasourceID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointGetCursorBillingDatasourceByID, routeParams, noRequest{})
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if statusCode != http.StatusOK {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out externalBillingDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toCursorBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}

	return normalized, nil
}

// ValidateAnthropicBillingDatasource validates an Anthropic billing datasource before creation.
func (c *Client) ValidateAnthropicBillingDatasource(ctx context.Context, req AnthropicBillingDatasourceRequest) error {
	body, statusCode, err := doEndpoint(ctx, c, endpointValidateAnthropicBillingDatasource, req.toAPIRequest())
	if err != nil {
		return err
	}

	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return nil
	}

	return unexpectedStatusError(statusCode, body)
}

// CreateAnthropicBillingDatasource creates an Anthropic billing datasource and returns its API representation.
func (c *Client) CreateAnthropicBillingDatasource(ctx context.Context, req AnthropicBillingDatasourceRequest) (*AnthropicBillingDatasource, error) {
	body, statusCode, err := doEndpoint(ctx, c, endpointCreateAnthropicBillingDatasource, req.toAPIRequest())
	if err != nil {
		return nil, err
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out externalBillingDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toAnthropicBillingDatasource()
	if normalized.ID == "" {
		return nil, errors.New("create response did not include datasource id")
	}

	return normalized, nil
}

// GetAnthropicBillingDatasource gets an Anthropic billing datasource by ID.
func (c *Client) GetAnthropicBillingDatasource(ctx context.Context, datasourceID string) (*AnthropicBillingDatasource, error) {
	routeParams := billingDatasourceByIDRouteParams{ID: datasourceID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointGetAnthropicBillingDatasourceByID, routeParams, noRequest{})
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if statusCode != http.StatusOK {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out externalBillingDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toAnthropicBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}

	return normalized, nil
}

// ValidateElasticCloudBillingDatasource validates an Elastic Cloud billing datasource before creation.
func (c *Client) ValidateElasticCloudBillingDatasource(ctx context.Context, req ElasticCloudBillingDatasourceRequest) error {
	body, statusCode, err := doEndpoint(ctx, c, endpointValidateElasticCloudBillingDatasource, req.toAPIRequest())
	if err != nil {
		return err
	}

	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return nil
	}

	return unexpectedStatusError(statusCode, body)
}

// CreateElasticCloudBillingDatasource creates an Elastic Cloud billing datasource and returns its API representation.
func (c *Client) CreateElasticCloudBillingDatasource(ctx context.Context, req ElasticCloudBillingDatasourceRequest) (*ElasticCloudBillingDatasource, error) {
	body, statusCode, err := doEndpoint(ctx, c, endpointCreateElasticCloudBillingDatasource, req.toAPIRequest())
	if err != nil {
		return nil, err
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out elasticCloudBillingDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toElasticCloudBillingDatasource()
	if normalized.ID == "" {
		return nil, errors.New("create response did not include datasource id")
	}

	return normalized, nil
}

// GetElasticCloudBillingDatasource gets an Elastic Cloud billing datasource by ID.
func (c *Client) GetElasticCloudBillingDatasource(ctx context.Context, datasourceID string) (*ElasticCloudBillingDatasource, error) {
	routeParams := billingDatasourceByIDRouteParams{ID: datasourceID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointGetElasticCloudBillingDatasourceByID, routeParams, noRequest{})
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if statusCode != http.StatusOK {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out elasticCloudBillingDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toElasticCloudBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}

	return normalized, nil
}

// ValidateAzureBillingDatasource validates an Azure billing datasource before creation.
func (c *Client) ValidateAzureBillingDatasource(ctx context.Context, req AzureBillingDatasourceRequest) error {
	body, statusCode, err := doEndpoint(ctx, c, endpointValidateAzureBillingDatasource, req.toAPIRequest())
	if err != nil {
		return err
	}

	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return nil
	}

	return unexpectedStatusError(statusCode, body)
}

// CreateAzureBillingDatasource creates an Azure billing datasource and returns its API representation.
func (c *Client) CreateAzureBillingDatasource(ctx context.Context, req AzureBillingDatasourceRequest) (*AzureBillingDatasource, error) {
	body, statusCode, err := doEndpoint(ctx, c, endpointCreateAzureBillingDatasource, req.toAPIRequest())
	if err != nil {
		return nil, err
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out azureBillingDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toAzureBillingDatasource()
	if normalized.ID == "" {
		return nil, errors.New("create response did not include datasource id")
	}

	return normalized, nil
}

// GetAzureBillingDatasource gets an Azure billing datasource by ID.
func (c *Client) GetAzureBillingDatasource(ctx context.Context, datasourceID string) (*AzureBillingDatasource, error) {
	routeParams := billingDatasourceByIDRouteParams{ID: datasourceID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointGetAzureBillingDatasourceByID, routeParams, noRequest{})
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if statusCode != http.StatusOK {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out azureBillingDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toAzureBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}

	return normalized, nil
}

// CreateTeam creates a team and returns its API representation.
func (c *Client) CreateTeam(ctx context.Context, req TeamCreateRequest) (*Team, error) {
	body, statusCode, err := doEndpoint(ctx, c, endpointCreateTeam, req.toAPIRequest())
	if err != nil {
		return nil, err
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out teamAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toTeam()
	if normalized.ID == "" {
		return nil, errors.New("create response did not include team id")
	}

	return normalized, nil
}

// GetTeam gets a team by ID.
func (c *Client) GetTeam(ctx context.Context, teamID string) (*Team, error) {
	routeParams := teamByIDRouteParams{ID: teamID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointGetTeamByID, routeParams, noRequest{})
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if statusCode != http.StatusOK {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out teamAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toTeam()
	if normalized.ID == "" {
		normalized.ID = teamID
	}

	return normalized, nil
}

// UpdateTeam updates a team and returns its API representation.
func (c *Client) UpdateTeam(ctx context.Context, teamID string, req TeamUpdateRequest) (*Team, error) {
	routeParams := teamByIDRouteParams{ID: teamID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointPatchTeamByID, routeParams, req.toAPIRequest())
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out teamAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toTeam()
	if normalized.ID == "" {
		normalized.ID = teamID
	}

	return normalized, nil
}

// DeleteTeam archives a team by ID.
func (c *Client) DeleteTeam(ctx context.Context, teamID string) error {
	routeParams := teamByIDRouteParams{ID: teamID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointDeleteTeamByID, routeParams, noRequest{})
	if err != nil {
		return err
	}

	if statusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return nil
	}

	return unexpectedStatusError(statusCode, body)
}

// AddTeamMember adds a member to the team.
func (c *Client) AddTeamMember(ctx context.Context, teamID string, req TeamMemberRequest) error {
	routeParams := teamByIDRouteParams{ID: teamID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointAddTeamMember, routeParams, req.toAPIRequest())
	if err != nil {
		return err
	}

	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return nil
	}

	return unexpectedStatusError(statusCode, body)
}

// RemoveTeamMember removes a member from the team by user ID.
func (c *Client) RemoveTeamMember(ctx context.Context, teamID, userID string) error {
	routeParams := teamMemberRouteParams{TeamID: teamID, UserID: userID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointRemoveTeamMember, routeParams, noRequest{})
	if err != nil {
		return err
	}

	if statusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return nil
	}

	return unexpectedStatusError(statusCode, body)
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

// ValidateMetricsDatasource validates a metrics datasource before create/update.
// If the API returns isSuccess=false, returns an error with the errors[] joined.
func (c *Client) ValidateMetricsDatasource(ctx context.Context, req MetricsDatasourceRequest) error {
	body, statusCode, err := doEndpoint(ctx, c, endpointValidateMetricsDatasource, req.toAPIRequest())
	if err != nil {
		return err
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return unexpectedStatusError(statusCode, body)
	}

	var out metricsDatasourceValidateAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("decode validation response: %w", err)
	}

	if out.IsSuccess {
		return nil
	}

	if len(out.Errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(out.Errors, "; "))
	}

	return fmt.Errorf("validation failed")
}

// CreateMetricsDatasource creates a metrics datasource and returns its API representation.
func (c *Client) CreateMetricsDatasource(ctx context.Context, req MetricsDatasourceRequest) (*MetricsDatasource, error) {
	body, statusCode, err := doEndpoint(ctx, c, endpointCreateMetricsDatasource, req.toAPIRequest())
	if err != nil {
		return nil, err
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out metricsDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toMetricsDatasource()
	if normalized.ID == "" {
		return nil, errors.New("create response did not include datasource id")
	}

	return normalized, nil
}

// GetMetricsDatasource gets a metrics datasource by ID.
func (c *Client) GetMetricsDatasource(ctx context.Context, datasourceID string) (*MetricsDatasource, error) {
	routeParams := metricsDatasourceByIDRouteParams{ID: datasourceID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointGetMetricsDatasourceByID, routeParams, noRequest{})
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if statusCode != http.StatusOK {
		return nil, unexpectedStatusError(statusCode, body)
	}

	var out metricsDatasourceAPIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	normalized := out.toMetricsDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}

	return normalized, nil
}

// UpdateMetricsDatasource updates a metrics datasource's metrics definition via PATCH.
func (c *Client) UpdateMetricsDatasource(ctx context.Context, datasourceID string, metricsDefinitions []MetricsDefinition) error {
	routeParams := metricsDatasourceByIDRouteParams{ID: datasourceID}
	patchReq := metricsDefinitionsToPatchRequest(metricsDefinitions)
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointPatchMetricsDatasourceByID, routeParams, patchReq)
	if err != nil {
		return err
	}

	if statusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return nil
	}

	return unexpectedStatusError(statusCode, body)
}

// DeleteMetricsDatasource deletes a metrics datasource by ID.
func (c *Client) DeleteMetricsDatasource(ctx context.Context, datasourceID string) error {
	routeParams := metricsDatasourceByIDRouteParams{ID: datasourceID}
	body, statusCode, err := doEndpointWithRouteParams(ctx, c, endpointDeleteMetricsDatasourceByID, routeParams, noRequest{})
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

func (r CursorBillingDatasourceRequest) toAPIRequest() externalBillingDatasourceAPIRequest {
	return externalBillingDatasourceAPIRequest{
		Type:        billingDatasourceTypeCursor,
		Name:        r.Name,
		AdminAPIKey: r.AdminAPIKey,
		StartDate:   r.StartDate,
		EndDate:     r.EndDate,
	}
}

func (r AnthropicBillingDatasourceRequest) toAPIRequest() externalBillingDatasourceAPIRequest {
	return externalBillingDatasourceAPIRequest{
		Type:        billingDatasourceTypeAnthropic,
		Name:        r.Name,
		AdminAPIKey: r.AdminAPIKey,
		StartDate:   r.StartDate,
		EndDate:     r.EndDate,
	}
}

func (r ElasticCloudBillingDatasourceRequest) toAPIRequest() elasticCloudBillingDatasourceAPIRequest {
	return elasticCloudBillingDatasourceAPIRequest{
		Type:           billingDatasourceTypeElasticCloud,
		Name:           r.Name,
		APIKey:         r.APIKey,
		OrganizationID: r.OrganizationID,
		StartDate:      r.StartDate,
		EndDate:        r.EndDate,
	}
}

func (r AzureBillingDatasourceRequest) toAPIRequest() azureBillingDatasourceAPIRequest {
	return azureBillingDatasourceAPIRequest{
		Type:               billingDatasourceTypeAzure,
		Name:               r.Name,
		SASURL:             r.SASURL,
		StorageAccountName: r.StorageAccountName,
		ContainerName:      r.ContainerName,
		ActualsPath:        r.ActualsPath,
		AmortizedPath:      r.AmortizedPath,
	}
}

func (r TeamCreateRequest) toAPIRequest() teamCreateAPIRequest {
	return teamCreateAPIRequest(r)
}

func (r TeamUpdateRequest) toAPIRequest() teamUpdateAPIRequest {
	return teamUpdateAPIRequest(r)
}

func (r TeamMemberRequest) toAPIRequest() teamMemberAPIRequest {
	return teamMemberAPIRequest(r)
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

func (r externalBillingDatasourceAPIResponse) toCursorBillingDatasource() *CursorBillingDatasource {
	return &CursorBillingDatasource{
		ID:         r.ID,
		Type:       r.Type,
		Status:     r.Status,
		Name:       r.Name,
		BQTableURI: r.BQTableURI,
		StartDate:  r.StartDate,
		EndDate:    r.EndDate,
	}
}

func (r externalBillingDatasourceAPIResponse) toAnthropicBillingDatasource() *AnthropicBillingDatasource {
	return &AnthropicBillingDatasource{
		ID:         r.ID,
		Type:       r.Type,
		Status:     r.Status,
		Name:       r.Name,
		BQTableURI: r.BQTableURI,
		StartDate:  r.StartDate,
		EndDate:    r.EndDate,
	}
}

func (r elasticCloudBillingDatasourceAPIResponse) toElasticCloudBillingDatasource() *ElasticCloudBillingDatasource {
	return &ElasticCloudBillingDatasource{
		ID:             r.ID,
		Type:           r.Type,
		Status:         r.Status,
		Name:           r.Name,
		OrganizationID: r.OrganizationID,
		BQTableURI:     r.BQTableURI,
		StartDate:      r.StartDate,
		EndDate:        r.EndDate,
	}
}

func (r azureBillingDatasourceAPIResponse) toAzureBillingDatasource() *AzureBillingDatasource {
	return &AzureBillingDatasource{
		ID:                 r.ID,
		Type:               r.Type,
		Status:             r.Status,
		Name:               r.Name,
		StorageAccountName: r.StorageAccountName,
		ContainerName:      r.ContainerName,
		ActualsPath:        r.ActualsPath,
		AmortizedPath:      r.AmortizedPath,
	}
}

func (r teamAPIResponse) toTeam() *Team {
	return &Team{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Visibility:  r.Visibility,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func (r MetricsDatasourceRequest) toAPIRequest() metricsDatasourceAPIRequest {
	defs := make([]metricsDefinitionAPI, len(r.MetricsDefinitions))
	for i, d := range r.MetricsDefinitions {
		defs[i] = metricsDefinitionAPI(d)
	}
	return metricsDatasourceAPIRequest{
		Type:              r.Type,
		Name:              r.Name,
		BucketName:        r.BucketName,
		Prefix:            r.Prefix,
		RoleARN:           r.RoleARN,
		MetricsDefinition: defs,
	}
}

func metricsDefinitionsToPatchRequest(defs []MetricsDefinition) metricsDatasourcePatchAPIRequest {
	apiDefs := make([]metricsDefinitionAPI, len(defs))
	for i, d := range defs {
		apiDefs[i] = metricsDefinitionAPI(d)
	}
	return metricsDatasourcePatchAPIRequest{MetricsDefinition: apiDefs}
}

func (r metricsDatasourceAPIResponse) toMetricsDatasource() *MetricsDatasource {
	defs := make([]MetricsDefinition, len(r.MetricsDefinition))
	for i, d := range r.MetricsDefinition {
		defs[i] = MetricsDefinition{
			MetricName:  d.MetricName,
			GapFilling:  d.GapFilling,
			Aggregation: d.Aggregation,
			ValueColumn: d.ValueColumn,
			DateColumn:  d.DateColumn,
			Dimensions:  append([]string(nil), d.Dimensions...),
			Unit:        d.Unit,
		}
	}
	return &MetricsDatasource{
		ID:                r.ID,
		Type:              r.Type,
		Status:            r.Status,
		Name:              r.Name,
		BucketName:        r.BucketName,
		Prefix:            r.Prefix,
		RoleARN:           r.RoleARN,
		MetricsDefinition: defs,
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
