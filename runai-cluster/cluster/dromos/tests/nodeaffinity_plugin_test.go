/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"path/filepath"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/plugins/mnnvl"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/plugins/nodeaffinity"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

const hostnameLabel = "kubernetes.io/hostname"

var _ = Describe("NodeAffinity Plugin Integration Tests", Serial, Ordered, func() {
	var (
		k8sManager     ctrl.Manager
		k8sClient      client.Client
		managerCtx     context.Context
		managerCancel  context.CancelFunc
		namespace      = "runai-nodeaffinitytest"
		controllerName = "nodeaffinity-plugin-test-controller"
	)

	BeforeAll(func() {
		managerCtx, managerCancel = context.WithCancel(ctx)

		absCertDir, err := filepath.Abs(webhookCertDir)
		Expect(err).NotTo(HaveOccurred())

		k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme:  scheme.Scheme,
			Metrics: metricsserver.Options{BindAddress: "0"},
			WebhookServer: webhook.NewServer(webhook.Options{
				CertDir: absCertDir,
				Port:    webhookPort,
			}),
		})
		Expect(err).NotTo(HaveOccurred())

		k8sClient = k8sManager.GetClient()

		// Both node-affinity writers run together: this is the combination the
		// framework's mutation loop has to converge on.
		registry := framework.NewRegistry(
			nodeaffinity.NewPlugin(k8sClient),
			mnnvl.NewPlugin(false),
		)
		fw := framework.NewFramework(registry, nil)
		nodeAffinityController := controller.
			NewWorkloadRunnerController(k8sClient, fw, k8sManager.GetEventRecorderFor(controllerName)).
			WithName(controllerName)
		Expect(nodeAffinityController.SetupWithManager(managerCtx, k8sManager)).To(Succeed())

		setupWebhooks(fw, k8sClient, managerCtx, k8sManager)

		startManager(k8sManager, managerCtx, managerCancel)

		err = k8sClient.Create(managerCtx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		})
		if !apierrors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		waitForSetupToComplete(k8sClient, managerCtx, webhookPort)
	})

	AfterAll(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		deleteAllForType(managerCtx, k8sClient, &nvapi.ComputeDomainList{})
		managerCancel()
		waitForWebhookToBeClosed(webhookUrlValidate)
	})

	AfterEach(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		deleteAllForType(managerCtx, k8sClient, &nvapi.ComputeDomainList{})
	})

	It("rejects a WorkloadRunner whose node-affinity instruction is invalid", func() {
		jobset := test.NewJobSetBuilder("test-workload-na-reject", namespace).WithDefault().Build()
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{Name: "test-runner-na-reject", Namespace: namespace},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Template: marshalRawExtension(jobset),
				Instructions: newNodeAffinityInstructions(corev1.NodeSelectorRequirement{
					Key:      hostnameLabel,
					Operator: corev1.NodeSelectorOpNotIn,
				}),
			},
		}

		err := k8sClient.Create(managerCtx, runner)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("values"))
	})

	It("rejects a runner whose karta cannot express node affinity", func() {
		karta := test.NewValidJobSetKartaWithFragmentedPodSpec()
		karta.Spec.StructureDefinition.ChildComponents[0].SpecDefinition.FragmentedPodSpecDefinition.NodeAffinityPath = nil
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(karta))

		jobset := test.NewJobSetBuilder("test-workload-na-nopath", namespace).WithDefault().Build()
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{Name: "test-runner-na-nopath", Namespace: namespace},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Template:     marshalRawExtension(jobset),
				Instructions: newNodeAffinityInstructions(hostnameExclusion("node-a")),
			},
		}

		err := k8sClient.Create(managerCtx, runner)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("cannot express node affinity"))
	})

	It("admits a runner with an instruction when no karta exists yet", func() {
		jobset := test.NewJobSetBuilder("test-workload-na-nokarta", namespace).WithDefault().Build()
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{Name: "test-runner-na-nokarta", Namespace: namespace},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Template:     marshalRawExtension(jobset),
				Instructions: newNodeAffinityInstructions(hostnameExclusion("node-a")),
			},
		}

		Expect(k8sClient.Create(managerCtx, runner)).To(Succeed())
	})

	It("lands node exclusions on the child's pod affinity", func() {
		workloadName := "test-workload-na-exclude"
		jobset := test.NewJobSetBuilder(workloadName, namespace).WithDefault().Build()
		instructions := newNodeAffinityInstructions(hostnameExclusion("node-a", "node-b"))

		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
		runner := createWorkloadRunner(managerCtx, k8sClient, "test-runner-na-exclude", namespace, jobset, instructions)

		createdJobSet := waitForJobSet(managerCtx, k8sClient, workloadName, namespace)
		expectRequirementOnEveryReplicatedJob(createdJobSet, hostnameExclusion("node-a", "node-b"))

		Eventually(func(g Gomega) {
			updated := &dromosv1alpha1.WorkloadRunner{}
			g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), updated)).To(Succeed())
			g.Expect(updated.Status.Plugins).To(HaveKey(nodeaffinity.PluginName))

			pluginStatus := updated.Status.Plugins[nodeaffinity.PluginName]
			g.Expect(pluginStatus.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
			g.Expect(pluginStatus.Issues).To(BeEmpty())
			g.Expect(pluginStatus.AdditionalInfo).NotTo(BeNil())
			g.Expect(string(pluginStatus.AdditionalInfo.Raw)).To(ContainSubstring(`["node-a","node-b"]`))
		}, timeout, interval).Should(Succeed())
	})

	It("converges with mnnvl enabled on the same workload", func() {
		workloadName := "test-workload-na-mnnvl"
		jobset := test.NewJobSetBuilder(workloadName, namespace).WithDefault().WithGPU(1).Build()

		// Two terms on purpose: an OR is the shape a cartesian merge re-multiplies,
		// and mnnvl merges into the same selector - the worst case for convergence.
		instructions := newNodeAffinityInstructionsFrom(
			[]corev1.NodeSelectorTerm{
				{MatchExpressions: []corev1.NodeSelectorRequirement{hostnameExclusion("node-a")}},
				{MatchExpressions: []corev1.NodeSelectorRequirement{{
					Key: "zone", Operator: corev1.NodeSelectorOpIn, Values: []string{"us-east-1"},
				}}},
			})
		instructions.GpuOptimizations = &dromosv1alpha1.GpuOptimizations{
			Mnnvl: &dromosv1alpha1.MnnvlConfig{
				Mode:                            dromosv1alpha1.MnnvlModeRequired,
				GpuNetworkAccelerationLabelKeys: []string{gpuNetworkAccelerationLabel},
			},
		}

		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
		runner := createWorkloadRunner(managerCtx, k8sClient, "test-runner-na-mnnvl", namespace, jobset, instructions)

		// Only applied if the mutation converged; otherwise no JobSet ever appears.
		createdJobSet := waitForJobSet(managerCtx, k8sClient, workloadName, namespace)

		// Both OR branches survive, and mnnvl's requirement is ANDed into each.
		forEachReplicatedJobNodeSelector(createdJobSet, func(terms []corev1.NodeSelectorTerm) {
			Expect(terms).To(HaveLen(2))
			for _, term := range terms {
				Expect(term.MatchExpressions).To(ContainElement(corev1.NodeSelectorRequirement{
					Key:      gpuNetworkAccelerationLabel,
					Operator: corev1.NodeSelectorOpExists,
				}))
			}
			Expect(terms[0].MatchExpressions).To(ContainElement(hostnameExclusion("node-a")))
		})

		Eventually(func(g Gomega) {
			updated := &dromosv1alpha1.WorkloadRunner{}
			g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), updated)).To(Succeed())
			g.Expect(updated.Status.Issues).To(BeEmpty())
			g.Expect(updated.Status.Plugins[nodeaffinity.PluginName].Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
			g.Expect(updated.Status.Plugins[mnnvl.PluginName].Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
		}, timeout, interval).Should(Succeed())

		// The controller keeps re-reconciling; the child must not drift.
		Consistently(func(g Gomega) {
			current := &jobsetv1alpha2.JobSet{}
			g.Expect(k8sClient.Get(managerCtx, client.ObjectKeyFromObject(createdJobSet), current)).To(Succeed())
			g.Expect(current.Spec.ReplicatedJobs).To(Equal(createdJobSet.Spec.ReplicatedJobs))
		}, "2s", interval).Should(Succeed())
	})
})

func newNodeAffinityInstructions(requirements ...corev1.NodeSelectorRequirement) *dromosv1alpha1.Instructions {
	return newNodeAffinityInstructionsFrom([]corev1.NodeSelectorTerm{{MatchExpressions: requirements}})
}

func newNodeAffinityInstructionsFrom(terms []corev1.NodeSelectorTerm) *dromosv1alpha1.Instructions {
	return &dromosv1alpha1.Instructions{
		NodeAffinity: &dromosv1alpha1.NodeAffinityConfig{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: terms,
				},
			},
		},
	}
}

func hostnameExclusion(hostnames ...string) corev1.NodeSelectorRequirement {
	return corev1.NodeSelectorRequirement{
		Key:      hostnameLabel,
		Operator: corev1.NodeSelectorOpNotIn,
		Values:   hostnames,
	}
}

func waitForJobSet(ctx context.Context, k8sClient client.Client, name, namespace string) *jobsetv1alpha2.JobSet {
	jobSet := &jobsetv1alpha2.JobSet{}
	Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, jobSet)
	}, timeout, interval).Should(Succeed())
	return jobSet
}

func expectRequirementOnEveryReplicatedJob(jobSet *jobsetv1alpha2.JobSet, expected corev1.NodeSelectorRequirement) {
	forEachReplicatedJobNodeSelector(jobSet, func(terms []corev1.NodeSelectorTerm) {
		Expect(terms).To(HaveLen(1))
		Expect(terms[0].MatchExpressions).To(ContainElement(expected))
	})
}

func forEachReplicatedJobNodeSelector(jobSet *jobsetv1alpha2.JobSet, assert func([]corev1.NodeSelectorTerm)) {
	Expect(jobSet.Spec.ReplicatedJobs).NotTo(BeEmpty())
	for _, replicatedJob := range jobSet.Spec.ReplicatedJobs {
		affinity := replicatedJob.Template.Spec.Template.Spec.Affinity
		Expect(affinity).NotTo(BeNil())
		Expect(affinity.NodeAffinity).NotTo(BeNil())

		required := affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		Expect(required).NotTo(BeNil())
		assert(required.NodeSelectorTerms)
	}
}
