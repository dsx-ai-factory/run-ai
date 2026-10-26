/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cliflags"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/schedulig/affinity"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/submit"
	nodes "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/node"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/flags/update"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	policyconstants "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/constants"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	CommandAnnotationKey            = "runai-submit-command"
	UpdateCommandAnnotationKey      = "runai-update-command"
	DefaultWaitForWorkloadSubmitMin = 1
	DefaultPodRunningTimeoutMin     = 10
	completionsString               = "completions"
	parallelismString               = "parallelism"
	priorityClassString             = "priorityClass"
	servingPortString               = "servingPort"
	autoScalingString               = "autoScaling"
	servingConfigurationString      = "servingConfiguration"
)

var (
	CommandAnnotationStr         = CommandAnnotationKey
	errInvalidRestartPolicyInput = fmt.Errorf("invalid input, cannot set backoff limit when restart policy is %s", openapi.RESTARTPOLICY_NEVER)
)

type GeneralFlags struct {
	General              submit.General
	Container            submit.Container
	Compute              submit.Compute
	Storage              submit.Storage
	Network              submit.Network
	Lifecycle            submit.Lifecycle
	Security             submit.Security
	Scheduling           submit.Scheduling
	DistributedTraining  submit.DistributedTraining
	DistributedInference submit.DistributedInference
	Inference            submit.Inference
	Nim                  submit.NimService
}

type GeneralFullSpec struct {
	openapi.DistributedSpecSpec
	InferenceSpec
	PriorityClass openapi.NullableString `json:"priorityClass,omitempty"`
	Completions   openapi.NullableInt32  `json:"completions,omitempty"`
	Parallelism   openapi.NullableInt32  `json:"parallelism,omitempty"`
}

func (gfs *GeneralFullSpec) MarshalJSON() ([]byte, error) {
	byteArr, err := json.Marshal(gfs.DistributedSpecSpec)
	if err != nil {
		return nil, err
	}
	var toSerialize map[string]interface{}
	err = json.Unmarshal(byteArr, &toSerialize)
	if err != nil {
		return nil, err
	}

	if gfs.Completions.IsSet() {
		toSerialize[completionsString] = gfs.Completions.Get()
	}
	if gfs.Parallelism.IsSet() {
		toSerialize[parallelismString] = gfs.Parallelism.Get()
	}
	if gfs.PriorityClass.IsSet() {
		toSerialize[priorityClassString] = gfs.PriorityClass.Get()
	}
	if gfs.ServingPort.IsSet() {
		toSerialize[servingPortString] = gfs.ServingPort.Get()
	}
	if gfs.AutoScaling.IsSet() {
		toSerialize[autoScalingString] = gfs.AutoScaling.Get()
	}
	if gfs.ServingConfiguration.IsSet() {
		toSerialize[servingConfigurationString] = gfs.ServingConfiguration.Get()
	}
	return json.Marshal(toSerialize)
}

func (gfs *GeneralFullSpec) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &gfs)
}

// GeneralSpec using training spec that includes the workspace as well.
type GeneralSpec struct {
	fullSpec     GeneralFullSpec
	flags        GeneralFlags
	workloadType spectype.Workload
	framework    api.Framework
	grouper      *cliflags.FlagsByGroups
	templateId   string
}

func NewTrainingSpec() *GeneralSpec {
	return &GeneralSpec{
		workloadType: spectype.Training,
		framework:    api.Runai,
	}
}

func NewWorkspaceSpec() *GeneralSpec {
	return &GeneralSpec{
		workloadType: spectype.Workspace,
		framework:    api.Runai,
	}
}

func NewDistributedSpec(framework api.Framework) *GeneralSpec {
	workloadType := spectype.TrainingDistributed
	if framework == api.LWS {
		workloadType = spectype.InferenceDistributed
	}
	return &GeneralSpec{
		workloadType: workloadType,
		framework:    framework,
	}
}

func (gs *GeneralSpec) GetType() spectype.Workload {
	return gs.workloadType
}

func (gs *GeneralSpec) GetFramework() api.Framework {
	return gs.framework
}

func (gs *GeneralSpec) GetFlags() *GeneralFlags {
	return &gs.flags
}

func (gs *GeneralSpec) SetScaleFlags(cmd *cobra.Command) {
	switch gs.workloadType {
	case spectype.InferenceDistributed:
		scaleFlags := pflag.NewFlagSet("scale", pflag.ContinueOnError)
		scaleFlags.Int32Var(&gs.flags.DistributedInference.Replicas, consts.AttributeNameReplicas, 0, consts.AttributeUsageReplicas)

		scaleFlags.BoolVar(&gs.flags.General.DryRun, consts.AttributeNameDryRun, false, consts.AttributeUsageDryRun)
		cmd.Flags().AddFlagSet(scaleFlags)
	case spectype.NimService:
		scaleFlags := pflag.NewFlagSet("scale", pflag.ContinueOnError)
		scaleFlags.Int32Var(&gs.flags.Nim.Replicas, consts.AttributeNameReplicas, 0, consts.AttributeUsageReplicas)

		scaleFlags.BoolVar(&gs.flags.General.DryRun, consts.AttributeNameDryRun, false, consts.AttributeUsageDryRun)
		cmd.Flags().AddFlagSet(scaleFlags)
	}
}

func (gs *GeneralSpec) SetFlags(cmd *cobra.Command, project *string) {

	// adding submit dist attribute, that will fail once we delete the deprecated code.
	_ = fmt.Sprintf(consts.DepSubmitDistMsg)
	// check if grouper already created, if it does exist, this is the 'submit-dist' commands
	exist := gs.getGrouper(cmd)

	// since we are only populating flag sets, we can't register completion function - the flags must be registered to the cmd before
	completionFnsToRegister := make(map[string]completion.CompletionFn)

	// in this section we set the cmd flags for all workload types (workspace, training, inference, distributed)
	general := gs.grouper.GetOrAddFlagSet(consts.GroupNameGeneral)
	generalFlagSet(general, project, &gs.flags.General, gs.GetType())

	container := gs.grouper.GetOrAddFlagSet(consts.GroupNameContainer)
	submit.ContainerFlagSet(container, &gs.flags.Container, gs.GetType(), gs.framework)
	if gs.workloadType == spectype.NimService {
		container.StringArrayVar(&gs.flags.Nim.ImagePullSecrets, consts.AttributeNameImagePullSecrets, nil, consts.AttributeUsageImagePullSecrets)
	}

	compute := gs.grouper.GetOrAddFlagSet(consts.GroupNameCompute)
	submit.ComputerFlagSet(compute, &gs.flags.Compute, gs.GetType(), gs.framework, exist)

	if gs.workloadType != spectype.NimService {
		storage := gs.grouper.GetOrAddFlagSet(consts.GroupNameStorage)
		submit.StorageFlagSet(storage, &gs.flags.Storage, gs.GetType(), gs.framework)
	}

	if gs.workloadType != spectype.InferenceDistributed && gs.workloadType != spectype.NimService {
		network := gs.grouper.GetOrAddFlagSet(consts.GroupNameNetwork)
		submit.NetworkFlagSet(network, &gs.flags.Network)
	}

	lifecycle := gs.grouper.GetOrAddFlagSet(consts.GroupNameLifecycle)
	submit.LifecycleFlagSet(lifecycle, &gs.flags.Lifecycle, gs.workloadType, gs.framework)

	security := gs.grouper.GetOrAddFlagSet(consts.GroupNameSecurity)
	submit.SecurityFlagSet(security, &gs.flags.Security, gs.workloadType, gs.framework)

	scheduling := gs.grouper.GetOrAddFlagSet(consts.GroupNameScheduling)
	submit.SchedulingFlagSet(scheduling, &gs.flags.Scheduling, gs.workloadType, completionFnsToRegister)

	if gs.workloadType == spectype.Inference {
		inference := gs.grouper.GetOrAddFlagSet(consts.GroupNameInference)
		submit.InferenceCommonFlagSet(inference, &gs.flags.Inference)
	}

	if gs.workloadType == spectype.NimService {
		nimConfig := gs.grouper.GetOrAddFlagSet(consts.GroupNameNimConfiguration)
		submit.NimConfigurationFlagSet(nimConfig, &gs.flags.Nim)

		scaling := gs.grouper.GetOrAddFlagSet(consts.GroupNameScaling)
		submit.NimScalingFlagSet(scaling, &gs.flags.Nim)

		lifecycle := gs.grouper.GetOrAddFlagSet(consts.GroupNameLifecycle)
		submit.NimLifecycleFlagSet(lifecycle, &gs.flags.Nim)

		probes := gs.grouper.GetOrAddFlagSet(consts.GroupNameProbes)
		submit.NimProbesFlagSet(probes, &gs.flags.Nim)
	}

	if gs.workloadType == spectype.TrainingDistributed {
		distributed := gs.grouper.GetOrAddFlagSet(consts.GroupNameDistributed)
		submit.DistributedCommonFlagSet(distributed, &gs.flags.DistributedTraining)
		switch gs.framework {
		case api.MPI:
			submit.MPIFlagSet(distributed, &gs.flags.DistributedTraining)
		case api.XGBoost:
			submit.XgboostFlagSet(distributed, &gs.flags.DistributedTraining)
		case api.TF:
			submit.TensorFlowFlagSet(distributed, &gs.flags.DistributedTraining)
		case api.PyTorch:
			submit.PyTorchFlagSet(distributed, &gs.flags.DistributedTraining)
		case api.Jax:
			submit.JaxFlagSet(distributed, &gs.flags.DistributedTraining)
		default:
			submit.DistributedFlagSet(distributed, &gs.flags.DistributedTraining)
		}
	}

	if gs.workloadType == spectype.InferenceDistributed {
		distributed := gs.grouper.GetOrAddFlagSet(consts.GroupNameDistributedInference)
		submit.DistributedInferenceFlagSet(distributed, &gs.flags.DistributedInference)

		// Add distributed inference specific lifecycle flags
		lifecycle := gs.grouper.GetOrAddFlagSet(consts.GroupNameLifecycle)
		submit.DistInferenceLifecycleFlagSet(lifecycle, &gs.flags.Lifecycle)
		// Add worker-specific scheduling flags for distributed inference
		scheduling := gs.grouper.GetOrAddFlagSet(consts.GroupNameScheduling)
		submit.DistributedInferenceSchedulingFlagSet(scheduling, &gs.flags.DistributedInference)
	}

	gs.grouper.UpdateFlagsByGroupsToCmd(completionFnsToRegister)
	if gs.framework == api.PyTorch || gs.framework == api.TF {
		cmd.MarkFlagsMutuallyExclusive(consts.AttributeNameNoMaster, consts.AttributeNameMasterDatasourceAsset)
	}
}

func (gs *GeneralSpec) SetUpdateFlags(cmd *cobra.Command) {
	switch gs.workloadType {
	case spectype.Inference:
		gs.setInferenceUpdateFlags(cmd)
	case spectype.NimService:
		gs.setNimUpdateFlags(cmd)
	default:
		return
	}
}

func (gs *GeneralSpec) setInferenceUpdateFlags(cmd *cobra.Command) {
	_ = gs.getGrouper(cmd)

	// since we are only populating flag sets, we can't register completion function - the flags must be registered to the cmd before
	completionFnsToRegister := make(map[string]completion.CompletionFn)

	general := gs.grouper.GetOrAddFlagSet(consts.GroupNameGeneral)
	general.BoolVar(&gs.flags.General.DryRun, consts.AttributeNameDryRun, false, consts.AttributeUsageDryRun)
	_ = general.MarkHidden(consts.AttributeNameDryRun)

	container := gs.grouper.GetOrAddFlagSet(consts.GroupNameContainer)
	update.ContainerUpdateFlagSet(container, &gs.flags.Container, gs.GetType())

	inference := gs.grouper.GetOrAddFlagSet(consts.GroupNameInference)
	update.InferenceUpdateCommonFlagSet(inference, &gs.flags.Inference)

	compute := gs.grouper.GetOrAddFlagSet(consts.GroupNameCompute)
	update.InferenceUpdateComputeFlagSet(compute, &gs.flags.Compute)

	scheduling := gs.grouper.GetOrAddFlagSet(consts.GroupNameScheduling)
	update.InferenceUpdateNodeSelectionFlagSet(scheduling, &gs.flags.Scheduling, completionFnsToRegister)

	gs.grouper.UpdateFlagsByGroupsToCmd(completionFnsToRegister)
}

func (gs *GeneralSpec) setNimUpdateFlags(cmd *cobra.Command) {
	_ = gs.getGrouper(cmd)

	completionFnsToRegister := make(map[string]completion.CompletionFn)

	fieldRemoval := gs.grouper.GetOrAddFlagSet("Field Removal")
	fieldRemoval.StringSliceVar(&gs.flags.Nim.Remove, consts.AttributeNameRemove, nil, consts.AttributeUsageRemove)

	general := gs.grouper.GetOrAddFlagSet(consts.GroupNameGeneral)
	general.BoolVar(&gs.flags.General.DryRun, consts.AttributeNameDryRun, false, consts.AttributeUsageDryRun)
	_ = general.MarkHidden(consts.AttributeNameDryRun)
	general.StringArrayVar(&gs.flags.General.Labels, consts.AttributeNameLabel, []string{}, consts.AttributeUsageLabel)
	general.StringArrayVar(&gs.flags.General.Annotations, consts.AttributeNameAnnotation, []string{}, consts.AttributeUsageAnnotation)

	container := gs.grouper.GetOrAddFlagSet(consts.GroupNameContainer)
	update.ContainerUpdateFlagSet(container, &gs.flags.Container, gs.GetType())
	container.StringArrayVar(&gs.flags.Nim.ImagePullSecrets, consts.AttributeNameImagePullSecrets, nil, consts.AttributeUsageImagePullSecrets)

	nimConfig := gs.grouper.GetOrAddFlagSet(consts.GroupNameNimConfiguration)
	submit.NimConfigurationUpdateFlagSet(nimConfig, &gs.flags.Nim)

	scaling := gs.grouper.GetOrAddFlagSet(consts.GroupNameScaling)
	submit.NimScalingUpdateFlagSet(scaling, &gs.flags.Nim)

	compute := gs.grouper.GetOrAddFlagSet(consts.GroupNameCompute)
	update.NimUpdateComputeFlagSet(compute, &gs.flags.Compute)

	lifecycle := gs.grouper.GetOrAddFlagSet(consts.GroupNameLifecycle)
	submit.NimLifecycleUpdateFlagSet(lifecycle, &gs.flags.Lifecycle)
	lifecycle.StringVar(&gs.flags.Nim.Category, consts.AttributeNameCategory, "", consts.AttributeUsageCategory)
	lifecycle.StringVar(&gs.flags.Nim.PriorityClass, consts.AttributeNamePriorityClass, "", consts.AttributeUsagePriorityClass)

	security := gs.grouper.GetOrAddFlagSet(consts.GroupNameSecurity)
	submit.NimSecurityUpdateFlagSet(security, &gs.flags.Security)

	scheduling := gs.grouper.GetOrAddFlagSet(consts.GroupNameScheduling)
	update.NimUpdateSchedulingFlagSet(scheduling, &gs.flags.Scheduling)

	probes := gs.grouper.GetOrAddFlagSet(consts.GroupNameProbes)
	submit.NimProbesFlagSet(probes, &gs.flags.Nim)

	gs.grouper.UpdateFlagsByGroupsToCmd(completionFnsToRegister)
}

func (gs *GeneralSpec) getGrouper(cmd *cobra.Command) bool {
	if gs.grouper == nil {
		grouper := cliflags.NewFlagsByGroups(cmd)
		gs.grouper = &grouper
		return false
	}
	return true
}

func (gs *GeneralSpec) HasPrefixName() bool {
	return len(gs.flags.General.PrefixName) != 0 || gs.flags.General.UseNameAsPrefix
}

func (gs *GeneralSpec) IsDryRun() bool {
	return gs.flags.General.DryRun
}

func (gs *GeneralSpec) GetUserPrefix() string {
	return gs.flags.General.PrefixName
}

func (gs *GeneralSpec) SetPrefix(prefix string) {
	gs.flags.General.PrefixName = prefix
}

func (gs *GeneralSpec) SetCliCommandAsAnnotation(command string) {
	gs.flags.General.Annotations = append(gs.flags.General.Annotations, command)
}

func (gs *GeneralSpec) GetTemplateName() string {
	return gs.flags.General.Template
}

func (gs *GeneralSpec) SetTemplateId(templateId string) {
	gs.templateId = templateId
}

func (gs *GeneralSpec) GetTemplateId() string {
	return gs.templateId
}

func (gs *GeneralSpec) generateSpec(ctx context.Context, clusterId string, nodesService *nodes.BackendNodeService) error {
	var (
		err error
	)

	// reset any previously set spec, to avoid using any remainders from previous calls
	// (e.g. if master spec is built using the same GeneralSpec object that was used for the worker)
	gs.fullSpec = GeneralFullSpec{}

	err = gs.setGeneralProperties()
	if err != nil {
		return err
	}
	err = gs.setContainerProperties()
	if err != nil {
		return err
	}
	err = gs.setComputeProperties()
	if err != nil {
		return err
	}
	err = gs.setStorageProperties()
	if err != nil {
		return err
	}
	err = gs.setNetworkProperties()
	if err != nil {
		return err
	}
	gs.setLifecycleProperties()
	err = gs.setSecurityProperties()
	if err != nil {
		return err
	}
	err = gs.setSchedulingProperties(ctx, clusterId, nodesService)
	if err != nil {
		return err
	}
	if gs.GetType() == spectype.Inference {
		err = gs.setInferenceProperties()
	}
	if gs.GetType() == spectype.InferenceDistributed {
		err = gs.setDistributedProperties()
	}
	if gs.GetType() == spectype.TrainingDistributed {
		err = gs.setDistributedProperties()
	}
	if err != nil {
		return err
	}
	return nil
}

func (gs *GeneralSpec) GenerateWorkspaceSpec(ctx context.Context, clusterId string, nodesService *nodes.BackendNodeService) (*openapi.WorkspaceSpecSpec, error) {
	var (
		workspaceSpec = openapi.WorkspaceSpecSpec{}
		err           error
		bytes         []byte
	)
	err = gs.generateSpec(ctx, clusterId, nodesService)
	if err != nil {
		return nil, err
	}

	bytes, err = json.Marshal(&gs.fullSpec)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &workspaceSpec)
	if err != nil {
		return nil, err
	}
	return &workspaceSpec, nil
}

func (gs *GeneralSpec) GenerateTrainingSpec(ctx context.Context, clusterId string, nodesService *nodes.BackendNodeService) (*openapi.TrainingSpecSpec, error) {
	var (
		trainingSpec = openapi.TrainingSpecSpec{}
		err          error
		bytes        []byte
	)
	err = gs.generateSpec(ctx, clusterId, nodesService)
	if err != nil {
		return nil, err
	}

	bytes, err = json.Marshal(&gs.fullSpec)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &trainingSpec)
	if err != nil {
		return nil, err
	}
	return &trainingSpec, nil
}

func (gs *GeneralSpec) GenerateDistributedSpec(ctx context.Context, clusterId string, nodesService *nodes.BackendNodeService) (*openapi.DistributedSpec, error) {
	var (
		workerSpec = openapi.DistributedSpecSpec{}
		leaderSpec = openapi.DistributedSpecSpec{}
		err        error
		bytes      []byte
	)

	distributedSpec := openapi.NewDistributedSpec()

	// Special handling for Distributed Inference
	// For InferenceDistributed: base flags = leader, workers inherit or override based on --worker-as-leader flag
	if gs.workloadType == spectype.InferenceDistributed {
		// Generate leader spec using base flags
		err = gs.generateSpec(ctx, clusterId, nodesService)
		if err != nil {
			return nil, err
		}

		bytes, err = json.Marshal(&gs.fullSpec)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, &leaderSpec)
		if err != nil {
			return nil, err
		}

		// Check --worker-as-leader flag (default: true)
		// When true: workers inherit leader config unless overridden by --worker-* flags
		// When false: workers need explicit configuration
		workerAsLeader := gs.flags.DistributedInference.WorkerAsLeader

		// Determine if we need a separate worker spec
		needsSeparateWorkerSpec := !workerAsLeader || uniqueWorkerSpecRequired(&gs.flags)

		if needsSeparateWorkerSpec {
			// Generate separate worker spec with worker-specific overrides
			workerSpecPtr, err := GenerateWorkerSpec(gs, ctx, clusterId, nodesService)
			if err != nil {
				return nil, err
			}
			workerSpec = *workerSpecPtr

			// Set leader as master spec, worker as main spec
			distributedSpec.SetSpec(workerSpec)
			masterSpecFromLeader := openapi.MasterSpec{}
			bytes, err = json.Marshal(&leaderSpec)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(bytes, &masterSpecFromLeader)
			if err != nil {
				return nil, err
			}
			distributedSpec.SetMasterSpec(masterSpecFromLeader)
			distributedSpec.SetMasterSpecSameAsWorker(false)
		} else {
			// worker-as-leader=true AND no worker-specific overrides
			// Use same spec for both leader and workers
			distributedSpec.SetSpec(leaderSpec)
			distributedSpec.SetMasterSpecSameAsWorker(true)
		}

		return distributedSpec, nil
	}

	// Standard distributed training logic (MPI, PyTorch, TF, etc.)
	// For training: base flags = worker, master flags override for master
	err = gs.generateSpec(ctx, clusterId, nodesService)
	if err != nil {
		return nil, err
	}

	bytes, err = json.Marshal(&gs.fullSpec)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &workerSpec)
	if err != nil {
		return nil, err
	}

	distributedSpec.SetSpec(workerSpec)
	if !gs.flags.DistributedTraining.NoMaster && uniqueMasterSpecRequired(&gs.flags) {
		masterSpec, err := GenerateMasterSpec(gs, ctx, clusterId, nodesService) // TODO gs is not reset! it has the worker fields in it, can't rely on this func to reset everything (e.g. spec.Compute)
		if err != nil {
			return nil, err
		}
		distributedSpec.SetMasterSpecSameAsWorker(false)
		distributedSpec.SetMasterSpec(*masterSpec)
	} else if gs.flags.DistributedTraining.NoMaster {
		distributedSpec.SetMasterSpecSameAsWorker(false)
	} else {
		distributedSpec.SetMasterSpecSameAsWorker(true)
	}
	return distributedSpec, nil
}

func (gs *GeneralSpec) setGeneralProperties() error {
	err := gs.setLabels()
	if err != nil {
		return err
	}
	err = gs.setAnnotations()
	if err != nil {
		return err
	}
	return nil
}

func (gs *GeneralSpec) setLabels() error {
	var (
		general = &gs.flags.General
		labels  = &gs.fullSpec.Labels
	)

	if len(general.Labels) < 1 {
		return nil
	}
	for _, label := range general.Labels {
		split, err := specset.SplitEqualSepValidateTwoSection(label)
		if err != nil {
			return err
		}
		*labels = append(*labels, openapi.Label{
			Name:  *openapi.NewNullableString(&split[0]),
			Value: *openapi.NewNullableString(&split[1]),
		})
	}
	return nil
}

func (gs *GeneralSpec) setAnnotations() error {
	var (
		general     = &gs.flags.General
		annotations = &gs.fullSpec.Annotations
	)

	if len(general.Annotations) < 1 {
		return nil
	}
	for _, annot := range general.Annotations {
		// this is internal annotation, we know it might contain double `=` so we handle it (because we need it)
		if strings.HasPrefix(annot, CommandAnnotationKey) {
			handleCommandAnnotation(annot, annotations)
		} else if strings.HasPrefix(annot, UpdateCommandAnnotationKey) {
			// ignore the command annotation, it will be handled later
			continue
		} else {
			split, err := specset.SplitEqualSepValidateTwoSection(annot)
			if err != nil {
				return err
			}
			*annotations = append(*annotations, openapi.Annotation{
				Name:  *openapi.NewNullableString(&split[0]),
				Value: *openapi.NewNullableString(&split[1]),
			})
		}
	}
	return nil
}

func handleCommandAnnotation(annotation string, annotations *[]openapi.Annotation) {
	cutString := fmt.Sprintf("%s=", CommandAnnotationKey)
	value, ok := strings.CutPrefix(annotation, cutString)
	if !ok {
		return
	}

	for i := range *annotations {
		if (*annotations)[i].GetName() == CommandAnnotationStr {
			(*annotations)[i].SetValue(value)
			return
		}
	}
	*annotations = append(*annotations, openapi.Annotation{
		Name:  *openapi.NewNullableString(&CommandAnnotationStr),
		Value: *openapi.NewNullableString(&value),
	})
}

func (gs *GeneralSpec) setContainerProperties() error {
	var (
		spec      = &gs.fullSpec
		container = &gs.flags.Container
	)
	ecLen := len(container.ExecuteCommand)
	eaLen := len(container.ExecuteArgs)

	if ecLen > 0 && eaLen > 0 {
		specset.SetNullableString(container.ExecuteCommand, &spec.Command)
		specset.SetNullableString(container.ExecuteArgs, &spec.Args)
	} else if ecLen > 0 {
		specset.SetNullableString(container.ExecuteCommand, &spec.Command)
		spec.Args.Unset()
	} else {
		specset.SetNullableString(container.ExecuteArgs, &spec.Args)
		spec.Command.Unset()
	}

	specset.SetNullableString(container.Image, &spec.Image)
	err := specset.SetGenericType[openapi.ImagePullPolicy, openapi.NullableImagePullPolicy](
		container.ImagePullPolicy, openapi.NewImagePullPolicyFromValue,
		&spec.ImagePullPolicy, openapi.NewNullableImagePullPolicy)
	if err != nil && !errors.Is(err, specset.ErrEmptyValue) {
		return err
	}

	specset.SetNullableString(container.WorkingDir, &spec.WorkingDir)
	spec.SetCreateHomeDir(container.CreateHomeDir)
	if !container.Attach {
		specset.SetNullableBool(container.TTY, &spec.Tty)
		specset.SetNullableBool(container.Stdin, &spec.Stdin)
	} else {
		specset.SetNullableBool(true, &spec.Tty)
		specset.SetNullableBool(true, &spec.Stdin)
	}

	err = specset.SetEnvironmentVariables(
		container.EnvironmentVariables, container.FieldRef, container.EnvSecret,
		container.EnvMyCredentials,
		spec.SetEnvironmentVariables,
	)
	if err != nil {
		return err
	}

	if len(container.ImagePullMyCredentials) > 0 {
		err = specset.SetImagePullSecrets(container.ImagePullMyCredentials, func(ips []openapi.ImagePullSecret) {
			merged := append([]openapi.ImagePullSecret{}, spec.GetImagePullSecrets()...)
			spec.SetImagePullSecrets(append(merged, ips...))
		})
		if err != nil {
			return err
		}
	}

	if container.PodAffinityPreferred != specset.EmptyString {
		pa := openapi.NewPodAffinity()
		pa.SetType(openapi.PODAFFINITYTYPE_PREFERRED)
		pa.SetKey(container.PodAffinityPreferred)
		spec.SetPodAffinity(*pa)
	}

	if container.PodAffinityRequired != specset.EmptyString {
		pa := openapi.NewPodAffinity()
		pa.SetType(openapi.PODAFFINITYTYPE_REQUIRED)
		pa.SetKey(container.PodAffinityRequired)
		spec.SetPodAffinity(*pa)
	}

	return nil
}

func (gs *GeneralSpec) setComputeProperties() error {
	var (
		fc   = &gs.flags.Compute
		spec = &gs.fullSpec
	)
	compute, err := fc.GenerateComputeSpec()
	if err != nil {
		return err
	}

	if compute != nil {
		spec.SetCompute(*compute)
	}
	return nil
}

func (gs *GeneralSpec) setStorageProperties() error {
	var (
		s    = &gs.flags.Storage
		spec = &gs.fullSpec
	)

	storage, err := s.GenerateStorageSpec()
	if err != nil {
		return err
	}
	if storage != nil {
		spec.SetStorage(*storage)
	}
	return nil
}

func (gs *GeneralSpec) setNetworkProperties() error {
	var (
		network = &gs.flags.Network
		spec    = &gs.fullSpec
	)

	if len(network.Ports) > 0 {
		ports, err := network.GeneratePorts()
		if err != nil {
			return err
		}
		spec.SetPorts(ports)
	}

	if len(network.ExternalUrls) > 0 {
		eUrl, err := network.GenerateExternalUrls()
		if err != nil {
			return err
		}
		spec.SetExposedUrls(eUrl)
	}
	return nil
}

func (gs *GeneralSpec) setLifecycleProperties() {
	var (
		lifecycle = &gs.flags.Lifecycle
		spec      = &gs.fullSpec
	)

	if lifecycle.AutoDeletionTimeAfterCompletion.ToSec() > 0 {
		spec.SetAutoDeletionTimeAfterCompletionSeconds(lifecycle.AutoDeletionTimeAfterCompletion.ToSec())
	}

	if lifecycle.TerminationGracePeriod.ToSec() > 0 {
		spec.SetTerminationGracePeriodSeconds(lifecycle.TerminationGracePeriod.ToSec())
	}

	if lifecycle.BackoffLimit.IsSet() {
		spec.SetBackoffLimit(int64(*lifecycle.BackoffLimit.Get()))
	}

	if lifecycle.RestartPolicy.IsSet() {
		spec.SetRestartPolicy(*lifecycle.RestartPolicy.Get())
	}

	// Note: Distributed Inference-specific restart policy is handled separately in DistributedInferenceSpec generation
	// See GenerateDistributedInferenceSpec for Distributed Inference policy handling

	if lifecycle.Preemptible { // preemptible is a workspace flag
		spec.SetPriorityClass(policyconstants.PriorityClassInteractivePreemptible)
	}

	if lifecycle.Preemptibility.IsSet() {
		spec.Preemptibility = lifecycle.Preemptibility
	}

	if lifecycle.Completions > 0 {
		spec.SetCompletions(lifecycle.Completions)
	}

	if lifecycle.Parallelism > 0 {
		spec.SetParallelism(lifecycle.Parallelism)
	}
}

func (gs *GeneralSpec) SetSecurityRunAsUserSettings(uid, gid int64, supplementalGroups []int) {
	security := &gs.flags.Security
	security.RunAsUid = uid
	security.RunAsGid = gid
	security.SupplementalGroups = supplementalGroups
	security.UidGidSource = "custom"
}

func (gs *GeneralSpec) UseSecurityRunAsUserSettings() bool {
	return gs.flags.Security.RunAsUser
}

func (gs *GeneralSpec) ValidateSecurityRunAsUserFlags() error {
	return submit.ValidateSecurityRunAsUserFlags(&gs.flags.Security)
}

func (gs *GeneralSpec) setSecurityProperties() error {
	var (
		security = &gs.flags.Security
		spec     = &gs.fullSpec
	)

	s, err := security.GenerateSecuritySpec()
	if err != nil {
		return err
	}
	if s != nil {
		spec.SetSecurity(*s)
	}
	return nil
}

func (gs *GeneralSpec) setSchedulingProperties(ctx context.Context, clusterId string, nodesService *nodes.BackendNodeService) error {
	var (
		scheduling = &gs.flags.Scheduling
		spec       = &gs.fullSpec
	)

	if len(scheduling.NodePools) > 0 {
		spec.SetNodePools(scheduling.NodePools)
	}

	if len(scheduling.ExcludeNodes) > 0 {
		actualNodesToExclude, nonExistingNodes, err := nodesService.CheckIfNodesExist(ctx, clusterId, scheduling.ExcludeNodes)
		if err != nil {
			// user doesn't have permissions to list nodes - skip validation
			if isForbiddenError(err) {
				actualNodesToExclude = scheduling.ExcludeNodes
			} else {
				return err
			}
		}

		if len(nonExistingNodes) > 0 {
			//nolint:staticcheck // user facing message
			return fmt.Errorf("The following node names to exclude do not exist in the cluster: %s", strings.Join(nonExistingNodes, ", "))
		}

		affinityRequired := affinity.BuildNodeAffinityRequired(actualNodesToExclude)
		if affinityRequired != nil {
			spec.SetNodeAffinityRequired(*affinityRequired)
		}
	}

	if len(scheduling.Toleration) > 0 {
		t, err := scheduling.GenerateTolerations()
		if err != nil {
			return err
		}
		spec.SetTolerations(t)
	}
	//nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
	specset.SetNullableString(scheduling.NodeType, &spec.NodeType)

	if len(scheduling.PriorityClass) > 0 {
		spec.SetPriorityClass(scheduling.PriorityClass)
	}

	return nil
}

func (gs *GeneralSpec) setInferenceProperties() error {
	var (
		inference = &gs.flags.Inference
		spec      = &gs.fullSpec
	)

	// Set the serving port
	if len(inference.ServingPort) > 0 {
		servingPort, err := submit.ParseServingPort(inference.ServingPort)
		if err != nil {
			return err
		}
		spec.ServingPort.Set(servingPort)
	}

	// set the auto-scaling
	gs.setAutoScaling(inference, spec)

	gs.setServingConfiguration(inference, spec)

	return nil
}

func (gs *GeneralSpec) setDistributedProperties() error {
	var (
		distributed = &gs.flags.DistributedTraining
		spec        = &gs.fullSpec
	)

	if distributed.DistributedFramework != "" {
		framework, err := distributed.GenerateToDistributedFramework()
		if err != nil {
			return err
		}
		spec.SetDistributedFramework(*framework)
	}

	if distributed.Workers.IsSet() && distributed.Workers.Get() != nil {
		spec.SetNumWorkers(*distributed.Workers.Get())
	}
	if distributed.SegmentSize.IsSet() && distributed.SegmentSize.Get() != nil {
		spec.SetSegmentSize(*distributed.SegmentSize.Get())
	}
	if distributed.MinReplica > 0 {
		spec.SetMinReplicas(distributed.MinReplica)
	}
	if distributed.MaxReplica > 0 {
		spec.SetMaxReplicas(distributed.MaxReplica)
	}
	if distributed.SlotsPerWorker > 0 {
		spec.SetSlotsPerWorker(distributed.SlotsPerWorker)
	}
	if distributed.SshAuthMountPath != "" {
		spec.SetSshAuthMountPath(distributed.SshAuthMountPath)
	}
	if distributed.LauncherCreationPolicy.IsSet() {
		spec.MpiLauncherCreationPolicy = distributed.LauncherCreationPolicy
	}

	if distributed.CleanPodPolicy != "" {
		policy, err := openapi.NewCleanPodPolicyFromValue(distributed.CleanPodPolicy)
		if err != nil {
			return err
		}

		spec.SetCleanPodPolicy(*policy)
	}

	// Note: Distributed Inference specific fields (Replicas, StartupPolicy, RestartPolicy)
	// are handled in a separate DistributedInferenceSpec generation flow (see inference_spec.go).
	// The current DistributedSpec is used for distributed training (MPI, PyTorch, TF, etc.)

	return nil
}

func (gs *GeneralSpec) SetExecute(execute string) {
	if gs.flags.Container.Command {
		gs.flags.Container.ExecuteCommand = execute
	} else {
		gs.flags.Container.ExecuteArgs = execute
	}
}

func (gs *GeneralSpec) Clear() {
	gs.flags.General.Annotations = []string{}
	gs.fullSpec.Annotations = []openapi.Annotation{}
	gs.flags.Compute.GpuDeviceRequest.Unset()
	gs.flags.Compute.GpuPortionRequest.Unset()
	gs.flags.Compute.GpuPortionLimit.Unset()
	gs.flags.Compute.GpuMemoryPortionRequest.Unset()
	gs.flags.Compute.GpuMemoryPortionLimit.Unset()
	gs.flags.Compute.GpuComputeMode = ""
	gs.flags.Compute.CpuCoreRequest.Unset()
	gs.flags.Compute.CpuCoreLimit.Unset()
	gs.flags.Compute.MasterGpuDeviceRequest.Unset()
	gs.flags.Compute.MasterGpuPortionRequest.Unset()
	gs.flags.Compute.MasterGpuPortionLimit.Unset()
	gs.flags.Compute.MasterGpuMemoryPortionRequest.Unset()
	gs.flags.Compute.MasterGpuMemoryPortionLimit.Unset()
	gs.flags.Compute.MasterGpuComputeMode = ""
	gs.flags.Compute.WorkerGpuMemoryPortionRequest.Unset()
	gs.flags.Compute.WorkerGpuMemoryPortionLimit.Unset()
	gs.flags.Compute.WorkerGpuComputeMode = ""
	gs.flags.Compute.MasterCpuCoreRequest.Unset()
	gs.flags.Compute.MasterCpuCoreLimit.Unset()
	gs.flags.DistributedTraining.MinReplica = 0
	gs.flags.DistributedTraining.MaxReplica = 0
	gs.flags.DistributedTraining.SegmentSize.Unset()
	gs.flags.DistributedTraining.SlotsPerWorker = 0
	gs.flags.DistributedTraining.SshAuthMountPath = ""
	gs.flags.DistributedTraining.LauncherCreationPolicy.Unset()
	gs.flags.Security.RunAsUser = false
	gs.flags.Security.RunAsUid = 0
	gs.flags.Security.RunAsGid = 0
	gs.flags.Security.UidGidSource = ""
	gs.flags.Security.SupplementalGroups = nil
	gs.flags.General.PrefixName = ""
	gs.flags.Lifecycle.MasterRestartPolicy.Unset()
	gs.flags.Lifecycle.RestartPolicy.Unset()
}

func (gs *GeneralSpec) Attach() bool {
	return gs.flags.Container.Attach
}

func (gs *GeneralSpec) GetWaitForWorkloadSubmit() time.Duration {
	if gs.flags.Container.WaitForWorkloadSubmit != 0 {
		return gs.flags.Container.WaitForWorkloadSubmit
	}
	return time.Minute * DefaultWaitForWorkloadSubmitMin
}

func (gs *GeneralSpec) GetPodRunningTimeout() time.Duration {
	if gs.flags.Container.PodRunningTimeout != 0 {
		return gs.flags.Container.PodRunningTimeout
	}
	return time.Minute * DefaultPodRunningTimeoutMin
}

// SetPriorityClass gets a reference to the given NewNullableString and assigns it to the PriorityClass field.
func (gfs *GeneralFullSpec) SetPriorityClass(v string) {
	gfs.PriorityClass = *openapi.NewNullableString(&v)
}

// SetCompletions gets a reference to the given NullableInt32 and assigns it to the Completions field.
func (gfs *GeneralFullSpec) SetCompletions(v int32) {
	gfs.Completions = *openapi.NewNullableInt32(&v)
}

// SetParallelism gets a reference to the given NullableInt32 and assigns it to the Parallelism field.
func (gfs *GeneralFullSpec) SetParallelism(v int32) {
	gfs.Parallelism = *openapi.NewNullableInt32(&v)
}

// validateBackoffRestartPolicy If RestartPolicy=Never set on both, master and workers, block the option to set backoff limit
func (gs *GeneralSpec) validateBackoffRestartPolicy() error {
	never := openapi.RESTARTPOLICY_NEVER
	lc := gs.flags.Lifecycle
	if gs.workloadType != spectype.TrainingDistributed && gs.workloadType != spectype.InferenceDistributed {
		return nil
	}
	if !lc.BackoffLimit.IsSet() || *lc.BackoffLimit.Get() == 0 {
		return nil
	}
	if !lc.RestartPolicy.IsSet() || *lc.RestartPolicy.Get() != never {
		return nil
	}
	if lc.MasterRestartPolicy.IsSet() && *lc.MasterRestartPolicy.Get() != never {
		return nil
	}
	return errInvalidRestartPolicyInput
}

func (gs *GeneralSpec) Validate() error {
	err := gs.validateBackoffRestartPolicy()
	if err != nil {
		return err
	}
	topologyKeys := []struct {
		name  string
		value string
	}{
		{consts.AttributeNamePreferredPodAffinityKey, gs.flags.Container.PodAffinityPreferred},
		{consts.AttributeNameRequiredPodAffinityKey, gs.flags.Container.PodAffinityRequired},
		{consts.AttributeNameWorkerPreferredPodTopologyKey, gs.flags.DistributedInference.WorkerPreferredPodTopology},
		{consts.AttributeNameWorkerRequiredPodTopologyKey, gs.flags.DistributedInference.WorkerRequiredPodTopology},
	}
	for _, key := range topologyKeys {
		if strings.Contains(key.value, "=") {
			return fmt.Errorf(
				"invalid value %q for --%s: expected a Kubernetes label key only; do not include '=VALUE'",
				key.value,
				key.name,
			)
		}
	}
	return nil
}

// GetEnvironmentAsset returns the environment asset name
func (gs *GeneralSpec) GetEnvironmentAsset() string {
	return gs.flags.Container.EnvironmentAsset
}

// SetEnvironmentAsset sets the environment asset (name or resolved ID)
func (gs *GeneralSpec) SetEnvironmentAsset(value string) {
	gs.flags.Container.EnvironmentAsset = value
}

// GetComputeAsset returns the compute asset name
func (gs *GeneralSpec) GetComputeAsset() string {
	return gs.flags.Container.ComputeAsset
}

// SetComputeAsset sets the compute asset (name or resolved ID)
func (gs *GeneralSpec) SetComputeAsset(value string) {
	gs.flags.Container.ComputeAsset = value
}

// GetRawDatasourceAssets returns the raw datasource values from --datasource flag
func (gs *GeneralSpec) GetRawDatasourceAssets() []string {
	return gs.flags.Container.RawDatasourceAssets
}

// GetDatasourceAssetIds returns the resolved datasource IDs
func (gs *GeneralSpec) GetDatasourceAssetIds() []string {
	return gs.flags.Container.DatasourceAssetIds
}

// SetDatasourceAssetIds sets the resolved datasource IDs
func (gs *GeneralSpec) SetDatasourceAssetIds(ids []string) {
	gs.flags.Container.DatasourceAssetIds = ids
}

// AddDatasourceAssetId adds a resolved datasource ID to the list
func (gs *GeneralSpec) AddDatasourceAssetId(id string) {
	gs.flags.Container.DatasourceAssetIds = append(gs.flags.Container.DatasourceAssetIds, id)
}

// GetRawMasterDatasourceAssets returns the raw datasource values from --master-datasource.
func (gs *GeneralSpec) GetRawMasterDatasourceAssets() []string {
	return gs.flags.Container.RawMasterDatasourceAssets
}

// AddMasterDatasourceAssetId adds a resolved datasource ID for the master pod.
func (gs *GeneralSpec) AddMasterDatasourceAssetId(id string) {
	gs.flags.Container.MasterDatasourceAssetIds = append(gs.flags.Container.MasterDatasourceAssetIds, id)
}

// SetMasterDatasourceAssetIds sets the resolved datasource IDs for the master pod.
func (gs *GeneralSpec) SetMasterDatasourceAssetIds(ids []string) {
	gs.flags.Container.MasterDatasourceAssetIds = ids
}

func (gs *GeneralSpec) GetRawDataVolumes() []string {
	return gs.flags.Storage.GetRawDataVolumes()
}

func (gs *GeneralSpec) SetResolvedDataVolumes(resolved []openapi.DataVolumeInstance) {
	gs.flags.Storage.SetResolvedDataVolumes(resolved)
}

// GetAssetSubmissionIds returns AssetSubmissionIds if any asset IDs are configured, nil otherwise.
func (gs *GeneralSpec) GetAssetSubmissionIds() *openapi.AssetSubmissionIds {
	envId := gs.flags.Container.EnvironmentAsset
	computeId := gs.flags.Container.ComputeAsset
	datasourceIds := gs.flags.Container.DatasourceAssetIds

	if envId == "" && computeId == "" && len(datasourceIds) == 0 {
		return nil
	}

	assets := openapi.NewAssetSubmissionIds()
	if envId != "" {
		assets.SetEnvironment(envId)
	}
	if computeId != "" {
		assets.SetCompute(computeId)
	}
	if len(datasourceIds) > 0 {
		assets.SetDatasources(datasourceIds)
	}
	return assets
}

// GetMasterAssetSubmissionIds returns master pod asset IDs, if configured.
func (gs *GeneralSpec) GetMasterAssetSubmissionIds() *openapi.AssetSubmissionIds {
	if gs.flags.DistributedTraining.NoMaster {
		return nil
	}

	datasourceIds := gs.flags.Container.MasterDatasourceAssetIds
	if len(datasourceIds) == 0 {
		return nil
	}

	assets := openapi.NewAssetSubmissionIds()
	assets.SetDatasources(datasourceIds)
	return assets
}

// ProcessRawEnvironment processes the --environment flag values
// This should be called after flag parsing but before spec generation
func (gs *GeneralSpec) ProcessRawEnvironment(w io.Writer) error {
	return gs.flags.Container.ProcessRawEnvironment(w)
}

func generalFlagSet(generalFlags *pflag.FlagSet, project *string, generalSpec *submit.General, getType spectype.Workload) {
	generalFlags.StringVarP(project, consts.ProjectStr, "p", "", consts.ProjectAttributeExplanation)

	if getType == spectype.NimService {
		generalFlags.BoolVar(&generalSpec.UseNameAsPrefix, consts.AttributeNameUseNameAsPrefix, false, consts.AttributeUsageUseNameAsPrefix)
	} else if getType != spectype.Submit {
		generalFlags.StringVar(&generalSpec.PrefixName, consts.AttributeNamePrefixName, "", consts.AttributeUsagePrefix)
	} else {
		generalFlags.StringVar(&generalSpec.PrefixName, consts.AttributeNameJobNamePrefix, "", consts.AttributeUsagePrefix)
	}

	generalFlags.StringArrayVar(&generalSpec.Labels, consts.AttributeNameLabel, []string{}, consts.AttributeUsageLabel)
	generalFlags.StringArrayVar(&generalSpec.Annotations, consts.AttributeNameAnnotation, []string{}, consts.AttributeUsageAnnotation)
	generalFlags.BoolVar(&generalSpec.DryRun, consts.AttributeNameDryRun, false, consts.AttributeUsageDryRun)
	_ = generalFlags.MarkHidden(consts.AttributeNameDryRun)

	if getType != spectype.NimService {
		generalFlags.StringVar(&generalSpec.Template, consts.AttributeNameTemplate, "", consts.AttributeUsageTemplate)
	}
}

func isForbiddenError(err error) bool {
	var cliErr *clierrors.HttpError
	return errors.As(err, &cliErr) && strings.Contains(cliErr.Status, http.StatusText(http.StatusForbidden))
}
