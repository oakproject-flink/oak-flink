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
)

// ListJobs returns all jobs
// Endpoint: GET /jobs
// Available since: Flink 1.0
func (c *Client) ListJobs(ctx context.Context) ([]Job, error) {
	resp, err := c.doRequest(ctx, "GET", "/jobs", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	var overview JobsOverview
	if err := unmarshalResponse(resp, &overview); err != nil {
		return nil, err
	}

	return overview.Jobs, nil
}

// GetJob returns details for a specific job
// Endpoint: GET /jobs/:jobid
// Available since: Flink 1.0
func (c *Client) GetJob(ctx context.Context, jobID string) (*JobDetails, error) {
	path := fmt.Sprintf("/jobs/%s", jobID)

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get job %s: %w", jobID, err)
	}

	var details JobDetails
	if err := unmarshalResponse(resp, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

// CancelJob cancels a running job
// Endpoint: PATCH /jobs/:jobid
// Available since: Flink 1.0
func (c *Client) CancelJob(ctx context.Context, jobID string) error {
	path := fmt.Sprintf("/jobs/%s", jobID)

	_, err := c.doRequest(ctx, "PATCH", path, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel job %s: %w", jobID, err)
	}

	return nil
}

// GetJobConfig returns the configuration of a job
// Endpoint: GET /jobs/:jobid/config
// Available since: Flink 1.2
func (c *Client) GetJobConfig(ctx context.Context, jobID string) (*ConfigResponse, error) {
	path := fmt.Sprintf("/jobs/%s/config", jobID)

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get job config for %s: %w", jobID, err)
	}

	// The job config endpoint returns a different structure than cluster config
	// It wraps config in "execution-config" field as a map
	var jobConfig JobConfigResponse
	if err := unmarshalResponse(resp, &jobConfig); err != nil {
		return nil, err
	}

	// Convert the execution-config map to ConfigEntry array for consistency
	entries := make([]ConfigEntry, 0, len(jobConfig.ExecutionConfig))
	for key, value := range jobConfig.ExecutionConfig {
		entries = append(entries, ConfigEntry{
			Key:   key,
			Value: fmt.Sprintf("%v", value),
		})
	}

	return &ConfigResponse{Entries: entries}, nil
}