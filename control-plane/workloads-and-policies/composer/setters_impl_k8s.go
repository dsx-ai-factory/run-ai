/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	NvidiaResourceName = "nvidia.com/gpu"

	LabelPreemptibilityName = "kai.scheduler/preemptibility"

	ResourceCpu    = "cpu"
	ResourceMemory = "memory"

	// DefaultNodePoolName is the reserved name of the cluster-default node pool.
	// A pod that targets the default node pool must NOT carry the node-pool
	// label, which is expressed via a "DoesNotExist" requirement.
	DefaultNodePoolName = "default"

	// NodePoolNameLabel is the label key used by the cluster to associate
	// nodes (and pods) with a non-default node pool.
	NodePoolNameLabel = "runai/node-pool"
)

// EnvVarsToK8sEnv store array of environment variables info, collected from env variables compose function,
// to k8s environment variables array. An entry whose Name already exists in target is skipped
// (first-wins) so callers can layer additional env vars without clobbering user-supplied values.
func EnvVarsToK8sEnv(envVarsInfo []EnvVarInfo, target any, cookie any) *workloadsApi.Error {
	if len(envVarsInfo) == 0 {
		return nil
	}

	targetPtr, ok := target.(*[]corev1.EnvVar)
	if !ok {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("EnvVarsToK8sEnv: target type %T is not of type *[]corev1.EnvVar", target))
	}

	if *targetPtr == nil {
		*targetPtr = make([]corev1.EnvVar, 0)
	}

	existingNames := datastructs.NewSet[string]()
	for _, ev := range *targetPtr {
		existingNames.Add(ev.Name)
	}

	for _, envVarInfo := range envVarsInfo {
		if envVarInfo.Deleted {
			continue
		}

		if existingNames.Has(envVarInfo.Name) {
			continue
		}
		existingNames.Add(envVarInfo.Name)

		result := corev1.EnvVar{
			Name: envVarInfo.Name,
		}

		switch envVarInfo.ValueType {
		case EnvValueText:
			result.Value = envVarInfo.ValueFrom
		case EnvValueSecret:
			result.ValueFrom = &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: envVarInfo.ValueFrom,
					},
					Key: envVarInfo.Key,
				},
			}
		case EnvValueConfigMap:
			result.ValueFrom = &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: envVarInfo.ValueFrom,
					},
					Key: envVarInfo.Key,
				},
			}
		case EnvValuePodFieldRef:
			result.ValueFrom = &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: envVarInfo.ValueFrom,
				},
			}
		case EnvValueUserCredentials:
			//WAIT_FOR_OFER what to do with credentials?
			continue
		}

		*targetPtr = append(*targetPtr, result)
	}

	return nil
}

func StringToEnvVar(src *string, target any, cookie any) *workloadsApi.Error {
	if src != nil {
		targetPtr, ok := target.(*[]corev1.EnvVar)
		if !ok {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("StringToEnvVar: target type %T is not of type *[]corev1.EnvVar", target))
		}

		if *targetPtr == nil {
			*targetPtr = make([]corev1.EnvVar, 0)
		}

		key, ok := cookie.(string)
		if !ok {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("StringToEnvVar: cookie type %T is not string", cookie))
		}

		*targetPtr = append(*targetPtr, corev1.EnvVar{
			Name:  key,
			Value: *src,
		})
	}
	return nil
}

type IntOrFloatType interface {
	IntType | float64
}

func FloatToK8sResourceList(src *float64, target any, cookie any) *workloadsApi.Error {
	return NumberToK8sResourceList[float64](src, target, cookie)
}

func IntToK8sResourceList[T IntType](src *T, target any, cookie any) *workloadsApi.Error {
	return NumberToK8sResourceList[T](src, target, cookie)
}

func NumberToK8sResourceList[T IntOrFloatType](src *T, target any, cookie any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*corev1.ResourceList)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("NumberToK8sResourceList: target type %T is not of type *corev1.ResourceList", target))
	}

	if *targetPtr == nil {
		*targetPtr = corev1.ResourceList{}
	}

	resourceKey, ok := cookie.(string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("FloatToK8sResourceList: cookie type %T is not *string", cookie))
	}

	var err error
	var quantity resource.Quantity

	if *src >= 1 {
		quantity, err = resource.ParseQuantity(fmt.Sprintf("%v", *src))
	} else {
		milli := int64(math.Round(float64(*src) * 1000))
		quantity, err = resource.ParseQuantity(fmt.Sprintf("%dm", milli))
	}

	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("invalid resource quantity: %v", *src))
	}

	(*targetPtr)[corev1.ResourceName(resourceKey)] = quantity

	return nil
}

func StringToK8sResourceList(src *string, target any, cookie any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*corev1.ResourceList)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("StringToK8sResourceList: target type %T is not of type *ResourceList", target))
	}

	if *targetPtr == nil {
		*targetPtr = corev1.ResourceList{}
	}

	resourceKey, ok := cookie.(string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("StringToK8sResourceList: cookie type %T is not *string", cookie))
	}

	quantity, err := resource.ParseQuantity(*src)
	if err != nil {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("invalid resource quantity: %v", *src))
	}

	(*targetPtr)[corev1.ResourceName(resourceKey)] = quantity

	return nil
}

func IsEmptyRequests(requirements *corev1.ResourceRequirements) bool {
	if len(requirements.Requests) != 0 {
		return false
	}
	if len(requirements.Limits) != 0 {
		return false
	}
	if len(requirements.Claims) != 0 {
		return false
	}
	return true
}

func MapToK8sTolerations(src map[string]*workloadsApi.Toleration, target any, _ any) *workloadsApi.Error {
	targetPtr, ok := target.(*[]corev1.Toleration)
	if !ok {
		return workloadsApi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("MapToK8sTolerations: target type %T is not of type *[]corev1.Toleration", target))
	}

	if *targetPtr == nil {
		*targetPtr = make([]corev1.Toleration, 0)
	}

	for _, id := range SortedMapKeys(src) {
		toleration := src[id]
		if toleration == nil {
			continue
		}

		result := corev1.Toleration{
			Key:      convutils.Deref(toleration.Key.Get(), ""),
			Operator: corev1.TolerationOperator(*toleration.Operator.Get()),
		}

		// set value if provided and operator is Equal
		if convutils.HasValue(toleration.Value) {
			result.Value = *toleration.Value.Get()
		}

		// set effect if provided and not ANY
		if *toleration.Effect.Get() != workloadsApi.TOLERATIONEFFECT_ANY {
			result.Effect = corev1.TaintEffect(*toleration.Effect.Get())
		}

		// set tolerationSeconds if provided and effect is NoExecute
		if convutils.HasValue(toleration.Seconds) {
			tolerationSeconds := int64(*toleration.Seconds.Get())
			result.TolerationSeconds = &tolerationSeconds
		}

		*targetPtr = append(*targetPtr, result)
	}

	return nil
}

func WholeGpusToK8sResources(src *GpuSizingInfo, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*corev1.ResourceRequirements)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("target type %T is not *corev1.ResourceRequirements", target))
	}

	// set gpu resources limits
	// the setting should be avoided if fractional gpu is used
	// note that a portion >= 1 is not considered fractional for this matter
	isFractionalGpu := src.IsFractionalPortion() || (src.MemoryRequest != nil && *src.MemoryRequest != "")
	if isFractionalGpu {
		return nil
	}

	numDevices := convutils.Deref(src.NumDevices, 0)

	// a whole-gpu portion request (>= 1) with no explicit device count implies that many whole
	// GPUs, matching the native controller and the distributed slots/nproc derivation which treat
	// a portion of N as N whole GPUs.
	if numDevices == 0 && src.PortionRequest != nil && *src.PortionRequest >= 1 {
		numDevices = int32(*src.PortionRequest)
	}

	if numDevices > 0 {
		err := IntToK8sResourceList[int32](&numDevices, &targetPtr.Limits, NvidiaResourceName)
		if err != nil {
			return err
		}
		err = IntToK8sResourceList[int32](&numDevices, &targetPtr.Requests, NvidiaResourceName)
		if err != nil {
			return err
		}
	}

	return nil
}

func ExtendedResourcesToK8sResources(src map[string]*workloadsApi.ExtendedResource, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(*corev1.ResourceRequirements)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("ExtendedResourcesToK8sResources: target type %T is not *corev1.ResourceRequirements", target))
	}

	if targetPtr.Requests == nil {
		targetPtr.Requests = corev1.ResourceList{}
	}
	if targetPtr.Limits == nil {
		targetPtr.Limits = corev1.ResourceList{}
	}

	for _, key := range SortedMapKeys(src) {
		extendedResource := src[key]
		if extendedResource == nil || convutils.BoolIsTrue(extendedResource.Exclude.Get()) {
			continue
		}

		resourceName := extendedResource.Resource.Get()
		if resourceName == nil {
			continue
		}

		quantityStr := extendedResource.Quantity.Get()
		if quantityStr == nil {
			continue
		}

		quantity, err := resource.ParseQuantity(*quantityStr)
		if err != nil {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("ExtendedResourcesToK8sResources: invalid quantity %q for resource %q: %v", *quantityStr, *resourceName, err))
		}

		// Extended resources are non-overcommittable: Kubernetes requires the limit to be set
		// and to equal the request, so mirror the single API quantity into both.
		targetPtr.Requests[corev1.ResourceName(*resourceName)] = quantity
		targetPtr.Limits[corev1.ResourceName(*resourceName)] = quantity
	}

	return nil
}

func StringToSupplementalGroups(src *string, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*[]int64)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("StringToSupplementalGroups: target type %T is not *[]int64", target))
	}

	parts := strings.Split(*src, ",")
	groups := make([]int64, 0, len(parts))

	for _, part := range parts {
		gid, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("StringToSupplementalGroups: invalid group id %q: %v", part, err))
		}
		groups = append(groups, gid)
	}

	*targetPtr = groups
	return nil
}

func MapToK8sCapabilities(src map[string]*workloadsApi.Capability, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(**corev1.Capabilities)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("MapToK8sCapabilities: target type %T is not **corev1.Capabilities", target))
	}

	if *targetPtr == nil {
		*targetPtr = &corev1.Capabilities{}
	}

	for _, key := range SortedMapKeys(src) {
		capability := src[key]
		if capability == nil {
			continue
		}

		capName := corev1.Capability(string(*capability))
		(*targetPtr).Add = append((*targetPtr).Add, capName)
	}

	return nil
}

type VolumesAndMounts struct {
	Volume      []corev1.Volume
	VolumeMount []corev1.VolumeMount

	Config *openapi.WorkloadV2Configuration
}

func MapHostPathToK8sVol(src map[string]*workloadsApi.HostPathInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(*VolumesAndMounts)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("MapHostPathToK8sVol: target type %T is not *VolumesAndMounts", target))
	}

	for _, key := range SortedMapKeys(src) {
		instance := src[key]
		if instance == nil {
			continue
		}

		sourcePath := convutils.Deref(instance.Path.Get(), "")
		mountPath := convutils.Deref(instance.MountPath.Get(), "")

		volume := corev1.Volume{
			Name: key,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: sourcePath,
				},
			},
		}
		targetPtr.Volume = append(targetPtr.Volume, volume)

		volumeMount := corev1.VolumeMount{
			Name:      key,
			MountPath: mountPath,
			ReadOnly:  convutils.Deref(instance.ReadOnly.Get(), false),
		}

		if convutils.HasValue(instance.MountPropagation) && *instance.MountPropagation.Get() == workloadsApi.HOSTPATHMOUNTPROPAGATION_HOST_TO_CONTAINER {
			hostToContainer := corev1.MountPropagationHostToContainer
			volumeMount.MountPropagation = &hostToContainer
		}

		targetPtr.VolumeMount = append(targetPtr.VolumeMount, volumeMount)
	}

	return nil
}

func MapNfsToK8sVol(src map[string]*workloadsApi.NfsInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(*VolumesAndMounts)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("MapNfsToK8sVol: target type %T is not *VolumesAndMounts", target))
	}

	for _, key := range SortedMapKeys(src) {
		instance := src[key]
		if instance == nil {
			continue
		}

		server := convutils.Deref(instance.Server.Get(), "")
		path := convutils.Deref(instance.Path.Get(), "")
		mountPath := convutils.Deref(instance.MountPath.Get(), "")

		volume := corev1.Volume{
			Name: key,
			VolumeSource: corev1.VolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server: server,
					Path:   path,
				},
			},
		}
		targetPtr.Volume = append(targetPtr.Volume, volume)

		volumeMount := corev1.VolumeMount{
			Name:      key,
			MountPath: mountPath,
			ReadOnly:  convutils.Deref(instance.ReadOnly.Get(), false),
		}
		targetPtr.VolumeMount = append(targetPtr.VolumeMount, volumeMount)
	}

	return nil
}

func MapConfigMapToK8sVol(src map[string]*workloadsApi.ConfigMapInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(*VolumesAndMounts)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("MapConfigMapToK8sVol: target type %T is not *VolumesAndMounts", target))
	}

	for _, key := range SortedMapKeys(src) {
		instance := src[key]
		if instance == nil {
			continue
		}

		configMapName := convutils.Deref(instance.ConfigMap.Get(), "")
		mountPath := convutils.Deref(instance.MountPath.Get(), "")

		volume := corev1.Volume{
			Name: key,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
				},
			},
		}
		targetPtr.Volume = append(targetPtr.Volume, volume)

		volumeMount := corev1.VolumeMount{
			Name:      key,
			MountPath: mountPath,
		}

		if convutils.IsStringHasValue(instance.SubPath.Get()) {
			volumeMount.SubPath = *instance.SubPath.Get()
		}

		targetPtr.VolumeMount = append(targetPtr.VolumeMount, volumeMount)
	}

	return nil
}

func MapEmptyDirToK8sVol(src map[string]*workloadsApi.EmptyDirInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(*VolumesAndMounts)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("MapEmptyDirToK8sVol: target type %T is not *VolumesAndMounts", target))
	}

	for _, key := range SortedMapKeys(src) {
		instance := src[key]
		if instance == nil {
			continue
		}

		mountPath := convutils.Deref(instance.Path.Get(), "")

		emptyDir := &corev1.EmptyDirVolumeSource{}

		if convutils.HasValue(instance.Medium) && *instance.Medium.Get() == "Memory" {
			emptyDir.Medium = corev1.StorageMediumMemory
		}

		if convutils.HasValue(instance.SizeLimit) {
			quantity, err := resource.ParseQuantity(*instance.SizeLimit.Get())
			if err == nil {
				emptyDir.SizeLimit = &quantity
			}
		}

		volume := corev1.Volume{
			Name: key,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: emptyDir,
			},
		}
		targetPtr.Volume = append(targetPtr.Volume, volume)

		volumeMount := corev1.VolumeMount{
			Name:      key,
			MountPath: mountPath,
		}
		targetPtr.VolumeMount = append(targetPtr.VolumeMount, volumeMount)
	}

	return nil
}

func MapSecretToK8sVol(src map[string]*workloadsApi.SecretInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(*VolumesAndMounts)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("MapSecretToK8sVol: target type %T is not *VolumesAndMounts", target))
	}

	for _, key := range SortedMapKeys(src) {
		instance := src[key]
		if instance == nil {
			continue
		}

		secretName := convutils.Deref(instance.Secret.Get(), "")
		mountPath := convutils.Deref(instance.MountPath.Get(), "")

		volume := corev1.Volume{
			Name: key,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		}
		targetPtr.Volume = append(targetPtr.Volume, volume)

		volumeMount := corev1.VolumeMount{
			Name:      key,
			MountPath: mountPath,
		}
		targetPtr.VolumeMount = append(targetPtr.VolumeMount, volumeMount)
	}

	return nil
}

func MapGitToWorkloadK8s(info *GitSyncInfo, target any, _ any) *workloadsApi.Error {
	if info == nil || len(info.Items) == 0 {
		return nil
	}

	podSpec, ok := target.(*corev1.PodSpec)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("MapGitToWorkloadK8s: target type %T is not *corev1.PodSpec", target))
	}

	if len(podSpec.Containers) == 0 {
		return workloadsApi.NewError(http.StatusInternalServerError, "MapGitToWorkloadK8s: PodSpec has no main container to mount git-sync code into")
	}

	initContainers, err := BuildGitSyncInitContainers(info)
	if err != nil {
		return err
	}

	podSpec.InitContainers = append(podSpec.InitContainers, initContainers...)
	podSpec.Volumes = append(podSpec.Volumes, BuildGitSyncVolumes(info)...)
	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, BuildGitSyncMainVolumeMounts(info)...)
	return nil
}

func UrlsToK8sContainers(info *ExposedUrlInfo, target any, _ any) *workloadsApi.Error {
	if info == nil || len(info.Items) == 0 {
		return nil
	}

	podSpec, ok := target.(*corev1.PodSpec)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("UrlsToK8sContainers: target type %T is not *corev1.PodSpec", target))
	}

	const defaultIngressHttpPort = int32(80)
	allContainerPorts := datastructs.NewSet[int32]()

	var portsToAdd []corev1.ContainerPort

	for _, nick := range SortedMapKeys(info.Items) {
		item := info.Items[nick]
		if item == nil {
			continue
		}

		portValue := defaultIngressHttpPort
		if item.ContainerPort != nil {
			portValue = *item.ContainerPort
		}

		if _, exists := allContainerPorts[portValue]; exists {
			continue
		}

		portsToAdd = append(portsToAdd, corev1.ContainerPort{
			Name:          fmt.Sprintf("port-%d", portValue),
			ContainerPort: portValue,
			Protocol:      corev1.ProtocolTCP,
		})

		allContainerPorts.Add(portValue)
	}

	for containerIdx := range podSpec.Containers {
		podSpec.Containers[containerIdx].Ports = append(podSpec.Containers[containerIdx].Ports, portsToAdd...)
	}

	return nil
}

func NodeAffinityRequiredToK8s(src [][]corev1.NodeSelectorRequirement, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(**corev1.NodeAffinity)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("NodeAffinityRequiredToK8s: target type %T is not **corev1.NodeAffinity", target))
	}

	if *targetPtr == nil {
		*targetPtr = &corev1.NodeAffinity{}
	}

	nodeSelector := &corev1.NodeSelector{
		NodeSelectorTerms: make([]corev1.NodeSelectorTerm, 0, len(src)),
	}

	for _, termReqs := range src {
		k8sTerm := corev1.NodeSelectorTerm{
			MatchExpressions: termReqs,
		}
		nodeSelector.NodeSelectorTerms = append(nodeSelector.NodeSelectorTerms, k8sTerm)
	}

	(*targetPtr).RequiredDuringSchedulingIgnoredDuringExecution = nodeSelector

	return nil
}

// PodAffinityToK8s converts workloadsApi.PodAffinity to corev1.PodAffinity.
// The cookie should contain the workload name (string) used as the label selector value.
func PodAffinityToK8s(src *workloadsApi.PodAffinity, target any, cookie any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(**corev1.PodAffinity)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("PodAffinityToK8s: target type %T is not **corev1.PodAffinity", target))
	}

	workloadName, ok := cookie.(string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("PodAffinityToK8s: cookie type %T is not string (workload name)", cookie))
	}

	topologyKey := convutils.Deref(src.Key.Get(), "")
	if topologyKey == "" {
		return nil
	}

	if *targetPtr == nil {
		*targetPtr = &corev1.PodAffinity{}
	}

	affinityTerm := corev1.PodAffinityTerm{
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"release": workloadName,
			},
		},
		TopologyKey: topologyKey,
	}

	affinityType := convutils.Deref((*workloadsApi.PodAffinityType)(src.Type.Get()), "")
	switch affinityType {
	case workloadsApi.PODAFFINITYTYPE_REQUIRED:
		(*targetPtr).RequiredDuringSchedulingIgnoredDuringExecution = append(
			(*targetPtr).RequiredDuringSchedulingIgnoredDuringExecution,
			affinityTerm,
		)
	case workloadsApi.PODAFFINITYTYPE_PREFERRED:
		(*targetPtr).PreferredDuringSchedulingIgnoredDuringExecution = append(
			(*targetPtr).PreferredDuringSchedulingIgnoredDuringExecution,
			corev1.WeightedPodAffinityTerm{
				Weight:          100,
				PodAffinityTerm: affinityTerm,
			},
		)
	}

	return nil
}

func ProbeToK8s(src *workloadsApi.Probe, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(**corev1.Probe)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("ProbeToK8s: target type %T is not **corev1.Probe", target))
	}

	handler := src.Handler.Get()
	if handler == nil || handler.HttpGet == nil {
		return nil
	}

	httpGet := handler.HttpGet
	port := convutils.Deref(httpGet.Port.Get(), int32(0))
	if port == 0 {
		return nil
	}

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.IntOrString{Type: intstr.Int, IntVal: port},
			},
		},
	}

	if httpGet.Path.IsSet() && httpGet.Path.Get() != nil {
		probe.HTTPGet.Path = *httpGet.Path.Get()
	}

	if httpGet.Host.IsSet() && httpGet.Host.Get() != nil {
		probe.HTTPGet.Host = *httpGet.Host.Get()
	}

	if httpGet.Scheme.IsSet() && httpGet.Scheme.Get() != nil {
		probe.HTTPGet.Scheme = corev1.URIScheme(strings.ToUpper(string(*httpGet.Scheme.Get())))
	}

	if src.InitialDelaySeconds.IsSet() && src.InitialDelaySeconds.Get() != nil {
		probe.InitialDelaySeconds = *src.InitialDelaySeconds.Get()
	}

	if src.TimeoutSeconds.IsSet() && src.TimeoutSeconds.Get() != nil {
		probe.TimeoutSeconds = *src.TimeoutSeconds.Get()
	}

	if src.PeriodSeconds.IsSet() && src.PeriodSeconds.Get() != nil {
		probe.PeriodSeconds = *src.PeriodSeconds.Get()
	}

	if src.SuccessThreshold.IsSet() && src.SuccessThreshold.Get() != nil {
		probe.SuccessThreshold = *src.SuccessThreshold.Get()
	}

	if src.FailureThreshold.IsSet() && src.FailureThreshold.Get() != nil {
		probe.FailureThreshold = *src.FailureThreshold.Get()
	}

	*targetPtr = probe

	return nil
}

// NodePoolsToK8s mirrors the behaviour of the cluster-side admission controller
// (see runai-cluster .../pkg/mutators/pod-mutators/project_node_pools_mutator.go
// and .../pkg/patch/affinity.go): for each requested node pool we add a
// match-expression to the node affinity. The default node pool is expressed as
// "label DoesNotExist"; any other node pool is expressed as the pool's
// labelKey In [labelValue]. Pools in the Deleting phase are silently skipped,
// while unknown pools result in an error.
func NodePoolsToK8s(src *NodePoolsInfo, target any, _ any) *workloadsApi.Error {
	if src == nil || len(src.RequestedNodePools) == 0 {
		return nil
	}

	targetPtr, ok := target.(**corev1.NodeAffinity)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("NodePoolsToK8s: target type %T is not **corev1.NodeAffinity", target))
	}

	nodePoolsByName := make(map[string]clusterApi.Nodepool, len(src.ActualNodePools))
	for _, np := range src.ActualNodePools {
		nodePoolsByName[np.GetName()] = np
	}

	matchExpressionsToAdd := make([]corev1.NodeSelectorRequirement, 0, len(src.RequestedNodePools))
	for _, nodePoolName := range src.RequestedNodePools {
		expr, skip, err := matchExpressionForNodePool(nodePoolName, nodePoolsByName)
		if err != nil {
			return err
		}
		if skip {
			continue
		}
		matchExpressionsToAdd = append(matchExpressionsToAdd, expr)
	}

	if len(matchExpressionsToAdd) == 0 {
		return nil
	}

	if *targetPtr == nil {
		*targetPtr = &corev1.NodeAffinity{}
	}
	addNodePoolMatchExpressionsToNodeAffinity(*targetPtr, matchExpressionsToAdd)

	return nil
}

// matchExpressionForNodePool returns the node-selector requirement for a single
// requested node pool. It returns skip=true when the pool exists but should be
// silently ignored (currently only when its phase is Deleting), and an error
// when the pool name is unknown.
func matchExpressionForNodePool(name string, nodePoolsByName map[string]clusterApi.Nodepool) (corev1.NodeSelectorRequirement, bool, *workloadsApi.Error) {
	if name == DefaultNodePoolName {
		return corev1.NodeSelectorRequirement{
			Key:      NodePoolNameLabel,
			Operator: corev1.NodeSelectorOpDoesNotExist,
		}, false, nil
	}

	np, ok := nodePoolsByName[name]
	if !ok {
		return corev1.NodeSelectorRequirement{}, false,
			workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("node pool %q does not exist", name))
	}

	if np.GetPhase() == clusterApi.NODEPOOLPHASE_DELETING {
		return corev1.NodeSelectorRequirement{}, true, nil
	}

	return corev1.NodeSelectorRequirement{
		Key:      np.GetLabelKey(),
		Operator: corev1.NodeSelectorOpIn,
		Values:   []string{np.GetLabelValue()},
	}, false, nil
}

// addNodePoolMatchExpressionsToNodeAffinity mimics the cluster's
// patch.AddNodeAffinityToPodSpec: requirements are AND-ed with the existing
// node affinity, and OR-ed amongst themselves. When existing terms are
// present, each new requirement replaces any existing requirement on the same
// key, expanding the term-set so that the cartesian product is preserved.
func addNodePoolMatchExpressionsToNodeAffinity(target *corev1.NodeAffinity, matchExpressionsToAdd []corev1.NodeSelectorRequirement) {
	if target.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		target.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{}
	}
	selector := target.RequiredDuringSchedulingIgnoredDuringExecution

	if len(selector.NodeSelectorTerms) == 0 {
		selector.NodeSelectorTerms = make([]corev1.NodeSelectorTerm, 0, len(matchExpressionsToAdd))
		for _, expr := range matchExpressionsToAdd {
			selector.NodeSelectorTerms = append(selector.NodeSelectorTerms, corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{expr},
			})
		}
		return
	}

	updatedTerms := make([]corev1.NodeSelectorTerm, 0, len(matchExpressionsToAdd)*len(selector.NodeSelectorTerms))
	for _, exprToAdd := range matchExpressionsToAdd {
		for _, term := range selector.NodeSelectorTerms {
			updatedExpressions := make([]corev1.NodeSelectorRequirement, 0, len(term.MatchExpressions)+1)
			for _, expr := range term.MatchExpressions {
				if expr.Key != exprToAdd.Key {
					updatedExpressions = append(updatedExpressions, expr)
				}
			}
			updatedExpressions = append(updatedExpressions, exprToAdd)
			term.MatchExpressions = updatedExpressions
			updatedTerms = append(updatedTerms, term)
		}
	}
	selector.NodeSelectorTerms = updatedTerms
}

// PreemptibilityToK8sLabel sets the preemptibility label on pod metadata labels (string setter)
func PreemptibilityToK8sLabel(src *string, target any, _ any) *workloadsApi.Error {
	if src == nil || *src == "" {
		return nil
	}

	targetPtr, ok := target.(*map[string]string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("PreemptibilityToK8sLabel: target type %T is not *map[string]string", target))
	}

	if *targetPtr == nil {
		*targetPtr = make(map[string]string)
	}

	(*targetPtr)[LabelPreemptibilityName] = *src

	return nil
}

// PreemptibleToK8sLabel sets the preemptibility label based on boolean preemptible flag (boolean setter)
// This is used for backward compatibility with old priority classes where preemptible=true means "preemptible"
func PreemptibleToK8sLabel(src *bool, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*map[string]string)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("PreemptibleToK8sLabel: target type %T is not *map[string]string", target))
	}

	if *targetPtr == nil {
		*targetPtr = make(map[string]string)
	}

	if *src {
		(*targetPtr)[LabelPreemptibilityName] = "preemptible"
	} else {
		(*targetPtr)[LabelPreemptibilityName] = "non-preemptible"
	}

	return nil
}

// MapDataVolumesToK8s converts data volume instances to PVC volumes.
// The cookie parameter should contain a map[string]string mapping data volume IDs to PVC claim names.
// This mimics the cluster behavior where data volumes are resolved to their underlying PVCs.
func MapDataVolumesToK8s(src map[string]*workloadsApi.DataVolumeInstance, target any, cookie any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(*VolumesAndMounts)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("MapDataVolumesToK8s: target type %T is not *VolumesAndMounts", target))
	}

	claimNameMap, ok := cookie.(map[string]string)
	if !ok || claimNameMap == nil {
		return workloadsApi.NewError(http.StatusInternalServerError, "MapDataVolumesToK8s: cookie must be a map[string]string containing data volume ID to claim name mapping")
	}

	for _, key := range SortedMapKeys(src) {
		instance := src[key]
		if instance == nil {
			continue
		}

		dataVolumeId := convutils.Deref(instance.Id.Get(), "")
		if dataVolumeId == "" {
			continue
		}

		claimName, found := claimNameMap[dataVolumeId]
		if !found || claimName == "" {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("dataVolume: claim name not found for data volume ID %s", dataVolumeId))
		}

		mountPath := convutils.Deref(instance.MountPath.Get(), "")

		volume := corev1.Volume{
			Name: key,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: claimName,
				},
			},
		}
		targetPtr.Volume = append(targetPtr.Volume, volume)

		volumeMount := corev1.VolumeMount{
			Name:      key,
			MountPath: mountPath,
			ReadOnly:  true,
		}
		targetPtr.VolumeMount = append(targetPtr.VolumeMount, volumeMount)
	}

	return nil
}

// StringToSeccompProfile allocates a SeccompProfile and sets its Type from src.
func StringToSeccompProfile(src *string, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(**corev1.SeccompProfile)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("StringToSeccompProfile: unsupported target type %T", target))
	}

	*targetPtr = &corev1.SeccompProfile{
		Type: corev1.SeccompProfileType(*src),
	}
	return nil
}
