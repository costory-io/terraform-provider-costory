package costoryapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientAWSBillingDatasourceCRUD(t *testing.T) {
	t.Parallel()

	var createCalls int
	var getCalls int
	var deleteCalls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == routeBillingDatasourceBase:
			createCalls++
			assertAWSCreateRequest(t, r)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"aws-ds-1","type":"AWS","status":"PENDING","name":"AWS Billing","bucketName":"billing-bucket","roleArn":"arn:aws:iam::123456789012:role/costory","prefix":"cur/","eksSplitDataEnabled":false}`))
		case r.Method == http.MethodGet && r.URL.Path == routeBillingDatasourceByID("aws-ds-1"):
			getCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"aws-ds-1","type":"AWS","status":"ACTIVE","name":"AWS Billing","bucketName":"billing-bucket","roleArn":"arn:aws:iam::123456789012:role/costory","prefix":"cur/","eksSplitDataEnabled":false,"startDate":"2025-01-01","eksSplit":true}`))
		case r.Method == http.MethodDelete && r.URL.Path == routeBillingDatasourceByID("aws-ds-1"):
			deleteCalls++
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())

	createRequest := AWSBillingDatasourceRequest{
		Name:                "AWS Billing",
		BucketName:          "billing-bucket",
		RoleARN:             "arn:aws:iam::123456789012:role/costory",
		Prefix:              "cur/",
		EKSSplitDataEnabled: boolPointer(false),
		StartDate:           stringPointer("2025-01-01"),
		EKSSplit:            boolPointer(true),
	}

	created, err := client.CreateAWSBillingDatasource(context.Background(), createRequest)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if created.ID != "aws-ds-1" {
		t.Fatalf("unexpected created id: got %q, want %q", created.ID, "aws-ds-1")
	}

	if created.Status == nil || *created.Status != "PENDING" {
		t.Fatalf("unexpected created status: got %#v", created.Status)
	}

	current, err := client.GetAWSBillingDatasource(context.Background(), "aws-ds-1")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}

	if current.BucketName != "billing-bucket" {
		t.Fatalf("unexpected bucket name: got %q", current.BucketName)
	}

	if current.EKSSplitDataEnabled == nil || *current.EKSSplitDataEnabled {
		t.Fatal("expected eks_split_data_enabled to be false")
	}

	if current.StartDate == nil || *current.StartDate != "2025-01-01" {
		t.Fatalf("unexpected start date: got %#v", current.StartDate)
	}

	if current.EKSSplit == nil || !*current.EKSSplit {
		t.Fatalf("unexpected eks split flag: got %#v", current.EKSSplit)
	}

	if current.Status == nil || *current.Status != "ACTIVE" {
		t.Fatalf("unexpected current status: got %#v", current.Status)
	}

	if err := client.DeleteBillingDatasource(context.Background(), "aws-ds-1"); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if createCalls != 1 || getCalls != 1 || deleteCalls != 1 {
		t.Fatalf(
			"unexpected call counters create/get/delete: %d/%d/%d",
			createCalls, getCalls, deleteCalls,
		)
	}
}

func TestClientGetAWSBillingDatasourceNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())

	_, err := client.GetAWSBillingDatasource(context.Background(), "missing-id")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestClientCreateAWSBillingDatasourceValidationError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != routeBillingDatasourceBase {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"aws_access_denied","reason":"Cannot access bucket with provided role"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())

	_, err := client.CreateAWSBillingDatasource(context.Background(), AWSBillingDatasourceRequest{
		Name:       "AWS Billing",
		BucketName: "billing-bucket",
		RoleARN:    "arn:aws:iam::123456789012:role/costory",
		Prefix:     "cur/",
	})
	if err == nil {
		t.Fatal("expected create error, got nil")
	}

	if got, want := err.Error(), "unexpected status code 403: error=aws_access_denied reason=Cannot access bucket with provided role"; got != want {
		t.Fatalf("unexpected create error message: got %q, want %q", got, want)
	}
}

func assertAWSCreateRequest(t *testing.T, r *http.Request) {
	t.Helper()

	if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
		t.Fatalf("unexpected auth header: got %q, want %q", got, want)
	}

	var payload awsBillingDatasourceAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("unable to decode request body: %v", err)
	}

	if payload.Type != billingDatasourceTypeAWS {
		t.Fatalf("unexpected datasource type: got %q, want %q", payload.Type, billingDatasourceTypeAWS)
	}

	if payload.Name != "AWS Billing" || payload.BucketName != "billing-bucket" || payload.RoleARN != "arn:aws:iam::123456789012:role/costory" || payload.Prefix != "cur/" {
		t.Fatalf("unexpected create payload: %#v", payload)
	}
}
