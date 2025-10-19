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

// Package restapi provides a client for the Apache Flink REST API.
//
// Supported Flink versions: 1.18.x through 2.1.x (inclusive)
//
// Example usage:
//
//	client := restapi.NewClient("http://localhost:8081")
//	overview, err := client.GetClusterOverview(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Flink version: %s\n", overview.FlinkVersion)
//
// The client supports configurable timeouts, retry logic, and version detection.
// For production use with SSL/TLS, use WithTLSConfig option (planned feature).
package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// Client is the main Flink REST API client
type Client struct {
	baseURL         string
	httpClient      *http.Client
	version         Version
	detectedVersion *Version // Cached detected version
	maxRetries      int
	retryDelay      time.Duration
}

// Version represents a Flink version range
type Version string

const (
	// Supported version ranges (Flink 1.18 - 2.1)
	Version1_18to1_19 Version = "1.18-1.19" // Flink 1.18 through 1.19
	Version2_0Plus    Version = "2.0+"      // Flink 2.0 and above
	VersionAuto       Version = "auto"      // Auto-detect version
)

// NewClient creates a new Flink REST API client.
// Returns an error if the baseURL is invalid.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	// Trim trailing slash
	baseURL = strings.TrimRight(baseURL, "/")

	// Basic URL validation
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return nil, fmt.Errorf("invalid baseURL: must start with http:// or https://")
	}

	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		version:    VersionAuto,
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
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

// WithRetries sets the maximum number of retries and delay between retries.
// Default is 3 retries with 1 second delay. Uses exponential backoff.
func WithRetries(maxRetries int, retryDelay time.Duration) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
		c.retryDelay = retryDelay
	}
}

// Close closes the HTTP client and cleans up resources.
// It's safe to call Close multiple times.
func (c *Client) Close() error {
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// doRequest executes an HTTP request with retry logic and handles common error cases
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Add exponential backoff delay after first failure
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * c.retryDelay
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

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
			lastErr = fmt.Errorf("failed to execute request: %w", err)
			// Retry on network errors
			continue
		}

		// Handle HTTP errors
		if resp.StatusCode >= 400 {
			defer resp.Body.Close()
			bodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return nil, fmt.Errorf("HTTP %d (failed to read response body: %w)", resp.StatusCode, readErr)
			}

			// Retry on 5xx errors (server errors)
			if resp.StatusCode >= 500 {
				lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
				continue
			}

			// Don't retry on 4xx errors (client errors)
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		}

		// Success
		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
}

// unmarshalResponse reads and unmarshals a JSON response
func unmarshalResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
