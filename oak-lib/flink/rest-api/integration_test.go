// Copyright 2025 Andrei Grigoriu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicablem law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build integration
// +build integration

package restapi

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const (
	flinkURL       = "http://localhost:8081"
	startupTimeout = 60 * time.Second
)

// TestMain sets up and tears down the Flink cluster for integration tests
func TestMain(m *testing.M) {
	// Get path to docker-compose file
	composeFile := filepath.Join("testdata", "docker-compose.yml")

	// Start Flink cluster
	fmt.Println("Starting Flink cluster...")
	cmd := exec.Command("docker", "compose", "-f", composeFile, "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to start Flink cluster: %v\n%s\n", err, string(output))
		os.Exit(1)
	}
	fmt.Printf("Docker Compose output:\n%s\n", string(output))

	// Show container status
	printContainerStatus()

	// Wait for Flink to be ready
	if err := waitForFlink(); err != nil {
		fmt.Printf("Flink cluster did not become ready: %v\n", err)
		printContainerLogs()
		cleanup(composeFile)
		os.Exit(1)
	}

	fmt.Println("Flink cluster is ready")

	// Start a test job for job-related tests
	if err := startTestJob(); err != nil {
		fmt.Printf("Failed to start test job: %v\n", err)
		printContainerLogs()
		cleanup(composeFile)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	cleanup(composeFile)

	os.Exit(code)
}

func printContainerStatus() {
	cmd := exec.Command("docker", "ps", "-a", "--filter", "name=flink", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
	output, err := cmd.CombinedOutput()
	if err == nil {
		fmt.Printf("\nContainer Status:\n%s\n", string(output))
	}
}

func printContainerLogs() {
	fmt.Println("\n=== JobManager Logs ===")
	cmd := exec.Command("docker", "logs", "--tail", "50", "flink-jobmanager-test")
	output, _ := cmd.CombinedOutput()
	fmt.Println(string(output))

	fmt.Println("\n=== TaskManager Logs ===")
	cmd = exec.Command("docker", "logs", "--tail", "50", "flink-taskmanager-test")
	output, _ = cmd.CombinedOutput()
	fmt.Println(string(output))
}

func waitForFlink() error {
	client := NewClient(flinkURL, WithTimeout(5*time.Second))
	ctx := context.Background()

	fmt.Print("Waiting for Flink to be ready")
	deadline := time.Now().Add(startupTimeout)
	attempts := 0
	for time.Now().Before(deadline) {
		attempts++
		_, err := client.GetClusterOverview(ctx)
		if err == nil {
			fmt.Printf(" ready after %d attempts!\n", attempts)
			return nil
		}

		fmt.Print(".")
		if attempts%10 == 0 {
			fmt.Printf(" (%d attempts)\n", attempts)
			printContainerStatus()
			fmt.Print("Continuing to wait")
		}

		time.Sleep(2 * time.Second)
	}

	fmt.Println()
	return fmt.Errorf("timeout waiting for Flink to start after %d attempts", attempts)
}

func startTestJob() error {
	// Extract a simple example JAR from the Flink container and upload it
	fmt.Println("Starting a test job...")

	// Copy the TopSpeedWindowing JAR from the container
	jarPath := filepath.Join("testdata", "TopSpeedWindowing.jar")
	cmd := exec.Command("docker", "cp", "flink-jobmanager-test:/opt/flink/examples/streaming/TopSpeedWindowing.jar", jarPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy JAR from container: %w", err)
	}

	client := NewClient(flinkURL, WithTimeout(30*time.Second))
	ctx := context.Background()

	// Upload the JAR
	fmt.Println("Uploading JAR...")
	uploadResp, err := client.UploadJar(ctx, jarPath)
	if err != nil {
		return fmt.Errorf("failed to upload JAR: %w", err)
	}
	fmt.Printf("JAR uploaded: %s\n", uploadResp.Filename)

	// Get the JAR ID from the filename (extract just the basename)
	jarID := filepath.Base(uploadResp.Filename)

	// Run the JAR
	fmt.Println("Running job...")
	runResp, err := client.RunJar(ctx, jarID, JarRunRequest{
		Parallelism: 1,
	})
	if err != nil {
		return fmt.Errorf("failed to run JAR: %w", err)
	}

	fmt.Printf("Job started with ID: %s\n", runResp.JobID)

	// Wait a bit for the job to actually start
	time.Sleep(3 * time.Second)

	// Verify the job is running
	jobs, err := client.ListJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	if len(jobs) == 0 {
		return fmt.Errorf("job was submitted but is not in the jobs list")
	}

	fmt.Printf("Job is running (status: %s)\n", jobs[0].Status)
	return nil
}

func cleanup(composeFile string) {
	fmt.Println("Stopping Flink cluster...")
	cmd := exec.Command("docker", "compose", "-f", composeFile, "down", "-v")
	cmd.Run()
}

func TestIntegration_GetClusterOverview(t *testing.T) {
	client := NewClient(flinkURL)
	ctx := context.Background()

	overview, err := client.GetClusterOverview(ctx)
	if err != nil {
		t.Fatalf("GetClusterOverview failed: %v", err)
	}

	// Verify Flink version
	if overview.FlinkVersion == "" {
		t.Error("FlinkVersion should not be empty")
	}

	t.Logf("Flink Version: %s", overview.FlinkVersion)
	t.Logf("Task Managers: %d", overview.TaskManagers)
	t.Logf("Total Slots: %d", overview.SlotsTotal)

	// Should have at least 1 task manager
	if overview.TaskManagers < 1 {
		t.Errorf("expected at least 1 task manager, got %d", overview.TaskManagers)
	}

	// Should have task slots
	if overview.SlotsTotal < 1 {
		t.Errorf("expected at least 1 slot, got %d", overview.SlotsTotal)
	}
}

func TestIntegration_DetectVersion(t *testing.T) {
	client := NewClient(flinkURL)
	ctx := context.Background()

	version, err := client.DetectVersion(ctx)
	if err != nil {
		t.Fatalf("DetectVersion failed: %v", err)
	}

	t.Logf("Detected version range: %s", version)

	// Should detect Flink 2.x
	if version != Version2_0Plus {
		t.Errorf("expected version 2.0+, got %s", version)
	}
}

func TestIntegration_GetConfig(t *testing.T) {
	client := NewClient(flinkURL)
	ctx := context.Background()

	config, err := client.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	if len(config.Entries) == 0 {
		t.Error("expected config entries, got none")
	}

	// Look for some expected config keys
	found := false
	for _, entry := range config.Entries {
		if entry.Key == "jobmanager.rpc.address" {
			found = true
			t.Logf("Found jobmanager.rpc.address: %s", entry.Value)
			break
		}
	}

	if !found {
		t.Error("expected to find jobmanager.rpc.address in config")
	}
}

func TestIntegration_ListJobs(t *testing.T) {
	client := NewClient(flinkURL)
	ctx := context.Background()

	jobs, err := client.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}

	// Fresh cluster should have no jobs
	t.Logf("Found %d jobs", len(jobs))

	// This should not error even with no jobs
	if jobs == nil {
		t.Error("jobs should not be nil")
	}
}

func TestIntegration_ContextCancellation(t *testing.T) {
	client := NewClient(flinkURL)

	// Create context with immediate timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a bit to ensure context is cancelled
	time.Sleep(10 * time.Millisecond)

	_, err := client.GetClusterOverview(ctx)
	if err == nil {
		t.Error("expected error due to context cancellation, got nil")
	}

	t.Logf("Got expected error: %v", err)
}

func TestIntegration_ConcurrentRequests(t *testing.T) {
	client := NewClient(flinkURL)
	ctx := context.Background()

	// Make multiple concurrent requests
	const numRequests = 10
	errChan := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			_, err := client.GetClusterOverview(ctx)
			errChan <- err
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("concurrent request %d failed: %v", i, err)
		}
	}

	t.Logf("Successfully completed %d concurrent requests", numRequests)
}

func TestIntegration_ErrorHandling(t *testing.T) {
	client := NewClient(flinkURL)
	ctx := context.Background()

	// Try to get a non-existent job
	_, err := client.GetJob(ctx, "non-existent-job-id")
	if err == nil {
		t.Error("expected error for non-existent job, got nil")
	}

	t.Logf("Got expected error for non-existent job: %v", err)
}

func TestIntegration_JAROperations(t *testing.T) {
	client := NewClient(flinkURL)
	ctx := context.Background()

	jarPath := "./testdata/TopSpeedWindowing.jar"

	// Test 1: Upload JAR
	t.Run("UploadJar", func(t *testing.T) {
		uploadResp, err := client.UploadJar(ctx, jarPath)
		if err != nil {
			t.Fatalf("UploadJar failed: %v", err)
		}

		if uploadResp.Filename == "" {
			t.Error("expected filename in upload response")
		}

		t.Logf("Uploaded JAR: %s (status: %s)", uploadResp.Filename, uploadResp.Status)
	})

	// Test 2: List JARs
	var jarID string
	t.Run("ListJars", func(t *testing.T) {
		jarsResp, err := client.ListJars(ctx)
		if err != nil {
			t.Fatalf("ListJars failed: %v", err)
		}

		if len(jarsResp.Files) == 0 {
			t.Fatal("expected at least one JAR in list")
		}

		// Get the JAR ID for next tests
		jarID = jarsResp.Files[0].ID
		t.Logf("Found %d JARs, first JAR ID: %s", len(jarsResp.Files), jarID)
	})

	// Test 3: Run JAR
	var jobID string
	t.Run("RunJar", func(t *testing.T) {
		if jarID == "" {
			t.Skip("no JAR ID available from previous test")
		}

		runReq := JarRunRequest{
			Parallelism: 1,
		}

		runResp, err := client.RunJar(ctx, jarID, runReq)
		if err != nil {
			t.Fatalf("RunJar failed: %v", err)
		}

		if runResp.JobID == "" {
			t.Error("expected job ID in run response")
		}

		jobID = runResp.JobID
		t.Logf("Started job from JAR: %s", jobID)

		// Cancel the job after starting to clean up
		if jobID != "" {
			time.Sleep(2 * time.Second) // Give job time to start
			if err := client.CancelJob(ctx, jobID); err != nil {
				t.Logf("Warning: failed to cancel test job %s: %v", jobID, err)
			} else {
				t.Logf("Cancelled test job: %s", jobID)
			}
		}
	})

	// Test 4: Delete JAR
	t.Run("DeleteJar", func(t *testing.T) {
		if jarID == "" {
			t.Skip("no JAR ID available from previous test")
		}

		err := client.DeleteJar(ctx, jarID)
		if err != nil {
			t.Fatalf("DeleteJar failed: %v", err)
		}

		t.Logf("Deleted JAR: %s", jarID)

		// Verify it's deleted by listing JARs
		jarsResp, err := client.ListJars(ctx)
		if err != nil {
			t.Fatalf("ListJars failed: %v", err)
		}

		// Check that the JAR is no longer in the list
		for _, jar := range jarsResp.Files {
			if jar.ID == jarID {
				t.Errorf("JAR %s still exists after deletion", jarID)
			}
		}
	})
}
