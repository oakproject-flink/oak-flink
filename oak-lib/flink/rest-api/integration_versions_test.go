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

//go:build integration_versions
// +build integration_versions

package restapi

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// FlinkVersion represents a Flink version to test
type FlinkVersion struct {
	Version        string  // e.g., "1.18.1"
	ComposeFile    string  // docker-compose filename
	ExpectedAPIVersion Version // expected version range from DetectVersion
}

// All Flink versions we support (1.18-2.1)
var supportedVersions = []FlinkVersion{
	{Version: "1.18.1", ComposeFile: "docker-compose-1.18.yml", ExpectedAPIVersion: Version1_18to1_19},
	{Version: "1.19.1", ComposeFile: "docker-compose-1.19.yml", ExpectedAPIVersion: Version1_18to1_19},
	{Version: "1.20.0", ComposeFile: "docker-compose-1.20.yml", ExpectedAPIVersion: Version2_0Plus},
	{Version: "2.0.0", ComposeFile: "docker-compose-2.0.yml", ExpectedAPIVersion: Version2_0Plus},
	{Version: "2.1.0", ComposeFile: "docker-compose-2.1.yml", ExpectedAPIVersion: Version2_0Plus},
}

const (
	versionTestFlinkURL       = "http://localhost:8081"
	versionTestStartupTimeout = 90 * time.Second
)

// TestAllFlinkVersions runs the full integration test suite against each supported Flink version
func TestAllFlinkVersions(t *testing.T) {
	for _, fv := range supportedVersions {
		fv := fv // capture range variable
		t.Run(fmt.Sprintf("Flink_%s", fv.Version), func(t *testing.T) {
			// Start Flink cluster for this version
			composeFile := filepath.Join("testdata", fv.ComposeFile)
			if err := startFlinkCluster(composeFile); err != nil {
				t.Fatalf("Failed to start Flink %s: %v", fv.Version, err)
			}
			defer cleanupFlinkCluster(composeFile)

			// Wait for Flink to be ready
			if err := waitForFlinkReady(t, fv.Version); err != nil {
				t.Fatalf("Flink %s did not become ready: %v", fv.Version, err)
			}

			t.Logf("✓ Flink %s cluster is ready", fv.Version)

			// Wait for TaskManager to register (can take a few seconds)
			if err := waitForTaskManager(t, fv.Version); err != nil {
				t.Logf("Warning: TaskManager may not have fully registered: %v", err)
			}

			// Run test suite
			runVersionTestSuite(t, fv)
		})
	}
}

// startFlinkCluster starts a Flink cluster using the specified docker-compose file
func startFlinkCluster(composeFile string) error {
	fmt.Printf("Starting Flink cluster with %s...\n", filepath.Base(composeFile))
	cmd := exec.Command("docker", "compose", "-f", composeFile, "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start cluster: %w\n%s", err, string(output))
	}
	return nil
}

// cleanupFlinkCluster stops and removes the Flink cluster
func cleanupFlinkCluster(composeFile string) {
	fmt.Printf("Stopping Flink cluster...\n")
	cmd := exec.Command("docker", "compose", "-f", composeFile, "down", "-v")
	cmd.Run()
}

// waitForFlinkReady waits for Flink to accept connections
func waitForFlinkReady(t *testing.T, version string) error {
	client, err := NewClient(versionTestFlinkURL, WithTimeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	deadline := time.Now().Add(versionTestStartupTimeout)
	attempts := 0

	for time.Now().Before(deadline) {
		attempts++
		_, err := client.GetClusterOverview(ctx)
		if err == nil {
			t.Logf("  Flink %s ready after %d attempts", version, attempts)
			return nil
		}

		if attempts%10 == 0 {
			t.Logf("  Waiting for Flink %s... (%d attempts)", version, attempts)
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout after %d attempts", attempts)
}

// waitForTaskManager waits for at least one TaskManager to register
func waitForTaskManager(t *testing.T, version string) error {
	client, err := NewClient(versionTestFlinkURL, WithTimeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	deadline := time.Now().Add(30 * time.Second)
	attempts := 0

	for time.Now().Before(deadline) {
		attempts++
		overview, err := client.GetClusterOverview(ctx)
		if err == nil && overview.TaskManagers >= 1 {
			t.Logf("  TaskManager registered after %d attempts", attempts)
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for TaskManager after %d attempts", attempts)
}

// runVersionTestSuite runs the core tests for a specific Flink version
func runVersionTestSuite(t *testing.T, fv FlinkVersion) {
	client, err := NewClient(versionTestFlinkURL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Test 1: Get Cluster Overview
	t.Run("ClusterOverview", func(t *testing.T) {
		overview, err := client.GetClusterOverview(ctx)
		if err != nil {
			t.Fatalf("GetClusterOverview failed: %v", err)
		}

		if overview.FlinkVersion == "" {
			t.Error("FlinkVersion should not be empty")
		}

		if overview.TaskManagers < 1 {
			t.Errorf("expected at least 1 task manager, got %d", overview.TaskManagers)
		}

		t.Logf("  Flink version: %s, TaskManagers: %d, Slots: %d",
			overview.FlinkVersion, overview.TaskManagers, overview.SlotsTotal)
	})

	// Test 2: Version Detection
	t.Run("VersionDetection", func(t *testing.T) {
		version, err := client.DetectVersion(ctx)
		if err != nil {
			t.Fatalf("DetectVersion failed: %v", err)
		}

		if version != fv.ExpectedAPIVersion {
			t.Errorf("expected version %s, got %s", fv.ExpectedAPIVersion, version)
		}

		t.Logf("  Detected API version: %s", version)
	})

	// Test 3: Get Config
	t.Run("GetConfig", func(t *testing.T) {
		config, err := client.GetConfig(ctx)
		if err != nil {
			t.Fatalf("GetConfig failed: %v", err)
		}

		if len(config.Entries) == 0 {
			t.Error("expected config entries, got none")
		}

		t.Logf("  Config entries: %d", len(config.Entries))
	})

	// Test 4: List Jobs
	t.Run("ListJobs", func(t *testing.T) {
		jobs, err := client.ListJobs(ctx)
		if err != nil {
			t.Fatalf("ListJobs failed: %v", err)
		}

		if jobs == nil {
			t.Error("jobs should not be nil")
		}

		t.Logf("  Jobs found: %d", len(jobs))
	})

	// Test 5: JAR Upload & List
	var testJobID string
	var testJarID string
	t.Run("JARUploadAndList", func(t *testing.T) {
		// Copy example JAR from container
		jarPath := filepath.Join("testdata", "TopSpeedWindowing.jar")
		cmd := exec.Command("docker", "cp", "flink-jobmanager-test:/opt/flink/examples/streaming/TopSpeedWindowing.jar", jarPath)
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to copy JAR from container: %v", err)
		}

		// Upload JAR
		uploadResp, err := client.UploadJar(ctx, jarPath)
		if err != nil {
			t.Fatalf("UploadJar failed: %v", err)
		}
		t.Logf("  JAR uploaded: %s", filepath.Base(uploadResp.Filename))
		testJarID = filepath.Base(uploadResp.Filename)

		// List JARs to verify upload
		jars, err := client.ListJars(ctx)
		if err != nil {
			t.Fatalf("ListJars failed: %v", err)
		}
		if len(jars.Files) == 0 {
			t.Error("expected at least one JAR after upload")
		}
		t.Logf("  JARs in system: %d", len(jars.Files))
	})

	// Test 6: Run JAR
	t.Run("RunJAR", func(t *testing.T) {
		if testJarID == "" {
			t.Skip("no JAR ID available from previous test")
		}

		runResp, err := client.RunJar(ctx, testJarID, JarRunRequest{Parallelism: 1})
		if err != nil {
			t.Fatalf("RunJar failed: %v", err)
		}
		testJobID = runResp.JobID
		t.Logf("  Job started: %s", testJobID)

		// Wait for job to actually start running
		time.Sleep(3 * time.Second)
	})

	// Test 7: Get Job Details
	t.Run("GetJobDetails", func(t *testing.T) {
		if testJobID == "" {
			t.Skip("no job ID available")
		}

		job, err := client.GetJob(ctx, testJobID)
		if err != nil {
			t.Fatalf("GetJob failed: %v", err)
		}
		t.Logf("  Job ID: %s, Status: %s, Vertices: %d", job.ID, job.Status, len(job.Vertices))

		if job.ID != testJobID {
			t.Errorf("expected job ID %s, got %s", testJobID, job.ID)
		}
		if len(job.Vertices) == 0 {
			t.Error("expected job to have vertices")
		}
	})

	// Test 8: Get Job Config
	t.Run("GetJobConfig", func(t *testing.T) {
		if testJobID == "" {
			t.Skip("no job ID available")
		}

		config, err := client.GetJobConfig(ctx, testJobID)
		if err != nil {
			t.Fatalf("GetJobConfig failed: %v", err)
		}
		t.Logf("  Job config entries: %d", len(config.Entries))

		if len(config.Entries) == 0 {
			t.Error("expected job config to have entries")
		}
	})

	// Test 9: Get Job Metrics
	t.Run("GetJobMetrics", func(t *testing.T) {
		if testJobID == "" {
			t.Skip("no job ID available")
		}

		metrics, err := client.GetJobMetrics(ctx, testJobID)
		if err != nil {
			t.Fatalf("GetJobMetrics failed: %v", err)
		}
		t.Logf("  Job metrics: %d", len(metrics.Metrics))

		if metrics.JobID != testJobID {
			t.Errorf("expected metrics for job %s, got %s", testJobID, metrics.JobID)
		}
	})

	// Test 10: Get Vertex Metrics
	t.Run("GetVertexMetrics", func(t *testing.T) {
		if testJobID == "" {
			t.Skip("no job ID available")
		}

		// Get job to find a vertex ID
		job, err := client.GetJob(ctx, testJobID)
		if err != nil || len(job.Vertices) == 0 {
			t.Skip("no vertices available")
		}

		vertexID := job.Vertices[0].ID
		metrics, err := client.GetVertexMetrics(ctx, testJobID, vertexID)
		if err != nil {
			t.Fatalf("GetVertexMetrics failed: %v", err)
		}
		t.Logf("  Vertex metrics: %d", len(metrics))
	})

	// Test 11: Trigger Savepoint
	var savepointTriggerID string
	t.Run("TriggerSavepoint", func(t *testing.T) {
		if testJobID == "" {
			t.Skip("no job ID available")
		}

		resp, err := client.TriggerSavepoint(ctx, testJobID, SavepointTriggerRequest{
			TargetDirectory: "file:///tmp/flink-savepoints",
			CancelJob:       false,
		})
		if err != nil {
			t.Fatalf("TriggerSavepoint failed: %v", err)
		}
		savepointTriggerID = resp.RequestID
		t.Logf("  Savepoint triggered: %s", savepointTriggerID)
	})

	// Test 12: Get Savepoint Status
	t.Run("GetSavepointStatus", func(t *testing.T) {
		if testJobID == "" || savepointTriggerID == "" {
			t.Skip("no savepoint trigger available")
		}

		// Poll for completion (with timeout)
		deadline := time.Now().Add(20 * time.Second)
		for time.Now().Before(deadline) {
			status, err := client.GetSavepointStatus(ctx, testJobID, savepointTriggerID)
			if err != nil {
				t.Fatalf("GetSavepointStatus failed: %v", err)
			}

			if status.Operation.Location != "" {
				t.Logf("  Savepoint completed: %s", status.Operation.Location)
				return
			}

			if status.Operation.FailureCause.Class != "" {
				t.Fatalf("Savepoint failed: %s", status.Operation.FailureCause.Class)
			}

			time.Sleep(2 * time.Second)
		}
		t.Log("  Savepoint still in progress (timeout)")
	})

	// Test 13: Cancel Job
	t.Run("CancelJob", func(t *testing.T) {
		if testJobID == "" {
			t.Skip("no job ID available")
		}

		err := client.CancelJob(ctx, testJobID)
		if err != nil {
			t.Fatalf("CancelJob failed: %v", err)
		}
		t.Logf("  Job cancelled: %s", testJobID)

		// Wait a bit for cancellation
		time.Sleep(2 * time.Second)
	})

	// Test 14: Delete JAR
	t.Run("DeleteJAR", func(t *testing.T) {
		if testJarID == "" {
			t.Skip("no JAR ID available")
		}

		err := client.DeleteJar(ctx, testJarID)
		if err != nil {
			t.Fatalf("DeleteJar failed: %v", err)
		}
		t.Logf("  JAR deleted: %s", testJarID)
	})

	// Test 15: Error Handling
	t.Run("ErrorHandling", func(t *testing.T) {
		_, err := client.GetJob(ctx, "non-existent-job-id")
		if err == nil {
			t.Error("expected error for non-existent job, got nil")
		}
		t.Logf("  Error handling works correctly")
	})
}

