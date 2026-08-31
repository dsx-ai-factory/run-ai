/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package delete

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/completion"
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
	instance          *config.Instance
	cmdHelper         helper.Cmd
	completionFactory completion.Factory
}

func NewCommandFactory(instance *config.Instance, cmdHelper helper.Cmd, completionFactory completion.Factory) *Factory {
	return &Factory{
		instance:          instance,
		cmdHelper:         cmdHelper,
		completionFactory: completionFactory,
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

	operationFn, err := f.getOperationFunc(framework, workloadType)
	if err != nil {
		return nil, err
	}

	searchFn, err := cmd.GetSearchFunc(framework, workloadType)
	if err != nil {
		return nil, err
	}

	completionFn, err := f.completionFactory.GetCompletionFn(framework, workloadType)
	if err != nil {
		return nil, err
	}

	return NewDeleteCommand(framework, nil, commandText, f.instance,
		f.cmdHelper, operationFn, searchFn, workloadType, completionFn), nil
}

func (f *Factory) IsAliasSupported(_ api.Framework, _ api.WorkloadType) bool {
	return false
}

func (f *Factory) GetAliasCommand(framework api.Framework, workload api.WorkloadType) ([]cmd.ActionCommand, error) {
	return nil, cmd.NewUnsupportedAliasError(framework, workload)
}

func (f *Factory) getCommandTexts(framework api.Framework, workloadType api.WorkloadType) (cmd.SearchableCommandTexts, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	if framework == api.Runai && workloadType == api.Workspace {
		return NewWorkloadOnlyTexts(workloadType), nil
	}

	if framework == api.Knative && workloadType == api.Inference {
		return NewInferenceTexts(), nil
	}

	if framework == api.LWS && workloadType == api.DistributedInference {
		return NewDistributedInferenceTexts(), nil
	}

	if framework == api.NimService {
		return NewNimInferenceDeleteTexts(), nil
	}

	return NewFrameworkTexts(workloadType, framework), nil
}

func (f *Factory) getOperationFunc(framework api.Framework, workloadType api.WorkloadType) (cmd.RunOperationFn, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	if framework == api.NimService {
		return nim.Delete, nil
	}

	switch workloadType {
	case api.Distributed:
		return distributed.Delete, nil
	case api.DistributedInference:
		return distributedinference.Delete, nil
	case api.Inference:
		return inference.Delete, nil
	case api.Training:
		return training.Delete, nil
	case api.Workspace:
		return workspace.Delete, nil
	default:
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
}
