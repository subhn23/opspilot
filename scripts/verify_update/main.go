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

	log.Println("Creating environment in database...")
	if err := db.Create(&env).Error; err != nil {
		log.Fatalf("Failed to create environment: %v", err)
	}

	log.Println("Success! 'manual-env-verification' created. Check your browser for the live update.")
}
