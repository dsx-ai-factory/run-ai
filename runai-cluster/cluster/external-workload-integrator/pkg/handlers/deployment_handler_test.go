/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"strconv"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clustercommonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/test"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	cpuLimit   int64 = 800
	cpuRequest int64 = 400
	memLimit   int64 = 8000
	memRequest int64 = 4000
	gpuLimit         = 5.5

	gpuFractionAnnotationValue    = 0.4
	gpuFractionAnnotationValueStr = "0.4"
	gpuMemoryAnnotationValue      = 5000
)

func TestHandlers(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Info().Msgf("Started handlers tests")
	RegisterFailHandler(Fail)
	test.MockEnv(t)
	RunSpecs(t, "handlers Test Suite")
}

var _ = Describe("TestDeploymentHandler", Ordered, func() {
	var (
		externalWorkload          *v1alpha1.ExternalWorkload
		expectedExternalWorkload  *v1alpha1.ExternalWorkload
		cachedClient              client.WithWatch
		deployment                *appsv1.Deployment
		deploymentTopOwnerHandler *DeploymentTopOwnerHandler

		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeAll(func() {
		scheme := scheme_init.Scheme()
		ns := test.BuildNamespace(test.TestNamespace, map[string]string{})
		cachedClient = fake.NewClientBuilder().WithStatusSubresource(&corev1.Pod{}).WithScheme(scheme).WithObjects(ns, test.TestNodePool, test.TestKaiNodePool).Build()

		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(func() {
			cancel()
		})

		deploymentTopOwnerHandler = NewDeploymentTopOwnerHandler(cachedClient, nil)
	})

	BeforeEach(func() {
		externalWorkload = &v1alpha1.ExternalWorkload{}
		deployment = test.TestDeployment.DeepCopy()
		expectedExternalWorkload = test.TestDeploymentExpectedBasicEw.DeepCopy()
		expectedExternalWorkload.Spec.Preemptibility = ""
	})

	It("Sanity", func() {
		err := deploymentTopOwnerHandler.Parse(ctx, deployment, externalWorkload)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)
	})
	It("Validate expected spec with some different values", func() {
		deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms =
			test.TestDeployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[:1]
		expectedExternalWorkload.Spec.RequestedNodePools = []string{test.TestNodePool.Name}

		deployment.Spec.Template.Spec.PriorityClassName = "some-priority-class"
		expectedExternalWorkload.Spec.PriorityClassName = "some-priority-class"

		var numOfReplicas int64 = 2
		deployment.Spec.Replicas = ptr.To(int32(numOfReplicas))
		expectedExternalWorkload.Spec.RequestedPods.Replicas = ptr.To(int32(numOfReplicas))

		deployment.Spec.Template.Spec.Containers[0].Image = "some-imagee"
		deployment.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:        *resource.NewQuantity(cpuLimit, resource.DecimalSI),
				corev1.ResourceMemory:     *resource.NewQuantity(memLimit, resource.DecimalSI),
				constants.GpuResourceName: *resource.NewMilliQuantity(int64(1000*gpuLimit), resource.DecimalSI),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:        *resource.NewQuantity(cpuRequest, resource.DecimalSI),
				corev1.ResourceMemory:     *resource.NewQuantity(memRequest, resource.DecimalSI),
				constants.GpuResourceName: *resource.NewQuantity(0, resource.DecimalSI),
			},
		}
		var numberOfContainers int64 = 2
		deployment.Spec.Template.Spec.Containers = []corev1.Container{deployment.Spec.Template.Spec.Containers[0],
			deployment.Spec.Template.Spec.Containers[0]}
		deployment.Spec.Template.Annotations[test.GpuFractionAnnotation] = gpuFractionAnnotationValueStr
		deployment.Spec.Template.Annotations[test.GpuMemoryAnnotation] = strconv.Itoa(gpuMemoryAnnotationValue)

		expectedExternalWorkload.Spec.ImageName = "some-imagee"
		expectedExternalWorkload.Spec.RequestedResources = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:              *resource.NewQuantity(numOfReplicas*numberOfContainers*cpuLimit, resource.DecimalSI),
				corev1.ResourceMemory:           *resource.NewQuantity(numOfReplicas*numberOfContainers*memLimit, resource.DecimalSI),
				constants.GpuResourceName:       *resource.NewMilliQuantity(numOfReplicas*int64(1000*(float64(numberOfContainers)*gpuLimit+gpuFractionAnnotationValue)), resource.DecimalSI),
				constants.GpuMemoryResourceName: *resource.NewQuantity(numOfReplicas*int64(gpuMemoryAnnotationValue)*test.GpuMemoryLimitConversion, resource.DecimalSI),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:              *resource.NewQuantity(numOfReplicas*numberOfContainers*cpuRequest, resource.DecimalSI),
				corev1.ResourceMemory:           *resource.NewQuantity(numOfReplicas*numberOfContainers*memRequest, resource.DecimalSI),
				constants.GpuResourceName:       *resource.NewMilliQuantity(numOfReplicas*int64(1000*gpuFractionAnnotationValue), resource.DecimalSI),
				constants.GpuMemoryResourceName: *resource.NewQuantity(numOfReplicas*int64(gpuMemoryAnnotationValue)*test.GpuMemoryLimitConversion, resource.DecimalSI),
			},
		}

		err := deploymentTopOwnerHandler.Parse(ctx, deployment, externalWorkload)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)
	})
	It("Test observed status", func() {
		// no conditions - unknown status
		deployment.Status.Conditions = nil
		expectedExternalWorkload.Status.ObservedStatus.K8sStatus = ""
		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWUnknownPhase
		expectedExternalWorkload.Status.ObservedStatus.Message = "Progressing condition type not found"
		expectedExternalWorkload.Status.ObservedStatus.Conditions = []v1alpha1.ObservedCondition{}
		err := deploymentTopOwnerHandler.Parse(ctx, deployment, externalWorkload)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)

		// irrelevant condition - unknown status
		deployment.Status.Conditions = []appsv1.DeploymentCondition{{Type: appsv1.DeploymentAvailable}}
		expectedExternalWorkload.Status.ObservedStatus.Conditions = []v1alpha1.ObservedCondition{
			{Type: string(appsv1.DeploymentAvailable)}}
		err = deploymentTopOwnerHandler.Parse(ctx, deployment, externalWorkload)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)

		// unknown reason
		deployment.Status.Conditions = []appsv1.DeploymentCondition{{Type: appsv1.DeploymentProgressing,
			Reason: "blabla", Message: "some-message"}}
		expectedExternalWorkload.Status.ObservedStatus.K8sStatus = ""
		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWUnknownPhase
		expectedExternalWorkload.Status.ObservedStatus.Message = "some-message"
		expectedExternalWorkload.Status.ObservedStatus.Conditions = []v1alpha1.ObservedCondition{
			{Type: string(appsv1.DeploymentProgressing), Reason: "blabla", Message: "some-message"}}
		err = deploymentTopOwnerHandler.Parse(ctx, deployment, externalWorkload)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)

		// known reason but wrong conditionStatus (false)
		deployment.Status.Conditions = []appsv1.DeploymentCondition{{Type: appsv1.DeploymentProgressing,
			Reason: "NewReplicaSetCreated", Message: "some-message", Status: corev1.ConditionFalse}}
		expectedExternalWorkload.Status.ObservedStatus.Conditions = []v1alpha1.ObservedCondition{
			{Type: string(appsv1.DeploymentProgressing), Reason: "NewReplicaSetCreated",
				Message: "some-message", Status: corev1.ConditionFalse}}
		err = deploymentTopOwnerHandler.Parse(ctx, deployment, externalWorkload)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)

		// pending - progressing
		deployment.Status.Conditions = []appsv1.DeploymentCondition{{Type: appsv1.DeploymentProgressing,
			Reason: "FoundNewReplicaSet", Message: "some-message-2", Status: corev1.ConditionTrue}}
		expectedExternalWorkload.Status.ObservedStatus.K8sStatus = ProgressingDeploymentStatus
		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWPendingPhase
		expectedExternalWorkload.Status.ObservedStatus.Message = "some-message-2"
		expectedExternalWorkload.Status.ObservedStatus.Conditions = []v1alpha1.ObservedCondition{
			{Type: string(appsv1.DeploymentProgressing), Reason: "FoundNewReplicaSet",
				Message: "some-message-2", Status: corev1.ConditionTrue}}
		err = deploymentTopOwnerHandler.Parse(ctx, deployment, externalWorkload)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)

		// running - complete
		deployment.Status.Conditions = []appsv1.DeploymentCondition{{Type: appsv1.DeploymentProgressing,
			Reason: "NewReplicaSetAvailable", Message: "some-message-3", Status: corev1.ConditionTrue}}
		expectedExternalWorkload.Status.ObservedStatus.K8sStatus = CompleteDeploymentStatus
		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWRunningPhase
		expectedExternalWorkload.Status.ObservedStatus.Message = "some-message-3"
		expectedExternalWorkload.Status.ObservedStatus.Conditions = []v1alpha1.ObservedCondition{
			{Type: string(appsv1.DeploymentProgressing), Reason: "NewReplicaSetAvailable",
				Message: "some-message-3", Status: corev1.ConditionTrue}}
		err = deploymentTopOwnerHandler.Parse(ctx, deployment, externalWorkload)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)

		// degraded
		deployment.Status.Conditions = []appsv1.DeploymentCondition{{Type: appsv1.DeploymentProgressing,
			Reason: "ProgressDeadlineExceeded", Message: "some-message-4", Status: corev1.ConditionFalse}}
		expectedExternalWorkload.Status.ObservedStatus.K8sStatus = DegradedDeploymentStatus
		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWDegradedPhase
		expectedExternalWorkload.Status.ObservedStatus.Message = "some-message-4"
		expectedExternalWorkload.Status.ObservedStatus.Conditions = []v1alpha1.ObservedCondition{
			{Type: string(appsv1.DeploymentProgressing), Reason: "ProgressDeadlineExceeded",
				Message: "some-message-4", Status: corev1.ConditionFalse}}
		err = deploymentTopOwnerHandler.Parse(ctx, deployment, externalWorkload)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)
	})
	It("Test observed status - initializing by pods", func() {
		// first - create the running pods
		deploymentPod1 := test.TestPod.DeepCopy()
		deploymentPod1.Status.Phase = corev1.PodRunning
		deploymentPod1.Labels = map[string]string{
			commonconstants.TopOwnerUidLabel:        string(deployment.UID),
			commonconstants.TopOwnerWorkloadIdLabel: string(deployment.UID),
		}
		setOwnerReferenceForObject(deploymentPod1, deployment)
		deploymentPod2 := deploymentPod1.DeepCopy()
		deploymentPod2.Name = deploymentPod2.Name + "2"
		err := cachedClient.Create(ctx, deploymentPod1)
		Expect(err).To(BeNil())
		err = cachedClient.Create(ctx, deploymentPod2)
		Expect(err).To(BeNil())

		// initializing - by pods + progressing
		deployment.Status.Conditions = []appsv1.DeploymentCondition{{Type: appsv1.DeploymentProgressing,
			Reason: "FoundNewReplicaSet", Message: "some-message-3", Status: corev1.ConditionTrue}}
		expectedExternalWorkload.Status.ObservedStatus.K8sStatus = ProgressingDeploymentStatus
		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWInitializingPhase
		expectedExternalWorkload.Status.ObservedStatus.Message = "some-message-3"
		expectedExternalWorkload.Status.ObservedStatus.Conditions = []v1alpha1.ObservedCondition{
			{Type: string(appsv1.DeploymentProgressing), Reason: "FoundNewReplicaSet",
				Message: "some-message-3", Status: corev1.ConditionTrue}}
		err = deploymentTopOwnerHandler.Parse(ctx, deployment, externalWorkload)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)

		// now if one of the pods is pending - the status should be pending
		deploymentPod2.Status.Phase = corev1.PodPending
		err = cachedClient.Status().Update(ctx, deploymentPod2)
		Expect(err).To(BeNil())
		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWPendingPhase
		err = deploymentTopOwnerHandler.Parse(ctx, deployment, externalWorkload)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)

		// now one of the pods is Pending+ContainerCreating -
		// so deployment should be Initializing
		deploymentPod2.Annotations = map[string]string{clustercommonconstants.PodStatusAnnotation: "ContainerCreating"}
		err = cachedClient.Update(ctx, deploymentPod2)
		Expect(err).To(BeNil())
		expectedExternalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWInitializingPhase
		err = deploymentTopOwnerHandler.Parse(ctx, deployment, externalWorkload)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, externalWorkload, expectedExternalWorkload)
	})
	It("DRA GPU resources from pod ResourceClaims", func() {
		draFakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(test.TestDRAGpuClaim.DeepCopy(), test.TestDRAGpuClaimTemplate.DeepCopy()).Build()
		fakeKubeClient := kubefake.NewClientset(test.TestDRAGpuClaim.DeepCopy(), test.TestDRAGpuClaimTemplate.DeepCopy())
		draParser := dra.NewDRAParser(draFakeClient, fakeKubeClient)
		draHandler := NewDeploymentTopOwnerHandler(cachedClient, draParser)

		draDeployment := test.TestDeploymentWithDRA.DeepCopy()
		draExpected := test.TestDeploymentWithDRAExpectedEw.DeepCopy()
		draExpected.Spec.Preemptibility = ""

		draEw := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: test.TestNamespace,
			},
		}
		err := draHandler.Parse(ctx, draDeployment, draEw)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, draEw, draExpected)
	})
})

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
