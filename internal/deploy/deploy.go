package deploy

import (
	"context"
	"fmt"
	"log"
	"opspilot/internal/models"
	"os/exec"

	"gorm.io/gorm"
)

// Scanner abstracts the vulnerability scanning logic
type Scanner interface {
	Scan(ctx context.Context, imageName string) (bool, string, error)
}

// RealScanner uses the Trivy binary to scan images
type RealScanner struct{}

func (s *RealScanner) Scan(ctx context.Context, imageName string) (bool, string, error) {
	log.Printf("Trivy: Scanning image %s", imageName)
	cmd := exec.CommandContext(ctx, "trivy", "image", "--severity", "CRITICAL,HIGH", "--exit-code", "1", imageName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// If exit code is 1, vulnerabilities were found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, string(output), nil
		}
		return false, string(output), fmt.Errorf("trivy execution failed: %w", err)
	}

	return true, "No critical or high vulnerabilities found.", nil
}

type Deployer struct {
	DB      *gorm.DB
	Scanner Scanner
}

func NewDeployer(db *gorm.DB) *Deployer {
	return &Deployer{
		DB:      db,
		Scanner: &RealScanner{},
	}
}

// ScanImage runs a vulnerability scan on the image using the configured scanner
func (d *Deployer) ScanImage(ctx context.Context, imageName string) (bool, string, error) {
	return d.Scanner.Scan(ctx, imageName)
}

// BuildAndPush triggers a local docker build and pushes to the mirrored registry
func (d *Deployer) BuildAndPush(ctx context.Context, deploy *models.Deployment) error {
	d.updateStatus(deploy, "BUILDING")

	// ... (build logic) ...
	imageName := fmt.Sprintf("localhost:5000/app:%s", deploy.CommitHash)

	// SECURITY SCAN
	d.updateStatus(deploy, "SCANNING")
	safe, report, err := d.ScanImage(ctx, imageName)
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
