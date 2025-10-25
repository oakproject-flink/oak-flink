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
)

func TestListJobs(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		wantErr        bool
		wantJobCount   int
	}{
		{
			name: "multiple jobs",
			responseBody: `{
				"jobs": [
					{"id": "job1", "status": "RUNNING"},
					{"id": "job2", "status": "FINISHED"}
				]
			}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantJobCount:   2,
		},
		{
			name:           "no jobs",
			responseBody:   `{"jobs": []}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantJobCount:   0,
		},
		{
			name:           "server error",
			responseBody:   `{"error": "internal error"}`,
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/jobs" {
					t.Errorf("expected path /jobs, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatalf("NewClient failed: %v", err)
			}
			defer client.Close()

			ctx := context.Background()

			jobs, err := client.ListJobs(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListJobs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(jobs) != tt.wantJobCount {
				t.Errorf("got %d jobs, want %d", len(jobs), tt.wantJobCount)
			}
		})
	}
}

func TestGetJob(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		responseBody   string
		responseStatus int
		wantErr        bool
		wantVertices   int
	}{
		{
			name:  "successful response",
			jobID: "test-job-id",
			responseBody: `{
				"jid": "test-job-id",
				"name": "Test Job",
				"state": "RUNNING",
				"vertices": [
					{"id": "v1", "name": "Source", "parallelism": 2},
					{"id": "v2", "name": "Map", "parallelism": 4}
				],
				"plan": {"nodes": []}
			}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantVertices:   2,
		},
		{
			name:           "job not found",
			jobID:          "nonexistent",
			responseBody:   `{"errors": ["Job not found"]}`,
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/jobs/" + tt.jobID
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatalf("NewClient failed: %v", err)
			}
			defer client.Close()

			ctx := context.Background()

			job, err := client.GetJob(ctx, tt.jobID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if job.ID != tt.jobID {
					t.Errorf("job ID = %s, want %s", job.ID, tt.jobID)
				}
				if len(job.Vertices) != tt.wantVertices {
					t.Errorf("got %d vertices, want %d", len(job.Vertices), tt.wantVertices)
				}
			}
		})
	}
}

func TestCancelJob(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "successful cancel",
			jobID:          "test-job-id",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "job not found",
			jobID:          "nonexistent",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/jobs/" + tt.jobID
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH method, got %s", r.Method)
				}
				w.WriteHeader(tt.responseStatus)
			}))
			defer server.Close()

			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatalf("NewClient failed: %v", err)
			}
			defer client.Close()

			ctx := context.Background()

			err = client.CancelJob(ctx, tt.jobID)

			if (err != nil) != tt.wantErr {
				t.Errorf("CancelJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetJobConfig(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		responseBody   string
		responseStatus int
		wantErr        bool
		wantEntries    int
	}{
		{
			name:  "successful response",
			jobID: "test-job-id",
			responseBody: `{
				"jid": "test-job-id",
				"name": "Test Job",
				"execution-config": {
					"execution-mode": "PIPELINED",
					"restart-strategy": "Cluster Level Restart Strategy"
				}
			}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantEntries:    2,
		},
		{
			name:           "job not found",
			jobID:          "nonexistent",
			responseBody:   `{"errors": ["Job not found"]}`,
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/jobs/" + tt.jobID + "/config"
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatalf("NewClient failed: %v", err)
			}
			defer client.Close()

			ctx := context.Background()

			config, err := client.GetJobConfig(ctx, tt.jobID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetJobConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(config.Entries) != tt.wantEntries {
				t.Errorf("got %d entries, want %d", len(config.Entries), tt.wantEntries)
			}
		})
	}
}
