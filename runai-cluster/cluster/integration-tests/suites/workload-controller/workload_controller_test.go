/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_controller

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/integration-tests/common"
	"github.com/run-ai/runai/runai-cluster/cluster/integration-tests/config"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Workload Controller Test", Ordered, func() {
	var (
		testConfig         *config.TestConfig
		kubeClient         client.Client
		ctx                context.Context
		testCreatedObjects = map[string]client.Object{}
	)

	BeforeAll(func() {
		testConfig = config.GetTestConfig(Default)
		kubeClient = testConfig.KubeClient
		ctx = testConfig.Ctx
	})

	AfterAll(func() {
		for _, obj := range testCreatedObjects {
			if err := kubeClient.Delete(ctx, obj); err != nil {
				fmt.Printf("Failed to delete resource %v: %v\n", obj, err)
			}
		}
		testConfig.Cleanup()
	})
	It("Submit interactive workload and check image field", func() {
		Skip("Check why the pod phase is 'pending', the test works only on regular envs")
		podNamePrefix := common.GetRandomName("iw")
		interactiveWl, err := common.LoadWorkload("../../resources/workspace.yaml", common.GetRandomName("interactive-wl"), testConfig.TestNamespace.Name, podNamePrefix, &iwv2.InteractiveWorkload{}, nil)
		Expect(err).ToNot(HaveOccurred(), "Failed to load InteractiveWorkload")

		testCreatedObjects[podNamePrefix] = interactiveWl
		err = kubeClient.Create(ctx, interactiveWl)
		Expect(err).ToNot(HaveOccurred(), "Failed to create interactive workload")

		var createdPod *corev1.Pod
		err = common.WaitForCondition(ctx, 1*time.Second, 10*time.Second, func() bool {
			createdPod, _ = common.GetPodByPrefix(ctx, kubeClient, podNamePrefix, testConfig.TestNamespace.Name)
			return createdPod != nil && createdPod.Status.Phase == corev1.PodRunning
		})
		Expect(err).ToNot(HaveOccurred(),
			fmt.Sprintf("Expected one pod with a name starting with '%s' in namespace '%s'.",
				podNamePrefix, testConfig.TestNamespace.Name))
		Expect(createdPod.Spec.Containers[0].Image).To(Equal("ubuntu"), "Expected the pod's image to be 'ubuntu'.")
	})

	It("Submit training with new pvc", func() {
		Skip("Check why the pod phase is 'pending', the test works only on regular envs")
		podNamePrefix := common.GetRandomName("tw")
		pvcName := common.GetRandomName("pvc")
		optionalParams := map[string]string{"{{PVC_NAME}}": pvcName}
		trainingWl, err := common.LoadWorkload("../../resources/training-with-pvc.yaml", "tw-for-integration-test", testConfig.TestNamespace.Name, podNamePrefix, &twv2.TrainingWorkload{}, optionalParams)
		testCreatedObjects[podNamePrefix] = trainingWl

		err = kubeClient.Create(ctx, trainingWl)
		Expect(err).ToNot(HaveOccurred(), "Failed to create training with new pvc")

		common.ValidatePodStatus(ctx, kubeClient, podNamePrefix, testConfig.TestNamespace.Name, corev1.PodSucceeded)

		pvc, err := common.GetPVC(ctx, kubeClient, testConfig.TestNamespace.Name, pvcName)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to get PVC  '%s' in namespace '%s': %v", pvcName, testConfig.TestNamespace.Name, err))
		Expect(pvc.Status.Phase).To(Equal(corev1.ClaimBound), fmt.Sprintf("Expected PVC '%s' to be in 'Bound' phase, but it is in '%s' phase", pvcName, pvc.Status.Phase))
	})

	It("Submit training with autoDeletionTimeAfterCompletion", func() {
		Skip("Check why the pod phase is 'pending', the test works only on regular envs")
		podNamePrefix := common.GetRandomName("training")
		trainingWl, err := common.LoadWorkload("../../resources/training-with-auto-completion-after-deletion.yaml", common.GetRandomName("training-wl"), testConfig.TestNamespace.Name, podNamePrefix, &twv2.TrainingWorkload{}, nil)
		testCreatedObjects[podNamePrefix] = trainingWl

		err = kubeClient.Create(ctx, trainingWl)
		Expect(err).ToNot(HaveOccurred(), "Failed to create training")

		// Validate that pod status is completed
		common.ValidatePodStatus(ctx, kubeClient, podNamePrefix, testConfig.TestNamespace.Name, corev1.PodSucceeded)

		// Validate that pod was deleted
		common.ValidatePodNotFound(ctx, kubeClient, podNamePrefix, testConfig.TestNamespace.Name)
	})

	It("Submit distributed training with 2 workers", func() {
		Skip("Check why the pod phase is 'pending', the test works only on regular envs")
		podNamePrefix := common.GetRandomName("distributed")
		distributedWl, err := common.LoadWorkload("../../resources/distributed.yaml", common.GetRandomName("distributed-wl"), testConfig.TestNamespace.Name, podNamePrefix, &dwv2.DistributedWorkload{}, nil)
		testCreatedObjects[podNamePrefix] = distributedWl

		err = kubeClient.Create(ctx, distributedWl)
		Expect(err).ToNot(HaveOccurred(), "Failed to create distributed training")

		common.ValidatePodStatus(ctx, kubeClient, fmt.Sprintf("%s-chief-0", podNamePrefix), testConfig.TestNamespace.Name, corev1.PodRunning)
		common.ValidatePodStatus(ctx, kubeClient, fmt.Sprintf("%s-worker-0", podNamePrefix), testConfig.TestNamespace.Name, corev1.PodRunning)
		common.ValidatePodStatus(ctx, kubeClient, fmt.Sprintf("%s-worker-1", podNamePrefix), testConfig.TestNamespace.Name, corev1.PodRunning)
	})
})
