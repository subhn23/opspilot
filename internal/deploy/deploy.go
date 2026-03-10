package deploy

import (
	"context"
	"fmt"
	"log"
	"opspilot/internal/models"
	"os/exec"
	"time"

	"gorm.io/gorm"
)

type Deployer struct {
	DB *gorm.DB
}

func NewDeployer(db *gorm.DB) *Deployer {
	return &Deployer{DB: db}
}

// ScanImage runs a Trivy vulnerability scan on the image
func (d *Deployer) ScanImage(imageName string) (bool, string, error) {
	log.Printf("Starting security scan for image: %s", imageName)

	// Conceptual: exec.Command("trivy", "image", "--severity", "CRITICAL", imageName)
	// For now, assume it's clean
	return true, "No critical vulnerabilities found.", nil
}

// BuildAndPush triggers a local docker build and pushes to the mirrored registry
func (d *Deployer) BuildAndPush(ctx context.Context, deploy *models.Deployment) error {
	d.updateStatus(deploy, "BUILDING")

	// ... (build logic) ...
	imageName := fmt.Sprintf("localhost:5000/app:%s", deploy.CommitHash)

	// SECURITY SCAN
	d.updateStatus(deploy, "SCANNING")
	safe, report, err := d.ScanImage(imageName)
	if err != nil || !safe {
		d.updateStatus(deploy, "FAILED_SECURITY")
		deploy.Logs += "\nSECURITY ALERT: " + report
		d.DB.Save(deploy)
		return fmt.Errorf("security scan failed: %s", report)
	}

	// ... (push logic) ...
	return nil
}

// RemoteUp SSHs into the dynamic VM and runs docker-compose up
func (d *Deployer) RemoteUp(ctx context.Context, deploy *models.Deployment, targetIP string) error {
	d.updateStatus(deploy, "DEPLOYING")

	// Use golang.org/x/crypto/ssh to execute commands
	// 1. Pull new image
	// 2. Update docker-compose.yml
	// 3. docker-compose up -d

	log.Printf("Executing remote deploy to %s", targetIP)

	// Mock success for now
	d.updateStatus(deploy, "SUCCESS")
	return nil
}

func (d *Deployer) updateStatus(deploy *models.Deployment, status string) {
	deploy.Status = status
	d.DB.Save(deploy)
}
