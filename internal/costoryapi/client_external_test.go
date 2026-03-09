package costoryapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientCursorBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	var validateCalls int
	var createCalls int
	var getCalls int
	var deleteCalls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == routeBillingDatasourceValidate:
			validateCalls++
			assertExternalCreateRequest(t, r, billingDatasourceTypeCursor, "Cursor Billing", "sk_cursor_admin_123")
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == routeBillingDatasourceBase:
			createCalls++
			assertExternalCreateRequest(t, r, billingDatasourceTypeCursor, "Cursor Billing", "sk_cursor_admin_123")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"cursor-ds-1","type":"Cursor","name":"Cursor Billing","bqTableUri":"example-project.cursor_raw.cursor-ds-1","status":"ACTIVE"}`))
		case r.Method == http.MethodGet && r.URL.Path == routeBillingDatasourceByID("cursor-ds-1"):
			getCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"cursor-ds-1","type":"Cursor","name":"Cursor Billing","bqTableUri":"example-project.cursor_raw.cursor-ds-1","status":"ACTIVE","startDate":"2025-01-01","endDate":"2025-02-01"}`))
		case r.Method == http.MethodDelete && r.URL.Path == routeBillingDatasourceByID("cursor-ds-1"):
			deleteCalls++
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())

	createRequest := CursorBillingDatasourceRequest{
		Name:        "Cursor Billing",
		AdminAPIKey: "sk_cursor_admin_123",
		StartDate:   stringPointer("2025-01-01"),
		EndDate:     stringPointer("2025-02-01"),
	}

	if err := client.ValidateCursorBillingDatasource(context.Background(), createRequest); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}

	created, err := client.CreateCursorBillingDatasource(context.Background(), createRequest)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if created.ID != "cursor-ds-1" {
		t.Fatalf("unexpected created id: got %q, want %q", created.ID, "cursor-ds-1")
	}

	current, err := client.GetCursorBillingDatasource(context.Background(), "cursor-ds-1")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}

	if current.BQTableURI != "example-project.cursor_raw.cursor-ds-1" {
		t.Fatalf("unexpected bq table uri: got %q", current.BQTableURI)
	}

	if current.StartDate == nil || *current.StartDate != "2025-01-01" {
		t.Fatalf("unexpected start date: got %#v", current.StartDate)
	}

	if current.EndDate == nil || *current.EndDate != "2025-02-01" {
		t.Fatalf("unexpected end date: got %#v", current.EndDate)
	}

	if err := client.DeleteBillingDatasource(context.Background(), "cursor-ds-1"); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if validateCalls != 1 || createCalls != 1 || getCalls != 1 || deleteCalls != 1 {
		t.Fatalf(
			"unexpected call counters validate/create/get/delete: %d/%d/%d/%d",
			validateCalls, createCalls, getCalls, deleteCalls,
		)
	}
}

func TestClientAnthropicBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	var validateCalls int
	var createCalls int
	var getCalls int
	var deleteCalls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == routeBillingDatasourceValidate:
			validateCalls++
			assertExternalCreateRequest(t, r, billingDatasourceTypeAnthropic, "Anthropic Billing", "sk-ant-admin-xyz")
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == routeBillingDatasourceBase:
			createCalls++
			assertExternalCreateRequest(t, r, billingDatasourceTypeAnthropic, "Anthropic Billing", "sk-ant-admin-xyz")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"anthropic-ds-1","type":"Anthropic","name":"Anthropic Billing","bqTableUri":"example-project.anthropic_raw.anthropic-ds-1","status":"ACTIVE"}`))
		case r.Method == http.MethodGet && r.URL.Path == routeBillingDatasourceByID("anthropic-ds-1"):
			getCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"anthropic-ds-1","type":"Anthropic","name":"Anthropic Billing","bqTableUri":"example-project.anthropic_raw.anthropic-ds-1","status":"ACTIVE","startDate":"2025-03-01"}`))
		case r.Method == http.MethodDelete && r.URL.Path == routeBillingDatasourceByID("anthropic-ds-1"):
			deleteCalls++
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())

	createRequest := AnthropicBillingDatasourceRequest{
		Name:        "Anthropic Billing",
		AdminAPIKey: "sk-ant-admin-xyz",
		StartDate:   stringPointer("2025-03-01"),
	}

	if err := client.ValidateAnthropicBillingDatasource(context.Background(), createRequest); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}

	created, err := client.CreateAnthropicBillingDatasource(context.Background(), createRequest)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if created.ID != "anthropic-ds-1" {
		t.Fatalf("unexpected created id: got %q, want %q", created.ID, "anthropic-ds-1")
	}

	current, err := client.GetAnthropicBillingDatasource(context.Background(), "anthropic-ds-1")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}

	if current.BQTableURI != "example-project.anthropic_raw.anthropic-ds-1" {
		t.Fatalf("unexpected bq table uri: got %q", current.BQTableURI)
	}

	if current.StartDate == nil || *current.StartDate != "2025-03-01" {
		t.Fatalf("unexpected start date: got %#v", current.StartDate)
	}

	if err := client.DeleteBillingDatasource(context.Background(), "anthropic-ds-1"); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if validateCalls != 1 || createCalls != 1 || getCalls != 1 || deleteCalls != 1 {
		t.Fatalf(
			"unexpected call counters validate/create/get/delete: %d/%d/%d/%d",
			validateCalls, createCalls, getCalls, deleteCalls,
		)
	}
}

func assertExternalCreateRequest(t *testing.T, r *http.Request, expectedType, expectedName, expectedKey string) {
	t.Helper()

	if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
		t.Fatalf("unexpected auth header: got %q, want %q", got, want)
	}

	var payload externalBillingDatasourceAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("unable to decode request body: %v", err)
	}

	if payload.Type != expectedType {
		t.Fatalf("unexpected datasource type: got %q, want %q", payload.Type, expectedType)
	}

	if payload.Name != expectedName {
		t.Fatalf("unexpected datasource name: got %q, want %q", payload.Name, expectedName)
	}

	if payload.AdminAPIKey != expectedKey {
		t.Fatalf("unexpected admin api key: got %q, want %q", payload.AdminAPIKey, expectedKey)
	}
}
