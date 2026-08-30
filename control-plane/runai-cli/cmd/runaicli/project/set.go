package project

import (
	"strconv"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	project string
)

var Set = &cobra.Command{
	Use:   "set PROJECT_NAME",
	Short: consts.ProjectAttributeSetUsage,
	Args:  cobra.ExactArgs(1),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		project = args[0]
		validator := validators.NewProjectValidator(project)
		msg, err := validator.Execute()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, msg, err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		clusterId := instance.GetConfigValue(config.ClusterUUID)
		token := instance.GetToken()

		projectIdStr, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, project)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToDefaultProject, err)
		}

		projectId, err := strconv.ParseInt(projectIdStr, 10, 32)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToDefaultProject, err)
		}

		logger := instance.GetLogger()
		logger.Debug("setting project default value", zap.String("project", project))

		err = instance.SetProject(project, int32(projectId))
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToDefaultProject, err)
		}
		cmd.Printf(consts.SetProjectProperties, project)
		return nil
	},

	ValidArgsFunction: GetProjectNamesForArg,
}

func init() {
	Cmd.AddCommand(Set)
}
