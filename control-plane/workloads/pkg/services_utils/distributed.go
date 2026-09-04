/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_utils

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	FrameworkNotAvailableMsg   = "distributed framework %s is not available, make sure prerequisites are installed"
	FrameworkNotSuspendableMsg = "distributed framework %s is not suspendable, make sure prerequisites are installed"
	UnknownFrameworkMsg        = "unknown distributed framework %s"
)

// CheckDistributedFrameworkAvailable validates the distributed framework is available in the cluster (using data from cluster-service)
func CheckDistributedFrameworkAvailable(ctx context.Context, clusterId string, framework string) *api.Error {
	parsed, err := parseDistributedFramework(ctx, framework)
	if err != nil {
		return err
	}

	status, apiErr := getDistributedFrameworkStatus(ctx, clusterId, parsed)
	if apiErr != nil {
		return apiErr
	}

	if !status.Available {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf(FrameworkNotAvailableMsg, parsed), nil)
	}

	return nil
}

// CheckDistributedFrameworkSuspendable validates the distributed framework is suspendable (using data from cluster-service)
func CheckDistributedFrameworkSuspendable(ctx context.Context, clusterId string, framework string) *api.Error {
	parsed, err := parseDistributedFramework(ctx, framework)
	if err != nil {
		return err
	}

	status, apiErr := getDistributedFrameworkStatus(ctx, clusterId, parsed)
	if apiErr != nil {
		return apiErr
	}

	if !status.Available {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf(FrameworkNotAvailableMsg, parsed), nil)
	}

	if !status.Suspendable {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf(FrameworkNotSuspendableMsg, parsed), nil)
	}

	return nil
}

func getDistributedFrameworkStatus(ctx context.Context, clusterId string, framework workloadsApi.DistributedFramework) (*checker.DistributedFrameworkStatus, *api.Error) {
	clusterChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	clusterConfig, err := clusterChecker.ClusterConfiguration(ctx)
	if err != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, err.Error(), nil)
	}

	targetStatus, ok := clusterConfig.DistributedFrameworksStatus[framework]
	if !ok {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf(UnknownFrameworkMsg, framework), nil)
	}

	return &targetStatus, nil
}

// parseDistributedFramework is parsing all variation possible of distributed framework types to the expected format
// can be removed once we store the framework as enum in the db
func parseDistributedFramework(ctx context.Context, framework string) (workloadsApi.DistributedFramework, *api.Error) {
	switch strings.ToLower(framework) {
	case strings.ToLower(string(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)), "mpijob", "mpi":
		return workloadsApi.DISTRIBUTEDFRAMEWORK_MPI, nil
	case strings.ToLower(string(workloadsApi.DISTRIBUTEDFRAMEWORK_TF)), "tfjob", "tf", "tensorflow", "tensorflowjob":
		return workloadsApi.DISTRIBUTEDFRAMEWORK_TF, nil
	case strings.ToLower(string(workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH)), "pytorchjob", "pytorch":
		return workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH, nil
	case strings.ToLower(string(workloadsApi.DISTRIBUTEDFRAMEWORK_JAX)), "jaxjob", "jax":
		return workloadsApi.DISTRIBUTEDFRAMEWORK_JAX, nil
	case strings.ToLower(string(workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST)), "xgboostjob", "xgboost":
		return workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST, nil
	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("unable to parse distributed framework %s", framework)
		return "", api.NewError(http.StatusInternalServerError, fmt.Sprintf(UnknownFrameworkMsg, framework), nil)
	}
}
