package terraform

import (
	"context"
	"encoding/json"
	"opspilot/internal/models"
	"os"
	"path/filepath"
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

	baseDir, _ := os.MkdirTemp("", "tfbase")
	defer os.RemoveAll(baseDir)

	engine, _ := NewTFEngine(db, tmpDir)
	engine.BaseTemplatesDir = baseDir
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

	baseDir, _ := os.MkdirTemp("", "tfbase")
	defer os.RemoveAll(baseDir)

	engine, _ := NewTFEngine(db, tmpDir)
	engine.BaseTemplatesDir = baseDir
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

func TestTemplateMirroring(t *testing.T) {
	db := setupTestDB()
	tmpDir, _ := os.MkdirTemp("", "tfengine_test")
	defer os.RemoveAll(tmpDir)

	// Setup fake base templates
	baseDir := filepath.Join(tmpDir, "base")
	os.MkdirAll(baseDir, 0755)
	os.WriteFile(filepath.Join(baseDir, "main.tf"), []byte("resource..."), 0644)
	os.WriteFile(filepath.Join(baseDir, "variables.tf"), []byte("variable..."), 0644)
	os.WriteFile(filepath.Join(baseDir, "README.md"), []byte("not mirrored"), 0644)

	engine, _ := NewTFEngine(db, filepath.Join(tmpDir, "workspaces"))
	engine.BaseTemplatesDir = baseDir
	engine.ClientFactory = func(workingDir, execPath string) (TerraformClient, error) {
		return &MockTerraformClient{}, nil
	}

	workspace := "new-env"
	ctx := context.Background()

	// Initial setup
	_, err := engine.setupTF(ctx, workspace)
	if err != nil {
		t.Fatalf("setupTF failed: %v", err)
	}

	wsDir := filepath.Join(tmpDir, "workspaces", workspace)

	// 1. Verify files exist
	if _, err := os.Stat(filepath.Join(wsDir, "main.tf")); os.IsNotExist(err) {
		t.Error("main.tf was not mirrored")
	}
	if _, err := os.Stat(filepath.Join(wsDir, "variables.tf")); os.IsNotExist(err) {
		t.Error("variables.tf was not mirrored")
	}

	// 2. Verify non-tf file was NOT mirrored
	if _, err := os.Stat(filepath.Join(wsDir, "README.md")); err == nil {
		t.Error("README.md should not have been mirrored")
	}

	// 3. Verify content
	content, _ := os.ReadFile(filepath.Join(wsDir, "main.tf"))
	if string(content) != "resource..." {
		t.Errorf("Expected 'resource...', got %q", string(content))
	}
}
