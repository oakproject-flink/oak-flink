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

package flink

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
