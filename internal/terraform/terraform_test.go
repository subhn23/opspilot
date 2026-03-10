package terraform

import (
	"context"
	"encoding/json"
	"opspilot/internal/models"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-exec/tfexec"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockTerraformClient implements TerraformClient for testing
type MockTerraformClient struct {
	ApplyCalled   bool
	DestroyCalled bool
	InitCalled    bool
	OutputCalled  bool
	ApplyError    error
	DestroyError  error
	InitError     error
	OutputError   error
	MockOutputs   map[string]tfexec.OutputMeta
}

func (m *MockTerraformClient) Apply(ctx context.Context, opts ...tfexec.ApplyOption) error {
	m.ApplyCalled = true
	return m.ApplyError
}

func (m *MockTerraformClient) Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error {
	m.DestroyCalled = true
	return m.DestroyError
}

func (m *MockTerraformClient) Init(ctx context.Context, opts ...tfexec.InitOption) error {
	m.InitCalled = true
	return m.InitError
}

func (m *MockTerraformClient) Output(ctx context.Context, opts ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error) {
	m.OutputCalled = true
	return m.MockOutputs, m.OutputError
}

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Environment{})
	return db
}

func TestProvision(t *testing.T) {
	db := setupTestDB()
	tmpDir, _ := os.MkdirTemp("", "tfengine_test")
	defer os.RemoveAll(tmpDir)

	mockClient := &MockTerraformClient{
		MockOutputs: map[string]tfexec.OutputMeta{
			"vm_ip": {
				Value: []byte(`"10.0.0.100"`),
			},
		},
	}

	engine, _ := NewTFEngine(db, tmpDir)
	engine.ClientFactory = func(workingDir, execPath string) (TerraformClient, error) {
		return mockClient, nil
	}

	env := &models.Environment{
		ID:       uuid.New(),
		Name:     "test-env",
		HostNode: "pve1",
		VMID:     100,
	}
	db.Create(env)

	err := engine.Provision(context.Background(), env)

	if err != nil {
		t.Fatalf("Provision failed: %v", err)
	}
	if !mockClient.InitCalled {
		t.Error("Init was not called")
	}
	if !mockClient.ApplyCalled {
		t.Error("Apply was not called")
	}
	if !mockClient.OutputCalled {
		t.Error("Output was not called")
	}

	// Verify DB update
	var updatedEnv models.Environment
	db.First(&updatedEnv, env.ID)

	if updatedEnv.Status != "HEALTHY" {
		t.Errorf("Expected status HEALTHY, got %s", updatedEnv.Status)
	}

	// Need to unmarshal the JSON string value
	var ip string
	json.Unmarshal([]byte(`"10.0.0.100"`), &ip)
	if updatedEnv.IPAddress != ip {
		t.Errorf("Expected IP 10.0.0.100, got %s", updatedEnv.IPAddress)
	}
}

func TestDestroy(t *testing.T) {
	db := setupTestDB()
	tmpDir, _ := os.MkdirTemp("", "tfengine_test")
	defer os.RemoveAll(tmpDir)

	mockClient := &MockTerraformClient{}

	engine, _ := NewTFEngine(db, tmpDir)
	engine.ClientFactory = func(workingDir, execPath string) (TerraformClient, error) {
		return mockClient, nil
	}

	env := &models.Environment{
		ID:     uuid.New(),
		Name:   "test-env",
		Status: "HEALTHY",
	}
	db.Create(env)

	err := engine.Destroy(context.Background(), env)

	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}
	if !mockClient.InitCalled {
		t.Error("Init was not called")
	}
	if !mockClient.DestroyCalled {
		t.Error("Destroy was not called")
	}

	// Verify DB update
	var updatedEnv models.Environment
	db.First(&updatedEnv, env.ID)
	if updatedEnv.Status != "DESTROYED" {
		t.Errorf("Expected status DESTROYED, got %s", updatedEnv.Status)
	}
}
