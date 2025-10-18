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
	"strings"
	"testing"
)

func TestGetClusterOverview(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		wantErr        bool
		wantVersion    string
		wantTMs        int
	}{
		{
			name: "successful response",
			responseBody: `{
				"taskmanagers": 3,
				"slots-total": 12,
				"slots-available": 8,
				"jobs-running": 2,
				"jobs-finished": 5,
				"jobs-cancelled": 1,
				"jobs-failed": 0,
				"flink-version": "2.1.0",
				"flink-commit": "abc123"
			}`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantVersion:    "2.1.0",
			wantTMs:        3,
		},
		{
			name:           "server error",
			responseBody:   `{"error": "internal error"}`,
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
		{
			name:           "invalid json",
			responseBody:   `invalid json`,
			responseStatus: http.StatusOK,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/overview" {
					t.Errorf("expected path /overview, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			ctx := context.Background()

			overview, err := client.GetClusterOverview(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetClusterOverview() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if overview.FlinkVersion != tt.wantVersion {
					t.Errorf("FlinkVersion = %s, want %s", overview.FlinkVersion, tt.wantVersion)
				}
				if overview.TaskManagers != tt.wantTMs {
					t.Errorf("TaskManagers = %d, want %d", overview.TaskManagers, tt.wantTMs)
				}
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		wantErr        bool
		wantEntries    int
	}{
		{
			name: "successful response",
			responseBody: `[
				{"key": "jobmanager.rpc.address", "value": "localhost"},
				{"key": "rest.port", "value": "8081"},
				{"key": "parallelism.default", "value": "4"}
			]`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantEntries:    3,
		},
		{
			name:           "empty config",
			responseBody:   `[]`,
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantEntries:    0,
		},
		{
			name:           "server error",
			responseBody:   `{"error": "not found"}`,
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/jobmanager/config" {
					t.Errorf("expected path /jobmanager/config, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			ctx := context.Background()

			config, err := client.GetConfig(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(config.Entries) != tt.wantEntries {
					t.Errorf("got %d entries, want %d", len(config.Entries), tt.wantEntries)
				}
			}
		})
	}
}

func TestDetectVersion(t *testing.T) {
	tests := []struct {
		name        string
		flinkVer    string
		wantVersion Version
		wantErr     bool
		errContains string
	}{
		{
			name:        "Flink 2.1.0 - supported (max)",
			flinkVer:    "2.1.0",
			wantVersion: Version2_0Plus,
			wantErr:     false,
		},
		{
			name:        "Flink 2.0.0 - supported",
			flinkVer:    "2.0.0",
			wantVersion: Version2_0Plus,
			wantErr:     false,
		},
		{
			name:        "Flink 1.19.0 - supported",
			flinkVer:    "1.19.0",
			wantVersion: Version1_18to1_19,
			wantErr:     false,
		},
		{
			name:        "Flink 1.18.0 - supported (min)",
			flinkVer:    "1.18.0",
			wantVersion: Version1_18to1_19,
			wantErr:     false,
		},
		{
			name:        "Flink 1.17.0 - not supported (too old)",
			flinkVer:    "1.17.0",
			wantErr:     true,
			errContains: "not supported (minimum version: 1.18)",
		},
		{
			name:        "Flink 1.13.0 - not supported (too old)",
			flinkVer:    "1.13.0",
			wantErr:     true,
			errContains: "not supported (minimum version: 1.18)",
		},
		{
			name:        "Flink 1.8.0 - not supported (too old)",
			flinkVer:    "1.8.0",
			wantErr:     true,
			errContains: "not supported (minimum version: 1.18)",
		},
		{
			name:        "Flink 2.2.0 - not supported (too new)",
			flinkVer:    "2.2.0",
			wantErr:     true,
			errContains: "not supported (maximum version: 2.1)",
		},
		{
			name:        "Flink 3.0.0 - not supported (too new)",
			flinkVer:    "3.0.0",
			wantErr:     true,
			errContains: "not supported (maximum version: 2.1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := `{"flink-version": "` + tt.flinkVer + `", "taskmanagers": 1, "slots-total": 4}`
				w.Write([]byte(response))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			ctx := context.Background()

			version, err := client.DetectVersion(ctx)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("DetectVersion() expected error but got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("DetectVersion() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("DetectVersion() unexpected error = %v", err)
			}

			if version != tt.wantVersion {
				t.Errorf("DetectVersion() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}