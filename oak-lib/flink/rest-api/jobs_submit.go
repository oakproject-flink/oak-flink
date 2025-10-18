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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// JarUploadResponse represents the response from uploading a JAR
type JarUploadResponse struct {
	Filename string `json:"filename"`
	Status   string `json:"status"`
}

// JarRunRequest represents a request to run a JAR
type JarRunRequest struct {
	// EntryClass is the main class to execute
	EntryClass string `json:"entryClass,omitempty"`
	// ProgramArgs are arguments for the program
	ProgramArgs string `json:"programArgs,omitempty"`
	// Parallelism for the job
	Parallelism int `json:"parallelism,omitempty"`
	// SavepointPath to restore from
	SavepointPath string `json:"savepointPath,omitempty"`
	// AllowNonRestoredState allows job to start even if savepoint has extra state
	AllowNonRestoredState bool `json:"allowNonRestoredState,omitempty"`
}

// JarRunResponse represents the response from running a JAR
type JarRunResponse struct {
	JobID string `json:"jobid"`
}

// JarsListResponse represents the list of uploaded JARs
type JarsListResponse struct {
	Address string      `json:"address"`
	Files   []JarFile   `json:"files"`
}

// JarFile represents an uploaded JAR file
type JarFile struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Upload int64  `json:"uploaded"`
	Entry  []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"entry"`
}

// UploadJar uploads a JAR file to the Flink cluster
// Endpoint: POST /jars/upload
// Available since: Flink 1.0
func (c *Client) UploadJar(ctx context.Context, jarPath string) (*JarUploadResponse, error) {
	// Open the JAR file
	file, err := os.Open(jarPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open JAR file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("jarfile", filepath.Base(jarPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	contentType := writer.FormDataContentType()
	writer.Close()

	// Create request manually for multipart upload
	url := fmt.Sprintf("%s/jars/upload", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload JAR: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var uploadResp JarUploadResponse
	if err := unmarshalResponse(resp, &uploadResp); err != nil {
		return nil, err
	}

	return &uploadResp, nil
}

// ListJars lists all uploaded JARs
// Endpoint: GET /jars
// Available since: Flink 1.0
func (c *Client) ListJars(ctx context.Context) (*JarsListResponse, error) {
	resp, err := c.doRequest(ctx, "GET", "/jars", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list JARs: %w", err)
	}

	var jarsResp JarsListResponse
	if err := unmarshalResponse(resp, &jarsResp); err != nil {
		return nil, err
	}

	return &jarsResp, nil
}

// RunJar runs an uploaded JAR file
// Endpoint: POST /jars/:jarid/run
// Available since: Flink 1.0
func (c *Client) RunJar(ctx context.Context, jarID string, req JarRunRequest) (*JarRunResponse, error) {
	path := fmt.Sprintf("/jars/%s/run", jarID)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal run request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to run JAR %s: %w", jarID, err)
	}

	var runResp JarRunResponse
	if err := unmarshalResponse(resp, &runResp); err != nil {
		return nil, err
	}

	return &runResp, nil
}

// DeleteJar deletes an uploaded JAR file
// Endpoint: DELETE /jars/:jarid
// Available since: Flink 1.0
func (c *Client) DeleteJar(ctx context.Context, jarID string) error {
	path := fmt.Sprintf("/jars/%s", jarID)

	_, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete JAR %s: %w", jarID, err)
	}

	return nil
}
