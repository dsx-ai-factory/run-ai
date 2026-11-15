/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"strconv"

	dynamov1alpha1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

const (
	deploymentGpuRequestWorkerWithIngress = 3
)

var (
	DGDExpectedEwWithIngress = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			Type:              "Dynamo",
			PriorityClassName: "very-high",
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(3)),
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(2*deploymentGpuRequestWorkerWithIngress, resource.DecimalSI),
				},
			},
			Urls:             []string{"https://llm-api.example.com"},
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

	DGDWithIngress = getSecureLlmDGD()
)

func getSecureLlmDGD() dynamov1alpha1.DynamoGraphDeployment {

	return dynamov1alpha1.DynamoGraphDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "nvidia.com/v1alpha1",
			Kind:       "DynamoGraphDeployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "secure-llm",
			Namespace:  "runai-proj001",
			UID:        types.UID("cb0da452-2233-477d-923b-7505347f2e93"),
			Finalizers: []string{"nvidia.com/finalizer"},
			Annotations: map[string]string{
				//nolint:lll // test data JSON annotation
				"kubectl.kubernetes.io/last-applied-configuration": `{"apiVersion":"nvidia.com/v1alpha1","kind":"DynamoGraphDeployment","metadata":{"annotations":{},"name":"secure-llm","namespace":"runai-proj001"},"spec":{"services":{"Frontend":{"componentType":"frontend","dynamoNamespace":"secure-llm","extraPodSpec":{"mainContainer":{"image":"nvcr.io/nvidia/ai-dynamo/vllm-runtime:0.5.0"}},"ingress":{"annotations":{"cert-manager.io/cluster-issuer":"letsencrypt-prod","nginx.ingress.kubernetes.io/ssl-redirect":"true"},"enabled":true,"host":"api","hostPrefix":"llm-","hostSuffix":"example.com","ingressControllerClassName":"nginx","tls":{"secretName":"my-tls-secret"}},"replicas":1},"VllmWorker":{"componentType":"worker","dynamoNamespace":"secure-llm","envFromSecret":"hf-token-secret","extraPodSpec":{"mainContainer":{"args":["--model","Qwen/Qwen3-0.6B"],"command":["python3","-m","dynamo.vllm"],"image":"nvcr.io/nvidia/ai-dynamo/vllm-runtime:0.5.0","workingDir":"/workspace/components/backends/vllm"}},"replicas":2,"resources":{"limits":{"gpu":"3"}}}}}}`,
			},
		},
		Spec: dynamov1alpha1.DynamoGraphDeploymentSpec{
			Services: map[string]*dynamov1alpha1.DynamoComponentDeploymentSharedSpec{
				"Frontend": {
					ComponentType:   "frontend",
					DynamoNamespace: ptr.To("secure-llm"),
					Replicas:        ptr.To(int32(1)),
					ExtraPodSpec: &dynamov1alpha1.ExtraPodSpec{
						MainContainer: &corev1.Container{
							Image:     "nvcr.io/nvidia/ai-dynamo/vllm-runtime:0.5.0",
							Name:      "",
							Resources: corev1.ResourceRequirements{},
						},
					},
					Ingress: &dynamov1alpha1.IngressSpec{
						Enabled:                    true,
						Host:                       "api",
						HostPrefix:                 ptr.To("llm-"),
						HostSuffix:                 ptr.To("example.com"),
						IngressControllerClassName: ptr.To("nginx"),
						Annotations: map[string]string{
							"cert-manager.io/cluster-issuer":           "letsencrypt-prod",
							"nginx.ingress.kubernetes.io/ssl-redirect": "true",
						},
						TLS: &dynamov1alpha1.IngressTLSSpec{
							SecretName: "my-tls-secret",
						},
					},
				},
				"VllmWorker": {
					ComponentType:   "worker",
					DynamoNamespace: ptr.To("secure-llm"),
					Replicas:        ptr.To(int32(2)),
					EnvFromSecret:   ptr.To("hf-token-secret"),
					Resources: &dynamov1alpha1.Resources{
						Limits: &dynamov1alpha1.ResourceItem{
							GPU: strconv.Itoa(deploymentGpuRequestWorkerWithIngress),
						},
					},
					ExtraPodSpec: &dynamov1alpha1.ExtraPodSpec{
						MainContainer: &corev1.Container{
							Image:      "nvcr.io/nvidia/ai-dynamo/vllm-runtime:0.5.0",
							Name:       "",
							WorkingDir: "/workspace/components/backends/vllm",
							Command:    []string{"python3", "-m", "dynamo.vllm"},
							Args:       []string{"--model", "Qwen/Qwen3-0.6B"},
							Resources:  corev1.ResourceRequirements{},
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
}
