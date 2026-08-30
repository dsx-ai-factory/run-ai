package tensorflow

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	cmdfactory "github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

// Cmd represents the cluster command
var Cmd = &cobra.Command{
	Use:        "tensorflow",
	Short:      consts.TensorflowShortMsg,
	Args:       cobra.NoArgs,
	Aliases:    []string{"tf"},
	GroupID:    consts.GroupTrainingID,
	Deprecated: consts.DepFrameworkMsg,

	PersistentPreRunE: cmdfactory.WorkloadsPersistentPreRunE,
	PreRunE:           cmdfactory.WorkloadsPreRunE,
}

func Init(cmd *cobra.Command) {
	cmdfactory.WorkloadsCmdInit(cmd, api.TF, api.Distributed)
}

func init() {
	Init(Cmd)
}
