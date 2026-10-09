package costoryapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPutEnrichmentReturnsOneErrorPerMessage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/enrichments/02_namespace_to_org" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":["priority must be an integer >= 0","name has an empty table suffix"]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())
	err := client.PutEnrichment(context.Background(), "02_namespace_to_org", map[string]any{"name": "02_namespace_to_org"})
	var statusErr *StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected StatusError, got %v", err)
	}
	if statusErr.Status != http.StatusBadRequest || len(statusErr.Errors) != 2 {
		t.Fatalf("unexpected status error: %+v", statusErr)
	}
}

func TestDeleteMissingEnrichmentIsNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())
	err := client.DeleteEnrichment(context.Background(), "gone")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestPublishEnrichments(t *testing.T) {
	t.Parallel()

	status := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/enrichments/draft/publish" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(status)
		switch status {
		case http.StatusOK:
			_, _ = w.Write([]byte(`{"changes":{"added":["b"],"changed":[],"removed":["a"]},"workflowStarted":false,"workflowError":"please wait"}`))
		case http.StatusBadRequest:
			_, _ = w.Write([]byte(`{"errors":["b: cos_unknown is neither a known column nor a target of an earlier group"]}`))
		}
	}))
	defer server.Close()
	client := NewClient(server.URL, "test-token", server.Client())

	result, err := client.PublishEnrichments(context.Background())
	if err != nil || result == nil || result.Changes.Added[0] != "b" || result.Changes.Removed[0] != "a" || result.WorkflowError != "please wait" {
		t.Fatalf("unexpected publish result: %+v, %v", result, err)
	}

	status = http.StatusNotFound
	result, err = client.PublishEnrichments(context.Background())
	if err != nil || result != nil {
		t.Fatalf("expected nothing to publish, got %+v, %v", result, err)
	}

	status = http.StatusBadRequest
	_, err = client.PublishEnrichments(context.Background())
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.Status != http.StatusBadRequest || len(statusErr.Errors) != 1 {
		t.Fatalf("expected the compile error, got %v", err)
	}
}
