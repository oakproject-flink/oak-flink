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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// TriggerSavepoint triggers a savepoint for a job
// Endpoint: POST /jobs/:jobid/savepoints
// Available since: Flink 1.2
func (c *Client) TriggerSavepoint(ctx context.Context, jobID string, req SavepointTriggerRequest) (*SavepointTriggerResponse, error) {
	path := fmt.Sprintf("/jobs/%s/savepoints", jobID)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal savepoint request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to trigger savepoint for job %s: %w", jobID, err)
	}

	var savepointResp SavepointTriggerResponse
	if err := unmarshalResponse(resp, &savepointResp); err != nil {
		return nil, err
	}

	return &savepointResp, nil
}

// GetSavepointStatus retrieves the status of a savepoint operation
// Endpoint: GET /jobs/:jobid/savepoints/:triggerid
// Available since: Flink 1.2
func (c *Client) GetSavepointStatus(ctx context.Context, jobID, triggerID string) (*SavepointStatus, error) {
	path := fmt.Sprintf("/jobs/%s/savepoints/%s", jobID, triggerID)

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get savepoint status for job %s, trigger %s: %w", jobID, triggerID, err)
	}

	var status SavepointStatus
	if err := unmarshalResponse(resp, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// StopJobWithSavepoint stops a job with a savepoint
// This is a version-specific implementation
// Endpoint: POST /jobs/:jobid/stop (Flink 1.11+)
func (c *Client) StopJobWithSavepoint(ctx context.Context, jobID string, targetDirectory string) (*SavepointTriggerResponse, error) {
	// Use different endpoints based on version
	switch c.version {
	case Version1_8to1_12:
		// Older versions: use trigger savepoint with cancel flag
		return c.TriggerSavepoint(ctx, jobID, SavepointTriggerRequest{
			TargetDirectory: targetDirectory,
			CancelJob:       true,
		})
	default:
		// Flink 1.11+ has dedicated stop endpoint
		path := fmt.Sprintf("/jobs/%s/stop", jobID)

		req := struct {
			TargetDirectory string `json:"targetDirectory"`
			Drain           bool   `json:"drain"`
		}{
			TargetDirectory: targetDirectory,
			Drain:           false,
		}

		body, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal stop request: %w", err)
		}

		resp, err := c.doRequest(ctx, "POST", path, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to stop job %s with savepoint: %w", jobID, err)
		}

		var savepointResp SavepointTriggerResponse
		if err := unmarshalResponse(resp, &savepointResp); err != nil {
			return nil, err
		}

		return &savepointResp, nil
	}
}
