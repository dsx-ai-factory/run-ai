package factory

import (
	internalApi "github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/attach"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/completion"
	deletecmd "github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/delete"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/describe"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/exec"
	listcmd "github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/pf"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/scale"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/submitdistbc"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/suspend"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/logs"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/submit"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/update"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
)

type Commands struct {
	Instance  *config.Instance
	CmdHelper helper.Cmd
	Factories []cmd.Factory
}

func NewCommands(instance *config.Instance, cmdHelper helper.Cmd) Commands {
	listFactory := listcmd.NewCommandFactory(instance, cmdHelper)
	completionFactory := completion.NewCompletionFactory(listFactory, InstanceInit)

	factoriesFn := []cmd.Factory{
		listFactory,
		deletecmd.NewCommandFactory(instance, cmdHelper, completionFactory),
		submit.NewCommandFactory(instance, cmdHelper),
		update.NewCommandFactory(instance, cmdHelper),
		scale.NewCommandFactory(instance, cmdHelper, completionFactory),
		submitdistbc.NewCommandFactory(instance, cmdHelper),
		logs.NewCommandFactory(instance, cmdHelper, completionFactory),
		pf.NewCommandFactory(instance, cmdHelper, completionFactory),
		describe.NewCommandFactory(instance, cmdHelper, completionFactory),
		suspend.NewResumeFactory(instance, cmdHelper, completionFactory),
		suspend.NewSuspendFactory(instance, cmdHelper, completionFactory),
		attach.NewCommandFactory(instance, cmdHelper, completionFactory),
		exec.NewCommandFactory(instance, cmdHelper, completionFactory),
	}

	return Commands{
		Instance:  instance,
		CmdHelper: cmdHelper,
		Factories: factoriesFn,
	}
}

func (c Commands) GetAllSupportedCommands(framework internalApi.Framework, workloadType internalApi.WorkloadType) ([]cmd.ActionCommand, error) {
	var commands []cmd.ActionCommand
	var err error
	for _, commandFactory := range c.Factories {
		commands, err = appendIfSupported(commands, commandFactory, framework, workloadType)
		if err != nil {
			return nil, err
		}
	}

	return commands, nil
}

func appendIfSupported(commands []cmd.ActionCommand, commandFactory cmd.Factory, framework internalApi.Framework, workloadType internalApi.WorkloadType) ([]cmd.ActionCommand, error) {
	if !commandFactory.IsSupported(framework, workloadType) {
		return commands, nil
	}
	command, err := commandFactory.GetCommand(framework, workloadType)
	if err != nil {
		// shouldn't happen
		return nil, err
	}
	commands = append(commands, command)

	if !commandFactory.IsAliasSupported(framework, workloadType) {
		return commands, nil
	}
	aliasCommands, err := commandFactory.GetAliasCommand(framework, workloadType)
	if err != nil {
		// shouldn't happen
		return nil, err
	}
	return append(commands, aliasCommands...), nil
}
