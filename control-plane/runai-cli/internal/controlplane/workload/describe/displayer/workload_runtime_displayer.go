package displayer

// Type definitions
// WorkloadRuntimeConfig represents runtime configuration for general workloads,
// distributed inference leaders, and workers.
type WorkloadRuntimeConfig struct {
	// Container Configuration
	Image                *string           `json:"image,omitempty"`
	ImagePullPolicy      *string           `json:"imagePullPolicy,omitempty"`
	ImagePullSecrets     []string          `json:"imagePullSecrets,omitempty"`
	Command              *string           `json:"command,omitempty"`
	Arguments            *string           `json:"arguments,omitempty"`
	WorkingDir           *string           `json:"workingDir,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`

	// Scheduling Configuration
	NodeType    *string  `json:"nodeType,omitempty"`
	Tolerations []string `json:"tolerations,omitempty"`

	// Security Configuration
	RunAsUser    *int64   `json:"runAsUser,omitempty"`
	RunAsGroup   *int64   `json:"runAsGroup,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`

	// Health Configuration
	ReadinessProbe *string `json:"readinessProbe,omitempty"`

	// Resources (optional - may be nil for general runtime if Compute/Storage are top-level)
	Compute *Compute `json:"compute,omitempty"`
	Storage *Storage `json:"storage,omitempty"`

	// Metadata (optional - used by Leader/Worker for pod labels)
	Labels map[string]string `json:"labels,omitempty"`

	// Workload-level fields (optional - not used by Leader/Worker)
	Preemptible *bool `json:"preemptible,omitempty"`
}
