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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("http://localhost:8081")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client.baseURL != "http://localhost:8081" {
		t.Errorf("expected baseURL to be http://localhost:8081, got %s", client.baseURL)
	}

	if client.version != VersionAuto {
		t.Errorf("expected version to be auto, got %s", client.version)
	}

	// Test invalid URL
	_, err = NewClient("invalid-url")
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	client, err := NewClient("http://localhost:8081",
		WithHTTPClient(httpClient),
		WithVersion(Version1_18to1_19),
		WithTimeout(5*time.Second),
		WithRetries(5, 2*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client.version != Version1_18to1_19 {
		t.Errorf("expected version to be 1.18-1.19, got %s", client.version)
	}

	if client.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected timeout to be 5s, got %v", client.httpClient.Timeout)
	}

	if client.maxRetries != 5 {
		t.Errorf("expected maxRetries to be 5, got %d", client.maxRetries)
	}

	if client.retryDelay != 2*time.Second {
		t.Errorf("expected retryDelay to be 2s, got %v", client.retryDelay)
	}
}

func TestContextCancellation(t *testing.T) {
	// Server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = client.ListJobs(ctx)
	if err == nil {
		t.Error("expected error due to context timeout, got nil")
	}
}