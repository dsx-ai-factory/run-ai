package standard

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	cmdfactory "github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

const (
	ShortMsg = "Runs a single inference process on one node. Suitable for smaller models or simpler inference tasks."
)

// Cmd represents the standard inference command
var Cmd = &cobra.Command{
	Use:     "standard",
	Short:   ShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupInferenceID,

	PersistentPreRunE: cmdfactory.WorkloadsPersistentPreRunE,
	PreRunE:           cmdfactory.WorkloadsPreRunE,
}

func Init(cmd *cobra.Command) {
	cmdfactory.WorkloadsCmdInit(cmd, api.Knative, api.Inference)
}

func init() {
	Init(Cmd)
}
