package costoryapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientMetricsDatasourceCRUD(t *testing.T) {
	t.Parallel()

	var validateCalls int
	var createCalls int
	var getCalls int
	var patchCalls int
	var deleteCalls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == routeMetricsDatasourceValidate:
			validateCalls++
			assertMetricsCreateRequest(t, r)
			_, _ = w.Write([]byte(`{"isSuccess":true,"errors":[]}`))
		case r.Method == http.MethodPost && r.URL.Path == routeMetricsDatasourceBase:
			createCalls++
			assertMetricsCreateRequest(t, r)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"metrics-ds-1","type":"AwsS3V2","status":"PENDING","name":"S3 Metrics","bucketName":"metrics-bucket","prefix":"metrics/","roleArn":"arn:aws:iam::123456789012:role/costory","metricsDefinition":[{"metricName":"Usage","gapFilling":"ZERO","aggregation":"SUM","valueColumn":"usage_amount","dateColumn":"usage_start_date","dimensions":["service"],"unit":"count"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == routeMetricsDatasourceByID("metrics-ds-1"):
			getCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"metrics-ds-1","type":"AwsS3V2","status":"ACTIVE","name":"S3 Metrics","bucketName":"metrics-bucket","prefix":"metrics/","roleArn":"arn:aws:iam::123456789012:role/costory","metricsDefinition":[{"metricName":"Usage","gapFilling":"ZERO","aggregation":"SUM","valueColumn":"usage_amount","dateColumn":"usage_start_date","dimensions":["service"],"unit":"count"}]}`))
		case r.Method == http.MethodPatch && r.URL.Path == routeMetricsDatasourceByID("metrics-ds-1"):
			patchCalls++
			assertMetricsPatchRequest(t, r)
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == routeMetricsDatasourceByID("metrics-ds-1"):
			deleteCalls++
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())

	createRequest := MetricsDatasourceRequest{
		Name:       "S3 Metrics",
		Type:       metricsDatasourceTypeS3V2,
		BucketName: "metrics-bucket",
		Prefix:     "metrics/",
		RoleARN:    "arn:aws:iam::123456789012:role/costory",
		MetricsDefinitions: []MetricsDefinition{
			{
				MetricName:  "Usage",
				GapFilling:  "ZERO",
				Aggregation: "SUM",
				ValueColumn: "usage_amount",
				DateColumn:  "usage_start_date",
				Dimensions:  []string{"service"},
				Unit:        "count",
			},
		},
	}

	if err := client.ValidateMetricsDatasource(context.Background(), createRequest); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}

	created, err := client.CreateMetricsDatasource(context.Background(), createRequest)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if created.ID != "metrics-ds-1" {
		t.Fatalf("unexpected created id: got %q, want %q", created.ID, "metrics-ds-1")
	}

	current, err := client.GetMetricsDatasource(context.Background(), "metrics-ds-1")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}

	if current.BucketName != "metrics-bucket" {
		t.Fatalf("unexpected bucket name: got %q", current.BucketName)
	}

	if len(current.MetricsDefinition) != 1 || current.MetricsDefinition[0].MetricName != "Usage" {
		t.Fatalf("unexpected metrics definition: got %#v", current.MetricsDefinition)
	}

	updatedDefs := []MetricsDefinition{
		{
			MetricName:  "Usage",
			GapFilling:  "ZERO",
			Aggregation: "SUM",
			ValueColumn: "usage_amount",
			DateColumn:  "usage_start_date",
			Dimensions:  []string{"service", "region"},
			Unit:        "count",
		},
	}

	if err := client.UpdateMetricsDatasource(context.Background(), "metrics-ds-1", updatedDefs); err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	if err := client.DeleteMetricsDatasource(context.Background(), "metrics-ds-1"); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if validateCalls != 1 || createCalls != 1 || getCalls != 1 || patchCalls != 1 || deleteCalls != 1 {
		t.Fatalf(
			"unexpected call counters validate/create/get/patch/delete: %d/%d/%d/%d/%d",
			validateCalls, createCalls, getCalls, patchCalls, deleteCalls,
		)
	}
}

func TestClientGetMetricsDatasourceNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())

	_, err := client.GetMetricsDatasource(context.Background(), "missing-id")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestClientValidateMetricsDatasourceFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != routeMetricsDatasourceValidate {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"isSuccess":false,"errors":["Invalid bucket name","Role cannot access bucket"]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())

	err := client.ValidateMetricsDatasource(context.Background(), MetricsDatasourceRequest{
		Name:       "S3 Metrics",
		Type:       metricsDatasourceTypeS3V2,
		BucketName: "bad-bucket",
		Prefix:     "",
		RoleARN:    "arn:aws:iam::123456789012:role/costory",
		MetricsDefinitions: []MetricsDefinition{
			{
				MetricName:  "Usage",
				GapFilling:  "ZERO",
				Aggregation: "SUM",
				ValueColumn: "usage_amount",
				DateColumn:  "usage_start_date",
			},
		},
	})
	if err == nil {
		t.Fatal("expected validate error, got nil")
	}

	if got := err.Error(); got != "validation failed: Invalid bucket name; Role cannot access bucket" {
		t.Fatalf("unexpected validate error message: got %q", got)
	}
}

func assertMetricsCreateRequest(t *testing.T, r *http.Request) {
	t.Helper()

	if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
		t.Fatalf("unexpected auth header: got %q, want %q", got, want)
	}

	var payload metricsDatasourceAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("unable to decode request body: %v", err)
	}

	if payload.Type != metricsDatasourceTypeS3V2 {
		t.Fatalf("unexpected datasource type: got %q, want %q", payload.Type, metricsDatasourceTypeS3V2)
	}

	if payload.Name != "S3 Metrics" || payload.BucketName != "metrics-bucket" || payload.RoleARN != "arn:aws:iam::123456789012:role/costory" {
		t.Fatalf("unexpected create payload: %#v", payload)
	}

	if len(payload.MetricsDefinition) != 1 || payload.MetricsDefinition[0].MetricName != "Usage" {
		t.Fatalf("unexpected metrics definition: got %#v", payload.MetricsDefinition)
	}
}

func assertMetricsPatchRequest(t *testing.T, r *http.Request) {
	t.Helper()

	if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
		t.Fatalf("unexpected auth header: got %q, want %q", got, want)
	}

	var payload metricsDatasourcePatchAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("unable to decode request body: %v", err)
	}

	if len(payload.MetricsDefinition) != 1 || payload.MetricsDefinition[0].MetricName != "Usage" {
		t.Fatalf("unexpected patch payload: got %#v", payload.MetricsDefinition)
	}

	if len(payload.MetricsDefinition[0].Dimensions) != 2 {
		t.Fatalf("unexpected dimensions count: got %d", len(payload.MetricsDefinition[0].Dimensions))
	}
}
