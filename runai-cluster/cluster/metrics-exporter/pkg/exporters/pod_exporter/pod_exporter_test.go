/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kubeaischedv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

const (
	podName                 = "my-pod"
	podNamespace            = "my-ns"
	numberOfFractionDevices = "3"
)

func TestGetName(t *testing.T) {
	pod := createPod()
	exporter := NewPodExporter(context.TODO(), nil, pod)
	assert.Equal(t, podName, exporter.GetName())
}

func TestGetNamespace(t *testing.T) {
	pod := createPod()
	exporter := NewPodExporter(context.TODO(), nil, pod)
	assert.Equal(t, podNamespace, exporter.GetNamespace())
}

func TestGetNodeName(t *testing.T) {
	expectedNodeName := "node-1"

	pod := createPod()
	pod.Spec.NodeName = expectedNodeName
	exporter := NewPodExporter(context.TODO(), nil, pod)
	assert.Equal(t, expectedNodeName, exporter.GetNodeName())
}

func TestGetUUID(t *testing.T) {
	expectedUUID := "uid"

	pod := createPod()
	pod.UID = types.UID(expectedUUID)
	exporter := NewPodExporter(context.TODO(), nil, pod)
	assert.Equal(t, expectedUUID, exporter.GetUUID())
}

func TestGetAllocatedGpusPodNotActive(t *testing.T) {
	pod := createPod()
	exporter := NewPodExporter(context.TODO(), nil, pod)

	allocatedGpus := exporter.GetAllocatedGpus()
	assert.Equal(t, float64(0), allocatedGpus)
}

func TestGetAllocatedGpusActivePod(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Annotations[podAllocatedGPUsAnnotationName] = "1"
	pod.Annotations[podAllocatedMigGPUsAnnotationName] = "0"

	exporter := NewPodExporter(context.TODO(), nil, pod)
	allocatedGpus := exporter.GetAllocatedGpus()
	assert.Equal(t, float64(1), allocatedGpus)
}

func TestGetAllocatedGpusNotFound(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning

	exporter := NewPodExporter(context.TODO(), nil, pod)
	allocatedGpus := exporter.GetAllocatedGpus()
	assert.Equal(t, float64(0), allocatedGpus)
}

func TestGetAllocatedGpusNotValid(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Annotations[podAllocatedGPUsAnnotationName] = "1"
	pod.Annotations[podAllocatedMigGPUsAnnotationName] = "invalidValue"

	exporter := NewPodExporter(context.TODO(), nil, pod)
	allocatedGpus := exporter.GetAllocatedGpus()
	assert.Equal(t, float64(0), allocatedGpus)
}

func TestGetAllocatedCpusPodNotActive(t *testing.T) {
	pod := createPod()

	exporter := NewPodExporter(context.TODO(), nil, pod)
	allocatedCpus := exporter.GetAllocatedCpus()
	assert.Equal(t, float64(0), allocatedCpus)
}

func TestGetAllocatedCpusActivePod(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Spec.Containers = []v1.Container{
		{
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					v1.ResourceCPU: resource.MustParse("1"),
				},
			},
		},
		{
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					v1.ResourceCPU: resource.MustParse("1"),
				},
			},
		},
	}

	exporter := NewPodExporter(context.TODO(), nil, pod)
	allocatedCpus := exporter.GetAllocatedCpus()
	assert.Equal(t, float64(2*1000), allocatedCpus)
}

func TestGetAllocatedMemoryPodNotActive(t *testing.T) {
	pod := createPod()
	exporter := NewPodExporter(context.TODO(), nil, pod)
	allocatedMemory := exporter.GetAllocatedMemory()
	assert.Equal(t, float64(0), allocatedMemory)
}

func TestGetAllocatedMemoryActivePod(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Spec.Containers = []v1.Container{
		{
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					v1.ResourceMemory: resource.MustParse("250"),
				},
			},
		},
		{
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					v1.ResourceMemory: resource.MustParse("250"),
				},
			},
		},
	}

	exporter := NewPodExporter(context.TODO(), nil, pod)
	allocatedMemory := exporter.GetAllocatedMemory()
	assert.Equal(t, float64(500), allocatedMemory)
}

func TestGetGpuIndexPodNotBound(t *testing.T) {
	pod := createPod()
	exporter := NewPodExporter(context.TODO(), nil, pod)
	gpuIndex, err := exporter.GetGpuIndex()
	assert.Nil(t, err)
	assert.Equal(t, "", gpuIndex)
}

func TestGetGpuIndexMigPod(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Annotations[receivedResourceTypeAnnotationName] = clustercommonconstants.ReceivedTypeMigInstance

	migDeviceInfo, err := addDeviceInfoForPod(pod)
	assert.Nil(t, err)
	expectedGpuIndex := migDeviceInfo.GPUIndex

	exporter := NewPodExporter(context.TODO(), nil, pod)
	gpuIndex, err := exporter.GetGpuIndex()
	assert.Nil(t, err)
	assert.Equal(t, expectedGpuIndex, gpuIndex)
}

func TestGetGpuIndexMigPodDeviceInfoError(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Annotations[receivedResourceTypeAnnotationName] = clustercommonconstants.ReceivedTypeMigInstance

	exporter := NewPodExporter(context.TODO(), nil, pod)
	gpuIndex, err := exporter.GetGpuIndex()
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("could not find %s annotation", allocatedMigDeviceAnnotationName), err.Error())
	assert.Equal(t, "", gpuIndex)
}

func TestGetGpuIndexNoGpuIndex(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning

	exporter := NewPodExporter(context.TODO(), nil, pod)
	gpuIndex, err := exporter.GetGpuIndex()
	assert.Nil(t, err)
	assert.Equal(t, "", gpuIndex)
}

func TestGetGpuIndexNoGpuGroup(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning

	exporter := NewPodExporter(context.TODO(), nil, pod)
	gpuIndex, err := exporter.GetGpuIndex()
	assert.Nil(t, err)
	assert.Equal(t, "", gpuIndex)
}

func TestGetGpuIndexOpportunisticPod(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Labels[clustercommonconstants.GPUGroup] = "gpu-group"
	pod.Annotations[runaiOpportunisticAnnotationName] = "opportunistic"

	exporter := NewPodExporter(context.TODO(), nil, pod)
	gpuIndex, err := exporter.GetGpuIndex()
	assert.Nil(t, err)
	assert.Equal(t, opportunisticPodGpuIndex, gpuIndex)
}

func TestGetGpuIndexFractionPod(t *testing.T) {
	expectedGPUIndex := "3"

	cm := createConfigMap(expectedGPUIndex)
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Labels[clustercommonconstants.GPUGroup] = "gpu-group"
	pod.Annotations[receivedResourceTypeAnnotationName] = clustercommonconstants.ReceivedTypeFraction
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{{
			Env: []v1.EnvVar{{
				Name: clustercommonconstants.NvidiaVisibleDevices,
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "my-configmap",
						},
						Key: clustercommonconstants.NvidiaVisibleDevices,
					},
				},
			}},
		}},
	}

	exporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{cm}), pod)
	gpuIndex, err := exporter.GetGpuIndex()
	assert.Nil(t, err)
	assert.Equal(t, expectedGPUIndex, gpuIndex)
}

func TestGetGpuIndexFractionPod_CDI(t *testing.T) {
	cdiDeviceName := "k8s.device-plugin.nvidia.com/gpu=GPU-e7be50b5-06a7-2924-4f04-854c45c50685"
	expectedGPUIndex := "GPU-e7be50b5-06a7-2924-4f04-854c45c50685"

	cm := createConfigMap(cdiDeviceName)
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Labels[clustercommonconstants.GPUGroup] = "gpu-group"
	pod.Annotations[receivedResourceTypeAnnotationName] = clustercommonconstants.ReceivedTypeFraction
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{{
			Env: []v1.EnvVar{{
				Name: clustercommonconstants.NvidiaVisibleDevices,
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "my-configmap",
						},
						Key: clustercommonconstants.NvidiaVisibleDevices,
					},
				},
			}},
		}},
	}

	exporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{cm}), pod)
	gpuIndex, err := exporter.GetGpuIndex()
	assert.Nil(t, err)
	assert.Equal(t, expectedGPUIndex, gpuIndex)
}

func TestGetGpuIndexMultiFractionPod_CDI(t *testing.T) {
	cdiDeviceNames := "k8s.device-plugin.nvidia.com/gpu=GPU-abc123,k8s.device-plugin.nvidia.com/gpu=GPU-def456"
	expectedGPUIndex := "GPU-abc123,GPU-def456"
	expectedNumDevices := "2"

	cm := createConfigMap(cdiDeviceNames)
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Labels[clustercommonconstants.GPUGroup] = "gpu-group"
	pod.Annotations[receivedResourceTypeAnnotationName] = clustercommonconstants.ReceivedTypeFraction
	pod.Annotations[numberOfDevicesAnnotationName] = "2"
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{{
			Env: []v1.EnvVar{{
				Name: clustercommonconstants.NvidiaVisibleDevices,
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "my-configmap",
						},
						Key: clustercommonconstants.NvidiaVisibleDevices,
					},
				},
			}},
		}},
	}

	exporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{cm}), pod)
	gpuIndex, err := exporter.GetGpuIndex()
	assert.Nil(t, err)
	assert.Equal(t, expectedGPUIndex, gpuIndex)
	numDevices, err := exporter.GetGpuNumDevices()
	assert.Nil(t, err)
	assert.Equal(t, expectedNumDevices, numDevices)
}

func TestStripCDIDevicePrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain numeric index", "3", "3"},
		{"plain UUID", "GPU-e7be50b5-06a7-2924-4f04-854c45c50685", "GPU-e7be50b5-06a7-2924-4f04-854c45c50685"},
		{"CDI device name", "k8s.device-plugin.nvidia.com/gpu=GPU-e7be50b5-06a7-2924-4f04-854c45c50685", "GPU-e7be50b5-06a7-2924-4f04-854c45c50685"},
		{"multi CDI", "k8s.device-plugin.nvidia.com/gpu=GPU-abc,k8s.device-plugin.nvidia.com/gpu=GPU-def", "GPU-abc,GPU-def"},
		{"multi plain", "3,5,6", "3,5,6"},
		{"empty", "", ""},
		{"all", "all", "all"},
		{"unknown", "unknown", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, stripCDIDevicePrefix(tt.input))
		})
	}
}

func TestGetGpuIndexMultiFractionPod(t *testing.T) {
	expectedGPUIndex := "3,5,6"
	expectedNumDevices := "3"

	cm := createConfigMap(expectedGPUIndex)
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Labels[clustercommonconstants.GPUGroup] = "gpu-group"
	pod.Annotations[receivedResourceTypeAnnotationName] = clustercommonconstants.ReceivedTypeFraction
	pod.Annotations[numberOfDevicesAnnotationName] = numberOfFractionDevices
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{{
			Env: []v1.EnvVar{{
				Name: clustercommonconstants.NvidiaVisibleDevices,
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "my-configmap",
						},
						Key: clustercommonconstants.NvidiaVisibleDevices,
					},
				},
			}},
		}},
	}

	exporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{cm}), pod)
	gpuIndex, err := exporter.GetGpuIndex()
	assert.Nil(t, err)
	assert.Equal(t, expectedGPUIndex, gpuIndex)
	numDevices, err := exporter.GetGpuNumDevices()
	assert.Nil(t, err)
	assert.Equal(t, expectedNumDevices, numDevices)
}

func TestGetGpuIndexFractionPod_withNodeScheduler(t *testing.T) {
	expectedGPUIndex := "all"

	cm := createConfigMap(expectedGPUIndex)
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Labels[clustercommonconstants.GPUGroup] = "gpu-group"
	pod.Annotations[receivedResourceTypeAnnotationName] = clustercommonconstants.ReceivedTypeFraction
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{{
			Env: []v1.EnvVar{{
				Name: clustercommonconstants.NvidiaVisibleDevices,
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "my-configmap",
						},
						Key: clustercommonconstants.NvidiaVisibleDevices,
					},
				},
			}},
		}},
	}

	exporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{cm}), pod)
	gpuIndex, err := exporter.GetGpuIndex()
	assert.Nil(t, err)
	assert.Equal(t, unknownDevice, gpuIndex)
}

func TestGetGpuIndexWholeGpuPod_withNodeScheduler(t *testing.T) {
	expectedGPUIndex := "all"

	cm := createConfigMap(expectedGPUIndex)
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{{
			EnvFrom: []v1.EnvFromSource{{
				ConfigMapRef: &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "my-configmap",
					},
				},
			}},
		}},
	}

	exporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{cm}), pod)
	gpuIndex, err := exporter.GetGpuIndex()
	assert.Nil(t, err)
	assert.Equal(t, unknownDevice, gpuIndex)
}

func TestGetGpuInstanceIdNonMigPod(t *testing.T) {
	pod := createPod()
	exporter := NewPodExporter(context.TODO(), nil, pod)
	gpuInstanceId, err := exporter.GetGpuInstanceId()
	assert.Nil(t, err)
	assert.Equal(t, "", gpuInstanceId)
}

func TestGetGpuInstanceIdMigPod(t *testing.T) {
	pod := createPod()
	pod.Annotations[receivedResourceTypeAnnotationName] = clustercommonconstants.ReceivedTypeMigInstance
	migDeviceInfo, err := addDeviceInfoForPod(pod)
	assert.Nil(t, err)
	expectedGpuInstanceId := strconv.Itoa(migDeviceInfo.GpuInstanceId)

	exporter := NewPodExporter(context.TODO(), nil, pod)
	gpuInstanceId, err := exporter.GetGpuInstanceId()
	assert.Nil(t, err)
	assert.Equal(t, expectedGpuInstanceId, gpuInstanceId)
}

func TestGetGpuInstanceIdMigPodDeviceInfoError(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Annotations[receivedResourceTypeAnnotationName] = clustercommonconstants.ReceivedTypeMigInstance

	exporter := NewPodExporter(context.TODO(), nil, pod)
	gpuInstanceId, err := exporter.GetGpuInstanceId()
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("could not find %s annotation", allocatedMigDeviceAnnotationName), err.Error())
	assert.Equal(t, "", gpuInstanceId)
}

func TestGetPhase(t *testing.T) {
	expectedPhase := "pod-phase"
	pod := createPod()
	pod.Annotations[podStatusAnnotationName] = expectedPhase

	exporter := NewPodExporter(context.TODO(), nil, pod)
	phase, err := exporter.GetPhase()
	assert.Nil(t, err)
	assert.Equal(t, expectedPhase, phase)
}

func TestGetPhaseNotFound(t *testing.T) {
	pod := createPod()
	exporter := NewPodExporter(context.TODO(), nil, pod)
	phase, err := exporter.GetPhase()
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("could not find %s annotation", podStatusAnnotationName), err.Error())
	assert.Equal(t, "", phase)
}

func TestIsBoundTrue(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning

	exporter := NewPodExporter(context.TODO(), nil, pod)
	assert.True(t, exporter.IsBound())
}

func TestIsBoundFalse(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodPending

	exporter := NewPodExporter(context.TODO(), nil, pod)
	assert.False(t, exporter.IsBound())
}

func TestExtractGPUIndexFromPodEnvVar(t *testing.T) {
	pod := createPod()
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{{
			Env: []v1.EnvVar{{
				Name:  clustercommonconstants.NvidiaVisibleDevices,
				Value: "all",
			}},
		}},
	}

	exporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{}), pod)

	extractedGPUIndex, err := exporter.extractGPUIndex()
	assert.Nil(t, err)

	assert.Equal(t, unknownDevice, extractedGPUIndex)
}

func TestExtractGPUIndexFromPodEnvVarWithConfigMap(t *testing.T) {
	expectedGPUIndex := "3"

	cm := createConfigMap(expectedGPUIndex)
	pod := createPod()
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{{
			Env: []v1.EnvVar{{
				Name: clustercommonconstants.NvidiaVisibleDevices,
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "my-configmap",
						},
						Key: clustercommonconstants.NvidiaVisibleDevices,
					},
				},
			}},
		}},
	}

	exporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{cm}), pod)

	extractedGPUIndex, err := exporter.extractGPUIndex()
	assert.Nil(t, err)

	assert.Equal(t, expectedGPUIndex, extractedGPUIndex)
}

func TestExtractGPUIndexFromPodEnvVarNoValue(t *testing.T) {
	pod := createPod()
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{{
			Env: []v1.EnvVar{{
				Name: clustercommonconstants.NvidiaVisibleDevices,
			}},
		}},
	}

	exporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{}), pod)

	extractedGPUIndex, err := exporter.extractGPUIndex()
	assert.NotNil(t, err)
	assert.Equal(t, "", extractedGPUIndex)
	assert.Contains(t, err.Error(), "failed to read value")
}

func TestExtractGPUIndexFromPodEnvVarNoConfigMapRef(t *testing.T) {
	pod := createPod()
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{{
			Env: []v1.EnvVar{{
				Name:      clustercommonconstants.NvidiaVisibleDevices,
				ValueFrom: &v1.EnvVarSource{},
			}},
		}},
	}

	exporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{}), pod)

	extractedGPUIndex, err := exporter.extractGPUIndex()
	assert.NotNil(t, err)
	assert.Equal(t, "", extractedGPUIndex)
	assert.Contains(t, err.Error(), "failed to read config map reference")
}

func TestExtractGPUIndexFromPodEnvConfigMapNotFound(t *testing.T) {
	pod := createPod()
	pod.Spec = v1.PodSpec{
		Containers: []v1.Container{{
			Env: []v1.EnvVar{{
				Name: clustercommonconstants.NvidiaVisibleDevices,
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "my-configmap",
						},
						Key: clustercommonconstants.NvidiaVisibleDevices,
					},
				},
			}},
		}},
	}

	exporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{}), pod)

	extractedGPUIndex, err := exporter.extractGPUIndex()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to get config map")
	assert.Equal(t, "", extractedGPUIndex)
}

func TestGetMIGDeviceInfoNonMigPod(t *testing.T) {
	pod := createPod()

	exporter := NewPodExporter(context.TODO(), nil, pod)

	deviceInfo, err := exporter.getMIGDeviceInfo()
	assert.NotNil(t, err)
	assert.Equal(t, "asked for MIG Device info for non-mig pod", err.Error())
	assert.Nil(t, deviceInfo)
}

func TestGetMIGDeviceInfo(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Annotations[receivedResourceTypeAnnotationName] = clustercommonconstants.ReceivedTypeMigInstance
	migDeviceInfo, err := addDeviceInfoForPod(pod)
	assert.Nil(t, err)

	exporter := NewPodExporter(context.TODO(), nil, pod)
	assert.Nil(t, err)

	deviceInfo, err := exporter.getMIGDeviceInfo()
	assert.Nil(t, err)
	assert.Equal(t, migDeviceInfo.DeviceUUID, deviceInfo.DeviceUUID)
	assert.Equal(t, migDeviceInfo.Name, deviceInfo.Name)
	assert.Equal(t, migDeviceInfo.GPUIndex, deviceInfo.GPUIndex)
	assert.Equal(t, migDeviceInfo.GPUAmount, deviceInfo.GPUAmount)
	assert.Equal(t, migDeviceInfo.Position, deviceInfo.Position)
	assert.Equal(t, migDeviceInfo.GpuInstanceId, deviceInfo.GpuInstanceId)
	assert.Equal(t, migDeviceInfo.GPUProduct, deviceInfo.GPUProduct)
}

func TestGetMIGDeviceInfoFailedToUnmarshal(t *testing.T) {
	pod := createPod()
	pod.Status.Phase = v1.PodRunning
	pod.Annotations[receivedResourceTypeAnnotationName] = clustercommonconstants.ReceivedTypeMigInstance
	pod.Annotations[allocatedMigDeviceAnnotationName] =
		"{\"name\":\"1g.5gb\",\"position\":\"invalid-position\",\"gpuamount\":1," +
			"\"deviceUUID\":\"device-uuid\",\"gpuinstanceid\":9,\"gpuindex\":\"0\"}"

	exporter := NewPodExporter(context.TODO(), nil, pod)

	deviceInfo, err := exporter.getMIGDeviceInfo()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(),
		fmt.Sprintf("could not unmarshal %s annotation value", allocatedMigDeviceAnnotationName))
	assert.Nil(t, deviceInfo)
}

func TestGetGpusFromResourceRequest(t *testing.T) {
	pod := createPod()
	pod.Spec.Containers = []v1.Container{
		{
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					clustercommonconstants.NvidiaGpuResourceName: resource.MustParse("1"),
				},
			},
		},
		{
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					clustercommonconstants.NvidiaGpuResourceName: resource.MustParse("1"),
				},
			},
		},
	}
	podExporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{pod}), pod)
	gpus := podExporter.GetGpusFromResourceRequest()
	assert.Equal(t, float64(2), gpus)
}

func TestGetGpusFromResourceRequest_Zero(t *testing.T) {
	pod := createPod()
	pod.Spec.Containers = []v1.Container{
		{
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{},
			},
		},
		{
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{},
			},
		},
	}
	podExporter := NewPodExporter(context.TODO(), createFakeClient([]client.Object{pod}), pod)
	gpus := podExporter.GetGpusFromResourceRequest()
	assert.Equal(t, float64(0), gpus)
}

func createFakeClient(objects []client.Object) client.Client {
	scheme := runtime.NewScheme()
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(v2alpha2.AddToScheme(scheme))
	utilruntime.Must(kubeaischedv2alpha2.AddToScheme(scheme))
	utilruntime.Must(kaiv2.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func createPod() *v1.Pod {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        podName,
			Namespace:   podNamespace,
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		},
		Spec: v1.PodSpec{},
	}
	return &pod
}

func addDeviceInfoForPod(pod *v1.Pod) (*utils.MigDeviceInfo, error) {
	deviceInfo := utils.MigDeviceInfo{
		Name:          "1g.5gb",
		Position:      6,
		GPUAmount:     1,
		DeviceUUID:    "device-uuid",
		GpuInstanceId: 9,
		GPUIndex:      "0",
		GPUProduct:    "",
	}
	deviceInfoJson, err := json.Marshal(deviceInfo)
	if err != nil {
		log.Printf("failed to marshal MigDeviceInfo to json %v, error: %s", deviceInfo, err.Error())
		return nil, err
	}

	pod.Annotations[allocatedMigDeviceAnnotationName] = string(deviceInfoJson)
	return &deviceInfo, nil
}

func createConfigMap(visibleDeviceIndex string) *v1.ConfigMap {
	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-configmap",
			Namespace: podNamespace,
		},
		Data: map[string]string{
			clustercommonconstants.NvidiaVisibleDevices: visibleDeviceIndex,
		},
	}
	return &cm
}
