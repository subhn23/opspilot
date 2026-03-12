package deploy

import (
	"context"
	"fmt"
	"opspilot/internal/crypto"
	"opspilot/internal/models"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MockScanner struct {
	Safe   bool
	Report string
	Err    error
}

func (m *MockScanner) Scan(ctx context.Context, imageName string) (bool, string, error) {
	return m.Safe, m.Report, m.Err
}

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Deployment{}, &models.AuditLog{}, &models.TargetHost{}, &models.Environment{})
	return db
}

func TestScanImage(t *testing.T) {
	ctx := context.Background()

	t.Run("Safe Image", func(t *testing.T) {
		mock := &MockScanner{Safe: true, Report: "Clean"}
		deployer := &Deployer{Scanner: mock, Git: &MockGitClient{}, Docker: &MockDockerClient{}, Federation: &FederatedClient{}}

		safe, report, err := deployer.ScanImage(ctx, "test-image")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !safe {
			t.Error("Expected image to be safe")
		}
		if report != "Clean" {
			t.Errorf("Expected report 'Clean', got %s", report)
		}
	})

	t.Run("Unsafe Image", func(t *testing.T) {
		mock := &MockScanner{Safe: false, Report: "Vulnerability Found"}
		deployer := &Deployer{DB: setupTestDB(), Scanner: mock, Git: &MockGitClient{}, Docker: &MockDockerClient{}, Federation: &FederatedClient{}}

		deploy := &models.Deployment{CommitHash: "unsafe123"}
		deployer.DB.Create(deploy)

		safe, report, err := deployer.ScanImage(ctx, "unsafe-image")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if safe {
			t.Error("Expected image to be unsafe")
		}
		if report != "Vulnerability Found" {
			t.Errorf("Expected report 'Vulnerability Found', got %s", report)
		}

		// Test integration in BuildAndPush
		err = deployer.BuildAndPush(ctx, deploy)
		if err == nil {
			t.Fatal("Expected error in BuildAndPush for unsafe image")
		}

		var updated models.Deployment
		deployer.DB.First(&updated, deploy.ID)
		if updated.Status != "FAILED_SECURITY" {
			t.Errorf("Expected status FAILED_SECURITY, got %s", updated.Status)
		}

		// Verify Audit Log
		var auditEntry models.AuditLog
		err = deployer.DB.Where("action = ?", "SECURITY_FAILURE").First(&auditEntry).Error
		if err != nil {
			t.Errorf("Failed to find audit log entry for SECURITY_FAILURE: %v", err)
		}
	})
}

type MockSSHClient struct {
	CommandsRun  []string
	MockOutput   string
	MockErr      error
	ConfigLoaded bool
}

func (m *MockSSHClient) RunCommand(ctx context.Context, addr, command string) (string, error) {
	m.CommandsRun = append(m.CommandsRun, command)
	return m.MockOutput, m.MockErr
}

func (m *MockSSHClient) Configure(user, privateKey string) {
	if user != "" && privateKey != "" {
		m.ConfigLoaded = true
	}
}

func TestRemoteUp(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	mockSSH := &MockSSHClient{MockOutput: "Done"}
	deployer := &Deployer{DB: db, SSH: mockSSH, Git: &MockGitClient{}, Docker: &MockDockerClient{}, Federation: &FederatedClient{}}

	deploy := &models.Deployment{
		CommitHash: "abc1234",
	}
	db.Create(deploy)

	err := deployer.RemoteUp(ctx, deploy, "10.0.0.50")

	if err != nil {
		t.Fatalf("RemoteUp failed: %v", err)
	}

	if len(mockSSH.CommandsRun) != 2 {
		t.Errorf("Expected 2 commands, ran %d", len(mockSSH.CommandsRun))
	}

	var updated models.Deployment
	db.First(&updated, deploy.ID)
	if updated.Status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", updated.Status)
	}

	// Verify Audit Log
	var auditEntry models.AuditLog
	err = db.Where("action = ?", "DEPLOY_SUCCESS").First(&auditEntry).Error
	if err != nil {
		t.Errorf("Failed to find audit log entry: %v", err)
	}
}

func TestRemoteUpWithTargetHost(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	mockSSH := &MockSSHClient{MockOutput: "Done"}
	deployer := &Deployer{DB: db, SSH: mockSSH, Git: &MockGitClient{}, Docker: &MockDockerClient{}, Federation: &FederatedClient{}}

	os.Setenv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	defer os.Unsetenv("ENCRYPTION_KEY")

	// 1. Create TargetHost with encrypted AuthData
	authData := "fake-ssh-key"
	encrypted, _ := crypto.Encrypt(authData)

	host := models.TargetHost{
		Name:     "Dynamic Host",
		Type:     "remote_ssh",
		Endpoint: "192.168.1.50",
		AuthData: encrypted,
	}
	db.Create(&host)

	// 2. Create Environment linked to Host
	env := models.Environment{
		Name:         "Dynamic Env",
		TargetHostID: &host.ID,
	}
	db.Create(&env)

	deploy := &models.Deployment{
		EnvironmentID: env.ID,
		CommitHash:    "abc1234",
	}
	db.Create(deploy)

	err := deployer.RemoteUp(ctx, deploy, "192.168.1.50")
	if err != nil {
		t.Fatalf("RemoteUp failed: %v", err)
	}

	if !mockSSH.ConfigLoaded {
		t.Error("SSH config (key) was not loaded from TargetHost")
	}
}

func TestBuildAndPush(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	mockScanner := &MockScanner{Safe: true, Report: "All good"}
	deployer := &Deployer{DB: db, Scanner: mockScanner, Git: &MockGitClient{}, Docker: &MockDockerClient{}, Federation: &FederatedClient{}}

	deploy := &models.Deployment{
		CommitHash: "feat123",
	}
	db.Create(deploy)

	err := deployer.BuildAndPush(ctx, deploy)

	if err != nil {
		t.Fatalf("BuildAndPush failed: %v", err)
	}

	var updated models.Deployment
	db.First(&updated, deploy.ID)
	if updated.Status != "PUSHED" {
		t.Errorf("Expected status PUSHED, got %s", updated.Status)
	}
}

type MockGitClient struct {
	CloneCalled    bool
	CheckoutCalled bool
	CloneError     error
	CheckoutError  error
}

func (m *MockGitClient) Clone(ctx context.Context, repoURL, targetDir string) error {
	m.CloneCalled = true
	return m.CloneError
}

func (m *MockGitClient) Checkout(ctx context.Context, targetDir, commitHash string) error {
	m.CheckoutCalled = true
	return m.CheckoutError
}

func TestGitIntegration(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	mockGit := &MockGitClient{}
	mockDocker := &MockDockerClient{}
	mockScanner := &MockScanner{Safe: true}
	deployer := &Deployer{DB: db, Scanner: mockScanner, Git: mockGit, Docker: mockDocker, Federation: &FederatedClient{}}

	deploy := &models.Deployment{
		CommitHash: "feat123",
		Branch:     "main",
	}
	db.Create(deploy)

	err := deployer.BuildAndPush(ctx, deploy)

	if err != nil {
		t.Fatalf("BuildAndPush failed: %v", err)
	}

	if !mockGit.CloneCalled {
		t.Error("Git Clone was not called")
	}
	if !mockGit.CheckoutCalled {
		t.Error("Git Checkout was not called")
	}

	// Test Failure
	mockGit.CloneError = fmt.Errorf("git error")
	err = deployer.BuildAndPush(ctx, deploy)
	if err == nil {
		t.Error("Expected error from git clone")
	}

	var updated models.Deployment
	db.First(&updated, deploy.ID)
	if updated.Status != "FAILED_BUILD" {
		t.Errorf("Expected status FAILED_BUILD, got %s", updated.Status)
	}
}

type MockDockerClient struct {
	LoginCalled bool
	BuildCalled bool
	PushCalled  bool
	LoginErr    error
	BuildErr    error
	PushErr     error
}

func (m *MockDockerClient) Login(ctx context.Context, user, pass, registry string) error {
	m.LoginCalled = true
	return m.LoginErr
}

func (m *MockDockerClient) Build(ctx context.Context, workingDir, tag string) error {
	m.BuildCalled = true
	return m.BuildErr
}

func (m *MockDockerClient) Push(ctx context.Context, tag string) error {
	m.PushCalled = true
	return m.PushErr
}

func TestDockerIntegration(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	mockDocker := &MockDockerClient{}
	mockGit := &MockGitClient{}
	mockScanner := &MockScanner{Safe: true}
	deployer := &Deployer{DB: db, Scanner: mockScanner, Git: mockGit, Docker: mockDocker, Federation: &FederatedClient{}}

	os.Setenv("REGISTRY_URL", "localhost:5000")
	os.Setenv("REGISTRY_USER", "user")
	os.Setenv("REGISTRY_PASS", "pass")
	defer os.Unsetenv("REGISTRY_URL")
	defer os.Unsetenv("REGISTRY_USER")
	defer os.Unsetenv("REGISTRY_PASS")

	deploy := &models.Deployment{
		CommitHash: "feat123",
	}
	db.Create(deploy)

	err := deployer.BuildAndPush(ctx, deploy)

	if err != nil {
		t.Fatalf("BuildAndPush failed: %v", err)
	}

	if !mockDocker.LoginCalled {
		t.Error("Docker Login was not called")
	}
	if !mockDocker.BuildCalled {
		t.Error("Docker Build was not called")
	}
	if !mockDocker.PushCalled {
		t.Error("Docker Push was not called")
	}
}
