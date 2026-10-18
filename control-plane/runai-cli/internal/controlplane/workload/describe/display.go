/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/distributedinference"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/inference"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/nim"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/field"
)

type Displayer interface {
	AsTable() error
	AsJSON() error
	AsYAML() error
}

type WorkloadDisplayer struct {
	workload *Workload
	options  displayer.DisplayOptions
}

func NewWorkloadDisplayer(workload *Workload, options displayer.DisplayOptions) *WorkloadDisplayer {
	return &WorkloadDisplayer{workload: workload, options: options}
}

func (wd *WorkloadDisplayer) AsTable() error {
	if wd.options.ShowGeneral {
		wd.DisplayGeneralAndRuntime()
	}

	if wd.workload.Inference != nil {
		wd.DisplayInference()
	}

	if wd.workload.DistributedInference != nil {
		wd.DisplayDistributedInference()
	}

	if wd.workload.Nim != nil {
		wd.DisplayNim()
	}

	if wd.workload.MasterConfig != nil {
		wd.displayRuntimeConfig(wd.workload.MasterConfig, "Master Configuration", true)
	}

	var replicas, workersPerReplica *int32
	if wd.workload.DistributedInference != nil {
		replicas = wd.workload.DistributedInference.Replicas
		workersPerReplica = wd.workload.DistributedInference.Workers
	}

	compute := wd.workload.Compute
	if wd.workload.Nim != nil {
		compute.GPUComputeMode = wd.workload.Nim.GpuComputeMode
	}

	sections := []struct {
		show    bool
		section *displayer.Section
	}{
		{true, displayer.NewIssuesSection(wd.workload.Issues)},
		{wd.options.ShowCompute, displayer.NewComputeSection(compute)},
		{wd.options.ShowStorage, displayer.NewStorageSection(wd.workload.Storage)},
		{wd.options.ShowNetworks, displayer.NewNetworksSection(wd.workload.Networks)},
		{true, displayer.NewNetworkTopologySection(wd.workload.NetworkTopology)},
		{wd.options.ShowPods, displayer.NewPodsSection(wd.workload.PodsInfo.Pods, wd.options.IncludeContainers, wd.workload.PodsInfo.RunningPods, replicas, workersPerReplica)},
		{wd.options.ShowEvents, displayer.NewEventsSection(wd.workload.Events)},
	}

	for _, s := range sections {
		if !s.show {
			continue
		}
		if err := displayer.DisplaySection(s.section); err != nil {
			return err
		}
	}

	return nil
}

func (wd *WorkloadDisplayer) DisplayGeneralAndRuntime() {
	columnPadding := displayer.GeneralAndRuntimeColumnPadding
	if wd.workload.GPUAcceleration != nil {
		columnPadding += displayer.MnnvlPaddingOffset
	}

	// For Distributed Inference, hide fields that are shown in Leader/Worker sections
	isDistributedInference := wd.workload.DistributedInference != nil
	isNim := wd.workload.Nim != nil

	fields := []displayer.DisplayField{
		{Name: displayer.FieldName, Value: wd.workload.General.Name},
		{Name: displayer.FieldID, Value: wd.workload.General.ID},
		{Name: displayer.FieldType, Value: wd.workload.General.Type},
		{Name: displayer.FieldPhase, Value: wd.workload.General.Phase},
		{Name: displayer.FieldPhaseMessage, Value: field.StringOrEmpty(wd.workload.General.PhaseMessage)},
		{Name: displayer.FieldCreatedAt, Value: wd.workload.General.CreatedAt.Format(displayer.DateFormat)},
		{Name: displayer.FieldUpdatedAt, Value: wd.workload.General.UpdatedAt.Format(displayer.DateFormat)},
		{Name: displayer.FieldProject, Value: wd.workload.General.Project},
		{Name: displayer.FieldDepartment, Value: wd.workload.General.Department},
		{Name: displayer.FieldSubmittedBy, Value: field.StringOrEmpty(wd.workload.General.SubmittedBy)},
		{Name: displayer.FieldPriority, Value: field.StringOrEmpty(wd.workload.General.Priority)},
		{Name: displayer.FieldCategory, Value: wd.workload.General.Category},
		{Name: displayer.FieldPreemptibility, Value: field.CapitalizeFirst(field.StringOrEmpty(wd.workload.General.Preemptibility))},
		{Name: wd.getCurrentNodePoolFieldName(), Value: strings.Join(wd.workload.Compute.CurrentNodePools, ", ")},
	}

	if isNim {
		if wd.workload.Nim.Image != nil {
			fields = append(fields,
				displayer.DisplayField{Name: displayer.FieldImageSingular, Value: field.StringOrEmpty(wd.workload.Nim.Image)},
			)
		}
	} else {
		fields = append(fields,
			displayer.DisplayField{Name: wd.getRequestedNodePoolFieldName(), Value: strings.Join(wd.workload.Compute.RequestedNodePools, ", ")},
			displayer.DisplayField{Name: displayer.FieldTopology, Value: field.StringOrEmpty(wd.workload.General.Topology)},
			displayer.DisplayField{Name: displayer.FieldDistributedFramework, Value: field.StringOrEmpty(wd.workload.General.DistributedFramework)},
			displayer.DisplayField{Name: wd.getImageFieldName(), Value: strings.Join(wd.workload.Runtime.Images, ", ")},
		)
	}

	fields = append(fields, displayer.DisplayField{})

	// Add these fields only for non-Distributed Inference workloads (they appear in Leader/Worker sections for Dist Inference)
	if !isDistributedInference && !isNim {
		additionalFields := []displayer.DisplayField{
			{Name: displayer.FieldImagePullPolicy, Value: field.StringOrEmpty(wd.workload.Runtime.ImagePullPolicy)},
			{Name: displayer.FieldImagePullSecrets, Value: strings.Join(wd.workload.Runtime.ImagePullSecrets, ", ")},
			{Name: displayer.FieldWorkingDir, Value: field.StringOrEmpty(wd.workload.Runtime.WorkingDir)},
			{Name: displayer.FieldNodeType, Value: field.StringOrEmpty(wd.workload.Runtime.NodeType)},
			{Name: displayer.FieldRunAsUser, Value: field.Int64OrEmpty(wd.workload.Runtime.RunAsUser)},
			{Name: displayer.FieldRunAsGroup, Value: field.Int64OrEmpty(wd.workload.Runtime.RunAsGroup)},
			{Name: displayer.FieldCapabilities, Value: strings.Join(wd.workload.Runtime.Capabilities, ", ")},
			{Name: displayer.FieldTolerations, Value: strings.Join(wd.workload.Runtime.Tolerations, ", ")},
			{Name: displayer.FieldReadinessProbe, Value: field.StringOrEmpty(wd.workload.Runtime.ReadinessProbe)},
		}
		// Insert before the empty field at the end
		fields = append(fields[:len(fields)-1], append(additionalFields, fields[len(fields)-1])...)
	}

	if len(wd.workload.Spec.MasterLabels) > 0 {
		fields = append(fields, displayer.DisplayField{Name: displayer.FieldWorkerLabels, Value: displayer.GetKeyValueDisplay(wd.workload.Spec.Labels)})
		fields = append(fields, displayer.DisplayField{Name: displayer.FieldMasterLabels, Value: displayer.GetKeyValueDisplay(wd.workload.Spec.MasterLabels)})
	} else if len(wd.workload.Spec.Labels) > 0 {
		fields = append(fields, displayer.DisplayField{Name: displayer.FieldLabels, Value: displayer.GetKeyValueDisplay(wd.workload.Spec.Labels)})
	}
	if len(wd.workload.Spec.Annotations) > 0 {
		fields = append(fields, displayer.DisplayField{Name: displayer.FieldAnnotations, Value: displayer.GetKeyValueDisplay(wd.workload.Spec.Annotations)})
	}

	// For Distributed Inference, Command/Arguments/Env Vars are shown in Leader/Worker sections, so skip them here
	if !isDistributedInference && !isNim {
		if wd.workload.General.Command != nil {
			fields = append(fields, displayer.DisplayField{Name: displayer.FieldCommand, Value: *wd.workload.General.Command})
		}

		if wd.workload.General.Arguments != nil {
			fields = append(fields, displayer.DisplayField{Name: displayer.FieldArguments, Value: *wd.workload.General.Arguments})
		}
	}

	if !isDistributedInference {
		fields = append(fields, displayer.DisplayField{Name: displayer.FieldEnvVars, Value: displayer.GetEnvironmentVariablesDisplay(wd.workload.General.EnvironmentVariables, columnPadding)})
	}

	if wd.workload.General.RunaiCliCommand != nil {
		fields = append(fields, displayer.DisplayField{Name: displayer.FieldCLICommand, Value: *wd.workload.General.RunaiCliCommand})
	}

	if wd.workload.GPUAcceleration != nil {
		fields = append(fields, displayer.DisplayField{Name: displayer.FieldGPUNetworkAcceleration, Value: *wd.workload.GPUAcceleration})
	}

	displayer.DisplayFields(fields, false, wd.workload.GPUAcceleration != nil)
}

func (wd *WorkloadDisplayer) DisplayInference() {
	fields := []displayer.DisplayField{
		{Name: inference.FieldLatestCreatedRevisionName, Value: field.StringOrEmpty(wd.workload.Inference.LatestCreatedRevisionName)},
		{Name: inference.FieldLatestReadyRevisionName, Value: field.StringOrEmpty(wd.workload.Inference.LatestReadyRevisionName)},
		{Name: inference.FieldMinReplicasName, Value: field.Int32OrEmpty(wd.workload.Inference.MinReplicas)},
		{Name: inference.FieldMaxReplicasName, Value: field.Int32OrEmpty(wd.workload.Inference.MaxReplicas)},
		{Name: inference.FieldInitialReplicasName, Value: field.Int32OrEmpty(wd.workload.Inference.InitialReplicas)},
		{Name: inference.FieldInitializationTimeoutSecondsName, Value: field.Int32OrEmpty(wd.workload.Inference.InitializationTimeoutSeconds)},
		{Name: inference.FieldRequestTimeoutSecondsName, Value: field.Int32OrEmpty(wd.workload.Inference.RequestTimeoutSeconds)},
		{Name: inference.FieldActivationReplicasName, Value: field.Int32OrEmpty(wd.workload.Inference.ActivationReplicas)},
		{Name: inference.FieldMetricName, Value: field.StringOrEmpty(wd.workload.Inference.Metric)},
		{Name: inference.FieldMetricThresholdPercentageName, Value: field.Float32OrEmpty(wd.workload.Inference.MetricThresholdPercentage)},
		{Name: inference.FieldMetricThresholdName, Value: field.Int32OrEmpty(wd.workload.Inference.MetricThreshold)},
		{Name: inference.FieldScaleDownDelaySecondsName, Value: field.Int32OrEmpty(wd.workload.Inference.ScaleDownDelaySeconds)},
		{Name: inference.FieldScaleToZeroRetentionSecondsName, Value: field.Int32OrEmpty(wd.workload.Inference.ScaleToZeroRetentionSeconds)},
		{Name: inference.FieldConcurrencyHardLimitName, Value: field.Int32OrEmpty(wd.workload.Inference.ConcurrencyHardLimit)},
	}

	displayer.DisplayFieldsWithSectionHeader(fields, inference.TitleServing, true)
}

func (wd *WorkloadDisplayer) DisplayDistributedInference() {
	configFields := []displayer.DisplayField{
		{Name: distributedinference.FieldRestartPolicy, Value: field.StringOrEmpty(wd.workload.DistributedInference.RestartPolicy)},
		{Name: distributedinference.FieldStartupPolicy, Value: field.StringOrEmpty(wd.workload.DistributedInference.StartupPolicy)},
		{Name: distributedinference.FieldReplicasRequested, Value: field.Int32OrEmpty(wd.workload.DistributedInference.Replicas)},
		{Name: distributedinference.FieldReplicasReady, Value: field.Int32OrEmpty(wd.workload.DistributedInference.ReplicasReady)},
		{Name: distributedinference.FieldWorkersPerReplica, Value: field.Int32OrEmpty(wd.workload.DistributedInference.Workers)},
	}
	displayer.DisplayFieldsWithSectionHeader(configFields, distributedinference.TitleDistributedInference, true)

	servingFields := []displayer.DisplayField{
		{Name: distributedinference.FieldServingPort, Value: field.Int32OrEmpty(wd.workload.DistributedInference.ServingPort)},
		{Name: distributedinference.FieldServingURL, Value: field.StringOrEmpty(wd.workload.DistributedInference.ServingURL)},
		{Name: distributedinference.FieldAuthorizationType, Value: field.StringOrEmpty(wd.workload.DistributedInference.AuthorizationType)},
		{Name: distributedinference.FieldAuthorizedUsers, Value: strings.Join(wd.workload.DistributedInference.AuthorizedUsers, ", ")},
		{Name: distributedinference.FieldAuthorizedGroups, Value: strings.Join(wd.workload.DistributedInference.AuthorizedGroups, ", ")},
		{Name: distributedinference.FieldExposeExternally, Value: field.BoolOrEmpty(wd.workload.DistributedInference.ExposeExternally)},
	}
	displayer.DisplayFieldsWithSectionHeader(servingFields, distributedinference.TitleServingConfiguration, true)

	if wd.workload.DistributedInference.Leader != nil {
		wd.displayRuntimeConfig(wd.workload.DistributedInference.Leader, distributedinference.TitleLeaderConfiguration, true)
	}

	if wd.workload.DistributedInference.Worker != nil {
		wd.displayRuntimeConfig(wd.workload.DistributedInference.Worker, distributedinference.TitleWorkerConfiguration, true)
	}
}

func (wd *WorkloadDisplayer) DisplayNim() {
	basicFields := []displayer.DisplayField{
		{Name: nim.FieldNgcAuthSecret, Value: field.StringOrEmpty(wd.workload.Nim.NgcAuthSecret)},
		{Name: nim.FieldReplicasRequested, Value: field.Int32OrEmpty(wd.workload.Nim.ReplicasRequested)},
		{Name: nim.FieldReplicasReady, Value: field.Int32OrEmpty(wd.workload.Nim.ReplicasReady)},
		{Name: nim.FieldMultiNode, Value: field.BoolToEnabled(wd.workload.Nim.MultiNodeEnabled)},
	}
	if field.BoolIsTrue(wd.workload.Nim.MultiNodeEnabled) {
		basicFields = append(basicFields,
			displayer.DisplayField{Name: nim.FieldWorkersPerReplica, Value: field.Int32OrEmpty(wd.workload.Nim.Workers)},
		)
	}
	displayer.DisplayFieldsWithSectionHeader(basicFields, nim.TitleNimService, true)

	if wd.workload.Nim.ModelStoreType != nil {
		modelStoreFields := []displayer.DisplayField{
			{Name: nim.FieldModelStoreType, Value: field.StringOrEmpty(wd.workload.Nim.ModelStoreType)},
			{Name: nim.FieldModelStoreClaimName, Value: field.StringOrEmpty(wd.workload.Nim.ModelStoreClaimName)},
			{Name: nim.FieldModelStoreAccessMode, Value: field.StringOrEmpty(wd.workload.Nim.ModelStoreAccessMode)},
			{Name: nim.FieldModelStoreCacheName, Value: field.StringOrEmpty(wd.workload.Nim.ModelStoreCacheName)},
			{Name: nim.FieldModelStoreProfile, Value: field.StringOrEmpty(wd.workload.Nim.ModelStoreProfile)},
		}
		displayer.DisplayFieldsWithSectionHeader(modelStoreFields, nim.TitleModelStore, true)
	}

	servingFields := []displayer.DisplayField{
		{Name: nim.FieldServingPort, Value: field.Int32OrEmpty(wd.workload.Nim.ServingPort)},
		{Name: nim.FieldGrpcPort, Value: field.Int32OrEmpty(wd.workload.Nim.GrpcPort)},
		{Name: nim.FieldMetricsPort, Value: field.Int32OrEmpty(wd.workload.Nim.MetricsPort)},
		{Name: nim.FieldExposeExternally, Value: field.BoolOrEmpty(wd.workload.Nim.ExposeExternally)},
		{Name: nim.FieldExposedProtocol, Value: field.StringOrEmpty(wd.workload.Nim.ExposedProtocol)},
		{Name: nim.FieldServingURL, Value: field.StringOrEmpty(wd.workload.Nim.ServingURL)},
		{Name: nim.FieldServiceType, Value: field.StringOrEmpty(wd.workload.Nim.ServiceType)},
	}
	displayer.DisplayFieldsWithSectionHeader(servingFields, nim.TitleServingConfiguration, true)

	if wd.workload.Nim.AutoscalingEnabled != nil && *wd.workload.Nim.AutoscalingEnabled {
		autoscalingFields := []displayer.DisplayField{
			{Name: nim.FieldMinReplicas, Value: field.Int32OrEmpty(wd.workload.Nim.MinReplicas)},
			{Name: nim.FieldMaxReplicas, Value: field.Int32OrEmpty(wd.workload.Nim.MaxReplicas)},
			{Name: nim.FieldMetric, Value: field.StringOrEmpty(wd.workload.Nim.Metric)},
			{Name: nim.FieldMetricThreshold, Value: field.Int32OrEmpty(wd.workload.Nim.MetricThreshold)},
			{Name: nim.FieldScaleWindow, Value: field.Int32OrEmpty(wd.workload.Nim.ScaleWindowSeconds)},
		}
		displayer.DisplayFieldsWithSectionHeader(autoscalingFields, nim.TitleAutoscaling, true)
	}

}

// displayRuntimeConfig displays a WorkloadRuntimeConfig (used for general runtime, leader, or worker)
func (wd *WorkloadDisplayer) displayRuntimeConfig(config *displayer.WorkloadRuntimeConfig, title string, indented bool) {
	if config == nil {
		return
	}

	columnPadding := displayer.GeneralAndRuntimeColumnPadding
	if indented {
		columnPadding += 4
	}

	fields := []displayer.DisplayField{
		{Name: displayer.FieldImageSingular, Value: field.StringOrEmpty(config.Image)},
		{Name: displayer.FieldImagePullPolicy, Value: field.StringOrEmpty(config.ImagePullPolicy)},
		{Name: displayer.FieldImagePullSecrets, Value: strings.Join(config.ImagePullSecrets, ", ")},
		{Name: displayer.FieldRunAsUser, Value: field.Int64OrEmpty(config.RunAsUser)},
		{Name: displayer.FieldRunAsGroup, Value: field.Int64OrEmpty(config.RunAsGroup)},
		{Name: displayer.FieldCapabilities, Value: strings.Join(config.Capabilities, ", ")},
		{Name: displayer.FieldTolerations, Value: strings.Join(config.Tolerations, ", ")},
		{Name: displayer.FieldReadinessProbe, Value: field.StringOrEmpty(config.ReadinessProbe)},
		{Name: displayer.FieldWorkingDir, Value: field.StringOrEmpty(config.WorkingDir)},
		{Name: displayer.FieldCommand, Value: field.StringOrEmpty(config.Command)},
		{Name: displayer.FieldArguments, Value: field.StringOrEmpty(config.Arguments)},
		{Name: displayer.FieldEnvVars, Value: displayer.GetEnvironmentVariablesDisplayForMap(config.EnvironmentVariables, columnPadding)},
	}

	// Add labels if present (used by Leader/Worker)
	if len(config.Labels) > 0 {
		fields = append(fields, displayer.DisplayField{
			Name:  displayer.FieldLabels,
			Value: displayer.FormatMapAsKeyValue(config.Labels),
		})
	}

	// Add node type
	fields = append(fields, displayer.DisplayField{
		Name:  displayer.FieldNodeType,
		Value: field.StringOrEmpty(config.NodeType),
	})

	// Add preemptible if present (used by general workloads)
	if config.Preemptible != nil {
		fields = append(fields, displayer.DisplayField{
			Name:  displayer.FieldPreemptible,
			Value: field.BoolOrEmpty(config.Preemptible),
		})
	}

	displayer.DisplayFieldsWithSectionHeader(fields, title, indented)

	// Display nested sections
	if config.Compute != nil {
		displayer.DisplayIndentedSection(displayer.NewComputeSection(*config.Compute), displayer.TitleCompute)
	}

	if config.Storage != nil {
		displayer.DisplayIndentedSection(displayer.NewStorageSection(config.Storage), displayer.TitleStorage)
	}
}

func (wd *WorkloadDisplayer) getImageFieldName() string {
	return getPluralizedFieldName(len(wd.workload.Runtime.Images), displayer.FieldImageSingular, displayer.FieldImagePlural)
}

func (wd *WorkloadDisplayer) getRequestedNodePoolFieldName() string {
	return getPluralizedFieldName(len(wd.workload.Compute.RequestedNodePools), displayer.FieldRequestedNodePoolSingular, displayer.FieldRequestedNodePoolPlural)
}

func (wd *WorkloadDisplayer) getCurrentNodePoolFieldName() string {
	return getPluralizedFieldName(len(wd.workload.Compute.CurrentNodePools), displayer.FieldCurrentNodePoolSingular, displayer.FieldCurrentNodePoolPlural)
}

// getPluralizedFieldName returns singular or plural field name based on count
func getPluralizedFieldName(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func (wd *WorkloadDisplayer) AsJSON() error {
	jsonData, err := json.MarshalIndent(wd.workload, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling to JSON: %w", err)
	}
	fmt.Println(string(jsonData)) //nolint:forbidigo
	return nil
}

func (wd *WorkloadDisplayer) AsYAML() error {
	yamlData, err := yaml.Marshal(wd.workload)
	if err != nil {
		return fmt.Errorf("error marshalling data: %w", err)
	}
	fmt.Println(string(yamlData)) //nolint:forbidigo
	return nil
}
