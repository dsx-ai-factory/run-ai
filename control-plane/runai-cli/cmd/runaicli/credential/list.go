package credential

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"
	credCompletion "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/credential"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	commonconfig "github.com/run-ai/runai/control-plane/runai-cli/internal/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	credlist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/credential/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	All     bool
	Project string
	Output  list.Output
	Type    string
}

var (
	listOptions ListOptions
)

var List = &cobra.Command{
	Use:     "list",
	Short:   ListShortMsg,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Example: ListExample,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := commonconfig.SetOutput(&listOptions.Output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
		}

		err = listOptions.Output.RawFlagValidator()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags, err)
		}

		err = credlist.HandleColumnsFlags(listOptions.Output, cmd, DefaultCredentialColumnsSet)
		if err != nil {
			return err
		}

		if listOptions.Type != "" {
			credType := strings.TrimSpace(listOptions.Type)
			if credType == "" || !allowedCredentialTypes.Contains(credType) {
				return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags,
					fmt.Errorf("invalid credential type %q; must be one of: accessKey, dockerRegistry, genericSecret, ngcApiKey, password", listOptions.Type))
			}
			listOptions.Type = credType
		}

		if !listOptions.All {
			if listOptions.Project == "" {
				listOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
			}
			cmdHelper.AddValidation(validators.NewProjectValidator(listOptions.Project))
		}

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		token := instance.GetToken()
		clusterId := instance.GetConfigValue(config.ClusterUUID)

		projectName := ""
		if !listOptions.All {
			projectName = listOptions.Project
		}

		handler, err := credlist.List(instance.GetTxn(), cpUrl, token, projectName, clusterId, listOptions.Type)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListCredentials, err)
		}

		if err := renderList(cmd, listOptions.Output, handler); err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListCredentials, err)
		}
		return nil
	},
}

func renderList(cmd *cobra.Command, output list.Output, handler *credlist.Output) error {
	printOut, err := output.Generate(handler)
	if err != nil {
		return err
	}

	cmd.Println(printOut)
	return nil
}

func init() {
	Cmd.AddCommand(List)

	list.FlagOutputType(List, &listOptions.Output)
	list.FlagOutputTypeWithColumns(List, &listOptions.Output)

	List.Flags().BoolVarP(&listOptions.All, consts.ProjectAll, "A", false,
		ListAllFlagUsage)
	List.Flags().StringVarP(&listOptions.Project, consts.ProjectStr, consts.ProjectShort, "",
		ProjectFlagUsage)
	List.Flags().StringVar(&listOptions.Type, TypeFlag, "", TypeFlagUsage)

	_ = List.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)
	_ = List.RegisterFlagCompletionFunc(list.ColumnsStr, credCompletion.ColumnsFlagCompletion)
	_ = List.RegisterFlagCompletionFunc(TypeFlag, credCompletion.TypeFlagCompletion)

	List.MarkFlagsMutuallyExclusive(consts.ProjectAll, consts.ProjectStr)
}
