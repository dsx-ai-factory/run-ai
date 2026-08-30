package workspace

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	cmdfactory "github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

// Cmd represents the cluster command
var Cmd = &cobra.Command{
	Use:     "workspace",
	Short:   consts.WorkspaceShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupWorkloadID,

	PersistentPreRunE: cmdfactory.WorkloadsPersistentPreRunE,
	PreRunE:           cmdfactory.WorkloadsPreRunE,
}

func Init(cmd *cobra.Command) {
	cmdfactory.WorkloadsCmdInit(cmd, api.Runai, api.Workspace)
}

func init() {
	Init(Cmd)
}
