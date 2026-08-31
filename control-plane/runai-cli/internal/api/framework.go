/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"fmt"
	"strings"

	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

type Framework int

const (
	UndefinedFramework Framework = iota
	Runai
	MPI
	PyTorch
	Jax
	TF
	XGBoost
	Knative
	LWS
	AnyDistributed
	NimService
)

const (
	// EmptyFrameworkName is the name of the empty framework
	EmptyFrameworkName = ""
	//StandardFrameworkName is the name of the runai framework
	StandardFrameworkName = "standard"
	// RunaiFrameworkName is the name of the Run:AI framework
	RunaiFrameworkName = "runai"
	// MPIFrameworkName is the name of the MPI framework
	MPIFrameworkName = "mpi"
	// PyTorchFrameworkName is the name of the PyTorch framework
	PyTorchFrameworkName = "pytorch"
	// JaxFrameworkName is the name of the JAX framework
	JaxFrameworkName = "jax"
	// TensorFlowFrameworkName is the name of the TensorFlow framework
	TensorFlowFrameworkName = "tf"
	// XGBoostFrameworkName is the name of the XGBoost framework
	XGBoostFrameworkName = "xgboost"
	// KnativeFrameworkName is the name of the Knative framework
	KnativeFrameworkName = "knative"
	// LWSFrameworkName is the name of the LeaderWorkerSet framework
	LWSFrameworkName = "lws"
	// UndefinedFrameworkName is the name of the undefined framework
	UndefinedFrameworkName = "undefined"
	// AnyDistributedFrameworkName is the name of any distributed framework
	AnyDistributedFrameworkName = "any"
	// NIMFrameworkName is the name of the NIM framework
	NIMFrameworkName        = "nim"
	NimServiceFrameworkName = "nimservice"

	// ClusterApiFrameworkExternal is the framework path used in cluster API for external workloads (e.g. NIM)
	ClusterApiFrameworkExternal = "external"

	// ClusterAPINimServiceNamePrefix is the cluster API workload name prefix for NIM (exec, logs, port-forward).
	ClusterAPINimServiceNamePrefix = "ew-nimservice-"
)

func (f Framework) String() string {
	switch {
	case f < Runai || f > NimService:
		return UndefinedFrameworkName
	default:
		return [...]string{
			UndefinedFrameworkName, RunaiFrameworkName, MPIFrameworkName, PyTorchFrameworkName, JaxFrameworkName,
			TensorFlowFrameworkName, XGBoostFrameworkName, KnativeFrameworkName, LWSFrameworkName, AnyDistributedFrameworkName, NimServiceFrameworkName,
		}[f]
	}
}

func GetWorkloadFramework() []string {
	return []string{
		RunaiFrameworkName, KnativeFrameworkName, LWSFrameworkName, StandardFrameworkName,
		MPIFrameworkName, PyTorchFrameworkName, JaxFrameworkName, TensorFlowFrameworkName,
		XGBoostFrameworkName, AnyDistributedFrameworkName, NimServiceFrameworkName,
	}
}

func (f Framework) CLIType() string {
	switch f {
	case Runai:
		return strings.ToLower(StandardFrameworkName)
	case NimService:
		return strings.ToLower(NIMFrameworkName)
	default:
		return strings.ToLower(f.String())
	}
}

// ClusterApiFramework returns the framework name to use in cluster API URLs
func (f Framework) ClusterApiFramework() string {
	switch f {
	case LWS:
		return RunaiFrameworkName
	case NimService:
		return ClusterApiFrameworkExternal
	default:
		return f.String()
	}
}

// ClusterAPIIdentifiers returns the workload name and framework to use in cluster API URLs (exec, logs, port-forward).
// For NimService the cluster resource is named ew-nimservice-<name> and the path uses external/external; other frameworks are unchanged.
func (f Framework) ClusterAPIIdentifiers(userName string) (clusterName string, clusterFramework Framework) {
	switch f {
	case NimService:
		return ClusterAPINimServiceNamePrefix + userName, NimService
	default:
		return userName, f
	}
}

func (f Framework) JobKind() string {
	switch f {
	case Runai:
		return clusterconstants.KindRunaiJob
	case MPI:
		return clusterconstants.KindMpiJob
	case PyTorch:
		return clusterconstants.KindPyTorchJob
	case Jax:
		return clusterconstants.KindJAXJob
	case TF:
		return clusterconstants.KindTFJob
	case XGBoost:
		return clusterconstants.KindXGBoostJob
	case NimService:
		return clusterconstants.KindNIMService
	default:
		return UndefinedFrameworkName
	}
}

func ParseWorkloadFramework(s string) (Framework, error) {
	switch strings.ToLower(s) {
	case RunaiFrameworkName, StandardFrameworkName:
		return Runai, nil
	case KnativeFrameworkName:
		return Knative, nil
	case LWSFrameworkName:
		return LWS, nil
	case MPIFrameworkName:
		return MPI, nil
	case PyTorchFrameworkName:
		return PyTorch, nil
	case JaxFrameworkName:
		return Jax, nil
	case TensorFlowFrameworkName:
		return TF, nil
	case XGBoostFrameworkName:
		return XGBoost, nil
	case EmptyFrameworkName, AnyDistributedFrameworkName:
		return AnyDistributed, nil
	case NIMFrameworkName, NimServiceFrameworkName:
		return NimService, nil
	default:
		return -1, fmt.Errorf("unknown workload framework: %s", s)
	}
}
