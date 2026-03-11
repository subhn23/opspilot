package metrics

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type MockDockerClient struct {
	Containers []container.Summary
	Stats      container.StatsResponse
}

func (m *MockDockerClient) ContainerList(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error) {
	return client.ContainerListResult{
		Items: m.Containers,
	}, nil
}

func (m *MockDockerClient) ContainerStats(ctx context.Context, containerID string, options client.ContainerStatsOptions) (client.ContainerStatsResult, error) {
	b, _ := json.Marshal(m.Stats)
	return client.ContainerStatsResult{
		Body: io.NopCloser(strings.NewReader(string(b))),
	}, nil
}

func TestMetricCollector_Scrape(t *testing.T) {
	mockDocker := &MockDockerClient{
		Containers: []container.Summary{
			{ID: "test-id", Names: []string{"/test-container"}},
		},
		Stats: container.StatsResponse{
			CPUStats: container.CPUStats{
				CPUUsage: container.CPUUsage{
					TotalUsage: 200,
				},
				SystemUsage: 2000,
				OnlineCPUs:  1,
			},
			PreCPUStats: container.CPUStats{
				CPUUsage: container.CPUUsage{
					TotalUsage: 100,
				},
				SystemUsage: 1000,
			},
			MemoryStats: container.MemoryStats{
				Usage: 512,
			},
		},
	}

	collector := &MetricCollector{Docker: mockDocker}
	metrics, err := collector.Scrape(context.Background())
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	if len(metrics) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(metrics))
	}

	m := metrics[0]
	if m.ContainerID != "test-id" {
		t.Errorf("Expected ContainerID test-id, got %s", m.ContainerID)
	}

	// (200-100)/(2000-1000) * 1 * 100 = 10%
	if m.CPUUsage != 10.0 {
		t.Errorf("Expected CPUUsage 10.0, got %f", m.CPUUsage)
	}

	if m.MemoryUsage != 512 {
		t.Errorf("Expected MemoryUsage 512, got %d", m.MemoryUsage)
	}
}

func TestMetricCollector_Push(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/write" {
			t.Errorf("Expected /write path, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	collector := &MetricCollector{
		VictoriaMetricsURL: server.URL,
	}

	metrics := []Metric{
		{
			ContainerID:   "test-id",
			ContainerName: "test-container",
			CPUUsage:      10.5,
			MemoryUsage:   1024,
			Timestamp:     time.Now(),
		},
	}

	err := collector.Push(context.Background(), metrics)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}
}
