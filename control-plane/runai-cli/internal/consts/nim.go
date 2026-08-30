package consts

const (
	AttributeNameRemove = "remove"
	AttributeNameForce  = "force"

	AttributeUsageNimWorkers = "Number of worker nodes for multi-node NIM. Cannot be used with autoscaling (--min-replicas/--max-replicas)"

	AttributeUsageRemove = "Remove entire field sections. Format: autoscaling|environmentVariables|labels|annotations|tolerations|probes|compute|multiNode|modelStore"

	AttributeUsageForceScale = "Skip the confirmation prompt when scaling a workload that has autoscaling enabled"

	RemovableFieldAutoscaling          = "autoscaling"
	RemovableFieldReplicas             = "replicas"
	RemovableFieldEnvironmentVariables = "environmentVariables"
	RemovableFieldLabels               = "labels"
	RemovableFieldAnnotations          = "annotations"
	RemovableFieldTolerations          = "tolerations"
	RemovableFieldProbes               = "probes"
	RemovableFieldCompute              = "compute"
	RemovableFieldMultiNode            = "multiNode"
	RemovableFieldModelStore           = "modelStore"
)
