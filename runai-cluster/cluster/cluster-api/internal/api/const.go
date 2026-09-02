/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"fmt"
	"net/http"
	"strings"

	clusterCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

type WorkloadType string

const (
	Workspace            WorkloadType = "workspace"
	Training             WorkloadType = "training"
	Distributed          WorkloadType = "distributed"
	Inference            WorkloadType = "inference"
	DistributedInference WorkloadType = "distributedinference"
	External             WorkloadType = "external"

	ClusterAPIBaseUrl = "/cluster-api/api/v1/{projectName}/workloads/{workloadType}/{workloadFramework}/{workloadName}"
)

var BasePathSplitterSize = len(strings.Split(ClusterAPIBaseUrl, "/")) + 1

func (w WorkloadType) String() string {
	return string(w)
}

func (w WorkloadType) ToK8sResource() string {
	switch w {
	case Workspace:
		return clusterCommon.KindInteractiveWorkload

	case Training:
		return clusterCommon.KindTrainingWorkload

	case Distributed:
		return clusterCommon.KindDistributedWorkload

	case Inference:
		return clusterCommon.KindInferenceWorkload

	case DistributedInference:
		return clusterCommon.KindDistributedInferenceWorkload
	}

	return ""
}

func ValidateWorkloadType(workloadType string) bool {
	switch WorkloadType(workloadType) {
	case Workspace, Training, Distributed, Inference, DistributedInference, External:
		return true
	}

	return false
}

const (
	LogsStr         = "logs"
	LogsDownloadStr = "logs/download"
	ExecStr         = "exec"
	AttachStr       = "attach"
	PortForwardStr  = "port-forward"
	StatusStr       = "status"
)

type WorkloadAction int

const (
	PortForward WorkloadAction = iota
	Logs
	LogsDownload
	Exec
	Attach
	Status
)

func (a WorkloadAction) String() string {
	switch a {
	case PortForward:
		return PortForwardStr

	case Logs:
		return LogsStr

	case Exec:
		return ExecStr

	case Attach:
		return AttachStr

	case Status:
		return StatusStr

	default:
		return ""
	}
}

func GetEndpointPath(r *http.Request) string {
	pathArray := strings.SplitAfterN(r.URL.Path, "/", BasePathSplitterSize)
	if len(pathArray) != BasePathSplitterSize {
		return ""
	}

	return pathArray[BasePathSplitterSize-1]
}

func FromStringAction(action string) WorkloadAction {
	switch action {
	case PortForwardStr:
		return PortForward

	case LogsStr:
		return Logs

	case LogsDownloadStr:
		return LogsDownload

	case ExecStr:
		return Exec

	case AttachStr:
		return Attach

	case StatusStr:
		return Status
	}

	return -1
}

type JobType string

const (
	Unknown                 JobType = "unknown"
	RunaiJob                JobType = "runai-job"
	InferenceJob            JobType = "inference-job"
	DistributedInferenceJob JobType = "distributedinference-job"
	MPIJob                  JobType = "mpi-job"
	PyTorchJob              JobType = "pytorch-job"
	TFJob                   JobType = "tf-job"
	XGBoostJob              JobType = "xgboost-job"
	JAXJob                  JobType = "jax-job"
	ExternalJob             JobType = "external-job"
)

func (j JobType) String() string {
	return string(j)
}

// TODO: consolidate with cli code
type WorkloadFramework string

func (f WorkloadFramework) String() string {
	return string(f)
}

const (
	RunaiFramework          WorkloadFramework = RunaiFrameworkName
	KnativeFramework        WorkloadFramework = KnativeFrameworkName
	LwsFramework            WorkloadFramework = LwsFrameworkName
	MPIFramework            WorkloadFramework = MPIFrameworkName
	PyTorchFramework        WorkloadFramework = PyTorchFrameworkName
	TFFramework             WorkloadFramework = TensorFlowFrameworkName
	XGBoostFramework        WorkloadFramework = XGBoostFrameworkName
	AnyDistributedFramework WorkloadFramework = AnyDistributedFrameworkName
	JAXFramework            WorkloadFramework = JAXFrameworkName
	ExternalFramework       WorkloadFramework = ExternalFrameworkName
)

const (
	// RunaiFrameworkName is the name of the Run:AI framework
	RunaiFrameworkName = "runai"
	// KnativeFrameworkName is the name of the Knative framework
	KnativeFrameworkName = "knative"
	// LwsFrameworkName is the name of the Leader Worker Set framework
	LwsFrameworkName = "lws"
	// MPIFrameworkName is the name of the MPI framework
	MPIFrameworkName = "mpi"
	// PyTorchFrameworkName is the name of the PyTorch framework
	PyTorchFrameworkName = "pytorch"
	// TensorFlowFrameworkName is the name of the TensorFlow framework
	TensorFlowFrameworkName = "tf"
	// XGBoostFrameworkName is the name of the XGBoost framework
	XGBoostFrameworkName = "xgboost"
	// JAXFrameworkName is the name of the JAX framework.
	JAXFrameworkName = "jax"
	// AnyDistributedFrameworkName is the name of the any distributed framework
	AnyDistributedFrameworkName = "any"
	// ExternalFrameworkName is the name of the external framework
	ExternalFrameworkName = "external"
)

func IsValidWorkloadFramework(workloadFramework string) bool {
	switch WorkloadFramework(strings.ToLower(workloadFramework)) {
	case RunaiFramework, MPIFramework, PyTorchFramework, TFFramework,
		XGBoostFramework, JAXFramework, AnyDistributedFrameworkName,
		KnativeFramework, LwsFramework, ExternalFrameworkName:
		return true
	}

	return false
}

func (f WorkloadFramework) JobType() (JobType, error) {
	switch f {
	case RunaiFramework:
		return RunaiJob, nil

	case MPIFramework:
		return MPIJob, nil

	case PyTorchFramework:
		return PyTorchJob, nil

	case TFFramework:
		return TFJob, nil

	case XGBoostFramework:
		return XGBoostJob, nil

	case JAXFrameworkName:
		return JAXJob, nil

	case KnativeFramework:
		return InferenceJob, nil

	case ExternalFramework:
		return ExternalJob, nil

	default:
		return "", fmt.Errorf("unsupported workload framework %s", f)
	}
}

const (
	MasterWorkerIndex = "0"
)
