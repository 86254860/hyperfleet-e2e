package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
)

// HyperFleetClient is a wrapper around the generated Client that provides
// convenience methods and better error handling for E2E tests.
type HyperFleetClient struct {
	*openapi.Client
}

// NewHyperFleetClient creates a new HyperFleet API client.
func NewHyperFleetClient(baseURL string, httpClient *http.Client) (*HyperFleetClient, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	client, err := openapi.NewClient(baseURL, openapi.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &HyperFleetClient{
		Client: client,
	}, nil
}

// handleHTTPResponse is a generic helper for processing HTTP responses.
// It handles status code validation, response body decoding, and error formatting.
func handleHTTPResponse[T any](resp *http.Response, expectedStatus int, action string) (*T, error) {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != expectedStatus {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected status code %d for %s (failed to read response body: %w)",
				resp.StatusCode, action, err)
		}
		return nil, fmt.Errorf("unexpected status code %d for %s: %s",
			resp.StatusCode, action, string(body))
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode %s response: %w", action, err)
	}

	return &result, nil
}
