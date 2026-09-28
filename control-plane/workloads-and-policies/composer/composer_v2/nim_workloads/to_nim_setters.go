/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nim_workloads

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	"github.com/distribution/reference"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsManagerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func imageToNimImage(src *string, target any, cookie any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*v1alpha1.Image)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("imageToNimImage: target type %T is not *v1alpha1.Image", target))
	}

	// parse the image reference
	ref, err := reference.ParseNormalizedNamed(*src)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("invalid image reference: %v", err))
	}

	// extract repository name
	repository := reference.FamiliarName(ref)

	// extract tag or digest. tag is mandatory in NIMService but not in our api
	var tag string
	if tagged, ok := ref.(reference.Tagged); ok {
		tag = tagged.Tag()
	} else {
		tag = "latest"
	}

	(*targetPtr).Repository = repository
	(*targetPtr).Tag = tag

	return nil
}

func servingPortToServiceSetter(src *workloadsManagerApi.NimServiceServingPort, target any, _ any) *workloadsApi.Error {
	targetPtr, ok := target.(*v1alpha1.Service)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("servingPortToServiceSetter: target type %T is not a *nim1alpha1.Service", target))
	}

	if convutils.HasValue(src.ServiceType) {
		(*targetPtr).Type = corev1.ServiceType(*src.ServiceType.Get())
	} else {
		(*targetPtr).Type = corev1.ServiceTypeClusterIP
	}

	(*targetPtr).Port = src.Port.Get()
	(*targetPtr).GRPCPort = src.GrpcPort.Get()
	(*targetPtr).MetricsPort = src.MetricsPort.Get()

	return nil
}

func readinessProbeSetter(src *workloadsApi.Probe, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*v1alpha1.Probe)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("target type %T is not a *v1alpha1.Probe", target))
	}

	readinessHandler := src.Handler.Get()

	// httpGet.port is mandatory
	// init httpGet probe handler
	probeHandler := &corev1.ProbeHandler{
		HTTPGet: &corev1.HTTPGetAction{
			Port: intstr.IntOrString{Type: intstr.Int, IntVal: *readinessHandler.HttpGet.Port.Get()},
		},
	}

	// set optional httpGet fields
	if convutils.HasValue(readinessHandler.HttpGet.Path) {
		probeHandler.HTTPGet.Path = *readinessHandler.HttpGet.Path.Get()
	}

	if convutils.HasValue(readinessHandler.HttpGet.Host) {
		probeHandler.HTTPGet.Host = *readinessHandler.HttpGet.Host.Get()
	}

	if convutils.HasValue(readinessHandler.HttpGet.Scheme) {
		probeHandler.HTTPGet.Scheme = corev1.URIScheme(strings.ToUpper(string(*readinessHandler.HttpGet.Scheme.Get())))
	}

	// init probe
	probe := &corev1.Probe{
		ProbeHandler: *probeHandler,
	}

	// set optional fields
	if convutils.HasValue(src.InitialDelaySeconds) {
		probe.InitialDelaySeconds = *src.InitialDelaySeconds.Get()
	}

	if convutils.HasValue(src.TimeoutSeconds) {
		probe.TimeoutSeconds = *src.TimeoutSeconds.Get()
	}

	if convutils.HasValue(src.PeriodSeconds) {
		probe.PeriodSeconds = *src.PeriodSeconds.Get()
	}

	if convutils.HasValue(src.SuccessThreshold) {
		probe.SuccessThreshold = *src.SuccessThreshold.Get()
	}

	if convutils.HasValue(src.FailureThreshold) {
		probe.FailureThreshold = *src.FailureThreshold.Get()
	}

	targetPtr.Enabled = convutils.Ref(true)
	targetPtr.Probe = probe

	return nil
}

func nimCacheSetter(src *workloadsManagerApi.NimCache, target any, _ any) *workloadsApi.Error {
	targetPtr, ok := target.(*v1alpha1.NIMCacheVolSpec)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("modelStore.nimCache: target type %T is not NIMCacheVolSpec", target))
	}

	targetPtr.Name = convutils.Deref(src.Name.Get(), "")

	if convutils.HasValue(src.Profile) {
		targetPtr.Profile = *src.Profile.Get()
	}

	return nil
}

func nimPvcSetter(src map[string]*workloadsApi.PvcInstance, target any, _ any) *workloadsApi.Error {
	if len(src) != 1 {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("expecting 1 pvc instance but got %d", len(src)))
	}

	targetPtr, ok := target.(*v1alpha1.NIMServiceStorage)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("modelStore.nimPvc: target type %T is not *v1alpha1.PersistentVolumeClaim", target))
	}

	for _, pvcInst := range src {
		targetPtr.ReadOnly = convutils.Ref(convutils.BoolIsTrue(pvcInst.ReadOnly.Get()))

		targetPvc := &targetPtr.PVC

		if convutils.BoolIsTrue(pvcInst.ExistingPvc.Get()) {
			targetPvc.Create = convutils.Ref(false)
			targetPvc.Name = convutils.Deref(pvcInst.ClaimName.Get(), "")
		} else {
			targetPvc.Create = convutils.Ref(true)
		}

		if convutils.HasValue(pvcInst.ClaimInfo) {
			claimInfo := pvcInst.GetClaimInfo()

			if convutils.HasValue(claimInfo.StorageClass) {
				targetPvc.StorageClass = claimInfo.GetStorageClass()
			}

			if convutils.HasValue(claimInfo.Size) {
				targetPvc.Size = claimInfo.GetSize()
			}

			if convutils.HasValue(claimInfo.AccessModes) {
				accessModes := claimInfo.GetAccessModes()

				if convutils.BoolIsTrue(accessModes.ReadWriteOnce.Get()) {
					targetPvc.VolumeAccessMode = corev1.ReadWriteOnce
				} else if convutils.BoolIsTrue(accessModes.ReadOnlyMany.Get()) {
					targetPvc.VolumeAccessMode = corev1.ReadOnlyMany
				} else if convutils.BoolIsTrue(accessModes.ReadWriteMany.Get()) {
					targetPvc.VolumeAccessMode = corev1.ReadWriteMany
				}
			}
		}
	}

	return nil
}

func gpuSizingInfoSetter(src *composer.GpuSizingInfo, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*corev1.ResourceRequirements)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("target type %T is not *corev1.ResourceRequirements", target))
	}

	// set gpu resources limits
	// the setting should be avoided if fractional gpu is used
	// note that a portion > 1 is not considered fractional for this matter
	isFractionalGpu := (src.PortionRequest != nil && *src.PortionRequest < 1 && *src.PortionRequest != 0) || (src.MemoryRequest != nil && *src.MemoryRequest != "")

	if !isFractionalGpu && convutils.Deref(src.NumDevices, 0) > 0 {
		err := composer.IntToK8sResourceList[int32](src.NumDevices, &targetPtr.Limits, composer.NvidiaResourceName)
		if err != nil {
			return err
		}
		err = composer.IntToK8sResourceList[int32](src.NumDevices, &targetPtr.Requests, composer.NvidiaResourceName)
		if err != nil {
			return err
		}
	}

	return nil
}

func nimScaleSetter(src *workloadsManagerApi.NimServiceSpecAllOfAutoscaling, target any, cookie any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*v1alpha1.HorizontalPodAutoscalerSpec)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("autoscaling: target type %T is not a *workloadsManagerApi.HorizontalPodAutoscalerSpec", target))
	}

	minReplicas := convutils.Deref(src.MinReplicas.Get(), autoscalingDefaultMinReplicas)
	maxReplicas := convutils.Deref(src.MaxReplicas.Get(), minReplicas)

	metricSpec := autoscalingv2.MetricSpec{}

	metric := convutils.Deref(src.Metric.Get(), "")

	switch metric {
	case autoscalingMetricCpuUtilization:
		metricSpec.Type = autoscalingv2.ResourceMetricSourceType
		metricSpec.Resource = &autoscalingv2.ResourceMetricSource{
			Name: corev1.ResourceCPU,
			Target: autoscalingv2.MetricTarget{
				Type:               autoscalingv2.UtilizationMetricType,
				AverageUtilization: src.MetricThreshold.Get(),
			},
		}
	case autoscalingMetricMemoryUtilization:
		quantity, err := resource.ParseQuantity(fmt.Sprintf("%d", *src.MetricThreshold.Get()))
		if err != nil {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("failed to parse metricThreshold '%d': %v", *src.MetricThreshold.Get(), err))
		}

		metricSpec.Type = autoscalingv2.ResourceMetricSourceType
		metricSpec.Resource = &autoscalingv2.ResourceMetricSource{
			Name: corev1.ResourceMemory,
			Target: autoscalingv2.MetricTarget{
				Type:         autoscalingv2.UtilizationMetricType,
				AverageValue: convutils.Ref(quantity),
			},
		}
	default:
		quantity, err := resource.ParseQuantity(fmt.Sprintf("%d", *src.MetricThreshold.Get()))
		if err != nil {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("failed to parse metricThreshold '%d': %v", *src.MetricThreshold.Get(), err))
		}

		metricSpec.Type = autoscalingv2.PodsMetricSourceType
		metricSpec.Pods = &autoscalingv2.PodsMetricSource{
			Metric: autoscalingv2.MetricIdentifier{
				Name: metric,
			},
			Target: autoscalingv2.MetricTarget{
				Type:         autoscalingv2.AverageValueMetricType,
				AverageValue: convutils.Ref(quantity),
			},
		}
	}

	var behavior *autoscalingv2.HorizontalPodAutoscalerBehavior

	if src.ScaleWindowSeconds.Get() != nil {
		behavior = &autoscalingv2.HorizontalPodAutoscalerBehavior{
			ScaleUp: &autoscalingv2.HPAScalingRules{
				StabilizationWindowSeconds: src.ScaleWindowSeconds.Get(),
			},
			ScaleDown: &autoscalingv2.HPAScalingRules{
				StabilizationWindowSeconds: src.ScaleWindowSeconds.Get(),
			},
		}
	}

	*targetPtr = v1alpha1.HorizontalPodAutoscalerSpec{
		MinReplicas: convutils.Ref(minReplicas),
		MaxReplicas: maxReplicas,
		Metrics: []autoscalingv2.MetricSpec{
			metricSpec,
		},
		Behavior: behavior,
	}

	return nil
}
