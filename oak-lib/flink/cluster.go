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

package flink

import (
	"context"
	"fmt"
)

// GetClusterOverview returns overview information about the Flink cluster
// Endpoint: GET /overview
// Available since: Flink 1.0
func (c *Client) GetClusterOverview(ctx context.Context) (*ClusterOverview, error) {
	resp, err := c.doRequest(ctx, "GET", "/overview", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster overview: %w", err)
	}

	var overview ClusterOverview
	if err := unmarshalResponse(resp, &overview); err != nil {
		return nil, err
	}

	return &overview, nil
}

// GetConfig returns the Flink cluster configuration
// Endpoint: GET /config
// Available since: Flink 1.2
func (c *Client) GetConfig(ctx context.Context) (*ConfigResponse, error) {
	resp, err := c.doRequest(ctx, "GET", "/config", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster config: %w", err)
	}

	var config ConfigResponse
	if err := unmarshalResponse(resp, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DetectVersion attempts to auto-detect the Flink version from the cluster
func (c *Client) DetectVersion(ctx context.Context) (Version, error) {
	overview, err := c.GetClusterOverview(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to detect version: %w", err)
	}

	// Parse version string and map to version range
	version := overview.FlinkVersion

	// Simple version mapping - can be enhanced
	switch {
	case version >= "2.0.0":
		return Version2_0Plus, nil
	case version >= "1.18.0":
		return Version1_18to1_19, nil
	case version >= "1.13.0":
		return Version1_13to1_17, nil
	case version >= "1.8.0":
		return Version1_8to1_12, nil
	default:
		return Version1_8to1_12, nil // fallback
	}
}