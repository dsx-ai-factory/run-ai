package submit

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/status"
	"go.uber.org/zap"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/attach"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/cluster"
	apitype "github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/inference"

	argshandler "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/args"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributed"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributedinference"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/nim"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/storage/datavolume"
	computeservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/compute"
	datasourceservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/datasource"
	datavolumeservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/datavolume"
	environmentservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/environment"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	templateservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/template"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/training"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workspace"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	EnvFormat            = "%s=%s"
	failedClusterStatus  = "failed to get cluster status. please check cluster availability"
	errEmptyTemplateFlag = "flag --template requires a non-empty value. Use --template <template_name>"
)

var (
	instance  = config.GetInstance()
	cpUrl     string
	cmdHelper helper.Cmd
	projectId string
	clusterId string
	token     string
	txn       *transaction.Txn

	Project string
	Name    string

	Spec *spec.GeneralSpec
)

func Pre(cmd *cobra.Command, args []string) error {
	var err error
	if Project == "" {
		Project = instance.GetConfigValue(config.ClusterProjectName)
	}

	Name, err = argshandler.ValidateArgsAndExtractName(cmd, &args)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
	}
	err = argshandler.BothAreSet(len(Name) != 0, Spec.HasPrefixName())
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
	}
	if execute, found := argshandler.GetExecuteString(&args); found {
		Spec.SetExecute(execute)
	}

	if cmd.Flags().Changed(consts.AttributeNameTemplate) && Spec.GetTemplateName() == "" {
		return cmdtxn.CmdErrorValidationMessage(cmd, "Error: %v", errors.New(errEmptyTemplateFlag))
	}

	// Process the --environment flag values to separate env vars from asset names
	err = Spec.ProcessRawEnvironment(cmd.ErrOrStderr())
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
	}

	err = Spec.Validate()
	if err != nil {
		return cmdtxn.CmdErrorVersionMessage(cmd, err)
	}
	cpUrl = instance.GetConfigValue(config.ControlPlaneUrl)
	cmdHelper.AddValidation(validators.NewProjectValidator(Project))
	cmdHelper.AddValidation(validators.NewNameValidator(Name))
	cmdHelper.AddValidation(validators.NewNameValidator(Spec.GetUserPrefix()))

	if msg, err := cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}

	if Name == specset.EmptyString && !Spec.HasPrefixName() {
		if Spec.GetType() == spectype.Workspace || Spec.GetType() == spectype.Inference {
			Spec.SetPrefix(string(Spec.GetType()))
		} else if Spec.GetFramework() != apitype.UndefinedFramework {
			Spec.SetPrefix(Spec.GetFramework().String())
		} else {
			Spec.SetPrefix(apitype.Runai.String())
		}
	}

	if Spec.UseSecurityRunAsUserSettings() {
		err = Spec.ValidateSecurityRunAsUserFlags()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags, err)
		}
	}

	return nil
}

// resolveTemplateId resolves template name to template ID if template is specified
func resolveTemplateId() error {
	templateName := Spec.GetTemplateName()
	if templateName == "" {
		return nil
	}

	templateId, err := templateservice.GetId(txn, cpUrl, token, templateName, clusterId, projectId, Spec.GetType())
	if err != nil {
		return err
	}

	Spec.SetTemplateId(templateId)
	return nil
}

// resolveEnvironmentAsset resolves environment name to environment ID
// Also validates that the environment supports the current workload type
func resolveEnvironmentAsset() error {
	envName := Spec.GetEnvironmentAsset()
	if envName == "" {
		return nil
	}

	envId, err := environmentservice.GetId(txn, cpUrl, token, envName, clusterId, projectId, Spec.GetType())
	if err != nil {
		return err
	}

	Spec.SetEnvironmentAsset(envId)
	return nil
}

// resolveComputeAsset resolves compute name to compute ID
func resolveComputeAsset() error {
	computeName := Spec.GetComputeAsset()
	if computeName == "" {
		return nil
	}

	computeId, err := computeservice.GetId(txn, cpUrl, token, computeName, clusterId, projectId)
	if err != nil {
		return err
	}

	Spec.SetComputeAsset(computeId)
	return nil
}

// parseDatasourceValue parses "type=TYPE,name=NAME" format
func parseDatasourceValue(value string) (name, dsType string, err error) {
	invalidFormatErr := fmt.Errorf("invalid datasource format '%s': required format is type=TYPE,name=NAME", value)

	mp := parser.NewStrictMapParser("datasource", mapset.NewSet[string]("name", "type"))
	am := parser.NewAttributeMapper(value, mp)
	if err := am.Parse(); err != nil {
		return "", "", invalidFormatErr
	}

	nameVal := am.Get("name")
	typeVal := am.Get("type")
	if nameVal == nil || typeVal == nil {
		return "", "", invalidFormatErr
	}

	name = strings.TrimSpace(*nameVal)
	dsType = strings.TrimSpace(*typeVal)
	if name == "" || dsType == "" {
		return "", "", invalidFormatErr
	}

	return name, dsType, nil
}

// resolveDatasourceAssets resolves datasource names to datasource IDs
func resolveDatasourceAssets() error {
	// Reset IDs before resolving
	Spec.SetDatasourceAssetIds(nil)
	Spec.SetMasterDatasourceAssetIds(nil)

	resolve := func(rawDatasources []string, addID func(string)) error {
		for _, rawDs := range rawDatasources {
			name, dsType, err := parseDatasourceValue(rawDs)
			if err != nil {
				return err
			}
			dsId, err := datasourceservice.GetId(txn, cpUrl, token, name, dsType, clusterId, projectId)
			if err != nil {
				return err
			}
			addID(dsId)
		}
		return nil
	}

	if err := resolve(Spec.GetRawDatasourceAssets(), Spec.AddDatasourceAssetId); err != nil {
		return err
	}
	return resolve(Spec.GetRawMasterDatasourceAssets(), Spec.AddMasterDatasourceAssetId)
}

// resolveDataVolumeAssets parses data volume flags and resolves names to IDs
func resolveDataVolumeAssets() error {
	rawDataVolumes := Spec.GetRawDataVolumes()
	if len(rawDataVolumes) == 0 {
		return nil
	}

	parsedDataVolumes, err := datavolume.ParseDataVolumeList(rawDataVolumes)
	if err != nil {
		return err
	}

	names := make([]string, len(parsedDataVolumes))
	for i, parsed := range parsedDataVolumes {
		names[i] = parsed.Name
	}

	nameToIdMap, err := datavolumeservice.GetNameToIdMap(txn, cpUrl, token, projectId, names)
	if err != nil {
		return err
	}

	resolvedInstances, err := datavolume.NewDataVolumeInstanceList(parsedDataVolumes, nameToIdMap)
	if err != nil {
		return err
	}

	Spec.SetResolvedDataVolumes(resolvedInstances)
	return nil
}

func Run(cmd *cobra.Command) error {
	var (
		err      error
		metadata *status.Metadata
	)

	txn = instance.GetTxn()
	token = instance.GetToken()
	clusterId = instance.GetConfigValue(config.ClusterUUID)

	projectId, err = projectservice.GetProjectId(txn, cpUrl, token, clusterId, Project)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
	}

	if Name == specset.EmptyString {
		Name = Spec.GetUserPrefix()
	}

	setCliCommandAsAnnotation()
	if Spec.UseSecurityRunAsUserSettings() {
		uid, gid, sup := calculateSecurityRunAsUserSettings(token)
		setSecurityRunAsUserSettings(uid, gid, sup)
	}

	// Resolve template ID before submitting
	if err = resolveTemplateId(); err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
	}

	// Resolve environment name to ID before submitting
	if err = resolveEnvironmentAsset(); err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
	}

	// Resolve compute name to ID before submitting
	if err = resolveComputeAsset(); err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
	}

	// Resolve datasource names to IDs before submitting
	if err = resolveDatasourceAssets(); err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
	}

	// Resolve data volume names to IDs before submitting
	if err = resolveDataVolumeAssets(); err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
	}

	if !Spec.Attach() {
		_, _, err = submit(cmd)
		return err
	}

	// attach mode
	var clusterUrl *url.URL
	metadata, err = cluster.GetStatus()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, failedClusterStatus, err)
	}
	clusterUrl, err = cluster.GetClusterApiUrl()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, failedClusterStatus, err)
	}
	instance.GetLogger().Debug("cluster metadata", zap.Any("metadata", metadata))

	var attachOptions = &attach.Options{
		Project:     Project,
		Stdin:       true,
		TTY:         true,
		WaitTimeout: Spec.GetPodRunningTimeout(),
	}

	submitResponse, workloadType, err := submit(cmd)
	if err != nil {
		return err
	}
	attachOptions.WorkloadName = submitResponse
	attachOptions.WorkloadType = workloadType
	attachOptions.WorkloadFramework = Spec.GetFramework()

	err = attachSubmittedWorkload(cmd, attachOptions, clusterUrl.String(), metadata)
	if err != nil {
		return err
	}
	return nil

}

func attachSubmittedWorkload(cmd *cobra.Command, attachOptions *attach.Options, clusterUrl string, metadata *status.Metadata) error {
	workload := apis.NewWorkload(attachOptions.Project, attachOptions.WorkloadType, attachOptions.WorkloadName, attachOptions.WorkloadFramework)
	err := cluster.WaitForWorkloadInClusterRunning(txn.GetCtx(), workload, clusterUrl, token, Spec.GetWaitForWorkloadSubmit())
	if err != nil {
		return err
	}
	return attach.Attach(txn.GetCtx(), cmd, attachOptions, token, clusterUrl, metadata)
}

func submit(cmd *cobra.Command) (string, apitype.WorkloadType, error) {
	var submitResponse string
	var err error
	var workloadType apitype.WorkloadType

	switch Spec.GetType() {
	case spectype.Inference:
		submitResponse, err = inference.Submit(txn, cpUrl, token, Name, projectId, clusterId, Spec)
		if err != nil {
			return "", apitype.Undefined, cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
		}
		workloadType = apitype.Inference
		printSubmitSuccessMessage(cmd, submitResponse)
	case spectype.InferenceDistributed:
		submitResponse, err = distributedinference.Submit(txn, cpUrl, token, Name, projectId, clusterId, Spec)
		if err != nil {
			return "", apitype.Undefined, cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
		}
		workloadType = apitype.DistributedInference
		printSubmitSuccessMessage(cmd, submitResponse)
	case spectype.NimService:
		submitResponse, err = nim.Submit(txn, cpUrl, token, Name, projectId, clusterId, Spec)
		if err != nil {
			return "", apitype.Undefined, cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
		}
		workloadType = apitype.Inference
		printSubmitSuccessMessage(cmd, submitResponse)
	case spectype.TrainingDistributed:
		submitResponse, err = distributed.Submit(txn, cpUrl, token, Name, projectId, clusterId, Spec)
		if err != nil {
			return "", apitype.Undefined, cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmitDistributed, err)
		}
		workloadType = apitype.Distributed
		printSubmitSuccessMessage(cmd, submitResponse)
	case spectype.Workspace:
		submitResponse, err = workspace.Submit(txn, cpUrl, token, Name, projectId, clusterId, Spec)
		if err != nil {
			return "", apitype.Undefined, cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
		}
		workloadType = apitype.Workspace
		printSubmitSuccessMessage(cmd, submitResponse)
	case spectype.Training:
		submitResponse, err = training.Submit(txn, cpUrl, token, Name, projectId, clusterId, Spec)
		if err != nil {
			return "", apitype.Undefined, cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmitTraining, err)
		}
		workloadType = apitype.Training
		printSubmitSuccessMessage(cmd, submitResponse)
	default:
		return "", apitype.Undefined, fmt.Errorf("submit command does not support type: %s", Spec.GetType().DisplayName())
	}

	return submitResponse, workloadType, nil
}

func printSubmitSuccessMessage(cmd *cobra.Command, submitResponse string) {
	if Spec.IsDryRun() {
		cmd.Println(submitResponse)
		return
	}

	switch workloadType := Spec.GetType(); workloadType {
	case spectype.Inference:
		cmd.Println(fmt.Sprintf(consts.SuccessSubmit, workloadType.DisplayName(), submitResponse, fmt.Sprintf(consts.FollowingInference, submitResponse)))
	case spectype.Workspace:
		cmd.Println(fmt.Sprintf(consts.SuccessSubmit, workloadType.DisplayName(), submitResponse, fmt.Sprintf(consts.FollowingWorkspace, submitResponse)))
	case spectype.Training:
		cmd.Println(fmt.Sprintf(consts.SuccessSubmit, workloadType.DisplayName(), submitResponse, fmt.Sprintf(consts.FollowingTraining, submitResponse)))
	case spectype.InferenceDistributed:
		cmd.Println(fmt.Sprintf(consts.SuccessSubmit, workloadType.DisplayName(), submitResponse, fmt.Sprintf(consts.FollowingDistributedInference, submitResponse)))
	case spectype.NimService:
		cmd.Println(fmt.Sprintf(consts.SuccessSubmit, workloadType.DisplayName(), submitResponse, fmt.Sprintf(consts.FollowingNimInference, submitResponse)))
	case spectype.TrainingDistributed:
		var msg string
		switch framework := Spec.GetFramework(); framework {
		case apitype.MPI, apitype.PyTorch, apitype.Jax, apitype.TF, apitype.XGBoost:
			msg = fmt.Sprintf(consts.FollowingDistributedTraining, framework, submitResponse)
		default:
			msg = consts.FollowingTraining
		}
		cmd.Println(fmt.Sprintf(consts.SuccessSubmit, workloadType.DisplayName(), submitResponse, msg))
	}
}

func setCliCommandAsAnnotation() {
	osArgs := os.Args
	arguments, _ := argshandler.GetExecuteString(&osArgs)
	Spec.SetCliCommandAsAnnotation(fmt.Sprintf(EnvFormat, spec.CommandAnnotationKey, arguments))
}
