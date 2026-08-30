package api

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
)

type Workload struct {
	Type      api.WorkloadType
	Name      string
	Project   string
	ClusterID string
	ID        string
	Framework api.Framework
}
