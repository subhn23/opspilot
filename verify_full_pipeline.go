package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"opspilot/internal/metrics"
	"time"

	"github.com/moby/moby/client"
)

func main() {
	// 1. Setup Mock VictoriaMetrics
	receivedData := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/write" {
			body, _ := io.ReadAll(r.Body)
			receivedData = string(body)
			w.WriteHeader(http.StatusNoContent)
		} else if r.URL.Path == "/api/v1/query_range" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[]}}`))
		}
	}))
	defer server.Close()

	// 2. Setup Collector
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	collector := &metrics.MetricCollector{
		Docker:             dockerCli,
		VictoriaMetricsURL: server.URL,
	}

	// 3. Run Pipeline
	fmt.Println("1. Scraping Docker stats...")
	results, err := collector.Scrape(context.Background())
	if err != nil {
		log.Fatalf("Scrape failed: %v", err)
	}
	fmt.Printf("   Found %d containers.\n", len(results))

	if len(results) > 0 {
		fmt.Println("2. Pushing to Mock VictoriaMetrics...")
		err = collector.Push(context.Background(), results)
		if err != nil {
			log.Fatalf("Push failed: %v", err)
		}
		fmt.Println("   Push successful.")
		fmt.Printf("   Data Received by Mock:\n%s", receivedData)
	}

	fmt.Println("3. Testing Query API...")
	query := "docker_metrics_cpu_usage"
	data, err := collector.QueryRange(context.Background(), query, time.Now().Add(-1*time.Hour), time.Now(), "1m")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	fmt.Println("   Query successful.")
	fmt.Printf("   Query Result: %s\n", data)

	fmt.Println("\nFull pipeline verification PASSED.")
}
