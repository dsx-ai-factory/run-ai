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
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/submit"
	nodes "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/node"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"k8s.io/utils/ptr"
)

const (
	RemoveSuffix = "=-"

	FrameworkNoSupportMaster                 = "distributed framework type (%s) does not support no master"
	FrameworkNoSupportMinMaxRep              = "distributed framework type (%s) does not support no min/max replica"
	FrameworkNoSupportSlotsPerWorker         = "distributed framework type (%s) does not support slots per worker"
	FrameworkNoSupportSshAuthMountPath       = "distributed framework type (%s) does not support ssh auth mount path"
	FrameworkNoSupportLauncherCreationPolicy = "distributed framework type (%s) does not support launcher creation policy path"
	FrameworkNoSupportMaxEqMin               = "distributed max replica value (%d) must be equal or greater then min replica value (%d)"
	FrameworkNoSupportSegmentSize            = "distributed framework type (%s) does not support segment size"
	SegmentSizeMustBePositive                = "--segment-size must be greater than 0"
)

// TODO remove after removing BC submit-dist
func (gs *GeneralSpec) CheckDistributeFlags() error {
	dist := &gs.flags.DistributedTraining
	framework, err := dist.GenerateToDistributedFramework()
	if err != nil {
		return err
	}
	if dist.NoMaster &&
		(*framework != openapi.DISTRIBUTEDFRAMEWORK_TF &&
			*framework != openapi.DISTRIBUTEDFRAMEWORK_PY_TORCH &&
			*framework != openapi.DISTRIBUTEDFRAMEWORK_JAX) {
		return fmt.Errorf(FrameworkNoSupportMaster, string(*framework))
	}
	if dist.MaxReplica > 0 || dist.MinReplica > 0 {
		if *framework != openapi.DISTRIBUTEDFRAMEWORK_PY_TORCH {
			return fmt.Errorf(FrameworkNoSupportMinMaxRep, string(*framework))
		}
		if dist.MaxReplica > 0 && dist.MinReplica > dist.MaxReplica {
			return fmt.Errorf(FrameworkNoSupportMaxEqMin, dist.MaxReplica, dist.MinReplica)
		}
	}
	if dist.SegmentSize.IsSet() {
		segmentSize := dist.SegmentSize.Get()
		if *framework != openapi.DISTRIBUTEDFRAMEWORK_PY_TORCH {
			return fmt.Errorf(FrameworkNoSupportSegmentSize, string(*framework))
		}
		if segmentSize == nil || *segmentSize < 1 {
			return errors.New(SegmentSizeMustBePositive)
		}
	}
	if dist.SlotsPerWorker > 0 && *framework != openapi.DISTRIBUTEDFRAMEWORK_MPI {
		return fmt.Errorf(FrameworkNoSupportSlotsPerWorker, string(*framework))
	}
	if dist.SshAuthMountPath != "" && *framework != openapi.DISTRIBUTEDFRAMEWORK_MPI {
		return fmt.Errorf(FrameworkNoSupportSshAuthMountPath, string(*framework))
	}
	if dist.LauncherCreationPolicy.IsSet() && *framework != openapi.DISTRIBUTEDFRAMEWORK_MPI {
		return fmt.Errorf(FrameworkNoSupportLauncherCreationPolicy, string(*framework))
	}
	return nil
}

func (gs *GeneralSpec) CheckDistributeMpiFlags() error {
	dist := &gs.flags.DistributedTraining
	framework, err := dist.GenerateToDistributedFramework()
	if err != nil {
		return err
	}
	if *framework != openapi.DISTRIBUTEDFRAMEWORK_MPI {
		return nil
	}

	if gs.flags.Compute.MasterGpuDeviceRequest.IsSet() && *gs.flags.Compute.MasterGpuDeviceRequest.Get() > 0 ||
		gs.flags.Compute.MasterGpuPortionLimit.IsSet() && *gs.flags.Compute.MasterGpuPortionLimit.Get() > 0 ||
		gs.flags.Compute.MasterGpuPortionRequest.IsSet() && *gs.flags.Compute.MasterGpuPortionRequest.Get() > 0 ||
		gs.flags.Compute.MasterGpuMemoryPortionRequest.IsSet() && *gs.flags.Compute.MasterGpuMemoryPortionRequest.Get() > 0 ||
		gs.flags.Compute.MasterGpuMemoryPortionLimit.IsSet() && *gs.flags.Compute.MasterGpuMemoryPortionLimit.Get() > 0 ||
		gs.flags.Compute.MasterGpuComputeMode != "" {
		return errors.New(consts.ErrorSubmitMpiWithGpu) //nolint:staticcheck // user facing message
	}

	return nil
}

func GenerateMasterSpec(gs *GeneralSpec, ctx context.Context, clusterId string, nodesService *nodes.BackendNodeService) (*openapi.MasterSpec, error) {
	// Master overrides must not affect subsequent worker spec generation.
	originalFlags := gs.flags
	defer func() { gs.flags = originalFlags }()

	err := updateWorkerSpec(gs)
	if err != nil {
		return nil, err
	}
	err = gs.generateSpec(ctx, clusterId, nodesService)
	if err != nil {
		return nil, err
	}
	return generateMasterSpec(&gs.fullSpec)
}

func GenerateWorkerSpec(gs *GeneralSpec, ctx context.Context, clusterId string, nodesService *nodes.BackendNodeService) (*openapi.DistributedSpecSpec, error) {
	// Preserve original flags so worker-specific overrides don't leak back
	originalFlags := gs.flags
	restoreFlags := func() {
		gs.flags = originalFlags
	}

	updateWorkerFlags(gs, gs.flags.DistributedInference.WorkerAsLeader)

	err := gs.generateSpec(ctx, clusterId, nodesService)
	if err != nil {
		restoreFlags()
		return nil, err
	}

	workerSpec := gs.fullSpec.DistributedSpecSpec
	restoreFlags()

	return &workerSpec, nil
}

func updateWorkerSpec(gs *GeneralSpec) error {
	flags := &gs.flags
	err := updateEnvironment(&flags.Container, &gs.fullSpec)
	updateFlags(gs)
	return err
}

func updateEnvironment(container *submit.Container, specEnv *GeneralFullSpec) error {
	if len(container.MasterEnvironmentVariables) == 0 {
		return nil
	}

	var (
		masterEnvList []string
		masterEnvMap  = map[string]string{}
		err           error
	)
	for _, env := range specEnv.GetEnvironmentVariables() {
		masterEnvMap[env.GetName()] = env.GetValue()
	}

	err = preventEnvironments(&container.MasterEnvironmentVariables, &masterEnvMap)
	if err != nil {
		return err
	}
	err = addEnvironments(&container.MasterEnvironmentVariables, &masterEnvMap)
	if err != nil {
		return err
	}
	for key, value := range masterEnvMap {
		masterEnvList = append(masterEnvList, fmt.Sprintf("%s=%s", key, value))
	}
	container.EnvironmentVariables = masterEnvList
	return nil
}

func preventEnvironments(masterEnv *[]string, workerEnv *map[string]string) error {
	for _, env := range *masterEnv {
		if strings.HasSuffix(env, RemoveSuffix) {
			key, _ := strings.CutSuffix(env, RemoveSuffix)
			if _, found := (*workerEnv)[key]; found {
				delete(*workerEnv, key)
			} else {
				return fmt.Errorf("couldn't find environment to remove, env: %s", key)
			}
		}
	}
	return nil
}

func addEnvironments(masterEnv *[]string, workerEnv *map[string]string) error {
	for _, env := range *masterEnv {
		if !strings.HasSuffix(env, RemoveSuffix) {
			envSplit, err := specset.SplitEqualSepValidateTwoSection(env)
			if err != nil {
				return err
			}
			(*workerEnv)[envSplit[0]] = envSplit[1]
		}
	}
	return nil
}

// mergeEnvVars merges key=value slices. Worker values override leader values with the same key.
func mergeEnvVars(leaderVars, workerVars []string) []string {
	envMap := make(map[string]string)
	for _, env := range leaderVars {
		if name := getEnvVarName(env); name != "" {
			envMap[name] = env
		}
	}
	for _, env := range workerVars {
		if name := getEnvVarName(env); name != "" {
			envMap[name] = env
		}
	}
	result := make([]string, 0, len(envMap))
	for _, env := range envMap {
		result = append(result, env)
	}
	return result
}

func getEnvVarName(env string) string {
	parts, err := specset.SplitEqualSepValidateTwoSection(env)
	if err != nil {
		return ""
	}
	return parts[0]
}

// mergeOrReplace merges when workerAsLeader is true, replaces when false.
func mergeOrReplace(leaderVars, workerVars []string, workerAsLeader bool) []string {
	if workerAsLeader {
		return mergeEnvVars(leaderVars, workerVars)
	}
	return workerVars
}

// overrideOrClear returns workerValue if provided, otherwise returns leaderValue if workerAsLeader is true, or empty string if false.
func overrideOrClear(leaderValue, workerValue string, workerAsLeader bool) string {
	if workerValue != "" {
		return workerValue
	}
	if workerAsLeader {
		return leaderValue
	}
	return ""
}

func updateFlags(gs *GeneralSpec) {
	container := &gs.flags.Container
	compute := &gs.flags.Compute
	storage := &gs.flags.Storage
	lifeCycle := &gs.flags.Lifecycle

	if len(container.MasterCommand) > 0 || len(container.MasterArgs) > 0 {
		container.ExecuteCommand = container.MasterCommand
		container.ExecuteArgs = container.MasterArgs
	}

	if storage.MasterNoPVC {
		storage.NewPvc = []string{}
		storage.ExistingPvc = []string{}
	}

	if lifeCycle.MasterRestartPolicy.IsSet() {
		lifeCycle.RestartPolicy = lifeCycle.MasterRestartPolicy
	}

	// MPI can't and should not have master GPU.
	if gs.framework == api.MPI {
		compute.GpuDeviceRequest.Unset()
		compute.GpuPortionRequest.Set(ptr.To(float64(0)))
		compute.GpuMemoryPortionRequest.Unset()
		compute.GpuMemoryPortionLimit.Unset()
		compute.GpuComputeMode = ""
	} else {
		if convutils.HasValue(compute.MasterGpuDeviceRequest) {
			compute.GpuDeviceRequest.Set(compute.MasterGpuDeviceRequest.Get())
			compute.GpuPortionRequest.Set(ptr.To(float64(0)))
			compute.GpuMemoryPortionRequest.Unset()
			if *compute.MasterGpuDeviceRequest.Get() == 0 {
				// A CPU-only master must not inherit the workers' fractional GPU settings.
				compute.GpuRequestType = ""
				compute.GpuPortionLimit.Unset()
				compute.GpuMemoryPortionLimit.Unset()
				compute.GpuComputeMode = ""
				compute.GpuMemoryRequest = ""
				compute.GpuMemoryLimit = ""
			}
		}
		if compute.MasterGpuPortionRequest.IsSet() && *compute.MasterGpuPortionRequest.Get() > 0 || compute.MasterGpuMemoryPortionRequest.IsSet() {
			compute.GpuPortionRequest = compute.MasterGpuPortionRequest
			compute.GpuMemoryPortionRequest = compute.MasterGpuMemoryPortionRequest
			// Keep explicit device counts for fractions, without inheriting worker counts or changing legacy whole-GPU precedence.
			hasMasterDevices := convutils.HasValue(compute.MasterGpuDeviceRequest) && *compute.MasterGpuDeviceRequest.Get() > 0
			legacyWholeGPUs := convutils.HasValue(compute.MasterGpuPortionRequest) && *compute.MasterGpuPortionRequest.Get() > 1
			if !hasMasterDevices || legacyWholeGPUs {
				compute.GpuDeviceRequest.Unset()
			}
		}
		if compute.MasterGpuPortionLimit.IsSet() && *compute.MasterGpuPortionLimit.Get() > 0 || compute.MasterGpuMemoryPortionLimit.IsSet() {
			compute.GpuPortionLimit = compute.MasterGpuPortionLimit
			compute.GpuMemoryPortionLimit = compute.MasterGpuMemoryPortionLimit
		}
		if compute.MasterGpuComputeMode != "" {
			compute.GpuComputeMode = compute.MasterGpuComputeMode
		}
	}
	if compute.MasterExtendedResources != nil {
		compute.ExtendedResources = compute.MasterExtendedResources
	}
	if compute.MasterCpuCoreRequest.IsSet() && *compute.MasterCpuCoreRequest.Get() > 0 {
		compute.CpuCoreRequest.Set(compute.MasterCpuCoreRequest.Get())
	}
	if compute.MasterCpuCoreLimit.IsSet() && *compute.MasterCpuCoreLimit.Get() > 0 {
		compute.CpuCoreLimit.Set(compute.MasterCpuCoreLimit.Get())
	}
	if compute.MasterCpuMemoryRequest != "" {
		compute.CpuMemoryRequest = compute.MasterCpuMemoryRequest
	}
	if compute.MasterCpuMemoryLimit != "" {
		compute.CpuMemoryLimit = compute.MasterCpuMemoryLimit
	}
}

func updateWorkerFlags(gs *GeneralSpec, workerAsLeader bool) {
	container := &gs.flags.Container
	compute := &gs.flags.Compute
	storage := &gs.flags.Storage
	security := &gs.flags.Security
	distributed := &gs.flags.DistributedInference
	general := &gs.flags.General

	if len(distributed.WorkerAnnotations) > 0 {
		general.Annotations = mergeOrReplace(general.Annotations, distributed.WorkerAnnotations, workerAsLeader)
	}
	if len(distributed.WorkerLabels) > 0 {
		general.Labels = mergeOrReplace(general.Labels, distributed.WorkerLabels, workerAsLeader)
	}
	if container.WorkerImage != "" {
		container.Image = container.WorkerImage
	}
	if container.WorkerImagePullPolicy != "" {
		container.ImagePullPolicy = container.WorkerImagePullPolicy
	}
	container.ExecuteCommand = overrideOrClear(container.ExecuteCommand, container.WorkerCommand, workerAsLeader)
	container.ExecuteArgs = overrideOrClear(container.ExecuteArgs, container.WorkerArgs, workerAsLeader)
	if container.WorkerWorkingDir != "" {
		container.WorkingDir = container.WorkerWorkingDir
	}
	if len(container.WorkerEnvironmentVariables) > 0 {
		container.EnvironmentVariables = mergeOrReplace(container.EnvironmentVariables, container.WorkerEnvironmentVariables, workerAsLeader)
	}
	if len(container.WorkerFieldRef) > 0 {
		container.FieldRef = mergeOrReplace(container.FieldRef, container.WorkerFieldRef, workerAsLeader)
	}
	if len(container.WorkerEnvSecret) > 0 {
		container.EnvSecret = mergeOrReplace(container.EnvSecret, container.WorkerEnvSecret, workerAsLeader)
	}
	if container.WorkerCreateHomeDir {
		container.CreateHomeDir = container.WorkerCreateHomeDir
	}

	if compute.WorkerGpuDevicesRequest.IsSet() && *compute.WorkerGpuDevicesRequest.Get() > 0 {
		compute.GpuDeviceRequest.Set(compute.WorkerGpuDevicesRequest.Get())
	}
	if compute.WorkerGpuPortionRequest.IsSet() && *compute.WorkerGpuPortionRequest.Get() > 0 || compute.WorkerGpuMemoryPortionRequest.IsSet() {
		compute.GpuPortionRequest = compute.WorkerGpuPortionRequest
		compute.GpuMemoryPortionRequest = compute.WorkerGpuMemoryPortionRequest
	}
	if compute.WorkerGpuPortionLimit.IsSet() && *compute.WorkerGpuPortionLimit.Get() > 0 || compute.WorkerGpuMemoryPortionLimit.IsSet() {
		compute.GpuPortionLimit = compute.WorkerGpuPortionLimit
		compute.GpuMemoryPortionLimit = compute.WorkerGpuMemoryPortionLimit
	}
	if compute.WorkerGpuComputeMode != "" {
		compute.GpuComputeMode = compute.WorkerGpuComputeMode
	}
	if compute.WorkerGpuMemoryRequest != "" {
		compute.GpuMemoryRequest = compute.WorkerGpuMemoryRequest
	}
	if compute.WorkerGpuMemoryLimit != "" {
		compute.GpuMemoryLimit = compute.WorkerGpuMemoryLimit
	}
	if compute.WorkerCpuCoreRequest.IsSet() && *compute.WorkerCpuCoreRequest.Get() > 0 {
		compute.CpuCoreRequest.Set(compute.WorkerCpuCoreRequest.Get())
	}
	if compute.WorkerCpuCoreLimit.IsSet() && *compute.WorkerCpuCoreLimit.Get() > 0 {
		compute.CpuCoreLimit.Set(compute.WorkerCpuCoreLimit.Get())
	}
	if compute.WorkerCpuMemoryRequest != "" {
		compute.CpuMemoryRequest = compute.WorkerCpuMemoryRequest
	}
	if compute.WorkerCpuMemoryLimit != "" {
		compute.CpuMemoryLimit = compute.WorkerCpuMemoryLimit
	}
	if len(compute.WorkerExtendedResources) > 0 {
		compute.ExtendedResources = compute.WorkerExtendedResources
	}

	if len(storage.WorkerExistingPvc) > 0 {
		storage.ExistingPvc = storage.WorkerExistingPvc
	}
	if len(storage.WorkerNewPvc) > 0 {
		storage.NewPvc = storage.WorkerNewPvc
	}
	if len(storage.WorkerSecretVolume) > 0 {
		storage.SecretVolume = storage.WorkerSecretVolume
	}
	if len(storage.WorkerConfigmapMapVolume) > 0 {
		storage.ConfigMapVolume = storage.WorkerConfigmapMapVolume
	}
	if len(storage.WorkerEmptyDirVolume) > 0 {
		storage.EmptyDirVolume = storage.WorkerEmptyDirVolume
	}

	if security.WorkerRunAsUid != 0 {
		security.RunAsUid = int64(security.WorkerRunAsUid)
	}
	if security.WorkerRunAsGid != 0 {
		security.RunAsGid = int64(security.WorkerRunAsGid)
	}
	if security.WorkerRunAsNonRoot {
		security.RunAsNonRoot = security.WorkerRunAsNonRoot
	}
	if len(security.WorkerSupplementalGroups) > 0 {
		security.SupplementalGroups = security.WorkerSupplementalGroups
	}
	if len(security.WorkerCapabilities) > 0 {
		security.Capabilities = security.WorkerCapabilities
	}
	if security.WorkerPrivileged {
		security.Privileged = security.WorkerPrivileged
	}
	if security.WorkerAllowPrivilegeEscalation {
		security.AllowPrivilegeEscalation = security.WorkerAllowPrivilegeEscalation
	}
	if security.WorkerHostIpc {
		security.HostIpc = security.WorkerHostIpc
	}
	if security.WorkerHostNetwork {
		security.HostNetwork = security.WorkerHostNetwork
	}

	// Worker scheduling updates
	scheduling := &gs.flags.Scheduling
	if distributed.WorkerNodeType != "" {
		scheduling.NodeType = distributed.WorkerNodeType
	}
	if len(distributed.WorkerExcludeNodes) > 0 {
		scheduling.ExcludeNodes = distributed.WorkerExcludeNodes
	}
	if len(distributed.WorkerTolerations) > 0 {
		scheduling.Toleration = distributed.WorkerTolerations
	}
	// Worker pod topology overrides
	if distributed.WorkerRequiredPodTopology != "" {
		container.PodAffinityRequired = distributed.WorkerRequiredPodTopology
		// Clear preferred if required is set
		container.PodAffinityPreferred = ""
	}
	if distributed.WorkerPreferredPodTopology != "" {
		container.PodAffinityPreferred = distributed.WorkerPreferredPodTopology
		// Clear required if preferred is set
		container.PodAffinityRequired = ""
	}
}

func generateMasterSpec(gs *GeneralFullSpec) (*openapi.MasterSpec, error) {
	masterSpec := openapi.MasterSpec{}
	bytes, err := json.Marshal(gs)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &masterSpec)
	if err != nil {
		return nil, err
	}
	return &masterSpec, nil
}

func uniqueMasterSpecRequired(gf *GeneralFlags) bool {
	if gf.DistributedTraining.DistributedFramework == api.MPI.String() &&
		(gf.Compute.GpuMemoryPortionRequest.IsSet() || gf.Compute.GpuMemoryPortionLimit.IsSet() || gf.Compute.GpuComputeMode != "" ||
			gf.Compute.GpuPortionRequest.IsSet() && *gf.Compute.GpuPortionRequest.Get() > 0 ||
			gf.Compute.GpuDeviceRequest.IsSet() && *gf.Compute.GpuDeviceRequest.Get() > 0 ||
			gf.Compute.GpuPortionLimit.IsSet() && *gf.Compute.GpuPortionLimit.Get() > 0) {
		return true
	}

	if len(gf.Container.MasterArgs) > 0 ||
		len(gf.Container.MasterCommand) > 0 ||
		len(gf.Container.MasterEnvironmentVariables) > 0 ||
		len(gf.Container.MasterDatasourceAssetIds) > 0 ||
		len(gf.Compute.MasterExtendedResources) > 0 ||
		convutils.HasValue(gf.Compute.MasterGpuDeviceRequest) ||
		gf.Compute.MasterGpuMemoryPortionRequest.IsSet() ||
		gf.Compute.MasterGpuMemoryPortionLimit.IsSet() ||
		gf.Compute.MasterGpuComputeMode != "" ||
		gf.Compute.MasterGpuPortionLimit.IsSet() && *gf.Compute.MasterGpuPortionLimit.Get() > 0 ||
		gf.Compute.MasterGpuPortionRequest.IsSet() && *gf.Compute.MasterGpuPortionRequest.Get() > 0 ||
		gf.Compute.MasterCpuCoreRequest.IsSet() && *gf.Compute.MasterCpuCoreRequest.Get() > 0 ||
		gf.Compute.MasterCpuCoreLimit.IsSet() && *gf.Compute.MasterCpuCoreLimit.Get() > 0 ||
		gf.Compute.MasterCpuMemoryLimit != "" ||
		gf.Compute.MasterCpuMemoryRequest != "" ||
		gf.Storage.MasterNoPVC ||
		gf.Lifecycle.MasterRestartPolicy.IsSet() {
		return true
	}
	return false
}

func uniqueWorkerSpecRequired(gf *GeneralFlags) bool {
	if gf.DistributedInference.WorkerPreferredPodTopology != "" ||
		gf.DistributedInference.WorkerRequiredPodTopology != "" {
		return true
	}

	if gf.Container.WorkerImage != "" ||
		gf.Container.WorkerImagePullPolicy != "" ||
		gf.Container.WorkerCommand != "" ||
		gf.Container.WorkerArgs != "" ||
		gf.Container.WorkerWorkingDir != "" ||
		len(gf.Container.WorkerEnvironmentVariables) > 0 ||
		len(gf.Container.WorkerFieldRef) > 0 ||
		len(gf.Container.WorkerEnvSecret) > 0 ||
		gf.Container.WorkerCreateHomeDir {
		return true
	}

	if gf.Compute.WorkerGpuMemoryPortionRequest.IsSet() ||
		gf.Compute.WorkerGpuMemoryPortionLimit.IsSet() ||
		gf.Compute.WorkerGpuComputeMode != "" ||
		gf.Compute.WorkerGpuDevicesRequest.IsSet() && *gf.Compute.WorkerGpuDevicesRequest.Get() > 0 ||
		gf.Compute.WorkerGpuPortionRequest.IsSet() && *gf.Compute.WorkerGpuPortionRequest.Get() > 0 ||
		gf.Compute.WorkerGpuPortionLimit.IsSet() && *gf.Compute.WorkerGpuPortionLimit.Get() > 0 ||
		gf.Compute.WorkerGpuMemoryRequest != "" ||
		gf.Compute.WorkerGpuMemoryLimit != "" ||
		gf.Compute.WorkerCpuCoreRequest.IsSet() && *gf.Compute.WorkerCpuCoreRequest.Get() > 0 ||
		gf.Compute.WorkerCpuCoreLimit.IsSet() && *gf.Compute.WorkerCpuCoreLimit.Get() > 0 ||
		gf.Compute.WorkerCpuMemoryRequest != "" ||
		gf.Compute.WorkerCpuMemoryLimit != "" ||
		len(gf.Compute.WorkerExtendedResources) > 0 {
		return true
	}

	if len(gf.Storage.WorkerExistingPvc) > 0 ||
		len(gf.Storage.WorkerNewPvc) > 0 ||
		len(gf.Storage.WorkerSecretVolume) > 0 ||
		len(gf.Storage.WorkerConfigmapMapVolume) > 0 ||
		len(gf.Storage.WorkerEmptyDirVolume) > 0 {
		return true
	}

	if gf.Security.WorkerRunAsUid != 0 ||
		gf.Security.WorkerRunAsGid != 0 ||
		gf.Security.WorkerRunAsNonRoot ||
		len(gf.Security.WorkerSupplementalGroups) > 0 ||
		len(gf.Security.WorkerCapabilities) > 0 ||
		gf.Security.WorkerPrivileged ||
		gf.Security.WorkerAllowPrivilegeEscalation {
		return true
	}

	return false
}
