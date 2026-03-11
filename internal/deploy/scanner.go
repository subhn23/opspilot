package deploy

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

// Scanner abstracts the vulnerability scanning logic
type Scanner interface {
	Scan(ctx context.Context, imageName string) (bool, string, error)
}

var execCommand = exec.CommandContext

// RealScanner uses the Trivy binary to scan images
type RealScanner struct{}

func (s *RealScanner) Scan(ctx context.Context, imageName string) (bool, string, error) {
	log.Printf("Trivy: Scanning image %s", imageName)
	cmd := execCommand(ctx, "trivy", "image", "--severity", "CRITICAL,HIGH", "--exit-code", "1", imageName)
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
