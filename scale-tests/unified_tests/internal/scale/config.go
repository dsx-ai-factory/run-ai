package scale

import (
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type LoadMethod string

type ResourceRequirements string

const (
	Small  ResourceRequirements = "small"
	Medium ResourceRequirements = "medium"
	Large  ResourceRequirements = "large"
	Custom ResourceRequirements = "custom"
)

type ClusterConfig struct {
	NodepoolCount        int                  `mapstructure:"nodepool_count"`
	NodesPerPool         int                  `mapstructure:"nodes_per_pool"`
	ControlPlaneURL      string               `mapstructure:"control_plane_url"`
	ControlPlaneUsername string               `mapstructure:"control_plane_username" sensitive:"true"`
	ControlPlanePassword string               `mapstructure:"control_plane_password" sensitive:"true"`
	ClusterID            string               `mapstructure:"cluster_id"`
	ControlPlaneQPS      int                  `mapstructure:"control_plane_qps"`
	KubeQPS              int                  `mapstructure:"kube_qps"`
	OperatorVersion      string               `mapstructure:"operator_version"`
	CpVersion            string               `mapstructure:"cp_version"`
	KwokEnabled          bool                 `mapstructure:"kwok_enabled"`
	ClusterCleanup       bool                 `mapstructure:"cluster_cleanup"`
	BatchSize            int                  `mapstructure:"batch_size"`
	ScaleGroup           string               `mapstructure:"scale_group"`
	Resources            ResourceRequirements `mapstructure:"resources"`
	GpuProduct           []string             `mapstructure:"gpu_product"`
	GPUsPerNode          int                  `mapstructure:"gpus_per_node"`

	// Prometheus configuration
	PrometheusRemoteWriteEnabled  bool   `mapstructure:"prometheus_remote_write_enabled"`
	PrometheusRemoteWriteURL      string `mapstructure:"prometheus_remote_write_url" sensitive:"true"`
	PrometheusRemoteWriteUsername string `mapstructure:"prometheus_remote_write_username" sensitive:"true"`
	PrometheusRemoteWritePassword string `mapstructure:"prometheus_remote_write_password" sensitive:"true"`

	// Custom resource specifications (only used when Resources="custom")
	// Scheduling Services
	SchedulingServicesCPURequest    string `mapstructure:"scheduling_services_cpu_request"`
	SchedulingServicesCPULimit      string `mapstructure:"scheduling_services_cpu_limit"`
	SchedulingServicesMemoryRequest string `mapstructure:"scheduling_services_memory_request"`
	SchedulingServicesMemoryLimit   string `mapstructure:"scheduling_services_memory_limit"`

	// Workload Services
	WorkloadServicesCPURequest    string `mapstructure:"workload_services_cpu_request"`
	WorkloadServicesCPULimit      string `mapstructure:"workload_services_cpu_limit"`
	WorkloadServicesMemoryRequest string `mapstructure:"workload_services_memory_request"`
	WorkloadServicesMemoryLimit   string `mapstructure:"workload_services_memory_limit"`

	// Sync Services
	SyncServicesCPURequest    string `mapstructure:"sync_services_cpu_request"`
	SyncServicesCPULimit      string `mapstructure:"sync_services_cpu_limit"`
	SyncServicesMemoryRequest string `mapstructure:"sync_services_memory_request"`
	SyncServicesMemoryLimit   string `mapstructure:"sync_services_memory_limit"`

	// Training Operator
	TrainingOperatorCPURequest    string `mapstructure:"training_operator_cpu_request"`
	TrainingOperatorCPULimit      string `mapstructure:"training_operator_cpu_limit"`
	TrainingOperatorMemoryRequest string `mapstructure:"training_operator_memory_request"`
	TrainingOperatorMemoryLimit   string `mapstructure:"training_operator_memory_limit"`

	// Fake GPU Operator
	FakeGPUOperatorCPURequest    string `mapstructure:"fake_gpu_operator_cpu_request"`
	FakeGPUOperatorCPULimit      string `mapstructure:"fake_gpu_operator_cpu_limit"`
	FakeGPUOperatorMemoryRequest string `mapstructure:"fake_gpu_operator_memory_request"`
	FakeGPUOperatorMemoryLimit   string `mapstructure:"fake_gpu_operator_memory_limit"`

	// PostgreSQL
	PostgresCPURequest    string `mapstructure:"postgres_cpu_request"`
	PostgresCPULimit      string `mapstructure:"postgres_cpu_limit"`
	PostgresMemoryRequest string `mapstructure:"postgres_memory_request"`
	PostgresMemoryLimit   string `mapstructure:"postgres_memory_limit"`

	// GitHub workflow information
	GitHubActor       string `mapstructure:"github_actor"`
	GitHubWorkflowURL string `mapstructure:"github_workflow_url"`
}

func (c *ClusterConfig) SetDefaults() {
	c.NodepoolCount = 3
	c.NodesPerPool = 60
	// ControlPlaneURL keeps a harmless local default. Credentials intentionally have
	// NO default and must be supplied at runtime via the CONTROL_PLANE_USERNAME and
	// CONTROL_PLANE_PASSWORD env vars (injected from secrets in CI), so that no
	// credentials are committed to this repository.
	c.ControlPlaneURL = "http://localhost:8080"
	c.ControlPlaneUsername = ""
	c.ControlPlanePassword = ""
	c.ControlPlaneQPS = 20
	c.KubeQPS = 100
	c.OperatorVersion = "dev"
	c.CpVersion = "dev"
	c.KwokEnabled = true
	c.ClusterCleanup = true
	c.BatchSize = 100
	c.ScaleGroup = "manual"
	c.Resources = Small
	c.GPUsPerNode = 8
}

func ConstructConfig() *ClusterConfig {
	clusterConfig, err := config.GetR[*ClusterConfig]()
	if err != nil {
		logging.Fatalf("failed to register config, error: %v. Service is terminating", err)
	}
	return clusterConfig
}
