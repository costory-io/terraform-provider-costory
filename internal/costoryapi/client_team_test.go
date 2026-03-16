package costoryapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientTeamCRUD(t *testing.T) {
	t.Parallel()

	var createCalls int
	var getCalls int
	var patchCalls int
	var deleteCalls int
	var addCalls int
	var removeCalls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == routeTeamsBase:
			createCalls++
			assertTeamCreateRequest(t, r)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"team-1","name":"Engineering","description":"Core platform team","visibility":"PRIVATE","createdAt":"2026-03-07T12:34:56.789Z","updatedAt":"2026-03-07T12:34:56.789Z"}`))
		case r.Method == http.MethodGet && r.URL.Path == routeTeamByID("team-1"):
			getCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"team-1","name":"Engineering","description":"Core platform team","visibility":"PRIVATE","createdAt":"2026-03-07T12:34:56.789Z","updatedAt":"2026-03-07T12:34:56.789Z"}`))
		case r.Method == http.MethodPatch && r.URL.Path == routeTeamByID("team-1"):
			patchCalls++
			assertTeamUpdateRequest(t, r)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"team-1","name":"Engineering","description":"Updated description","visibility":"PUBLIC","createdAt":"2026-03-07T12:34:56.789Z","updatedAt":"2026-03-08T12:34:56.789Z"}`))
		case r.Method == http.MethodDelete && r.URL.Path == routeTeamByID("team-1"):
			deleteCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))
		case r.Method == http.MethodPost && r.URL.Path == routeTeamMembersByID("team-1"):
			addCalls++
			assertTeamMemberAddRequest(t, r)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))
		case r.Method == http.MethodDelete && r.URL.Path == routeTeamMemberByID("team-1", "user-1"):
			removeCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())

	description := "Core platform team"
	visibility := "PRIVATE"
	created, err := client.CreateTeam(context.Background(), TeamCreateRequest{
		Name:        "Engineering",
		Description: &description,
		Visibility:  &visibility,
	})
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if created.ID != "team-1" {
		t.Fatalf("unexpected team id: got %q, want %q", created.ID, "team-1")
	}

	current, err := client.GetTeam(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}

	if current.Visibility != "PRIVATE" {
		t.Fatalf("unexpected visibility: got %q", current.Visibility)
	}

	updatedDescription := "Updated description"
	updatedVisibility := "PUBLIC"
	updated, err := client.UpdateTeam(context.Background(), "team-1", TeamUpdateRequest{
		Description: &updatedDescription,
		Visibility:  &updatedVisibility,
	})
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	if updated.Description != "Updated description" {
		t.Fatalf("unexpected updated description: got %q", updated.Description)
	}

	role := "OWNER"
	userID := "user-1"
	if err := client.AddTeamMember(context.Background(), "team-1", TeamMemberRequest{
		UserID: &userID,
		Role:   &role,
	}); err != nil {
		t.Fatalf("unexpected add team member error: %v", err)
	}

	if err := client.RemoveTeamMember(context.Background(), "team-1", "user-1"); err != nil {
		t.Fatalf("unexpected remove team member error: %v", err)
	}

	if err := client.DeleteTeam(context.Background(), "team-1"); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if createCalls != 1 || getCalls != 1 || patchCalls != 1 || deleteCalls != 1 || addCalls != 1 || removeCalls != 1 {
		t.Fatalf(
			"unexpected call counters create/get/patch/delete/add/remove: %d/%d/%d/%d/%d/%d",
			createCalls, getCalls, patchCalls, deleteCalls, addCalls, removeCalls,
		)
	}
}

func TestClientGetTeamNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())

	_, err := client.GetTeam(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func assertTeamCreateRequest(t *testing.T, r *http.Request) {
	t.Helper()

	if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
		t.Fatalf("unexpected auth header: got %q, want %q", got, want)
	}

	var payload teamCreateAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("unable to decode request body: %v", err)
	}

	if payload.Name != "Engineering" {
		t.Fatalf("unexpected team name: got %q", payload.Name)
	}

	if payload.Description == nil || *payload.Description != "Core platform team" {
		t.Fatalf("unexpected team description: %#v", payload.Description)
	}

	if payload.Visibility == nil || *payload.Visibility != "PRIVATE" {
		t.Fatalf("unexpected team visibility: %#v", payload.Visibility)
	}
}

func assertTeamUpdateRequest(t *testing.T, r *http.Request) {
	t.Helper()

	var payload teamUpdateAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("unable to decode request body: %v", err)
	}

	if payload.Name != nil {
		t.Fatalf("unexpected name in update payload: %v", payload.Name)
	}

	if payload.Description == nil || *payload.Description != "Updated description" {
		t.Fatalf("unexpected description in update payload: %#v", payload.Description)
	}

	if payload.Visibility == nil || *payload.Visibility != "PUBLIC" {
		t.Fatalf("unexpected visibility in update payload: %#v", payload.Visibility)
	}
}

func assertTeamMemberAddRequest(t *testing.T, r *http.Request) {
	t.Helper()

	var payload teamMemberAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("unable to decode request body: %v", err)
	}

	if payload.UserID == nil || *payload.UserID != "user-1" {
		t.Fatalf("unexpected user id: %#v", payload.UserID)
	}

	if payload.Role == nil || *payload.Role != "OWNER" {
		t.Fatalf("unexpected role: %#v", payload.Role)
	}
}
