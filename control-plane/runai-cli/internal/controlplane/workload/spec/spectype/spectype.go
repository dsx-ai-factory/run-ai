package spectype

type Workload string

const (
	Workspace            Workload = "workspace"
	Inference            Workload = "inference"
	Training             Workload = "training"
	Submit               Workload = "submit"
	TrainingDistributed  Workload = "training-distributed"
	InferenceDistributed Workload = "inference-distributed"
	NimService           Workload = "inference-nim"
)

// DisplayName returns the user-friendly display name for the workload type
func (w Workload) DisplayName() string {
	switch w {
	case TrainingDistributed, InferenceDistributed:
		return "distributed"
	case NimService:
		return "nim"
	default:
		return string(w)
	}
}

// APIString returns the API WorkloadType string for this workload type
func (w Workload) APIString() string {
	switch w {
	case Workspace:
		return "Workspace"
	case Training:
		return "Training"
	case TrainingDistributed:
		return "Distributed"
	case Inference:
		return "Inference"
	case InferenceDistributed:
		return "DistributedInference"
	case NimService:
		return "NIMService"
	default:
		return string(w)
	}
}

// APIStringToDisplayName converts an API workload type string to a user-friendly display name
func APIStringToDisplayName(apiString string) string {
	switch apiString {
	case "Workspace":
		return "workspace"
	case "Training":
		return "training"
	case "Distributed":
		return "distributed training"
	case "Inference":
		return "inference"
	case "DistributedInference":
		return "distributed inference"
	case "NIMService":
		return "nim service"
	default:
		return apiString
	}
}
