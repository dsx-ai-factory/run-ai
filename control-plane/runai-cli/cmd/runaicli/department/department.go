package department

import (
	"fmt"
	"strconv"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/department/model"

	commAttr "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/attributes"

	departmentservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/department"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"

	token "github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config/attributes"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	"github.com/spf13/cobra"
)

var (
	cmdHelper   helper.Cmd
	interactive attributes.EnableFlag
	instance    = config.GetInstance()
	txn         *transaction.Txn
)

// Cmd represents the department command
var Cmd = &cobra.Command{
	Use:     "department",
	Aliases: []string{"departments", "dep", "deps"},
	Short:   consts.DepartmentShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupOrganizationID,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := initDepartmentInstance(cmd)
		if err != nil {
			// err is already formatted
			return err
		}

		if instance.GetOptions().Quiet {
			if err := commAttr.EnableQuietMode(cmd); err != nil {
				return err
			}
		}

		return nil
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		options := instance.GetOptions()
		if !cmdHelper.CheckInteractive(txn, interactive, options.Interactive) {
			cmd.HelpFunc()(cmd, args)
			return cmdtxn.CmdErrorSetMissingParams(cmd)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var filters []model.Filter
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		clusterId := instance.GetConfigValue(config.ClusterUUID)
		token := instance.GetToken()

		filters = append(filters, *model.NewClusterIDFilter(clusterId))
		handler, err := departmentservice.GetDepartmentList(txn, cpUrl, token, filters, options.limit, options.offset)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListDepartments, err)
		}

		selectedDepartment, err := table.Selection(handler)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListDepartments, err)
		}
		if selectedDepartment == nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListDepartments, fmt.Errorf(consts.ErrorSelectionEmpty))
		}

		selectedDepartmentName, exist := (*selectedDepartment)["Name"]
		if !exist {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListDepartments, fmt.Errorf(consts.ListDepartmentsNotFound))
		}

		foundDepartment := false
		for _, department := range handler.GetList() {
			if department.Name == selectedDepartmentName {
				departmentId, err := strconv.Atoi(department.Id)
				if err != nil {
					return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListDepartments, err)
				}
				foundDepartment = true
				err = instance.SetDepartment(department.Name, int32(departmentId))
				if err != nil {
					return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListDepartments, err)
				}
				break
			}
		}

		if !foundDepartment {
			err = fmt.Errorf(consts.ErrorNoDepartmentFound, selectedDepartmentName)
		}

		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorNoDepartmentFound, err)
		}
		cmd.Printf(consts.SetDepartmentProperties, selectedDepartmentName)
		return nil
	},
}

func init() {
	Cmd.Flags().Var(&interactive, consts.Interactive, consts.InteractiveArgumentUsage)
	_ = Cmd.RegisterFlagCompletionFunc(consts.Interactive, attributes.EnableFlagCompletion)
	SetFlags(Cmd)
}

func initDepartmentInstance(cmd *cobra.Command) error {
	instance = config.GetInstance()

	txn = instance.GetTxn()
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterUrl := instance.GetConfigValue(config.ClusterDomain)
	cmdHelper.AddValidation(validators.NewCpValidator(cpUrl))
	cmdHelper.AddValidation(validators.NewClusterValidator(clusterUrl))

	if msg, err := cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}
	tokenStr, err := token.GetToken(txn.GetCtx())
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ReLoginMsg, err)
	}
	instance.SetToken(tokenStr)

	return nil
}
