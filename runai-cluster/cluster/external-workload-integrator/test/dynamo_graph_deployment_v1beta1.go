package test

import (
	dynamov1beta1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1beta1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const dgdV1beta1Image = "nvcr.io/nvidia/ai-dynamo/vllm-runtime:0.5.0"

// DynamoGraphDeploymentV1beta1Basic is a v1beta1 DGD served as the preferred
// nvidia.com/v1beta1 version: spec.components is a list, each component uses a
// standard pod template, and the topology name lives under
// spec.topologyConstraint.clusterTopologyName. Frontend (1 replica) + worker
// (1 replica, 5 GPUs) => 2 pods total, 5 GPU limits.
var (
	DynamoGraphDeploymentV1beta1Basic = &dynamov1beta1.DynamoGraphDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DynamoGraphDeployment",
			APIVersion: "nvidia.com/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "b1e7a0d2-1c3f-4a6b-9d8e-2f5c7a9b0d11",
			Name:              "vllm-agg-v1beta1",
			Namespace:         TestNamespace,
			Finalizers:        []string{"nvidia.com/finalizer"},
			Generation:        2,
			CreationTimestamp: metav1.Now(),
		},
		Spec: dynamov1beta1.DynamoGraphDeploymentSpec{
			TopologyConstraint: &dynamov1beta1.SpecTopologyConstraint{
				ClusterTopologyName: "grove-topology",
				PackDomain:          dynamov1beta1.TopologyDomain("rack"),
			},
			Components: []dynamov1beta1.DynamoComponentDeploymentSharedSpec{
				{
					ComponentName: "Frontend",
					ComponentType: dynamov1beta1.ComponentTypeFrontend,
					Replicas:      ptr.To(int32(1)),
					PodTemplate: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "main", Image: dgdV1beta1Image},
							},
						},
					},
				},
				{
					ComponentName: "VllmWorker",
					ComponentType: dynamov1beta1.ComponentTypeWorker,
					Replicas:      ptr.To(int32(1)),
					PodTemplate: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "main",
									Image: dgdV1beta1Image,
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											constants.GpuResourceName: resource.MustParse("5"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		// Status intentionally left empty: status-to-phase mapping is covered by the
		// handler unit tests. Asserting conditions here would compare metav1.Time
		// values that differ by timezone object after the fake-client round-trip.
	}

	DynamoGraphDeploymentV1beta1BasicExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(2)),
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: resource.MustParse("5"),
				},
			},
			Topology: v1alpha1.Topology{
				Name: "grove-topology",
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				// No state/conditions on the source DGD => Creating phase.
				Phase: v1alpha1.ExWCreatingPhase,
			},
		},
	}
)
