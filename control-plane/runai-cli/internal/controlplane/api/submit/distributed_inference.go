/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/spf13/pflag"
)

// DistributedInference holds configuration for distributed inference workloads
type DistributedInference struct {
	ServingPort       string
	Replicas          int32
	Workers           openapi.NullableInt32
	WorkerAsLeader    bool
	StartupPolicy     openapi.NullableDistributedInferenceStartupPolicy
	Mnnvl             string
	WorkerAnnotations []string
	WorkerLabels      []string
	// Worker scheduling fields
	WorkerNodeType             string
	WorkerExcludeNodes         []string
	WorkerTolerations          []string
	WorkerRequiredPodTopology  string
	WorkerPreferredPodTopology string
}

func DistributedInferenceFlagSet(flagsDistributed *pflag.FlagSet, specDistributed *DistributedInference) {
	// Distributed Inference specific flags
	flagsDistributed.Int32Var(&specDistributed.Replicas, consts.AttributeNameReplicas, 1, consts.AttributeUsageReplicas)
	flagsDistributed.Var(cli.NewNullableNonNegativeInt32Flag(&specDistributed.Workers), consts.AttributeNameWorkers, consts.AttributeUsageWorkers)
	flagsDistributed.BoolVar(&specDistributed.WorkerAsLeader, consts.AttributeNameWorkerAsLeader, true, consts.AttributeUsageWorkerAsLeader)
	flagsDistributed.Var(cli.NewDistributedInferenceStartupPolicyFlag(&specDistributed.StartupPolicy), consts.AttributeNameStartupPolicy, consts.AttributeUsageStartupPolicy)
	flagsDistributed.Var(cli.NewMnnvlModeFlag(&specDistributed.Mnnvl), consts.AttributeNameMnnvl, consts.AttributeUsageMnnvl)
	flagsDistributed.StringArrayVar(&specDistributed.WorkerAnnotations, consts.AttributeNameWorkerAnnotation, []string{}, consts.AttributeUsageWorkerAnnotation)
	flagsDistributed.StringArrayVar(&specDistributed.WorkerLabels, consts.AttributeNameWorkerLabel, []string{}, consts.AttributeUsageWorkerLabel)

	// ServingPort flag (using distributed inference specific usage message)
	flagsDistributed.StringVar(&specDistributed.ServingPort, consts.AttributeNameServingPort, "", consts.AttributeUsageServingPortDistributed)
}

// DistributedInferenceComputeFlagSet adds worker-specific compute flags for distributed inference
func DistributedInferenceComputeFlagSet(flagsCompute *pflag.FlagSet, compute *Compute) {
	flagsCompute.Var(cli.NewNullableNonNegativeInt32Flag(&compute.WorkerGpuDevicesRequest), consts.AttributeNameWorkerGpuDevicesRequest, consts.AttributeUsageWorkerGpuDevicesRequest)
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&compute.WorkerGpuPortionRequest), consts.AttributeNameWorkerGpuPortionRequest, consts.AttributeUsageWorkerGpuPortionRequest)
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&compute.WorkerGpuPortionLimit), consts.AttributeNameWorkerGpuPortionLimit, consts.AttributeUsageWorkerGpuPortionLimit)
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&compute.WorkerGpuMemoryPortionRequest), consts.AttributeNameWorkerGpuMemoryPortionRequest, consts.AttributeUsageWorkerGpuMemoryPortionRequest)
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&compute.WorkerGpuMemoryPortionLimit), consts.AttributeNameWorkerGpuMemoryPortionLimit, consts.AttributeUsageWorkerGpuMemoryPortionLimit)
	flagsCompute.StringVar(&compute.WorkerGpuComputeMode, consts.AttributeNameWorkerGpuComputeMode, "", consts.AttributeUsageGpuComputeMode)
	if err := flagsCompute.MarkDeprecated(consts.AttributeNameWorkerGpuPortionRequest,
		"use --"+consts.AttributeNameWorkerGpuMemoryPortionRequest+" for GPU memory fractions, or --"+
			consts.AttributeNameWorkerGpuDevicesRequest+" for whole-GPU counts"); err != nil {
		panic(err)
	}
	if err := flagsCompute.MarkDeprecated(consts.AttributeNameWorkerGpuPortionLimit, "use --"+consts.AttributeNameWorkerGpuMemoryPortionLimit+" instead"); err != nil {
		panic(err)
	}
	flagsCompute.StringVar(&compute.WorkerGpuMemoryRequest, consts.AttributeNameWorkerGpuMemoryRequest, "", consts.AttributeUsageWorkerGpuMemoryRequest)
	flagsCompute.StringVar(&compute.WorkerGpuMemoryLimit, consts.AttributeNameWorkerGpuMemoryLimit, "", consts.AttributeUsageWorkerGpuMemoryLimit)
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&compute.WorkerCpuCoreRequest), consts.AttributeNameWorkerCpuCoreRequest, consts.AttributeUsageWorkerCpuCoreRequest)
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&compute.WorkerCpuCoreLimit), consts.AttributeNameWorkerCpuCoreLimit, consts.AttributeUsageWorkerCpuCoreLimit)
	flagsCompute.StringVar(&compute.WorkerCpuMemoryRequest, consts.AttributeNameWorkerCpuMemoryRequest, "", consts.AttributeUsageWorkerCpuMemoryRequest)
	flagsCompute.StringVar(&compute.WorkerCpuMemoryLimit, consts.AttributeNameWorkerCpuMemoryLimit, "", consts.AttributeUsageWorkerCpuMemoryLimit)
	flagsCompute.StringArrayVar(&compute.WorkerExtendedResources, consts.AttributeNameWorkerExtendedResource, []string{}, consts.AttributeUsageWorkerExtendedResource)
}

// DistributedInferenceStorageFlagSet adds worker-specific storage flags for distributed inference.
// Only storage types supported by the DistributedInferenceLeaderWorkerSpec schema are exposed:
// configMapVolume, emptyDirVolume, pvc, secretVolume.
func DistributedInferenceStorageFlagSet(flagsStorage *pflag.FlagSet, storage *Storage) {
	flagsStorage.StringArrayVar(&storage.WorkerExistingPvc, consts.AttributeNameWorkerExistingPvc, []string{}, consts.AttributeUsageWorkerExistingPvc)
	flagsStorage.StringArrayVar(&storage.WorkerNewPvc, consts.AttributeNameWorkerNewPvc, []string{}, consts.AttributeUsageWorkerNewPvc)
	flagsStorage.StringArrayVar(&storage.WorkerSecretVolume, consts.AttributeNameWorkerSecretVolume, []string{}, consts.AttributeUsageWorkerSecretVolume)
	flagsStorage.StringArrayVar(&storage.WorkerConfigmapMapVolume, consts.AttributeNameWorkerConfigmapMapVolume, []string{}, consts.AttributeUsageWorkerConfigmapMapVolume)
	flagsStorage.StringArrayVar(&storage.WorkerEmptyDirVolume, consts.AttributeNameWorkerEmptyDirVolume, []string{}, consts.AttributeUsageWorkerEmptyDirVolume)
}

// DistributedInferenceSecurityFlagSet adds worker-specific security flags for distributed inference
func DistributedInferenceSecurityFlagSet(flagsSecurity *pflag.FlagSet, security *Security) {
	flagsSecurity.IntVar(&security.WorkerRunAsUid, consts.AttributeNameWorkerRunAsUid, 0, consts.AttributeUsageWorkerRunAsUid)
	flagsSecurity.IntVar(&security.WorkerRunAsGid, consts.AttributeNameWorkerRunAsGid, 0, consts.AttributeUsageWorkerRunAsGid)
	flagsSecurity.BoolVar(&security.WorkerRunAsNonRoot, consts.AttributeNameWorkerRunAsNonRoot, false, consts.AttributeUsageWorkerRunAsNonRoot)
	flagsSecurity.IntSliceVar(&security.WorkerSupplementalGroups, consts.AttributeNameWorkerSupplementalGroups, nil, consts.AttributeUsageWorkerSupplementalGroups)
	flagsSecurity.StringArrayVar(&security.WorkerCapabilities, consts.AttributeNameWorkerCapability, []string{}, consts.AttributeUsageWorkerCapability)
	flagsSecurity.BoolVar(&security.WorkerPrivileged, consts.AttributeNameWorkerPrivileged, false, consts.AttributeUsageWorkerPrivileged)
	flagsSecurity.BoolVar(&security.WorkerAllowPrivilegeEscalation, consts.AttributeNameWorkerAllowPrivilegeEscalation, false, consts.AttributeUsageWorkerAllowPrivilegeEscalation)
	flagsSecurity.BoolVar(&security.WorkerHostIpc, consts.AttributeNameWorkerHostIpc, false, consts.AttributeUsageWorkerHostIpc)
	flagsSecurity.BoolVar(&security.WorkerHostNetwork, consts.AttributeNameWorkerHostNetwork, false, consts.AttributeUsageWorkerHostNetwork)
}

// DistributedInferenceContainerFlagSet adds worker-specific container flags for distributed inference
func DistributedInferenceContainerFlagSet(flagsContainer *pflag.FlagSet, container *Container) {
	flagsContainer.StringVar(&container.WorkerImage, consts.AttributeNameWorkerImage, "", consts.AttributeUsageWorkerImage)
	flagsContainer.StringVar(&container.WorkerImagePullPolicy, consts.AttributeNameWorkerImagePullPolicy, "", consts.AttributeUsageWorkerImagePullPolicy)
	flagsContainer.StringVar(&container.WorkerCommand, consts.AttributeNameWorkerCommand, "", consts.AttributeUsageWorkerCommand)
	flagsContainer.StringVar(&container.WorkerArgs, consts.AttributeNameWorkerArgs, "", consts.AttributeUsageWorkerArgs)
	flagsContainer.StringVar(&container.WorkerWorkingDir, consts.AttributeNameWorkerWorkingDir, "", consts.AttributeUsageWorkerWorkingDir)
	flagsContainer.StringArrayVar(&container.WorkerEnvironmentVariables, consts.AttributeNameWorkerEnvironment, []string{}, consts.AttributeUsageWorkerEnvironment)
	flagsContainer.StringArrayVar(&container.WorkerFieldRef, consts.AttributeNameWorkerEnvPodFieldRef, []string{}, consts.AttributeUsageWorkerEnvPodFieldRef)
	flagsContainer.StringArrayVar(&container.WorkerEnvSecret, consts.AttributeNameWorkerEnvSecret, []string{}, consts.AttributeUsageWorkerEnvSecret)
	flagsContainer.BoolVar(&container.WorkerCreateHomeDir, consts.AttributeNameWorkerCreateHomeDir, false, consts.AttributeUsageWorkerCreateHomeDir)
}

// DistributedInferenceSchedulingFlagSet adds worker-specific scheduling flags for distributed inference
func DistributedInferenceSchedulingFlagSet(flagsScheduling *pflag.FlagSet, distributed *DistributedInference) {
	flagsScheduling.StringVar(&distributed.WorkerNodeType, consts.AttributeNameWorkerNodeType, "", consts.AttributeUsageWorkerNodeType)
	flagsScheduling.StringArrayVar(&distributed.WorkerExcludeNodes, consts.AttributeNameWorkerExcludeNode, []string{}, consts.AttributeUsageWorkerExcludeNode)
	flagsScheduling.StringArrayVar(&distributed.WorkerTolerations, consts.AttributeNameWorkerToleration, []string{}, consts.AttributeUsageWorkerToleration)
	flagsScheduling.StringVar(&distributed.WorkerRequiredPodTopology, consts.AttributeNameWorkerRequiredPodTopologyKey, "", consts.AttributeUsageWorkerRequiredPodTopologyKey)
	flagsScheduling.StringVar(&distributed.WorkerPreferredPodTopology, consts.AttributeNameWorkerPreferredPodTopologyKey, "", consts.AttributeUsageWorkerPreferredPodTopologyKey)
}

// ParseDistributedInferenceServingPort parses serving port for distributed inference
func ParseDistributedInferenceServingPort(input string) (*openapi.DistributedInferenceServingPort, error) {
	pairs, err := parseServingPortInput(input)
	if err != nil {
		return nil, err
	}

	sp := openapi.NewDistributedInferenceServingPort()
	for _, kv := range pairs {
		if err := assignValueDistributedInferenceServingPort(sp, kv.Key, kv.Value); err != nil {
			return nil, err
		}
	}
	return sp, nil
}

func assignValueDistributedInferenceServingPort(sp *openapi.DistributedInferenceServingPort, key, value string) error {
	if value == "" && key != keyExposeExternally {
		return fmt.Errorf(errInvalidValueRequiresKey, key)
	}

	switch key {
	case keyContainer:
		port, err := validateContainerPort(value)
		if err != nil {
			return err
		}
		sp.Port = *openapi.NewNullableInt32(openapi.PtrInt32(int32(port)))
	case keyAuthorizationType:
		enumVal, err := openapi.NewDistributedInferenceServingPortAccessAuthorizationTypeEnumFromValue(value)
		if err != nil {
			return err
		}
		sp.AuthorizationType = *openapi.NewNullableDistributedInferenceServingPortAccessAuthorizationTypeEnum(enumVal)
	case keyAuthorizedUsers:
		sp.AuthorizedUsers = strings.Split(value, separatorAuthorizedEntities)
	case keyAuthorizedGroups:
		sp.AuthorizedGroups = strings.Split(value, separatorAuthorizedEntities)
	case keyProtocol:
		var protocol openapi.NullableDistributedInferenceServingPortProtocol
		protocol.Set((*openapi.DistributedInferenceServingPortProtocol)(&value))
		sp.Protocol = protocol
	case keyExposeExternally:
		var exposeExternally bool
		switch value {
		case "", "true":
			exposeExternally = true
		case "false":
			exposeExternally = false
		default:
			return fmt.Errorf(errInvalidBooleanValue, keyExposeExternally, value)
		}
		sp.ExposeExternally.Set(openapi.PtrBool(exposeExternally))
	case keyExposedUrl:
		sp.ExposedUrl.Set(&value)
	default:
		return fmt.Errorf(errInvalidServingPortKey, key)
	}
	return nil
}
