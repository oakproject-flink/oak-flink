# Oak API - gRPC Protocol Definitions

This module contains the Protocol Buffer definitions for Oak Agent ↔ Oak Server communication.

## Prerequisites

### Install Protocol Buffer Compiler (protoc)

**Windows:**
```powershell
# Using winget (recommended - pre-installed on Windows 10/11)
winget install protoc

# Or using Chocolatey
choco install protoc

# Or using Scoop
scoop install protobuf

# Or download manually from GitHub releases
# https://github.com/protocolbuffers/protobuf/releases
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install protobuf-compiler

# Or download latest from releases (v33.0 as of Oct 2025)
# https://github.com/protocolbuffers/protobuf/releases
```

**macOS:**
```bash
brew install protobuf
```

### Install Go Plugins

```bash
cd api
make install-tools
```

Or manually:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Generate gRPC Code

```bash
cd api
make generate
```

This will generate `*.pb.go` files in `proto/oak/v1/`.

## Protocol Overview

### Communication Pattern

```
Agent → Server: AgentMessage (stream)
Server → Agent: ServerMessage (stream)
```

### Message Types

**Agent → Server:**
- `AgentRegistration` - Register with server on connection
- `Heartbeat` - Periodic keepalive (every 30s)
- `MetricsReport` - Flink job metrics
- `EventReport` - Important events (job failures, scaling, etc.)
- `CommandResult` - Response to server commands

**Server → Agent:**
- `RegistrationAck` - Acknowledge registration
- `Command` - Execute action (scale, savepoint, cancel, etc.)
- `ConfigUpdate` - Update agent configuration

### Authentication Flow

1. Agent registers via HTTPS (port 8080) with API key
2. Server issues client certificate
3. Agent connects gRPC (port 9090) with mTLS
4. Bidirectional streaming begins

## Usage Example

### Server Side

```go
import oakv1 "github.com/oakproject-flink/oak-flink/api/proto/oak/v1"

// Implement the service
type oakServer struct {
	oakv1.UnimplementedOakServiceServer
}

func (s *oakServer) AgentStream(stream oakv1.OakService_AgentStreamServer) error {
	// Handle bidirectional stream
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		// Process message
		switch payload := msg.Payload.(type) {
		case *oakv1.AgentMessage_Registration:
			// Handle registration
		case *oakv1.AgentMessage_Metrics:
			// Handle metrics
		}

		// Send response
		err = stream.Send(&oakv1.ServerMessage{
			// ...
		})
	}
}
```

### Client Side (Agent)

```go
import oakv1 "github.com/oakproject-flink/oak-flink/api/proto/oak/v1"

// Connect to server
conn, err := grpc.Dial("oak-server:9090",
	grpc.WithTransportCredentials(creds))
client := oakv1.NewOakServiceClient(conn)

// Start bidirectional stream
stream, err := client.AgentStream(context.Background())

// Send registration
stream.Send(&oakv1.AgentMessage{
	Payload: &oakv1.AgentMessage_Registration{
		Registration: &oakv1.AgentRegistration{
			ClusterId:   "prod-cluster-1",
			ClusterName: "Production Cluster",
			// ...
		},
	},
})

// Receive messages from server
go func() {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return
		}
		// Handle server message
	}
}()

// Send heartbeats
ticker := time.NewTicker(30 * time.Second)
for range ticker.C {
	stream.Send(&oakv1.AgentMessage{
		Payload: &oakv1.AgentMessage_Heartbeat{
			Heartbeat: &oakv1.Heartbeat{
				// ...
			},
		},
	})
}
```

## Makefile Commands

```bash
# Generate gRPC code
make generate

# Install required tools
make install-tools

# Clean generated files
make clean
```

## Versioning

The proto files are versioned under `oak/v1`. Future breaking changes will create new versions (`oak/v2`, etc.) to maintain backward compatibility.