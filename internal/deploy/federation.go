package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"opspilot/internal/models"
)

// FederatedClient handles communication between OpsPilot instances
type FederatedClient struct {
	HTTPClient *http.Client
}

// Deploy sends a deployment request to a Worker node
func (f *FederatedClient) Deploy(ctx context.Context, workerURL, token string, req models.FederationRequest) (string, error) {
	if f.HTTPClient == nil {
		f.HTTPClient = http.DefaultClient
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/federation/deploy", workerURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Federation-Token", token)

	resp, err := f.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request to worker: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return string(body), fmt.Errorf("worker returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Status string `json:"status"`
		Logs   string `json:"logs"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return string(body), fmt.Errorf("failed to unmarshal worker response: %w", err)
	}

	return result.Logs, nil
}
