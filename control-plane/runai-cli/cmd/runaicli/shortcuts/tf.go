package shortcuts

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/training/tensorflow"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

var Tensorflow = &cobra.Command{
	Use:        tensorflow.Cmd.Use,
	Short:      fmt.Sprintf("alias for %s", tensorflow.Cmd.Short),
	Aliases:    tensorflow.Cmd.Aliases,
	Args:       tensorflow.Cmd.Args,
	GroupID:    consts.GroupWorkloadID,
	Deprecated: tensorflow.Cmd.Deprecated,

	PersistentPreRunE: tensorflow.Cmd.PersistentPreRunE,
	PreRunE:           tensorflow.Cmd.PersistentPreRunE,
}

func init() {
	tensorflow.Init(Tensorflow)
}
