# Flink REST API Client

A Go client library for Apache Flink's REST API with support for versions 1.8 through 2.1+.

## Features

- **Version-aware**: Supports Flink 1.8 through 2.1+ with automatic version detection
- **Context-based**: All operations are async-ready using `context.Context`
- **Type-safe**: Strong typing for all API requests and responses
- **Minimal duplication**: Shared code for common endpoints, version-specific overrides only when needed

## Installation

```go
import "github.com/oakproject-flink/oak-flink/oak-lib/flink"
```

## Usage

### Basic Client Creation

```go
// Create client with auto-detection
client := flink.NewClient("http://localhost:8081")

// Create client with specific version
client := flink.NewClient("http://localhost:8081",
    flink.WithVersion(flink.Version1_18to1_19),
    flink.WithTimeout(30 * time.Second),
)
```

### List Jobs

```go
ctx := context.Background()

jobs, err := client.ListJobs(ctx)
if err != nil {
    log.Fatal(err)
}

for _, job := range jobs {
    fmt.Printf("Job: %s (%s) - Status: %s\n", job.Name, job.ID, job.Status)
}
```

### Get Job Details

```go
details, err := client.GetJob(ctx, "job-id-here")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Job %s has %d vertices\n", details.Name, len(details.Vertices))
```

### Trigger Savepoint

```go
resp, err := client.TriggerSavepoint(ctx, "job-id", flink.SavepointTriggerRequest{
    TargetDirectory: "s3://bucket/savepoints",
    CancelJob:       false,
})

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Savepoint triggered: %s\n", resp.RequestID)

// Poll for status
status, err := client.GetSavepointStatus(ctx, "job-id", resp.RequestID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Savepoint location: %s\n", status.Operation.Location)
```

### Get Metrics

```go
// Get all metrics
metrics, err := client.GetJobMetrics(ctx, "job-id")

// Get specific metrics
metrics, err := client.GetJobMetrics(ctx, "job-id",
    flink.MetricNumRecordsIn,
    flink.MetricNumRecordsOut,
    flink.MetricBackPressuredTime,
)

for name, value := range metrics.Metrics {
    fmt.Printf("%s: %.2f\n", name, value)
}
```

### Cluster Overview

```go
overview, err := client.GetClusterOverview(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Flink %s\n", overview.FlinkVersion)
fmt.Printf("Task Managers: %d\n", overview.TaskManagers)
fmt.Printf("Total Slots: %d (Available: %d)\n",
    overview.SlotsTotal, overview.SlotsAvailable)
```

## Supported Versions

The client groups Flink versions into ranges based on REST API compatibility:

- **Version1_8to1_12**: Flink 1.8 through 1.12
- **Version1_13to1_17**: Flink 1.13 through 1.17
- **Version1_18to1_19**: Flink 1.18 through 1.19
- **Version2_0Plus**: Flink 2.0 and above
- **VersionAuto**: Auto-detect (default)

Most endpoints are identical across versions. Version-specific behavior is handled internally (e.g., `StopJobWithSavepoint` uses different endpoints for different versions).

## API Coverage

### Jobs
- ✅ List all jobs
- ✅ Get job details
- ✅ Cancel job
- ✅ Get job configuration

### Savepoints
- ✅ Trigger savepoint
- ✅ Get savepoint status
- ✅ Stop job with savepoint (version-aware)

### Metrics
- ✅ Get job metrics
- ✅ Get vertex metrics
- ✅ Predefined metric constants

### Cluster
- ✅ Get cluster overview
- ✅ Get cluster configuration
- ✅ Auto-detect Flink version

## Testing

```bash
cd oak-lib
go test ./flink/... -v
```

## Thread Safety

The client is safe for concurrent use. All methods accept `context.Context` for cancellation and timeout control.

## Error Handling

All methods return descriptive errors that wrap the underlying cause:

```go
jobs, err := client.ListJobs(ctx)
if err != nil {
    // Error includes context like "failed to list jobs: HTTP 404: ..."
    log.Printf("Error: %v", err)
}
```

## License

Copyright 2025 Andrei Grigoriu

Licensed under the Apache License, Version 2.0.