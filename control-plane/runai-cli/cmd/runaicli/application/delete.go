package application

import (
	"fmt"

	applicationservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/application"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"

	projectCmd "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/template"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	Project string
	Force   bool
	Yes     bool
}

var deleteFlags deleteOptions

var Delete = &cobra.Command{
	Use:               "delete APPLICATION_NAME",
	Short:             DeleteShortMsg,
	Args:              cobra.ExactArgs(1),
	Example:           DeleteExample,
	ValidArgsFunction: GetAIApplicationNamesForArg,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if deleteFlags.Project == "" {
			deleteFlags.Project = instance.GetConfigValue(config.ClusterProjectName)
		}
		cmdHelper.AddValidation(validators.NewProjectValidator(deleteFlags.Project))

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		confirmed, err := confirmApplicationDeletion(appName, deleteFlags.Project)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDeleteAIApplication, err)
		}
		if !confirmed {
			cmd.Println(DeleteAbortedMsg)
			return nil
		}

		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		clusterId := instance.GetConfigValue(config.ClusterUUID)
		token := instance.GetToken()

		if _, err = projectservice.GetProjectId(txn, cpUrl, token, clusterId, deleteFlags.Project); err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDeleteAIApplication, err)
		}

		app, err := applicationservice.GetAIApplicationByName(txn, cpUrl, token, clusterId, deleteFlags.Project, appName)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDeleteAIApplication, err)
		}
		if app == nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDeleteAIApplication,
				fmt.Errorf(ErrAppNotFound, appName, deleteFlags.Project))
		}

		// The delete API is idempotent (a repeated delete returns 202), so the
		// "already being deleted" state has to be detected client side.
		if app.DesiredPhase.Get() != nil && *app.DesiredPhase.Get() == desiredPhaseDeleted ||
			app.GetActionStatus() == actionStatusDeleting {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDeleteAIApplication,
				fmt.Errorf(ErrAppBeingDeleted, appName))
		}

		if err = applicationservice.DeleteAIApplication(txn, cpUrl, token, app.GetId()); err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDeleteAIApplication, err)
		}

		cmd.Println(fmt.Sprintf(DeleteSuccessMsg, appName))
		return nil
	},
}

func confirmApplicationDeletion(appName, projectName string) (bool, error) {
	if deleteFlags.Force || deleteFlags.Yes {
		return true, nil
	}
	return template.YesNoQuestion(fmt.Sprintf(DeleteConfirmMsg, appName, projectName))
}

func init() {
	Cmd.AddCommand(Delete)

	Delete.Flags().StringVarP(&deleteFlags.Project, consts.ProjectStr, consts.ProjectShort, "", ProjectFlagUsage)
	Delete.Flags().BoolVar(&deleteFlags.Force, consts.AttributeNameForce, false, ForceFlagUsage)
	Delete.Flags().BoolVarP(&deleteFlags.Yes, attributeNameYes, "y", false, YesFlagUsage)
	_ = Delete.RegisterFlagCompletionFunc(consts.ProjectStr, projectCmd.GetProjectNamesForFlag)
}
