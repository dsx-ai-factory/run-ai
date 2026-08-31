/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributed"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributedinference"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/inference"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/nim"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/training"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workspace"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
)

type Factory struct {
	instance  *config.Instance
	cmdHelper helper.Cmd
}

func NewCommandFactory(instance *config.Instance, cmdHelper helper.Cmd) *Factory {
	return &Factory{
		instance:  instance,
		cmdHelper: cmdHelper,
	}
}

func (f *Factory) IsSupported(framework api.Framework, workloadType api.WorkloadType) bool {
	switch framework {
	case api.MPI, api.XGBoost, api.TF, api.PyTorch, api.Jax:
		return workloadType == api.Distributed
	case api.Runai:
		return workloadType == api.Training || workloadType == api.Workspace
	case api.Knative:
		return workloadType == api.Inference
	case api.LWS:
		return workloadType == api.DistributedInference
	case api.NimService:
		return workloadType == api.Inference
	default:
		return false
	}
}

func (f *Factory) GetCommand(framework api.Framework, workloadType api.WorkloadType) (cmd.ActionCommand, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	commandText, err := f.getCommandTexts(framework, workloadType)
	if err != nil {
		return nil, err
	}

	listFn, err := f.getListFunc(framework, workloadType)
	if err != nil {
		return nil, err
	}

	return NewListCommand(framework, nil, f.instance, commandText, f.cmdHelper, listFn, workloadType), nil
}

func (f *Factory) getCommandTexts(framework api.Framework, workloadType api.WorkloadType) (cmd.CommandTexts, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	if framework == api.Runai && workloadType == api.Workspace {
		return NewWorkloadOnlyListTexts(workloadType), nil
	}
	if framework == api.Knative && workloadType == api.Inference {
		return NewInferenceStandardListTexts(workloadType), nil
	}
	if framework == api.LWS && workloadType == api.DistributedInference {
		return NewInferenceDistributedListTexts(workloadType), nil
	}
	if framework == api.NimService {
		return NewInferenceNimListTexts(workloadType), nil
	}

	return NewFrameworkListTexts(workloadType, framework), nil
}

func (f *Factory) getListFunc(framework api.Framework, workloadType api.WorkloadType) (cmd.ListWorkloadsFn, error) {
	if framework == api.NimService {
		return nim.List, nil
	}

	switch workloadType {
	case api.Distributed:
		return distributed.List, nil
	case api.DistributedInference:
		return distributedinference.List, nil
	case api.Inference:
		return inference.List, nil
	case api.Training:
		return training.List, nil
	case api.Workspace:
		return workspace.List, nil
	default:
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
}

func (f *Factory) IsAliasSupported(_ api.Framework, _ api.WorkloadType) bool {
	return false
}

func (f *Factory) GetAliasCommand(framework api.Framework, workload api.WorkloadType) ([]cmd.ActionCommand, error) {
	return nil, cmd.NewUnsupportedAliasError(framework, workload)
}
