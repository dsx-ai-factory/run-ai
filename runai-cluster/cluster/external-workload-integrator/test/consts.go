/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"encoding/json"
	"strconv"

	upstreamnimv1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	devv1alpha2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	kservev1beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/kserve/v1beta1"
	kubeflowv1beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/kubeflownotebook/v1beta1"
	nim "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nim/v1alpha1"
	nvcfv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nvcf/v1"
	seldonv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/seldon/v1"
	seldonv1alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/seldon/v1alpha2"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	workloadTypeConfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/workload-type-config"
	commonconsts "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/priority"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kubeflowv2beta1 "github.com/kubeflow/mpi-operator/pkg/apis/kubeflow/v2beta1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	rayv1alpha1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TestNamespace           = "ns-proj1"
	RunAiTestNamespace      = "runai-proj001"
	TestNonProjectNamespace = "some-namespace"
	testProjectName         = "proj1"
	testProjectId           = "1"
	testDepartmentName      = "dep1"
	testDepartmentId        = "2"

	testDeploymentName             = "depl1"
	testDeploymentUID              = "depl1-UID"
	testDeploymentNumberOfReplicas = 3
	testImage                      = "ubuntu"
	testDeploymentPriority         = "some-priority"
	testDeploymentCpuLimit         = 100
	testDeploymentMemoryLimit      = 1000
	testDeploymentGpuLimit         = 10
	testDeploymentCpuRequest       = 50
	testDeploymentMemoryRequest    = 500
	testDeploymentGpuRequest       = 5
	testDeploymentFracGpuLimitStr  = "0.5"
	testDeploymentFracGpuLimit     = 0.5
	testDeploymentGpuMemoryLimit   = 10000

	testPodGroupName = "pod-group-12121212121"
	testPodGroupId   = "pod-group-id"

	npAName = "nodepool-a"
	npAKey  = "keya"
	npAVal  = "vala"

	// annotations used to submit external workloads with fraction/memory
	GpuFractionAnnotation = "gpu-fraction"
	GpuMemoryAnnotation   = "gpu-memory"

	GpuMemoryLimitConversion int64 = 1000 * 1000 // convert from MB to Bytes
)

var (
	TestDeployment = &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:       testDeploymentUID,
			Name:      testDeploymentName,
			Namespace: TestNamespace,
			Labels: map[string]string{
				constants.ManagedByLabelKey: "Helm",
				"otherLabel":                "otherval",
			},
			Annotations: map[string]string{
				constants.HelmReleaseAnnotationKey: "rel",
				"otherAnnotation":                  "otheraal",
			},
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(testDeploymentNumberOfReplicas)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						GpuFractionAnnotation: testDeploymentFracGpuLimitStr,
						GpuMemoryAnnotation:   strconv.Itoa(testDeploymentGpuMemoryLimit),
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
												Key:      npAKey,
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{npAVal},
											},
										},
									},
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      v1alpha1.NodePoolNameLabel,
												Operator: corev1.NodeSelectorOpDoesNotExist,
											},
										},
									},
								},
							},
						},
					},
					PriorityClassName: testDeploymentPriority,
					Containers: []corev1.Container{
						{
							Name:  "container-name",
							Image: testImage,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuLimit, resource.DecimalSI),
									corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryLimit, resource.DecimalSI),
									constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuLimit, resource.DecimalSI),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuRequest, resource.DecimalSI),
									corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryRequest, resource.DecimalSI),
									constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest, resource.DecimalSI),
								},
							},
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:    appsv1.DeploymentProgressing,
					Status:  corev1.ConditionTrue,
					Reason:  "NewReplicaSetAvailable",
					Message: "ReplicaSet \"depl1-5467849487\" has successfully progressed.",
				},
			},
		},
	}
	testDeploymentChild1RS                 *appsv1.ReplicaSet
	testDeploymentChild2Pod                *corev1.Pod
	testDeploymentChildPodWithPodGroup     *corev1.Pod
	testDeploymentChildPodBadSchedulerName *corev1.Pod

	// using the regular flow of the external_workload_handler
	testDeploymentExpectedEw *v1alpha1.ExternalWorkload
	// TestDeploymentExpectedBasicEw only using HandleDeploymentTopOwner
	TestDeploymentExpectedBasicEw = &v1alpha1.ExternalWorkload{
		//TypeMeta: metav1.TypeMeta{
		//	Kind: "Deployment",
		//},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				commonconsts.WorkloadCategoryLabel: "Deploy",
				constants.ManagedByLabelKey:        "Helm",
			},
			Annotations: map[string]string{
				constants.HelmReleaseAnnotationKey: "rel",
			},
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(testDeploymentNumberOfReplicas)),
			},
			ImageName:          testImage,
			PriorityClassName:  testDeploymentPriority,
			Preemptibility:     "preemptible",
			RequestedNodePools: []string{npAName, v1alpha1.DefaultNodePoolName},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:              *resource.NewQuantity(testDeploymentNumberOfReplicas*testDeploymentCpuLimit, resource.DecimalSI),
					corev1.ResourceMemory:           *resource.NewQuantity(testDeploymentNumberOfReplicas*testDeploymentMemoryLimit, resource.DecimalSI),
					constants.GpuResourceName:       *resource.NewMilliQuantity(1000*testDeploymentNumberOfReplicas*(testDeploymentGpuLimit+testDeploymentFracGpuLimit), resource.DecimalSI),
					constants.GpuMemoryResourceName: *resource.NewQuantity(GpuMemoryLimitConversion*testDeploymentNumberOfReplicas*testDeploymentGpuMemoryLimit, resource.DecimalSI),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:              *resource.NewQuantity(testDeploymentNumberOfReplicas*testDeploymentCpuRequest, resource.DecimalSI),
					corev1.ResourceMemory:           *resource.NewQuantity(testDeploymentNumberOfReplicas*testDeploymentMemoryRequest, resource.DecimalSI),
					constants.GpuResourceName:       *resource.NewMilliQuantity(1000*testDeploymentNumberOfReplicas*(testDeploymentGpuRequest+testDeploymentFracGpuLimit), resource.DecimalSI),
					constants.GpuMemoryResourceName: *resource.NewQuantity(GpuMemoryLimitConversion*testDeploymentNumberOfReplicas*testDeploymentGpuMemoryLimit, resource.DecimalSI),
				},
			},
			ProjectName:    testProjectName,
			ProjectId:      testProjectId,
			DepartmentName: testDepartmentName,
			DepartmentId:   testDepartmentId,
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: "Complete",
				Message:   "ReplicaSet \"depl1-5467849487\" has successfully progressed.",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(appsv1.DeploymentProgressing),
						Status:  corev1.ConditionTrue,
						Reason:  "NewReplicaSetAvailable",
						Message: "ReplicaSet \"depl1-5467849487\" has successfully progressed.",
					},
				},
			},
		},
	}

	TestPod = &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "some-pod-UID",
			Name:              "some-pod-5467849487-vqbfj",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.PodSpec{
			SchedulerName: constants.RunaiSchedulerName,
		},
	}
	TestPodExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			Type:              "Pod",
			ProjectName:       testProjectName,
			ProjectId:         testProjectId,
			CreationTimestamp: TestPod.CreationTimestamp,
		},
	}

	// TestDeploymentExpectedEw - simple expected EW for Deployment treated as top owner
	TestDeploymentExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			Type:        "Deployment",
			ProjectName: testProjectName,
			ProjectId:   testProjectId,
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:   v1alpha1.ExWUnknownPhase,
				Message: "Progressing condition type not found",
			},
		},
	}

	testPod2 = &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:       "pod-2-UID",
			Name:      "pod-2-5467849487-vqbfj",
			Namespace: TestNamespace,
			Labels: map[string]string{
				clusterconstants.LabelPreemptibilityName: "preemptible",
			},
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.PodSpec{
			SchedulerName:     constants.RunaiSchedulerName,
			PriorityClassName: "system-critical",
			Containers: []corev1.Container{
				{
					Image: testImage,
				},
			},
		},
		Status: corev1.PodStatus{
			Phase:   corev1.PodRunning,
			Message: "some-message",
			Reason:  "some-reason",
			Conditions: []corev1.PodCondition{
				{
					Type:    corev1.ContainersReady,
					Status:  corev1.ConditionTrue,
					Reason:  "some-reason",
					Message: "some-message",
				},
			},
		},
	}
	testPodExpectedEw2 = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			Type:              "Pod",
			ProjectName:       testProjectName,
			ProjectId:         testProjectId,
			CreationTimestamp: TestPod.CreationTimestamp,
			PriorityClassName: "system-critical",
			Preemptibility:    "preemptible",
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			ImageName: testImage,
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: "Running",
				Message:   "some-message",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(corev1.ContainersReady),
						Status:  corev1.ConditionTrue,
						Reason:  "some-reason",
						Message: "some-message",
					},
				},
			},
		},
	}
	testPodWithPodGroupAnnotation *corev1.Pod

	TestReplicaSet = &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "rs1-UID",
			Name:              "rs1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.ReplicaSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					SchedulerName: constants.RunaiSchedulerName,
					Containers: []corev1.Container{
						{
							Image: testImage,
						},
					},
				},
			},
		},
		Status: appsv1.ReplicaSetStatus{
			AvailableReplicas: 1,
			Conditions: []appsv1.ReplicaSetCondition{
				{
					Type:    appsv1.ReplicaSetReplicaFailure,
					Status:  corev1.ConditionFalse,
					Message: "msg1",
					Reason:  "reason1",
				},
			},
		},
	}
	testReplicaSetChild1Pod       *corev1.Pod
	testReplicaSetExpectedEw      *v1alpha1.ExternalWorkload
	TestReplicaSetExpectedBasicEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			ImageName:          testImage,
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: "Ready",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(appsv1.ReplicaSetReplicaFailure),
						Status:  corev1.ConditionFalse,
						Message: "msg1",
						Reason:  "reason1",
					},
				},
			},
		},
	}

	TestStatefulSet = &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "ss1-UID",
			Name:              "ss1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: ptr.To(int32(0)),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					SchedulerName: constants.RunaiSchedulerName,
					Containers: []corev1.Container{
						{
							Image: testImage,
						},
					},
				},
			},
		},
		Status: appsv1.StatefulSetStatus{
			AvailableReplicas: 0,
			ReadyReplicas:     0,
			Conditions: []appsv1.StatefulSetCondition{
				{
					Type:    appsv1.StatefulSetConditionType("Blabla"),
					Status:  corev1.ConditionFalse,
					Message: "msg2",
					Reason:  "reason12",
				},
			},
		},
	}
	testStatefulSetChild1Pod       *corev1.Pod
	testStatefulSetExpectedEw      *v1alpha1.ExternalWorkload
	TestStatefulSetExpectedBasicEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(0)),
			},
			ImageName:          testImage,
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
			RequestedResources: corev1.ResourceRequirements{},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWStoppedPhase,
				K8sStatus: "Suspended",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    "Blabla",
						Status:  corev1.ConditionFalse,
						Message: "msg2",
						Reason:  "reason12",
					},
				},
			},
		},
	}

	testNimStatefulSet = &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:       "ss1-UID-55",
			Name:      "ss1-3242355",
			Namespace: TestNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": "nemollm-inference",
			},
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: ptr.To(int32(0)),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					SchedulerName: constants.RunaiSchedulerName,
					Containers: []corev1.Container{
						{
							Image: testImage,
						},
					},
				},
			},
		},
	}
	testNimStatefulSetChild1Pod       *corev1.Pod
	testNimStatefulSetExpectedEw      *v1alpha1.ExternalWorkload
	TestNimStatefulSetExpectedBasicEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(0)),
			},
			ImageName:          testImage,
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
			RequestedResources: corev1.ResourceRequirements{},
			Type:               "NVIDIA-NIM",
			Group:              "apps",
			PriorityClassName:  "very-high",
			Preemptibility:     "preemptible",
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWStoppedPhase,
				K8sStatus: "Suspended",
			},
		},
	}

	testNimCache = &nim.NIMCache{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NIMCache",
			APIVersion: "apps.nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "nimcache-UID-55",
			Name:              "nimcache-3242355",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: upstreamnimv1alpha1.NIMCacheSpec{
			Resources: upstreamnimv1alpha1.Resources{
				CPU:    *resource.NewQuantity(100, resource.DecimalSI),
				Memory: *resource.NewQuantity(10, resource.DecimalSI),
			},
		},
		Status: upstreamnimv1alpha1.NIMCacheStatus{
			State: nim.NimCacheStatusReady,
			Conditions: []metav1.Condition{
				{
					Type:    nim.NimCacheConditionJobCreated,
					Status:  metav1.ConditionTrue,
					Reason:  "JobCreated",
					Message: "The Job to cache NIM was successfully created",
				},
				{
					Type:    nim.NimCacheConditionJobCompleted,
					Status:  metav1.ConditionTrue,
					Reason:  "JobCompleted",
					Message: "The Job to cache NIM has successfully completed",
				},
			},
		},
	}
	TestNimCacheExpectedBasicEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedResources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewQuantity(100, resource.DecimalSI),
					corev1.ResourceMemory: *resource.NewQuantity(10, resource.DecimalSI),
				},
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWCompletedPhase,
				K8sStatus: nim.NimCacheStatusReady,
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    nim.NimCacheConditionJobCreated,
						Status:  corev1.ConditionTrue,
						Reason:  "JobCreated",
						Message: "The Job to cache NIM was successfully created",
					},
					{
						Type:    nim.NimCacheConditionJobCompleted,
						Status:  corev1.ConditionTrue,
						Reason:  "JobCompleted",
						Message: "The Job to cache NIM has successfully completed",
					},
				},
			},
		},
	}
	testNimService = &nim.NIMService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NIMService",
			APIVersion: "apps.nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "nimservice-UID-55",
			Name:              "nimservice-3242355",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nim.NIMServiceSpec{
			NIMServiceSpec: upstreamnimv1alpha1.NIMServiceSpec{
				Replicas: 2,
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest, resource.DecimalSI),
					},
				},
			},
		},
		Status: upstreamnimv1alpha1.NIMServiceStatus{
			State: nim.NIMServiceStatusPending,
			Conditions: []metav1.Condition{
				{
					Type:   nim.NIMServiceStatusReady,
					Status: metav1.ConditionFalse,
					Reason: "Ready",
				},
			},
		},
	}
	TestNimServiceExpectedBasicEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(2)),
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest*2, resource.DecimalSI),
				},
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWPendingPhase,
				K8sStatus: nim.NIMServiceStatusPending,
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:   nim.NIMServiceStatusReady,
						Status: corev1.ConditionFalse,
						Reason: "Ready",
					},
				},
			},
		},
	}

	// Test case for NIMService with update in progress
	testNimServiceUpdating = &nim.NIMService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NIMService",
			APIVersion: "apps.nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "nimservice-UID-updating",
			Name:              "nimservice-updating",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nim.NIMServiceSpec{
			NIMServiceSpec: upstreamnimv1alpha1.NIMServiceSpec{
				Replicas: 5,
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest, resource.DecimalSI),
					},
				},
			},
		},
		Status: upstreamnimv1alpha1.NIMServiceStatus{
			State: nim.NIMServiceStatusNotReady,
			Conditions: []metav1.Condition{
				{
					Type:    nim.NIMServiceStatusReady,
					Status:  metav1.ConditionFalse,
					Reason:  "NotReady",
					Message: "Waiting for deployment \"nimservice-updating\" rollout to finish: 2 out of 5 new replicas have been updated...",
				},
			},
		},
	}

	TestNimServiceUpdatingExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(5)),
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest*5, resource.DecimalSI),
				},
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWInitializingPhase,
				K8sStatus: nim.NIMServiceStatusNotReady,
				Message:   "Update in progress: Waiting for deployment \"nimservice-updating\" rollout to finish: 2 out of 5 new replicas have been updated...",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    nim.NIMServiceStatusReady,
						Status:  corev1.ConditionFalse,
						Reason:  "NotReady",
						Message: "Waiting for deployment \"nimservice-updating\" rollout to finish: 2 out of 5 new replicas have been updated...",
					},
				},
			},
		},
	}

	// Test case for NIMService during initial creation (initializing)
	testNimServiceCreating = &nim.NIMService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NIMService",
			APIVersion: "apps.nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "nimservice-UID-creating",
			Name:              "nimservice-creating",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nim.NIMServiceSpec{
			NIMServiceSpec: upstreamnimv1alpha1.NIMServiceSpec{
				Replicas: 1,
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest, resource.DecimalSI),
					},
				},
			},
		},
		Status: upstreamnimv1alpha1.NIMServiceStatus{
			State: nim.NIMServiceStatusNotReady,
			Conditions: []metav1.Condition{
				{
					Type:    nim.NIMServiceStatusReady,
					Status:  metav1.ConditionFalse,
					Reason:  "NotReady",
					Message: "Waiting for deployment \"nimservice-creating\" rollout to finish: 0 of 1 updated replicas are available...",
				},
			},
		},
	}

	TestNimServiceCreatingExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest, resource.DecimalSI),
				},
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWInitializingPhase,
				K8sStatus: nim.NIMServiceStatusNotReady,
				Message:   "Waiting for deployment \"nimservice-creating\" rollout to finish: 0 of 1 updated replicas are available...",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    nim.NIMServiceStatusReady,
						Status:  corev1.ConditionFalse,
						Reason:  "NotReady",
						Message: "Waiting for deployment \"nimservice-creating\" rollout to finish: 0 of 1 updated replicas are available...",
					},
				},
			},
		},
	}

	// Test case for NIMService in Failed state with no message in conditions
	testNimServiceFailed = &nim.NIMService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NIMService",
			APIVersion: "apps.nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "nimservice-UID-failed",
			Name:              "nimservice-failed",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nim.NIMServiceSpec{
			NIMServiceSpec: upstreamnimv1alpha1.NIMServiceSpec{
				Replicas: 1,
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest, resource.DecimalSI),
					},
				},
			},
		},
		Status: upstreamnimv1alpha1.NIMServiceStatus{
			State: nim.NIMServiceStatusFailed,
			Conditions: []metav1.Condition{
				{
					Type:   nim.NIMServiceStatusReady,
					Status: metav1.ConditionFalse,
					Reason: "Ready",
					// No message - simulating ingress URL conflict where message is not set
				},
			},
		},
	}

	TestNimServiceFailedExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest, resource.DecimalSI),
				},
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWFailedPhase,
				K8sStatus: nim.NIMServiceStatusFailed,
				Message:   "Failed: Ready", // Now captures reason when message is empty
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:   nim.NIMServiceStatusReady,
						Status: corev1.ConditionFalse,
						Reason: "Ready",
					},
				},
			},
		},
	}

	// Test case for NIMService in Failed state with error message
	testNimServiceFailedWithMessage = &nim.NIMService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NIMService",
			APIVersion: "apps.nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "nimservice-UID-failed-msg",
			Name:              "nimservice-failed-msg",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nim.NIMServiceSpec{
			NIMServiceSpec: upstreamnimv1alpha1.NIMServiceSpec{
				Replicas: 1,
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest, resource.DecimalSI),
					},
				},
			},
		},
		Status: upstreamnimv1alpha1.NIMServiceStatus{
			State: nim.NIMServiceStatusFailed,
			Conditions: []metav1.Condition{
				{
					Type:    nim.NIMServiceStatusReady,
					Status:  metav1.ConditionFalse,
					Reason:  "IngressConflict",
					Message: "Ingress URL conflict: URL already used by another service",
				},
			},
		},
	}

	TestNimServiceFailedWithMessageExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest, resource.DecimalSI),
				},
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWFailedPhase,
				K8sStatus: nim.NIMServiceStatusFailed,
				Message:   "Ingress URL conflict: URL already used by another service",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    nim.NIMServiceStatusReady,
						Status:  corev1.ConditionFalse,
						Reason:  "IngressConflict",
						Message: "Ingress URL conflict: URL already used by another service",
					},
				},
			},
		},
	}

	// Test cases for NIMService with multiNode configuration
	// These use our local nim types which have MultiNode support
	testNimServiceMultiNode = &nim.NIMService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NIMService",
			APIVersion: "apps.nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "nimservice-UID-multinode",
			Name:              "nimservice-multinode",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nim.NIMServiceSpec{
			NIMServiceSpec: upstreamnimv1alpha1.NIMServiceSpec{
				Replicas: 2, // 2 replicas * 4 nodes per replica = 8 total pods
				Image: upstreamnimv1alpha1.Image{
					Repository: "nvcr.io/nvidia/nim",
					Tag:        "latest",
				},
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						constants.GpuResourceName: *resource.NewQuantity(1, resource.DecimalSI),
					},
				},
			},
			MultiNode: &nim.MultiNodeConfig{
				Parallelism: &nim.ParallelismSpec{
					Pipeline: ptr.To(4), // 1 leader + 3 workers = 4 nodes per replica
					Tensor:   ptr.To(8),
				},
			},
		},
		Status: upstreamnimv1alpha1.NIMServiceStatus{
			State: nim.NIMServiceStatusReady,
			Conditions: []metav1.Condition{
				{
					Type:   nim.NIMServiceStatusReady,
					Status: metav1.ConditionTrue,
					Reason: "Ready",
				},
			},
		},
	}

	// Expected EW for multiNode NIMService: 2 replicas * 4 nodes = 8 total pods
	// Resources are multiplied by total pods
	TestNimServiceMultiNodeExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(8)), // 2 replicas * 4 nodes per replica
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(8, resource.DecimalSI), // 8 pods * 1 GPU each
				},
			},
			ImageName: "nvcr.io/nvidia/nim:latest",
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: nim.NIMServiceStatusReady,
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:   nim.NIMServiceStatusReady,
						Status: corev1.ConditionTrue,
						Reason: "Ready",
					},
				},
			},
		},
	}

	// Test case for single replica multiNode NIMService
	testNimServiceMultiNodeSingleReplica = &nim.NIMService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NIMService",
			APIVersion: "apps.nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "nimservice-UID-multinode-single",
			Name:              "nimservice-multinode-single",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nim.NIMServiceSpec{
			NIMServiceSpec: upstreamnimv1alpha1.NIMServiceSpec{
				Replicas: 1, // 1 replica * 2 nodes per replica = 2 total pods
				Image: upstreamnimv1alpha1.Image{
					Repository: "nvcr.io/nvidia/nim",
					Tag:        "v1.0",
				},
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						constants.GpuResourceName: *resource.NewQuantity(2, resource.DecimalSI),
					},
				},
			},
			MultiNode: &nim.MultiNodeConfig{
				Parallelism: &nim.ParallelismSpec{
					Pipeline: ptr.To(2), // 1 leader + 1 worker = 2 nodes per replica
					Tensor:   ptr.To(4),
				},
			},
		},
		Status: upstreamnimv1alpha1.NIMServiceStatus{
			State: nim.NIMServiceStatusPending,
			Conditions: []metav1.Condition{
				{
					Type:   nim.NIMServiceStatusReady,
					Status: metav1.ConditionFalse,
					Reason: "Pending",
				},
			},
		},
	}

	// Expected EW for single replica multiNode NIMService: 1 replica * 2 nodes = 2 total pods
	TestNimServiceMultiNodeSingleReplicaExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(2)), // 1 replica * 2 nodes per replica
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(4, resource.DecimalSI), // 2 pods * 2 GPUs each
				},
			},
			ImageName: "nvcr.io/nvidia/nim:v1.0",
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWPendingPhase,
				K8sStatus: nim.NIMServiceStatusPending,
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:   nim.NIMServiceStatusReady,
						Status: corev1.ConditionFalse,
						Reason: "Pending",
					},
				},
			},
		},
	}

	// Test case for NIMService without multiNode (default behavior)
	testNimServiceNoMultiNode = &nim.NIMService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NIMService",
			APIVersion: "apps.nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "nimservice-UID-no-multinode",
			Name:              "nimservice-no-multinode",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nim.NIMServiceSpec{
			NIMServiceSpec: upstreamnimv1alpha1.NIMServiceSpec{
				Replicas: 3, // 3 replicas * 1 (no multiNode) = 3 total pods
				Image: upstreamnimv1alpha1.Image{
					Repository: "nvcr.io/nvidia/nim",
					Tag:        "v2.0",
				},
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						constants.GpuResourceName: *resource.NewQuantity(1, resource.DecimalSI),
					},
				},
			},
			// No MultiNode field - should default to 1 node per replica
		},
		Status: upstreamnimv1alpha1.NIMServiceStatus{
			State: nim.NIMServiceStatusReady,
			Conditions: []metav1.Condition{
				{
					Type:   nim.NIMServiceStatusReady,
					Status: metav1.ConditionTrue,
					Reason: "Ready",
				},
			},
		},
	}

	// Expected EW for NIMService without multiNode: 3 replicas * 1 = 3 total pods
	TestNimServiceNoMultiNodeExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(3)), // 3 replicas * 1 node per replica (no multiNode)
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(3, resource.DecimalSI), // 3 pods * 1 GPU each
				},
			},
			ImageName: "nvcr.io/nvidia/nim:v2.0",
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: nim.NIMServiceStatusReady,
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:   nim.NIMServiceStatusReady,
						Status: corev1.ConditionTrue,
						Reason: "Ready",
					},
				},
			},
		},
	}

	// DRA test fixtures

	// ResourceClaimTemplate for DRA GPU tests
	TestDRAGpuClaimTemplate = &resourcev1.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dra-gpu-template",
			Namespace: TestNamespace,
		},
		Spec: resourcev1.ResourceClaimTemplateSpec{
			Spec: resourcev1.ResourceClaimSpec{
				Devices: resourcev1.DeviceClaim{
					Requests: []resourcev1.DeviceRequest{
						{
							Name: "gpu",
							Exactly: &resourcev1.ExactDeviceRequest{
								DeviceClassName: "gpu.nvidia.com",
								AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
								Count:           4,
							},
						},
					},
				},
			},
		},
	}

	// ResourceClaim for DRA GPU tests
	TestDRAGpuClaim = &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dra-gpu-claim",
			Namespace: TestNamespace,
		},
		Spec: resourcev1.ResourceClaimSpec{
			Devices: resourcev1.DeviceClaim{
				Requests: []resourcev1.DeviceRequest{
					{
						Name: "gpu",
						Exactly: &resourcev1.ExactDeviceRequest{
							DeviceClassName: "gpu.nvidia.com",
							AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
							Count:           2,
						},
					},
				},
			},
		},
	}

	// Deployment with DRA claims (no traditional GPU resources)
	TestDeploymentWithDRA = &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "depl-dra-UID",
			Name:              "depl-dra",
			Namespace:         TestNamespace,
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
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:    appsv1.DeploymentProgressing,
					Status:  corev1.ConditionTrue,
					Reason:  "NewReplicaSetAvailable",
					Message: "ReplicaSet \"depl-dra-abc\" has successfully progressed.",
				},
			},
		},
	}

	// Expected EW for Deployment with DRA: 2 replicas * 4 GPUs = 8 GPUs total
	TestDeploymentWithDRAExpectedEw = &v1alpha1.ExternalWorkload{
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
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: "Complete",
				Message:   "ReplicaSet \"depl-dra-abc\" has successfully progressed.",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(appsv1.DeploymentProgressing),
						Status:  corev1.ConditionTrue,
						Reason:  "NewReplicaSetAvailable",
						Message: "ReplicaSet \"depl-dra-abc\" has successfully progressed.",
					},
				},
			},
		},
	}

	// NIMService with DRA ResourceClaim
	TestNimServiceWithDRAClaim = &nim.NIMService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NIMService",
			APIVersion: "apps.nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "nimservice-UID-dra-claim",
			Name:              "nimservice-dra-claim",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nim.NIMServiceSpec{
			NIMServiceSpec: upstreamnimv1alpha1.NIMServiceSpec{
				Replicas: 3,
				Image: upstreamnimv1alpha1.Image{
					Repository: "nvcr.io/nvidia/nim",
					Tag:        "dra",
				},
			},
			DRAResources: []nim.DRAResource{
				{
					ResourceClaimName: ptr.To("dra-gpu-claim"),
				},
			},
		},
		Status: upstreamnimv1alpha1.NIMServiceStatus{
			State: nim.NIMServiceStatusReady,
			Conditions: []metav1.Condition{
				{
					Type:   nim.NIMServiceStatusReady,
					Status: metav1.ConditionTrue,
					Reason: "Ready",
				},
			},
		},
	}

	// Expected EW for NIMService with DRA claim: 3 replicas * 2 GPUs = 6 GPUs total
	TestNimServiceWithDRAClaimExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(3)),
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(6, resource.DecimalSI), // 3 replicas * 2 GPUs
				},
				Requests: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(6, resource.DecimalSI),
				},
			},
			ImageName: "nvcr.io/nvidia/nim:dra",
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: nim.NIMServiceStatusReady,
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:   nim.NIMServiceStatusReady,
						Status: corev1.ConditionTrue,
						Reason: "Ready",
					},
				},
			},
		},
	}

	// NIMService with DRA ResourceClaimTemplate
	TestNimServiceWithDRATemplate = &nim.NIMService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NIMService",
			APIVersion: "apps.nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "nimservice-UID-dra-template",
			Name:              "nimservice-dra-template",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nim.NIMServiceSpec{
			NIMServiceSpec: upstreamnimv1alpha1.NIMServiceSpec{
				Replicas: 2,
				Image: upstreamnimv1alpha1.Image{
					Repository: "nvcr.io/nvidia/nim",
					Tag:        "dra-tmpl",
				},
			},
			DRAResources: []nim.DRAResource{
				{
					ResourceClaimTemplateName: ptr.To("dra-gpu-template"),
				},
			},
		},
		Status: upstreamnimv1alpha1.NIMServiceStatus{
			State: nim.NIMServiceStatusReady,
			Conditions: []metav1.Condition{
				{
					Type:   nim.NIMServiceStatusReady,
					Status: metav1.ConditionTrue,
					Reason: "Ready",
				},
			},
		},
	}

	// Expected EW for NIMService with DRA template: 2 replicas * 4 GPUs = 8 GPUs total
	TestNimServiceWithDRATemplateExpectedEw = &v1alpha1.ExternalWorkload{
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
			ImageName: "nvcr.io/nvidia/nim:dra-tmpl",
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: nim.NIMServiceStatusReady,
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:   nim.NIMServiceStatusReady,
						Status: corev1.ConditionTrue,
						Reason: "Ready",
					},
				},
			},
		},
	}

	// NIMService requesting a fractional GPU via run:ai annotations, with no nvidia.com/gpu
	// in spec.resources.
	TestNimServiceWithFractionalGpu = &nim.NIMService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NIMService",
			APIVersion: "apps.nvidia.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "nimservice-UID-fractional-gpu",
			Name:              "nimservice-fractional-gpu",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nim.NIMServiceSpec{
			NIMServiceSpec: upstreamnimv1alpha1.NIMServiceSpec{
				Replicas: 1,
				Image: upstreamnimv1alpha1.Image{
					Repository: "nvcr.io/nim/nvidia/nv-embedqa-e5-v5",
					Tag:        "1.5.1",
				},
				Annotations: map[string]string{
					"gpu-fraction": "0.5",
				},
			},
		},
		Status: upstreamnimv1alpha1.NIMServiceStatus{
			State: nim.NIMServiceStatusReady,
			Conditions: []metav1.Condition{
				{
					Type:   nim.NIMServiceStatusReady,
					Status: metav1.ConditionTrue,
					Reason: "Ready",
				},
			},
		},
	}

	// Expected EW for the fractional-GPU NIMService: 0.5 GPU requested via annotation, 1 replica.
	TestNimServiceWithFractionalGpuExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: resource.MustParse("0.5"),
				},
				Requests: corev1.ResourceList{
					constants.GpuResourceName: resource.MustParse("0.5"),
				},
			},
			ImageName: "nvcr.io/nim/nvidia/nv-embedqa-e5-v5:1.5.1",
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: nim.NIMServiceStatusReady,
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:   nim.NIMServiceStatusReady,
						Status: corev1.ConditionTrue,
						Reason: "Ready",
					},
				},
			},
		},
	}

	TestNvcfFunction = &nvcfv1.SPOTRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SPOTRequest",
			APIVersion: "egx.nvidia.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "sr-uid-123",
			Name:              "sr-name-123",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nvcfv1.SPOTRequestSpec{
			FunctionID:        "some-function-id",
			FunctionVersionID: "some-function-version-id",
		},
		Status: nvcfv1.SPOTRequestStatus{},
	}
	TestNvcfFunctionChildPod        *corev1.Pod
	testNvcfFunctionExpectedEw      *v1alpha1.ExternalWorkload
	TestNvcfFunctionExpectedBasicEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			Type:                 "NVCF Function",
			OriginalWorkloadUID:  "some-function-version-id",                    // important! this is unique to nvcf handler - the id is not taken from the original object
			OriginalWorkloadName: "some-function-name-some-function-version-id", // important! this is unique to nvcf handler - the name is not taken from the original object
			PriorityClassName:    priority.InferencePriorityClassName,
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest, resource.DecimalSI),
				},
			},
			Urls:             []string{"https://my.domain.com/v2/nvcf/pexec/functions/some-function-id/versions/some-function-version-id"},
			AdditionalFields: ptr.To("{\"function-id\":\"some-function-id\",\"function-version-id\":\"some-function-version-id\",\"function-version-name\":\"some-function-name\"}"),
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase: v1alpha1.ExWPendingPhase,
			},
		},
	}

	TestBatchJob = &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "job1-UID",
			Name:              "job1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: batchv1.JobSpec{
			Parallelism: ptr.To(int32(2)),
			Completions: ptr.To(int32(4)),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					SchedulerName: constants.RunaiSchedulerName,
					Containers: []corev1.Container{
						{
							Image: testImage,
						},
					},
				},
			},
		},
		Status: batchv1.JobStatus{
			Active: 1,
			Ready:  ptr.To(int32(1)),
			Conditions: []batchv1.JobCondition{
				{
					Type:    batchv1.JobFailureTarget,
					Status:  corev1.ConditionFalse,
					Message: "msg3",
					Reason:  "reason3",
				},
				{
					Type:    batchv1.JobFailed,
					Status:  corev1.ConditionFalse,
					Message: "msg4",
					Reason:  "reason4",
				},
			},
		},
	}
	testBatchJobChild1Pod       *corev1.Pod
	testBatchJobChild2Pod       *corev1.Pod
	testBatchJobExpectedEw      *v1alpha1.ExternalWorkload
	TestBatchJobExpectedBasicEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Parallelism: ptr.To(int32(2)),
				Completions: ptr.To(int32(4)),
			},
			ImageName:          testImage,
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: "Active",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(batchv1.JobFailureTarget),
						Status:  corev1.ConditionFalse,
						Message: "msg3",
						Reason:  "reason3",
					},
					{
						Type:    string(batchv1.JobFailed),
						Status:  corev1.ConditionFalse,
						Message: "msg4",
						Reason:  "reason4",
					},
				},
			},
		},
	}

	TestCronJob = &batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "cronjob1-UID",
			Name:              "cronjob1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: batchv1.CronJobSpec{
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: constants.RunaiSchedulerName,
							Containers: []corev1.Container{
								{
									Image: testImage,
								},
							},
						},
					},
				},
			},
		},
		Status: batchv1.CronJobStatus{
			Active: []corev1.ObjectReference{
				{
					Kind: "Job",
					Name: TestBatchJob.Name,
					UID:  TestBatchJob.UID,
				},
			},
		},
	}
	testCronJobChild1Job       *batchv1.Job
	testCronJobChild2Pod       *corev1.Pod
	testCronJobExpectedEw      *v1alpha1.ExternalWorkload
	TestCronJobExpectedBasicEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: nil,
			},
			ImageName:          testImage,
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: "Active",
			},
		},
	}

	TestRayJob = &rayv1alpha1.RayJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RayJob",
			APIVersion: "ray.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "rayjob1-UID",
			Name:              "rayjob1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: rayv1alpha1.RayJobSpec{
			RayClusterSpec: &rayv1alpha1.RayClusterSpec{
				HeadGroupSpec: rayv1alpha1.HeadGroupSpec{
					Template: TestDeployment.Spec.Template,
				},
				WorkerGroupSpecs: []rayv1alpha1.WorkerGroupSpec{
					{
						GroupName:   "worker-group-1",
						MinReplicas: ptr.To(int32(2)),
						MaxReplicas: ptr.To(int32(4)),
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								SchedulerName: constants.RunaiSchedulerName,
								Containers: []corev1.Container{
									{
										Image: testImage,
										Resources: corev1.ResourceRequirements{
											Limits: corev1.ResourceList{
												corev1.ResourceCPU: *resource.NewQuantity(testDeploymentCpuLimit, resource.DecimalSI),
											},
											Requests: corev1.ResourceList{
												corev1.ResourceMemory: *resource.NewQuantity(testDeploymentMemoryRequest, resource.DecimalSI),
											},
										},
									},
								},
							},
						},
					},
					{
						GroupName:   "worker-group-2",
						MinReplicas: ptr.To(int32(3)),
						MaxReplicas: ptr.To(int32(5)),
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								SchedulerName: constants.RunaiSchedulerName,
								Containers: []corev1.Container{
									{
										Image: testImage,
										Resources: corev1.ResourceRequirements{
											Limits: corev1.ResourceList{
												corev1.ResourceMemory: *resource.NewQuantity(testDeploymentMemoryLimit, resource.DecimalSI),
											},
											Requests: corev1.ResourceList{
												corev1.ResourceCPU: *resource.NewQuantity(testDeploymentCpuRequest, resource.DecimalSI),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Status: rayv1alpha1.RayJobStatus{
			JobStatus:           rayv1alpha1.JobStatusStopped,
			JobDeploymentStatus: rayv1alpha1.JobDeploymentStatusSuspended,
			Message:             "some-message",
			RayClusterStatus: rayv1alpha1.RayClusterStatus{
				State:  rayv1alpha1.Unhealthy,
				Reason: "some-reason",
			},
		},
	}
	TestRayJobV1   = buildRayJobV1(TestRayJob)
	TestRayCluster = &rayv1alpha1.RayCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RayCluster",
			APIVersion: "ray.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "raycluster1-UID",
			Name:              "raycluster1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: rayv1alpha1.RayClusterSpec{
			HeadGroupSpec: rayv1alpha1.HeadGroupSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						SchedulerName: constants.RunaiSchedulerName,
						Containers: []corev1.Container{
							{
								Image: testImage,
							},
						},
					},
				},
			},
		},
		Status: rayv1alpha1.RayClusterStatus{
			State:  rayv1alpha1.Ready,
			Reason: "some-reason2",
		},
	}
	TestRayClusterV1 = buildRayClusterV1(TestRayCluster)
	TestRayService   = &rayv1alpha1.RayService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RayService",
			APIVersion: "ray.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "rayservice1-UID",
			Name:              "rayservice1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: rayv1alpha1.RayServiceSpec{
			RayClusterSpec: rayv1alpha1.RayClusterSpec{
				HeadGroupSpec: rayv1alpha1.HeadGroupSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: constants.RunaiSchedulerName,
						},
					},
				},
			},
		},
		Status: rayv1alpha1.RayServiceStatuses{
			ServiceStatus: "Failed",
			ActiveServiceStatus: rayv1alpha1.RayServiceStatus{
				RayClusterStatus: rayv1alpha1.RayClusterStatus{
					State:  rayv1alpha1.Failed,
					Reason: "some-reason3",
				},
			},
		},
	}
	TestRayServiceV1          = buildRayServiceV1(TestRayService)
	testRayJobChildRayCluster *rayv1alpha1.RayCluster
	testRayJobChild1Pod       *corev1.Pod
	testRayJobChild2Pod       *corev1.Pod
	testRayClusterChildPod    *corev1.Pod
	TestRayJobExpectedEw      = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas:    ptr.To(int32(1)),
				MinReplicas: ptr.To(int32(5)),
				MaxReplicas: ptr.To(int32(9)),
			},
			ImageName:          testImage,
			PriorityClassName:  testDeploymentPriority,
			Preemptibility:     "preemptible",
			RequestedNodePools: []string{npAName, v1alpha1.DefaultNodePoolName},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:              *resource.NewQuantity(testDeploymentCpuLimit+testDeploymentCpuLimit*2, resource.DecimalSI),
					corev1.ResourceMemory:           *resource.NewQuantity(testDeploymentMemoryLimit+testDeploymentMemoryLimit*3, resource.DecimalSI),
					constants.GpuResourceName:       *resource.NewMilliQuantity(1000*(testDeploymentGpuLimit+testDeploymentFracGpuLimit), resource.DecimalSI),
					constants.GpuMemoryResourceName: *resource.NewQuantity(GpuMemoryLimitConversion*testDeploymentGpuMemoryLimit, resource.DecimalSI),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:              *resource.NewQuantity(testDeploymentCpuRequest+testDeploymentCpuRequest*3, resource.DecimalSI),
					corev1.ResourceMemory:           *resource.NewQuantity(testDeploymentMemoryRequest+testDeploymentMemoryRequest*2, resource.DecimalSI),
					constants.GpuResourceName:       *resource.NewMilliQuantity(1000*(testDeploymentGpuRequest+testDeploymentFracGpuLimit), resource.DecimalSI),
					constants.GpuMemoryResourceName: *resource.NewQuantity(GpuMemoryLimitConversion*testDeploymentGpuMemoryLimit, resource.DecimalSI),
				},
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWDegradedPhase,
				K8sStatus: string(rayv1alpha1.Unhealthy),
				Message:   "some-reason",
			},
		},
	}
	TestRayClusterExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			ImageName:          testImage,
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
			PriorityClassName:  "low",
			Preemptibility:     "preemptible",
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: string(rayv1alpha1.Ready),
				Message:   "some-reason2",
			},
		},
	}
	TestRayServiceExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWFailedPhase,
				K8sStatus: string(rayv1alpha1.Failed),
				Message:   "some-reason3",
			},
		},
	}

	TestKnativeService = &servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:       "knative-service-UID",
			Name:      "knative-service-3242344",
			Namespace: TestNamespace,
			Labels: map[string]string{
				commonconsts.ControlPlaneWorkloadNameLabelForWorkload: "some-workload",
				commonconsts.ControlPlaneWorkloadKindLabelForWorkload: "NotInf",
				constants.ManagedByLabelKey:                           "Helm",
				"otherKey":                                            "otherVal",
				constants.DynamoGraphDeploymentNameLabelKey:           "dynamograph-deployment-name",
			},
			Annotations: map[string]string{
				commonconsts.ControlPlaneWorkloadIdAnnotationForWorkload: "runai-inference-workload-id",
				clusterconstants.AnnotationJobUrl:                        "https://env.example.com/team-a/test, https://env.example.com/team-a/test/url-2",
				constants.HelmReleaseAnnotationKey:                       "rel",
				"otherKey":                                               "otherVal",
			},
			CreationTimestamp: metav1.Now(),
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"autoscaling.knative.dev/min-scale": "2",
						},
					},
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							SchedulerName: constants.RunaiSchedulerName,
							Containers: []corev1.Container{
								{
									Image: testImage,
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuLimit, resource.DecimalSI),
											corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryRequest, resource.DecimalSI),
											constants.GpuResourceName: *resource.NewMilliQuantity(1000*testDeploymentGpuLimit, resource.DecimalSI),
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuRequest, resource.DecimalSI),
											corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryRequest, resource.DecimalSI),
											constants.GpuResourceName: *resource.NewMilliQuantity(1000*testDeploymentGpuRequest, resource.DecimalSI),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Status: servingv1.ServiceStatus{
			Status: duckv1.Status{
				Conditions: duckv1.Conditions{
					{
						Type:   servingv1.ServiceConditionReady,
						Status: corev1.ConditionTrue,
						Reason: "some-reason",
					},
				},
			},
			RouteStatusFields: servingv1.RouteStatusFields{
				URL: &apis.URL{
					Scheme: "http",
					Host:   "depl1.runai-test.svc.cluster.local",
				},
			},
		},
	}
	TestKnativeServiceInference   = TestKnativeService.DeepCopy()
	testKnativeChildConfiguration = &servingv1.Configuration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Configuration",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:       "knative-configuration-UID",
			Name:      "knative-configuration-3242344",
			Namespace: TestNamespace,
		},
	}
	testKnativeChildRevision = &servingv1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:       "knative-revision-UID",
			Name:      "knative-revision-3242344",
			Namespace: TestNamespace,
		},
	}
	TestKnativeServiceExpectedEw = &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				constants.ManagedByLabelKey:                 "Helm",
				constants.DynamoGraphDeploymentNameLabelKey: "dynamograph-deployment-name",
			},
			Annotations: map[string]string{
				commonconsts.ControlPlaneWorkloadIdAnnotationForWorkload: "runai-inference-workload-id",
				constants.HelmReleaseAnnotationKey:                       "rel",
			},
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas:    nil,
				MinReplicas: ptr.To(int32(2)),
				MaxReplicas: nil,
			},
			ImageName:          testImage,
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(2*testDeploymentCpuLimit, resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(2*testDeploymentMemoryRequest, resource.DecimalSI),
					constants.GpuResourceName: *resource.NewMilliQuantity(2*1000*testDeploymentGpuLimit, resource.DecimalSI),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(2*testDeploymentCpuRequest, resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(2*testDeploymentMemoryRequest, resource.DecimalSI),
					constants.GpuResourceName: *resource.NewMilliQuantity(2*1000*testDeploymentGpuRequest, resource.DecimalSI),
				},
			},
			Urls: []string{"https://env.example.com/team-a/test, https://env.example.com/team-a/test/url-2", "http://depl1.runai-test.svc.cluster.local"},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: "Ready",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:   string(servingv1.ServiceConditionReady),
						Status: corev1.ConditionTrue,
						Reason: "some-reason",
					},
				},
			},
		},
	}
	TestKnativeServiceInferenceExpectedEw = TestKnativeServiceExpectedEw.DeepCopy()

	TestSeldonDeployment1 = &seldonv1.SeldonDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SeldonDeployment",
			APIVersion: "machinelearning.seldon.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "seldondeploy1-UID",
			Name:              "seldondeploy1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: seldonv1.SeldonDeploymentSpec{
			Replicas: ptr.To(int32(2)),
			Predictors: []seldonv1.PredictorSpec{
				{
					ComponentSpecs: []*seldonv1.SeldonPodSpec{
						{
							Replicas: ptr.To(int32(1)),
							HpaSpec: &seldonv1.SeldonHpaSpec{
								MinReplicas: ptr.To(int32(1)),
							},
							Spec: corev1.PodSpec{
								SchedulerName: constants.RunaiSchedulerName,
								Containers: []corev1.Container{
									{
										Image: testImage,
										Resources: corev1.ResourceRequirements{
											Limits: corev1.ResourceList{
												corev1.ResourceCPU: *resource.NewQuantity(testDeploymentCpuLimit, resource.DecimalSI),
											},
										},
									},
								},
							},
						},
						{
							Spec: corev1.PodSpec{
								SchedulerName: constants.RunaiSchedulerName,
								Containers: []corev1.Container{
									{
										Image: testImage,
										Resources: corev1.ResourceRequirements{
											Limits: corev1.ResourceList{
												corev1.ResourceCPU: *resource.NewQuantity(testDeploymentCpuLimit, resource.DecimalSI),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Status: seldonv1.SeldonDeploymentStatus{
			State:       seldonv1.StatusStateAvailable,
			Description: "some-description",
			Status: duckv1.Status{
				Conditions: duckv1.Conditions{
					{
						Type:    apis.ConditionReady,
						Status:  corev1.ConditionTrue,
						Message: "some-message",
						Reason:  "some-reason",
					},
				},
			},
		},
	}
	TestSeldonDeployment2 = &seldonv1alpha2.SeldonDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SeldonDeployment",
			APIVersion: "machinelearning.seldon.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "seldondeploy2-UID",
			Name:              "seldondeploy2-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec:   TestSeldonDeployment1.Spec,
		Status: TestSeldonDeployment1.Status,
	}
	TestSeldonDeploymentExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas:    ptr.To(int32(3)),
				MinReplicas: ptr.To(int32(1)),
				MaxReplicas: nil,
			},
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU: *resource.NewQuantity(testDeploymentCpuLimit*3, resource.DecimalSI),
				},
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: string(seldonv1.StatusStateAvailable),
				Message:   "some-description",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(apis.ConditionReady),
						Status:  corev1.ConditionTrue,
						Message: "some-message",
						Reason:  "some-reason",
					},
				},
			},
		},
	}

	// kubeflow
	TestTFJob1 = &kubeflowv1.TFJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TFJob",
			APIVersion: "kubeflow.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "tfjob1-UID",
			Name:              "tfjob1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
			Annotations: map[string]string{
				commonconsts.ControlPlaneWorkloadIdAnnotationForWorkload: "tfjob-workloadid",
			},
		},
		Spec: kubeflowv1.TFJobSpec{
			TFReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
				"PS": {
					Replicas: ptr.To(int32(2)),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: constants.RunaiSchedulerName,
							Containers: []corev1.Container{
								{
									Image: testImage,
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuLimit, resource.DecimalSI),
											corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryRequest, resource.DecimalSI),
											constants.GpuResourceName: *resource.NewMilliQuantity(1000*testDeploymentGpuLimit, resource.DecimalSI),
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuRequest, resource.DecimalSI),
											corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryRequest, resource.DecimalSI),
											constants.GpuResourceName: *resource.NewMilliQuantity(1000*testDeploymentGpuRequest, resource.DecimalSI),
										},
									},
								},
							},
						},
					},
				},
				"Worker": {
					Replicas: ptr.To(int32(1)),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: constants.RunaiSchedulerName,
							Containers: []corev1.Container{
								{
									Image: testImage,
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuLimit, resource.DecimalSI),
											corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryRequest, resource.DecimalSI),
											constants.GpuResourceName: *resource.NewMilliQuantity(1000*testDeploymentGpuLimit, resource.DecimalSI),
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuRequest, resource.DecimalSI),
											corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryRequest, resource.DecimalSI),
											constants.GpuResourceName: *resource.NewMilliQuantity(1000*testDeploymentGpuRequest, resource.DecimalSI),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Status: kubeflowv1.JobStatus{
			Conditions: []kubeflowv1.JobCondition{
				{
					Type:    kubeflowv1.JobRunning,
					Status:  corev1.ConditionTrue,
					Message: "some-messagee",
				},
			},
		},
	}
	TestPytorchJob1 = &kubeflowv1.PyTorchJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PyTorchJob",
			APIVersion: "kubeflow.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "pytorchjob1-UID",
			Name:              "pytorchjob1-3242344",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
			Labels: map[string]string{
				commonconsts.WorkloadCategoryLabel: "Train",
				constants.ManagedByLabelKey:        "Helm",
			},
			Annotations: map[string]string{
				constants.HelmReleaseAnnotationKey: "rel",
				"otherKey":                         "val",
			},
		},
		Spec: kubeflowv1.PyTorchJobSpec{
			PyTorchReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
				"Master": {
					Replicas: ptr.To(int32(3)),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: constants.RunaiSchedulerName,
							Containers: []corev1.Container{
								{
									Image: testImage,
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuLimit, resource.DecimalSI),
											corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryRequest, resource.DecimalSI),
											constants.GpuResourceName: *resource.NewMilliQuantity(1000*testDeploymentGpuLimit, resource.DecimalSI),
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuRequest, resource.DecimalSI),
											corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryRequest, resource.DecimalSI),
											constants.GpuResourceName: *resource.NewMilliQuantity(1000*testDeploymentGpuRequest, resource.DecimalSI),
										},
									},
								},
							},
						},
					},
				},
				"Worker": {
					Replicas: ptr.To(int32(2)),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: constants.RunaiSchedulerName,
							Containers: []corev1.Container{
								{
									Image: testImage,
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:        *resource.NewQuantity(2*testDeploymentCpuLimit, resource.DecimalSI),
											corev1.ResourceMemory:     *resource.NewQuantity(2*testDeploymentMemoryRequest, resource.DecimalSI),
											constants.GpuResourceName: *resource.NewMilliQuantity(2*1000*testDeploymentGpuLimit, resource.DecimalSI),
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:        *resource.NewQuantity(2*testDeploymentCpuRequest, resource.DecimalSI),
											corev1.ResourceMemory:     *resource.NewQuantity(2*testDeploymentMemoryRequest, resource.DecimalSI),
											constants.GpuResourceName: *resource.NewMilliQuantity(2*1000*testDeploymentGpuRequest, resource.DecimalSI),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Status: kubeflowv1.JobStatus{
			Conditions: []kubeflowv1.JobCondition{
				{
					Type:    kubeflowv1.JobSucceeded,
					Status:  corev1.ConditionTrue,
					Message: "some-messagee1",
				},
				{
					Type:    kubeflowv1.JobCreated,
					Status:  corev1.ConditionTrue,
					Message: "some-messagee2",
				},
			},
		},
	}
	TestXGBoostJob1 = &kubeflowv1.XGBoostJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "XGBoostJob",
			APIVersion: "kubeflow.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "xgboostjob1-UID",
			Name:              "xgboostjob1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubeflowv1.XGBoostJobSpec{
			XGBReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
				"Master": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: constants.RunaiSchedulerName,
							Containers: []corev1.Container{
								{
									Image: testImage,
								},
							},
						},
					},
				},
			},
		},
		Status: kubeflowv1.JobStatus{
			Conditions: []kubeflowv1.JobCondition{
				{
					Type:    kubeflowv1.JobRestarting,
					Status:  corev1.ConditionTrue,
					Message: "some-messagee3",
				},
			},
		},
	}
	TestJAXJob1 = &kubeflowv1.JAXJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "JAXJob",
			APIVersion: "kubeflow.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "jaxjob1-UID",
			Name:              "jaxjob1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubeflowv1.JAXJobSpec{
			JAXReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
				"Master": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: constants.RunaiSchedulerName,
							Containers: []corev1.Container{
								{
									Image: testImage,
								},
							},
						},
					},
				},
			},
		},
		Status: kubeflowv1.JobStatus{
			Conditions: []kubeflowv1.JobCondition{
				{
					Type:    kubeflowv1.JobRunning,
					Status:  corev1.ConditionTrue,
					Message: "some-messagee3",
				},
			},
		},
	}
	TestMPIJobV1 = &kubeflowv1.MPIJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MPIJob",
			APIVersion: "kubeflow.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "mpijob1-UID",
			Name:              "mpijob1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubeflowv1.MPIJobSpec{
			MPIReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
				"Launcher": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: constants.RunaiSchedulerName,
							Containers: []corev1.Container{
								{
									Image: testImage,
								},
							},
						},
					},
				},
			},
		},
		Status: kubeflowv1.JobStatus{
			Conditions: []kubeflowv1.JobCondition{
				{
					Type:    kubeflowv1.JobFailed,
					Status:  corev1.ConditionTrue,
					Message: "some-messagee4",
				},
			},
		},
	}
	TestMPIJobV2beta1 = &kubeflowv2beta1.MPIJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MPIJob",
			APIVersion: "kubeflow.org/v2beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "mpijob2-UID",
			Name:              "mpijob2-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubeflowv2beta1.MPIJobSpec{
			MPIReplicaSpecs: map[kubeflowv2beta1.MPIReplicaType]*kubeflowv2beta1.ReplicaSpec{
				"Launcher": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: constants.RunaiSchedulerName,
							Containers: []corev1.Container{
								{
									Image: testImage,
								},
							},
						},
					},
				},
			},
		},
		Status: kubeflowv2beta1.JobStatus{
			Conditions: []kubeflowv2beta1.JobCondition{
				{
					Type:    kubeflowv2beta1.JobSuspended,
					Status:  corev1.ConditionTrue,
					Message: "some-messagee5",
				},
			},
		},
	}
	TestNotebook1 = &kubeflowv1beta1.Notebook{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Notebook",
			APIVersion: "kubeflow.org/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "notebook1-UID",
			Name:              "notebook1-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubeflowv1beta1.NotebookSpec{
			Template: kubeflowv1beta1.NotebookTemplateSpec{
				Spec: corev1.PodSpec{
					SchedulerName: constants.RunaiSchedulerName,
					Containers: []corev1.Container{
						{
							Image: testImage,
						},
					},
				},
			},
		},
		Status: kubeflowv1beta1.NotebookStatus{
			Conditions: []kubeflowv1beta1.NotebookCondition{
				{
					Type:    "Running",
					Status:  string(corev1.ConditionTrue),
					Message: "some-messagee6",
				},
			},
		},
	}
	testPytorchJobChild1Pod *corev1.Pod
	TestTFJobExpectedEw     = &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				commonconsts.ControlPlaneWorkloadIdAnnotationForWorkload: "tfjob-workloadid",
			},
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(3)),
			},
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(3*testDeploymentCpuLimit, resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(3*testDeploymentMemoryRequest, resource.DecimalSI),
					constants.GpuResourceName: *resource.NewMilliQuantity(3*1000*testDeploymentGpuLimit, resource.DecimalSI),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(3*testDeploymentCpuRequest, resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(3*testDeploymentMemoryRequest, resource.DecimalSI),
					constants.GpuResourceName: *resource.NewMilliQuantity(3*1000*testDeploymentGpuRequest, resource.DecimalSI),
				},
			},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: string(kubeflowv1.JobRunning),
				Message:   "some-messagee",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(kubeflowv1.JobRunning),
						Status:  corev1.ConditionTrue,
						Message: "some-messagee",
					},
				},
			},
		},
	}
	TestPytorchJobExpectedEw = &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				commonconsts.WorkloadCategoryLabel: "Train",
				constants.ManagedByLabelKey:        "Helm",
			},
			Annotations: map[string]string{
				constants.HelmReleaseAnnotationKey: "rel",
			},
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(5)),
			},
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(7*testDeploymentCpuLimit, resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(7*testDeploymentMemoryRequest, resource.DecimalSI),
					constants.GpuResourceName: *resource.NewMilliQuantity(7*1000*testDeploymentGpuLimit, resource.DecimalSI),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(7*testDeploymentCpuRequest, resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(7*testDeploymentMemoryRequest, resource.DecimalSI),
					constants.GpuResourceName: *resource.NewMilliQuantity(7*1000*testDeploymentGpuRequest, resource.DecimalSI),
				},
			},
			PriorityClassName: "low",
			Preemptibility:    "non-preemptible",
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWCompletedPhase,
				K8sStatus: string(kubeflowv1.JobSucceeded),
				Message:   "some-messagee1",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(kubeflowv1.JobSucceeded),
						Status:  corev1.ConditionTrue,
						Message: "some-messagee1",
					},
					{
						Type:    string(kubeflowv1.JobCreated),
						Status:  corev1.ConditionTrue,
						Message: "some-messagee2",
					},
				},
			},
		},
	}
	TestXGBoostJobExpectedEw = &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{commonconsts.WorkloadCategoryLabel: "Train"},
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWPendingPhase,
				K8sStatus: string(kubeflowv1.JobRestarting),
				Message:   "some-messagee3",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(kubeflowv1.JobRestarting),
						Status:  corev1.ConditionTrue,
						Message: "some-messagee3",
					},
				},
			},
		},
	}
	TestJAXJobExpectedEw = &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{commonconsts.WorkloadCategoryLabel: "Train"},
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: string(kubeflowv1.JobRunning),
				Message:   "some-messagee3",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(kubeflowv1.JobRunning),
						Status:  corev1.ConditionTrue,
						Message: "some-messagee3",
					},
				},
			},
		},
	}
	TestMPIJobV1ExpectedEw = &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{commonconsts.WorkloadCategoryLabel: "Train"},
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWFailedPhase,
				K8sStatus: string(kubeflowv1.JobFailed),
				Message:   "some-messagee4",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(kubeflowv1.JobFailed),
						Status:  corev1.ConditionTrue,
						Message: "some-messagee4",
					},
				},
			},
		},
	}
	TestMPIJobV2beta1ExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWStoppedPhase,
				K8sStatus: string(kubeflowv2beta1.JobSuspended),
				Message:   "some-messagee5",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(kubeflowv2beta1.JobSuspended),
						Status:  corev1.ConditionTrue,
						Message: "some-messagee5",
					},
				},
			},
		},
	}
	TestNotebookExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWRunningPhase,
				K8sStatus: "Running",
				Message:   "some-messagee6",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    "Running",
						Status:  corev1.ConditionTrue,
						Message: "some-messagee6",
					},
				},
			},
		},
	}

	TestKubevirtVMI = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubevirt.io/v1",
			"kind":       "VirtualMachineInstance",
			"metadata": map[string]interface{}{
				"uid":               "kubevirtvmi-UID",
				"name":              "kubevirtvmi-3242344",
				"namespace":         TestNamespace,
				"labels":            map[string]interface{}{},
				"creationTimestamp": metav1.Now().Format(metav1.RFC3339Micro),
			},
			"spec": map[string]interface{}{
				"schedulerName": constants.RunaiSchedulerName,
			},
			"status": map[string]interface{}{
				"phase":  string(kubevirtv1.Scheduled),
				"reason": "some-reasonnn",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   string(kubevirtv1.VirtualMachineInstanceProvisioning),
						"status": string(corev1.ConditionTrue),
						"reason": "some-reasonnn2",
					},
				},
			},
		},
	}
	TestKubevirtVMIExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedNodePools: []string{v1alpha1.DefaultNodePoolName},
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWInitializingPhase,
				K8sStatus: string(kubevirtv1.Scheduled),
				Message:   "some-reasonnn",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:   string(kubevirtv1.VirtualMachineInstanceProvisioning),
						Status: corev1.ConditionTrue,
						Reason: "some-reasonnn2",
					},
				},
			},
		},
	}

	TestDevWorkspace = &devv1alpha2.DevWorkspace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DevWorkspace",
			APIVersion: "workspace.devfile.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "devworkspace1-UID",
			Name:              "devworkspace11-3242344",
			Namespace:         TestNamespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Status: devv1alpha2.DevWorkspaceStatus{
			Phase:   devv1alpha2.DevWorkspaceStatusFailed,
			Message: "some-messageeee8",
			Conditions: []devv1alpha2.DevWorkspaceCondition{
				{
					Type:    devv1alpha2.DevWorkspaceFailedStart,
					Status:  corev1.ConditionTrue,
					Reason:  "some-reasonnn3",
					Message: "some-messagee3",
				},
				{
					Type:    devv1alpha2.DevWorkspaceError,
					Status:  corev1.ConditionTrue,
					Reason:  "some-reasonnn4",
					Message: "some-messagee4",
				},
			},
		},
	}
	TestDevWorkspaceExpectedEw = &v1alpha1.ExternalWorkload{
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWFailedPhase,
				K8sStatus: string(devv1alpha2.DevWorkspaceStatusFailed),
				Message:   "some-messageeee8",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(devv1alpha2.DevWorkspaceFailedStart),
						Status:  corev1.ConditionTrue,
						Reason:  "some-reasonnn3",
						Message: "some-messagee3",
					},
					{
						Type:    string(devv1alpha2.DevWorkspaceError),
						Status:  corev1.ConditionTrue,
						Reason:  "some-reasonnn4",
						Message: "some-messagee4",
					},
				},
			},
		},
	}
	testTestDevWorkspaceChild1Pod *corev1.Pod

	testRunaiWorkload           = &unstructured.Unstructured{}
	testChildPodOfRunaiWorkload *corev1.Pod

	testUnrecognizedOwnerWorkload       = &unstructured.Unstructured{}
	testChildPodWithUnrecognizedOwner   *corev1.Pod
	unrecognizedWorkloadOwnerReferences []metav1.OwnerReference

	TestNodePool = &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			UID:  "npa-uid",
			Name: npAName,
		},
		Spec: v1alpha1.NodePoolSpec{
			LabelKey:   npAKey,
			LabelValue: npAVal,
		},
	}

	// kai.resources twin of TestNodePool. Requested node pools are resolved through
	// node-pool-utils/converter, which reads whichever CRD UseKaiNodePools selects, so both
	// are seeded with identical content.
	TestKaiNodePool = &kaiorg.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			UID:  "npa-uid",
			Name: npAName,
		},
		Spec: kaiorg.NodePoolSpec{
			LabelKey:   npAKey,
			LabelValue: npAVal,
		},
	}

	TestPodGroup = &kaiv2alpha2.PodGroup{
		TypeMeta: metav1.TypeMeta{
			Kind: "PodGroup",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:       testPodGroupId,
			Name:      testPodGroupName,
			Namespace: TestNamespace,
			Labels: map[string]string{
				"some-key1": "some-value1",
				"some-key2": "some-value2",
			},
			Annotations: map[string]string{
				"some-key11": "some-value11",
				"some-key22": "some-value22",
			},
		},
	}

	WorkloadTypeConfigs = []workloadTypeConfig.WorkloadTypeConfig{
		{
			TypeName:       "Deployment",
			Group:          "apps",
			CategoryName:   "Deploy",
			PriorityName:   "very-high",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "Deployment",
			Group:          "some-other-group",
			CategoryName:   "Train",
			PriorityName:   "low",
			Preemptibility: "non-preemptible",
		},
		{
			TypeName:       "TrainingWorkload",
			Group:          "run.ai",
			CategoryName:   "Train",
			PriorityName:   "low",
			Preemptibility: "non-preemptible",
		},
		{
			TypeName:       "InferenceWorkload",
			Group:          "run.ai",
			CategoryName:   "Deploy",
			PriorityName:   "very-high",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "StatefulSet",
			Group:          "apps",
			CategoryName:   "Deploy",
			PriorityName:   "very-high",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "PyTorchJob",
			Group:          "kubeflow.org",
			CategoryName:   "Train",
			PriorityName:   "low",
			Preemptibility: "non-preemptible",
		},
		{
			TypeName:       "RayJob",
			Group:          "ray.io",
			CategoryName:   "Train",
			PriorityName:   "low",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "RayCluster",
			Group:          "ray.io",
			CategoryName:   "Train",
			PriorityName:   "low",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "RayCluster",
			Group:          "some-other-grouppp",
			CategoryName:   "Train222",
			PriorityName:   "low222",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "Pod",
			Group:          "",
			CategoryName:   "Deploy",
			PriorityName:   "very-high",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "DevWorkspace",
			Group:          "workspace.devfile.io",
			CategoryName:   "Build",
			PriorityName:   "high",
			Preemptibility: "non-preemptible",
		},
	}
	WorkloadConfigMap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterconstants.WorkloadsConfigMapName,
			Namespace: clusterconstants.WorkloadsConfigMapNamespace,
		},
		Data: map[string]string{},
	}

	TestKserveInferenceService = &kservev1beta1.InferenceService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "InferenceService",
			APIVersion: "serving.kserve.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "kserve1-UID",
			Name:              "kserve1-3242344",
			Namespace:         TestNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: kservev1beta1.InferenceServiceSpec{
			Predictor: kservev1beta1.PredictorSpec{
				ComponentExtensionSpec: kservev1beta1.ComponentExtensionSpec{
					Annotations: map[string]string{
						GpuFractionAnnotation: testDeploymentFracGpuLimitStr,
					},
				},
				Model: &kservev1beta1.ModelSpec{
					ModelFormat: kservev1beta1.ModelFormat{
						Name: "triton",
					},
					PredictorExtensionSpec: kservev1beta1.PredictorExtensionSpec{
						StorageURI: ptr.To("s3"),
						Container: corev1.Container{
							Image: "registry.example.com/test-image",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    *resource.NewQuantity(2, resource.DecimalSI),
									corev1.ResourceMemory: resource.MustParse("24Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    *resource.NewQuantity(2, resource.DecimalSI),
									corev1.ResourceMemory: resource.MustParse("24Gi"),
								},
							},
						},
					},
				},
			},
		},
		Status: kservev1beta1.InferenceServiceStatus{
			Status: duckv1.Status{
				Conditions: duckv1.Conditions{
					{
						Type:   "LatestDeploymentReady",
						Status: "True",
					},
					{
						Type:   "PredictorConfigurationReady",
						Status: "True",
					},
					{
						Type:   "PredictorReady",
						Status: "True",
					},
					{
						Type:   "RoutesReady",
						Status: "True",
					},
				},
			},
		},
	}
	KartaForDeployment = &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "karta-v1-apps-deployment",
			Labels: v1alpha1.GetKartaLabels("", "apps", "v1", "Deployment", ""),
		},
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "deployment",
					Kind: &kartav1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
				},
				ChildComponents: []kartav1alpha1.ComponentDefinition{
					{
						Name:     "replicaset",
						Kind:     &kartav1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"},
						OwnerRef: ptr.To("deployment"),
					},
				},
			},
		},
		Status: kartav1alpha1.KartaStatus{
			Conditions: []metav1.Condition{
				{Type: string(v1alpha1.RunaiReady), Status: metav1.ConditionTrue},
			},
		},
	}
	KartaForKserveInferenceService = &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "karta-v1beta1-serving-kserve-io-inferenceservice",
			Labels: v1alpha1.GetKartaLabels("", "serving.kserve.io", "v1beta1", "InferenceService", ""),
		},
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "inferenceservice",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "serving.kserve.io",
						Version: "v1beta1",
						Kind:    "InferenceService",
					},
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
							ResourcesPath:         ptr.To(".spec.domain.resources"),
							PriorityClassNamePath: ptr.To(".spec.priorityClassName"),
							NodeAffinityPath:      ptr.To(".spec.affinity.nodeAffinity"),
						},
					},
					StatusDefinition: &kartav1alpha1.StatusDefinition{
						ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
							Path:             ".status.conditions",
							TypeFieldName:    "type",
							StatusFieldName:  "status",
							MessageFieldName: ptr.To("message"),
						},
						StatusMappings: kartav1alpha1.StatusMappings{
							Running: []kartav1alpha1.StatusMatcher{
								{
									ByConditions: []kartav1alpha1.ExpectedCondition{
										{
											Type:   "PredictorReady",
											Status: ptr.To("True"),
										},
										{
											Type:   "RoutesReady",
											Status: ptr.To("True"),
										},
										{
											Type:   "LatestDeploymentReady",
											Status: ptr.To("True"),
										},
									},
								},
							},
							Failed: []kartav1alpha1.StatusMatcher{
								{
									ByConditions: []kartav1alpha1.ExpectedCondition{
										{
											Type:   "PredictorReady",
											Status: ptr.To("False"),
										},
										{
											Type:   "PredictorConfigurationReady",
											Status: ptr.To("False"),
										},
										{
											Type:   "RoutesReady",
											Status: ptr.To("False"),
										},
									},
								},
							},
						},
					},
				},
				ChildComponents: []kartav1alpha1.ComponentDefinition{
					{
						Name:     "predictor",
						OwnerRef: ptr.To("inferenceservice"),
						SpecDefinition: &kartav1alpha1.SpecDefinition{
							PodSpecPath:  ptr.To(".spec.predictor"),
							MetadataPath: ptr.To(".spec.predictor"),
							FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
								ContainerPath: ptr.To(".spec.predictor.model"),
							},
						},
						ScaleDefinition: &kartav1alpha1.ScaleDefinition{
							MinReplicasPath: ptr.To(".spec.predictor.minReplicas"),
							MaxReplicasPath: ptr.To(".spec.predictor.maxReplicas"),
						},
					},
					{
						Name:     "transformer",
						OwnerRef: ptr.To("inferenceservice"),
						SpecDefinition: &kartav1alpha1.SpecDefinition{
							PodSpecPath:  ptr.To(".spec.transformer"),
							MetadataPath: ptr.To(".spec.transformer"),
						},
						ScaleDefinition: &kartav1alpha1.ScaleDefinition{
							MinReplicasPath: ptr.To(".spec.transformer.minReplicas"),
							MaxReplicasPath: ptr.To(".spec.transformer.maxReplicas"),
						},
					},
				},
			},
		},
	}
	TestKserveExpectedEw = &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{
			RequestedPods: v1alpha1.RequestedPods{
				Replicas: ptr.To(int32(1)),
			},
			RequestedResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(2, resource.DecimalSI),
					corev1.ResourceMemory:     resource.MustParse("24Gi"),
					constants.GpuResourceName: *resource.NewMilliQuantity(1000*(testDeploymentFracGpuLimit), resource.DecimalSI),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(2, resource.DecimalSI),
					corev1.ResourceMemory:     resource.MustParse("24Gi"),
					constants.GpuResourceName: *resource.NewMilliQuantity(1000*(testDeploymentFracGpuLimit), resource.DecimalSI),
				},
			},
			ImageName: "registry.example.com/test-image",
		},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
				Phase: v1alpha1.ExWRunningPhase,
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:   "LatestDeploymentReady",
						Status: "True",
					},
					{
						Type:   "PredictorConfigurationReady",
						Status: "True",
					},
					{
						Type:   "PredictorReady",
						Status: "True",
					},
					{
						Type:   "RoutesReady",
						Status: "True",
					},
				},
			},
		},
	}
)

func InitVars() {
	// deployments
	testDeploymentChild1RS = TestReplicaSet.DeepCopy()
	setOwnerReferenceForObject(testDeploymentChild1RS, TestDeployment)
	testDeploymentChild2Pod = TestPod.DeepCopy()
	setOwnerReferenceForObject(testDeploymentChild2Pod, testDeploymentChild1RS)
	testDeploymentChildPodBadSchedulerName = testDeploymentChild2Pod.DeepCopy()
	testDeploymentChildPodBadSchedulerName.Spec.SchedulerName = "default-scheduler"
	testDeploymentChildPodWithPodGroup = testDeploymentChild2Pod.DeepCopy()
	testDeploymentChildPodWithPodGroup.Annotations = map[string]string{
		commonconsts.PodGroupAnnotationForPod: testPodGroupName,
	}
	testDeploymentExpectedEw = TestDeploymentExpectedBasicEw.DeepCopy()
	testDeploymentExpectedEw.Spec.Type = "Deployment"
	fillStaticInfoForExpectedEW(testDeploymentExpectedEw)

	// replicasets
	testReplicaSetChild1Pod = TestPod.DeepCopy()
	setOwnerReferenceForObject(testReplicaSetChild1Pod, TestReplicaSet)
	testReplicaSetExpectedEw = TestReplicaSetExpectedBasicEw.DeepCopy()
	testReplicaSetExpectedEw.Spec.Type = "ReplicaSet"
	fillStaticInfoForExpectedEW(testReplicaSetExpectedEw)

	// statefulset
	testStatefulSetChild1Pod = TestPod.DeepCopy()
	setOwnerReferenceForObject(testStatefulSetChild1Pod, TestStatefulSet)
	testStatefulSetChild1Pod.Annotations = map[string]string{
		commonconsts.PodGroupAnnotationForPod: testPodGroupName,
	}
	testStatefulSetExpectedEw = TestStatefulSetExpectedBasicEw.DeepCopy()
	testStatefulSetExpectedEw.Spec.Type = "StatefulSet"
	fillStaticInfoForExpectedEW(testStatefulSetExpectedEw)

	// Nim StatefulSet
	testNimStatefulSetChild1Pod = TestPod.DeepCopy()
	setOwnerReferenceForObject(testNimStatefulSetChild1Pod, testNimStatefulSet)
	testNimStatefulSetChild1Pod.Annotations = map[string]string{
		commonconsts.PodGroupAnnotationForPod: testPodGroupName,
	}
	testNimStatefulSetExpectedEw = TestNimStatefulSetExpectedBasicEw.DeepCopy()
	testNimStatefulSetExpectedEw.Spec.Type = "NVIDIA-NIM"
	fillStaticInfoForExpectedEW(testNimStatefulSetExpectedEw)

	// NVCF Function
	TestNvcfFunctionChildPod = TestPod.DeepCopy()
	setOwnerReferenceForObject(TestNvcfFunctionChildPod, TestNvcfFunction)
	TestNvcfFunctionChildPod.Annotations = map[string]string{
		commonconsts.PodGroupAnnotationForPod: testPodGroupName,
	}
	TestNvcfFunctionChildPod.Labels = map[string]string{
		"function-version-id": "some-function-version-id",
	}
	TestNvcfFunctionChildPod.Spec.InitContainers = []corev1.Container{
		{
			Image: "don't-use-this-image",
			Env: []corev1.EnvVar{
				{
					Name:  "INFERENCE_CONTAINER",
					Value: "some-image",
				},
				{
					Name:  "FUNCTION_ID",
					Value: "some-function-id",
				},
				{
					Name:  "FUNCTION_VERSION_ID",
					Value: "some-function-version-id",
				},
				{
					Name:  "FUNCTION_NAME",
					Value: "some-function-name",
				},
				{
					Name:  "INFERENCE_PROTOCOL",
					Value: "REST",
				},
				{
					Name:  "NVCF_FQDN",
					Value: "https://my.domain.com",
				},
			},
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest, resource.DecimalSI),
				},
			},
		},
	}
	TestNvcfFunctionChildPod.Status.Phase = corev1.PodPending
	testNvcfFunctionExpectedEw = TestNvcfFunctionExpectedBasicEw.DeepCopy()
	fillStaticInfoForExpectedEW(testNvcfFunctionExpectedEw)

	// batch job
	testBatchJobChild1Pod = TestPod.DeepCopy()
	testBatchJobChild2Pod = testPod2.DeepCopy()
	setOwnerReferenceForObject(testBatchJobChild1Pod, TestBatchJob)
	setOwnerReferenceForObject(testBatchJobChild2Pod, TestBatchJob)
	testBatchJobExpectedEw = TestBatchJobExpectedBasicEw.DeepCopy()
	testBatchJobExpectedEw.Spec.Type = "Job"
	fillStaticInfoForExpectedEW(testBatchJobExpectedEw)

	// cronjob
	testCronJobChild1Job = TestBatchJob.DeepCopy()
	testCronJobChild2Pod = TestPod.DeepCopy()
	setOwnerReferenceForObject(testCronJobChild1Job, TestCronJob)
	setOwnerReferenceForObject(testCronJobChild2Pod, TestCronJob)
	testCronJobExpectedEw = TestCronJobExpectedBasicEw.DeepCopy()
	testCronJobExpectedEw.Spec.Type = "CronJob"
	fillStaticInfoForExpectedEW(testCronJobExpectedEw)

	// rays
	testRayJobChildRayCluster = TestRayCluster.DeepCopy()
	testRayJobChild1Pod = TestPod.DeepCopy()
	testRayJobChild2Pod = testPod2.DeepCopy()
	setOwnerReferenceForObject(testRayJobChildRayCluster, TestRayJob)
	setOwnerReferenceForObject(testRayJobChild1Pod, TestRayJob)
	setOwnerReferenceForObject(testRayJobChild2Pod, TestRayJob)
	testRayClusterChildPod = TestPod.DeepCopy()
	setOwnerReferenceForObject(testRayClusterChildPod, TestRayCluster)
	testRayClusterChildPod.Annotations = map[string]string{
		commonconsts.PodGroupAnnotationForPod: testPodGroupName,
	}

	testRunaiWorkload.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "run.ai",
		Version: "v2alpha1",
		Kind:    "TrainingWorkload",
	})
	testRunaiWorkload.SetNamespace(TestNamespace)
	testRunaiWorkload.SetName("training-workload-1")
	testRunaiWorkload.SetCreationTimestamp(metav1.Now())
	testRunaiWorkload.SetUID("some-uidddd-0")
	testChildPodOfRunaiWorkload = TestPod.DeepCopy()
	testChildPodOfRunaiWorkload.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         testRunaiWorkload.GetAPIVersion(),
			BlockOwnerDeletion: ptr.To(true),
			Controller:         ptr.To(true),
			Kind:               testRunaiWorkload.GetKind(),
			Name:               testRunaiWorkload.GetName(),
			UID:                testRunaiWorkload.GetUID(),
		},
	}

	// knative
	setOwnerReferenceForObject(testKnativeChildConfiguration, TestKnativeService)
	setOwnerReferenceForObject(testKnativeChildRevision, testKnativeChildConfiguration)
	TestKnativeServiceInference.Labels[commonconsts.ControlPlaneWorkloadKindLabelForWorkload] = "InferenceWorkload"
	TestKnativeServiceInferenceExpectedEw.Spec.Type = "Inference"

	// kubeflow
	testPytorchJobChild1Pod = testPod2.DeepCopy()
	setOwnerReferenceForObject(testPytorchJobChild1Pod, TestPytorchJob1)

	// devworkspace
	testTestDevWorkspaceChild1Pod = TestPod.DeepCopy()
	setOwnerReferenceForObject(testTestDevWorkspaceChild1Pod, TestDevWorkspace)

	testUnrecognizedOwnerWorkload.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "made.up.group",
		Version: "v1",
		Kind:    "MadeUpKind",
	})
	testUnrecognizedOwnerWorkload.SetNamespace(TestNamespace)
	testUnrecognizedOwnerWorkload.SetName("unrecognized-object-name")
	testUnrecognizedOwnerWorkload.SetCreationTimestamp(metav1.Now())
	testUnrecognizedOwnerWorkload.SetUID("some-uidddd")
	unrecognizedWorkloadOwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         testUnrecognizedOwnerWorkload.GetAPIVersion(),
			BlockOwnerDeletion: ptr.To(true),
			Controller:         ptr.To(true),
			Kind:               testUnrecognizedOwnerWorkload.GetKind(),
			Name:               testUnrecognizedOwnerWorkload.GetName(),
			UID:                testUnrecognizedOwnerWorkload.GetUID(),
		},
	}
	testChildPodWithUnrecognizedOwner = TestPod.DeepCopy()
	testChildPodWithUnrecognizedOwner.OwnerReferences = unrecognizedWorkloadOwnerReferences

	testPodWithPodGroupAnnotation = testPod2.DeepCopy()
	testPodWithPodGroupAnnotation.Annotations = map[string]string{
		commonconsts.PodGroupAnnotationForPod: testPodGroupName,
	}

	workloadTypeConfigsJson, _ := json.Marshal(WorkloadTypeConfigs)
	WorkloadConfigMap.Data[clusterconstants.WorkloadsConfigMapTypesKey] = string(workloadTypeConfigsJson)
}

func setOwnerReferenceForObject(object, owner client.Object) {
	object.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         owner.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			BlockOwnerDeletion: ptr.To(true),
			Controller:         ptr.To(true),
			Kind:               owner.GetObjectKind().GroupVersionKind().Kind,
			Name:               owner.GetName(),
			UID:                owner.GetUID(),
		},
	})
}

func fillStaticInfoForExpectedEW(ew *v1alpha1.ExternalWorkload) {
	ew.Spec.ProjectName = testProjectName
	ew.Spec.ProjectId = testProjectId
}
