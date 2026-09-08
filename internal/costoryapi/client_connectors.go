package costoryapi

import (
	"context"
	"errors"
)

const (
	billingDatasourceTypeAnthropicClaudeAi = "AnthropicClaudeAi"
	billingDatasourceTypeClickHouseCloud   = "ClickHouseCloud"
	billingDatasourceTypeCloudflare        = "Cloudflare"
	billingDatasourceTypeScaleway          = "Scaleway"
	billingDatasourceTypeOpenAI            = "OpenAI"
	billingDatasourceTypeCustomBigQuery    = "CustomBigQuery"
	billingDatasourceTypeConfluent         = "CONFLUENT"
	billingDatasourceTypeDatadog           = "Datadog"
	billingDatasourceTypeAiven             = "Aiven"
	billingDatasourceTypeGCPCUDExport      = "GCP_CUD_EXPORT"
	billingDatasourceTypeSnowflake         = "SNOWFLAKE"
)

// OpenAIBillingDatasourceRequest is the Terraform input used to create/validate an OpenAI billing datasource.
type OpenAIBillingDatasourceRequest struct {
	Name        string
	AdminAPIKey string
	StartDate   *string
}

// OpenAIBillingDatasource is the normalized datasource payload returned by the Costory API.
type OpenAIBillingDatasource struct {
	ID         string
	Type       string
	Status     *string
	Name       string
	BQTableURI string
	StartDate  *string
}

// AnthropicClaudeAiBillingDatasourceRequest is the Terraform input used to create/validate an Anthropic Claude AI billing datasource.
type AnthropicClaudeAiBillingDatasourceRequest struct {
	Name            string
	AnalyticsAPIKey string
	AccountName     string
	StartDate       *string
}

// AnthropicClaudeAiBillingDatasource is the normalized datasource payload returned by the Costory API.
type AnthropicClaudeAiBillingDatasource struct {
	ID         string
	Type       string
	Status     *string
	Name       string
	BQTableURI string
	StartDate  *string
}

// ClickHouseCloudBillingDatasourceRequest is the Terraform input used to create/validate a ClickHouse Cloud billing datasource.
type ClickHouseCloudBillingDatasourceRequest struct {
	Name           string
	KeyID          string
	KeySecret      string
	OrganizationID string
	StartDate      *string
}

// ClickHouseCloudBillingDatasource is the normalized datasource payload returned by the Costory API.
type ClickHouseCloudBillingDatasource struct {
	ID         string
	Type       string
	Status     *string
	Name       string
	BQTableURI string
	StartDate  *string
}

// CloudflareBillingDatasourceRequest is the Terraform input used to create/validate a Cloudflare billing datasource.
type CloudflareBillingDatasourceRequest struct {
	Name      string
	APIToken  string
	AccountID string
	StartDate *string
}

// CloudflareBillingDatasource is the normalized datasource payload returned by the Costory API.
type CloudflareBillingDatasource struct {
	ID         string
	Type       string
	Status     *string
	Name       string
	BQTableURI string
	StartDate  *string
}

// ScalewayBillingDatasourceRequest is the Terraform input used to create/validate a Scaleway billing datasource.
type ScalewayBillingDatasourceRequest struct {
	Name           string
	SecretKey      string
	OrganizationID string
	AccessKey      *string
	StartDate      *string
}

// ScalewayBillingDatasource is the normalized datasource payload returned by the Costory API.
type ScalewayBillingDatasource struct {
	ID         string
	Type       string
	Status     *string
	Name       string
	BQTableURI string
	StartDate  *string
}

// ConfluentBillingDatasourceRequest is the Terraform input used to create/validate a Confluent billing datasource.
type ConfluentBillingDatasourceRequest struct {
	Name      string
	APIKey    string
	APISecret string
}

// ConfluentBillingDatasource is the normalized datasource payload returned by the Costory API.
type ConfluentBillingDatasource struct {
	ID         string
	Type       string
	Status     *string
	Name       string
	BQTableURI string
}

// DatadogBillingDatasourceRequest is the Terraform input used to create/validate a Datadog billing datasource.
type DatadogBillingDatasourceRequest struct {
	Name           string
	APIKey         string
	ApplicationKey string
	Region         string
}

// DatadogBillingDatasource is the normalized datasource payload returned by the Costory API.
type DatadogBillingDatasource struct {
	ID            string
	Type          string
	Status        *string
	Name          string
	IntegrationID string
	BQTableURI    string
}

// AivenBillingDatasourceRequest is the Terraform input used to create/validate an Aiven billing datasource.
type AivenBillingDatasourceRequest struct {
	Name           string
	APISecret      string
	OrganizationID string
}

// AivenBillingDatasource is the normalized datasource payload returned by the Costory API.
type AivenBillingDatasource struct {
	ID             string
	Type           string
	Status         *string
	Name           string
	OrganizationID string
	BQTableURI     string
}

// GCPCUDExportBillingDatasourceRequest is the Terraform input used to create/validate a GCP CUD export billing datasource.
type GCPCUDExportBillingDatasourceRequest struct {
	Name        string
	BQTablePath string
}

// GCPCUDExportBillingDatasource is the normalized datasource payload returned by the Costory API.
type GCPCUDExportBillingDatasource struct {
	ID          string
	Type        string
	Status      *string
	Name        string
	BQTablePath string
}

// SnowflakeBillingDatasourceRequest is the Terraform input used to create/validate a Snowflake billing datasource.
type SnowflakeBillingDatasourceRequest struct {
	Name          string
	IntegrationID string
}

// SnowflakeBillingDatasource is the normalized datasource payload returned by the Costory API.
type SnowflakeBillingDatasource struct {
	ID            string
	Type          string
	Status        *string
	Name          string
	IntegrationID string
	BQTableURIs   []string
}

// CustomBigQueryBillingDatasourceRequest is the Terraform input used to create/validate a custom BigQuery billing datasource.
type CustomBigQueryBillingDatasourceRequest struct {
	Name             string
	BQTablePath      string
	ProviderName     string
	BillingAccountID *string
	StartDate        *string
	EndDate          *string
	Mapping          map[string]string
}

// CustomBigQueryBillingDatasource is the normalized datasource payload returned by the Costory API.
type CustomBigQueryBillingDatasource struct {
	ID               string
	Type             string
	Status           *string
	Name             string
	BQTablePath      string
	ProviderName     string
	BillingAccountID *string
	StartDate        *string
	EndDate          *string
	Mapping          map[string]string
}

type openaiBillingDatasourceAPIRequest struct {
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	AdminAPIKey string  `json:"adminApiKey"`
	StartDate   *string `json:"startDate,omitempty"`
}

type openaiBillingDatasourceAPIResponse struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Status     *string `json:"status"`
	Name       string  `json:"name"`
	BQTableURI string  `json:"bqTableUri"`
	StartDate  *string `json:"startDate"`
}

type anthropicClaudeAiBillingDatasourceAPIRequest struct {
	Type            string  `json:"type"`
	Name            string  `json:"name"`
	AnalyticsAPIKey string  `json:"analyticsApiKey"`
	AccountName     string  `json:"accountName"`
	StartDate       *string `json:"startDate,omitempty"`
}

type anthropicClaudeAiBillingDatasourceAPIResponse struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Status     *string `json:"status"`
	Name       string  `json:"name"`
	BQTableURI string  `json:"bqTableUri"`
	StartDate  *string `json:"startDate"`
}

type clickHouseCloudBillingDatasourceAPIRequest struct {
	Type           string  `json:"type"`
	Name           string  `json:"name"`
	KeyID          string  `json:"keyId"`
	KeySecret      string  `json:"keySecret"`
	OrganizationID string  `json:"organizationId"`
	StartDate      *string `json:"startDate,omitempty"`
}

type clickHouseCloudBillingDatasourceAPIResponse struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Status     *string `json:"status"`
	Name       string  `json:"name"`
	BQTableURI string  `json:"bqTableUri"`
	StartDate  *string `json:"startDate"`
}

type cloudflareBillingDatasourceAPIRequest struct {
	Type      string  `json:"type"`
	Name      string  `json:"name"`
	APIToken  string  `json:"apiToken"`
	AccountID string  `json:"accountId"`
	StartDate *string `json:"startDate,omitempty"`
}

type cloudflareBillingDatasourceAPIResponse struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Status     *string `json:"status"`
	Name       string  `json:"name"`
	BQTableURI string  `json:"bqTableUri"`
	StartDate  *string `json:"startDate"`
}

type scalewayBillingDatasourceAPIRequest struct {
	Type           string  `json:"type"`
	Name           string  `json:"name"`
	SecretKey      string  `json:"secretKey"`
	OrganizationID string  `json:"organizationId"`
	AccessKey      *string `json:"accessKey,omitempty"`
	StartDate      *string `json:"startDate,omitempty"`
}

type scalewayBillingDatasourceAPIResponse struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Status     *string `json:"status"`
	Name       string  `json:"name"`
	BQTableURI string  `json:"bqTableUri"`
	StartDate  *string `json:"startDate"`
}

type confluentBillingDatasourceAPIRequest struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	APIKey    string `json:"apiKey"`
	APISecret string `json:"apiSecret"`
}

type confluentBillingDatasourceAPIResponse struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Status     *string `json:"status"`
	Name       string  `json:"name"`
	BQTableURI string  `json:"bqTableUri"`
}

type datadogBillingDatasourceAPIRequest struct {
	Type           string `json:"type"`
	Name           string `json:"name"`
	APIKey         string `json:"apiKey"`
	ApplicationKey string `json:"applicationKey"`
	Region         string `json:"region"`
}

type datadogBillingDatasourceAPIResponse struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	Status        *string `json:"status"`
	Name          string  `json:"name"`
	IntegrationID string  `json:"integrationId"`
	BQTableURI    string  `json:"bqTableUri"`
}

type aivenBillingDatasourceAPIRequest struct {
	Type           string `json:"type"`
	Name           string `json:"name"`
	APISecret      string `json:"apiSecret"`
	OrganizationID string `json:"organizationId"`
}

type aivenBillingDatasourceAPIResponse struct {
	ID             string  `json:"id"`
	Type           string  `json:"type"`
	Status         *string `json:"status"`
	Name           string  `json:"name"`
	OrganizationID string  `json:"organizationId"`
	BQTableURI     string  `json:"bqTableUri"`
}

type gcpCUDExportBillingDatasourceAPIRequest struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	BQTablePath string `json:"bqTablePath"`
}

type gcpCUDExportBillingDatasourceAPIResponse struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Status      *string `json:"status"`
	Name        string  `json:"name"`
	BQTablePath string  `json:"bqTablePath"`
}

type snowflakeBillingDatasourceAPIRequest struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	IntegrationID string `json:"integrationId"`
}

type snowflakeBillingDatasourceAPIResponse struct {
	ID            string   `json:"id"`
	Type          string   `json:"type"`
	Status        *string  `json:"status"`
	Name          string   `json:"name"`
	IntegrationID string   `json:"integrationId"`
	BQTableURIs   []string `json:"bqTableUris"`
}

type customBigQueryBillingDatasourceAPIRequest struct {
	Type             string            `json:"type"`
	Name             string            `json:"name"`
	BQTablePath      string            `json:"bqTablePath"`
	ProviderName     string            `json:"providerName"`
	BillingAccountID *string           `json:"billingAccountId,omitempty"`
	StartDate        *string           `json:"startDate,omitempty"`
	EndDate          *string           `json:"endDate,omitempty"`
	Mapping          map[string]string `json:"mapping"`
}

type customBigQueryBillingDatasourceAPIResponse struct {
	ID               string            `json:"id"`
	Type             string            `json:"type"`
	Status           *string           `json:"status"`
	Name             string            `json:"name"`
	BQTablePath      string            `json:"bqTablePath"`
	ProviderName     string            `json:"providerName"`
	BillingAccountID *string           `json:"billingAccountId"`
	StartDate        *string           `json:"startDate"`
	EndDate          *string           `json:"endDate"`
	Mapping          map[string]string `json:"mapping"`
}

func (r OpenAIBillingDatasourceRequest) toAPIRequest() openaiBillingDatasourceAPIRequest {
	return openaiBillingDatasourceAPIRequest{
		Type:        billingDatasourceTypeOpenAI,
		Name:        r.Name,
		AdminAPIKey: r.AdminAPIKey,
		StartDate:   r.StartDate,
	}
}

func (r AnthropicClaudeAiBillingDatasourceRequest) toAPIRequest() anthropicClaudeAiBillingDatasourceAPIRequest {
	return anthropicClaudeAiBillingDatasourceAPIRequest{
		Type:            billingDatasourceTypeAnthropicClaudeAi,
		Name:            r.Name,
		AnalyticsAPIKey: r.AnalyticsAPIKey,
		AccountName:     r.AccountName,
		StartDate:       r.StartDate,
	}
}

func (r ClickHouseCloudBillingDatasourceRequest) toAPIRequest() clickHouseCloudBillingDatasourceAPIRequest {
	return clickHouseCloudBillingDatasourceAPIRequest{
		Type:           billingDatasourceTypeClickHouseCloud,
		Name:           r.Name,
		KeyID:          r.KeyID,
		KeySecret:      r.KeySecret,
		OrganizationID: r.OrganizationID,
		StartDate:      r.StartDate,
	}
}

func (r CloudflareBillingDatasourceRequest) toAPIRequest() cloudflareBillingDatasourceAPIRequest {
	return cloudflareBillingDatasourceAPIRequest{
		Type:      billingDatasourceTypeCloudflare,
		Name:      r.Name,
		APIToken:  r.APIToken,
		AccountID: r.AccountID,
		StartDate: r.StartDate,
	}
}

func (r ScalewayBillingDatasourceRequest) toAPIRequest() scalewayBillingDatasourceAPIRequest {
	return scalewayBillingDatasourceAPIRequest{
		Type:           billingDatasourceTypeScaleway,
		Name:           r.Name,
		SecretKey:      r.SecretKey,
		OrganizationID: r.OrganizationID,
		AccessKey:      r.AccessKey,
		StartDate:      r.StartDate,
	}
}

func (r ConfluentBillingDatasourceRequest) toAPIRequest() confluentBillingDatasourceAPIRequest {
	return confluentBillingDatasourceAPIRequest{
		Type:      billingDatasourceTypeConfluent,
		Name:      r.Name,
		APIKey:    r.APIKey,
		APISecret: r.APISecret,
	}
}

func (r DatadogBillingDatasourceRequest) toAPIRequest() datadogBillingDatasourceAPIRequest {
	return datadogBillingDatasourceAPIRequest{
		Type:           billingDatasourceTypeDatadog,
		Name:           r.Name,
		APIKey:         r.APIKey,
		ApplicationKey: r.ApplicationKey,
		Region:         r.Region,
	}
}

func (r AivenBillingDatasourceRequest) toAPIRequest() aivenBillingDatasourceAPIRequest {
	return aivenBillingDatasourceAPIRequest{
		Type:           billingDatasourceTypeAiven,
		Name:           r.Name,
		APISecret:      r.APISecret,
		OrganizationID: r.OrganizationID,
	}
}

func (r GCPCUDExportBillingDatasourceRequest) toAPIRequest() gcpCUDExportBillingDatasourceAPIRequest {
	return gcpCUDExportBillingDatasourceAPIRequest{
		Type:        billingDatasourceTypeGCPCUDExport,
		Name:        r.Name,
		BQTablePath: r.BQTablePath,
	}
}

func (r SnowflakeBillingDatasourceRequest) toAPIRequest() snowflakeBillingDatasourceAPIRequest {
	return snowflakeBillingDatasourceAPIRequest{
		Type:          billingDatasourceTypeSnowflake,
		Name:          r.Name,
		IntegrationID: r.IntegrationID,
	}
}

func (r CustomBigQueryBillingDatasourceRequest) toAPIRequest() customBigQueryBillingDatasourceAPIRequest {
	return customBigQueryBillingDatasourceAPIRequest{
		Type:             billingDatasourceTypeCustomBigQuery,
		Name:             r.Name,
		BQTablePath:      r.BQTablePath,
		ProviderName:     r.ProviderName,
		BillingAccountID: r.BillingAccountID,
		StartDate:        r.StartDate,
		EndDate:          r.EndDate,
		Mapping:          r.Mapping,
	}
}

func (r openaiBillingDatasourceAPIResponse) toOpenAIBillingDatasource() *OpenAIBillingDatasource {
	return &OpenAIBillingDatasource{
		ID:         r.ID,
		Type:       r.Type,
		Status:     r.Status,
		Name:       r.Name,
		BQTableURI: r.BQTableURI,
		StartDate:  r.StartDate,
	}
}

func (r anthropicClaudeAiBillingDatasourceAPIResponse) toAnthropicClaudeAiBillingDatasource() *AnthropicClaudeAiBillingDatasource {
	return &AnthropicClaudeAiBillingDatasource{
		ID:         r.ID,
		Type:       r.Type,
		Status:     r.Status,
		Name:       r.Name,
		BQTableURI: r.BQTableURI,
		StartDate:  r.StartDate,
	}
}

func (r clickHouseCloudBillingDatasourceAPIResponse) toClickHouseCloudBillingDatasource() *ClickHouseCloudBillingDatasource {
	return &ClickHouseCloudBillingDatasource{
		ID:         r.ID,
		Type:       r.Type,
		Status:     r.Status,
		Name:       r.Name,
		BQTableURI: r.BQTableURI,
		StartDate:  r.StartDate,
	}
}

func (r cloudflareBillingDatasourceAPIResponse) toCloudflareBillingDatasource() *CloudflareBillingDatasource {
	return &CloudflareBillingDatasource{
		ID:         r.ID,
		Type:       r.Type,
		Status:     r.Status,
		Name:       r.Name,
		BQTableURI: r.BQTableURI,
		StartDate:  r.StartDate,
	}
}

func (r scalewayBillingDatasourceAPIResponse) toScalewayBillingDatasource() *ScalewayBillingDatasource {
	return &ScalewayBillingDatasource{
		ID:         r.ID,
		Type:       r.Type,
		Status:     r.Status,
		Name:       r.Name,
		BQTableURI: r.BQTableURI,
		StartDate:  r.StartDate,
	}
}

func (r confluentBillingDatasourceAPIResponse) toConfluentBillingDatasource() *ConfluentBillingDatasource {
	return &ConfluentBillingDatasource{
		ID:         r.ID,
		Type:       r.Type,
		Status:     r.Status,
		Name:       r.Name,
		BQTableURI: r.BQTableURI,
	}
}

func (r datadogBillingDatasourceAPIResponse) toDatadogBillingDatasource() *DatadogBillingDatasource {
	return &DatadogBillingDatasource{
		ID:            r.ID,
		Type:          r.Type,
		Status:        r.Status,
		Name:          r.Name,
		IntegrationID: r.IntegrationID,
		BQTableURI:    r.BQTableURI,
	}
}

func (r aivenBillingDatasourceAPIResponse) toAivenBillingDatasource() *AivenBillingDatasource {
	return &AivenBillingDatasource{
		ID:             r.ID,
		Type:           r.Type,
		Status:         r.Status,
		Name:           r.Name,
		OrganizationID: r.OrganizationID,
		BQTableURI:     r.BQTableURI,
	}
}

func (r gcpCUDExportBillingDatasourceAPIResponse) toGCPCUDExportBillingDatasource() *GCPCUDExportBillingDatasource {
	return &GCPCUDExportBillingDatasource{
		ID:          r.ID,
		Type:        r.Type,
		Status:      r.Status,
		Name:        r.Name,
		BQTablePath: r.BQTablePath,
	}
}

func (r snowflakeBillingDatasourceAPIResponse) toSnowflakeBillingDatasource() *SnowflakeBillingDatasource {
	uris := r.BQTableURIs
	if uris == nil {
		uris = []string{}
	}
	return &SnowflakeBillingDatasource{
		ID:            r.ID,
		Type:          r.Type,
		Status:        r.Status,
		Name:          r.Name,
		IntegrationID: r.IntegrationID,
		BQTableURIs:   uris,
	}
}

func (r customBigQueryBillingDatasourceAPIResponse) toCustomBigQueryBillingDatasource() *CustomBigQueryBillingDatasource {
	mapping := r.Mapping
	if mapping == nil {
		mapping = map[string]string{}
	}
	return &CustomBigQueryBillingDatasource{
		ID:               r.ID,
		Type:             r.Type,
		Status:           r.Status,
		Name:             r.Name,
		BQTablePath:      r.BQTablePath,
		ProviderName:     r.ProviderName,
		BillingAccountID: r.BillingAccountID,
		StartDate:        r.StartDate,
		EndDate:          r.EndDate,
		Mapping:          mapping,
	}
}

func missingDatasourceIDError() error {
	return errors.New("create response did not include datasource id")
}

// ValidateOpenAIBillingDatasource validates an OpenAI billing datasource before creation.
func (c *Client) ValidateOpenAIBillingDatasource(ctx context.Context, req OpenAIBillingDatasourceRequest) error {
	return c.validateBillingDatasource(ctx, req.toAPIRequest())
}

// CreateOpenAIBillingDatasource creates an OpenAI billing datasource and returns its API representation.
func (c *Client) CreateOpenAIBillingDatasource(ctx context.Context, req OpenAIBillingDatasourceRequest) (*OpenAIBillingDatasource, error) {
	var out openaiBillingDatasourceAPIResponse
	if err := c.createBillingDatasource(ctx, req.toAPIRequest(), &out); err != nil {
		return nil, err
	}
	normalized := out.toOpenAIBillingDatasource()
	if normalized.ID == "" {
		return nil, missingDatasourceIDError()
	}
	return normalized, nil
}

// GetOpenAIBillingDatasource gets an OpenAI billing datasource by ID.
func (c *Client) GetOpenAIBillingDatasource(ctx context.Context, datasourceID string) (*OpenAIBillingDatasource, error) {
	var out openaiBillingDatasourceAPIResponse
	if err := c.getBillingDatasource(ctx, datasourceID, &out); err != nil {
		return nil, err
	}
	normalized := out.toOpenAIBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}
	return normalized, nil
}

// ValidateAnthropicClaudeAiBillingDatasource validates an Anthropic Claude AI billing datasource before creation.
func (c *Client) ValidateAnthropicClaudeAiBillingDatasource(ctx context.Context, req AnthropicClaudeAiBillingDatasourceRequest) error {
	return c.validateBillingDatasource(ctx, req.toAPIRequest())
}

// CreateAnthropicClaudeAiBillingDatasource creates an Anthropic Claude AI billing datasource and returns its API representation.
func (c *Client) CreateAnthropicClaudeAiBillingDatasource(ctx context.Context, req AnthropicClaudeAiBillingDatasourceRequest) (*AnthropicClaudeAiBillingDatasource, error) {
	var out anthropicClaudeAiBillingDatasourceAPIResponse
	if err := c.createBillingDatasource(ctx, req.toAPIRequest(), &out); err != nil {
		return nil, err
	}
	normalized := out.toAnthropicClaudeAiBillingDatasource()
	if normalized.ID == "" {
		return nil, missingDatasourceIDError()
	}
	return normalized, nil
}

// GetAnthropicClaudeAiBillingDatasource gets an Anthropic Claude AI billing datasource by ID.
func (c *Client) GetAnthropicClaudeAiBillingDatasource(ctx context.Context, datasourceID string) (*AnthropicClaudeAiBillingDatasource, error) {
	var out anthropicClaudeAiBillingDatasourceAPIResponse
	if err := c.getBillingDatasource(ctx, datasourceID, &out); err != nil {
		return nil, err
	}
	normalized := out.toAnthropicClaudeAiBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}
	return normalized, nil
}

// ValidateClickHouseCloudBillingDatasource validates a ClickHouse Cloud billing datasource before creation.
func (c *Client) ValidateClickHouseCloudBillingDatasource(ctx context.Context, req ClickHouseCloudBillingDatasourceRequest) error {
	return c.validateBillingDatasource(ctx, req.toAPIRequest())
}

// CreateClickHouseCloudBillingDatasource creates a ClickHouse Cloud billing datasource and returns its API representation.
func (c *Client) CreateClickHouseCloudBillingDatasource(ctx context.Context, req ClickHouseCloudBillingDatasourceRequest) (*ClickHouseCloudBillingDatasource, error) {
	var out clickHouseCloudBillingDatasourceAPIResponse
	if err := c.createBillingDatasource(ctx, req.toAPIRequest(), &out); err != nil {
		return nil, err
	}
	normalized := out.toClickHouseCloudBillingDatasource()
	if normalized.ID == "" {
		return nil, missingDatasourceIDError()
	}
	return normalized, nil
}

// GetClickHouseCloudBillingDatasource gets a ClickHouse Cloud billing datasource by ID.
func (c *Client) GetClickHouseCloudBillingDatasource(ctx context.Context, datasourceID string) (*ClickHouseCloudBillingDatasource, error) {
	var out clickHouseCloudBillingDatasourceAPIResponse
	if err := c.getBillingDatasource(ctx, datasourceID, &out); err != nil {
		return nil, err
	}
	normalized := out.toClickHouseCloudBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}
	return normalized, nil
}

// ValidateCloudflareBillingDatasource validates a Cloudflare billing datasource before creation.
func (c *Client) ValidateCloudflareBillingDatasource(ctx context.Context, req CloudflareBillingDatasourceRequest) error {
	return c.validateBillingDatasource(ctx, req.toAPIRequest())
}

// CreateCloudflareBillingDatasource creates a Cloudflare billing datasource and returns its API representation.
func (c *Client) CreateCloudflareBillingDatasource(ctx context.Context, req CloudflareBillingDatasourceRequest) (*CloudflareBillingDatasource, error) {
	var out cloudflareBillingDatasourceAPIResponse
	if err := c.createBillingDatasource(ctx, req.toAPIRequest(), &out); err != nil {
		return nil, err
	}
	normalized := out.toCloudflareBillingDatasource()
	if normalized.ID == "" {
		return nil, missingDatasourceIDError()
	}
	return normalized, nil
}

// GetCloudflareBillingDatasource gets a Cloudflare billing datasource by ID.
func (c *Client) GetCloudflareBillingDatasource(ctx context.Context, datasourceID string) (*CloudflareBillingDatasource, error) {
	var out cloudflareBillingDatasourceAPIResponse
	if err := c.getBillingDatasource(ctx, datasourceID, &out); err != nil {
		return nil, err
	}
	normalized := out.toCloudflareBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}
	return normalized, nil
}

// ValidateScalewayBillingDatasource validates a Scaleway billing datasource before creation.
func (c *Client) ValidateScalewayBillingDatasource(ctx context.Context, req ScalewayBillingDatasourceRequest) error {
	return c.validateBillingDatasource(ctx, req.toAPIRequest())
}

// CreateScalewayBillingDatasource creates a Scaleway billing datasource and returns its API representation.
func (c *Client) CreateScalewayBillingDatasource(ctx context.Context, req ScalewayBillingDatasourceRequest) (*ScalewayBillingDatasource, error) {
	var out scalewayBillingDatasourceAPIResponse
	if err := c.createBillingDatasource(ctx, req.toAPIRequest(), &out); err != nil {
		return nil, err
	}
	normalized := out.toScalewayBillingDatasource()
	if normalized.ID == "" {
		return nil, missingDatasourceIDError()
	}
	return normalized, nil
}

// GetScalewayBillingDatasource gets a Scaleway billing datasource by ID.
func (c *Client) GetScalewayBillingDatasource(ctx context.Context, datasourceID string) (*ScalewayBillingDatasource, error) {
	var out scalewayBillingDatasourceAPIResponse
	if err := c.getBillingDatasource(ctx, datasourceID, &out); err != nil {
		return nil, err
	}
	normalized := out.toScalewayBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}
	return normalized, nil
}

// ValidateConfluentBillingDatasource validates a Confluent billing datasource before creation.
func (c *Client) ValidateConfluentBillingDatasource(ctx context.Context, req ConfluentBillingDatasourceRequest) error {
	return c.validateBillingDatasource(ctx, req.toAPIRequest())
}

// CreateConfluentBillingDatasource creates a Confluent billing datasource and returns its API representation.
func (c *Client) CreateConfluentBillingDatasource(ctx context.Context, req ConfluentBillingDatasourceRequest) (*ConfluentBillingDatasource, error) {
	var out confluentBillingDatasourceAPIResponse
	if err := c.createBillingDatasource(ctx, req.toAPIRequest(), &out); err != nil {
		return nil, err
	}
	normalized := out.toConfluentBillingDatasource()
	if normalized.ID == "" {
		return nil, missingDatasourceIDError()
	}
	return normalized, nil
}

// GetConfluentBillingDatasource gets a Confluent billing datasource by ID.
func (c *Client) GetConfluentBillingDatasource(ctx context.Context, datasourceID string) (*ConfluentBillingDatasource, error) {
	var out confluentBillingDatasourceAPIResponse
	if err := c.getBillingDatasource(ctx, datasourceID, &out); err != nil {
		return nil, err
	}
	normalized := out.toConfluentBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}
	return normalized, nil
}

// ValidateDatadogBillingDatasource validates a Datadog billing datasource before creation.
func (c *Client) ValidateDatadogBillingDatasource(ctx context.Context, req DatadogBillingDatasourceRequest) error {
	return c.validateBillingDatasource(ctx, req.toAPIRequest())
}

// CreateDatadogBillingDatasource creates a Datadog billing datasource and returns its API representation.
func (c *Client) CreateDatadogBillingDatasource(ctx context.Context, req DatadogBillingDatasourceRequest) (*DatadogBillingDatasource, error) {
	var out datadogBillingDatasourceAPIResponse
	if err := c.createBillingDatasource(ctx, req.toAPIRequest(), &out); err != nil {
		return nil, err
	}
	normalized := out.toDatadogBillingDatasource()
	if normalized.ID == "" {
		return nil, missingDatasourceIDError()
	}
	return normalized, nil
}

// GetDatadogBillingDatasource gets a Datadog billing datasource by ID.
func (c *Client) GetDatadogBillingDatasource(ctx context.Context, datasourceID string) (*DatadogBillingDatasource, error) {
	var out datadogBillingDatasourceAPIResponse
	if err := c.getBillingDatasource(ctx, datasourceID, &out); err != nil {
		return nil, err
	}
	normalized := out.toDatadogBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}
	return normalized, nil
}

// ValidateAivenBillingDatasource validates an Aiven billing datasource before creation.
func (c *Client) ValidateAivenBillingDatasource(ctx context.Context, req AivenBillingDatasourceRequest) error {
	return c.validateBillingDatasource(ctx, req.toAPIRequest())
}

// CreateAivenBillingDatasource creates an Aiven billing datasource and returns its API representation.
func (c *Client) CreateAivenBillingDatasource(ctx context.Context, req AivenBillingDatasourceRequest) (*AivenBillingDatasource, error) {
	var out aivenBillingDatasourceAPIResponse
	if err := c.createBillingDatasource(ctx, req.toAPIRequest(), &out); err != nil {
		return nil, err
	}
	normalized := out.toAivenBillingDatasource()
	if normalized.ID == "" {
		return nil, missingDatasourceIDError()
	}
	return normalized, nil
}

// GetAivenBillingDatasource gets an Aiven billing datasource by ID.
func (c *Client) GetAivenBillingDatasource(ctx context.Context, datasourceID string) (*AivenBillingDatasource, error) {
	var out aivenBillingDatasourceAPIResponse
	if err := c.getBillingDatasource(ctx, datasourceID, &out); err != nil {
		return nil, err
	}
	normalized := out.toAivenBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}
	return normalized, nil
}

// ValidateGCPCUDExportBillingDatasource validates a GCP CUD export billing datasource before creation.
func (c *Client) ValidateGCPCUDExportBillingDatasource(ctx context.Context, req GCPCUDExportBillingDatasourceRequest) error {
	return c.validateBillingDatasource(ctx, req.toAPIRequest())
}

// CreateGCPCUDExportBillingDatasource creates a GCP CUD export billing datasource and returns its API representation.
func (c *Client) CreateGCPCUDExportBillingDatasource(ctx context.Context, req GCPCUDExportBillingDatasourceRequest) (*GCPCUDExportBillingDatasource, error) {
	var out gcpCUDExportBillingDatasourceAPIResponse
	if err := c.createBillingDatasource(ctx, req.toAPIRequest(), &out); err != nil {
		return nil, err
	}
	normalized := out.toGCPCUDExportBillingDatasource()
	if normalized.ID == "" {
		return nil, missingDatasourceIDError()
	}
	return normalized, nil
}

// GetGCPCUDExportBillingDatasource gets a GCP CUD export billing datasource by ID.
func (c *Client) GetGCPCUDExportBillingDatasource(ctx context.Context, datasourceID string) (*GCPCUDExportBillingDatasource, error) {
	var out gcpCUDExportBillingDatasourceAPIResponse
	if err := c.getBillingDatasource(ctx, datasourceID, &out); err != nil {
		return nil, err
	}
	normalized := out.toGCPCUDExportBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}
	return normalized, nil
}

// ValidateSnowflakeBillingDatasource validates a Snowflake billing datasource before creation.
func (c *Client) ValidateSnowflakeBillingDatasource(ctx context.Context, req SnowflakeBillingDatasourceRequest) error {
	return c.validateBillingDatasource(ctx, req.toAPIRequest())
}

// CreateSnowflakeBillingDatasource creates a Snowflake billing datasource and returns its API representation.
func (c *Client) CreateSnowflakeBillingDatasource(ctx context.Context, req SnowflakeBillingDatasourceRequest) (*SnowflakeBillingDatasource, error) {
	var out snowflakeBillingDatasourceAPIResponse
	if err := c.createBillingDatasource(ctx, req.toAPIRequest(), &out); err != nil {
		return nil, err
	}
	normalized := out.toSnowflakeBillingDatasource()
	if normalized.ID == "" {
		return nil, missingDatasourceIDError()
	}
	return normalized, nil
}

// GetSnowflakeBillingDatasource gets a Snowflake billing datasource by ID.
func (c *Client) GetSnowflakeBillingDatasource(ctx context.Context, datasourceID string) (*SnowflakeBillingDatasource, error) {
	var out snowflakeBillingDatasourceAPIResponse
	if err := c.getBillingDatasource(ctx, datasourceID, &out); err != nil {
		return nil, err
	}
	normalized := out.toSnowflakeBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}
	return normalized, nil
}

// ValidateCustomBigQueryBillingDatasource validates a custom BigQuery billing datasource before creation.
func (c *Client) ValidateCustomBigQueryBillingDatasource(ctx context.Context, req CustomBigQueryBillingDatasourceRequest) error {
	return c.validateBillingDatasource(ctx, req.toAPIRequest())
}

// CreateCustomBigQueryBillingDatasource creates a custom BigQuery billing datasource and returns its API representation.
func (c *Client) CreateCustomBigQueryBillingDatasource(ctx context.Context, req CustomBigQueryBillingDatasourceRequest) (*CustomBigQueryBillingDatasource, error) {
	var out customBigQueryBillingDatasourceAPIResponse
	if err := c.createBillingDatasource(ctx, req.toAPIRequest(), &out); err != nil {
		return nil, err
	}
	normalized := out.toCustomBigQueryBillingDatasource()
	if normalized.ID == "" {
		return nil, missingDatasourceIDError()
	}
	return normalized, nil
}

// GetCustomBigQueryBillingDatasource gets a custom BigQuery billing datasource by ID.
func (c *Client) GetCustomBigQueryBillingDatasource(ctx context.Context, datasourceID string) (*CustomBigQueryBillingDatasource, error) {
	var out customBigQueryBillingDatasourceAPIResponse
	if err := c.getBillingDatasource(ctx, datasourceID, &out); err != nil {
		return nil, err
	}
	normalized := out.toCustomBigQueryBillingDatasource()
	if normalized.ID == "" {
		normalized.ID = datasourceID
	}
	return normalized, nil
}
