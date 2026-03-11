package main

import (
	"context"
	"fmt"
	"log"
	"opspilot/internal/metrics"

	"github.com/moby/moby/client"
)

func main() {
	// Initialize real Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	collector := &metrics.MetricCollector{
		Docker: cli,
	}

	fmt.Println("Scraping real Docker stats...")
	results, err := collector.Scrape(context.Background())
	if err != nil {
		log.Fatalf("Scrape failed: %v", err)
	}

	if len(results) == 0 {
		fmt.Println("No active containers found.")
		return
	}

	for _, m := range results {
		fmt.Printf("Container: %s (%s)\n", m.ContainerName, m.ContainerID[:12])
		fmt.Printf("  CPU Usage: %.2f%%\n", m.CPUUsage)
		fmt.Printf("  Mem Usage: %d bytes\n", m.MemoryUsage)
	}
}
