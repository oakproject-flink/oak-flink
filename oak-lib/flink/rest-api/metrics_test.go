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

func TestGetJobMetrics(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		responseBody   string
		responseStatus int
		wantErr        bool
		wantMetrics    int
	}{
		{
			name:  "successful response with metrics",
			jobID: "test-job-id",
			responseBody: `[
				{"id": "numRecordsIn", "value": "12345"},
				{"id": "numRecordsOut", "value": "12340"},
				{"id": "numBytesIn", "value": "1024000"}
			]`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantMetrics:    3,
		},
		{
			name:           "no metrics",
			jobID:          "test-job-id",
			responseBody:   `[]`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantMetrics:    0,
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
				expectedPath := "/jobs/" + tt.jobID + "/metrics"
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

			metrics, err := client.GetJobMetrics(ctx, tt.jobID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetJobMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if metrics.JobID != tt.jobID {
					t.Errorf("job ID = %s, want %s", metrics.JobID, tt.jobID)
				}
				if len(metrics.Metrics) != tt.wantMetrics {
					t.Errorf("got %d metrics, want %d", len(metrics.Metrics), tt.wantMetrics)
				}
			}
		})
	}
}

func TestGetVertexMetrics(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		vertexID       string
		responseBody   string
		responseStatus int
		wantErr        bool
		wantMetrics    int
	}{
		{
			name:     "successful response",
			jobID:    "test-job-id",
			vertexID: "vertex-1",
			responseBody: `[
				{"id": "backPressuredTime", "value": "100"},
				{"id": "busyTime", "value": "5000"}
			]`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantMetrics:    2,
		},
		{
			name:           "vertex not found",
			jobID:          "test-job-id",
			vertexID:       "nonexistent",
			responseBody:   `{"errors": ["Vertex not found"]}`,
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/jobs/" + tt.jobID + "/vertices/" + tt.vertexID + "/metrics"
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

			metrics, err := client.GetVertexMetrics(ctx, tt.jobID, tt.vertexID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetVertexMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(metrics) != tt.wantMetrics {
				t.Errorf("got %d metrics, want %d", len(metrics), tt.wantMetrics)
			}
		})
	}
}
