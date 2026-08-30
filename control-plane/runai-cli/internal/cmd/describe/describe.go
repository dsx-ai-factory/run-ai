package describe

import (
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workload"
	internalApi "github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type DescribeFn func(cmd *cobra.Command, args []string, options workload.DescribeCmdOptions, txn *transaction.Txn) error

type Command struct {
	cmd.BaseCommand
	options      workload.DescribeCmdOptions
	texts        cmd.CommandTexts
	workloadType internalApi.WorkloadType
	framework    internalApi.Framework
	completionFn completion.CompletionFn
	describeFn   DescribeFn
}

func NewCommand(framework internalApi.Framework, workloadType internalApi.WorkloadType, initBaseCommandFn func() *cobra.Command,
	texts cmd.CommandTexts, instance *config.Instance, cmdHelper helper.Cmd, completionFn completion.CompletionFn, describeFn DescribeFn) *Command {
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
		texts:        texts,
		workloadType: workloadType,
		framework:    framework,
		completionFn: completionFn,
		describeFn:   describeFn,
	}
}

func (c *Command) GetCommand() *cobra.Command {
	return c.buildCommand()
}

func (c *Command) buildCommand() *cobra.Command {
	c.Cmd.Use = "describe [WORKLOAD_NAME]"
	c.Cmd.Short = c.texts.GetShort()
	c.Cmd.Args = cobra.ExactArgs(1)
	c.Cmd.Example = c.texts.GetExample()
	c.options.Framework = c.framework
	c.options.WorkloadType = c.workloadType
	c.Cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		return workload.DescribeProjectValidatorPre(cmd, &c.options)
	}
	c.Cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return c.describeFn(cmd, args, c.options, c.GetTxnFn())
	}
	c.Cmd.ValidArgsFunction = c.completionFn
	return c.Cmd
}

func (c *Command) RegisterFlags() error {
	workload.BindDescribeFlags(c.Cmd, &c.options)
	return nil
}
