package costoryapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestClientOpenAIBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	req := OpenAIBillingDatasourceRequest{
		Name:        "OpenAI Billing",
		AdminAPIKey: "sk-openai-admin",
		StartDate:   stringPointer("2025-01-01"),
	}

	runConnectorCRUDTest(t, connectorCRUDTest{
		id:         "openai-ds-1",
		want:       map[string]any{"type": "OpenAI", "name": "OpenAI Billing", "adminApiKey": "sk-openai-admin", "startDate": "2025-01-01"},
		createBody: `{"id":"openai-ds-1","type":"OpenAI","name":"OpenAI Billing","bqTableUri":"proj.openai_raw.openai-ds-1","status":"ACTIVE"}`,
		getBody:    `{"id":"openai-ds-1","type":"OpenAI","name":"OpenAI Billing","bqTableUri":"proj.openai_raw.openai-ds-1","status":"ACTIVE","startDate":"2025-01-01"}`,
		run: func(t *testing.T, client *Client) {
			if err := client.ValidateOpenAIBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected validate error: %v", err)
			}
			created, err := client.CreateOpenAIBillingDatasource(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected create error: %v", err)
			}
			if created.ID != "openai-ds-1" {
				t.Fatalf("unexpected created id: got %q", created.ID)
			}
			current, err := client.GetOpenAIBillingDatasource(context.Background(), "openai-ds-1")
			if err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}
			if current.BQTableURI != "proj.openai_raw.openai-ds-1" {
				t.Fatalf("unexpected bq table uri: got %q", current.BQTableURI)
			}
		},
	})
}

func TestClientAnthropicClaudeAiBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	req := AnthropicClaudeAiBillingDatasourceRequest{
		Name:            "Claude AI Billing",
		AnalyticsAPIKey: "sk-ant-analytics",
		AccountName:     "acme",
	}

	runConnectorCRUDTest(t, connectorCRUDTest{
		id:         "claude-ds-1",
		want:       map[string]any{"type": "AnthropicClaudeAi", "name": "Claude AI Billing", "analyticsApiKey": "sk-ant-analytics", "accountName": "acme"},
		createBody: `{"id":"claude-ds-1","type":"AnthropicClaudeAi","name":"Claude AI Billing","bqTableUri":"proj.claude_raw.claude-ds-1","status":"ACTIVE"}`,
		getBody:    `{"id":"claude-ds-1","type":"AnthropicClaudeAi","name":"Claude AI Billing","bqTableUri":"proj.claude_raw.claude-ds-1","status":"ACTIVE"}`,
		run: func(t *testing.T, client *Client) {
			if err := client.ValidateAnthropicClaudeAiBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected validate error: %v", err)
			}
			created, err := client.CreateAnthropicClaudeAiBillingDatasource(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected create error: %v", err)
			}
			if created.ID != "claude-ds-1" {
				t.Fatalf("unexpected created id: got %q", created.ID)
			}
			if _, err := client.GetAnthropicClaudeAiBillingDatasource(context.Background(), "claude-ds-1"); err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}
		},
	})
}

func TestClientClickHouseCloudBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	req := ClickHouseCloudBillingDatasourceRequest{
		Name:           "ClickHouse Billing",
		KeyID:          "key-id",
		KeySecret:      "key-secret",
		OrganizationID: "org-ch",
	}

	runConnectorCRUDTest(t, connectorCRUDTest{
		id: "ch-ds-1",
		want: map[string]any{
			"type": "ClickHouseCloud", "name": "ClickHouse Billing",
			"keyId": "key-id", "keySecret": "key-secret", "organizationId": "org-ch",
		},
		createBody: `{"id":"ch-ds-1","type":"ClickHouseCloud","name":"ClickHouse Billing","bqTableUri":"proj.ch_raw.ch-ds-1","status":"ACTIVE"}`,
		getBody:    `{"id":"ch-ds-1","type":"ClickHouseCloud","name":"ClickHouse Billing","bqTableUri":"proj.ch_raw.ch-ds-1","status":"ACTIVE"}`,
		run: func(t *testing.T, client *Client) {
			if err := client.ValidateClickHouseCloudBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected validate error: %v", err)
			}
			if _, err := client.CreateClickHouseCloudBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected create error: %v", err)
			}
			if _, err := client.GetClickHouseCloudBillingDatasource(context.Background(), "ch-ds-1"); err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}
		},
	})
}

func TestClientCloudflareBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	req := CloudflareBillingDatasourceRequest{
		Name:      "Cloudflare Billing",
		APIToken:  "cf-token",
		AccountID: "cf-account",
	}

	runConnectorCRUDTest(t, connectorCRUDTest{
		id:         "cf-ds-1",
		want:       map[string]any{"type": "Cloudflare", "name": "Cloudflare Billing", "apiToken": "cf-token", "accountId": "cf-account"},
		createBody: `{"id":"cf-ds-1","type":"Cloudflare","name":"Cloudflare Billing","bqTableUri":"proj.cf_raw.cf-ds-1","status":"ACTIVE"}`,
		getBody:    `{"id":"cf-ds-1","type":"Cloudflare","name":"Cloudflare Billing","bqTableUri":"proj.cf_raw.cf-ds-1","status":"ACTIVE"}`,
		run: func(t *testing.T, client *Client) {
			if err := client.ValidateCloudflareBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected validate error: %v", err)
			}
			if _, err := client.CreateCloudflareBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected create error: %v", err)
			}
			if _, err := client.GetCloudflareBillingDatasource(context.Background(), "cf-ds-1"); err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}
		},
	})
}

func TestClientScalewayBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	req := ScalewayBillingDatasourceRequest{
		Name:           "Scaleway Billing",
		SecretKey:      "scw-secret",
		OrganizationID: "scw-org",
		AccessKey:      stringPointer("scw-access"),
	}

	runConnectorCRUDTest(t, connectorCRUDTest{
		id: "scw-ds-1",
		want: map[string]any{
			"type": "Scaleway", "name": "Scaleway Billing",
			"secretKey": "scw-secret", "organizationId": "scw-org", "accessKey": "scw-access",
		},
		createBody: `{"id":"scw-ds-1","type":"Scaleway","name":"Scaleway Billing","bqTableUri":"proj.scw_raw.scw-ds-1","status":"ACTIVE"}`,
		getBody:    `{"id":"scw-ds-1","type":"Scaleway","name":"Scaleway Billing","bqTableUri":"proj.scw_raw.scw-ds-1","status":"ACTIVE"}`,
		run: func(t *testing.T, client *Client) {
			if err := client.ValidateScalewayBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected validate error: %v", err)
			}
			if _, err := client.CreateScalewayBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected create error: %v", err)
			}
			if _, err := client.GetScalewayBillingDatasource(context.Background(), "scw-ds-1"); err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}
		},
	})
}

func TestClientConfluentBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	req := ConfluentBillingDatasourceRequest{
		Name:      "Confluent Billing",
		APIKey:    "ccloud-key",
		APISecret: "ccloud-secret",
	}

	runConnectorCRUDTest(t, connectorCRUDTest{
		id:         "ccloud-ds-1",
		want:       map[string]any{"type": "CONFLUENT", "name": "Confluent Billing", "apiKey": "ccloud-key", "apiSecret": "ccloud-secret"},
		createBody: `{"id":"ccloud-ds-1","type":"CONFLUENT","name":"Confluent Billing","bqTableUri":"proj.ccloud_raw.ccloud-ds-1","status":"ACTIVE"}`,
		getBody:    `{"id":"ccloud-ds-1","type":"CONFLUENT","name":"Confluent Billing","bqTableUri":"proj.ccloud_raw.ccloud-ds-1","status":"ACTIVE"}`,
		run: func(t *testing.T, client *Client) {
			if err := client.ValidateConfluentBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected validate error: %v", err)
			}
			if _, err := client.CreateConfluentBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected create error: %v", err)
			}
			if _, err := client.GetConfluentBillingDatasource(context.Background(), "ccloud-ds-1"); err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}
		},
	})
}

func TestClientDatadogBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	req := DatadogBillingDatasourceRequest{
		Name:           "Datadog Billing",
		APIKey:         "dd-api",
		ApplicationKey: "dd-app",
		Region:         "datadoghq.eu",
	}

	runConnectorCRUDTest(t, connectorCRUDTest{
		id: "dd-ds-1",
		want: map[string]any{
			"type": "Datadog", "name": "Datadog Billing",
			"apiKey": "dd-api", "applicationKey": "dd-app", "region": "datadoghq.eu",
		},
		createBody: `{"id":"dd-ds-1","type":"Datadog","name":"Datadog Billing","integrationId":"dd-int-1","bqTableUri":"proj.dd_raw.dd-ds-1","status":"ACTIVE"}`,
		getBody:    `{"id":"dd-ds-1","type":"Datadog","name":"Datadog Billing","integrationId":"dd-int-1","bqTableUri":"proj.dd_raw.dd-ds-1","status":"ACTIVE"}`,
		run: func(t *testing.T, client *Client) {
			if err := client.ValidateDatadogBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected validate error: %v", err)
			}
			created, err := client.CreateDatadogBillingDatasource(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected create error: %v", err)
			}
			if created.IntegrationID != "dd-int-1" {
				t.Fatalf("unexpected integration id: got %q", created.IntegrationID)
			}
			if _, err := client.GetDatadogBillingDatasource(context.Background(), "dd-ds-1"); err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}
		},
	})
}

func TestClientAivenBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	req := AivenBillingDatasourceRequest{
		Name:           "Aiven Billing",
		APISecret:      "aiven-token",
		OrganizationID: "aiven-org",
	}

	runConnectorCRUDTest(t, connectorCRUDTest{
		id:         "aiven-ds-1",
		want:       map[string]any{"type": "Aiven", "name": "Aiven Billing", "apiSecret": "aiven-token", "organizationId": "aiven-org"},
		createBody: `{"id":"aiven-ds-1","type":"Aiven","name":"Aiven Billing","organizationId":"aiven-org","bqTableUri":"proj.aiven_raw.aiven-ds-1","status":"ACTIVE"}`,
		getBody:    `{"id":"aiven-ds-1","type":"Aiven","name":"Aiven Billing","organizationId":"aiven-org","bqTableUri":"proj.aiven_raw.aiven-ds-1","status":"ACTIVE"}`,
		run: func(t *testing.T, client *Client) {
			if err := client.ValidateAivenBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected validate error: %v", err)
			}
			created, err := client.CreateAivenBillingDatasource(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected create error: %v", err)
			}
			if created.OrganizationID != "aiven-org" {
				t.Fatalf("unexpected organization id: got %q", created.OrganizationID)
			}
			if _, err := client.GetAivenBillingDatasource(context.Background(), "aiven-ds-1"); err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}
		},
	})
}

func TestClientGCPCUDExportBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	req := GCPCUDExportBillingDatasourceRequest{
		Name:        "GCP CUD Export",
		BQTablePath: "my-project.billing.cud_export",
	}

	runConnectorCRUDTest(t, connectorCRUDTest{
		id:         "cud-ds-1",
		want:       map[string]any{"type": "GCP_CUD_EXPORT", "name": "GCP CUD Export", "bqTablePath": "my-project.billing.cud_export"},
		createBody: `{"id":"cud-ds-1","type":"GCP_CUD_EXPORT","name":"GCP CUD Export","bqTablePath":"my-project.billing.cud_export","status":"ACTIVE"}`,
		getBody:    `{"id":"cud-ds-1","type":"GCP_CUD_EXPORT","name":"GCP CUD Export","bqTablePath":"my-project.billing.cud_export","status":"ACTIVE"}`,
		run: func(t *testing.T, client *Client) {
			if err := client.ValidateGCPCUDExportBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected validate error: %v", err)
			}
			created, err := client.CreateGCPCUDExportBillingDatasource(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected create error: %v", err)
			}
			if created.BQTablePath != "my-project.billing.cud_export" {
				t.Fatalf("unexpected bq table path: got %q", created.BQTablePath)
			}
			if _, err := client.GetGCPCUDExportBillingDatasource(context.Background(), "cud-ds-1"); err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}
		},
	})
}

func TestClientSnowflakeBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	req := SnowflakeBillingDatasourceRequest{
		Name:          "Snowflake Billing",
		IntegrationID: "sf-int-1",
	}

	runConnectorCRUDTest(t, connectorCRUDTest{
		id:         "sf-ds-1",
		want:       map[string]any{"type": "SNOWFLAKE", "name": "Snowflake Billing", "integrationId": "sf-int-1"},
		createBody: `{"id":"sf-ds-1","type":"SNOWFLAKE","name":"Snowflake Billing","integrationId":"sf-int-1","bqTableUris":["proj.sf_raw.a"],"status":"ACTIVE"}`,
		getBody:    `{"id":"sf-ds-1","type":"SNOWFLAKE","name":"Snowflake Billing","integrationId":"sf-int-1","bqTableUris":["proj.sf_raw.a"],"status":"ACTIVE"}`,
		run: func(t *testing.T, client *Client) {
			if err := client.ValidateSnowflakeBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected validate error: %v", err)
			}
			created, err := client.CreateSnowflakeBillingDatasource(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected create error: %v", err)
			}
			if !reflect.DeepEqual(created.BQTableURIs, []string{"proj.sf_raw.a"}) {
				t.Fatalf("unexpected bq table uris: got %#v", created.BQTableURIs)
			}
			if _, err := client.GetSnowflakeBillingDatasource(context.Background(), "sf-ds-1"); err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}
		},
	})
}

func TestClientCustomBigQueryBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	req := CustomBigQueryBillingDatasourceRequest{
		Name:         "Custom BQ Billing",
		BQTablePath:  "my-project.custom.costs",
		ProviderName: "acme",
		Mapping: map[string]string{
			"billedCost":  "cost",
			"serviceName": "sku",
		},
	}

	runConnectorCRUDTest(t, connectorCRUDTest{
		id: "cbq-ds-1",
		want: map[string]any{
			"type": "CustomBigQuery", "name": "Custom BQ Billing",
			"bqTablePath": "my-project.custom.costs", "providerName": "acme",
			"mapping": map[string]any{"billedCost": "cost", "serviceName": "sku"},
		},
		createBody: `{"id":"cbq-ds-1","type":"CustomBigQuery","name":"Custom BQ Billing","bqTablePath":"my-project.custom.costs","providerName":"acme","mapping":{"billedCost":"cost","serviceName":"sku"},"status":"ACTIVE"}`,
		getBody:    `{"id":"cbq-ds-1","type":"CustomBigQuery","name":"Custom BQ Billing","bqTablePath":"my-project.custom.costs","providerName":"acme","mapping":{"billedCost":"cost","serviceName":"sku"},"status":"ACTIVE"}`,
		run: func(t *testing.T, client *Client) {
			if err := client.ValidateCustomBigQueryBillingDatasource(context.Background(), req); err != nil {
				t.Fatalf("unexpected validate error: %v", err)
			}
			created, err := client.CreateCustomBigQueryBillingDatasource(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected create error: %v", err)
			}
			if created.Mapping["billedCost"] != "cost" {
				t.Fatalf("unexpected mapping: got %#v", created.Mapping)
			}
			if _, err := client.GetCustomBigQueryBillingDatasource(context.Background(), "cbq-ds-1"); err != nil {
				t.Fatalf("unexpected get error: %v", err)
			}
		},
	})
}

type connectorCRUDTest struct {
	id         string
	want       map[string]any
	createBody string
	getBody    string
	run        func(t *testing.T, client *Client)
}

func runConnectorCRUDTest(t *testing.T, test connectorCRUDTest) {
	t.Helper()

	var validateCalls int
	var createCalls int
	var getCalls int
	var deleteCalls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == routeBillingDatasourceValidate:
			validateCalls++
			assertJSONObject(t, r, test.want)
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == routeBillingDatasourceBase:
			createCalls++
			assertJSONObject(t, r, test.want)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(test.createBody))
		case r.Method == http.MethodGet && r.URL.Path == routeBillingDatasourceByID(test.id):
			getCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(test.getBody))
		case r.Method == http.MethodDelete && r.URL.Path == routeBillingDatasourceByID(test.id):
			deleteCalls++
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())
	test.run(t, client)

	if err := client.DeleteBillingDatasource(context.Background(), test.id); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if validateCalls != 1 || createCalls != 1 || getCalls != 1 || deleteCalls != 1 {
		t.Fatalf(
			"unexpected call counters validate/create/get/delete: %d/%d/%d/%d",
			validateCalls, createCalls, getCalls, deleteCalls,
		)
	}
}

func assertJSONObject(t *testing.T, r *http.Request, want map[string]any) {
	t.Helper()

	if got, wantAuth := r.Header.Get("Authorization"), "Bearer test-token"; got != wantAuth {
		t.Fatalf("unexpected auth header: got %q, want %q", got, wantAuth)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("unable to read request body: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unable to decode request body: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected request body: got %#v, want %#v", got, want)
	}
}
