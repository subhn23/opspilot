package ssh

import (
	"context"
	"testing"
)

func TestRealSSHClient_RunCommand_Failure(t *testing.T) {
	client := &RealSSHClient{
		User: "root",
		PrivateKey: "invalid-key",
	}
	
	_, err := client.RunCommand(context.Background(), "127.0.0.1:22", "ls")
	if err == nil {
		t.Error("Expected error for invalid private key, got nil")
	}
}
