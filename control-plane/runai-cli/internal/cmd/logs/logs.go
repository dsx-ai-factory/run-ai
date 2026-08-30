package logs

import (
	commonlogs "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/logs"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type Command struct {
	cmd.BaseCommand
	options      commonlogs.Options
	texts        cmd.CommandTexts
	workloadType api.WorkloadType
	completionFn completion.CompletionFn
}

func NewLogsCommand(framework api.Framework, initBaseCommandFn func() *cobra.Command, workloadType api.WorkloadType,
	text cmd.CommandTexts, instance *config.Instance, cmdHelper helper.Cmd, completionFn completion.CompletionFn) *Command {
	if initBaseCommandFn == nil {
		initBaseCommandFn = cmd.NewEmptyCobraCommand
	}
	return &Command{
		BaseCommand: cmd.BaseCommand{
			Cmd:       initBaseCommandFn(),
			Framework: framework,
			GetTxnFn:  instance.GetTxn,
			Instance:  instance,
			CmdHelper: cmdHelper,
		},
		options:      commonlogs.Options{},
		workloadType: workloadType,
		texts:        text,
		completionFn: completionFn,
	}
}
func (l *Command) GetCommand() *cobra.Command {
	return l.buildCommand()
}

func (l *Command) buildCommand() *cobra.Command {
	l.Cmd.Use = "logs [WORKLOAD_NAME]"
	l.Cmd.Short = l.texts.GetShort()
	l.Cmd.Long = l.texts.GetLong()
	l.Cmd.Example = l.texts.GetExample()
	l.Cmd.Args = cobra.ExactArgs(1)
	l.Cmd.PreRunE = l.preRunE
	l.Cmd.RunE = l.runE
	l.Cmd.ValidArgsFunction = l.completionFn
	return l.Cmd
}

func (l *Command) preRunE(cmd *cobra.Command, args []string) error {
	l.options.WorkloadType = l.workloadType
	l.options.WorkloadFramework = l.Framework
	return commonlogs.Pre(cmd, args, &l.options)
}

func (l *Command) runE(cmd *cobra.Command, args []string) error {
	return commonlogs.Run(cmd, &l.options, l.GetTxnFn())
}

func (l *Command) RegisterFlags() error {
	commonlogs.FlagOutputType(l.Cmd, &l.options)
	return nil
}
