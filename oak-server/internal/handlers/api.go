package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/oakproject-flink/oak-flink/oak-server/web/templates/components"
)

// APIJobs returns jobs list as HTML for HTMX
func APIJobs(c echo.Context) error {
	// TODO: Fetch real jobs from services
	jobs := []map[string]interface{}{
		{
			"id":          "job-001",
			"name":        "User Analytics Pipeline",
			"status":      "running",
			"cluster":     "prod-cluster-1",
			"parallelism": 8,
			"uptime":      "2d 5h",
		},
		{
			"id":          "job-002",
			"name":        "Event Processing Stream",
			"status":      "running",
			"cluster":     "prod-cluster-1",
			"parallelism": 16,
			"uptime":      "1d 12h",
		},
		{
			"id":          "job-003",
			"name":        "Real-time Recommendations",
			"status":      "failing",
			"cluster":     "prod-cluster-2",
			"parallelism": 4,
			"uptime":      "0d 3h",
		},
	}

	return components.JobsTable(jobs).Render(c.Request().Context(), c.Response())
}

// MetricsStream sends real-time metrics via Server-Sent Events (SSE)
func MetricsStream(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// TODO: Fetch real metrics from services
			metrics := map[string]interface{}{
				"timestamp":   time.Now().Unix(),
				"cpuUsage":    rand.Float64() * 100,
				"memoryUsage": rand.Float64() * 100,
				"throughput":  rand.Intn(10000) + 5000,
				"backpressure": rand.Float64() * 10,
			}

			data, _ := json.Marshal(metrics)
			fmt.Fprintf(c.Response(), "data: %s\n\n", data)
			c.Response().Flush()

		case <-c.Request().Context().Done():
			return nil
		}
	}
}
