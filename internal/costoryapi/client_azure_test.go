package costoryapi

import (
	"context"
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
			_, _ = w.Write([]byte(`{"id":"azure-ds-1","type":"Azure","name":"Azure Billing","clientId":"11111111-1111-1111-1111-111111111111","tenantId":"22222222-2222-2222-2222-222222222222","storageAccountName":"storageaccount","containerName":"billing-exports","actualsPath":"actuals","amortizedPath":"amortized","status":"ACTIVE"}`))
		case r.Method == http.MethodGet && r.URL.Path == routeBillingDatasourceByID("azure-ds-1"):
			getCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"type":"Azure","name":"Azure Billing","clientId":"11111111-1111-1111-1111-111111111111","tenantId":"22222222-2222-2222-2222-222222222222","storageAccountName":"storageaccount","containerName":"billing-exports","actualsPath":"actuals","amortizedPath":"amortized","status":"ACTIVE"}`))
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
		ClientID:           "11111111-1111-1111-1111-111111111111",
		TenantID:           "22222222-2222-2222-2222-222222222222",
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

	if current.ClientID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected client id: got %q", current.ClientID)
	}

	if current.TenantID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("unexpected tenant id: got %q", current.TenantID)
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

	assertJSONObject(t, r, map[string]any{
		"type":               billingDatasourceTypeAzure,
		"name":               "Azure Billing",
		"clientId":           "11111111-1111-1111-1111-111111111111",
		"tenantId":           "22222222-2222-2222-2222-222222222222",
		"storageAccountName": "storageaccount",
		"containerName":      "billing-exports",
		"actualsPath":        "actuals",
		"amortizedPath":      "amortized",
	})
}
