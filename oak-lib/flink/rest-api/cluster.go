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
// Endpoint: GET /jobmanager/config
// Available since: Flink 1.2
func (c *Client) GetConfig(ctx context.Context) (*ConfigResponse, error) {
	resp, err := c.doRequest(ctx, "GET", "/jobmanager/config", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster config: %w", err)
	}

	var entries []ConfigEntry
	if err := unmarshalResponse(resp, &entries); err != nil {
		return nil, err
	}

	return &ConfigResponse{Entries: entries}, nil
}

// parseVersion parses a semantic version string and returns major, minor
func parseVersion(version string) (major, minor int, err error) {
	// Remove any prefix like "v" if present
	version = strings.TrimPrefix(version, "v")

	// Split by dots and parse
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid version format: %s", version)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minor version: %w", err)
	}

	return major, minor, nil
}

// DetectVersion attempts to auto-detect the Flink version from the cluster.
// The detected version is cached to avoid redundant API calls.
// Supported versions: Flink 1.18 through 2.1 (inclusive)
func (c *Client) DetectVersion(ctx context.Context) (Version, error) {
	// Return cached version if available
	if c.detectedVersion != nil {
		return *c.detectedVersion, nil
	}

	overview, err := c.GetClusterOverview(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to detect version: %w", err)
	}

	version := overview.FlinkVersion
	major, minor, err := parseVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to parse Flink version %s: %w", version, err)
	}

	// Check if version is in supported range [1.18, 2.1]
	// Version < 1.18: Not supported
	if major < 1 || (major == 1 && minor < 18) {
		return "", fmt.Errorf("Flink version %s is not supported (minimum version: 1.18)", version)
	}

	// Version > 2.1: Not supported
	if major > 2 || (major == 2 && minor > 1) {
		return "", fmt.Errorf("Flink version %s is not supported (maximum version: 2.1)", version)
	}

	// Version is in supported range [1.18, 2.1]
	// Map to appropriate version constant and cache it
	// Note: Flink 1.20+ uses the same REST API as 2.0+
	var detected Version
	if major >= 2 || (major == 1 && minor >= 20) {
		detected = Version2_0Plus
	} else {
		detected = Version1_18to1_19
	}

	c.detectedVersion = &detected
	return detected, nil
}