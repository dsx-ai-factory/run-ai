/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/inference"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/nim"
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
	case api.Knative:
		return workloadType == api.Inference
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

	workloadSpec, err := f.getWorkloadSpec(framework, workloadType)
	if err != nil {
		return nil, err
	}

	updateWorkloadFn, err := f.getUpdateWorkloadFn(framework, workloadType)
	if err != nil {
		return nil, err
	}

	searchWorkloadFn, err := f.getSearchWorkloadFn(framework, workloadType)
	if err != nil {
		return nil, err
	}

	return NewCommand(framework, workloadSpec, nil, commandText, f.instance, f.cmdHelper, searchWorkloadFn, updateWorkloadFn), nil
}

func (f *Factory) getWorkloadSpec(framework api.Framework, workloadType api.WorkloadType) (*spec.GeneralSpec, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
	switch framework {
	case api.Knative:
		return spec.NewInferenceSpec(), nil
	case api.NimService:
		return spec.NewNimSpec(), nil
	default:
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
}

func (f *Factory) getSearchWorkloadFn(framework api.Framework, workloadType api.WorkloadType) (cmd.SearchWorkloadFn, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
	if framework == api.Knative && workloadType == api.Inference {
		return inference.GetByNameAndProj, nil
	}
	if framework == api.NimService && workloadType == api.Inference {
		return nim.GetByNameAndProj, nil
	}
	return nil, cmd.NewUnsupportedError(framework, workloadType)
}

func (f *Factory) getUpdateWorkloadFn(framework api.Framework, workloadType api.WorkloadType) (cmd.UpdateWorkloadFn, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
	if framework == api.Knative && workloadType == api.Inference {
		return inference.Update, nil
	}
	if framework == api.NimService && workloadType == api.Inference {
		return nim.Update, nil
	}
	return nil, cmd.NewUnsupportedError(framework, workloadType)
}

func (f *Factory) getCommandTexts(framework api.Framework, workloadType api.WorkloadType) (cmd.SearchableCommandTexts, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
	if framework == api.Knative && workloadType == api.Inference {
		return NewWorkloadOnlyText(workloadType), nil
	}
	if framework == api.NimService && workloadType == api.Inference {
		return NewNimUpdateText(), nil
	}
	return nil, cmd.NewUnsupportedError(framework, workloadType)
}

func (f *Factory) IsAliasSupported(_ api.Framework, _ api.WorkloadType) bool {
	return false
}

func (f *Factory) GetAliasCommand(framework api.Framework, workload api.WorkloadType) ([]cmd.ActionCommand, error) {
	return nil, cmd.NewUnsupportedAliasError(framework, workload)
}
