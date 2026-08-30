package cluster

import (
	commCluster "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/cluster"
	commonconfig "github.com/run-ai/runai/control-plane/runai-cli/internal/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/list/manipulate"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"
	"github.com/spf13/cobra"
)

var (
	Output listCommon.Output
)

// List represents the List command
// set as a global variable to be used by alias commands for backward compatibility
// when the alias commands are removed, this variable should be return to be a local variable
var List = &cobra.Command{
	Use:   "list",
	Short: consts.ClusterListShortMsg,
	Args:  cobra.NoArgs,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := commonconfig.SetOutput(&Output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
		}

		err = Output.RawFlagValidator()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags, err)
		}

		err = HandleColumnsFlags(Output, cmd)
		if err != nil {
			return err
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		token := instance.GetToken()

		handler, err := cluster.GetClusterList(txn, cpUrl, token)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListClusters, err)
		}

		var printOut string
		printOut, err = Output.Generate(handler)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListClusters, err)
		}
		cmd.Println(printOut)
		return nil
	},
}

func init() {
	Cmd.AddCommand(List)

	listCommon.FlagOutputType(List, &Output)
	listCommon.FlagOutputTypeWithColumns(List, &Output)

	// Register autocomplete function for columns flag
	_ = List.RegisterFlagCompletionFunc(listCommon.ColumnsStr, commCluster.ColumnsFlagCompletion)
}

func HandleColumnsFlags(output listCommon.Output, cmd *cobra.Command) error {
	var err error
	userColumns := output.GetColumns()

	if len(userColumns) > 0 {
		_, err = manipulate.ParseClusterColumns(userColumns)
		if err != nil {
			return err
		}
	}
	if output.IsSetDefaultColumns() {
		err = settings.GetInstance().SetClusterColumns(userColumns)
		if err != nil {
			return err
		}
		if len(userColumns) == 0 {
			cmd.Printf(consts.DefaultClusterColumnsSet, manipulate.GetDefaultColumns())
		} else {
			cmd.Printf(consts.DefaultClusterColumnsSet, userColumns)
		}
	}
	return nil
}
