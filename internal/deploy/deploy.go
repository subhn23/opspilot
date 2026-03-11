package deploy

import (
	"context"
	"fmt"
	"log"
	"opspilot/internal/auth"
	"opspilot/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SSHClient interface {
	RunCommand(ctx context.Context, addr, command string) (string, error)
}

// RealSSHClient uses golang.org/x/crypto/ssh to execute remote commands
type RealSSHClient struct {
	User       string
	PrivateKey string
}

func (s *RealSSHClient) RunCommand(ctx context.Context, addr, command string) (string, error) {
	// Placeholder for actual SSH logic using golang.org/x/crypto/ssh
	return "", fmt.Errorf("SSH execution not yet fully implemented")
}

type Deployer struct {
	DB      *gorm.DB
	Scanner Scanner
	SSH     SSHClient
}

func NewDeployer(db *gorm.DB) *Deployer {
	return &Deployer{
		DB:      db,
		Scanner: &RealScanner{},
		SSH:     &RealSSHClient{User: "root"},
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
		auth.LogAction(d.DB, uuid.Nil, "SECURITY_FAILURE", deploy.CommitHash, imageName, report)
		return fmt.Errorf("security scan failed: %s", report)
	}

	// ... (push logic) ...
	d.updateStatus(deploy, "PUSHING")
	log.Printf("Pushing image %s to registry", imageName)
	deploy.Logs += fmt.Sprintf("\n$ docker push %s\n(Mocked push success)", imageName)

	d.updateStatus(deploy, "PUSHED")
	return nil
}

// RemoteUp SSHs into the dynamic VM and runs docker-compose up
func (d *Deployer) RemoteUp(ctx context.Context, deploy *models.Deployment, targetIP string) error {
	d.updateStatus(deploy, "DEPLOYING")

	imageName := fmt.Sprintf("localhost:5000/app:%s", deploy.CommitHash)

	// Command sequence
	commands := []string{
		fmt.Sprintf("docker pull %s", imageName),
		"docker-compose up -d",
	}

	log.Printf("Executing remote deploy to %s", targetIP)

	for _, cmd := range commands {
		output, err := d.SSH.RunCommand(ctx, targetIP+":22", cmd)
		deploy.Logs += fmt.Sprintf("\n$ %s\n%s", cmd, output)
		if err != nil {
			d.updateStatus(deploy, "FAILED_DEPLOY")
			d.DB.Save(deploy)
			return fmt.Errorf("remote command failed: %s: %w", cmd, err)
		}
	}

	d.updateStatus(deploy, "SUCCESS")

	// Audit Log
	auth.LogAction(d.DB, uuid.Nil, "DEPLOY_SUCCESS", deploy.CommitHash, targetIP, "Remote deployment successful")

	return nil
}

func (d *Deployer) updateStatus(deploy *models.Deployment, status string) {
	deploy.Status = status
	d.DB.Save(deploy)
}
