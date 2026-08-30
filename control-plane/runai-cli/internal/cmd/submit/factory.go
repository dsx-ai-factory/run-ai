package submit

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
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

	workloadSpec, err := f.getWorkloadSpec(framework, workloadType)
	if err != nil {
		return nil, err
	}

	return NewCommand(framework, workloadSpec, nil, commandText, f.instance, f.cmdHelper), nil
}

func (f *Factory) getWorkloadSpec(framework api.Framework, workloadType api.WorkloadType) (*spec.GeneralSpec, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	switch workloadType {
	case api.Distributed:
		return spec.NewDistributedSpec(framework), nil
	case api.DistributedInference:
		return spec.NewDistributedSpec(framework), nil
	case api.Training:
		return spec.NewTrainingSpec(), nil
	case api.Workspace:
		return spec.NewWorkspaceSpec(), nil
	case api.Inference:
		if framework == api.NimService {
			return spec.NewNimSpec(), nil
		}
		return spec.NewInferenceSpec(), nil
	default:
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
}

func (f *Factory) getCommandTexts(framework api.Framework, workloadType api.WorkloadType) (cmd.CommandTexts, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
	if (framework == api.Runai && workloadType == api.Workspace) ||
		(framework == api.Knative && workloadType == api.Inference) ||
		(framework == api.LWS && workloadType == api.DistributedInference) ||
		(framework == api.NimService && workloadType == api.Inference) {
		if framework == api.NimService {
			return NewNimSubmitText(), nil
		}
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
