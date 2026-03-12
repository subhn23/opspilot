package proxy

import (
	"fmt"
	"net"
	"time"
)

// VerifyPorts checks if the application is correctly listening on the expected ports
func VerifyPorts(ports []string) error {
	for _, port := range ports {
		address := net.JoinHostPort("127.0.0.1", port)
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err != nil {
			return fmt.Errorf("port %s is not reachable: %w", port, err)
		}
		conn.Close()
	}
	return nil
}
