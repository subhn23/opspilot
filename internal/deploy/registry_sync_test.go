package deploy

import (
	"context"
	"opspilot/internal/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MockImageService struct {
	Images []string
	Pulled []string
	Pushed []string
}

func (m *MockImageService) ListImages(ctx context.Context, registryAddr string) ([]string, error) {
	return m.Images, nil
}

func (m *MockImageService) PullImage(ctx context.Context, image string) error {
	m.Pulled = append(m.Pulled, image)
	return nil
}

func (m *MockImageService) PushImage(ctx context.Context, image, targetRegistry string) error {
	m.Pushed = append(m.Pushed, image)
	return nil
}

func TestSyncNodes(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.AuditLog{})

	service := &MockImageService{
		Images: []string{"app:v1", "db:latest"},
	}

	sync := NewRegistrySync(db, "host1:5000", "host2:5000", service)
	err := sync.SyncNodes(context.Background())

	if err != nil {
		t.Fatalf("SyncNodes failed: %v", err)
	}

	if len(service.Pulled) != 2 {
		t.Errorf("Expected 2 pulled images, got %d", len(service.Pulled))
	}

	if len(service.Pushed) != 2 {
		t.Errorf("Expected 2 pushed images, got %d", len(service.Pushed))
	}

	// Verify Audit Log
	var auditEntry models.AuditLog
	err = db.Where("action = ?", "REGISTRY_SYNC").First(&auditEntry).Error
	if err != nil {
		t.Errorf("Failed to find audit log entry: %v", err)
	}
}
