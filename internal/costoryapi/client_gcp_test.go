package costoryapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGCPBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	var validateCalls int
	var createCalls int
	var getCalls int
	var deleteCalls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == routeBillingDatasourceValidate:
			validateCalls++
			assertGCPCreateRequest(t, r)
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == routeBillingDatasourceBase:
			createCalls++
			assertGCPCreateRequest(t, r)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"gcp-ds-1","type":"GCP","name":"GCP Billing","bqUri":"project.dataset.table","isDetailedBilling":true}`))
		case r.Method == http.MethodGet && r.URL.Path == routeBillingDatasourceByID("gcp-ds-1"):
			getCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"gcp-ds-1","type":"GCP","name":"GCP Billing","bqUri":"project.dataset.table","isDetailedBilling":true,"startDate":"2025-01-01"}`))
		case r.Method == http.MethodDelete && r.URL.Path == routeBillingDatasourceByID("gcp-ds-1"):
			deleteCalls++
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-slug", "test-token", server.Client())

	createRequest := GCPBillingDatasourceRequest{
		Name:              "GCP Billing",
		BQURI:             "project.dataset.table",
		IsDetailedBilling: boolPointer(true),
		StartDate:         stringPointer("2025-01-01"),
	}

	if err := client.ValidateGCPBillingDatasource(context.Background(), createRequest); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}

	created, err := client.CreateGCPBillingDatasource(context.Background(), createRequest)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if created.ID != "gcp-ds-1" {
		t.Fatalf("unexpected created id: got %q, want %q", created.ID, "gcp-ds-1")
	}

	current, err := client.GetGCPBillingDatasource(context.Background(), "gcp-ds-1")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}

	if current.BQURI != "project.dataset.table" {
		t.Fatalf("unexpected bq uri: got %q", current.BQURI)
	}

	if current.IsDetailedBilling == nil || !*current.IsDetailedBilling {
		t.Fatal("expected detailed billing to be true")
	}

	if current.StartDate == nil || *current.StartDate != "2025-01-01" {
		t.Fatalf("unexpected start date: got %#v", current.StartDate)
	}

	if err := client.DeleteBillingDatasource(context.Background(), "gcp-ds-1"); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if validateCalls != 1 || createCalls != 1 || getCalls != 1 || deleteCalls != 1 {
		t.Fatalf(
			"unexpected call counters validate/create/get/delete: %d/%d/%d/%d",
			validateCalls, createCalls, getCalls, deleteCalls,
		)
	}
}

func TestClientGetGCPBillingDatasourceNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-slug", "test-token", server.Client())

	_, err := client.GetGCPBillingDatasource(context.Background(), "missing-id")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func assertGCPCreateRequest(t *testing.T, r *http.Request) {
	t.Helper()

	if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
		t.Fatalf("unexpected auth header: got %q, want %q", got, want)
	}

	if got, want := r.Header.Get("X-Costory-Slug"), "test-slug"; got != want {
		t.Fatalf("unexpected slug header: got %q, want %q", got, want)
	}

	var payload gcpBillingDatasourceAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("unable to decode request body: %v", err)
	}

	if payload.Type != billingDatasourceTypeGCP {
		t.Fatalf("unexpected datasource type: got %q, want %q", payload.Type, billingDatasourceTypeGCP)
	}

	if payload.Name != "GCP Billing" || payload.BQTablePath != "project.dataset.table" {
		t.Fatalf("unexpected create payload: %#v", payload)
	}
}

func boolPointer(value bool) *bool {
	return &value
}

func stringPointer(value string) *string {
	return &value
}
