package distributed

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	cmdfactory "github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

// Cmd represents the distributed inference command
var Cmd = &cobra.Command{
	Use:     "distributed",
	Short:   consts.DistributedInferenceShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupInferenceID,

	PersistentPreRunE: cmdfactory.WorkloadsPersistentPreRunE,
	PreRunE:           cmdfactory.WorkloadsPreRunE,
}

func Init(cmd *cobra.Command) {
	cmdfactory.WorkloadsCmdInit(cmd, api.LWS, api.DistributedInference)
}

func init() {
	Init(Cmd)
}
