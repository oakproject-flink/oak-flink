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
	"os"
	"path/filepath"
	"testing"
)

func TestUploadJar(t *testing.T) {
	// Create a temporary JAR file for testing
	tmpDir := t.TempDir()
	jarPath := filepath.Join(tmpDir, "test.jar")
	if err := os.WriteFile(jarPath, []byte("fake jar content"), 0644); err != nil {
		t.Fatalf("failed to create test JAR: %v", err)
	}

	tests := []struct {
		name           string
		jarPath        string
		responseBody   string
		responseStatus int
		wantErr        bool
		wantFilename   string
	}{
		{
			name:    "successful upload",
			jarPath: jarPath,
			responseBody: `{
				"filename": "/tmp/flink-web-upload/test.jar",
				"status": "success"
			}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantFilename:   "/tmp/flink-web-upload/test.jar",
		},
		{
			name:           "file not found",
			jarPath:        "/nonexistent/file.jar",
			responseBody:   ``,
			responseStatus: http.StatusOK,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "file not found" {
				// Skip server setup for file not found test
				client, err := NewClient("http://localhost:8081")
				if err != nil {
					t.Fatalf("NewClient failed: %v", err)
				}
				defer client.Close()

				ctx := context.Background()
				_, err = client.UploadJar(ctx, tt.jarPath)
				if err == nil {
					t.Error("expected error for non-existent file")
				}
				return
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/jars/upload" {
					t.Errorf("expected path /jars/upload, got %s", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("expected POST method, got %s", r.Method)
				}

				// Verify multipart form
				if err := r.ParseMultipartForm(32 << 20); err != nil {
					t.Errorf("failed to parse multipart form: %v", err)
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

			resp, err := client.UploadJar(ctx, tt.jarPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("UploadJar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp.Filename != tt.wantFilename {
				t.Errorf("filename = %s, want %s", resp.Filename, tt.wantFilename)
			}
		})
	}
}

func TestListJars(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		wantErr        bool
		wantJarCount   int
	}{
		{
			name: "successful response",
			responseBody: `{
				"address": "http://localhost:8081",
				"files": [
					{"id": "jar1", "name": "test1.jar", "uploaded": 1234567890},
					{"id": "jar2", "name": "test2.jar", "uploaded": 1234567891}
				]
			}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantJarCount:   2,
		},
		{
			name:           "empty list",
			responseBody:   `{"address": "http://localhost:8081", "files": []}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantJarCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/jars" {
					t.Errorf("expected path /jars, got %s", r.URL.Path)
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

			jars, err := client.ListJars(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListJars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(jars.Files) != tt.wantJarCount {
				t.Errorf("got %d JARs, want %d", len(jars.Files), tt.wantJarCount)
			}
		})
	}
}

func TestRunJar(t *testing.T) {
	tests := []struct {
		name           string
		jarID          string
		request        JarRunRequest
		responseBody   string
		responseStatus int
		wantErr        bool
		wantJobID      string
	}{
		{
			name:  "successful run",
			jarID: "test-jar-id",
			request: JarRunRequest{
				EntryClass:  "com.example.Main",
				Parallelism: 4,
			},
			responseBody:   `{"jobid": "job-123"}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantJobID:      "job-123",
		},
		{
			name:           "jar not found",
			jarID:          "nonexistent",
			request:        JarRunRequest{},
			responseBody:   `{"errors": ["JAR not found"]}`,
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/jars/" + tt.jarID + "/run"
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("expected POST method, got %s", r.Method)
				}

				// Verify request body
				body, _ := io.ReadAll(r.Body)
				var req JarRunRequest
				if err := json.Unmarshal(body, &req); err == nil {
					if req.EntryClass != tt.request.EntryClass {
						t.Errorf("entry class = %s, want %s", req.EntryClass, tt.request.EntryClass)
					}
					if req.Parallelism != tt.request.Parallelism {
						t.Errorf("parallelism = %d, want %d", req.Parallelism, tt.request.Parallelism)
					}
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

			resp, err := client.RunJar(ctx, tt.jarID, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunJar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp.JobID != tt.wantJobID {
				t.Errorf("job ID = %s, want %s", resp.JobID, tt.wantJobID)
			}
		})
	}
}

func TestDeleteJar(t *testing.T) {
	tests := []struct {
		name           string
		jarID          string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "successful delete",
			jarID:          "test-jar-id",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "jar not found",
			jarID:          "nonexistent",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/jars/" + tt.jarID
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE method, got %s", r.Method)
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

			err = client.DeleteJar(ctx, tt.jarID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteJar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
