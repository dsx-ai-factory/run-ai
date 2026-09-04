/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	rayv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/ray/v1"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	commonconsts "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/test"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testNamespace = "ns-proj1"

	testImage    = "test-image:latest"
	testPriority = "some-priority"

	testCpuLimit        = 100
	testMemoryLimit     = 1000
	testGpuLimit        = 10
	testCpuRequest      = 50
	testMemoryRequest   = 500
	testGpuRequest      = 5
	testFracGpuLimitStr = "0.5"
	testFracGpuLimit    = 0.5
	testGpuMemoryLimit  = 10000

	GpuFractionAnnotation              = "gpu-fraction"
	GpuFractionDevicesAnnotation       = "gpu-fraction-num-devices"
	GpuMemoryAnnotation                = "gpu-memory"
	GpuMemoryLimitConversion     int64 = 1000 * 1000 // convert from MB to Bytes

)

type TestData struct {
	testName                 string
	objectToParse            client.Object
	karta                    *kartav1alpha1.Karta
	expectedExternalWorkload *v1alpha1.ExternalWorkload
}

var (
	podKarta = &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{
			Name: "karta-pod-test",
		},
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "Pod",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Pod",
					},
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodSpecPath:  ptr.To(".spec"),
						MetadataPath: ptr.To(".metadata"),
					},
					StatusDefinition: &kartav1alpha1.StatusDefinition{
						PhaseDefinition: &kartav1alpha1.PhaseDefinition{
							Path: ".status.phase",
						},
						ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
							Path:             ".status.conditions",
							TypeFieldName:    "type",
							StatusFieldName:  "status",
							MessageFieldName: ptr.To("message"),
						},
						StatusMappings: kartav1alpha1.StatusMappings{
							Initializing: []kartav1alpha1.StatusMatcher{
								{ByConditions: []kartav1alpha1.ExpectedCondition{{Type: "Initialized", Status: ptr.To("True")}}},
								{ByConditions: []kartav1alpha1.ExpectedCondition{{Type: "PodScheduled", Status: ptr.To("True")}}},
							},
							Running:   []kartav1alpha1.StatusMatcher{{ByPhase: "Running"}},
							Completed: []kartav1alpha1.StatusMatcher{{ByPhase: "Succeeded"}},
							Failed:    []kartav1alpha1.StatusMatcher{{ByPhase: "Failed"}},
						},
					},
				},
			},
		},
	}

	deploymentKarta = &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{
			Name: "karta-deployment-test",
		},
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "Deployment",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "apps",
						Version: "v1",
						Kind:    "Deployment",
					},
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.template"),
					},
					ScaleDefinition: &kartav1alpha1.ScaleDefinition{
						ReplicasPath: ptr.To(".spec.replicas"),
					},
				},
			},
		},
	}

	rayClusterKarta = &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{
			Name: "karta-raycluster-test",
		},
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "RayCluster",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "ray.io",
						Version: "v1",
						Kind:    "RayCluster",
					},
				},
				ChildComponents: []kartav1alpha1.ComponentDefinition{
					{
						Name:     "headGroup",
						OwnerRef: ptr.To("RayCluster"),
						SpecDefinition: &kartav1alpha1.SpecDefinition{
							PodTemplateSpecPath: ptr.To(".spec.headGroupSpec.template"),
						},
					},
					{
						Name:           "workerGroup",
						OwnerRef:       ptr.To("RayCluster"),
						InstanceIdPath: ptr.To(".spec.workerGroupSpecs[]?.groupName"),
						SpecDefinition: &kartav1alpha1.SpecDefinition{
							PodTemplateSpecPath: ptr.To(".spec.workerGroupSpecs[]?.template"),
						},
						ScaleDefinition: &kartav1alpha1.ScaleDefinition{
							ReplicasPath:    ptr.To(".spec.workerGroupSpecs[]?.replicas"),
							MinReplicasPath: ptr.To(".spec.workerGroupSpecs[]?.minReplicas"),
							MaxReplicasPath: ptr.To(".spec.workerGroupSpecs[]?.maxReplicas"),
						},
					},
				},
				AdditionalChildKinds: []kartav1alpha1.GroupVersionKind{
					{
						Group:   "ray.io",
						Version: "v1",
						Kind:    "RayJob",
					},
					{
						Group:   "ray.io",
						Version: "v1",
						Kind:    "RayService",
					},
				},
			},
		},
	}

	rayJobKarta = &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{
			Name: "karta-rayjob-test",
		},
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "RayJob",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "ray.io",
						Version: "v1",
						Kind:    "RayJob",
					},
				},
				ChildComponents: []kartav1alpha1.ComponentDefinition{
					{
						Name:     "headGroup",
						OwnerRef: ptr.To("RayJob"),
						SpecDefinition: &kartav1alpha1.SpecDefinition{
							PodTemplateSpecPath: ptr.To(".spec.rayClusterSpec.headGroupSpec.template"),
						},
					},
					{
						Name:           "workerGroup",
						OwnerRef:       ptr.To("RayJob"),
						InstanceIdPath: ptr.To(".spec.rayClusterSpec.workerGroupSpecs[]?.groupName"),
						SpecDefinition: &kartav1alpha1.SpecDefinition{
							PodTemplateSpecPath: ptr.To(".spec.rayClusterSpec.workerGroupSpecs[]?.template"),
						},
						ScaleDefinition: &kartav1alpha1.ScaleDefinition{
							ReplicasPath:    ptr.To(".spec.rayClusterSpec.workerGroupSpecs[]?.replicas"),
							MinReplicasPath: ptr.To(".spec.rayClusterSpec.workerGroupSpecs[]?.minReplicas"),
							MaxReplicasPath: ptr.To(".spec.rayClusterSpec.workerGroupSpecs[]?.maxReplicas"),
						},
					},
				},
				AdditionalChildKinds: []kartav1alpha1.GroupVersionKind{
					{
						Group:   "ray.io",
						Version: "v1",
						Kind:    "RayCluster",
					},
					{
						Group:   "ray.io",
						Version: "v1",
						Kind:    "RayService",
					},
				},
			},
		},
	}

	kubevirtVMKarta = &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{
			Name: "karta-kubevirt-test",
		},
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "VirtualMachineInstance",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "kubevirt.io",
						Version: "v1",
						Kind:    "VirtualMachineInstance",
					},
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
							ResourcesPath:         ptr.To(".spec.domain.resources"),
							PriorityClassNamePath: ptr.To(".spec.priorityClassName"),
							NodeAffinityPath:      ptr.To(".spec.affinity.nodeAffinity"),
						},
					},
				},
			},
		},
	}
)

var _ = Describe("TestKartaHandler", Ordered, func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc

		fakeClient   client.Client
		kartaHandler *KartaTopOwnerHandler
	)

	BeforeAll(func() {
		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(func() {
			cancel()
		})

		scheme := getTestScheme()
		ns := test.BuildNamespace(testNamespace, map[string]string{constants.RunaiQueueLabel: "proj"})
		workloadConfigMap := test.CreateWorkloadsConfigMapObject()
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&v1alpha1.ExternalWorkload{}).
			WithObjects(ns, workloadConfigMap, test.TestNodePool, test.TestKaiNodePool).Build()

		kartaHandler = NewKartaTopOwnerHandler(fakeClient, nil)
	})

	DescribeTable("Sanity - externalworkload exists with correct info for all supported types",
		Ordered,
		func(testData *TestData) {
			result := &v1alpha1.ExternalWorkload{}
			err := kartaHandler.Parse(ctx, testData.objectToParse, testData.karta, result)
			Expect(err).ToNot(HaveOccurred(), "Failed to parse object for test: %s", testData.testName)

			test.ValidateExternalWorkloadInnerSpec(Default, 1, result, testData.expectedExternalWorkload)
		},
		Entry("Test KartaHandler with Pod", &TestData{
			testName: "Pod",
			objectToParse: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					UID:               "pod-2-UID",
					Name:              "pod-2-5467849487-vqbfj",
					Namespace:         testNamespace,
					CreationTimestamp: metav1.Now(),
					Labels:            map[string]string{},
					Annotations: map[string]string{
						GpuFractionAnnotation: testFracGpuLimitStr,
						GpuMemoryAnnotation:   strconv.Itoa(testGpuMemoryLimit),
					},
				},
				Spec: corev1.PodSpec{
					SchedulerName: constants.RunaiSchedulerName,
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      test.TestNodePool.Spec.LabelKey,
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{test.TestNodePool.Spec.LabelValue},
											},
										},
									},
								},
							},
						},
					},
					PriorityClassName: testPriority,
					Containers: []corev1.Container{
						{
							Image: testImage,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:        *resource.NewQuantity(testCpuLimit, resource.DecimalSI),
									corev1.ResourceMemory:     *resource.NewQuantity(testMemoryLimit, resource.DecimalSI),
									constants.GpuResourceName: *resource.NewQuantity(testGpuLimit, resource.DecimalSI),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:        *resource.NewQuantity(testCpuRequest, resource.DecimalSI),
									corev1.ResourceMemory:     *resource.NewQuantity(testMemoryRequest, resource.DecimalSI),
									constants.GpuResourceName: *resource.NewQuantity(testGpuRequest, resource.DecimalSI),
								},
							},
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					Conditions: []corev1.PodCondition{
						{
							Type:    corev1.PodInitialized,
							Status:  corev1.ConditionTrue,
							Message: "pod init",
						},
					},
				},
			},
			karta: podKarta,
			expectedExternalWorkload: &v1alpha1.ExternalWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						commonconsts.WorkloadCategoryLabel: "Train",
					},
				},
				Spec: v1alpha1.ExternalWorkloadSpec{
					RequestedPods: v1alpha1.RequestedPods{
						Replicas: ptr.To(int32(1)),
					},
					ImageName:          testImage,
					PriorityClassName:  testPriority,
					RequestedNodePools: []string{test.TestNodePool.Name},
					RequestedResources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:              *resource.NewQuantity(testCpuLimit, resource.DecimalSI),
							corev1.ResourceMemory:           *resource.NewQuantity(testMemoryLimit, resource.DecimalSI),
							constants.GpuResourceName:       *resource.NewMilliQuantity(1000*(testGpuLimit+testFracGpuLimit), resource.DecimalSI),
							constants.GpuMemoryResourceName: *resource.NewQuantity(GpuMemoryLimitConversion*testGpuMemoryLimit, resource.DecimalSI),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:              *resource.NewQuantity(testCpuRequest, resource.DecimalSI),
							corev1.ResourceMemory:           *resource.NewQuantity(testMemoryRequest, resource.DecimalSI),
							constants.GpuResourceName:       *resource.NewMilliQuantity(1000*(testGpuRequest+testFracGpuLimit), resource.DecimalSI),
							constants.GpuMemoryResourceName: *resource.NewQuantity(GpuMemoryLimitConversion*testGpuMemoryLimit, resource.DecimalSI),
						},
					},
				},
				Status: v1alpha1.ExternalWorkloadStatus{
					ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
						Phase:     v1alpha1.ExWInitializingPhase,
						K8sStatus: "Pending",
						Conditions: []v1alpha1.ObservedCondition{
							{
								Type:    "Initialized",
								Status:  corev1.ConditionTrue,
								Message: "pod init",
							},
						},
					},
				},
			},
		}),
		Entry("Test KartaHandler with Pod 2", &TestData{
			testName: "Pod2",
			objectToParse: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					UID:               "some-pod-UID",
					Name:              "some-pod-5467849487-vqbfj",
					Namespace:         testNamespace,
					Labels:            map[string]string{},
					CreationTimestamp: metav1.Now(),
				},
				Spec: *test.TestPod.Spec.DeepCopy(),
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					Conditions: []corev1.PodCondition{
						{
							Type:    corev1.PodInitialized,
							Status:  corev1.ConditionFalse,
							Message: "pod not init",
						},
						{
							Type:    corev1.PodReady,
							Status:  corev1.ConditionTrue,
							Message: "pod ready",
							Reason:  "some reason",
						},
					},
				},
			},
			karta: podKarta,
			expectedExternalWorkload: &v1alpha1.ExternalWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: *test.TestPodExpectedEw.Spec.DeepCopy(),
				Status: v1alpha1.ExternalWorkloadStatus{
					ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
						Phase:     v1alpha1.ExWRunningPhase,
						K8sStatus: "Running",
						Conditions: []v1alpha1.ObservedCondition{
							{
								Type:    "Initialized",
								Status:  corev1.ConditionFalse,
								Message: "pod not init",
							},
							{
								Type:    "Ready",
								Status:  corev1.ConditionTrue,
								Message: "pod ready",
							},
						},
					},
				},
			},
		}),
		Entry("Test KartaHandler with Deployment", &TestData{
			testName: "Deployment",
			objectToParse: &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					UID:       "deploy-1-UID",
					Name:      "deploy-1",
					Namespace: testNamespace,
					Annotations: map[string]string{
						"otherAnnotation": "otheraal",
					},
					CreationTimestamp: metav1.Now(),
				},
				Spec: *test.TestDeployment.Spec.DeepCopy(),
			},
			karta: deploymentKarta,
			expectedExternalWorkload: &v1alpha1.ExternalWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: *test.TestDeploymentExpectedBasicEw.Spec.DeepCopy(),
				Status: v1alpha1.ExternalWorkloadStatus{
					ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
						Phase: v1alpha1.ExWUnknownPhase,
					},
				},
			},
		}),
		Entry("Test KartaHandler with RayCluster", &TestData{
			testName: "Test KartaHandler with RayCluster",
			objectToParse: &rayv1.RayCluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RayCluster",
					APIVersion: "ray.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					UID:               "raycluster1-UID",
					Name:              "raycluster1-3242344",
					Namespace:         testNamespace,
					Labels:            map[string]string{},
					CreationTimestamp: metav1.Now(),
				},
				Spec: *test.TestRayClusterV1.Spec.DeepCopy(),
			},
			karta: rayClusterKarta,
			expectedExternalWorkload: &v1alpha1.ExternalWorkload{
				Spec: *test.TestRayClusterExpectedEw.Spec.DeepCopy(),
				Status: v1alpha1.ExternalWorkloadStatus{
					ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
						Phase: v1alpha1.ExWUnknownPhase,
					},
				},
			},
		}),
		Entry("Test KartaHandler with RayJob", &TestData{
			testName: "RayJob",
			objectToParse: &rayv1.RayJob{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RayJob",
					APIVersion: "ray.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					UID:               "rayjob1-UID",
					Name:              "rayjob1-3242344",
					Namespace:         testNamespace,
					Labels:            map[string]string{},
					CreationTimestamp: metav1.Now(),
				},
				Spec: *test.TestRayJobV1.Spec.DeepCopy(),
			},
			karta: rayJobKarta,
			expectedExternalWorkload: &v1alpha1.ExternalWorkload{
				Spec: *test.TestRayJobExpectedEw.Spec.DeepCopy(),
				Status: v1alpha1.ExternalWorkloadStatus{
					ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
						Phase: v1alpha1.ExWUnknownPhase,
					},
				},
			},
		}),
		Entry("Test KartaHandler with Kubevirt VirtualMachineInstance", &TestData{
			testName: "VirtualMachineInstance",
			objectToParse: &kubevirtv1.VirtualMachineInstance{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VirtualMachineInstance",
					APIVersion: "kubevirt.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					UID:               "kubevirtvm1-UID",
					Name:              "kubevirtvm1-3242344",
					Namespace:         testNamespace,
					Labels:            map[string]string{},
					CreationTimestamp: metav1.Now(),
				},
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					PriorityClassName: testPriority,
					Domain: kubevirtv1.DomainSpec{
						Resources: kubevirtv1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(testCpuRequest, resource.DecimalSI),
								corev1.ResourceMemory: *resource.NewQuantity(testMemoryRequest, resource.DecimalSI),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewQuantity(testCpuLimit, resource.DecimalSI),
								corev1.ResourceMemory: *resource.NewQuantity(testMemoryLimit, resource.DecimalSI),
							},
						},
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      test.TestNodePool.Spec.LabelKey,
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{test.TestNodePool.Spec.LabelValue},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			karta: kubevirtVMKarta,
			expectedExternalWorkload: &v1alpha1.ExternalWorkload{
				Spec: v1alpha1.ExternalWorkloadSpec{
					PriorityClassName: testPriority,
					RequestedResources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    *resource.NewQuantity(testCpuRequest, resource.DecimalSI),
							corev1.ResourceMemory: *resource.NewQuantity(testMemoryRequest, resource.DecimalSI),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    *resource.NewQuantity(testCpuLimit, resource.DecimalSI),
							corev1.ResourceMemory: *resource.NewQuantity(testMemoryLimit, resource.DecimalSI),
						},
					},
					RequestedNodePools: []string{test.TestNodePool.Name},
					RequestedPods: v1alpha1.RequestedPods{
						Replicas: ptr.To(int32(1)),
					},
				},
				Status: v1alpha1.ExternalWorkloadStatus{
					ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
						Phase: v1alpha1.ExWUnknownPhase,
					},
				},
			},
		}),
		Entry("Test KserveInferenceService", &TestData{
			testName:      "Test KserveInferenceService",
			objectToParse: test.TestKserveInferenceService,
			karta:         test.KartaForKserveInferenceService,
			expectedExternalWorkload: &v1alpha1.ExternalWorkload{
				Spec:   *test.TestKserveExpectedEw.Spec.DeepCopy(),
				Status: *test.TestKserveExpectedEw.Status.DeepCopy(),
			}}),
	)

	It("aggregates resources from nested child components", func() {
		leaderWorkerSet := &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "leaderworkerset.x-k8s.io/v1",
			"kind":       "LeaderWorkerSet",
			"metadata": map[string]any{
				"name":      "distributed-inference",
				"namespace": testNamespace,
			},
			"spec": map[string]any{
				"replicas": int64(2),
				"leaderWorkerTemplate": map[string]any{
					"size": int64(3),
					"leaderTemplate": map[string]any{
						"metadata": map[string]any{
							"annotations": map[string]any{
								GpuFractionAnnotation:        "0.5",
								GpuFractionDevicesAnnotation: "1",
							},
						},
						"spec": map[string]any{
							"containers": []any{map[string]any{
								"name":  "leader",
								"image": "leader:latest",
								"resources": map[string]any{
									"requests": map[string]any{"cpu": "200m", "memory": "64M"},
									"limits":   map[string]any{"cpu": "500m", "memory": "128M"},
								},
							}},
						},
					},
					"workerTemplate": map[string]any{
						"metadata": map[string]any{
							"annotations": map[string]any{
								GpuFractionAnnotation:        "0.25",
								GpuFractionDevicesAnnotation: "1",
							},
						},
						"spec": map[string]any{
							"containers": []any{map[string]any{
								"name":  "worker",
								"image": "worker:latest",
								"resources": map[string]any{
									"requests": map[string]any{"cpu": "100m", "memory": "64M"},
									"limits":   map[string]any{"cpu": "250m", "memory": "128M"},
								},
							}},
						},
					},
				},
			},
		}}
		leaderWorkerSetKarta := &kartav1alpha1.Karta{
			Spec: kartav1alpha1.KartaSpec{
				StructureDefinition: kartav1alpha1.StructureDefinition{
					RootComponent: kartav1alpha1.ComponentDefinition{
						Name: "leaderworkerset",
						Kind: &kartav1alpha1.GroupVersionKind{
							Group: "leaderworkerset.x-k8s.io", Version: "v1", Kind: "LeaderWorkerSet",
						},
					},
					ChildComponents: []kartav1alpha1.ComponentDefinition{
						{
							Name:     "group",
							OwnerRef: ptr.To("leaderworkerset"),
							ScaleDefinition: &kartav1alpha1.ScaleDefinition{
								ReplicasPath: ptr.To(".spec.leaderWorkerTemplate.size"),
							},
						},
						{
							Name:     "leader",
							OwnerRef: ptr.To("group"),
							SpecDefinition: &kartav1alpha1.SpecDefinition{
								PodTemplateSpecPath: ptr.To(".spec.leaderWorkerTemplate.leaderTemplate"),
							},
							ScaleDefinition: &kartav1alpha1.ScaleDefinition{
								ReplicasPath: ptr.To(".spec.replicas // 1"),
							},
						},
						{
							Name:     "worker",
							OwnerRef: ptr.To("group"),
							SpecDefinition: &kartav1alpha1.SpecDefinition{
								PodTemplateSpecPath: ptr.To(".spec.leaderWorkerTemplate.workerTemplate"),
							},
							ScaleDefinition: &kartav1alpha1.ScaleDefinition{
								ReplicasPath: ptr.To("(.spec.replicas // 1) * ((.spec.leaderWorkerTemplate.size // 1) - 1)"),
							},
						},
					},
				},
			},
		}

		result := &v1alpha1.ExternalWorkload{}
		err := kartaHandler.Parse(ctx, leaderWorkerSet, leaderWorkerSetKarta, result)
		Expect(err).ToNot(HaveOccurred())

		Expect(result.Spec.RequestedPods.Replicas).ToNot(BeNil())
		Expect(*result.Spec.RequestedPods.Replicas).To(Equal(int32(6)))
		Expect(result.Spec.RequestedResources.Requests.Cpu().Cmp(resource.MustParse("800m"))).To(Equal(0))
		Expect(result.Spec.RequestedResources.Limits.Cpu().Cmp(resource.MustParse("2"))).To(Equal(0))
		Expect(result.Spec.RequestedResources.Requests.Memory().Cmp(resource.MustParse("384M"))).To(Equal(0))
		Expect(result.Spec.RequestedResources.Limits.Memory().Cmp(resource.MustParse("768M"))).To(Equal(0))
		Expect(result.Spec.RequestedResources.Requests.Name(constants.GpuResourceName, resource.DecimalSI).
			Cmp(resource.MustParse("2"))).To(Equal(0))
		Expect(result.Spec.RequestedResources.Limits.Name(constants.GpuResourceName, resource.DecimalSI).
			Cmp(resource.MustParse("2"))).To(Equal(0))
	})

	It("does not inspect components below the first nested pod-bearing level", func() {
		workload := &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "example.run.ai/v1",
			"kind":       "NestedWorkload",
			"metadata": map[string]any{
				"name":      "nested-workload",
				"namespace": testNamespace,
			},
			"spec": map[string]any{
				"parentTemplate": map[string]any{
					"spec": map[string]any{
						"containers": []any{map[string]any{
							"name":  "parent",
							"image": "parent:latest",
							"resources": map[string]any{
								"requests": map[string]any{"cpu": "1"},
							},
						}},
					},
				},
				"nestedTemplate": map[string]any{
					"spec": map[string]any{
						"containers": []any{map[string]any{
							"name":  "nested",
							"image": "nested:latest",
							"resources": map[string]any{
								"requests": map[string]any{"cpu": "10"},
							},
						}},
					},
				},
			},
		}}
		karta := &kartav1alpha1.Karta{
			Spec: kartav1alpha1.KartaSpec{
				StructureDefinition: kartav1alpha1.StructureDefinition{
					RootComponent: kartav1alpha1.ComponentDefinition{
						Name: "root",
						Kind: &kartav1alpha1.GroupVersionKind{
							Group: "example.run.ai", Version: "v1", Kind: "NestedWorkload",
						},
					},
					ChildComponents: []kartav1alpha1.ComponentDefinition{
						{
							Name:     "group",
							OwnerRef: ptr.To("root"),
						},
						{
							Name:     "parent",
							OwnerRef: ptr.To("group"),
							SpecDefinition: &kartav1alpha1.SpecDefinition{
								PodTemplateSpecPath: ptr.To(".spec.parentTemplate"),
							},
							ScaleDefinition: &kartav1alpha1.ScaleDefinition{
								ReplicasPath: ptr.To("1"),
							},
						},
						{
							Name:     "nested",
							OwnerRef: ptr.To("parent"),
							SpecDefinition: &kartav1alpha1.SpecDefinition{
								PodTemplateSpecPath: ptr.To(".spec.nestedTemplate"),
							},
							ScaleDefinition: &kartav1alpha1.ScaleDefinition{
								ReplicasPath: ptr.To("1"),
							},
						},
					},
				},
			},
		}

		result := &v1alpha1.ExternalWorkload{}
		err := kartaHandler.Parse(ctx, workload, karta, result)
		Expect(err).ToNot(HaveOccurred())

		Expect(result.Spec.RequestedPods.Replicas).ToNot(BeNil())
		Expect(*result.Spec.RequestedPods.Replicas).To(Equal(int32(1)))
		Expect(result.Spec.RequestedResources.Requests.Cpu().Cmp(resource.MustParse("1"))).To(Equal(0))
	})
})

var _ = Describe("TestKartaHandler DRA", Ordered, func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc

		fakeClient   client.Client
		kartaHandler *KartaTopOwnerHandler
	)

	BeforeAll(func() {
		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(func() {
			cancel()
		})

		scheme := getTestScheme()
		ns := test.BuildNamespace(testNamespace, map[string]string{constants.RunaiQueueLabel: "proj"})
		workloadConfigMap := test.CreateWorkloadsConfigMapObject()
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&v1alpha1.ExternalWorkload{}).
			WithObjects(ns, workloadConfigMap, test.TestNodePool, test.TestKaiNodePool).Build()

		draFakeClient := fake.NewClientBuilder().WithScheme(getTestScheme()).WithObjects(test.TestDRAGpuClaim.DeepCopy(), test.TestDRAGpuClaimTemplate.DeepCopy()).Build()
		fakeKubeClient := kubefake.NewClientset(test.TestDRAGpuClaim.DeepCopy(), test.TestDRAGpuClaimTemplate.DeepCopy())
		draParser := dra.NewDRAParser(draFakeClient, fakeKubeClient)
		kartaHandler = NewKartaTopOwnerHandler(fakeClient, draParser)
	})

	It("Test KartaHandler with Deployment DRA ResourceClaims", func() {
		deployment := &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				UID:               "deploy-dra-UID",
				Name:              "deploy-dra",
				Namespace:         testNamespace,
				CreationTimestamp: metav1.Now(),
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To(int32(2)),
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						SchedulerName: constants.RunaiSchedulerName,
						Containers: []corev1.Container{
							{
								Name:  "main",
								Image: testImage,
							},
						},
						ResourceClaims: []corev1.PodResourceClaim{
							{
								Name:                      "gpu-claim",
								ResourceClaimTemplateName: ptr.To("dra-gpu-template"),
							},
						},
					},
				},
			},
		}

		expectedEw := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: testNamespace,
			},
			Spec: v1alpha1.ExternalWorkloadSpec{
				RequestedPods: v1alpha1.RequestedPods{
					Replicas: ptr.To(int32(2)),
				},
				RequestedResources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						constants.GpuResourceName: *resource.NewQuantity(8, resource.DecimalSI), // 2 replicas * 4 GPUs
					},
					Requests: corev1.ResourceList{
						constants.GpuResourceName: *resource.NewQuantity(8, resource.DecimalSI),
					},
				},
				ImageName:          testImage,
				RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
			},
			Status: v1alpha1.ExternalWorkloadStatus{
				ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
					Phase: v1alpha1.ExWUnknownPhase,
				},
			},
		}

		result := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: testNamespace,
			},
		}
		err := kartaHandler.Parse(ctx, deployment, deploymentKarta, result)
		Expect(err).ToNot(HaveOccurred())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, result, expectedEw)
	})
})

func getTestScheme() *runtime.Scheme {
	scheme := scheme_init.Scheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	return scheme
}
