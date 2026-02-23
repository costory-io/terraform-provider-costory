package costoryapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestClientGetServiceAccount(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("unexpected method: got %q, want %q", got, want)
		}

		if got, want := r.URL.Path, routeServiceAccount; got != want {
			t.Fatalf("unexpected path: got %q, want %q", got, want)
		}

		if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
			t.Fatalf("unexpected auth header: got %q, want %q", got, want)
		}

		if got, want := r.Header.Get("X-Costory-Slug"), "test-slug"; got != want {
			t.Fatalf("unexpected slug header: got %q, want %q", got, want)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"service_account":"sa-test","sub_ids":["sub-1","sub-2"]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-slug", "test-token", server.Client())

	got, err := client.GetServiceAccount(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := &ServiceAccountResponse{
		ServiceAccount: "sa-test",
		SubIDs:         []string{"sub-1", "sub-2"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected context response: got %#v, want %#v", got, want)
	}
}

func TestClientGetServiceAccountUnexpectedStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-slug", "test-token", server.Client())

	_, err := client.GetServiceAccount(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
