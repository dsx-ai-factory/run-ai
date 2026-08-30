package api

import (
	"errors"
	"strings"
)

type WorkloadType int

const (
	Undefined WorkloadType = iota
	Workspace
	Inference
	Training
	Distributed
	DistributedInference
)

func (w WorkloadType) CLIType() string {
	if w == Distributed {
		return strings.ToLower(training)
	}
	return strings.ToLower(w.String())
}

// ClusterApiType returns the workload type string used in cluster API URLs
// For distributed inference, the cluster API expects "distributedinference" (no hyphen)
func (w WorkloadType) ClusterApiType() string {
	switch w {
	case DistributedInference:
		return distributedInferenceClusterApi
	default:
		return w.String()
	}
}

const (
	workspace                      = "workspace"
	workspaceCapitalize            = "Workspace"
	training                       = "training"
	trainingCapitalize             = "Training"
	inference                      = "inference"
	inferenceCapitalize            = "Inference"
	distributed                    = "distributed"
	distributedCapitalize          = "Distributed"
	distributedInference           = "distributed inference"
	distributedInferenceCapitalize = "DistributedInference"
	distributedInferenceClusterApi = "distributedinference" // For cluster API URLs
	inferenceNim                   = "inference nim"
	inferenceNimCapitalize         = "InferenceNim"
)

const BackendNimServiceWorkloadType = "NIMService"

func (w WorkloadType) String() string {
	switch w {
	case Workspace:
		return workspace
	case Training:
		return training
	case Distributed:
		return distributed
	case Inference:
		return inference
	case DistributedInference:
		return distributedInference
	default:
		return "unknown"
	}
}

func (w WorkloadType) StringCapitalize() string {
	switch w {
	case Workspace:
		return workspaceCapitalize
	case Training:
		return trainingCapitalize
	case Inference:
		return inferenceCapitalize
	case Distributed:
		return distributedCapitalize
	case DistributedInference:
		return distributedInferenceCapitalize
	default:
		return "Unknown"
	}
}

func GetWorkloadTypes() []string {
	return []string{workspace, training, inference}
}

func (w WorkloadType) QueryParam() string {
	switch w {
	case Workspace:
		return "Workspace"
	case Training:
		return "Training||Distributed"
	case Distributed:
		return "Distributed"
	case Inference:
		return "Inference"
	case DistributedInference:
		return "DistributedInference"
	default:
		return "Unknown"
	}
}

func ParseWorkloadType(s string) (WorkloadType, error) {
	switch s {
	case workspace:
		return Workspace, nil
	case training:
		return Training, nil
	case inference, inferenceNim:
		return Inference, nil
	default:
		return -1, errors.New("unknown workload type")
	}
}

func (w WorkloadType) DefaultFramework() Framework {
	switch w {
	case Workspace, Training:
		return Runai
	case Inference:
		return Knative
	case Distributed:
		return AnyDistributed
	case DistributedInference:
		return LWS
	default:
		return UndefinedFramework
	}
}
