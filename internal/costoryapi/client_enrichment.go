package costoryapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// StatusError is an API failure that carries the enrichment `errors` array.
type StatusError struct {
	Status int
	Errors []string
}

func (e *StatusError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("unexpected status code %d", e.Status)
	}
	return strings.Join(e.Errors, "\n")
}

// PutEnrichment stores one enrichment document as a draft. Nothing is live
// until PublishEnrichments.
func (c *Client) PutEnrichment(ctx context.Context, name string, document map[string]any) error {
	body, statusCode, err := c.doJSON(ctx, http.MethodPut, enrichmentPath(name), document)
	if err != nil {
		return err
	}
	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return nil
	}
	return statusError(statusCode, body)
}

// GetEnrichment reads one live enrichment document.
func (c *Client) GetEnrichment(ctx context.Context, name string) (map[string]any, error) {
	body, statusCode, err := c.doJSON(ctx, http.MethodGet, enrichmentPath(name), nil)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if statusCode != http.StatusOK {
		return nil, statusError(statusCode, body)
	}
	var document map[string]any
	if err := json.Unmarshal(body, &document); err != nil {
		return nil, fmt.Errorf("decode enrichment: %w", err)
	}
	return document, nil
}

// DeleteEnrichment drafts the removal of a live enrichment document, or drops
// the draft of one that is not live. ErrNotFound when there is neither.
func (c *Client) DeleteEnrichment(ctx context.Context, name string) error {
	body, statusCode, err := c.doJSON(ctx, http.MethodDelete, enrichmentPath(name), nil)
	if err != nil {
		return err
	}
	if statusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if statusCode == http.StatusNoContent || statusCode == http.StatusOK {
		return nil
	}
	return statusError(statusCode, body)
}

// EnrichmentChanges lists enrichment names by change.
type EnrichmentChanges struct {
	Added   []string `json:"added"`
	Changed []string `json:"changed"`
	Removed []string `json:"removed"`
}

// EnrichmentPublishResult is the answer of a publish.
type EnrichmentPublishResult struct {
	Changes         EnrichmentChanges `json:"changes"`
	WorkflowStarted bool              `json:"workflowStarted"`
	WorkflowError   string            `json:"workflowError,omitempty"`
}

// PublishEnrichments compiles the live set with every draft applied and makes
// it live in one transaction, then starts the billing update. It returns nil
// when there is no draft. A set that does not compile is a StatusError (400),
// drafts that changed while publishing are a StatusError (409).
func (c *Client) PublishEnrichments(ctx context.Context) (*EnrichmentPublishResult, error) {
	body, statusCode, err := c.doJSON(ctx, http.MethodPost, "/enrichments/draft/publish", nil)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotFound {
		return nil, nil
	}
	if statusCode != http.StatusOK {
		return nil, statusError(statusCode, body)
	}
	var result EnrichmentPublishResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode enrichment publish: %w", err)
	}
	return &result, nil
}

func enrichmentPath(name string) string {
	return "/enrichments/" + url.PathEscape(name)
}

func statusError(statusCode int, body []byte) error {
	var payload struct {
		Errors []string `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && len(payload.Errors) > 0 {
		return &StatusError{Status: statusCode, Errors: payload.Errors}
	}
	return unexpectedStatusError(statusCode, body)
}
