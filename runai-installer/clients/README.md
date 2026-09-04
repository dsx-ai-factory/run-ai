# Run:ai API Client

This module contains an auto-generated Go client for the Run:ai Public API.

## Overview

The client is generated from the **runai-public.yaml** OpenAPI specification, which is the official Run:ai public API.

**Source:** `https://github.com/dsx-ai-factory/run-ai/blob/master/control-plane/infra/openapi/specs/runai-public.yaml`

## Available APIs

The generated client provides access to all Run:ai public APIs, including:

- **Clusters API** - Cluster management operations
- **Node Pools API** - Node pool configuration
- **Nodes API** - Node management and telemetry
- **Projects API** - Create, read, update, and delete projects
- **Departments API** - Manage departments
- **Authorization API** - Roles, access rules, permissions, and permission sets
- **Workloads API** - Workload management (trainings, workspaces, inferences)
- **Assets API** - Compute, credentials, environments, datasources
- **Policy API** - Workload policies
- **And more...**

## Usage

### Installation

Add the clients module to your `go.mod`:

```go
require github.com/dsx-ai-factory/run-ai/runai-installer/clients v0.0.0
replace github.com/dsx-ai-factory/run-ai/runai-installer/clients => ../clients
```

### Example: Using the API Client

The clients are organized into separate packages:
- `github.com/dsx-ai-factory/run-ai/runai-installer/clients/runai/authorization` - Roles, access rules, permissions
- `github.com/dsx-ai-factory/run-ai/runai-installer/clients/runai/cluster` - Clusters, node pools, nodes
- `github.com/dsx-ai-factory/run-ai/runai-installer/clients/runai/orgunit` - Projects, departments

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/dsx-ai-factory/run-ai/runai-installer/clients/runai/cluster"
)

func main() {
    // Create configuration
    cfg := cluster.NewConfiguration()
    cfg.Host = "app.run.ai"
    cfg.Scheme = "https"
    cfg.AddDefaultHeader("Authorization", "Bearer <your-token>")
    
    // Create client
    client := cluster.NewAPIClient(cfg)
    
    // Example: List clusters
    ctx := context.Background()
    
    clusters, resp, err := client.ClustersAPI.GetClusters(ctx).Execute()
    if err != nil {
        fmt.Printf("Error listing clusters: %v\n", err)
        return
    }
    
    fmt.Printf("Found %d clusters\n", len(clusters))
}
```

## Generating the Client

### Prerequisites

- Docker (for running openapi-generator)
- Make
- GitHub access to the `dsx-ai-factory/run-ai` repository (private)

### Authentication

The OpenAPI spec is fetched from a private GitHub repository. You need to authenticate using one of:

1. **GitHub Token** (recommended for CI):
   ```bash
   export GITHUB_TOKEN=your_github_token
   make generate
   ```

2. **GitHub CLI** (recommended for local development):
   ```bash
   gh auth login
   make generate
   ```

### Generate the Client

```bash
make all
# or
make generate
```

### Fetch Latest OpenAPI Spec

```bash
make fetch-spec
```

### Show All Available API Paths

```bash
make show-paths
```

### Clean Generated Files

```bash
make clean
```

### Validate OpenAPI Spec

```bash
make validate
```

## Configuration Options

Each generated client supports various configuration options:

```go
cfg := cluster.NewConfiguration()  // or authorization.NewConfiguration(), orgunit.NewConfiguration()

// Set the API host
cfg.Host = "app.run.ai"

// Set the scheme (https/http)
cfg.Scheme = "https"

// Set default headers (e.g., authentication)
cfg.AddDefaultHeader("Authorization", "Bearer <token>")

// Set custom HTTP client
cfg.HTTPClient = &http.Client{
    Timeout: 30 * time.Second,
}

// Enable debug mode
cfg.Debug = true
```

## Directory Structure

```
clients/
├── go.mod                    # Go module definition
├── Makefile                  # Generation and utility targets
├── README.md                 # This file
├── pkg/
│   └── clients/              # Shared RunAI client with token rotation
│       └── client.go
├── specs/                    # Downloaded OpenAPI specs (gitignored)
│   ├── authorization-service-api.yaml
│   ├── cluster-service-api.yaml
│   └── org-unit-service-api.yaml
└── runai/
    ├── authorization/        # Authorization service client (roles, access rules)
    │   ├── api_*.go
    │   ├── client.go
    │   ├── configuration.go
    │   └── model_*.go
    ├── cluster/              # Cluster service client (clusters, nodes, node pools)
    │   ├── api_*.go
    │   ├── client.go
    │   ├── configuration.go
    │   └── model_*.go
    └── orgunit/              # Org unit service client (projects, departments)
        ├── api_*.go
        ├── client.go
        ├── configuration.go
        └── model_*.go
```

## API Reference

For full API documentation, see:
- [Run:ai API Documentation](https://api-docs.run.ai/latest/)

## Versioning

The client is versioned together with the runai-installer-project monorepo. The generated code is committed to the repository to ensure reproducible builds.
