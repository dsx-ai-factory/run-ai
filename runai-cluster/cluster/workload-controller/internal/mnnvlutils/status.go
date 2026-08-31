/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvlutils

import (
	"context"
	"errors"
	"fmt"
	gpu_validator "github.com/run-ai/runai/runai-cluster/cluster/workload-common/gpu-validator"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	nodepoolbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders/nodepool"
	apijobs "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errorMsgNodePoolsUnsupported = "%w: '%s' do not support %s in %s mode"
)

var (
	SpecInvalidForRequiredErr         = fmt.Errorf("the provided spec cannot support %s in %s mode", mnnvl.UserFacingFeatureName, mnnvl.ModeRequired)
	ErrNodePoolsUnsupportedCapability = errors.New("some of the provided node pools do not support MNNVL capability")

	defaultStatus = Status{
		WorkerMode: mnnvl.ModeNone,
		MasterMode: mnnvl.ModeNone,
	}
)

type Status struct {
	MasterMode mnnvl.Mode
	WorkerMode mnnvl.Mode
}

func (s *Status) IsNetworkAccelerated() bool {
	return s.MasterMode != mnnvl.ModeNone || s.WorkerMode != mnnvl.ModeNone
}

func GenerateStatus(ctx context.Context, k8sClient client.Client, workload *apijobs.JobWorkload) (Status, error) {
	// starting with the default status - both None
	status := defaultStatus

	if workload.Kind != constants.KindDistributedWorkload {
		return status, nil
	}

	featureLabelValue, exist := workload.Labels[mnnvl.FeatureLabelKey]
	if !exist || featureLabelValue == mnnvl.None {
		return status, nil
	}

	// convert the feature label value to mode
	desiredMode, err := mnnvl.GetMode(featureLabelValue)
	if err != nil {
		return status, err
	}

	// check if the nodepools are valid for the desired mode
	nodepoolsValid, err := areNodepoolsValid(ctx, k8sClient, workload, desiredMode)
	if err != nil {
		return status, err
	}
	if !nodepoolsValid {
		return status, nil
	}

	// calculate the mode for the worker and store it in the status
	workerMode := calculateMode(desiredMode, workload.Spec)
	status.WorkerMode = workerMode

	// if workload has no master, no need to calculate any further
	if hasMaster(workload) {
		// if the master has its own spec, need to calculate the mode for it
		// else, the spec is same as the worker - use the mode we calculated for it
		if workload.MasterSpec != nil {
			masterMode := calculateMode(desiredMode, *workload.MasterSpec)
			status.MasterMode = masterMode
		} else {
			status.MasterMode = workerMode
		}
	}

	// if the mode is Required but neither spec is enabled for acceleration it's an error
	if desiredMode == mnnvl.ModeRequired && !status.IsNetworkAccelerated() {
		return defaultStatus, SpecInvalidForRequiredErr
	}

	return status, nil
}

func calculateMode(desiredMode mnnvl.Mode, spec apijobs.JobSpec) mnnvl.Mode {
	if desiredMode == mnnvl.ModeNone {
		return desiredMode
	}

	// Mode cannot be Preferred or Required if no GPUs were requested
	isValid := gpu_validator.HasRequestedGpus(&spec.ResourceAllocationFields)
	if !isValid {
		return mnnvl.ModeNone
	}

	return desiredMode
}

func hasMaster(workload *apijobs.JobWorkload) bool {
	return workload.MasterSpec != nil && workload.Spec.NoMaster.Deref() == false
}

// areNodepoolsValid checks if the nodepools support the mode
func areNodepoolsValid(ctx context.Context, k8sClient client.Client, workload *apijobs.JobWorkload, desiredMode mnnvl.Mode) (bool, error) {
	if desiredMode == mnnvl.ModeNone {
		return false, nil
	}

	nodepoolList, err := nodepoolbuilders.GetNodePools(ctx, k8sClient, workload.Spec.NodePools.Deref(), workload.Spec.NodePool.Deref())
	if err != nil {
		return false, err
	}
	unsupported := getUnsupportedNodePools(nodepoolList)

	if len(unsupported) > 0 && desiredMode == mnnvl.ModeRequired {
		return false, fmt.Errorf(errorMsgNodePoolsUnsupported,
			ErrNodePoolsUnsupportedCapability, strings.Join(unsupported, ","), mnnvl.UserFacingFeatureName, desiredMode)
	}

	return len(unsupported) == 0, nil
}
