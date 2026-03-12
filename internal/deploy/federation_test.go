package deploy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"opspilot/internal/models"
	"testing"
)

func TestFederatedClient_Deploy(t *testing.T) {
	// Mock Worker Node
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Federation-Token")
		if token != "valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req models.FederationRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.EnvironmentName == "fail" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "SUCCESS", "logs": "Remote logs"})
	}))
	defer server.Close()

	client := &FederatedClient{}
	req := models.FederationRequest{EnvironmentName: "test-env", CommitHash: "abc"}

	t.Run("Success", func(t *testing.T) {
		logs, err := client.Deploy(context.Background(), server.URL, "valid-token", req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if logs != "Remote logs" {
			t.Errorf("Expected 'Remote logs', got %s", logs)
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		_, err := client.Deploy(context.Background(), server.URL, "invalid", req)
		if err == nil {
			t.Error("Expected error for invalid token")
		}
	})

	t.Run("Worker Error", func(t *testing.T) {
		failReq := models.FederationRequest{EnvironmentName: "fail"}
		_, err := client.Deploy(context.Background(), server.URL, "valid-token", failReq)
		if err == nil {
			t.Error("Expected error for worker failure")
		}
	})
}
