package shortcuts

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/training/mpi"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

var Mpi = &cobra.Command{
	Use:     mpi.Cmd.Use,
	Short:   fmt.Sprintf("alias for %s", mpi.Cmd.Short),
	Aliases: mpi.Cmd.Aliases,
	Args:    mpi.Cmd.Args,
	GroupID: consts.GroupWorkloadID,

	PersistentPreRunE: mpi.Cmd.PersistentPreRunE,
	PreRunE:           mpi.Cmd.PersistentPreRunE,
}

func init() {
	mpi.Init(Mpi)
}
