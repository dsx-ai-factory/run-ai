package logs

import (
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/backward"
	commonlogs "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/logs"
	commWorkload "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	workloadType string
	LogOptions   commonlogs.Options
	instance     = config.GetInstance()
	txn          *transaction.Txn
)

// clusters represents the clusters command
// Deprecated: Use `runai workload logs` instead
var Cmd = &cobra.Command{
	Use:        "logs WORKLOAD_NAME",
	Short:      consts.LogsShortMsgDep,
	GroupID:    consts.GroupDeprecatedID,
	Args:       cobra.ExactArgs(1),
	Deprecated: consts.DepLogsMsg,
	// Do not add deprecated flag, it removes the command from help

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		txn = instance.GetTxn()
		return workload.Cmd.PersistentPreRunE(cmd, args)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := commonlogs.Pre(cmd, args, &LogOptions)
		if err != nil {
			return err
		}
		clusterWorkload, err := backward.HandleUniqWorkloadBackwardCompatiblePreE(txn.GetCtx(),
			cmd, args[0], workloadType, LogOptions.Project, consts.FailedLogsCommand)
		if err != nil {
			return err
		}
		LogOptions.WorkloadType = clusterWorkload.Type
		LogOptions.WorkloadFramework = clusterWorkload.Framework
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return commonlogs.Run(cmd, &LogOptions, txn)
	},
}

func init() {
	Cmd.Flags().StringVar(&workloadType, consts.WorkloadType, "", consts.WorkloadTypeUsage)
	commonlogs.FlagOutputType(Cmd, &LogOptions)
	_ = Cmd.RegisterFlagCompletionFunc(consts.WorkloadType, commWorkload.TypeFlagCompletion)
}
