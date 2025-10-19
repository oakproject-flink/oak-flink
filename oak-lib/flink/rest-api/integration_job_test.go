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

//go:build integration
// +build integration

package restapi

import (
	"context"
	"testing"
	"time"
)

func TestIntegration_RunningJob(t *testing.T) {
	client, err := NewClient(flinkURL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// List jobs - should have the test job we started in TestMain
	jobs, err := client.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}

	if len(jobs) == 0 {
		t.Fatal("expected at least 1 running job (test job started in setup)")
	}

	t.Logf("Found %d job(s)", len(jobs))

	// Get details of the first job
	job := jobs[0]
	t.Logf("Job ID: %s", job.ID)
	t.Logf("Job Name: %s", job.Name)
	t.Logf("Job Status: %s", job.Status)

	// Job should be running
	if job.Status != JobStatusRunning {
		t.Errorf("expected job status RUNNING, got %s", job.Status)
	}

	// Get detailed job information
	details, err := client.GetJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}

	t.Logf("Job has %d vertices", len(details.Vertices))
	if len(details.Vertices) == 0 {
		t.Error("expected job to have vertices")
	}
}

func TestIntegration_JobMetrics(t *testing.T) {
	client, err := NewClient(flinkURL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get running jobs
	jobs, err := client.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}

	if len(jobs) == 0 {
		t.Skip("No running jobs to test metrics")
	}

	job := jobs[0]

	// Get job metrics
	metrics, err := client.GetJobMetrics(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetJobMetrics failed: %v", err)
	}

	t.Logf("Retrieved %d metrics", len(metrics.Metrics))

	// Just verify we can get metrics
	if metrics.JobID != job.ID {
		t.Errorf("expected metrics for job %s, got %s", job.ID, metrics.JobID)
	}
}

func TestIntegration_TriggerSavepoint(t *testing.T) {
	client, err := NewClient(flinkURL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get running jobs
	jobs, err := client.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}

	if len(jobs) == 0 {
		t.Skip("No running jobs to test savepoint")
	}

	job := jobs[0]

	// Trigger savepoint
	t.Logf("Triggering savepoint for job %s", job.ID)
	savepointResp, err := client.TriggerSavepoint(ctx, job.ID, SavepointTriggerRequest{
		TargetDirectory: "file:///tmp/flink-savepoints",
		CancelJob:       false,
	})

	if err != nil {
		t.Fatalf("TriggerSavepoint failed: %v", err)
	}

	t.Logf("Savepoint triggered with request ID: %s", savepointResp.RequestID)

	// Poll for savepoint completion (with timeout)
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		status, err := client.GetSavepointStatus(ctx, job.ID, savepointResp.RequestID)
		if err != nil {
			t.Fatalf("GetSavepointStatus failed: %v", err)
		}

		t.Logf("Savepoint status: %+v", status)

		// Check if savepoint is complete
		if status.Operation.Location != "" {
			t.Logf("Savepoint completed at: %s", status.Operation.Location)
			return
		}

		// Check for failure
		if status.Operation.FailureCause.Class != "" {
			t.Fatalf("Savepoint failed: %s", status.Operation.FailureCause.Class)
		}

		time.Sleep(2 * time.Second)
	}

	t.Error("Savepoint did not complete within timeout")
}

func TestIntegration_CancelJob(t *testing.T) {
	// Note: This test will cancel the running WordCount job
	// It should be run last
	client, err := NewClient(flinkURL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get running jobs
	jobs, err := client.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}

	if len(jobs) == 0 {
		t.Skip("No running jobs to cancel")
	}

	job := jobs[0]

	// Cancel the job
	t.Logf("Canceling job %s", job.ID)
	err = client.CancelJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("CancelJob failed: %v", err)
	}

	// Wait a bit and verify job is canceled
	time.Sleep(3 * time.Second)

	details, err := client.GetJob(ctx, job.ID)
	if err != nil {
		// Job might not be found after cancellation, which is OK
		t.Logf("Job no longer found (expected after cancellation)")
		return
	}

	t.Logf("Job status after cancel: %s", details.Status)

	if details.Status == JobStatusRunning {
		t.Error("job should not be RUNNING after cancel")
	}
}