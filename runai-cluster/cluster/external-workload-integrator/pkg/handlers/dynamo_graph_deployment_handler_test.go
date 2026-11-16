/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	dynamov1alpha1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1alpha1"
	dynamov1beta1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1beta1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/test"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
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
		utilruntime.Must(kartav1alpha1.AddToScheme(scheme))
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

		// packDomain is a hard requirement, so it maps to RequiredPlacement only.
		Expect(ew.Spec.Topology.PreferredPlacement).To(BeEmpty())
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

		// Topology name comes from v1beta1 clusterTopologyName, placement from packDomain.
		Expect(ew.Spec.Topology.Name).To(Equal("grove-topology"))
		Expect(ew.Spec.Topology.PreferredPlacement).To(BeEmpty())
		Expect(ew.Spec.Topology.RequiredPlacement).To(Equal("rack"))
		// The component has no constraint of its own, so it inherits the deployment-level one.
		Expect(ew.Spec.ComponentTopologies).To(Equal([]v1alpha1.ComponentTopologyEntry{
			{
				SpecName: "VllmWorker",
				Topology: v1alpha1.Topology{Name: "grove-topology", RequiredPlacement: "rack"},
			},
		}))

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
		// No pack domain anywhere, so there is nothing to constrain placement to.
		Expect(ew.Spec.Topology.RequiredPlacement).To(BeEmpty())
		Expect(ew.Spec.ComponentTopologies).To(BeEmpty())
	})

	It("Parse v1beta1 DGD: each component's packDomain overrides the deployment-level one", func() {
		prefill := newV1beta1Component("PrefillWorker", dynamov1beta1.ComponentTypeWorker, 1, 1, dgdImage, nil, nil)
		prefill.TopologyConstraint = &dynamov1beta1.TopologyConstraint{PackDomain: "rack"}
		decode := newV1beta1Component("DecodeWorker", dynamov1beta1.ComponentTypeWorker, 1, 1, dgdImage, nil, nil)
		decode.TopologyConstraint = &dynamov1beta1.TopologyConstraint{PackDomain: "host"}
		// No constraint of its own, so it inherits the deployment-level zone.
		frontend := newV1beta1Component("Frontend", dynamov1beta1.ComponentTypeFrontend, 1, 1, dgdImage, nil, nil)

		dgd := newV1beta1DGD("dgd-v1beta1-component-topology", "grove-topology", prefill, decode, frontend)
		dgd.Spec.TopologyConstraint.PackDomain = "zone"

		Expect(cachedClient.Create(ctx, dgd)).To(BeNil())
		ew := &v1alpha1.ExternalWorkload{}
		Expect(dgdfTopOwnerHandler.Parse(ctx, dgd, ew)).To(BeNil())

		Expect(ew.Spec.Topology).To(Equal(v1alpha1.Topology{Name: "grove-topology", RequiredPlacement: "zone"}))
		// Sorted by component name, so the spec is stable across re-parses.
		Expect(ew.Spec.ComponentTopologies).To(Equal([]v1alpha1.ComponentTopologyEntry{
			{SpecName: "DecodeWorker", Topology: v1alpha1.Topology{Name: "grove-topology", RequiredPlacement: "host"}},
			{SpecName: "Frontend", Topology: v1alpha1.Topology{Name: "grove-topology", RequiredPlacement: "zone"}},
			{SpecName: "PrefillWorker", Topology: v1alpha1.Topology{Name: "grove-topology", RequiredPlacement: "rack"}},
		}))
	})

	It("Parse v1beta1 DGD: a component keeps its packDomain when the deployment has none", func() {
		prefill := newV1beta1Component("PrefillWorker", dynamov1beta1.ComponentTypeWorker, 1, 1, dgdImage, nil, nil)
		prefill.TopologyConstraint = &dynamov1beta1.TopologyConstraint{PackDomain: "rack"}
		frontend := newV1beta1Component("Frontend", dynamov1beta1.ComponentTypeFrontend, 1, 1, dgdImage, nil, nil)

		dgd := newV1beta1DGD("dgd-v1beta1-component-only-topology", "grove-topology", prefill, frontend)

		Expect(cachedClient.Create(ctx, dgd)).To(BeNil())
		ew := &v1alpha1.ExternalWorkload{}
		Expect(dgdfTopOwnerHandler.Parse(ctx, dgd, ew)).To(BeNil())

		// Nothing constrains the workload as a whole.
		Expect(ew.Spec.Topology).To(Equal(v1alpha1.Topology{Name: "grove-topology"}))
		// Frontend has no pack domain to inherit, so it gets no entry.
		Expect(ew.Spec.ComponentTopologies).To(Equal([]v1alpha1.ComponentTopologyEntry{
			{SpecName: "PrefillWorker", Topology: v1alpha1.Topology{Name: "grove-topology", RequiredPlacement: "rack"}},
		}))
	})

	It("Parse v1alpha1 DGD: each service's packDomain overrides the deployment-level one", func() {
		dgd := test.DynamoGraphDeploymentBasic.DeepCopy()
		dgd.Name = fmt.Sprintf("%s-service-topology", dgd.Name)
		dgd.Spec.TopologyConstraint = &dynamov1alpha1.SpecTopologyConstraint{
			TopologyProfile: "grove-topology",
			PackDomain:      "zone",
		}
		dgd.Spec.Services = map[string]*dynamov1alpha1.DynamoComponentDeploymentSharedSpec{
			"PrefillWorker": {
				ComponentType:      "worker",
				Replicas:           ptr.To(int32(1)),
				TopologyConstraint: &dynamov1alpha1.TopologyConstraint{PackDomain: "rack"},
			},
			// Inherits the deployment-level zone.
			"Frontend": {
				ComponentType: "frontend",
				Replicas:      ptr.To(int32(1)),
			},
		}

		Expect(cachedClient.Create(ctx, dgd)).To(BeNil())
		ew := &v1alpha1.ExternalWorkload{}
		Expect(dgdfTopOwnerHandler.Parse(ctx, dgd, ew)).To(BeNil())

		Expect(ew.Spec.Topology).To(Equal(v1alpha1.Topology{Name: "grove-topology", RequiredPlacement: "zone"}))
		Expect(ew.Spec.ComponentTopologies).To(Equal([]v1alpha1.ComponentTopologyEntry{
			{SpecName: "Frontend", Topology: v1alpha1.Topology{Name: "grove-topology", RequiredPlacement: "zone"}},
			{SpecName: "PrefillWorker", Topology: v1alpha1.Topology{Name: "grove-topology", RequiredPlacement: "rack"}},
		}))
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

	It("keeps existing resource totals when active beta components have no pod template", func() {
		worker := newV1beta1Component("worker", dynamov1beta1.ComponentTypeWorker, 2, 1, dgdImage,
			corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}, nil)
		dgd := newV1beta1DGD("missing-template", "", worker)
		ew := &v1alpha1.ExternalWorkload{}
		Expect(dgdfTopOwnerHandler.Parse(ctx, dgd, ew)).To(Succeed())
		dgd.Spec.Components[0].PodTemplate = nil
		Expect(dgdfTopOwnerHandler.Parse(ctx, dgd, ew)).To(Succeed())
		Expect(*ew.Spec.RequestedPods.Replicas).To(Equal(int32(2)))
		Expect(ew.Spec.RequestedResources.Requests.Cpu().String()).To(Equal("2"))
	})

	It("clears beta totals when all components scale to zero", func() {
		worker := newV1beta1Component("worker", dynamov1beta1.ComponentTypeWorker, 2, 1, dgdImage,
			corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}, nil)
		dgd := newV1beta1DGD("zero", "", worker)
		ew := &v1alpha1.ExternalWorkload{}
		Expect(dgdfTopOwnerHandler.Parse(ctx, dgd, ew)).To(Succeed())
		Expect(*ew.Spec.RequestedPods.Replicas).To(Equal(int32(2)))
		dgd.Spec.Components[0].Replicas = ptr.To(int32(0))
		Expect(dgdfTopOwnerHandler.Parse(ctx, dgd, ew)).To(Succeed())
		Expect(*ew.Spec.RequestedPods.Replicas).To(BeZero())
		Expect(ew.Spec.RequestedResources.Requests.Cpu().IsZero()).To(BeTrue())
	})

	DescribeTable("counts role templates from the original beta object",
		func(replicas *int32, nodeCount int32, explicitRoleReplicas, sharedTemplate bool, expectedPods int32, expectedCPU string) {
			object := &unstructured.Unstructured{}
			Expect(json.Unmarshal([]byte(`{
				"apiVersion": "nvidia.com/v1beta1",
				"kind": "DynamoGraphDeployment",
				"metadata": {"name": "roles"},
				"spec": {"components": [{
					"name": "worker", "type": "worker",
					"roles": [
						{"name": "leader", "podTemplate": {"spec": {"containers": [
							{"name": "main", "image": "leader:test", "resources": {"requests": {"cpu": "100m"}}}
						]}}},
						{"name": "worker", "podTemplate": {"spec": {"containers": [
							{"name": "main", "image": "worker:test", "resources": {"requests": {"cpu": "200m"}}}
						]}}}
					]
				}]}
			}`), &object.Object)).To(Succeed())
			component := object.Object["spec"].(map[string]interface{})["components"].([]interface{})[0].(map[string]interface{})
			if replicas != nil {
				component["replicas"] = int64(*replicas)
			}
			component["multinode"] = map[string]interface{}{}
			if nodeCount != 0 {
				component["multinode"].(map[string]interface{})["nodeCount"] = int64(nodeCount)
			}
			if explicitRoleReplicas {
				roles := component["roles"].([]interface{})
				roles[0].(map[string]interface{})["replicas"] = int64(1)
				roles[1].(map[string]interface{})["replicas"] = int64(nodeCount - 1)
			}
			if sharedTemplate {
				roles := component["roles"].([]interface{})
				component["podTemplate"] = roles[0].(map[string]interface{})["podTemplate"]
				for _, role := range roles {
					delete(role.(map[string]interface{}), "podTemplate")
				}
			}
			// The fake tracker decodes registered types on storage and would drop roles.
			roleClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(test.DynamoGraphDeploymentV1beta1Karta()).WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, out client.Object, opts ...client.GetOption) error {
					if key == client.ObjectKeyFromObject(object) {
						return runtime.DefaultUnstructuredConverter.FromUnstructured(object.DeepCopy().Object, out)
					}
					return c.Get(ctx, key, out, opts...)
				},
			}).Build()
			handler := NewDynamoGraphDeploymentTopOwnerHandler(roleClient, nil, func() client.Object { return &dynamov1beta1.DynamoGraphDeployment{} })
			fetched, err := handler.GetObject(ctx, object.GetName(), object.GetNamespace())
			Expect(err).NotTo(HaveOccurred())
			original := fetched.DeepCopyObject()
			ew := &v1alpha1.ExternalWorkload{}
			Expect(handler.Parse(ctx, fetched, ew)).To(Succeed())
			Expect(*ew.Spec.RequestedPods.Replicas).To(Equal(expectedPods))
			Expect(ew.Spec.RequestedResources.Requests.Cpu().String()).To(Equal(expectedCPU))
			Expect(fetched).To(Equal(original))
		},
		Entry("multiple component replicas", ptr.To(int32(2)), int32(3), false, false, int32(6), "1"),
		Entry("default replicas and node count", nil, int32(0), false, false, int32(2), "300m"),
		Entry("explicit role replicas", ptr.To(int32(2)), int32(3), true, false, int32(6), "1"),
		Entry("shared template with role overrides", ptr.To(int32(2)), int32(3), false, true, int32(6), "600m"),
		Entry("zero component replicas", ptr.To(int32(0)), int32(3), false, false, int32(0), "0"),
	)

	DescribeTable("parses ordinary beta workloads without role-aware RI",
		func(withOldKarta bool, roles []any) {
			var objects []client.Object
			if withOldKarta {
				old := test.DynamoGraphDeploymentV1beta1Karta()
				old.Spec.StructureDefinition.ChildComponents = []kartav1alpha1.ComponentDefinition{{Name: "component"}}
				objects = append(objects, old)
			}
			configuredClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(objects...).Build()
			handler := NewDynamoGraphDeploymentTopOwnerHandler(configuredClient, nil,
				func() client.Object { return &dynamov1beta1.DynamoGraphDeployment{} })
			worker := newV1beta1Component("worker", dynamov1beta1.ComponentTypeWorker, 2, 1, dgdImage,
				corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}, nil)
			raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(newV1beta1DGD("ordinary", "", worker))
			Expect(err).NotTo(HaveOccurred())
			if roles != nil {
				component := raw["spec"].(map[string]any)["components"].([]any)[0].(map[string]any)
				component["roles"] = roles
			}
			ew := &v1alpha1.ExternalWorkload{}
			Expect(handler.Parse(ctx, &unstructured.Unstructured{Object: raw}, ew)).To(Succeed())
			Expect(*ew.Spec.RequestedPods.Replicas).To(Equal(int32(2)))
			Expect(ew.Spec.RequestedResources.Requests.Cpu().String()).To(Equal("2"))
		},
		Entry("missing Karta", false, nil),
		Entry("older Karta", true, nil),
		Entry("empty roles without Karta", false, []any{}),
		Entry("roles using the shared template without Karta", false, []any{map[string]any{"name": "leader"}}),
		Entry("null role template with older Karta", true, []any{map[string]any{"name": "leader", "podTemplate": nil}}),
	)

	It("uses the installed Karta role expressions", func() {
		karta := test.DynamoGraphDeploymentV1beta1Karta()
		role := &karta.Spec.StructureDefinition.ChildComponents[0]
		role.InstanceIdPath = ptr.To(`.spec.components[] | .name + "/configured"`)
		role.SpecDefinition.PodTemplateSpecPath = ptr.To(".spec.components[] | .podTemplate")
		role.ScaleDefinition.ReplicasPath = ptr.To(".spec.components[] | 7")
		configuredClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(karta).Build()
		handler := NewDynamoGraphDeploymentTopOwnerHandler(configuredClient, nil,
			func() client.Object { return &dynamov1beta1.DynamoGraphDeployment{} })
		worker := newV1beta1Component("worker", dynamov1beta1.ComponentTypeWorker, 2, 1, dgdImage,
			corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}, nil)
		ew := &v1alpha1.ExternalWorkload{}
		raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(newV1beta1DGD("configured", "", worker))
		Expect(err).NotTo(HaveOccurred())
		component := raw["spec"].(map[string]any)["components"].([]any)[0].(map[string]any)
		component["roles"] = []any{map[string]any{"name": "leader", "podTemplate": component["podTemplate"]}}
		Expect(handler.Parse(ctx, &unstructured.Unstructured{Object: raw}, ew)).To(Succeed())
		Expect(*ew.Spec.RequestedPods.Replicas).To(Equal(int32(7)))
		Expect(ew.Spec.RequestedResources.Requests.Cpu().String()).To(Equal("7"))
	})

	DescribeTable("reports missing role extraction configuration for explicit role templates",
		func(withOldKarta bool, expectedError string) {
			var objects []client.Object
			if withOldKarta {
				old := test.DynamoGraphDeploymentV1beta1Karta()
				old.Spec.StructureDefinition.ChildComponents = nil
				objects = append(objects, old)
			}
			configuredClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(objects...).Build()
			handler := NewDynamoGraphDeploymentTopOwnerHandler(configuredClient, nil,
				func() client.Object { return &dynamov1beta1.DynamoGraphDeployment{} })
			object := &unstructured.Unstructured{}
			Expect(json.Unmarshal([]byte(`{
              "apiVersion":"nvidia.com/v1beta1","kind":"DynamoGraphDeployment",
              "spec":{"components":[{"name":"worker","roles":[
                {"name":"leader","podTemplate":{"spec":{"containers":[{"name":"main"}]}}}
              ]}]}
            }`), &object.Object)).To(Succeed())
			Expect(handler.Parse(ctx, object, &v1alpha1.ExternalWorkload{})).
				To(MatchError(ContainSubstring(expectedError)))
		},
		Entry("missing Karta", false, "get Dynamo v1beta1 Karta"),
		Entry("older Karta", true, "get Dynamo v1beta1 roles"),
	)

	It("keeps component order and topology when parsing role instances", func() {
		object := &unstructured.Unstructured{}
		Expect(json.Unmarshal([]byte(`{
          "apiVersion":"nvidia.com/v1beta1", "kind":"DynamoGraphDeployment",
          "metadata":{"name":"mixed"},
          "spec":{
            "topologyConstraint":{"clusterTopologyName":"topology","packDomain":"rack"},
            "components":[
              {"name":"frontend","type":"frontend","podTemplate":{"spec":{"containers":[
                {"name":"main","image":"frontend:test","resources":{"requests":{"cpu":"1"}}}
              ]}}},
              {"name":"worker","type":"worker","replicas":2,"roles":[
                {"name":"leader","podTemplate":{"spec":{"containers":[
                  {"name":"main","image":"worker:test","resources":{"requests":{"cpu":"200m"}}}
                ]}}}
              ]}
            ]
          }
        }`), &object.Object)).To(Succeed())
		configuredClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).
			WithObjects(test.DynamoGraphDeploymentV1beta1Karta()).Build()
		handler := NewDynamoGraphDeploymentTopOwnerHandler(configuredClient, nil,
			func() client.Object { return &dynamov1beta1.DynamoGraphDeployment{} })
		ew := &v1alpha1.ExternalWorkload{}
		original := object.DeepCopy()
		for range 2 {
			Expect(handler.Parse(ctx, object, ew)).To(Succeed())
			Expect(*ew.Spec.RequestedPods.Replicas).To(Equal(int32(3)))
			Expect(ew.Spec.RequestedResources.Requests.Cpu().String()).To(Equal("1400m"))
			Expect(ew.Spec.ImageName).To(Equal("frontend:test"))
			Expect(ew.Spec.ComponentTopologies).To(Equal([]v1alpha1.ComponentTopologyEntry{
				{SpecName: "frontend", Topology: v1alpha1.Topology{Name: "topology", RequiredPlacement: "rack"}},
				{SpecName: "worker", Topology: v1alpha1.Topology{Name: "topology", RequiredPlacement: "rack"}},
			}))
			Expect(object).To(Equal(original))
		}
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
