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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTriggerSavepoint(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		request        SavepointTriggerRequest
		responseBody   string
		responseStatus int
		wantErr        bool
		wantRequestID  string
	}{
		{
			name:  "successful savepoint trigger",
			jobID: "test-job-id",
			request: SavepointTriggerRequest{
				TargetDirectory: "/tmp/savepoints",
				CancelJob:       false,
			},
			responseBody:   `{"request-id": "sp-123"}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantRequestID:  "sp-123",
		},
		{
			name:  "savepoint with cancel",
			jobID: "test-job-id",
			request: SavepointTriggerRequest{
				TargetDirectory: "/tmp/savepoints",
				CancelJob:       true,
			},
			responseBody:   `{"request-id": "sp-456"}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantRequestID:  "sp-456",
		},
		{
			name:           "job not found",
			jobID:          "nonexistent",
			request:        SavepointTriggerRequest{},
			responseBody:   `{"errors": ["Job not found"]}`,
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/jobs/" + tt.jobID + "/savepoints"
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("expected POST method, got %s", r.Method)
				}

				// Verify request body
				body, _ := io.ReadAll(r.Body)
				var req SavepointTriggerRequest
				if err := json.Unmarshal(body, &req); err == nil {
					if req.TargetDirectory != tt.request.TargetDirectory {
						t.Errorf("target directory = %s, want %s", req.TargetDirectory, tt.request.TargetDirectory)
					}
					if req.CancelJob != tt.request.CancelJob {
						t.Errorf("cancel job = %v, want %v", req.CancelJob, tt.request.CancelJob)
					}
				}

				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			ctx := context.Background()

			resp, err := client.TriggerSavepoint(ctx, tt.jobID, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("TriggerSavepoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp.RequestID != tt.wantRequestID {
				t.Errorf("request ID = %s, want %s", resp.RequestID, tt.wantRequestID)
			}
		})
	}
}

func TestGetSavepointStatus(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		triggerID      string
		responseBody   string
		responseStatus int
		wantErr        bool
		wantStatusID   string
	}{
		{
			name:      "savepoint in progress",
			jobID:     "test-job-id",
			triggerID: "sp-123",
			responseBody: `{
				"status": {"id": "IN_PROGRESS"},
				"operation": {"location": ""}
			}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantStatusID:   "IN_PROGRESS",
		},
		{
			name:      "savepoint completed",
			jobID:     "test-job-id",
			triggerID: "sp-123",
			responseBody: `{
				"status": {"id": "COMPLETED"},
				"operation": {"location": "/tmp/savepoints/savepoint-123"}
			}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantStatusID:   "COMPLETED",
		},
		{
			name:           "trigger not found",
			jobID:          "test-job-id",
			triggerID:      "nonexistent",
			responseBody:   `{"errors": ["Trigger not found"]}`,
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/jobs/" + tt.jobID + "/savepoints/" + tt.triggerID
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			ctx := context.Background()

			status, err := client.GetSavepointStatus(ctx, tt.jobID, tt.triggerID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSavepointStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && status.Status.ID != tt.wantStatusID {
				t.Errorf("status ID = %s, want %s", status.Status.ID, tt.wantStatusID)
			}
		})
	}
}

func TestStopJobWithSavepoint(t *testing.T) {
	tests := []struct {
		name            string
		jobID           string
		targetDirectory string
		responseBody    string
		responseStatus  int
		wantErr         bool
		wantRequestID   string
	}{
		{
			name:            "successful stop with savepoint",
			jobID:           "test-job-id",
			targetDirectory: "/tmp/savepoints",
			responseBody:    `{"request-id": "sp-789"}`,
			responseStatus:  http.StatusOK,
			wantErr:         false,
			wantRequestID:   "sp-789",
		},
		{
			name:            "job not found",
			jobID:           "nonexistent",
			targetDirectory: "/tmp/savepoints",
			responseBody:    `{"errors": ["Job not found"]}`,
			responseStatus:  http.StatusNotFound,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/jobs/" + tt.jobID + "/stop"
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("expected POST method, got %s", r.Method)
				}
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			ctx := context.Background()

			resp, err := client.StopJobWithSavepoint(ctx, tt.jobID, tt.targetDirectory)

			if (err != nil) != tt.wantErr {
				t.Errorf("StopJobWithSavepoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp.RequestID != tt.wantRequestID {
				t.Errorf("request ID = %s, want %s", resp.RequestID, tt.wantRequestID)
			}
		})
	}
}