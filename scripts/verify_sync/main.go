package main

import (
	"context"
	"fmt"
	"log"
	"opspilot/internal/config"
	"opspilot/internal/deploy"
	"opspilot/internal/models"
	"time"

	"gorm.io/driver/sqlite"
)

type MockImageService struct{}

func (m *MockImageService) ListImages(ctx context.Context, registryAddr string) ([]string, error) {
	return []string{"web-app:latest", "api-server:v1.2.0"}, nil
}

func (m *MockImageService) PullImage(ctx context.Context, image string) error {
	fmt.Printf("Mock: Pulling %s\n", image)
	return nil
}

func (m *MockImageService) PushImage(ctx context.Context, image, targetRegistry string) error {
	fmt.Printf("Mock: Pushing %s to %s\n", image, targetRegistry)
	return nil
}

func main() {
	// 1. Setup in-memory DB
	db := config.InitDB(sqlite.Open(":memory:"))
	db.AutoMigrate(&models.AuditLog{})

	service := &MockImageService{}
	sync := deploy.NewRegistrySync(db, "host1:5000", "host2:5000", service)
	sync.Interval = 2 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Step 1: Running single synchronization...")
	if err := sync.SyncNodes(ctx); err != nil {
		log.Fatalf("SyncNodes failed: %v", err)
	}
	fmt.Println("Single sync successful.")

	fmt.Println("\nStep 2: Starting background worker (will run for a few seconds)...")
	// We'll use a shorter ticker for verification
	go sync.StartWorker(ctx)

	select {
	case <-ctx.Done():
		fmt.Println("\nBackground worker test complete (timeout reached).")
	}

	fmt.Println("\nManual verification of Registry Synchronization logic PASSED.")
}
