package training

import (
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/training/jax"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/training/mpi"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/training/pytorch"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/training/runaijob"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/training/tensorflow"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/training/xgboost"
	cmdfactory "github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

// shared with package
var (
	cmdHelper helper.Cmd
	instance  = config.GetInstance()
)

// Cmd represents the cluster command
var Cmd = &cobra.Command{
	Use:     "training",
	Short:   consts.TrainingShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupWorkloadID,

	PersistentPreRunE: cmdfactory.WorkloadsPersistentPreRunE,
	PreRunE:           cmdfactory.WorkloadsPreRunE,
}

func init() {
	Cmd.AddGroup(
		&cobra.Group{
			ID:    consts.GroupTrainingID,
			Title: consts.GroupTrainingTitle,
		},
		&cobra.Group{
			ID:    consts.GroupTrainingBaseID,
			Title: consts.GroupTrainingBaseTitle,
		},
		&cobra.Group{
			ID:    consts.GroupTrainingStdID,
			Title: consts.GroupTrainingStdTitle,
		},
	)

	runaijob.Init(Cmd)
	for _, c := range Cmd.Commands() {
		c.GroupID = consts.GroupTrainingStdID
	}

	Cmd.AddCommand(mpi.Cmd)
	Cmd.AddCommand(runaijob.Cmd)
	Cmd.AddCommand(xgboost.Cmd)
	Cmd.AddCommand(tensorflow.Cmd)
	Cmd.AddCommand(pytorch.Cmd)
	Cmd.AddCommand(jax.Cmd)

	// remove standard list command from training
	for _, c := range Cmd.Commands() {
		if c.Use == "list" {
			Cmd.RemoveCommand(c)
		}
	}

	Cmd.AddCommand(list)
}
