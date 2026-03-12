package proxy

import (
	"net"
	"testing"
)

func TestVerifyPorts(t *testing.T) {
	// Start a dummy listener
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer ln.Close()

	_, port, _ := net.SplitHostPort(ln.Addr().String())

	t.Run("Reachable", func(t *testing.T) {
		err := VerifyPorts([]string{port})
		if err != nil {
			t.Errorf("Expected port %s to be reachable, got error: %v", port, err)
		}
	})

	t.Run("Unreachable", func(t *testing.T) {
		err := VerifyPorts([]string{"9999"}) // Assume 9999 is free
		if err == nil {
			t.Error("Expected error for unreachable port, got nil")
		}
	})
}
