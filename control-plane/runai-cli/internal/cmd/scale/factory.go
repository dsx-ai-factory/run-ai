package scale

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributedinference"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/nim"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
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
	return (framework == api.LWS && workloadType == api.DistributedInference) ||
		(framework == api.NimService && workloadType == api.Inference)
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

	completionFn, err := f.completionFactory.GetCompletionFn(framework, workloadType)
	if err != nil {
		return nil, err
	}

	if framework == api.NimService {
		getAutoscalingInfoFn := f.getAutoscalingInfoFn(framework, workloadType)
		return NewCommandWithAutoscalingCheck(framework, workloadType, workloadSpec, nil, commandText, f.instance, f.cmdHelper, searchWorkloadFn, updateWorkloadFn, completionFn, getAutoscalingInfoFn), nil
	}

	return NewCommand(framework, workloadType, workloadSpec, nil, commandText, f.instance, f.cmdHelper, searchWorkloadFn, updateWorkloadFn, completionFn), nil
}

func (f *Factory) getAutoscalingInfoFn(framework api.Framework, workloadType api.WorkloadType) GetAutoscalingInfoFn {
	if framework == api.NimService {
		return func(txn *transaction.Txn, domain, token, uuid string) (*AutoscalingInfo, error) {
			info, err := nim.GetAutoscalingInfo(txn, domain, token, uuid)
			if err != nil {
				return nil, err
			}
			if info == nil {
				return nil, nil
			}
			return &AutoscalingInfo{
				Enabled:     info.Enabled,
				MinReplicas: info.MinReplicas,
				MaxReplicas: info.MaxReplicas,
			}, nil
		}
	}
	return nil
}

func (f *Factory) getWorkloadSpec(framework api.Framework, workloadType api.WorkloadType) (*spec.GeneralSpec, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
	if framework == api.NimService {
		return spec.NewNimSpec(), nil
	}
	return spec.NewDistributedSpec(api.LWS), nil
}

func (f *Factory) getSearchWorkloadFn(framework api.Framework, workloadType api.WorkloadType) (cmd.SearchWorkloadFn, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
	if framework == api.NimService {
		return nim.GetByNameAndProj, nil
	}
	return distributedinference.GetByNameAndProj, nil
}

func (f *Factory) getUpdateWorkloadFn(framework api.Framework, workloadType api.WorkloadType) (cmd.UpdateWorkloadFn, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
	if framework == api.NimService {
		return nim.Update, nil
	}
	return distributedinference.Update, nil
}

func (f *Factory) getCommandTexts(framework api.Framework, workloadType api.WorkloadType) (cmd.SearchableCommandTexts, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
	return NewScaleText(workloadType, framework), nil
}

func (f *Factory) IsAliasSupported(_ api.Framework, _ api.WorkloadType) bool {
	return false
}

func (f *Factory) GetAliasCommand(framework api.Framework, workload api.WorkloadType) ([]cmd.ActionCommand, error) {
	return nil, cmd.NewUnsupportedAliasError(framework, workload)
}
