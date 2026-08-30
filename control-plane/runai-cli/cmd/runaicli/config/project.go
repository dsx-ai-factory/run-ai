package config

import (
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

// projects set default project
// Deprecated: Use `runai project set` instead
var projects = &cobra.Command{
	Use:        "project PROJECT_NAME",
	Short:      consts.DepConfigProjectMsg,
	Aliases:    []string{"projects"},
	Args:       cobra.ExactArgs(1),
	Deprecated: consts.DepConfigProject,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return project.Cmd.PersistentPreRunE(cmd, args)
	},
	PreRunE: project.Set.PreRunE,
	RunE:    project.Set.RunE,
}

func init() {
	Config.AddCommand(projects)
}
