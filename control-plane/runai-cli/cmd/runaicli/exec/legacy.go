package exec

import (
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/backward"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/exec"
	commWorkload "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	workloadType string
	execOpt      exec.Options
	instance     = config.GetInstance()
	txn          *transaction.Txn
)

// Deprecated: Use `runai workload exec` instead
var Cmd = &cobra.Command{
	Use:        "exec WORKLOAD_NAME",
	Short:      consts.ExecShortMsgDep,
	GroupID:    consts.GroupDeprecatedID,
	Args:       cobra.MinimumNArgs(2),
	Deprecated: consts.DepExecMsg,
	// Do not add deprecated flag, it removes the command from help

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		txn = instance.GetTxn()
		return workload.Cmd.PersistentPreRunE(cmd, args)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := exec.Pre(cmd, args, &execOpt)
		if err != nil {
			return err
		}
		clusterWorkload, err := backward.HandleUniqWorkloadBackwardCompatiblePreE(txn.GetCtx(),
			cmd, args[0], workloadType, execOpt.Project, consts.ExecCommandError)
		if err != nil {
			return err
		}
		execOpt.WorkloadType = clusterWorkload.Type
		execOpt.WorkloadFramework = clusterWorkload.Framework
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return exec.Run(cmd, &execOpt, txn)
	},
}

func init() {
	Cmd.Flags().StringVar(&workloadType, consts.WorkloadType, "", consts.WorkloadTypeUsage)
	exec.FlagOutputType(Cmd, &execOpt)
	_ = Cmd.RegisterFlagCompletionFunc(consts.WorkloadType, commWorkload.TypeFlagCompletion)
}
