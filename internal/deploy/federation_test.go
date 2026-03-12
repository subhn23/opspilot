package deploy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"opspilot/internal/crypto"
	"opspilot/internal/models"
	"os"
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

func TestDeployRouting(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	
	os.Setenv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	defer os.Unsetenv("ENCRYPTION_KEY")

	// Mock Federated Client
	federationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "SUCCESS", "logs": "Remote Federated Logs"})
	}))
	defer federationServer.Close()

	mockSSH := &MockSSHClient{MockOutput: "SSH Done"}
	mockDocker := &MockDockerClient{}
	mockGit := &MockGitClient{}
	mockScanner := &MockScanner{Safe: true}
	
	deployer := &Deployer{
		DB:         db,
		Scanner:    mockScanner,
		SSH:        mockSSH,
		Git:        mockGit,
		Docker:     mockDocker,
		Federation: &FederatedClient{},
	}

	t.Run("Route to SSH", func(t *testing.T) {
		host := models.TargetHost{Name: "SSH Host", Type: "remote_ssh", Endpoint: "1.2.3.4"}
		db.Create(&host)
		env := models.Environment{Name: "SSH Env", TargetHostID: &host.ID}
		db.Create(&env)
		deploy := models.Deployment{EnvironmentID: env.ID, CommitHash: "abc"}
		db.Create(&deploy)

		err := deployer.Deploy(ctx, &deploy, &host)
		if err != nil {
			t.Fatalf("Deploy failed: %v", err)
		}

		if !mockDocker.BuildCalled {
			t.Error("Docker Build should have been called for SSH host")
		}
		if len(mockSSH.CommandsRun) == 0 {
			t.Error("SSH commands should have been run for SSH host")
		}
	})

	t.Run("Route to Federation", func(t *testing.T) {
		token, _ := crypto.Encrypt("secret-token")
		host := models.TargetHost{
			Name:     "Fed Host",
			Type:     "federated_opspilot",
			Endpoint: federationServer.URL,
			AuthData: token,
		}
		db.Create(&host)
		env := models.Environment{Name: "Fed Env", TargetHostID: &host.ID}
		db.Create(&env)
		deploy := models.Deployment{EnvironmentID: env.ID, CommitHash: "xyz"}
		db.Create(&deploy)

		// Reset mocks
		mockDocker.BuildCalled = false
		mockSSH.CommandsRun = nil

		err := deployer.Deploy(ctx, &deploy, &host)
		if err != nil {
			t.Fatalf("Federated deploy failed: %v", err)
		}

		if mockDocker.BuildCalled {
			t.Error("Docker Build should NOT have been called for federated host")
		}
		if len(mockSSH.CommandsRun) > 0 {
			t.Error("SSH commands should NOT have been run for federated host")
		}
		if !contains(deploy.Logs, "Remote Federated Logs") {
			t.Error("Remote logs were not captured in deployment record")
		}
	})
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
