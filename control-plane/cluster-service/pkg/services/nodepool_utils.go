/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/xhit/go-str2duration/v2"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	clients_local "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	"k8s.io/utils/ptr"
)

const (
	DefaultTimeBasedFairShareHistoricalUsageWeight = float32(1.0)
	DefaultTimeBasedFairShareDecayHalfLife         = "0h"
	DefaultTimeBasedFairShareWindowDuration        = "7d"
	DefaultTimeBasedFairShareWindowType            = api.TIMEBASEDFAIRSHAREWINDOWTYPE_SLIDING
)

func isCpuQuotaFFOn(ctx context.Context) bool {
	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	if tenantId == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Missing tenant id, assuming cpu_quota FF is Off")
		return false
	}
	setting, err := setting_pkg.Get[*settings.CPUResourcesQuota](*tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get cpu_quota setting, assuming the FF is Off. err: %v", err)
		return false
	}

	return setting.Enabled
}

func isAdjustableWeightsFFOn(ctx context.Context) bool {
	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	if tenantId == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Missing tenant id, assuming adjustable_weights FF is Off")
		return false
	}
	setting, err := setting_pkg.Get[*settings.ProjectManualOverQuota](*tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get adjustable_weights setting, assuming the FF is Off. err: %v", err)
		return false
	}

	return setting.Enabled
}

func isEditableNodepoolCapabilitiesOn() bool {
	return runai_config.Get[*config.EditableNodepoolCapabilities]().Enabled
}

func validateCreateParams(ctx context.Context, fields api.NodepoolCreateFields, capChecker clients_local.ClusterCapabilityChecker) error {
	validationErr := gpuNetworkAccelerationDetectionValidation(ctx, fields.GpuNetworkAccelerationDetection, capChecker)
	if validationErr != nil {
		return validationErr
	}

	validationErr = validateSchedulingConfiguration(ctx, fields.SchedulingConfiguration, capChecker, false, nil)
	if validationErr != nil {
		return validationErr
	}

	validationErr = validateSchedulingCapabilities(fields.SchedulingCapabilities)
	if validationErr != nil {
		return validationErr
	}

	return nil
}

func validateUpdateParams(ctx context.Context, fields api.NodepoolUpdateFields, capChecker clients_local.ClusterCapabilityChecker) error {
	return validateUpdatePatchParamsInner(ctx, fields, true, nil, capChecker)
}

func validatePatchParams(ctx context.Context, dbNodepool models.Nodepool, fields api.NodepoolUpdateFields, capChecker clients_local.ClusterCapabilityChecker) error {
	return validateUpdatePatchParamsInner(ctx, fields, false, &dbNodepool, capChecker)
}

func validateUpdatePatchParamsInner(ctx context.Context, fields api.NodepoolUpdateFields,
	isUpdate bool, dbNodepool *models.Nodepool, capChecker clients_local.ClusterCapabilityChecker,
) error {
	validationErr := validateSchedulingCapabilities(fields.SchedulingCapabilities)
	if validationErr != nil {
		return validationErr
	}

	if fields.GpuNetworkAccelerationDetection == nil && fields.SchedulingConfiguration == nil {
		// no need to even check capabilities if no need to
		return nil
	}

	validationErr = gpuNetworkAccelerationDetectionValidation(ctx, fields.GpuNetworkAccelerationDetection, capChecker)
	if validationErr != nil {
		return validationErr
	}

	validationErr = validateSchedulingConfiguration(ctx, fields.SchedulingConfiguration, capChecker, !isUpdate, dbNodepool)
	if validationErr != nil {
		return validationErr
	}

	return nil
}

func validateSchedulingConfiguration(
	ctx context.Context,
	schedulingConfiguration *api.SchedulingConfiguration,
	capChecker clients_local.ClusterCapabilityChecker,
	isPatch bool,
	dbNodepool *models.Nodepool,
) error {
	if schedulingConfiguration == nil {
		return nil
	}

	if err := validatePlacementStrategyEnums(schedulingConfiguration); err != nil {
		return err
	}

	if err := validateGpuDevicePlacementSupport(ctx, schedulingConfiguration, capChecker); err != nil {
		return err
	}

	if err := validateMinGuaranteedRuntime(ctx, schedulingConfiguration, capChecker); err != nil {
		return err
	}

	if err := validateTimeBasedFairShare(ctx, schedulingConfiguration.TimeBasedFairShare, capChecker, isPatch, dbNodepool); err != nil {
		return err
	}

	if err := validateNumaAwareScheduling(ctx, schedulingConfiguration, capChecker); err != nil {
		return err
	}

	return nil
}

func validateNumaAwareScheduling(ctx context.Context, schedulingConfiguration *api.SchedulingConfiguration, capChecker clients_local.ClusterCapabilityChecker) error {
	if schedulingConfiguration.NumaAwareScheduling == nil {
		return nil
	}

	clusterSupported, err := capChecker.ValidateClusterCapability(ctx, clients_local.LocalMinVersionOfNumaAwareScheduling)
	if err != nil {
		return fmt.Errorf("failed to check cluster support for numaAwareScheduling: %w", err)
	}
	if !clusterSupported {
		return fmt.Errorf("numaAwareScheduling supported from cluster version %s", clients_local.LocalMinVersionForSupporting(clients_local.LocalMinVersionOfNumaAwareScheduling))
	}

	return nil
}

func validatePlacementStrategyEnums(sc *api.SchedulingConfiguration) error {
	if sc.PlacementStrategy.Gpu != nil && !sc.PlacementStrategy.Gpu.IsValid() {
		return fmt.Errorf("invalid placementStrategy.gpu: %v", *sc.PlacementStrategy.Gpu)
	}
	if sc.PlacementStrategy.Cpu != nil && !sc.PlacementStrategy.Cpu.IsValid() {
		return fmt.Errorf("invalid placementStrategy.cpu: %v", *sc.PlacementStrategy.Cpu)
	}
	if sc.GpuPodsPlacementStrategy != nil {
		if sc.GpuPodsPlacementStrategy.Node != nil && !sc.GpuPodsPlacementStrategy.Node.IsValid() {
			return fmt.Errorf("invalid gpuPodsPlacementStrategy.node: %v", *sc.GpuPodsPlacementStrategy.Node)
		}
		if sc.GpuPodsPlacementStrategy.Device != nil && !sc.GpuPodsPlacementStrategy.Device.IsValid() {
			return fmt.Errorf("invalid gpuPodsPlacementStrategy.device: %v", *sc.GpuPodsPlacementStrategy.Device)
		}
	}
	if sc.CpuPodsPlacementStrategy != nil {
		if sc.CpuPodsPlacementStrategy.Node != nil && !sc.CpuPodsPlacementStrategy.Node.IsValid() {
			return fmt.Errorf("invalid cpuPodsPlacementStrategy.node: %v", *sc.CpuPodsPlacementStrategy.Node)
		}
	}
	return nil
}

// validateGpuDevicePlacementSupport enforces the cluster-version constraint:
// on clusters older than LocalMinVersionOfGpuDevicePlacementStrategy,
// gpuPodsPlacementStrategy.device cannot be set at all. The cluster can't honor
// an independent device strategy — it mirrors node onto device. To express
// placement on < 2.26 clusters, use gpuPodsPlacementStrategy.node only.
func validateGpuDevicePlacementSupport(
	ctx context.Context,
	schedulingConfiguration *api.SchedulingConfiguration,
	capChecker clients_local.ClusterCapabilityChecker,
) error {
	if schedulingConfiguration.GpuPodsPlacementStrategy == nil {
		return nil
	}
	if schedulingConfiguration.GpuPodsPlacementStrategy.Device == nil {
		return nil
	}
	supported, err := capChecker.ValidateClusterCapability(
		ctx, clients_local.LocalMinVersionOfGpuDevicePlacementStrategy,
	)
	if err != nil {
		return fmt.Errorf("failed to check cluster support for gpuPodsPlacementStrategy.device placement strategy: %w", err)
	}
	if supported {
		return nil
	}
	return fmt.Errorf(
		"on cluster versions older than %s, gpuPodsPlacementStrategy.device cannot be set",
		clients_local.LocalMinVersionForSupporting(clients_local.LocalMinVersionOfGpuDevicePlacementStrategy),
	)
}

func validateMinGuaranteedRuntime(ctx context.Context, schedulingConfiguration *api.SchedulingConfiguration, capChecker clients_local.ClusterCapabilityChecker) error {
	if schedulingConfiguration.MinGuaranteedRuntime == nil ||
		*schedulingConfiguration.MinGuaranteedRuntime == "" {
		return nil
	}

	clusterSupported, err := capChecker.ValidateClusterCapability(ctx, clients_local.LocalMinVersionOfMinGuaranteedRuntime)
	if err != nil {
		return fmt.Errorf("failed to check cluster support for MinGuaranteedRuntime. err: %v", err)
	}
	if !clusterSupported {
		return fmt.Errorf("MinGuaranteedRuntime supported from cluster version %s", clients_local.LocalMinVersionForSupporting(clients_local.LocalMinVersionOfMinGuaranteedRuntime))
	}

	minWorkloadRuntime := schedulingConfiguration.MinGuaranteedRuntime
	err = parseValidateDuration(minWorkloadRuntime, "MinGuaranteedRuntime")
	if err != nil {
		return err
	}

	return nil
}

func parseValidateDuration(durationStr *string, fieldName string) error {
	if durationStr == nil {
		return nil
	}

	duration, err := str2duration.ParseDuration(*durationStr)
	if err != nil {
		return fmt.Errorf("failed to parse %s '%s', error: %s", fieldName, *durationStr, err.Error())
	}
	if duration < 0 {
		return fmt.Errorf("%s '%s' cannot be negative", fieldName, *durationStr)
	}

	return nil
}

func validateTimeBasedFairShare(
	ctx context.Context,
	requestedTimeBasedFairShare *api.TimeBasedFairShare,
	capChecker clients_local.ClusterCapabilityChecker,
	isPatch bool,
	dbNodepool *models.Nodepool,
) error {
	if requestedTimeBasedFairShare == nil {
		return nil
	}

	tbfsEnabled := false
	if requestedTimeBasedFairShare.Enabled != nil {
		tbfsEnabled = *requestedTimeBasedFairShare.Enabled
	} else if isPatch && dbNodepool != nil && dbNodepool.TimeBasedFairShare.Enabled != nil {
		tbfsEnabled = *dbNodepool.TimeBasedFairShare.Enabled
	}

	if tbfsEnabled {
		clusterSupported, err := capChecker.ValidateClusterCapability(ctx, clients_local.LocalMinVersionOfTimeBasedFairShare)
		if err != nil {
			return fmt.Errorf("failed to check cluster support for Time-Based Fair Share. err: %v", err)
		}
		if !clusterSupported {
			return fmt.Errorf("Time-Based Fair Share supported from cluster version %s", clients_local.LocalMinVersionForSupporting(clients_local.LocalMinVersionOfTimeBasedFairShare))
		}
	}

	// When Time-Based Fair Share is disabled, no other fields are allowed.
	if !tbfsEnabled {
		if requestedTimeBasedFairShare.HistoricalUsageWeight != nil ||
			requestedTimeBasedFairShare.DecayHalfLife != nil ||
			requestedTimeBasedFairShare.WindowDuration != nil ||
			requestedTimeBasedFairShare.WindowType != nil ||
			requestedTimeBasedFairShare.TumblingWindowStartTime != nil {
			return fmt.Errorf("timeBasedFairShare fields must not be set when enabled is false")
		}
		return nil
	}

	err := parseValidateDuration(requestedTimeBasedFairShare.DecayHalfLife, "DecayHalfLife")
	if err != nil {
		return err
	}
	err = parseValidateDuration(requestedTimeBasedFairShare.WindowDuration, "WindowDuration")
	if err != nil {
		return err
	}

	err = validateTimeBasedFairShareWindowTypeAndStartTime(requestedTimeBasedFairShare, isPatch, dbNodepool)
	if err != nil {
		return err
	}

	return nil
}

func validateTimeBasedFairShareWindowTypeAndStartTime(
	requestedTimeBasedFairShare *api.TimeBasedFairShare,
	isPatch bool,
	dbNodepool *models.Nodepool,
) error {
	requestedWindowType := DefaultTimeBasedFairShareWindowType // sliding by default
	if requestedTimeBasedFairShare.WindowType != nil {
		requestedWindowType = *requestedTimeBasedFairShare.WindowType
	} else if isPatch && dbNodepool != nil && dbNodepool.TimeBasedFairShare.WindowType != nil {
		requestedWindowType = *dbNodepool.TimeBasedFairShare.WindowType
	}

	requestedTumblingWindowStartTime := time.Time{}
	if requestedTimeBasedFairShare.TumblingWindowStartTime != nil {
		requestedTumblingWindowStartTime = *requestedTimeBasedFairShare.TumblingWindowStartTime
	} else if isPatch && dbNodepool != nil && dbNodepool.TimeBasedFairShare.TumblingWindowStartTime != nil {
		requestedTumblingWindowStartTime = *dbNodepool.TimeBasedFairShare.TumblingWindowStartTime
	}

	if requestedWindowType == api.TIMEBASEDFAIRSHAREWINDOWTYPE_TUMBLING && requestedTumblingWindowStartTime.IsZero() {
		return fmt.Errorf("tumblingWindowStartTime must be set when windowType is tumbling")
	}

	return nil
}

func validateSchedulingCapabilities(sc api.SchedulingCapabilitiesNullable) error {
	if sc.CpuQuota != nil && *sc.CpuQuota && sc.Quota != nil && !*sc.Quota {
		return fmt.Errorf("SchedulingCapabilities.CpuQuota can only be enabled when SchedulingCapabilities.Quota is enabled")
	}

	// not checking min version for "ranks" - will be enforced in the org-unit-service

	return nil
}

func setNetworkTopology(ctx context.Context, tenantId int32, nodepool *models.Nodepool, networkTopologyId *string) *api.Error {
	if networkTopologyId == nil {
		nodepool.NetworkTopologyId = nil
		nodepool.NetworkTopologyName = nil
		return nil
	}
	networkTopologyUiId, parseErr := uuid.Parse(*networkTopologyId)
	if parseErr != nil {
		return &api.Error{Code: http.StatusBadRequest, Message: parseErr.Error()}
	}
	networkTopology, ntErr := dao.GetNetworkTopologyById(ctx, tenantId, networkTopologyUiId.String())
	if ntErr != nil {
		return ntErr
	}
	nodepool.NetworkTopologyId = &networkTopologyUiId
	nodepool.NetworkTopologyName = &networkTopology.Name

	return nil
}

// resolveRequestedPlacement returns the effective CPU node, GPU node, and GPU
// device values requested by the caller (in that order - cpu, gpuNode, gpuDevice),
// applying the precedence rules:
//   - The new gpuPodsPlacementStrategy / cpuPodsPlacementStrategy fields take
//     precedence over the deprecated placementStrategy.{cpu,gpu}.
//   - The deprecated placementStrategy.gpu updates BOTH node and device (it has
//     always been a combined GPU strategy); the new fields can selectively
//     override either of them.
//
// Fields the caller did not provide are returned as "" (empty PlacementStrategy).
func resolveRequestedPlacement(sc *api.SchedulingConfiguration) (api.PlacementStrategy, api.PlacementStrategy, api.PlacementStrategy) {
	if sc == nil {
		return "", "", ""
	}

	var cpu, gpuNode, gpuDevice api.PlacementStrategy

	// CPU node: new wins; else deprecated cpu.
	if sc.CpuPodsPlacementStrategy != nil && sc.CpuPodsPlacementStrategy.Node != nil {
		cpu = *sc.CpuPodsPlacementStrategy.Node
	} else if sc.PlacementStrategy.Cpu != nil {
		cpu = *sc.PlacementStrategy.Cpu
	}

	// GPU node + device: start from the deprecated combined gpu (mirrors to both),
	// then let any new field override.
	if sc.PlacementStrategy.Gpu != nil {
		gpuNode = *sc.PlacementStrategy.Gpu
		gpuDevice = *sc.PlacementStrategy.Gpu
	}
	if sc.GpuPodsPlacementStrategy != nil {
		if sc.GpuPodsPlacementStrategy.Node != nil {
			gpuNode = *sc.GpuPodsPlacementStrategy.Node
		}
		if sc.GpuPodsPlacementStrategy.Device != nil {
			gpuDevice = *sc.GpuPodsPlacementStrategy.Device
		}
	}
	return cpu, gpuNode, gpuDevice
}

// stripUnsupportedSchedulingFields removes fields from a response-bound
// SchedulingConfiguration that the target cluster cannot honor. On clusters
// older than LocalMinVersionOfGpuDevicePlacementStrategy, the agent mirrors
// the node-level strategy onto the device axis and the validator rejects
// divergent values — so returning gpuPodsPlacementStrategy.device in responses
// traps clients in a GET→PUT round-trip break.
func stripUnsupportedSchedulingFields(
	ctx context.Context,
	sc *api.SchedulingConfiguration,
	capChecker clients_local.ClusterCapabilityChecker,
) error {
	if sc == nil || sc.GpuPodsPlacementStrategy == nil || sc.GpuPodsPlacementStrategy.Device == nil {
		return nil
	}
	supportsDevicePlacement, err := capChecker.ValidateClusterCapability(
		ctx, clients_local.LocalMinVersionOfGpuDevicePlacementStrategy,
	)
	if err != nil {
		if errors.Is(err, dao.ErrClusterNotFound) {
			// Orphan nodepool — strip defensively rather than 500.
			logging.Warnf("Nodepool references missing cluster; stripping gpuPodsPlacementStrategy.device. err: %v", err)
			sc.GpuPodsPlacementStrategy.Device = nil
			return nil
		}
		return fmt.Errorf("failed to check cluster support for gpuPodsPlacementStrategy.device placement strategy: %w", err)
	}
	if !supportsDevicePlacement {
		sc.GpuPodsPlacementStrategy.Device = nil
	}
	return nil
}

// alignDeviceForClusterCapability forces device = node on clusters that do not
// yet support an independent device strategy. Validation rejects unequal
// caller-supplied values up front; this is the safety-net for cases where the
// device was implicitly derived (only node provided).
func alignDeviceForClusterCapability(
	ctx context.Context,
	gpuNode, gpuDevice api.PlacementStrategy,
	capChecker clients_local.ClusterCapabilityChecker,
) (api.PlacementStrategy, error) {
	supports, err := capChecker.ValidateClusterCapability(
		ctx, clients_local.LocalMinVersionOfGpuDevicePlacementStrategy,
	)
	if err != nil {
		return "", fmt.Errorf("failed to check cluster support for gpuDevice placement strategy: %w", err)
	}
	if !supports {
		return gpuNode, nil
	}
	return gpuDevice, nil
}

func setSchedulingConfigurationFieldsOnCreateUpdateInternal(
	ctx context.Context,
	nodePool *models.Nodepool,
	requestedSchedulingConfigurationFields *api.SchedulingConfiguration,
	capChecker clients_local.ClusterCapabilityChecker,
) error {
	if nodePool == nil {
		return nil
	}

	cpu, gpuNode, gpuDevice := resolveRequestedPlacement(requestedSchedulingConfigurationFields)

	if requestedSchedulingConfigurationFields != nil {
		resolvedDevice, err := alignDeviceForClusterCapability(ctx, gpuNode, gpuDevice, capChecker)
		if err != nil {
			return err
		}
		gpuDevice = resolvedDevice

		nodePool.MinGuaranteedRuntime = requestedSchedulingConfigurationFields.MinGuaranteedRuntime
		nodePool.TimeBasedFairShare = setTimeBasedFairShareDefaults(requestedSchedulingConfigurationFields.TimeBasedFairShare)
		nodePool.NumaAwareScheduling = requestedSchedulingConfigurationFields.NumaAwareScheduling
	} else {
		nodePool.TimeBasedFairShare = models.TimeBasedFairShare{Enabled: ptr.To(false)}
		nodePool.NumaAwareScheduling = nil
	}

	nodePool.CpuPlacementStrategy = cpu
	nodePool.GpuPlacementStrategy = gpuNode
	nodePool.GpuDevicePlacementStrategy = gpuDevice
	return nil
}

func setTimeBasedFairShareDefaults(input *api.TimeBasedFairShare) models.TimeBasedFairShare {
	// If disabled, we keep only the Enabled flag; validators already ensured no other fields are set.
	if input == nil || input.Enabled == nil || !*input.Enabled {
		return models.TimeBasedFairShare{Enabled: ptr.To(false)}
	}

	result := models.TimeBasedFairShare{
		Enabled:                 input.Enabled,
		HistoricalUsageWeight:   input.HistoricalUsageWeight,
		DecayHalfLife:           input.DecayHalfLife,
		WindowDuration:          input.WindowDuration,
		WindowType:              input.WindowType,
		TumblingWindowStartTime: input.TumblingWindowStartTime,
	}

	// Apply defaults when enabled and fields are not specified
	if result.HistoricalUsageWeight == nil {
		result.HistoricalUsageWeight = ptr.To(DefaultTimeBasedFairShareHistoricalUsageWeight)
	}
	if result.DecayHalfLife == nil {
		result.DecayHalfLife = ptr.To(DefaultTimeBasedFairShareDecayHalfLife)
	}
	if result.WindowDuration == nil {
		result.WindowDuration = ptr.To(DefaultTimeBasedFairShareWindowDuration)
	}
	if result.WindowType == nil {
		result.WindowType = ptr.To(DefaultTimeBasedFairShareWindowType)
	}

	// tumblingWindowStartTime has no default – leave as-is (can be nil)
	return result
}

func setSchedulingConfigurationFieldsOnCreate(
	ctx context.Context,
	nodePool *models.Nodepool,
	fields api.NodepoolCreateFields,
	capChecker clients_local.ClusterCapabilityChecker,
) error {
	return setSchedulingConfigurationFieldsOnCreateUpdateInternal(
		ctx,
		nodePool,
		fields.SchedulingConfiguration,
		capChecker)
}

func setSchedulingConfigurationFieldsOnUpdate(
	ctx context.Context,
	dbNodepool *models.Nodepool,
	fields api.NodepoolUpdateFields,
	capChecker clients_local.ClusterCapabilityChecker,
) error {
	return setSchedulingConfigurationFieldsOnCreateUpdateInternal(
		ctx,
		dbNodepool,
		fields.SchedulingConfiguration,
		capChecker)
}

func setSchedulingConfigurationFieldsOnPatch(ctx context.Context, dbNodepool *models.Nodepool, fields api.NodepoolUpdateFields, capChecker clients_local.ClusterCapabilityChecker) error {
	// the difference in Patch is that we only set the fields that are not nil

	schedulerConf := fields.SchedulingConfiguration
	if schedulerConf == nil {
		return nil
	}

	supportsGpuDevice, err := capChecker.ValidateClusterCapability(ctx, clients_local.LocalMinVersionOfGpuDevicePlacementStrategy)
	if err != nil {
		return fmt.Errorf("failed to check cluster support for gpuDevice placement strategy: %w", err)
	}

	// 1. Apply the deprecated combined gpu first: it mirrors to both node and device.
	if schedulerConf.PlacementStrategy.Gpu != nil && schedulerConf.PlacementStrategy.Gpu.IsValid() {
		dbNodepool.GpuPlacementStrategy = *schedulerConf.PlacementStrategy.Gpu
		dbNodepool.GpuDevicePlacementStrategy = *schedulerConf.PlacementStrategy.Gpu
	}
	if schedulerConf.PlacementStrategy.Cpu != nil && schedulerConf.PlacementStrategy.Cpu.IsValid() {
		dbNodepool.CpuPlacementStrategy = *schedulerConf.PlacementStrategy.Cpu
	}

	// 2. New fields override (new wins over old).
	if schedulerConf.GpuPodsPlacementStrategy != nil {
		if schedulerConf.GpuPodsPlacementStrategy.Node != nil && schedulerConf.GpuPodsPlacementStrategy.Node.IsValid() {
			dbNodepool.GpuPlacementStrategy = *schedulerConf.GpuPodsPlacementStrategy.Node
			if !supportsGpuDevice {
				dbNodepool.GpuDevicePlacementStrategy = *schedulerConf.GpuPodsPlacementStrategy.Node
			}
		}
		if schedulerConf.GpuPodsPlacementStrategy.Device != nil && schedulerConf.GpuPodsPlacementStrategy.Device.IsValid() {
			dbNodepool.GpuDevicePlacementStrategy = *schedulerConf.GpuPodsPlacementStrategy.Device
		}
	}
	if schedulerConf.CpuPodsPlacementStrategy != nil && schedulerConf.CpuPodsPlacementStrategy.Node != nil && schedulerConf.CpuPodsPlacementStrategy.Node.IsValid() {
		dbNodepool.CpuPlacementStrategy = *schedulerConf.CpuPodsPlacementStrategy.Node
	}

	if schedulerConf.MinGuaranteedRuntime != nil {
		dbNodepool.MinGuaranteedRuntime = schedulerConf.MinGuaranteedRuntime
	}
	if schedulerConf.TimeBasedFairShare != nil {
		dbNodepool.TimeBasedFairShare = setTimeBasedFairSharePatch(
			dbNodepool.TimeBasedFairShare,
			schedulerConf.TimeBasedFairShare,
		)
	}
	if schedulerConf.NumaAwareScheduling != nil {
		dbNodepool.NumaAwareScheduling = schedulerConf.NumaAwareScheduling
	}
	return nil
}

func setTimeBasedFairSharePatch(
	current models.TimeBasedFairShare,
	patch *api.TimeBasedFairShare,
) models.TimeBasedFairShare {
	if patch == nil {
		return current
	}

	// If disabling via patch, collapse to Enabled=false with no extra fields.
	if patch.Enabled != nil && !*patch.Enabled {
		return models.TimeBasedFairShare{
			Enabled: ptr.To(false),
		}
	}

	// From here on, Enabled=true.
	var result models.TimeBasedFairShare

	if current.Enabled != nil && *current.Enabled {
		// Start from existing config and override only specified fields.
		result = current
	} else {
		// Enabling for the first time or from disabled state – start from defaults.
		result = setTimeBasedFairShareDefaults(&api.TimeBasedFairShare{Enabled: ptr.To(true)})
	}

	if patch.HistoricalUsageWeight != nil {
		result.HistoricalUsageWeight = patch.HistoricalUsageWeight
	}
	if patch.DecayHalfLife != nil {
		result.DecayHalfLife = patch.DecayHalfLife
	}
	if patch.WindowDuration != nil {
		result.WindowDuration = patch.WindowDuration
	}
	if patch.WindowType != nil {
		result.WindowType = patch.WindowType
	}
	if patch.TumblingWindowStartTime != nil {
		result.TumblingWindowStartTime = patch.TumblingWindowStartTime
	}

	result.Enabled = ptr.To(true)
	return result
}

func setSchedulingCapabilitiesOnCreateUpdate(ctx context.Context,
	nodePool *models.Nodepool,
	fields api.SchedulingCapabilitiesNullable,
	isCreate bool,
) {
	if nodePool == nil {
		return
	}

	// create does not take capabilities from the request yet, so every capability gets its default
	if isCreate || !isEditableNodepoolCapabilitiesOn() {
		fields = api.SchedulingCapabilitiesNullable{}
	}
	nodePool.AllowAdjustableWeights = ptr.Deref(fields.AdjustableWeights, isAdjustableWeightsFFOn(ctx))
	nodePool.AllowQuota = ptr.Deref(fields.Quota, models.DefaultAllowQuota)
	nodePool.AllowRanks = ptr.Deref(fields.Ranks, models.DefaultAllowRanks)
	nodePool.AllowCpuQuota = ptr.Deref(fields.CpuQuota, isCpuQuotaFFOn(ctx))

	// an explicit cpuQuota with quota off is left for validateSchedulingMode to reject
	if !nodePool.AllowQuota && fields.CpuQuota == nil {
		nodePool.AllowCpuQuota = false
	}
}

func setSchedulingCapabilitiesOnPatch(dbNodepool *models.Nodepool, fields api.SchedulingCapabilitiesNullable) {
	if dbNodepool == nil || !isEditableNodepoolCapabilitiesOn() {
		return
	}

	// the difference in Patch is that we only set the fields that are not nil
	if fields.AdjustableWeights != nil {
		dbNodepool.AllowAdjustableWeights = *fields.AdjustableWeights
	}
	if fields.Quota != nil {
		dbNodepool.AllowQuota = *fields.Quota
	}
	if fields.Ranks != nil {
		dbNodepool.AllowRanks = *fields.Ranks
	}
	if fields.CpuQuota != nil {
		dbNodepool.AllowCpuQuota = *fields.CpuQuota
	}
}

func setStatusSchedulingConfigurationFromApiToDao(dbNodepool *models.Nodepool, apiStatus *api.NodepoolSyncFieldsStatus) {
	if apiStatus == nil {
		return
	}

	if apiStatus.SchedulingConfiguration != nil {
		cpu, gpuNode, gpuDevice := resolveRequestedPlacement(apiStatus.SchedulingConfiguration)
		dbNodepool.NodepoolStatus.CpuPlacementStrategy = cpu
		dbNodepool.NodepoolStatus.GpuPlacementStrategy = gpuNode
		dbNodepool.NodepoolStatus.GpuDevicePlacementStrategy = gpuDevice
		dbNodepool.NodepoolStatus.MinGuaranteedRuntime = apiStatus.SchedulingConfiguration.MinGuaranteedRuntime
	} else {
		dbNodepool.NodepoolStatus.CpuPlacementStrategy = convutils.Deref(apiStatus.PlacementStrategy.Cpu, "")
		dbNodepool.NodepoolStatus.GpuPlacementStrategy = convutils.Deref(apiStatus.PlacementStrategy.Gpu, "")
		dbNodepool.MinGuaranteedRuntime = nil
	}
}

// networkTopologyIdEqual distinguishes "no topology" from a topology whose id happens to
// be the zero UUID, which a plain Deref against uuid.Nil would conflate.
func networkTopologyIdEqual(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func clusterFieldsUpdated(left, right models.Nodepool) bool {
	return convutils.Deref(left.LabelKey, "") != convutils.Deref(right.LabelKey, "") ||
		convutils.Deref(left.LabelValue, "") != convutils.Deref(right.LabelValue, "") ||
		left.GpuPlacementStrategy != right.GpuPlacementStrategy ||
		left.GpuDevicePlacementStrategy != right.GpuDevicePlacementStrategy ||
		left.CpuPlacementStrategy != right.CpuPlacementStrategy ||
		left.MinGuaranteedRuntime != right.MinGuaranteedRuntime ||
		!timeBasedFairShareEqual(left.TimeBasedFairShare, right.TimeBasedFairShare) ||
		convutils.Deref(left.GPUNetworkAccelerationDetection, "") != convutils.Deref(right.GPUNetworkAccelerationDetection, "") ||
		left.GPUNetworkAccelerationLabelKey != right.GPUNetworkAccelerationLabelKey ||
		!networkTopologyIdEqual(left.NetworkTopologyId, right.NetworkTopologyId) ||
		convutils.Deref(left.NumaAwareScheduling, false) != convutils.Deref(right.NumaAwareScheduling, false)
}

func timeBasedFairShareEqual(left, right models.TimeBasedFairShare) bool {
	return convutils.Deref(left.Enabled, false) == convutils.Deref(right.Enabled, false) &&
		convutils.Deref(left.HistoricalUsageWeight, DefaultTimeBasedFairShareHistoricalUsageWeight) == convutils.Deref(right.HistoricalUsageWeight, DefaultTimeBasedFairShareHistoricalUsageWeight) &&
		convutils.Deref(left.DecayHalfLife, DefaultTimeBasedFairShareDecayHalfLife) == convutils.Deref(right.DecayHalfLife, DefaultTimeBasedFairShareDecayHalfLife) &&
		convutils.Deref(left.WindowDuration, DefaultTimeBasedFairShareWindowDuration) == convutils.Deref(right.WindowDuration, DefaultTimeBasedFairShareWindowDuration) &&
		convutils.Deref(left.WindowType, DefaultTimeBasedFairShareWindowType) == convutils.Deref(right.WindowType, DefaultTimeBasedFairShareWindowType) &&
		convutils.Deref(left.TumblingWindowStartTime, time.Time{}).Equal(convutils.Deref(right.TumblingWindowStartTime, time.Time{}))
}
