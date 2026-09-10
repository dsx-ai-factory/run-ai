/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"strconv"

	dynamov1alpha1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1alpha1"
	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	deploymentGpuRequest = 5
)

const (
	topologyName = "my-topology"
	// Legacy KAI-annotation path, independent of the pack domains below.
	topologyPreferredLevel = "host"
	// Deployment-level pack domain, and the narrower one a component overrides it with.
	topologyPackDomain          = "zone"
	topologyComponentPackDomain = "rack"
	topologyComponentName       = "Worker"
)

var (
	lastTransitionTime = metav1.Now()

	DynamoGraphDeploymentBasicExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			Type:              "Dynamo",
			PriorityClassName: "very-high",
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(2)),
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(deploymentGpuRequest, resource.DecimalSI),
				},
			},
			AdditionalFields: nil,
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				Message:   "All resources are ready",
				K8sStatus: "successful",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:               "Ready",
						Status:             "True",
						Reason:             string(constants.AllResourcesAreReady),
						Message:            "All resources are ready",
						LastTransitionTime: lastTransitionTime,
					},
				},
			},
		},
	}

	DynamoGraphDeploymentBasic = &dynamov1alpha1.DynamoGraphDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DynamoGraphDeployment",
			APIVersion: "nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:       "adb899d5-6f9d-4bf3-8257-59cbf9db671a",
			Name:      "vllm-agg",
			Namespace: RunAiTestNamespace,
			Annotations: map[string]string{
				"kubectl.kubernetes.io/last-applied-configuration": `{"apiVersion":"nvidia.com/v1alpha1","kind":"DynamoGraphDeployment","metadata":
{"annotations":{},"name":"vllm-agg","namespace":"runai-proj001"},"spec":{"services":{"Frontend":
{"componentType":"frontend","dynamoNamespace":"vllm-agg","extraPodSpec":{"mainContainer":
{"image":"nvcr.io/nvidia/ai-dynamo/vllm-runtime:0.5.0"}},"replicas":1},"VllmDecodeWorker":
{"componentType":"worker","dynamoNamespace":"vllm-agg","envFromSecret":"hf-token-secret","extraPodSpec":
{"mainContainer":{"args":["--model","Qwen/Qwen3-0.6B"],"command":["python3","-m","dynamo.vllm"],"image":"nvcr.io/nvidia/ai-dynamo/vllm-runtime:0.5.0",
"workingDir":"/workspace/components/backends/vllm"}},"replicas":1,"resources":{"limits":{"gpu":"0"}}}}}}`,
			},
			Finalizers:        []string{"nvidia.com/finalizer"},
			Generation:        2,
			CreationTimestamp: metav1.Now(),
		},
		Spec: dynamov1alpha1.DynamoGraphDeploymentSpec{
			Services: map[string]*dynamov1alpha1.DynamoComponentDeploymentSharedSpec{
				"Frontend": {
					ComponentType:   "frontend",
					DynamoNamespace: ptr.To("vllm-agg"),
					Replicas:        ptr.To(int32(1)),
				},
				"VllmDecodeWorker": {
					ComponentType:   "worker",
					DynamoNamespace: ptr.To("vllm-agg"),
					EnvFromSecret:   ptr.To("hf-token-secret"),
					Replicas:        ptr.To(int32(1)),
					Resources: &dynamov1alpha1.Resources{
						Limits: &dynamov1alpha1.ResourceItem{
							GPU: strconv.Itoa(deploymentGpuRequest),
						},
					},
				},
			},
		},
		Status: dynamov1alpha1.DynamoGraphDeploymentStatus{
			State: "successful",
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					Reason:             string(constants.AllResourcesAreReady),
					Message:            "All resources are ready",
					LastTransitionTime: lastTransitionTime,
				},
			},
		},
	}

	DGDWithTopologyInServiceSpec = &dynamov1alpha1.DynamoGraphDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DynamoGraphDeployment",
			APIVersion: "nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "dgd-topology-UID",
			Name:              "dgd-topology",
			Namespace:         RunAiTestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: dynamov1alpha1.DynamoGraphDeploymentSpec{
			Services: map[string]*dynamov1alpha1.DynamoComponentDeploymentSharedSpec{
				"Worker": {
					ComponentType: "worker",
					Replicas:      ptr.To(int32(1)),
					Annotations: map[string]string{
						clustercommonconstants.TopologyAnnotation:                   topologyName,
						clustercommonconstants.TopologyPreferredPlacementAnnotation: topologyPreferredLevel,
					},
				},
			},
		},
		Status: dynamov1alpha1.DynamoGraphDeploymentStatus{
			State: "successful",
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					Reason:             string(constants.AllResourcesAreReady),
					Message:            "All resources are ready",
					LastTransitionTime: lastTransitionTime,
				},
			},
		},
	}

	DGDWithTopologyInServiceSpecExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			Topology: v1alpha1.Topology{
				Name:               topologyName,
				PreferredPlacement: topologyPreferredLevel,
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				Message:   "All resources are ready",
				K8sStatus: "successful",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:               "Ready",
						Status:             "True",
						Reason:             string(constants.AllResourcesAreReady),
						Message:            "All resources are ready",
						LastTransitionTime: lastTransitionTime,
					},
				},
			},
		},
	}

	// DGDWithTopologyConstraint uses the Dynamo 1.1.0+ native deployment-level
	// spec.topologyConstraint. Only topologyProfile (the topology name) is surfaced
	// to the ExternalWorkload; packDomain is intentionally not mapped.
	DGDWithTopologyConstraint = &dynamov1alpha1.DynamoGraphDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DynamoGraphDeployment",
			APIVersion: "nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "dgd-topology-constraint-UID",
			Name:              "dgd-topology-constraint",
			Namespace:         RunAiTestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: dynamov1alpha1.DynamoGraphDeploymentSpec{
			TopologyConstraint: &dynamov1alpha1.SpecTopologyConstraint{
				TopologyProfile: topologyName,
				PackDomain:      dynamov1alpha1.TopologyDomain(topologyPackDomain),
			},
			Services: map[string]*dynamov1alpha1.DynamoComponentDeploymentSharedSpec{
				topologyComponentName: {
					ComponentType: "worker",
					Replicas:      ptr.To(int32(1)),
					TopologyConstraint: &dynamov1alpha1.TopologyConstraint{
						PackDomain: dynamov1alpha1.TopologyDomain(topologyComponentPackDomain),
					},
				},
			},
		},
		Status: dynamov1alpha1.DynamoGraphDeploymentStatus{
			State: "successful",
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					Reason:             string(constants.AllResourcesAreReady),
					Message:            "All resources are ready",
					LastTransitionTime: lastTransitionTime,
				},
			},
		},
	}

	DGDWithTopologyConstraintExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			Topology: v1alpha1.Topology{
				Name:              topologyName,
				RequiredPlacement: topologyPackDomain,
			},
			ComponentTopologies: []v1alpha1.ComponentTopologyEntry{
				{
					SpecName: topologyComponentName,
					Topology: v1alpha1.Topology{
						Name:              topologyName,
						RequiredPlacement: topologyComponentPackDomain,
					},
				},
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				Message:   "All resources are ready",
				K8sStatus: "successful",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:               "Ready",
						Status:             "True",
						Reason:             string(constants.AllResourcesAreReady),
						Message:            "All resources are ready",
						LastTransitionTime: lastTransitionTime,
					},
				},
			},
		},
	}
)
