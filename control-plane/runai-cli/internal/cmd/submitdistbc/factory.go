/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submitdistbc

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
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

	workloadSpec, err := f.getWorkloadSpec(framework, workloadType)
	if err != nil {
		return nil, err
	}

	return NewCommand(framework, workloadSpec, nil, commandText, f.instance, f.cmdHelper), nil
}

func (f *Factory) getWorkloadSpec(framework api.Framework, workloadType api.WorkloadType) (*spec.DistBCSpec, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	switch workloadType {
	case api.Distributed:
		return spec.NewDistributedBCSpec(framework), nil
	default:
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
}

func (f *Factory) getCommandTexts(framework api.Framework, workloadType api.WorkloadType) (cmd.CommandTexts, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	if (framework == api.Runai && workloadType == api.Workspace) ||
		(framework == api.Knative && workloadType == api.Inference) {
		return NewWorkloadOnlyText(workloadType), nil
	}

	return NewFrameworkText(workloadType, framework), nil
}

func (f *Factory) IsAliasSupported(_ api.Framework, _ api.WorkloadType) bool {
	return false
}

func (f *Factory) GetAliasCommand(framework api.Framework, workload api.WorkloadType) ([]cmd.ActionCommand, error) {
	return nil, cmd.NewUnsupportedAliasError(framework, workload)
}
