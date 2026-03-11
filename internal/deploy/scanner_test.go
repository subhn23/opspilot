package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

// TestProcess is used to mock the execution of the trivy command
func TestProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// The mock behavior depends on the arguments or environment variables
	arg := os.Getenv("MOCK_BEHAVIOR")
	switch arg {
	case "SAFE":
		fmt.Print("No vulnerabilities found")
		os.Exit(0)
	case "UNSAFE":
		fmt.Print("CRITICAL vulnerability found")
		os.Exit(1)
	default:
		fmt.Print("Command failed")
		os.Exit(2)
	}
}

func fakeExecCommand(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "MOCK_BEHAVIOR=" + os.Getenv("MOCK_BEHAVIOR")}
	return cmd
}

func TestRealScanner_Scan(t *testing.T) {
	oldExec := execCommand
	defer func() { execCommand = oldExec }()
	execCommand = fakeExecCommand

	scanner := &RealScanner{}
	ctx := context.Background()

	t.Run("Safe Image", func(t *testing.T) {
		os.Setenv("MOCK_BEHAVIOR", "SAFE")
		safe, report, err := scanner.Scan(ctx, "safe-image")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !safe {
			t.Error("Expected image to be safe")
		}
		if report != "No critical or high vulnerabilities found." {
			t.Errorf("Expected report 'No critical or high vulnerabilities found.', got %s", report)
		}
	})

	t.Run("Unsafe Image", func(t *testing.T) {
		os.Setenv("MOCK_BEHAVIOR", "UNSAFE")
		safe, report, err := scanner.Scan(ctx, "unsafe-image")
		if err != nil {
			t.Errorf("Expected no error from scanner, got %v", err)
		}
		if safe {
			t.Error("Expected image to be unsafe")
		}
		if report != "CRITICAL vulnerability found" {
			t.Errorf("Expected report 'CRITICAL vulnerability found', got %s", report)
		}
	})

	t.Run("Error Execution", func(t *testing.T) {
		os.Setenv("MOCK_BEHAVIOR", "FAIL")
		safe, _, err := scanner.Scan(ctx, "error-image")
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if safe {
			t.Error("Expected safe to be false")
		}
	})
}
