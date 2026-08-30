package shortcuts

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/training/pytorch"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

var Pytorch = &cobra.Command{
	Use:     pytorch.Cmd.Use,
	Short:   fmt.Sprintf("alias for %s", pytorch.Cmd.Short),
	Aliases: pytorch.Cmd.Aliases,
	Args:    pytorch.Cmd.Args,
	GroupID: consts.GroupWorkloadID,

	PersistentPreRunE: pytorch.Cmd.PersistentPreRunE,
	PreRunE:           pytorch.Cmd.PersistentPreRunE,
}

func init() {
	pytorch.Init(Pytorch)
}
