/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nim_workloads

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	nimv1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsManagerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	defaultReplicas = int32(1)

	ingressControllerNginx   = "nginx"
	ingressControllerHAProxy = "haproxy"
	defaultIngressClassName  = ingressControllerNginx

	portNameHttp = "api"
	portNameGrpc = "grpc"

	bytesToMB = 1000000

	autoscalingDefaultMinReplicas      = 1
	autoscalingMetricCpuUtilization    = "cpu"
	autoscalingMetricMemoryUtilization = "memory"
)

type FnNimServingPortSetter func(src *workloadsManagerApi.NimServiceServingPort, target any, cookie any) *workloadsApi.Error
type FnNimIngressSetter func(src *IngressInfo, target any, cookie any) *workloadsApi.Error
type FnNimCacheSetter func(src *workloadsManagerApi.NimCache, target any, cookie any) *workloadsApi.Error
type FnNimScaleSetter func(src *workloadsManagerApi.NimServiceSpecAllOfAutoscaling, target any, cookie any) *workloadsApi.Error

type NimServiceComposer struct {
	fc                  *composer.FieldsComposer
	clusterInfo         *clustersApi.DisplayedCluster
	forceOldFlowIngress bool // set by the service layer for legacy NIM services
	useNewFlowIngress   bool // computed once during Compose(); read by the service layer afterward
}

// SetForceOldFlowIngress tells the composer to keep ingress in the NIM CR even on >= 2.25 clusters.
// Used when updating a NIM service that was originally created with operator-managed ingress.
func (nsc *NimServiceComposer) SetForceOldFlowIngress(v bool) {
	nsc.forceOldFlowIngress = v
}

// UsedNewFlowIngress reports whether ingress was routed via workload v2 networking during composition.
// Valid only after Compose() has been called.
func (nsc *NimServiceComposer) UsedNewFlowIngress() bool {
	return nsc.useNewFlowIngress
}

func NewNimServiceComposer(
	workloadName string,
	requester string,
	clusterInfo *clustersApi.DisplayedCluster,
	projectName string,
	options composer.FieldsComposerOptions,
) *NimServiceComposer {
	result := NimServiceComposer{
		fc: &composer.FieldsComposer{
			WorkloadType:      composer.WorkloadType_NIM,
			WorkloadName:      workloadName,
			ProjectName:       projectName,
			SubmittedBy:       requester,
			Options:           options,
			ItemizedNumbering: make(itemized.ItemizedNumbering),
			OauthPortNumber:   composer.Oauth2ContainerDefaultPort,
		},
		clusterInfo: clusterInfo,
	}

	return &result
}

func (nsc *NimServiceComposer) Compose(ctx context.Context, src *workloadsManagerApi.NimServiceSpec,
	target *nimv1alpha1.NIMServiceSpec,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		// Compute once whether ingress should use workload v2 networking (>= 2.25) or stay in the NIM CR.
		// forceOldFlowIngress overrides this for legacy services that were created with operator-managed ingress.
		if !nsc.forceOldFlowIngress {
			if clusterVersion := nsc.clusterInfo.GetVersion(); clusterVersion != "" {
				if cap, capErr := checker.NewClusterCapabilitiesCheckerForVersion(clusterVersion); capErr == nil {
					nsc.useNewFlowIngress, _ = cap.ValidateClusterCapability(ctx, checker.MinVersionNIMIngressViaWorkloadV2)
				}
			}
		}

		runner := &composer.StepRunner{}

		var gpuSizingInfo composer.GpuSizingInfo

		runner.Batch(
			nsc.fc.CopyString(convutils.Ref(composer.RunaiSchedulerName), &target.SchedulerName, composer.StringToString, nil),
			nsc.fc.ComposeImage(src.Image.Get(), &target.Image, imageToNimImage, nil),
			nsc.fc.CopyString(composer.EnumToString(src.ImagePullPolicy.Get()), &target.Image.PullPolicy, composer.StringToString, nil),
			nsc.fc.ComposeImagePullSecrets(nil, composer.CastSlice[workloadsManagerApi.ImagePullSecret, workloadsApi.ImagePullSecret](src.ImagePullSecrets),
				&target.Image.PullSecrets, composer.AddStringsAoArray, nil),
			nsc.fc.ComposeEnvironmentVariables(nil, composer.CastSlice[workloadsManagerApi.EnvironmentVariable, workloadsApi.EnvironmentVariable](src.EnvironmentVariables),
				&target.Env, composer.EnvVarsToK8sEnv, nil, nil),
			nsc.fc.ComposeLabels(composer.CastSlice[workloadsManagerApi.Label, workloadsApi.Label](src.Labels),
				&target.Labels, composer.MapToMapString, nil),
			nsc.fc.ComposeAnnotations(composer.CastSlice[workloadsManagerApi.Annotation, workloadsApi.Annotation](src.Annotations),
				&target.Annotations, composer.MapToMapString, nil),
			nsc.fc.ComposeTolerations(itemized.TolerationPrefix, composer.CastSlice[workloadsManagerApi.Toleration, workloadsApi.Toleration](src.Tolerations),
				&target.Tolerations, composer.MapToK8sTolerations, nil),
			nsc.fc.CopyMandatoryString("ngcAuthSecret", src.NgcAuthSecret.Get(), &target.AuthSecret, composer.StringToString, nil),
			nsc.fc.CopyBoolean(convutils.Ref(true), &target.Metrics.Enabled, composer.BoolToBool, nil),

			nsc.processSecurityFields(src.Security.Get(), target),
			nsc.composeExpose(src.ServingPort.Get(), &target.Expose),
			nsc.processModelStoreFields(src.ModelStore.Get(), &target.Storage),
			nsc.processComputeFields(ctx, src.Compute.Get(), target, &gpuSizingInfo),
		)

		runner.Batch(
			nsc.composeAutoScaling(src.Autoscaling.Get(), target),
			nsc.composeReplicas(src.Replicas.Get(), src.Autoscaling.Get(), &target.Replicas, composer.NonZeroIntToInt, nil),
			nsc.composeMultiNode(src.MultiNode.Get(), src.Autoscaling.Get(), &gpuSizingInfo, &target.MultiNode),
		)

		if convutils.HasValue(src.Probes) {
			runner.Run(nsc.fc.ComposeReadinessProbe(composer.CastStruct[workloadsManagerApi.Probe, workloadsApi.Probe](src.Probes.Get().Readiness.Get()),
				&target.ReadinessProbe, readinessProbeSetter, nil))
		}

		return runner.Error()
	}
}

func (nsc *NimServiceComposer) processSecurityFields(src *workloadsManagerApi.NimServiceSpecAllOfSecurity, target *nimv1alpha1.NIMServiceSpec) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		runner := &composer.StepRunner{}

		runner.Batch(
			nsc.fc.CopyInt(src.RunAsUid.Get(), &target.UserID, composer.IntToInt, nil),
			nsc.fc.CopyInt(src.RunAsGid.Get(), &target.GroupID, composer.IntToInt, nil),
		)

		return runner.Error()
	}
}

func (nsc *NimServiceComposer) composeExpose(src *workloadsManagerApi.NimServiceServingPort, target *nimv1alpha1.Expose) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		var ingressInfo IngressInfo

		runner := &composer.StepRunner{}

		runner.Batch(
			nsc.composeService(src, &target.Service, servingPortToServiceSetter, nil),
			//nolint:staticcheck // Ingress is deprecated but still supported for backward compatibility
			nsc.composeIngress(src, &target.Ingress, nimIngressSetter, nil, &ingressInfo),
		)

		return runner.Error()
	}
}

func (nsc *NimServiceComposer) processModelStoreFields(src *workloadsManagerApi.NimServiceSpecAllOfModelStore, target *nimv1alpha1.NIMServiceStorage) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		// exactly one modelStore type is required
		if src == nil || (!convutils.HasValue(src.NimCache) && !convutils.HasValue(src.Pvc)) {
			return workloadsApi.NewError(http.StatusBadRequest, "modelStore: either nimCache or pvc is mandatory")
		}
		if convutils.HasValue(src.NimCache) && convutils.HasValue(src.Pvc) {
			return workloadsApi.NewError(http.StatusBadRequest, "modelStore: nimCache and pvc are mutually exclusive")
		}

		runner := &composer.StepRunner{}

		pvcAnnotations := make([]workloadsApi.Annotation, 0)

		runner.Batch(
			nsc.composeNimCache(src.NimCache.Get(), &target.NIMCache, nimCacheSetter, nil),
			nsc.composePvcStore(src.Pvc.Get(), target, nimPvcSetter, nil, &pvcAnnotations),
		)

		runner.Run(nsc.fc.ComposeAnnotations(pvcAnnotations, &target.PVC.Annotations, composer.MapToMapString, nil))

		return runner.Error()
	}
}

func (nsc *NimServiceComposer) composeNimCache(src *workloadsManagerApi.NimCache,
	target *nimv1alpha1.NIMCacheVolSpec, setter FnNimCacheSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		if !convutils.HasValue(src.Name) {
			return workloadsApi.NewError(http.StatusBadRequest, "modelStore.nimCache: name is mandatory")
		}

		return setter(src, target, cookie)
	}
}

func (nsc *NimServiceComposer) composePvcStore(src *workloadsManagerApi.NimServicePvcFields,
	target *nimv1alpha1.NIMServiceStorage, setter composer.FnMapToPvcSetter, cookie any, annotations *[]workloadsApi.Annotation,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		pvcInst := nimToPvc(src)
		if pvcInst == nil {
			return nil
		}

		if convutils.HasValue(pvcInst.ClaimInfo) && convutils.HasValue(pvcInst.GetClaimInfo().VolumeMode) {
			if *pvcInst.GetClaimInfo().VolumeMode.Get() == workloadsApi.PVCVOLUMEMODE_BLOCK {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("modelStore.pvc.claimInfo: volumeMode '%s' is not supported", workloadsApi.PVCVOLUMEMODE_BLOCK))
			}
		}

		err := nsc.fc.ComposePvc(itemized.PvcPrefix, []workloadsApi.PvcInstance{*pvcInst}, target, setter, cookie)()
		if err != nil {
			return err
		}

		if len(src.GetClaimInfo().AddedAttrValues) != 0 {
			for _, attr := range src.GetClaimInfo().AddedAttrValues {
				annotation := workloadsApi.Annotation{}

				annotation.Name.Set(convutils.Ref(attr.Key))

				if attr.Value != nil {
					annotation.Value.Set(attr.Value)
				}

				*annotations = append(*annotations, annotation)
			}
		}

		return nil
	}
}

func nimToPvc(src *workloadsManagerApi.NimServicePvcFields) *workloadsApi.PvcInstance {
	if src == nil {
		return nil
	}

	pvcInst := &workloadsApi.PvcInstance{}

	if convutils.HasValue(src.ExistingPvc) {
		pvcInst.ExistingPvc.Set(src.ExistingPvc.Get())
	}

	if convutils.HasValue(src.ClaimName) {
		pvcInst.ClaimName.Set(src.ClaimName.Get())
	}

	if convutils.HasValue(src.ReadOnly) {
		pvcInst.ReadOnly.Set(src.ReadOnly.Get())
	}

	if convutils.HasValue(src.ClaimInfo) {
		srcClaimInfo := src.ClaimInfo.Get()
		claimInfo := workloadsApi.ClaimInfo{}

		if convutils.HasValue(srcClaimInfo.Size) {
			claimInfo.Size.Set(srcClaimInfo.Size.Get())
		}

		if convutils.HasValue(srcClaimInfo.StorageClass) {
			claimInfo.StorageClass.Set(srcClaimInfo.StorageClass.Get())
		}

		if convutils.HasValue(srcClaimInfo.AccessModes) {
			accessModes := workloadsApi.PvcAccessModes{
				ReadOnlyMany:  *workloadsApi.NewNullableBool(convutils.Ref(convutils.BoolIsTrue(srcClaimInfo.AccessModes.Get().ReadOnlyMany.Get()))),
				ReadWriteOnce: *workloadsApi.NewNullableBool(convutils.Ref(convutils.BoolIsTrue(srcClaimInfo.AccessModes.Get().ReadWriteOnce.Get()))),
				ReadWriteMany: *workloadsApi.NewNullableBool(convutils.Ref(convutils.BoolIsTrue(srcClaimInfo.AccessModes.Get().ReadWriteMany.Get()))),
			}
			claimInfo.AccessModes.Set(&accessModes)
		}

		if convutils.HasValue(srcClaimInfo.VolumeMode) {
			claimInfo.VolumeMode.Set(composer.CastStruct[workloadsManagerApi.PvcVolumeMode, workloadsApi.PvcVolumeMode](srcClaimInfo.VolumeMode.Get()))
		}

		pvcInst.ClaimInfo.Set(&claimInfo)
	}

	return pvcInst
}

func (nsc *NimServiceComposer) processComputeFields(ctx context.Context, src *workloadsManagerApi.NimServiceSpecAllOfCompute,
	target *nimv1alpha1.NIMServiceSpec, gpuSizingInfo *composer.GpuSizingInfo,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		portion, err := gpucompat.ResolveLegacyGPUMemory("Request", src.GpuPortionRequest.Get(), src.GpuMemoryPortionRequest.Get())
		if err != nil {
			return workloadsApi.NewError(http.StatusBadRequest, err.Error())
		}
		limit, err := gpucompat.ResolveLegacyGPUMemory("Limit", src.GpuPortionLimit.Get(), src.GpuMemoryPortionLimit.Get())
		if err != nil {
			return workloadsApi.NewError(http.StatusBadRequest, err.Error())
		}

		runner := &composer.StepRunner{}

		if target.Resources == nil {
			target.Resources = &corev1.ResourceRequirements{}
		}

		runner.Batch(
			nsc.fc.ComposeCpuCoreRequest(src.CpuCoreRequest.Get(), &target.Resources.Requests, composer.FloatToK8sResourceList, composer.ResourceCpu),
			nsc.fc.ComposeCpuCoreLimit(src.CpuCoreLimit.Get(), src.CpuCoreRequest.Get(), &target.Resources.Limits, composer.FloatToK8sResourceList, composer.ResourceCpu),
			nsc.fc.ComposeCpuMemoryRequest(src.CpuMemoryRequest.Get(), &target.Resources.Requests, composer.StringToK8sResourceList, composer.ResourceMemory),
			nsc.fc.ComposeCpuMemoryLimit(src.CpuMemoryLimit.Get(), src.CpuMemoryRequest.Get(), &target.Resources.Limits, composer.StringToK8sResourceList, composer.ResourceMemory),

			nsc.fc.ComposeGpuRequestsAndLimits(ctx, src.GpuDevicesRequest.Get(),
				composer.CastStruct[workloadsManagerApi.GpuRequestType, workloadsApi.GpuRequestType](src.GpuRequestType.Get()),
				portion, limit,
				src.GpuMemoryRequest.Get(), src.GpuMemoryLimit.Get(), gpuSizingInfoSetter,
				target.Resources, nil, gpuSizingInfo),
		)

		if composer.IsEmptyRequests(target.Resources) {
			target.Resources = nil
		}

		if gpuSizingInfo.IsGpuRequested() {
			err := composer_v2.ValidateGpuVendorSupported(nsc.clusterInfo)
			if err != nil {
				return workloadsApi.NewError(err.Code, fmt.Sprintf("compute: failed to validate gpu vendor is nvidia: %v", err.Message))
			}
			if src.GpuComputeMode.Get() != nil {
				version := nsc.clusterInfo.GetVersion()
				if version == "" {
					return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuComputeMode: cluster version is unknown")
				}
				capabilities, err := checker.NewClusterCapabilitiesCheckerForVersion(version)
				if err != nil {
					return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("cluster version: %v", err))
				}
				supported, err := capabilities.ValidateClusterCapability(ctx, checker.MinVersionGpuComputeMode)
				if err != nil {
					return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("cluster capability check: %v", err))
				}
				nsc.fc.Options.GpuComputeModeSupported = supported
			}
		}

		nimService := nimv1alpha1.NIMService{}
		nimService.Name = nsc.fc.WorkloadName
		runner.Batch(
			nsc.fc.ComposeGpuComputeMode(composer.EnumToString(src.GpuComputeMode.Get()), gpuSizingInfo, nimService.GetContainerName(), &target.Annotations, composer.StringToMapString),
			nsc.composeMemoryLimitEnv(gpuSizingInfo, &target.Env, composer.StringToEnvVar, composer.GpuMemoryLimitEnvVar),
			nsc.composePortionLimitEnv(gpuSizingInfo, &target.Env, composer.StringToEnvVar, composer.GpuMemoryLimitEnvVar),
			nsc.composeMemoryReqAnnotation(gpuSizingInfo, &target.Annotations, composer.StringToMapString, composer.GpuMemoryAnnotation),
			nsc.composePortionReqAnnotation(gpuSizingInfo, &target.Annotations, composer.StringToMapString, composer.GpuFractionAnnotation),
			nsc.composeNumDevicesAnnotation(gpuSizingInfo, &target.Annotations, composer.StringToMapString, composer.GpuFractionDevicesAnnotation),
		)

		return runner.Error()
	}
}

func (nsc *NimServiceComposer) composeMemoryLimitEnv(src *composer.GpuSizingInfo, target any, setter composer.FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src.MemoryLimit != nil && src.MemoryRequest != nil {
			return setter(src.MemoryLimit, target, cookie)
		}
		return nil
	}
}

func (nsc *NimServiceComposer) composePortionLimitEnv(src *composer.GpuSizingInfo, target any, setter composer.FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src.PortionLimit != nil && src.PortionRequest != nil {
			value := strconv.FormatFloat(*src.PortionLimit, 'f', -1, 64)
			return setter(&value, target, cookie)
		}
		return nil
	}
}

func (nsc *NimServiceComposer) composeMemoryReqAnnotation(src *composer.GpuSizingInfo, target any, setter composer.FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src.MemoryRequest != nil {
			gpuMemoryQuantity, err := resource.ParseQuantity(*src.MemoryRequest)
			if err != nil { // not expected, already validated before
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf(composer.MemoryFormatError, "gpuMemoryLimit"))
			}

			quantityStr := strconv.Itoa(int(gpuMemoryQuantity.Value() / bytesToMB))

			return setter(&quantityStr, target, cookie)
		}
		return nil
	}
}

func (nsc *NimServiceComposer) composePortionReqAnnotation(src *composer.GpuSizingInfo, target any, setter composer.FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src.PortionRequest != nil {
			quantityStr := strconv.FormatFloat(*src.PortionRequest, 'f', -1, 64)
			return setter(&quantityStr, target, cookie)
		}
		return nil
	}
}

func (nsc *NimServiceComposer) composeNumDevicesAnnotation(src *composer.GpuSizingInfo, target any, setter composer.FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src.PortionRequest != nil {
			gpuDevices := 1
			if src.NumDevices != nil && *src.NumDevices > 0 {
				gpuDevices = int(*src.NumDevices)
			}

			devicesStr := fmt.Sprintf("%d", gpuDevices)

			return setter(&devicesStr, target, cookie)
		}
		return nil
	}
}

func (nsc *NimServiceComposer) composeReplicas(src *int32, autoScaling *workloadsManagerApi.NimServiceSpecAllOfAutoscaling,
	target any, setter composer.FnIntSetter[int32], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if autoScaling != nil {
			if src != nil {
				return workloadsApi.NewError(http.StatusBadRequest, "autoscaling: replicas is mutually exclusive with autoscaling")
			}
			// Leave replicas unset when the operator manages scaling.
			return nil
		}

		if src == nil {
			return setter(convutils.Ref(defaultReplicas), target, cookie)
		}

		return setter(src, target, cookie)
	}
}

func (nsc *NimServiceComposer) composeAutoScaling(src *workloadsManagerApi.NimServiceSpecAllOfAutoscaling,
	target *nimv1alpha1.NIMServiceSpec,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		// if autoscaling is enabled, set the inference platform to standalone, to cover the case that
		// the default platform is kserve, but the user requests HPA, which we run as standalone.
		// we don't support HPA over kserve.
		// when KPA support is implemented, this logic should be changed.
		target.InferencePlatform = nimv1alpha1.PlatformTypeStandalone

		runner := &composer.StepRunner{}

		runner.Batch(
			nsc.composeScaleConfig(src, &target.Scale.HPA, nimScaleSetter, nil),
			nsc.fc.CopyBoolean(convutils.Ref(true), &target.Scale.Enabled, composer.BoolToBool, nil),
		)

		return runner.Error()
	}
}

func (nsc *NimServiceComposer) composeScaleConfig(src *workloadsManagerApi.NimServiceSpecAllOfAutoscaling, target any, setter FnNimScaleSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src != nil {
			minReplicas := convutils.Deref(src.MinReplicas.Get(), autoscalingDefaultMinReplicas)
			maxReplicas := convutils.Deref(src.MaxReplicas.Get(), minReplicas)

			if maxReplicas < minReplicas {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("autoscaling: maxReplicas (%d) must be no less than minReplicas (%d)", maxReplicas, minReplicas))
			}

			if !convutils.HasValue(src.Metric) {
				return workloadsApi.NewError(http.StatusBadRequest, "autoscaling: metric is mandatory")
			}

			if !convutils.HasValue(src.MetricThreshold) {
				return workloadsApi.NewError(http.StatusBadRequest, "autoscaling: metricThreshold is mandatory")
			}
		}

		return setter(src, target, cookie)
	}
}

func (nsc *NimServiceComposer) composeMultiNode(src *workloadsManagerApi.NimServiceMultiNode, autoScaling *workloadsManagerApi.NimServiceSpecAllOfAutoscaling,
	gpuSizingInfo *composer.GpuSizingInfo, target **nimv1alpha1.NimServiceMultiNodeConfig,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		// If multiNode is not set, skip multi-node configuration
		if src == nil {
			return nil
		}

		if autoScaling != nil {
			return workloadsApi.NewError(http.StatusBadRequest, "autoscaling: multiNode is mutually exclusive with autoscaling")
		}

		// If multiNode is set, workers must be provided
		if !convutils.HasValue(src.Workers) {
			return workloadsApi.NewError(http.StatusBadRequest, "multiNode: workers is mandatory")
		}

		workers := *src.Workers.Get()

		// Get gpuDevicesRequest from compute
		var tensorParallelism uint32 = 1 // default to 1 if no GPU request

		gpuDevicesRequested := convutils.Deref(gpuSizingInfo.NumDevices, 0)
		if gpuDevicesRequested > 0 {
			tensorParallelism = uint32(gpuDevicesRequested)
		}

		// pipeline = workers + 1 (includes leader)
		pipelineParallelism := uint32(workers + 1)

		// Initialize MultiNode and set parallelism in NIMServiceSpec
		*target = &nimv1alpha1.NimServiceMultiNodeConfig{
			Parallelism: &nimv1alpha1.ParallelismSpec{
				Tensor:   convutils.Ref(tensorParallelism),
				Pipeline: convutils.Ref(pipelineParallelism),
			},
		}

		return nil
	}
}
