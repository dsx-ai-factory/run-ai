/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"path/filepath"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/plugins/mnnvl"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

const gpuNetworkAccelerationLabel = "nvidia.com/gpu.clique"

var _ = Describe("MNNVL Plugin Integration Tests", Serial, Ordered, func() {
	var (
		k8sManager          ctrl.Manager
		k8sClient           client.Client
		managerCtx          context.Context
		managerCancel       context.CancelFunc
		mnnvlNamespace      = "runai-mnnvltest"
		mnnvlControllerName = "mnnvl-plugin-test-controller"
	)

	BeforeAll(func() {
		managerCtx, managerCancel = context.WithCancel(ctx)

		// Use static certs from testdata
		absCertDir, err := filepath.Abs(webhookCertDir)
		Expect(err).NotTo(HaveOccurred())

		k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
			Metrics: metricsserver.Options{
				BindAddress: "0",
			},
			WebhookServer: webhook.NewServer(webhook.Options{
				CertDir: absCertDir,
				Port:    webhookPort,
			}),
		})
		Expect(err).NotTo(HaveOccurred())

		k8sClient = k8sManager.GetClient()

		// Setup controller with the dromos default OSS wiring: the mnnvl plugin with no
		// options, so it stamps dromos.run.ai/mnnvl-workload-name (the default) and its
		// default pod affinity groups peers by that same label.
		mnnvlPlugin := mnnvl.NewPlugin(false)
		Expect(mnnvlPlugin).NotTo(BeNil())

		registry := framework.NewRegistry(mnnvlPlugin)
		fw := framework.NewFramework(registry, nil)
		mnnvlController := controller.NewWorkloadRunnerController(k8sClient, fw, k8sManager.GetEventRecorderFor(mnnvlControllerName)).WithName(mnnvlControllerName)
		err = mnnvlController.SetupWithManager(managerCtx, k8sManager)
		Expect(err).NotTo(HaveOccurred())

		setupWebhooks(fw, k8sClient, managerCtx, k8sManager)

		startManager(k8sManager, managerCtx, managerCancel)
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: mnnvlNamespace,
			},
		}
		_ = k8sClient.Create(managerCtx, ns)

		waitForSetupToComplete(k8sClient, managerCtx, webhookPort)
	})

	AfterAll(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		deleteAllForType(managerCtx, k8sClient, &nvapi.ComputeDomainList{})
		managerCancel()
		waitForWebhookToBeClosed(webhookUrlValidate)
	})

	Context("WorkloadRunner with MNNVL", func() {
		var (
			runnerName string
		)
		BeforeEach(func() {
			runnerName = "test-runner-mnnvl"
		})

		AfterEach(func() {
			deleteAllTestResources(managerCtx, k8sClient)
			deleteAllForType(managerCtx, k8sClient, &nvapi.ComputeDomainList{})
		})

		It("should validate WorkloadRunner with Enabled mode", func() {
			workloadName := "test-workload-mnnvl-validate"
			jobset := test.NewJobSetBuilder(workloadName, mnnvlNamespace).WithDefault().WithGPU(1).Build()
			instructions := &dromosv1alpha1.Instructions{
				GpuOptimizations: &dromosv1alpha1.GpuOptimizations{
					Mnnvl: &dromosv1alpha1.MnnvlConfig{
						Mode:                            dromosv1alpha1.MnnvlModeRequired,
						GpuNetworkAccelerationLabelKeys: []string{gpuNetworkAccelerationLabel},
					},
				},
			}
			_ = createWorkloadRunner(managerCtx, k8sClient, "test-runner-validate", mnnvlNamespace, jobset, instructions)
		})

		It("should reject WorkloadRunner with invalid mode", func() {
			workloadName := "test-workload-mnnvl-reject"
			jobset := test.NewJobSetBuilder(workloadName, mnnvlNamespace).WithDefault().WithGPU(1).Build()
			runnerToReject := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runner-reject",
					Namespace: mnnvlNamespace,
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Instructions: &dromosv1alpha1.Instructions{
						GpuOptimizations: &dromosv1alpha1.GpuOptimizations{
							Mnnvl: &dromosv1alpha1.MnnvlConfig{
								Mode:                            "Preferred",
								GpuNetworkAccelerationLabelKeys: []string{gpuNetworkAccelerationLabel},
							},
						},
					},
					Template: marshalRawExtension(jobset),
				},
			}

			err := k8sClient.Create(managerCtx, runnerToReject)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unsupported value"))
		})

		It("should add resource claims, pod affinity, node affinity, and create ComputeDomain", func() {
			workloadName := "test-workload-mnnvl-add-all"
			jobset := test.NewJobSetBuilder(workloadName, mnnvlNamespace).WithDefault().WithGPU(1).Build()
			instructions := &dromosv1alpha1.Instructions{
				GpuOptimizations: &dromosv1alpha1.GpuOptimizations{
					Mnnvl: &dromosv1alpha1.MnnvlConfig{
						Mode:                            dromosv1alpha1.MnnvlModeRequired,
						GpuNetworkAccelerationLabelKeys: []string{gpuNetworkAccelerationLabel},
					},
				},
			}
			runner, createdJobSet := createWorkloadRunnerWithMNNVL(managerCtx, k8sClient, jobset, runnerName, mnnvlNamespace, workloadName, instructions)

			expectedTemplateClaimName := "cd-wlr-" + runnerName
			expectedClaimName := "compute-domain"
			verifyResourceClaims(createdJobSet, expectedClaimName, expectedTemplateClaimName)

			verifyPodAffinity(createdJobSet, runner, gpuNetworkAccelerationLabel)
			verifyNodeAffinity(createdJobSet, gpuNetworkAccelerationLabel)

			expectedCDName := "wlr-" + runnerName
			_ = verifyComputeDomain(managerCtx, k8sClient, createdJobSet, runner, expectedCDName, expectedTemplateClaimName, 2)

			updated := &dromosv1alpha1.WorkloadRunner{}
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{
					Name:      runner.Name,
					Namespace: runner.Namespace,
				}, updated)
			}, timeout, interval).Should(Succeed())

			Expect(updated).NotTo(BeNil())
			Expect(updated.Status.Plugins).To(HaveKey("mnnvl"))
			Expect(updated.Status.Plugins["mnnvl"].Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
			Expect(updated.Status.Plugins["mnnvl"].Conditions).To(HaveLen(1))
			Expect(updated.Status.Plugins["mnnvl"].Conditions[0].Type).To(Equal(mnnvl.ComputeDomainReadyConditionName))
			Expect(updated.Status.Plugins["mnnvl"].Conditions[0].Status).To(Equal(metav1.ConditionTrue))
			Expect(updated.Status.Plugins["mnnvl"].Issues).To(HaveLen(0))
			Expect(updated.Status.Phase).To(Equal(updated.Status.ResourcePhase))
		})

		It("should watch ComputeDomain force creation", func() {
			workloadName := "test-workload-mnnvl-watch"
			jobset := test.NewJobSetBuilder(workloadName, mnnvlNamespace).WithDefault().WithGPU(1).Build()
			instructions := &dromosv1alpha1.Instructions{
				GpuOptimizations: &dromosv1alpha1.GpuOptimizations{
					Mnnvl: &dromosv1alpha1.MnnvlConfig{
						Mode:                            dromosv1alpha1.MnnvlModeRequired,
						GpuNetworkAccelerationLabelKeys: []string{gpuNetworkAccelerationLabel},
					},
				},
			}
			runner, createdJobSet := createWorkloadRunnerWithMNNVL(managerCtx, k8sClient, jobset, runnerName, mnnvlNamespace, workloadName, instructions)

			expectedCDName := "wlr-" + runnerName
			expectedTemplateClaimName := "cd-wlr-" + runnerName
			computeDomain := verifyComputeDomain(managerCtx, k8sClient, createdJobSet, runner, expectedCDName, expectedTemplateClaimName, 2)
			computeDomain.Spec.NumNodes = 500
			err := k8sClient.Delete(managerCtx, computeDomain)
			Expect(err).NotTo(HaveOccurred())

			verifyComputeDomain(managerCtx, k8sClient, createdJobSet, runner, expectedCDName, expectedTemplateClaimName, 2)

		})

		It("should set plugin phase to Failed when Karta does not support MNNVL", func() {
			workloadName := "test-workload-mnnvl-bad-karta"
			jobset := test.NewJobSetBuilder(workloadName, mnnvlNamespace).WithDefault().WithGPU(1).Build()
			instructions := &dromosv1alpha1.Instructions{
				GpuOptimizations: &dromosv1alpha1.GpuOptimizations{
					Mnnvl: &dromosv1alpha1.MnnvlConfig{
						Mode:                            dromosv1alpha1.MnnvlModeRequired,
						GpuNetworkAccelerationLabelKeys: []string{gpuNetworkAccelerationLabel},
					},
				},
			}

			// Create Karta without SpecDefinition — MNNVL validation will fail during MutateResource
			karta := test.NewValidJobSetKarta()
			for i := range karta.Spec.StructureDefinition.ChildComponents {
				karta.Spec.StructureDefinition.ChildComponents[i].SpecDefinition = nil
			}
			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(karta))

			runner := createWorkloadRunner(managerCtx, k8sClient, "test-runner-bad-karta", mnnvlNamespace, jobset, instructions)

			updated := &dromosv1alpha1.WorkloadRunner{}
			Eventually(func(g Gomega) {
				err := k8sClient.Get(managerCtx, types.NamespacedName{
					Name:      runner.Name,
					Namespace: runner.Namespace,
				}, updated)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(updated.Status.Plugins).To(HaveKey("mnnvl"))
				g.Expect(updated.Status.Plugins["mnnvl"].Phase).To(Equal(dromosv1alpha1.PluginPhaseFailed))
			}, timeout, interval).Should(Succeed())
		})
	})
})

func createWorkloadRunnerWithMNNVL(managerCtx context.Context, k8sClient client.Client, jobset *jobsetv1alpha2.JobSet, runnerName, namespace string, workloadName string, instructions *dromosv1alpha1.Instructions) (*dromosv1alpha1.WorkloadRunner, *jobsetv1alpha2.JobSet) {
	_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))

	runner := createWorkloadRunner(managerCtx, k8sClient, runnerName, namespace, jobset, instructions)

	createdJobSet := &jobsetv1alpha2.JobSet{}
	Eventually(func() error {
		return k8sClient.Get(managerCtx, types.NamespacedName{
			Name:      workloadName,
			Namespace: namespace,
		}, createdJobSet)
	}, timeout, interval).Should(Succeed())
	return runner, createdJobSet
}

func verifyResourceClaims(jobSet *jobsetv1alpha2.JobSet, expectedClaimName string, expectedTemplateClaimName string) {
	Expect(jobSet.Spec.ReplicatedJobs).NotTo(BeEmpty())
	for _, replicatedJob := range jobSet.Spec.ReplicatedJobs {
		podSpec := replicatedJob.Template.Spec.Template.Spec

		// Verify container has resource claim
		Expect(podSpec.Containers[0].Resources.Claims).To(ContainElement(
			corev1.ResourceClaim{Name: expectedClaimName},
		))

		// Verify pod has resource claims with template
		Expect(podSpec.ResourceClaims).NotTo(BeEmpty())
		foundClaim := false
		for _, resourceClaim := range podSpec.ResourceClaims {
			if resourceClaim.Name == expectedClaimName {
				foundClaim = true
				Expect(resourceClaim.ResourceClaimTemplateName).NotTo(BeNil())
				Expect(*resourceClaim.ResourceClaimTemplateName).To(Equal(expectedTemplateClaimName))
				break
			}
		}
		Expect(foundClaim).To(BeTrue(), "Expected to find resource claim with name %s", expectedClaimName)
	}
}

func verifyPodAffinity(jobSet *jobsetv1alpha2.JobSet, runner *dromosv1alpha1.WorkloadRunner, expectedTopologyKey string) {
	for _, replicatedJob := range jobSet.Spec.ReplicatedJobs {
		podTemplate := replicatedJob.Template.Spec.Template

		// dromos stamps its own workload-name label on every GPU pod...
		Expect(podTemplate.ObjectMeta.Labels).To(HaveKeyWithValue(
			mnnvl.WorkloadNameLabel,
			runner.Name,
		))

		// ...and the default pod affinity groups peers by exactly that label.
		affinity := podTemplate.Spec.Affinity
		Expect(affinity).NotTo(BeNil())
		Expect(affinity.PodAffinity).NotTo(BeNil())
		Expect(affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution).NotTo(BeEmpty())

		term := affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0]
		Expect(term.PodAffinityTerm.TopologyKey).To(Equal(expectedTopologyKey))
		Expect(term.PodAffinityTerm.LabelSelector).NotTo(BeNil())
		Expect(term.PodAffinityTerm.LabelSelector.MatchLabels).To(HaveKeyWithValue(
			mnnvl.WorkloadNameLabel,
			runner.Name,
		))
	}
}

func verifyNodeAffinity(jobSet *jobsetv1alpha2.JobSet, expectedKey string) {
	for _, replicatedJob := range jobSet.Spec.ReplicatedJobs {
		affinity := replicatedJob.Template.Spec.Template.Spec.Affinity
		Expect(affinity).NotTo(BeNil())
		Expect(affinity.NodeAffinity).NotTo(BeNil())
		Expect(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).NotTo(BeNil())
		Expect(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(HaveLen(1))
		Expect(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions).To(ContainElement(corev1.NodeSelectorRequirement{
			Key:      expectedKey,
			Operator: corev1.NodeSelectorOpExists,
		}))
	}
}
func verifyComputeDomain(managerCtx context.Context, k8sClient client.Client, jobSet *jobsetv1alpha2.JobSet, runner *dromosv1alpha1.WorkloadRunner, expectedCDName string, expectedTemplateClaimName string, expectedNumNodes int) *nvapi.ComputeDomain {
	computeDomain := &nvapi.ComputeDomain{}
	Eventually(func(g Gomega) {
		err := k8sClient.Get(managerCtx, types.NamespacedName{
			Name:      expectedCDName,
			Namespace: jobSet.Namespace,
		}, computeDomain)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(computeDomain).NotTo(BeNil())
		g.Expect(computeDomain.DeletionTimestamp).To(BeNil())
		g.Expect(computeDomain.Name).To(Equal(expectedCDName))
		g.Expect(computeDomain.Namespace).To(Equal(jobSet.Namespace))
		g.Expect(computeDomain.Spec.NumNodes).To(Equal(expectedNumNodes))
		g.Expect(computeDomain.Spec.Channel).NotTo(BeNil())
		g.Expect(computeDomain.Spec.Channel.ResourceClaimTemplate).NotTo(BeNil())
		g.Expect(computeDomain.Spec.Channel.ResourceClaimTemplate.Name).To(Equal(expectedTemplateClaimName))
		g.Expect(computeDomain.OwnerReferences).To(HaveLen(1))
		g.Expect(computeDomain.OwnerReferences[0].UID).To(Equal(runner.UID))
	}, timeout, interval).Should(Succeed())

	return computeDomain
}
