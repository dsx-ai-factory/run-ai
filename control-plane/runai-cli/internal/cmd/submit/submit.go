package submit

import (
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/submit"
	workloadcommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/priorities"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/pvc"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type Command struct {
	cmd.BaseCommand
	generalSpec  *spec.GeneralSpec
	commandTexts cmd.CommandTexts
}

func NewCommand(specType api.Framework, workloadSpec *spec.GeneralSpec, initBaseCommandFn func() *cobra.Command,
	commandTexts cmd.CommandTexts,
	instance *config.Instance, cmdHelper helper.Cmd) *Command {
	if initBaseCommandFn == nil {
		initBaseCommandFn = cmd.NewEmptyCobraCommand
	}
	return &Command{
		BaseCommand: cmd.BaseCommand{
			Cmd:       initBaseCommandFn(),
			Framework: specType,
			GetTxnFn:  instance.GetTxn,
			Instance:  instance,
			CmdHelper: cmdHelper,
		},
		generalSpec:  workloadSpec,
		commandTexts: commandTexts,
	}
}

func (c *Command) GetCommand() *cobra.Command {
	return c.buildCommand()
}

func (c *Command) buildCommand() *cobra.Command {
	c.Cmd.Use = "submit"
	c.Cmd.Short = c.commandTexts.GetShort()
	c.Cmd.Long = c.commandTexts.GetLong()
	c.Cmd.Example = c.commandTexts.GetExample()
	c.Cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		submit.Spec = c.generalSpec
		return submit.Pre(cmd, args)
	}
	c.Cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if c.Framework != api.Runai && c.Framework != api.Knative && c.Framework != api.LWS && c.Framework != api.NimService {
			// ToDo should be removed after final approval
			err := c.generalSpec.CheckDistributeFlags()
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetBaseErrorMsg(), err)
			}
		}
		if c.Framework == api.MPI {
			err := c.generalSpec.CheckDistributeMpiFlags()
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetBaseErrorMsg(), err)
			}
		}
		return submit.Run(cmd)
	}
	return c.Cmd
}

func (c *Command) RegisterFlags() error {
	c.generalSpec.SetFlags(c.Cmd, &submit.Project)
	_ = c.Cmd.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)
	_ = c.Cmd.RegisterFlagCompletionFunc(consts.AttributeNameExistingPvc, pvc.GetExistingPVCFlag)
	_ = c.Cmd.RegisterFlagCompletionFunc(consts.AttributeNamePriorityClass, priorities.GetPrioritiesNamesForFlag)

	if c.generalSpec.GetType() != spectype.NimService {
		c.Cmd.MarkFlagsMutuallyExclusive(consts.AttributeNameRequiredPodAffinityKey, consts.AttributeNamePreferredPodAffinityKey)
	}

	if c.generalSpec.GetType() == spectype.Workspace {
		c.Cmd.MarkFlagsMutuallyExclusive(consts.AttributeNamePriorityClass, consts.AttributeNamePreemptible)
		c.Cmd.MarkFlagsMutuallyExclusive(consts.AttributeNamePreemptible, consts.AttributeNamePreemptibility)
	}

	_ = c.Cmd.RegisterFlagCompletionFunc(consts.AttributeNameCleanPodPolicy, workloadcommon.CleanPodPolicyFlagCompletion)
	_ = c.Cmd.RegisterFlagCompletionFunc(consts.AttributeNameRestartPolicy, workloadcommon.RestartPolicyFlagCompletion)
	_ = c.Cmd.RegisterFlagCompletionFunc(consts.AttributeNameLauncherCreationPolicy, workloadcommon.LauncherCreationPolicyFlagCompletion)
	_ = c.Cmd.RegisterFlagCompletionFunc(consts.AttributeNamePreemptibility, workloadcommon.PreemptibilityFlagCompletion)
	if c.generalSpec.GetType() == spectype.TrainingDistributed {
		_ = c.Cmd.RegisterFlagCompletionFunc(consts.AttributeNameMasterRestartPolicy, workloadcommon.RestartPolicyFlagCompletion)
	}
	return nil
}
