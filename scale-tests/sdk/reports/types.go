package reports

import (
	"time"
)

type SuiteReport struct {
	ID                 string                       `json:"id"`
	MetaData           map[string]string            `json:"metadata"`
	CreatedAt          time.Time                    `json:"created_at"`
	PerformanceMetrics map[string]PerformanceMetric `json:"performance_metrics"`
	Environment        Environment                  `json:"environment"`
}

type PerformanceMetric interface{}

type Environment struct {
	OperatorVersion       string                 `json:"operator_version"`
	ControlPlaneVersion   string                 `json:"control_plane_version"`
	InfrastructureDetails *InfrastructureDetails `json:"infrastructure_details,omitempty"`
}

type InfrastructureDetails struct {
	KubernetesVersion      string          `json:"kubernetes_version"`
	KubernetesDistribution string          `json:"kubernetes_distribution"`
	TotalNodes             int             `json:"total_nodes"`
	RealNodeCount          int             `json:"real_node_count"`
	FakeNodeCount          int             `json:"fake_node_count"`
	NodepoolCount          int             `json:"nodepool_count"`
	NodesPerPool           int             `json:"nodes_per_pool"`
	KwokEnabled            bool            `json:"kwok_enabled"`
	GpuProducts            []string        `json:"gpu_products"`
	ResourceTier           string          `json:"resource_tier"`
	RunaiResources         *RunaiResources `json:"runai_resources,omitempty"`
}

type RunaiResources struct {
	SchedulingServices *ComponentResources `json:"scheduling_services,omitempty"`
	SyncServices       *ComponentResources `json:"sync_services,omitempty"`
	WorkloadServices   *ComponentResources `json:"workload_services,omitempty"`
	WorkloadController *ControllerConfig   `json:"workload_controller,omitempty"`
}

type ComponentResources struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
}

type ControllerConfig struct {
	MaxConcurrentReconciles int                 `json:"max_concurrent_reconciles"`
	Resources               *ComponentResources `json:"resources,omitempty"`
}
