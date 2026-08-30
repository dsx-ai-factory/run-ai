package displayer

const (
	// Field names
	FieldPreemptible               = "Preemptible"
	FieldImageSingular             = "Image"
	FieldImagePlural               = "Images"
	FieldRequestedNodePoolSingular = "Requested Node Pool"
	FieldRequestedNodePoolPlural   = "Requested Node Pools"
	FieldCurrentNodePoolSingular   = "Current Node Pool"
	FieldCurrentNodePoolPlural     = "Current Node Pools"
	FieldImagePullPolicy           = "Image Pull Policy"
	FieldImagePullSecrets          = "Image Pull Secrets" //nolint:gosec // G101: This is a display field name, not a credential
	FieldWorkingDir                = "Working Directory"
	FieldNodeType                  = "Node Type"
	FieldRunAsUser                 = "Run as User"
	FieldRunAsGroup                = "Run as Group"
	FieldCapabilities              = "Capabilities"
	FieldTolerations               = "Tolerations"
	FieldReadinessProbe            = "Readiness Probe"
	FieldCommand                   = "Command"
	FieldArguments                 = "Arguments"
	FieldEnvVars                   = "Environment Variables"
	FieldGPUNetworkAcceleration    = "GPU NW acceleration (MNNVL) used"
)
