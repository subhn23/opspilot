package deploy

import (
	"context"
	"fmt"
	"log"
	"opspilot/internal/audit"
	"opspilot/internal/models"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

type SSHClient interface {
	RunCommand(ctx context.Context, addr, command string) (string, error)
}

type GitClient interface {
	Clone(ctx context.Context, repoURL, targetDir string) error
	Checkout(ctx context.Context, targetDir, commitHash string) error
}

type DockerClient interface {
	Login(ctx context.Context, user, pass, registry string) error
	Build(ctx context.Context, workingDir, tag string) error
	Push(ctx context.Context, tag string) error
}

type RealGitClient struct{}

func (g *RealGitClient) Clone(ctx context.Context, repoURL, targetDir string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, targetDir)
	return cmd.Run()
}

func (g *RealGitClient) Checkout(ctx context.Context, targetDir, commitHash string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", targetDir, "checkout", commitHash)
	return cmd.Run()
}

type RealDockerClient struct{}

func (d *RealDockerClient) Login(ctx context.Context, user, pass, registry string) error {
	cmd := exec.CommandContext(ctx, "docker", "login", "-u", user, "-p", pass, registry)
	return cmd.Run()
}

func (d *RealDockerClient) Build(ctx context.Context, workingDir, tag string) error {
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", tag, workingDir)
	return cmd.Run()
}

func (d *RealDockerClient) Push(ctx context.Context, tag string) error {
	cmd := exec.CommandContext(ctx, "docker", "push", tag)
	return cmd.Run()
}

// RealSSHClient uses golang.org/x/crypto/ssh to execute remote commands
type RealSSHClient struct {
	User       string
	PrivateKey string
}

func (s *RealSSHClient) RunCommand(ctx context.Context, addr, command string) (string, error) {
	if s.PrivateKey == "" {
		return "", fmt.Errorf("SSH private key not provided")
	}

	signer, err := ssh.ParsePrivateKey([]byte(s.PrivateKey))
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For dynamic VMs, ideally we'd use a known hosts file
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("failed to dial SSH: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("failed to run command: %w", err)
	}

	return string(output), nil
}

type Deployer struct {
	DB      *gorm.DB
	Scanner Scanner
	SSH     SSHClient
	Git     GitClient
	Docker  DockerClient
	RepoURL string
}

func NewDeployer(db *gorm.DB) *Deployer {
	return &Deployer{
		DB:      db,
		Scanner: &RealScanner{},
		SSH: &RealSSHClient{
			User:       getEnv("SSH_USER", "root"),
			PrivateKey: os.Getenv("SSH_PRIVATE_KEY"),
		},
		Git:     &RealGitClient{},
		Docker:  &RealDockerClient{},
		RepoURL: os.Getenv("PROJECT_REPO_URL"),
	}
}

// ScanImage runs a vulnerability scan on the image using the configured scanner
func (d *Deployer) ScanImage(ctx context.Context, imageName string) (bool, string, error) {
	return d.Scanner.Scan(ctx, imageName)
}

// BuildAndPush triggers a local docker build and pushes to the mirrored registry
func (d *Deployer) BuildAndPush(ctx context.Context, deploy *models.Deployment) error {
	d.updateStatus(deploy, "BUILDING")

	// 1. Clone & Checkout
	tmpDir, err := os.MkdirTemp("", "opspilot-build-*")
	if err != nil {
		d.updateStatus(deploy, "FAILED_BUILD")
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := d.Git.Clone(ctx, d.RepoURL, tmpDir); err != nil {
		d.updateStatus(deploy, "FAILED_BUILD")
		return fmt.Errorf("git clone failed: %w", err)
	}

	if err := d.Git.Checkout(ctx, tmpDir, deploy.CommitHash); err != nil {
		d.updateStatus(deploy, "FAILED_BUILD")
		return fmt.Errorf("git checkout failed: %w", err)
	}

	// 2. Docker Login
	registry := os.Getenv("REGISTRY_URL")
	user := os.Getenv("REGISTRY_USER")
	pass := os.Getenv("REGISTRY_PASS")
	if registry != "" && user != "" && pass != "" {
		if err := d.Docker.Login(ctx, user, pass, registry); err != nil {
			d.updateStatus(deploy, "FAILED_BUILD")
			return fmt.Errorf("docker login failed: %w", err)
		}
	}

	// 3. Build
	imageName := fmt.Sprintf("%s/app:%s", registry, deploy.CommitHash)
	if registry == "" {
		imageName = fmt.Sprintf("localhost:5000/app:%s", deploy.CommitHash)
	}

	if err := d.Docker.Build(ctx, tmpDir, imageName); err != nil {
		d.updateStatus(deploy, "FAILED_BUILD")
		return fmt.Errorf("docker build failed: %w", err)
	}

	// SECURITY SCAN
	d.updateStatus(deploy, "SCANNING")
	safe, report, err := d.ScanImage(ctx, imageName)
	if err != nil || !safe {
		d.updateStatus(deploy, "FAILED_SECURITY")
		deploy.Logs += "\nSECURITY ALERT: " + report
		d.DB.Save(deploy)
		audit.LogAction(d.DB, uuid.Nil, "SECURITY_FAILURE", deploy.CommitHash, imageName, report)
		return fmt.Errorf("security scan failed: %s", report)
	}

	// 4. Push
	d.updateStatus(deploy, "PUSHING")
	if err := d.Docker.Push(ctx, imageName); err != nil {
		d.updateStatus(deploy, "FAILED_PUSH")
		return fmt.Errorf("docker push failed: %w", err)
	}

	d.updateStatus(deploy, "PUSHED")
	return nil
}

// RemoteUp SSHs into the dynamic VM and runs docker-compose up
func (d *Deployer) RemoteUp(ctx context.Context, deploy *models.Deployment, targetIP string) error {
	d.updateStatus(deploy, "DEPLOYING")

	registry := os.Getenv("REGISTRY_URL")
	imageName := fmt.Sprintf("%s/app:%s", registry, deploy.CommitHash)
	if registry == "" {
		imageName = fmt.Sprintf("localhost:5000/app:%s", deploy.CommitHash)
	}

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
	audit.LogAction(d.DB, uuid.Nil, "DEPLOY_SUCCESS", deploy.CommitHash, targetIP, "Remote deployment successful")

	return nil
}

func (d *Deployer) updateStatus(deploy *models.Deployment, status string) {
	deploy.Status = status
	d.DB.Save(deploy)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
