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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the main Flink REST API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	version    Version
}

// Version represents a Flink version range
type Version string

const (
	// Version ranges - group versions with identical REST APIs
	Version1_8to1_12  Version = "1.8-1.12"   // Flink 1.8 through 1.12
	Version1_13to1_17 Version = "1.13-1.17"  // Flink 1.13 through 1.17
	Version1_18to1_19 Version = "1.18-1.19"  // Flink 1.18 through 1.19
	Version2_0Plus    Version = "2.0+"       // Flink 2.0 and above
	VersionAuto       Version = "auto"       // Auto-detect version
)

// NewClient creates a new Flink REST API client
func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		version: VersionAuto,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Option is a functional option for configuring the Client
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithVersion sets a specific Flink version range
func WithVersion(version Version) Option {
	return func(c *Client) {
		c.version = version
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// doRequest executes an HTTP request and handles common error cases
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// unmarshalResponse reads and unmarshals a JSON response
func unmarshalResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
