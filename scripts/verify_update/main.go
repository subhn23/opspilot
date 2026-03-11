package main

import (
	"log"
	"opspilot/internal/config"
	"opspilot/internal/models"
)

func main() {
	// Initialize DB using the default PostgreSQL dialector
	db := config.InitDB(nil)

	// Create a new environment to trigger the AfterSave hook
	env := models.Environment{
		Name:      "manual-env-verification",
		Status:    "HEALTHY",
		VMID:      888,
		IPAddress: "10.0.0.88",
		Type:      "dev",
		HostNode:  "host1",
	}

	log.Println("Ensuring environment exists in database...")
	if err := db.Where(models.Environment{VMID: 888}).FirstOrCreate(&env, models.Environment{
		Name:      "manual-env-verification",
		Status:    "HEALTHY",
		VMID:      888,
		IPAddress: "10.0.0.88",
		Type:      "dev",
		HostNode:  "host1",
	}).Error; err != nil {
		log.Fatalf("Failed to ensure environment: %v", err)
	}

	// Create a mock deployment for this environment to show a clickable container node
	deploy := models.Deployment{
		EnvironmentID: env.ID,
		CommitHash:    "deadbeef",
		Branch:        "main",
		Status:        "SUCCESS",
		ContainerID:   "test-container-123", // Clickable in UI
	}

	log.Println("Creating mock deployment in database...")
	if err := db.Create(&deploy).Error; err != nil {
		log.Printf("Note: Deployment might already exist or failed: %v", err)
	}

	log.Println("Success! 'manual-env-verification' and mock container created.")
	log.Println("1. Refresh your dashboard.")
	log.Println("2. Click the 'App (deadbee)' node.")
	log.Println("3. Confirm metrics appear in the table.")
}
