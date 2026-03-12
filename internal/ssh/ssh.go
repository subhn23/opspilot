package ssh

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClient interface defines the methods required for remote execution
type Client interface {
	RunCommand(ctx context.Context, addr, command string) (string, error)
	Configure(user, privateKey string)
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
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
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

func (s *RealSSHClient) Configure(user, privateKey string) {
	s.User = user
	s.PrivateKey = privateKey
}
