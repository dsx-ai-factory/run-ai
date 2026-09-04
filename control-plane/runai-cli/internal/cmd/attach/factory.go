/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attach

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
)

// Factory handles the creation of attach commands
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

	completionFn, err := f.completionFactory.GetCompletionFn(framework, workloadType)
	if err != nil {
		return nil, err
	}

	// Implement the NewAttachCommand function or use an appropriate command creation function
	return NewAttachCommand(framework, nil, workloadType, commandText, f.instance, f.cmdHelper, completionFn), nil
}

func (f *Factory) getCommandTexts(framework api.Framework, workloadType api.WorkloadType) (cmd.CommandTexts, error) {
	if f.IsSupported(framework, workloadType) {
		if workloadType == api.Workspace {
			return NewWorkloadOnlyTexts(workloadType), nil
		} else {
			return NewFrameworkTexts(workloadType, framework), nil
		}
	}
	return nil, cmd.NewUnsupportedError(framework, workloadType)
}

func (f *Factory) IsAliasSupported(_ api.Framework, _ api.WorkloadType) bool {
	return false
}

func (f *Factory) GetAliasCommand(framework api.Framework, workload api.WorkloadType) ([]cmd.ActionCommand, error) {
	return nil, cmd.NewUnsupportedAliasError(framework, workload)
}
