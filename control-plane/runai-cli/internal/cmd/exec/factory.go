package exec

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/completion"
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

	completionFn, err := f.completionFactory.GetCompletionFn(framework, workloadType)
	if err != nil {
		return nil, err
	}

	return NewExecCommand(framework, nil, workloadType, commandText, f.instance, f.cmdHelper, completionFn, nil, nil), nil
}

func (f *Factory) GetAliasCommand(framework api.Framework, workloadType api.WorkloadType) ([]cmd.ActionCommand, error) {
	if !f.IsAliasSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedAliasError(framework, workloadType)
	}

	commandText, err := f.getCommandAliasTexts(framework, workloadType)
	if err != nil {
		return nil, err
	}

	completionFn, err := f.completionFactory.GetCompletionFn(framework, workloadType)
	if err != nil {
		return nil, err
	}

	return []cmd.ActionCommand{NewBashCommand(framework, nil, workloadType, commandText, f.instance, f.cmdHelper, completionFn)}, nil
}

func (f *Factory) getCommandTexts(framework api.Framework, workloadType api.WorkloadType) (cmd.CommandTexts, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	if framework == api.Runai && workloadType == api.Workspace {
		return NewWorkloadOnlyExecTexts(workloadType), nil
	}
	if framework == api.Knative && workloadType == api.Inference {
		return NewInferenceExecTexts(), nil
	}
	if framework == api.LWS && workloadType == api.DistributedInference {
		return NewDistributedInferenceExecTexts(), nil
	}
	if framework == api.NimService {
		return NewNimInferenceExecTexts(), nil
	}
	return NewFrameworkExecTexts(workloadType, framework), nil
}

func (f *Factory) IsAliasSupported(framework api.Framework, workloadType api.WorkloadType) bool {
	return f.IsSupported(framework, workloadType)
}

func (f *Factory) getCommandAliasTexts(framework api.Framework, workloadType api.WorkloadType) (cmd.CommandTexts, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	if framework == api.Runai && workloadType == api.Workspace {
		return NewWorkloadOnlyBashTexts(workloadType), nil
	}
	if framework == api.Knative && workloadType == api.Inference {
		return NewInferenceBashTexts(), nil
	}
	if framework == api.LWS && workloadType == api.DistributedInference {
		return NewDistributedInferenceBashTexts(), nil
	}
	if framework == api.NimService {
		return NewNimInferenceBashTexts(), nil
	}
	return NewFrameworkBashTexts(workloadType, framework), nil
}
