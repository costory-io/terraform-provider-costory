package costoryapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientAzureBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	var validateCalls int
	var createCalls int
	var getCalls int
	var deleteCalls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == routeBillingDatasourceValidate:
			validateCalls++
			assertAzureCreateRequest(t, r)
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == routeBillingDatasourceBase:
			createCalls++
			assertAzureCreateRequest(t, r)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"azure-ds-1","type":"Azure","name":"Azure Billing","storageAccountName":"storageaccount","containerName":"billing-exports","actualsPath":"actuals","amortizedPath":"amortized","status":"ACTIVE"}`))
		case r.Method == http.MethodGet && r.URL.Path == routeBillingDatasourceByID("azure-ds-1"):
			getCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"type":"Azure","name":"Azure Billing","storageAccountName":"storageaccount","containerName":"billing-exports","actualsPath":"actuals","amortizedPath":"amortized","status":"ACTIVE"}`))
		case r.Method == http.MethodDelete && r.URL.Path == routeBillingDatasourceByID("azure-ds-1"):
			deleteCalls++
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())

	createRequest := AzureBillingDatasourceRequest{
		Name:               "Azure Billing",
		SASURL:             "https://account.blob.core.windows.net/billing-exports?sv=2024-01-01&sig=example",
		StorageAccountName: "storageaccount",
		ContainerName:      "billing-exports",
		ActualsPath:        "actuals",
		AmortizedPath:      "amortized",
	}

	if err := client.ValidateAzureBillingDatasource(context.Background(), createRequest); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}

	created, err := client.CreateAzureBillingDatasource(context.Background(), createRequest)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if created.ID != "azure-ds-1" {
		t.Fatalf("unexpected created id: got %q, want %q", created.ID, "azure-ds-1")
	}

	current, err := client.GetAzureBillingDatasource(context.Background(), "azure-ds-1")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}

	if current.ID != "azure-ds-1" {
		t.Fatalf("unexpected get id: got %q, want %q", current.ID, "azure-ds-1")
	}

	if current.StorageAccountName != "storageaccount" {
		t.Fatalf("unexpected storage account: got %q", current.StorageAccountName)
	}

	if current.ContainerName != "billing-exports" {
		t.Fatalf("unexpected container name: got %q", current.ContainerName)
	}

	if current.ActualsPath != "actuals" || current.AmortizedPath != "amortized" {
		t.Fatalf("unexpected paths: actuals=%q amortized=%q", current.ActualsPath, current.AmortizedPath)
	}

	if err := client.DeleteBillingDatasource(context.Background(), "azure-ds-1"); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if validateCalls != 1 || createCalls != 1 || getCalls != 1 || deleteCalls != 1 {
		t.Fatalf(
			"unexpected call counters validate/create/get/delete: %d/%d/%d/%d",
			validateCalls, createCalls, getCalls, deleteCalls,
		)
	}
}

func assertAzureCreateRequest(t *testing.T, r *http.Request) {
	t.Helper()

	if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
		t.Fatalf("unexpected auth header: got %q, want %q", got, want)
	}

	var payload azureBillingDatasourceAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("unable to decode request body: %v", err)
	}

	if payload.Type != billingDatasourceTypeAzure {
		t.Fatalf("unexpected datasource type: got %q, want %q", payload.Type, billingDatasourceTypeAzure)
	}

	if payload.Name != "Azure Billing" {
		t.Fatalf("unexpected datasource name: got %q", payload.Name)
	}

	if payload.SASURL == "" {
		t.Fatal("expected sas url to be set")
	}

	if payload.StorageAccountName != "storageaccount" || payload.ContainerName != "billing-exports" {
		t.Fatalf("unexpected storage/container: %q/%q", payload.StorageAccountName, payload.ContainerName)
	}

	if payload.ActualsPath != "actuals" || payload.AmortizedPath != "amortized" {
		t.Fatalf("unexpected paths: actuals=%q amortized=%q", payload.ActualsPath, payload.AmortizedPath)
	}
}
