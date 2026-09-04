/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	dynamov1alpha1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1alpha1"
	dynamov1beta1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1beta1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/test"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("TestDynamoGraphDeploymentHandler", Ordered, func() {
	var (
		externalWorkload            *v1alpha1.ExternalWorkload
		expectedDgdExternalWorkload *v1alpha1.ExternalWorkload
		cachedClient                client.WithWatch
		dynamoGraphDeployment       *dynamov1alpha1.DynamoGraphDeployment
		dgdfTopOwnerHandler         *DynamoGraphDeploymentTopOwnerHandler // dynamoGraphDeployment Top Owner Handler
		ctx                         context.Context
		cancel                      context.CancelFunc
	)

	BeforeAll(func() {
		scheme := scheme_init.Scheme()
		utilruntime.Must(dynamov1alpha1.AddToScheme(scheme))
		utilruntime.Must(dynamov1beta1.AddToScheme(scheme))
		ns := test.BuildNamespace(test.TestNamespace, map[string]string{})
		workloadConfigMap := test.CreateWorkloadsConfigMapObject()
		cachedClient = fake.NewClientBuilder().WithStatusSubresource(&corev1.Pod{}).WithScheme(scheme).WithObjects(ns, workloadConfigMap).Build()

		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(func() {
			cancel()
		})

		test.InitVars()

		dgdfTopOwnerHandler = NewDynamoGraphDeploymentTopOwnerHandler(cachedClient, nil,
			func() client.Object { return &dynamov1alpha1.DynamoGraphDeployment{} })
	})

	BeforeEach(func() {
		externalWorkload = &v1alpha1.ExternalWorkload{}
		dynamoGraphDeployment = test.DynamoGraphDeploymentBasic.DeepCopy()
		expectedDgdExternalWorkload = test.DynamoGraphDeploymentBasicExpectedEw.DeepCopy()
	})

	It("Validate Successful DGD", func() {
		// The fake client strips TypeMeta on Create; capture and restore the GVK
		// so Parse can resolve the kind for the priority-class lookup, as it would
		// in production where the object comes from a dynamic (unstructured) informer.
		gvk := dynamoGraphDeployment.GroupVersionKind()
		err := cachedClient.Create(ctx, dynamoGraphDeployment)
		Expect(err).To(BeNil())
		dynamoGraphDeployment.SetGroupVersionKind(gvk)

		err = dgdfTopOwnerHandler.Parse(ctx, dynamoGraphDeployment, externalWorkload)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedDgdExternalWorkload)

		err = cachedClient.Get(ctx, client.ObjectKeyFromObject(dynamoGraphDeployment), dynamoGraphDeployment)
		Expect(err).To(BeNil())
	})

	It("Validate Pending DGD", func() {
		lastTransitionTime := metav1.Now()
		name := fmt.Sprintf("%s-%d", dynamoGraphDeployment.Name, 1)
		dynamoGraphDeployment.Name = name
		message := "Resources not ready: secure-llm-vllmworker: Component deployment not\n      ready - Available condition not true"
		dynamoGraphDeployment.Status = dynamov1alpha1.DynamoGraphDeploymentStatus{
			State: "pending",
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionFalse,
					Reason:             string(constants.SomeResourcesNotReady),
					Message:            message,
					LastTransitionTime: lastTransitionTime,
				},
			},
		}
		expectedDgdExternalWorkload.Status = v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWPendingPhase,
				Message:   message,
				K8sStatus: "pending",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:               "Ready",
						Status:             "False",
						Reason:             string(constants.SomeResourcesNotReady),
						Message:            message,
						LastTransitionTime: lastTransitionTime,
					},
				},
			},
		}
		// The fake client strips TypeMeta on Create; capture and restore the GVK so
		// Parse can resolve the kind for the priority-class lookup (in production the
		// object comes from a dynamic informer that retains apiVersion/kind).
		gvk := dynamoGraphDeployment.GroupVersionKind()
		err := cachedClient.Create(ctx, dynamoGraphDeployment)
		Expect(err).To(BeNil())
		dynamoGraphDeployment.SetGroupVersionKind(gvk)

		err = dgdfTopOwnerHandler.Parse(ctx, dynamoGraphDeployment, externalWorkload)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedDgdExternalWorkload)

		err = cachedClient.Get(ctx, client.ObjectKeyFromObject(dynamoGraphDeployment), dynamoGraphDeployment)
		Expect(err).To(BeNil())
	})

	It("Validate Url DGD", func() {
		dynamoGraphDeployment = test.DGDWithIngress.DeepCopy()
		expectedDgdExternalWorkload = test.DGDExpectedEwWithIngress.DeepCopy()

		// The fake client strips TypeMeta on Create; capture and restore the GVK so
		// Parse can resolve the kind for the priority-class lookup.
		gvk := dynamoGraphDeployment.GroupVersionKind()
		err := cachedClient.Create(ctx, dynamoGraphDeployment)
		Expect(err).To(BeNil())
		dynamoGraphDeployment.SetGroupVersionKind(gvk)

		err = dgdfTopOwnerHandler.Parse(ctx, dynamoGraphDeployment, externalWorkload)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedDgdExternalWorkload)

		err = cachedClient.Get(ctx, client.ObjectKeyFromObject(dynamoGraphDeployment), dynamoGraphDeployment)
		Expect(err).To(BeNil())
	})

	It("Parse topology from Dynamo deployment-level topologyConstraint", func() {
		dgd := test.DGDWithTopologyConstraint.DeepCopy()
		expectedEw := test.DGDWithTopologyConstraintExpectedEw.DeepCopy()

		err := cachedClient.Create(ctx, dgd)
		Expect(err).To(BeNil())

		ew := &v1alpha1.ExternalWorkload{}
		err = dgdfTopOwnerHandler.Parse(ctx, dgd, ew)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, ew, expectedEw)

		// Only the topology name is surfaced; packDomain is intentionally not mapped.
		Expect(ew.Spec.Topology.PreferredPlacement).To(BeEmpty())
		Expect(ew.Spec.Topology.RequiredPlacement).To(BeEmpty())
	})

	It("Parse topology from service spec annotations", func() {
		dgd := test.DGDWithTopologyInServiceSpec.DeepCopy()
		expectedEw := test.DGDWithTopologyInServiceSpecExpectedEw.DeepCopy()

		err := cachedClient.Create(ctx, dgd)
		Expect(err).To(BeNil())

		ew := &v1alpha1.ExternalWorkload{}
		err = dgdfTopOwnerHandler.Parse(ctx, dgd, ew)
		Expect(err).To(BeNil())

		test.ValidateExternalWorkloadInnerSpec(Default, 1, ew, expectedEw)
	})

	It("DGD object-level topology annotations do not populate ExternalWorkload topology", func() {
		dgd := test.DynamoGraphDeploymentBasic.DeepCopy()
		dgd.Name = fmt.Sprintf("%s-no-svc-topology", dgd.Name)
		dgd.Annotations = map[string]string{
			"kai.scheduler/topology":                     "object-level-topology",
			"kai.scheduler/topology-preferred-placement": "some-rack",
		}

		err := cachedClient.Create(ctx, dgd)
		Expect(err).To(BeNil())

		ew := &v1alpha1.ExternalWorkload{}
		err = dgdfTopOwnerHandler.Parse(ctx, dgd, ew)
		Expect(err).To(BeNil())

		Expect(ew.Spec.Topology.Name).To(BeEmpty(), "topology should not be read from object-level annotations")
		Expect(ew.Spec.Topology.PreferredPlacement).To(BeEmpty())
		Expect(ew.Spec.Topology.RequiredPlacement).To(BeEmpty())
	})

	It("Parse v1beta1 DGD: topology, replicas and resources from components", func() {
		dgd := &dynamov1beta1.DynamoGraphDeployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DynamoGraphDeployment",
				APIVersion: "nvidia.com/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dgd-v1beta1",
				Namespace: test.TestNamespace,
			},
			Spec: dynamov1beta1.DynamoGraphDeploymentSpec{
				TopologyConstraint: &dynamov1beta1.SpecTopologyConstraint{
					ClusterTopologyName: "grove-topology",
					PackDomain:          dynamov1beta1.TopologyDomain("rack"),
				},
				Components: []dynamov1beta1.DynamoComponentDeploymentSharedSpec{
					{
						ComponentName: "VllmWorker",
						ComponentType: dynamov1beta1.ComponentTypeWorker,
						Replicas:      ptr.To(int32(2)),
						PodTemplate: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "main",
										Image: "nvcr.io/nvidia/ai-dynamo/vllm-runtime:0.5.0",
										Resources: corev1.ResourceRequirements{
											Requests: corev1.ResourceList{
												corev1.ResourceCPU:    resource.MustParse("2"),
												corev1.ResourceMemory: resource.MustParse("4Gi"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		err := cachedClient.Create(ctx, dgd)
		Expect(err).To(BeNil())

		ew := &v1alpha1.ExternalWorkload{}
		err = dgdfTopOwnerHandler.Parse(ctx, dgd, ew)
		Expect(err).To(BeNil())

		// Topology name comes from v1beta1 clusterTopologyName; packDomain not mapped.
		Expect(ew.Spec.Topology.Name).To(Equal("grove-topology"))
		Expect(ew.Spec.Topology.PreferredPlacement).To(BeEmpty())
		Expect(ew.Spec.Topology.RequiredPlacement).To(BeEmpty())

		// 2 replicas, image and resources aggregated from the component pod template.
		Expect(ew.Spec.RequestedPods.Replicas).ToNot(BeNil())
		Expect(*ew.Spec.RequestedPods.Replicas).To(Equal(int32(2)))
		Expect(ew.Spec.ImageName).To(Equal("nvcr.io/nvidia/ai-dynamo/vllm-runtime:0.5.0"))
		cpu := ew.Spec.RequestedResources.Requests[corev1.ResourceCPU]
		Expect(cpu.Value()).To(Equal(int64(4))) // 2 replicas * 2 CPU
	})

	It("Parse v1beta1 DGD: legacy KAI topology annotations on the deployment-level spec", func() {
		worker := newV1beta1Component("VllmWorker", dynamov1beta1.ComponentTypeWorker, 1, 1, dgdImage, nil, nil)
		// No topologyConstraint, so the deployment-level annotations are the source.
		dgd := newV1beta1DGD("dgd-v1beta1-legacy-topology", "", worker)
		dgd.Spec.Annotations = map[string]string{
			"kai.scheduler/topology":                     "legacy-topology",
			"kai.scheduler/topology-preferred-placement": "rack-1",
			"kai.scheduler/topology-required-placement":  "zone-a",
		}

		Expect(cachedClient.Create(ctx, dgd)).To(BeNil())
		ew := &v1alpha1.ExternalWorkload{}
		Expect(dgdfTopOwnerHandler.Parse(ctx, dgd, ew)).To(BeNil())

		Expect(ew.Spec.Topology.Name).To(Equal("legacy-topology"))
		Expect(ew.Spec.Topology.PreferredPlacement).To(Equal("rack-1"))
		Expect(ew.Spec.Topology.RequiredPlacement).To(Equal("zone-a"))
	})

	It("Parse v1beta1 DGD: clusterTopologyName takes precedence over legacy spec annotations", func() {
		worker := newV1beta1Component("VllmWorker", dynamov1beta1.ComponentTypeWorker, 1, 1, dgdImage, nil, nil)
		dgd := newV1beta1DGD("dgd-v1beta1-topology-precedence", "grove-topology", worker)
		dgd.Spec.Annotations = map[string]string{
			"kai.scheduler/topology":                     "legacy-topology",
			"kai.scheduler/topology-preferred-placement": "rack-1",
		}

		Expect(cachedClient.Create(ctx, dgd)).To(BeNil())
		ew := &v1alpha1.ExternalWorkload{}
		Expect(dgdfTopOwnerHandler.Parse(ctx, dgd, ew)).To(BeNil())

		// clusterTopologyName wins; the legacy annotations are not consulted.
		Expect(ew.Spec.Topology.Name).To(Equal("grove-topology"))
		Expect(ew.Spec.Topology.PreferredPlacement).To(BeEmpty())
		Expect(ew.Spec.Topology.RequiredPlacement).To(BeEmpty())
	})

	It("Parse v1beta1 DGD: aggregates replicas, requests and limits across components", func() {
		// frontend: 1 replica, single node, cpu/mem requests only.
		frontend := newV1beta1Component("Frontend", dynamov1beta1.ComponentTypeFrontend, 1, 1, dgdImage,
			corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1"), corev1.ResourceMemory: resource.MustParse("2Gi")},
			nil)
		// worker: 2 replicas x 2 nodes = 4 pods; cpu/mem requests + 2 GPUs per pod (limit).
		worker := newV1beta1Component("VllmWorker", dynamov1beta1.ComponentTypeWorker, 2, 2, dgdImage,
			corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("5"), corev1.ResourceMemory: resource.MustParse("8Gi")},
			corev1.ResourceList{constants.GpuResourceName: resource.MustParse("2")})
		dgd := newV1beta1DGD("dgd-v1beta1-multi", "grove-topology", frontend, worker)

		Expect(cachedClient.Create(ctx, dgd)).To(BeNil())
		ew := &v1alpha1.ExternalWorkload{}
		Expect(dgdfTopOwnerHandler.Parse(ctx, dgd, ew)).To(BeNil())

		// Pods: 1 (frontend) + 2*2 (worker) = 5.
		Expect(ew.Spec.RequestedPods.Replicas).ToNot(BeNil())
		Expect(*ew.Spec.RequestedPods.Replicas).To(Equal(int32(5)))

		// Requests sum across pods: cpu = 1*1 + 5*4 = 21; memory = 1*2Gi + 4*8Gi = 34Gi.
		reqCPU := ew.Spec.RequestedResources.Requests[corev1.ResourceCPU]
		Expect(reqCPU.Equal(resource.MustParse("21"))).To(BeTrue(), "cpu requests")
		reqMem := ew.Spec.RequestedResources.Requests[corev1.ResourceMemory]
		Expect(reqMem.Equal(resource.MustParse("34Gi"))).To(BeTrue(), "memory requests")

		// Limits sum across worker pods: gpu = 2*4 = 8.
		limGPU := ew.Spec.RequestedResources.Limits[constants.GpuResourceName]
		Expect(limGPU.Equal(resource.MustParse("8"))).To(BeTrue(), "gpu limits")

		Expect(ew.Spec.ImageName).To(Equal(dgdImage))
		Expect(ew.Spec.Topology.Name).To(Equal("grove-topology"))
	})

	It("Parse v1beta1 DGD: multinode component multiplies the pod count by node count", func() {
		// 3 replicas x 4 nodes = 12 pods.
		worker := newV1beta1Component("VllmWorker", dynamov1beta1.ComponentTypeWorker, 3, 4, dgdImage, nil, nil)
		dgd := newV1beta1DGD("dgd-v1beta1-multinode", "", worker)

		Expect(cachedClient.Create(ctx, dgd)).To(BeNil())
		ew := &v1alpha1.ExternalWorkload{}
		Expect(dgdfTopOwnerHandler.Parse(ctx, dgd, ew)).To(BeNil())

		Expect(ew.Spec.RequestedPods.Replicas).ToNot(BeNil())
		Expect(*ew.Spec.RequestedPods.Replicas).To(Equal(int32(12)))
	})

	It("Parse v1beta1 DGD: carries both requests and limits, scaled by replicas", func() {
		comp := newV1beta1Component("VllmWorker", dynamov1beta1.ComponentTypeWorker, 2, 1, dgdImage,
			corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2"), corev1.ResourceMemory: resource.MustParse("4Gi")},
			corev1.ResourceList{
				corev1.ResourceCPU:        resource.MustParse("4"),
				corev1.ResourceMemory:     resource.MustParse("8Gi"),
				constants.GpuResourceName: resource.MustParse("1"),
			})
		dgd := newV1beta1DGD("dgd-v1beta1-req-lim", "", comp)

		Expect(cachedClient.Create(ctx, dgd)).To(BeNil())
		ew := &v1alpha1.ExternalWorkload{}
		Expect(dgdfTopOwnerHandler.Parse(ctx, dgd, ew)).To(BeNil())

		Expect(*ew.Spec.RequestedPods.Replicas).To(Equal(int32(2)))

		reqCPU := ew.Spec.RequestedResources.Requests[corev1.ResourceCPU]
		Expect(reqCPU.Equal(resource.MustParse("4"))).To(BeTrue(), "cpu requests = 2*2")
		reqMem := ew.Spec.RequestedResources.Requests[corev1.ResourceMemory]
		Expect(reqMem.Equal(resource.MustParse("8Gi"))).To(BeTrue(), "memory requests = 2*4Gi")

		limCPU := ew.Spec.RequestedResources.Limits[corev1.ResourceCPU]
		Expect(limCPU.Equal(resource.MustParse("8"))).To(BeTrue(), "cpu limits = 2*4")
		limMem := ew.Spec.RequestedResources.Limits[corev1.ResourceMemory]
		Expect(limMem.Equal(resource.MustParse("16Gi"))).To(BeTrue(), "memory limits = 2*8Gi")
		limGPU := ew.Spec.RequestedResources.Limits[constants.GpuResourceName]
		Expect(limGPU.Equal(resource.MustParse("2"))).To(BeTrue(), "gpu limits = 2*1")
	})

	It("DRA GPU resources from ExtraPodSpec ResourceClaims", func() {
		draFakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(test.TestDRAGpuClaim.DeepCopy(), test.TestDRAGpuClaimTemplate.DeepCopy()).Build()
		fakeKubeClient := kubefake.NewClientset(test.TestDRAGpuClaim.DeepCopy(), test.TestDRAGpuClaimTemplate.DeepCopy())
		draParser := dra.NewDRAParser(draFakeClient, fakeKubeClient)
		draHandler := NewDynamoGraphDeploymentTopOwnerHandler(cachedClient, draParser,
			func() client.Object { return &dynamov1alpha1.DynamoGraphDeployment{} })

		dgd := &dynamov1alpha1.DynamoGraphDeployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DynamoGraphDeployment",
				APIVersion: "nvidia.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				UID:               "dgd-dra-UID",
				Name:              "dgd-dra",
				Namespace:         test.TestNamespace,
				CreationTimestamp: metav1.Now(),
			},
			Spec: dynamov1alpha1.DynamoGraphDeploymentSpec{
				Services: map[string]*dynamov1alpha1.DynamoComponentDeploymentSharedSpec{
					"Worker": {
						ComponentType:   "worker",
						DynamoNamespace: ptr.To("test-ns"),
						Replicas:        ptr.To(int32(2)),
						ExtraPodSpec: &dynamov1alpha1.ExtraPodSpec{
							PodSpec: &corev1.PodSpec{
								ResourceClaims: []corev1.PodResourceClaim{
									{
										Name:                      "gpu-claim",
										ResourceClaimTemplateName: ptr.To("dra-gpu-template"),
									},
								},
							},
							MainContainer: &corev1.Container{
								Image: "test-image:latest",
							},
						},
					},
				},
			},
			Status: dynamov1alpha1.DynamoGraphDeploymentStatus{
				State: "successful",
				Conditions: []metav1.Condition{
					{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
						Reason: string(constants.AllResourcesAreReady),
					},
				},
			},
		}

		expectedEw := &v1alpha1.ExternalWorkload{
			Spec: v1alpha1.ExternalWorkloadSpec{
				RequestedPods: v1alpha1.RequestedPods{
					Replicas: ptr.To(int32(2)),
				},
				RequestedResources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						constants.GpuResourceName: *resource.NewQuantity(8, resource.DecimalSI), // 2 replicas * 4 GPUs
					},
					Requests: corev1.ResourceList{
						constants.GpuResourceName: *resource.NewQuantity(8, resource.DecimalSI), // 2 replicas * 4 GPUs
					},
				},
				ImageName: "test-image:latest",
			},
			Status: v1alpha1.ExternalWorkloadStatus{
				ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
					Phase:     v1alpha1.ExWRunningPhase,
					K8sStatus: "successful",
					Conditions: []v1alpha1.ObservedCondition{
						{
							Type:   "Ready",
							Status: "True",
							Reason: string(constants.AllResourcesAreReady),
						},
					},
				},
			},
		}

		err := cachedClient.Create(ctx, dgd)
		Expect(err).To(BeNil())

		draEw := &v1alpha1.ExternalWorkload{}
		err = draHandler.Parse(ctx, dgd, draEw)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, draEw, expectedEw)
	})
})

const dgdImage = "nvcr.io/nvidia/ai-dynamo/vllm-runtime:0.5.0"

// newV1beta1Component builds a single v1beta1 DGD component with one "main"
// container whose requests/limits and replica/node-count drive the parsed EW.
// A nodeCount > 1 makes it a multinode component.
func newV1beta1Component(name string, componentType dynamov1beta1.ComponentType, replicas, nodeCount int32,
	image string, requests, limits corev1.ResourceList,
) dynamov1beta1.DynamoComponentDeploymentSharedSpec {
	component := dynamov1beta1.DynamoComponentDeploymentSharedSpec{
		ComponentName: name,
		ComponentType: componentType,
		Replicas:      ptr.To(replicas),
		PodTemplate: &corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:      "main",
						Image:     image,
						Resources: corev1.ResourceRequirements{Requests: requests, Limits: limits},
					},
				},
			},
		},
	}
	if nodeCount > 1 {
		component.Multinode = &dynamov1beta1.MultinodeSpec{NodeCount: nodeCount}
	}
	return component
}

// newV1beta1DGD assembles a v1beta1 DynamoGraphDeployment from components, with
// an optional deployment-level clusterTopologyName.
func newV1beta1DGD(name, clusterTopologyName string,
	components ...dynamov1beta1.DynamoComponentDeploymentSharedSpec,
) *dynamov1beta1.DynamoGraphDeployment {
	dgd := &dynamov1beta1.DynamoGraphDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DynamoGraphDeployment",
			APIVersion: "nvidia.com/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: test.TestNamespace,
		},
		Spec: dynamov1beta1.DynamoGraphDeploymentSpec{
			Components: components,
		},
	}
	if clusterTopologyName != "" {
		dgd.Spec.TopologyConstraint = &dynamov1beta1.SpecTopologyConstraint{
			ClusterTopologyName: clusterTopologyName,
		}
	}
	return dgd
}
