package suspend

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributed"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/training"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workspace"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
)

const (
	Suspend = "suspend"
	Resume  = "resume"
)

// Factory creates and manages suspend/resume commands
type Factory struct {
	action            string
	getTextFn         func(api.Framework, api.WorkloadType) (cmd.SearchableCommandTexts, error)
	getStateFn        func(api.Framework, api.WorkloadType) (cmd.RunOperationFn, error)
	instance          *config.Instance
	cmdHelper         helper.Cmd
	completionFactory completion.Factory
}

// IsSupported checks if the command is supported for given framework and workload type
func (f *Factory) IsSupported(framework api.Framework, workloadType api.WorkloadType) bool {
	return isSupported(framework, workloadType)
}

func isSupported(framework api.Framework, workloadType api.WorkloadType) bool {
	switch framework {
	case api.MPI, api.XGBoost, api.TF, api.PyTorch, api.Jax:
		return true
	case api.Runai:
		return workloadType == api.Training || workloadType == api.Workspace
	// Note: Distributed Inference does not support suspend/resume
	default:
		return false
	}
}

func (f *Factory) IsAliasSupported(_ api.Framework, _ api.WorkloadType) bool {
	return false
}

func (f *Factory) GetAliasCommand(framework api.Framework, workload api.WorkloadType) ([]cmd.ActionCommand, error) {
	return nil, cmd.NewUnsupportedAliasError(framework, workload)
}

func (f *Factory) GetCommand(framework api.Framework, workloadType api.WorkloadType) (cmd.ActionCommand, error) {
	if !f.IsSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	commandText, err := f.getTextFn(framework, workloadType)
	if err != nil {
		return nil, err
	}

	operationFn, err := f.getStateFn(framework, workloadType)
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

	bc := cmd.NewSearchBaseCommand(framework, cmd.NewEmptyCobraCommand, f.action, operationFn,
		searchFn, workloadType, f.instance, commandText, f.cmdHelper)
	bc.Cmd.ValidArgsFunction = completionFn

	return bc, nil
}

func getSuspendCommandTexts(framework api.Framework, workloadType api.WorkloadType) (cmd.SearchableCommandTexts, error) {
	if !isSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	if workloadType == api.Workspace {
		return NewWorkloadOnlySuspendTexts(workloadType), nil
	}
	return NewFrameworkSuspendTexts(workloadType, framework), nil
}

func getResumeCommandText(framework api.Framework, workloadType api.WorkloadType) (cmd.SearchableCommandTexts, error) {
	if !isSupported(framework, workloadType) {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	if workloadType == api.Workspace {
		return NewWorkloadOnlyResumeTexts(workloadType), nil
	}
	return NewFrameworkResumeTexts(workloadType, framework), nil
}

// NewSuspendFactory creates a new Factory for suspend commands
func NewSuspendFactory(instance *config.Instance, cmdHelper helper.Cmd, completionFactory completion.Factory) *Factory {
	return &Factory{
		action:            Suspend,
		getTextFn:         getSuspendCommandTexts,
		getStateFn:        getSuspendFn,
		instance:          instance,
		cmdHelper:         cmdHelper,
		completionFactory: completionFactory,
	}
}

// NewResumeFactory creates a new Factory for resume commands
func NewResumeFactory(instance *config.Instance, cmdHelper helper.Cmd, completionFactory completion.Factory) *Factory {
	return &Factory{
		action:            Resume,
		getTextFn:         getResumeCommandText,
		getStateFn:        getResumeFn,
		instance:          instance,
		cmdHelper:         cmdHelper,
		completionFactory: completionFactory,
	}
}

func getSuspendFn(framework api.Framework, workloadType api.WorkloadType) (cmd.RunOperationFn, error) {
	if framework <= api.UndefinedFramework || framework >= api.AnyDistributed {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	switch workloadType {
	case api.Distributed:
		return distributed.Suspend, nil
	case api.Training:
		return training.Suspend, nil
	case api.Workspace:
		return workspace.Suspend, nil
	default:
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
}

func getResumeFn(framework api.Framework, workloadType api.WorkloadType) (cmd.RunOperationFn, error) {
	if framework <= api.UndefinedFramework || framework >= api.AnyDistributed {
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}

	switch workloadType {
	case api.Distributed:
		return distributed.Resume, nil
	case api.Training:
		return training.Resume, nil
	case api.Workspace:
		return workspace.Resume, nil
	default:
		return nil, cmd.NewUnsupportedError(framework, workloadType)
	}
}
