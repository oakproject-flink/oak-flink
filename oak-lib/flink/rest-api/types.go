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

// JobStatus represents the state of a Flink job
type JobStatus string

const (
	JobStatusCreated     JobStatus = "CREATED"
	JobStatusRunning     JobStatus = "RUNNING"
	JobStatusFailing     JobStatus = "FAILING"
	JobStatusFailed      JobStatus = "FAILED"
	JobStatusCanceling   JobStatus = "CANCELING"
	JobStatusCanceled    JobStatus = "CANCELED"
	JobStatusFinished    JobStatus = "FINISHED"
	JobStatusRestarting  JobStatus = "RESTARTING"
	JobStatusSuspended   JobStatus = "SUSPENDED"
	JobStatusReconciling JobStatus = "RECONCILING"
)

// Job represents a Flink job
type Job struct {
	ID     string    `json:"id"`
	Name   string    `json:"name,omitempty"`
	Status JobStatus `json:"status"`
	// StartTime in milliseconds since epoch
	StartTime int64 `json:"start-time,omitempty"`
	// EndTime in milliseconds since epoch
	EndTime int64 `json:"end-time,omitempty"`
	// Duration in milliseconds
	Duration int64 `json:"duration,omitempty"`
	// Parallelism of the job
	Tasks struct {
		Total int `json:"total"`
	} `json:"tasks,omitempty"`
}

// JobsOverview represents the overview of all jobs
type JobsOverview struct {
	Jobs []Job `json:"jobs"`
}

// JobDetails represents detailed information about a specific job
type JobDetails struct {
	// Note: Job details endpoint uses "jid" and "state" instead of "id" and "status"
	ID        string    `json:"jid"`
	Name      string    `json:"name"`
	Status    JobStatus `json:"state"`
	StartTime int64     `json:"start-time,omitempty"`
	EndTime   int64     `json:"end-time,omitempty"`
	Duration  int64     `json:"duration,omitempty"`
	Vertices  []Vertex  `json:"vertices"`
	Plan      JobPlan   `json:"plan"`
}

// Vertex represents a vertex (operator) in the job graph
type Vertex struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Parallelism int       `json:"parallelism"`
	Status      JobStatus `json:"status"`
}

// JobPlan represents the execution plan
type JobPlan struct {
	Nodes []PlanNode `json:"nodes"`
}

// PlanNode represents a node in the execution plan
type PlanNode struct {
	ID          string `json:"id"`
	Parallelism int    `json:"parallelism"`
	Operator    string `json:"operator"`
}

// SavepointTriggerRequest is the request to trigger a savepoint
type SavepointTriggerRequest struct {
	// TargetDirectory is the directory where the savepoint will be stored
	TargetDirectory string `json:"target-directory,omitempty"`
	// CancelJob indicates whether to cancel the job after taking the savepoint
	CancelJob bool `json:"cancel-job,omitempty"`
}

// SavepointTriggerResponse is the response from triggering a savepoint
type SavepointTriggerResponse struct {
	// RequestID is the ID to track the savepoint operation
	RequestID string `json:"request-id"`
}

// SavepointStatus represents the status of a savepoint operation
type SavepointStatus struct {
	Status struct {
		ID string `json:"id"`
	} `json:"status"`
	Operation struct {
		// Location is the path to the completed savepoint
		Location string `json:"location,omitempty"`
		// FailureCause if the savepoint failed
		FailureCause struct {
			Class      string `json:"class"`
			StackTrace string `json:"stack-trace"`
		} `json:"failure-cause,omitempty"`
	} `json:"operation"`
}

// JobMetrics represents metrics for a job
type JobMetrics struct {
	// JobID is the job identifier
	JobID string
	// Metrics is a map of metric name to value
	Metrics map[string]float64
}

// MetricResponse represents the response from the metrics endpoint
type MetricResponse struct {
	Metrics []Metric `json:"metrics"`
}

// Metric represents a single metric
type Metric struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// ClusterOverview represents the Flink cluster overview
type ClusterOverview struct {
	// TaskManagers is the number of task managers
	TaskManagers int `json:"taskmanagers"`
	// SlotsTotal is the total number of task slots
	SlotsTotal int `json:"slots-total"`
	// SlotsAvailable is the number of available task slots
	SlotsAvailable int `json:"slots-available"`
	// JobsRunning is the number of running jobs
	JobsRunning int `json:"jobs-running"`
	// JobsFinished is the number of finished jobs
	JobsFinished int `json:"jobs-finished"`
	// JobsCancelled is the number of cancelled jobs
	JobsCancelled int `json:"jobs-cancelled"`
	// JobsFailed is the number of failed jobs
	JobsFailed int `json:"jobs-failed"`
	// FlinkVersion is the Flink version
	FlinkVersion string `json:"flink-version"`
	// FlinkCommit is the Flink commit ID
	FlinkCommit string `json:"flink-commit"`
}

// ConfigEntry represents a configuration entry
type ConfigEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ConfigResponse represents the configuration response
type ConfigResponse struct {
	Entries []ConfigEntry `json:"entries"`
}

// JobConfigResponse represents the job configuration response from /jobs/:jobid/config
// This has a different structure than the cluster config endpoint
type JobConfigResponse struct {
	JobID           string                 `json:"jid"`
	Name            string                 `json:"name"`
	ExecutionConfig map[string]interface{} `json:"execution-config"`
}
