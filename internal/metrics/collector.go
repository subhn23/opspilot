package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type Metric struct {
	ContainerID   string
	ContainerName string
	CPUUsage      float64
	MemoryUsage   uint64
	Timestamp     time.Time
}

type DockerClient interface {
	ContainerList(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error)
	ContainerStats(ctx context.Context, containerID string, options client.ContainerStatsOptions) (client.ContainerStatsResult, error)
}

type MetricCollector struct {
	Docker             DockerClient
	VictoriaMetricsURL string
	HTTPClient         *http.Client
}

func (m *MetricCollector) Scrape(ctx context.Context) ([]Metric, error) {
	resp, err := m.Docker.ContainerList(ctx, client.ContainerListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var metrics []Metric
	for _, c := range resp.Items {
		result, err := m.Docker.ContainerStats(ctx, c.ID, client.ContainerStatsOptions{Stream: false})
		if err != nil {
			continue
		}

		var statsJSON container.StatsResponse
		if err := json.NewDecoder(result.Body).Decode(&statsJSON); err != nil {
			result.Body.Close()
			continue
		}
		result.Body.Close()

		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
		}

		metrics = append(metrics, Metric{
			ContainerID:   c.ID,
			ContainerName: name,
			CPUUsage:      calculateCPUPercent(&statsJSON),
			MemoryUsage:   statsJSON.MemoryStats.Usage,
			Timestamp:     time.Now(),
		})
	}

	return metrics, nil
}

func (m *MetricCollector) Push(ctx context.Context, metrics []Metric) error {
	if m.VictoriaMetricsURL == "" {
		return fmt.Errorf("VictoriaMetrics URL not set")
	}

	if m.HTTPClient == nil {
		m.HTTPClient = &http.Client{Timeout: 5 * time.Second}
	}

	var buf bytes.Buffer
	for _, met := range metrics {
		// InfluxDB Line Protocol: measurement,tags fields timestamp
		// docker_metrics,container_id=xxx,container_name=yyy cpu_usage=10.5,memory_usage=1024 timestamp_ns
		line := fmt.Sprintf("docker_metrics,container_id=%s,container_name=%s cpu_usage=%f,memory_usage=%d %d\n",
			met.ContainerID, met.ContainerName, met.CPUUsage, met.MemoryUsage, met.Timestamp.UnixNano())
		buf.WriteString(line)
	}

	resp, err := m.HTTPClient.Post(m.VictoriaMetricsURL+"/write", "text/plain", &buf)
	if err != nil {
		return fmt.Errorf("failed to push to VictoriaMetrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("VictoriaMetrics returned error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func calculateCPUPercent(v *container.StatsResponse) float64 {
	cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage) - float64(v.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(v.CPUStats.SystemUsage) - float64(v.PreCPUStats.SystemUsage)

	onlineCPUs := float64(v.CPUStats.OnlineCPUs)
	if onlineCPUs == 0 {
		onlineCPUs = 1.0 // Fallback
	}

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		percent := (cpuDelta / systemDelta) * onlineCPUs * 100.0
		return percent
	}
	return 0.0
}
