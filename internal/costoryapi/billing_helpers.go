package costoryapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) validateBillingDatasource(ctx context.Context, request any) error {
	body, statusCode, err := c.doJSON(ctx, http.MethodPost, routeBillingDatasourceValidate, request)
	if err != nil {
		return err
	}

	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return nil
	}

	return unexpectedStatusError(statusCode, body)
}

func (c *Client) createBillingDatasource(ctx context.Context, request any, dest any) error {
	body, statusCode, err := c.doJSON(ctx, http.MethodPost, routeBillingDatasourceBase, request)
	if err != nil {
		return err
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return unexpectedStatusError(statusCode, body)
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}

func (c *Client) getBillingDatasource(ctx context.Context, datasourceID string, dest any) error {
	body, statusCode, err := c.doJSON(ctx, http.MethodGet, routeBillingDatasourceByID(datasourceID), nil)
	if err != nil {
		return err
	}

	if statusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if statusCode != http.StatusOK {
		return unexpectedStatusError(statusCode, body)
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}
