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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8081")

	if client.baseURL != "http://localhost:8081" {
		t.Errorf("expected baseURL to be http://localhost:8081, got %s", client.baseURL)
	}

	if client.version != VersionAuto {
		t.Errorf("expected version to be auto, got %s", client.version)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	client := NewClient("http://localhost:8081",
		WithHTTPClient(httpClient),
		WithVersion(Version1_18to1_19),
		WithTimeout(5*time.Second),
	)

	if client.version != Version1_18to1_19 {
		t.Errorf("expected version to be 1.18-1.19, got %s", client.version)
	}

	if client.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected timeout to be 5s, got %v", client.httpClient.Timeout)
	}
}

func TestListJobs(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jobs" {
			t.Errorf("expected path /jobs, got %s", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		response := JobsOverview{
			Jobs: []Job{
				{
					ID:        "job-123",
					Name:      "Test Job",
					Status:    JobStatusRunning,
					StartTime: 1234567890,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	jobs, err := client.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	if jobs[0].ID != "job-123" {
		t.Errorf("expected job ID 'job-123', got %s", jobs[0].ID)
	}

	if jobs[0].Status != JobStatusRunning {
		t.Errorf("expected status RUNNING, got %s", jobs[0].Status)
	}
}

func TestGetClusterOverview(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ClusterOverview{
			TaskManagers:   3,
			SlotsTotal:     12,
			SlotsAvailable: 4,
			JobsRunning:    2,
			FlinkVersion:   "1.18.0",
			FlinkCommit:    "abc123",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	overview, err := client.GetClusterOverview(ctx)
	if err != nil {
		t.Fatalf("GetClusterOverview failed: %v", err)
	}

	if overview.TaskManagers != 3 {
		t.Errorf("expected 3 task managers, got %d", overview.TaskManagers)
	}

	if overview.FlinkVersion != "1.18.0" {
		t.Errorf("expected Flink version 1.18.0, got %s", overview.FlinkVersion)
	}
}

func TestTriggerSavepoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		var req SavepointTriggerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.TargetDirectory != "/savepoints" {
			t.Errorf("expected target directory /savepoints, got %s", req.TargetDirectory)
		}

		response := SavepointTriggerResponse{
			RequestID: "trigger-123",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	resp, err := client.TriggerSavepoint(ctx, "job-123", SavepointTriggerRequest{
		TargetDirectory: "/savepoints",
		CancelJob:       false,
	})

	if err != nil {
		t.Fatalf("TriggerSavepoint failed: %v", err)
	}

	if resp.RequestID != "trigger-123" {
		t.Errorf("expected request ID 'trigger-123', got %s", resp.RequestID)
	}
}

func TestContextCancellation(t *testing.T) {
	// Server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.ListJobs(ctx)
	if err == nil {
		t.Error("expected error due to context timeout, got nil")
	}
}