// Copyright 2025 Andrei Grigoriu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package restapi

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// GetJobMetrics retrieves metrics for a specific job
// Endpoint: GET /jobs/:jobid/metrics
// Available since: Flink 1.0
// You can filter metrics by providing metric names in the query parameter
func (c *Client) GetJobMetrics(ctx context.Context, jobID string, metricNames ...string) (*JobMetrics, error) {
	path := fmt.Sprintf("/jobs/%s/metrics", jobID)

	// Add metric filter if provided
	if len(metricNames) > 0 {
		path = fmt.Sprintf("%s?get=%s", path, strings.Join(metricNames, ","))
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics for job %s: %w", jobID, err)
	}

	var metricResp []Metric
	if err := unmarshalResponse(resp, &metricResp); err != nil {
		return nil, err
	}

	// Convert to JobMetrics
	metrics := &JobMetrics{
		JobID:   jobID,
		Metrics: make(map[string]float64),
	}

	for _, m := range metricResp {
		// Try to parse metric value as float
		if val, err := strconv.ParseFloat(m.Value, 64); err == nil {
			metrics.Metrics[m.ID] = val
		}
	}

	return metrics, nil
}

// GetVertexMetrics retrieves metrics for a specific job vertex (operator)
// Endpoint: GET /jobs/:jobid/vertices/:vertexid/metrics
// Available since: Flink 1.0
func (c *Client) GetVertexMetrics(ctx context.Context, jobID, vertexID string, metricNames ...string) (map[string]float64, error) {
	path := fmt.Sprintf("/jobs/%s/vertices/%s/metrics", jobID, vertexID)

	// Add metric filter if provided
	if len(metricNames) > 0 {
		path = fmt.Sprintf("%s?get=%s", path, strings.Join(metricNames, ","))
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics for vertex %s in job %s: %w", vertexID, jobID, err)
	}

	var metricResp []Metric
	if err := unmarshalResponse(resp, &metricResp); err != nil {
		return nil, err
	}

	// Convert to map
	metrics := make(map[string]float64)
	for _, m := range metricResp {
		if val, err := strconv.ParseFloat(m.Value, 64); err == nil {
			metrics[m.ID] = val
		}
	}

	return metrics, nil
}

// Common metric names for convenience
const (
	// Job-level metrics
	MetricNumRecordsIn          = "numRecordsIn"
	MetricNumRecordsOut         = "numRecordsOut"
	MetricNumBytesIn            = "numBytesIn"
	MetricNumBytesOut           = "numBytesOut"
	MetricBackPressuredTime     = "backPressuredTimeMsPerSecond"
	MetricIdleTime              = "idleTimeMsPerSecond"
	MetricBusyTime              = "busyTimeMsPerSecond"
	MetricLastCheckpointDuration = "lastCheckpointDuration"
	MetricLastCheckpointSize    = "lastCheckpointSize"

	// Vertex-level metrics
	MetricNumRecordsInPerSecond  = "numRecordsInPerSecond"
	MetricNumRecordsOutPerSecond = "numRecordsOutPerSecond"
)