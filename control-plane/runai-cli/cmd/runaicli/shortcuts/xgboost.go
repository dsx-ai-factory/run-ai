package shortcuts

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/training/xgboost"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

var Xgboost = &cobra.Command{
	Use:        xgboost.Cmd.Use,
	Short:      fmt.Sprintf("alias for %s", xgboost.Cmd.Short),
	Aliases:    xgboost.Cmd.Aliases,
	Args:       xgboost.Cmd.Args,
	GroupID:    consts.GroupWorkloadID,
	Deprecated: xgboost.Cmd.Deprecated,

	PersistentPreRunE: xgboost.Cmd.PersistentPreRunE,
	PreRunE:           xgboost.Cmd.PersistentPreRunE,
}

func init() {
	xgboost.Init(Xgboost)
}
