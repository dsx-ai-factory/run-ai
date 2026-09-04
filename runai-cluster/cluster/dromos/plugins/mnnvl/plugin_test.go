/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/testutil"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/test"
	"github.com/google/uuid"
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

const (
	testNamespace    = "runai-test"
	testWorkloadName = "test-workload"
	testGPULabel     = "nvidia.com/gpu.clique"

	testPeerGroupingLabel = "test.dromos/peer-group-id"
)

// testOptions runs the plugin in its DEFAULT mode: stamp the workload-name label
// and group GPU pods by it. Provider-specific tests build their own.
func testOptions() []Option {
	return nil
}

// testProvidePodAffinity is a stand-in consumer provider that supplies pod
// affinity grouping by a neutral label instead of the default.
func testProvidePodAffinity(runner *dromosv1alpha1.WorkloadRunner, _ *unstructured.Unstructured) *corev1.PodAffinity {
	cfg := runner.Spec.Instructions.GpuOptimizations.Mnnvl
	terms := make([]corev1.WeightedPodAffinityTerm, 0, len(cfg.GpuNetworkAccelerationLabelKeys))
	for _, label := range cfg.GpuNetworkAccelerationLabelKeys {
		terms = append(terms, corev1.WeightedPodAffinityTerm{
			Weight: 10,
			PodAffinityTerm: corev1.PodAffinityTerm{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{testPeerGroupingLabel: string(runner.UID)},
				},
				TopologyKey: label,
			},
		})
	}
	if len(terms) == 0 {
		return nil
	}
	return &corev1.PodAffinity{PreferredDuringSchedulingIgnoredDuringExecution: terms}
}

func init() {
	_ = jobsetv1alpha2.AddToScheme(scheme.Scheme)
	_ = nvapi.AddToScheme(scheme.Scheme)
	_ = dromosv1alpha1.AddToScheme(scheme.Scheme)
}

var _ = Describe("MNNVL Plugin", func() {
	var (
		ctx   context.Context
		state *reconcile.State
	)

	BeforeEach(func() {
		ctx = context.Background()
		state = reconcile.New()
	})

	Describe("WebhookValidate", func() {
		It("passes validation with None mode", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeNone, testGPULabel))

			warnings, errs := plugin.WebhookValidate(ctx, nil, runner)

			Expect(warnings).To(BeEmpty())
			Expect(errs).To(BeEmpty())
		})

		It("passes validation with Required mode and valid config", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))

			warnings, errs := plugin.WebhookValidate(ctx, nil, runner)

			Expect(warnings).To(BeEmpty())
			Expect(errs).To(BeEmpty())
		})

		It("passes validation with Required mode but missing label keys", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired))

			warnings, errs := plugin.WebhookValidate(ctx, nil, runner)

			Expect(warnings).To(BeEmpty())
			Expect(errs).To(BeEmpty())
		})

		It("fails validation with unsupported mnnvl mode", func() {
			plugin := newTestPlugin(false)
			config := newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel)
			config.Mode = "Preferred"
			runner := newTestRunner(testWorkloadName, testNamespace, config)

			warnings, errs := plugin.WebhookValidate(ctx, nil, runner)

			Expect(warnings).To(BeEmpty())
			Expect(errs).To(HaveLen(1))
			Expect(errs[0].Type).To(Equal(field.ErrorTypeNotSupported))
		})
	})

	DescribeTable("MutateResource mutations for different types of pod definitions Kartas",
		func(testFunc func(*kartav1alpha1.Karta)) {
			By("using pod template spec Karta")
			testFunc(test.NewValidJobSetKartaWithPodTemplateSpec())

			By("using pod spec Karta")
			testFunc(test.NewValidJobSetKartaWithPodSpec())

			By("using fragmented pod spec Karta")
			testFunc(test.NewValidJobSetKartaWithFragmentedPodSpec())

		},
		Entry("node affinity",
			func(karta *kartav1alpha1.Karta) {
				plugin := newTestPlugin(false)
				runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
				jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()
				for i := range jobset.Spec.ReplicatedJobs {
					jobset.Spec.ReplicatedJobs[i].Template.Spec.Template.Spec.Affinity = &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "test-key",
												Operator: corev1.NodeSelectorOpExists,
											},
										},
									},
								},
							},
						},
					}
				}

				unstructuredJobset := convertToUnstructured(jobset)
				err := plugin.MutateResource(ctx, runner, unstructuredJobset, karta, state)
				Expect(err).NotTo(HaveOccurred())
				Expect(state.Issues()).To(HaveLen(0))

				convertFromUnstructured(unstructuredJobset, jobset)
				for _, replicatedJob := range jobset.Spec.ReplicatedJobs {
					affinity := replicatedJob.Template.Spec.Template.Spec.Affinity
					Expect(affinity).NotTo(BeNil())
					Expect(affinity.NodeAffinity).NotTo(BeNil())
					Expect(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).NotTo(BeNil())
					Expect(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(HaveLen(1))
					Expect(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions).To(ContainElement(corev1.NodeSelectorRequirement{
						Key:      testGPULabel,
						Operator: corev1.NodeSelectorOpExists,
					}))
				}
			},
		),
		Entry("resource claims",
			func(karta *kartav1alpha1.Karta) {
				plugin := newTestPlugin(false)
				runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
				jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

				unstructuredJobset := convertToUnstructured(jobset)
				err := plugin.MutateResource(ctx, runner, unstructuredJobset, karta, state)
				Expect(err).NotTo(HaveOccurred())
				Expect(state.Issues()).To(HaveLen(0))

				convertFromUnstructured(unstructuredJobset, jobset)

				for _, replicatedJob := range jobset.Spec.ReplicatedJobs {
					podSpec := replicatedJob.Template.Spec.Template.Spec
					claimName := getResourceClaimName()
					Expect(podSpec.Containers[0].Resources.Claims).To(ContainElement(
						corev1.ResourceClaim{Name: claimName}))
					Expect(podSpec.ResourceClaims).NotTo(BeEmpty())
					Expect(podSpec.ResourceClaims[0].Name).To(Equal(claimName))
					Expect(podSpec.ResourceClaims[0].ResourceClaimTemplateName).NotTo(BeNil())
				}

			},
		),
	)

	Describe("MutateResource pod affinity", func() {
		It("default: groups GPU pods by the workload-name label", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodTemplateSpec(), state)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.Issues()).To(HaveLen(0))

			convertFromUnstructured(unstructuredJobset, jobset)
			for _, replicatedJob := range jobset.Spec.ReplicatedJobs {
				template := replicatedJob.Template.Spec.Template
				pa := template.Spec.Affinity.PodAffinity
				Expect(pa).NotTo(BeNil())
				Expect(pa.PreferredDuringSchedulingIgnoredDuringExecution).NotTo(BeEmpty())
				term := pa.PreferredDuringSchedulingIgnoredDuringExecution[0]
				Expect(term.PodAffinityTerm.TopologyKey).To(Equal(testGPULabel))
				Expect(term.PodAffinityTerm.LabelSelector.MatchLabels).To(HaveKeyWithValue(WorkloadNameLabel, runner.Name))
				// the default path also stamps the label the affinity selects
				Expect(template.Labels).To(HaveKeyWithValue(WorkloadNameLabel, runner.Name))
			}
		})

		It("default with FragmentedPodSpec: groups by the workload-name label", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, test.NewValidJobSetKartaWithFragmentedPodSpec(), state)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.Issues()).To(HaveLen(0))

			convertFromUnstructured(unstructuredJobset, jobset)
			for _, replicatedJob := range jobset.Spec.ReplicatedJobs {
				template := replicatedJob.Template.Spec.Template
				pa := template.Spec.Affinity.PodAffinity
				Expect(pa).NotTo(BeNil())
				Expect(pa.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels).To(HaveKeyWithValue(WorkloadNameLabel, runner.Name))
				Expect(template.Labels).To(HaveKeyWithValue(WorkloadNameLabel, runner.Name))
			}
		})

		It("provider: supplies the consumer's selector instead of the default", func() {
			plugin := NewPlugin(false, WithPodAffinityProvider(testProvidePodAffinity), WithWorkloadLabelStampDisabled())
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodTemplateSpec(), state)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.Issues()).To(HaveLen(0))

			convertFromUnstructured(unstructuredJobset, jobset)
			for _, replicatedJob := range jobset.Spec.ReplicatedJobs {
				template := replicatedJob.Template.Spec.Template
				// the affinity uses the provider's selector, not the default label
				term := template.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0]
				Expect(term.PodAffinityTerm.LabelSelector.MatchLabels).To(HaveKeyWithValue(testPeerGroupingLabel, string(runner.UID)))
				Expect(term.PodAffinityTerm.LabelSelector.MatchLabels).NotTo(HaveKey(WorkloadNameLabel))
				// the provider path does not stamp the workload-name label
				Expect(template.Labels).NotTo(HaveKey(WorkloadNameLabel))
			}
		})

		It("provider returns nil: no pod affinity", func() {
			// A provider returning nil is how a consumer says "an external scheduler
			// owns placement" - dromos adds no pod affinity.
			nilProvider := func(*dromosv1alpha1.WorkloadRunner, *unstructured.Unstructured) *corev1.PodAffinity {
				return nil
			}
			plugin := NewPlugin(false, WithPodAffinityProvider(nilProvider), WithWorkloadLabelStampDisabled())
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodTemplateSpec(), state)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.Issues()).To(HaveLen(0))

			convertFromUnstructured(unstructuredJobset, jobset)
			for _, replicatedJob := range jobset.Spec.ReplicatedJobs {
				template := replicatedJob.Template.Spec.Template
				podSpec := template.Spec
				verifyEmptyPodAffinity(&podSpec)
			}
		})

		It("bare PodSpec: gets the default affinity but is not stamped (no metadata to carry the label)", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodSpec(), state)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.Issues()).To(HaveLen(0))

			convertFromUnstructured(unstructuredJobset, jobset)
			for _, replicatedJob := range jobset.Spec.ReplicatedJobs {
				template := replicatedJob.Template.Spec.Template
				pa := template.Spec.Affinity.PodAffinity
				Expect(pa).NotTo(BeNil())
				Expect(pa.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels).To(HaveKeyWithValue(WorkloadNameLabel, runner.Name))
				// a bare PodSpec has no metadata, so the label itself isn't stamped
				Expect(template.Labels).NotTo(HaveKey(WorkloadNameLabel))
			}
		})

		It("default: re-running MutateResource is idempotent (no duplicate terms)", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			Expect(plugin.MutateResource(ctx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodTemplateSpec(), state)).To(Succeed())
			Expect(plugin.MutateResource(ctx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodTemplateSpec(), state)).To(Succeed())
			Expect(state.Issues()).To(HaveLen(0))

			convertFromUnstructured(unstructuredJobset, jobset)
			for _, replicatedJob := range jobset.Spec.ReplicatedJobs {
				pa := replicatedJob.Template.Spec.Template.Spec.Affinity.PodAffinity
				Expect(pa.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))
			}
		})
	})

	Describe("mergePodAffinity", func() {
		preferred := func(topologyKey, selKey, selVal string) corev1.WeightedPodAffinityTerm {
			return corev1.WeightedPodAffinityTerm{
				Weight: 10,
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{selKey: selVal}},
					TopologyKey:   topologyKey,
				},
			}
		}

		It("preserves user terms and skips a src term whose topology key already exists", func() {
			dst := &corev1.PodAffinity{PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				preferred("user-topology", "app", "mine"),
				preferred("topo-a", "user", "x"),
			}}
			src := &corev1.PodAffinity{PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				preferred("topo-a", "workload-name", "w"),
				preferred("topo-b", "workload-name", "w"),
			}}

			mergePodAffinity(dst, src)

			Expect(topologyKeys(dst.PreferredDuringSchedulingIgnoredDuringExecution)).To(ConsistOf("user-topology", "topo-a", "topo-b"))
			// the user's topo-a term is kept, not replaced by src's
			Expect(dst.PreferredDuringSchedulingIgnoredDuringExecution[1].PodAffinityTerm.LabelSelector.MatchLabels).To(HaveKeyWithValue("user", "x"))
		})

		It("is idempotent - merging the same src twice adds nothing the second time", func() {
			dst := &corev1.PodAffinity{}
			src := &corev1.PodAffinity{PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{preferred("topo-a", "workload-name", "w")}}
			mergePodAffinity(dst, src)
			mergePodAffinity(dst, src)
			Expect(dst.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))
		})

		It("dedups required terms by topology key", func() {
			dst := &corev1.PodAffinity{RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{TopologyKey: "topo-a", LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"user": "x"}}},
			}}
			src := &corev1.PodAffinity{RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{TopologyKey: "topo-a", LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"workload-name": "w"}}},
				{TopologyKey: "topo-b", LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"workload-name": "w"}}},
			}}
			mergePodAffinity(dst, src)
			Expect(dst.RequiredDuringSchedulingIgnoredDuringExecution).To(HaveLen(2))
			Expect(dst.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels).To(HaveKeyWithValue("user", "x"))
		})

		It("no-ops on nil src", func() {
			dst := &corev1.PodAffinity{}
			mergePodAffinity(dst, nil)
			Expect(dst.PreferredDuringSchedulingIgnoredDuringExecution).To(BeEmpty())
		})
	})

	Describe("MutateResource edge cases", func() {
		It("fails validation when stamping the workload-name label and a FragmentedPodSpec lacks labelsPath", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			kartaMissingLabelsPath := test.NewValidJobSetKartaWithFragmentedPodSpec()
			for i := range kartaMissingLabelsPath.Spec.StructureDefinition.ChildComponents {
				kartaMissingLabelsPath.Spec.StructureDefinition.ChildComponents[i].SpecDefinition = &kartav1alpha1.SpecDefinition{
					FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
						ContainersPath:     ptr.To(".spec.containers"),
						ResourceClaimsPath: ptr.To(".spec.resourceClaims"),
						NodeAffinityPath:   ptr.To(".spec.affinity.nodeAffinity"),
						PodAffinityPath:    ptr.To(".spec.affinity.podAffinity"),
					},
				}
			}

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, kartaMissingLabelsPath, state)

			Expect(err).To(MatchError(ErrMissingLabelPath))
			Expect(state.Issues()).To(HaveLen(1))
		})

		It("skips mutation with None mode", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeNone, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			originalSpec := unstructuredJobset.Object["spec"]
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodTemplateSpec(), state)

			Expect(err).NotTo(HaveOccurred())
			Expect(state.Issues()).To(HaveLen(0))
			Expect(unstructuredJobset.Object["spec"]).To(Equal(originalSpec))
		})

		It("skips mutation with no MNNVL config", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			originalSpec := unstructuredJobset.Object["spec"]
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodTemplateSpec(), state)

			Expect(err).NotTo(HaveOccurred())
			Expect(state.Issues()).To(HaveLen(0))
			Expect(unstructuredJobset.Object["spec"]).To(Equal(originalSpec))
		})

		It("fails validation for Karta with FragmentedPodSpec missing resource/container paths", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			riMissingResourcePath := test.NewValidJobSetKartaWithFragmentedPodSpec()
			riMissingResourcePath.Spec.StructureDefinition.RootComponent.SpecDefinition = &kartav1alpha1.SpecDefinition{
				FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
					PodAffinityPath:  ptr.To(".spec.affinity.podAffinity"),
					NodeAffinityPath: ptr.To(".spec.affinity.nodeAffinity"),
				},
			}

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, riMissingResourcePath, state)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("requires containersPath or containerPath or resourceClaimsPath"))
			Expect(state.Issues()).To(HaveLen(1))

			issueItem := state.Issues()[0]
			Expect(issueItem.Error()).To(HavePrefix("[JobSet:test-workload] given karta valid-jobset-karta-fragmented does not support MNNVL:"))
			Expect(issueItem.Group).To(Equal(issue.Network))
			Expect(issueItem.GVK.Kind).To(Equal("JobSet"))
			Expect(issueItem.Name).To(Equal(testWorkloadName))
		})

		It("fails validation for Karta with FragmentedPodSpec missing pod affinity paths", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			riMissingAffinityPaths := test.NewValidJobSetKartaWithFragmentedPodSpec()
			for i := range riMissingAffinityPaths.Spec.StructureDefinition.ChildComponents {
				riMissingAffinityPaths.Spec.StructureDefinition.ChildComponents[i].SpecDefinition = &kartav1alpha1.SpecDefinition{
					FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
						ContainersPath:     ptr.To(".spec.containers"),
						ResourceClaimsPath: ptr.To(".spec.resourceClaims"),
						NodeAffinityPath:   ptr.To(".spec.affinity.nodeAffinity"),
					},
				}
			}

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, riMissingAffinityPaths, state)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("requires podAffinityPath"))
			Expect(state.Issues()).To(HaveLen(1))

			issueItem := state.Issues()[0]
			Expect(issueItem.Error()).To(HavePrefix("[JobSet:test-workload] given karta valid-jobset-karta-fragmented does not support MNNVL:"))
			Expect(issueItem.Group).To(Equal(issue.Network))
			Expect(issueItem.GVK.Kind).To(Equal("JobSet"))
			Expect(issueItem.Name).To(Equal(testWorkloadName))
		})

		It("fails validation for Karta with no Spec definition", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()
			riNoSpecDefinition := test.NewValidJobSetKartaWithPodTemplateSpec()
			riNoSpecDefinition.Spec.StructureDefinition.RootComponent.SpecDefinition = nil
			for i := range riNoSpecDefinition.Spec.StructureDefinition.ChildComponents {
				riNoSpecDefinition.Spec.StructureDefinition.ChildComponents[i].SpecDefinition = &kartav1alpha1.SpecDefinition{}
			}

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, riNoSpecDefinition, state)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("any component with a Spec definition"))
			Expect(state.Issues()).To(HaveLen(1))

			issueItem := state.Issues()[0]
			Expect(issueItem.Error()).To(Equal("[JobSet:test-workload] given karta valid-jobset-karta-podtemplatespec does not support MNNVL: karta does not have any component with a Spec definition"))
			Expect(issueItem.Group).To(Equal(issue.Network))
			Expect(issueItem.GVK.Kind).To(Equal("JobSet"))
			Expect(issueItem.Name).To(Equal(testWorkloadName))
		})

		It("succeeds validation if gpu network acceleration label keys are not set", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodTemplateSpec(), state)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("GPU network acceleration label keys are not set"))
			Expect(state.Issues()).To(HaveLen(1))

			issueItem := state.Issues()[0]
			Expect(issueItem.Error()).To(Equal("[JobSet:test-workload] GPU network acceleration label keys are not set"))
			Expect(issueItem.Group).To(Equal(issue.Compute))
			Expect(issueItem.GVK.Kind).To(Equal("JobSet"))
			Expect(issueItem.Name).To(Equal(testWorkloadName))
		})

		It("fails validation for Karta with FragmentedPodSpec missing node affinity paths", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().WithGPU(1).Build()

			riMissingAffinityPaths := test.NewValidJobSetKartaWithFragmentedPodSpec()
			for i := range riMissingAffinityPaths.Spec.StructureDefinition.ChildComponents {
				riMissingAffinityPaths.Spec.StructureDefinition.ChildComponents[i].SpecDefinition = &kartav1alpha1.SpecDefinition{
					FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
						ContainersPath:     ptr.To(".spec.containers"),
						ResourceClaimsPath: ptr.To(".spec.resourceClaims"),
						PodAffinityPath:    ptr.To(".spec.affinity.podAffinity"),
					},
				}
			}

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.MutateResource(ctx, runner, unstructuredJobset, riMissingAffinityPaths, state)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("requires nodeAffinityPath"))
			Expect(state.Issues()).To(HaveLen(1))

			issueItem := state.Issues()[0]
			Expect(issueItem.Error()).To(HavePrefix("[JobSet:test-workload] given karta valid-jobset-karta-fragmented does not support MNNVL:"))
			Expect(issueItem.Group).To(Equal(issue.Network))
			Expect(issueItem.GVK.Kind).To(Equal("JobSet"))
			Expect(issueItem.Name).To(Equal(testWorkloadName))
		})

	})

	Describe("HandleAdditionalResources", func() {
		var k8sClient client.Client
		var testCtx context.Context

		BeforeEach(func() {
			k8sClient = testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(scheme.Scheme).Build())
			testCtx = context.TODO()
		})

		It("creates ComputeDomain with correct name, node count, and owner reference", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithReplicas(3).WithGPU(2).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.HandleAdditionalResources(testCtx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodTemplateSpec(), k8sClient, state)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.Issues()).To(HaveLen(0))

			cdList := &nvapi.ComputeDomainList{}
			err = k8sClient.List(testCtx, cdList, client.InNamespace(testNamespace))
			Expect(err).NotTo(HaveOccurred())
			Expect(cdList.Items).To(HaveLen(1))

			cd := cdList.Items[0]
			Expect(cd.Name).To(ContainSubstring(testWorkloadName))
			Expect(cd.Spec.NumNodes).To(Equal(3))
			Expect(cd.OwnerReferences).To(HaveLen(1))
			Expect(cd.OwnerReferences[0].Name).To(Equal(testWorkloadName))
			Expect(cd.OwnerReferences[0].Kind).To(Equal("WorkloadRunner"))
			Expect(*cd.OwnerReferences[0].Controller).To(BeTrue())
		})

		It("elastic CD enabled - ComputeDomain use 0 nodes", func() {
			plugin := newTestPlugin(true)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithReplicas(2).WithGPU(1).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.HandleAdditionalResources(testCtx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodTemplateSpec(), k8sClient, state)

			Expect(err).NotTo(HaveOccurred())
			Expect(state.Issues()).To(HaveLen(0))

			cdList := &nvapi.ComputeDomainList{}
			err = k8sClient.List(testCtx, cdList, client.InNamespace(testNamespace))
			Expect(err).NotTo(HaveOccurred())
			Expect(cdList.Items).To(HaveLen(1))

			cd := cdList.Items[0]
			Expect(cd.Spec.NumNodes).To(Equal(0))
		})

		It("does not create ComputeDomain with None mode", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeNone, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithReplicas(2).WithGPU(1).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.HandleAdditionalResources(testCtx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodTemplateSpec(), k8sClient, state)

			Expect(err).NotTo(HaveOccurred())
			Expect(state.Issues()).To(HaveLen(0))

			cdList := &nvapi.ComputeDomainList{}
			err = k8sClient.List(testCtx, cdList, client.InNamespace(testNamespace))
			Expect(err).NotTo(HaveOccurred())
			Expect(cdList.Items).To(BeEmpty())
		})

		It("doesn't create ComputeDomain when no GPU resources found in workload", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithReplicas(2).Build()

			unstructuredJobset := convertToUnstructured(jobset)
			err := plugin.HandleAdditionalResources(testCtx, runner, unstructuredJobset, test.NewValidJobSetKartaWithPodTemplateSpec(), k8sClient, state)

			Expect(err).NotTo(HaveOccurred())
			Expect(state.Issues()).To(HaveLen(0))

			cdList := &nvapi.ComputeDomainList{}
			err = k8sClient.List(testCtx, cdList, client.InNamespace(testNamespace))
			Expect(err).NotTo(HaveOccurred())
			Expect(cdList.Items).To(BeEmpty())
		})
	})

	Describe("Status", func() {
		var k8sClient client.Client
		var testCtx context.Context

		BeforeEach(func() {
			k8sClient = fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
			testCtx = context.TODO()
		})

		It("returns error if status is nil", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)

			err := plugin.Status(testCtx, runner, k8sClient, nil, reconcile.New())

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("nil status for plugin mnnvl"))
		})

		It("if status is not point to runner .status.plugin[plugin-name], it will not change it", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			status := dromosv1alpha1.PluginStatus{}

			err := plugin.Status(testCtx, runner, k8sClient, &status, reconcile.New())

			Expect(err).NotTo(HaveOccurred())
			Expect(runner.Status.Plugins).ToNot(HaveKey(PluginName))
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
		})

		It("if status point to runner .status.plugin, it will change it", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			status := dromosv1alpha1.PluginStatus{}
			runner.Status.Plugins = make(map[string]*dromosv1alpha1.PluginStatus)
			runner.Status.Plugins["randomName"] = &status

			err := plugin.Status(testCtx, runner, k8sClient, &status, reconcile.New())

			Expect(err).NotTo(HaveOccurred())
			Expect(runner.Status.Plugins).To(HaveKey("randomName"))
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
		})

		It("plugin status is disabled when MNNVL config is nil", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			status := dromosv1alpha1.PluginStatus{}

			err := plugin.Status(testCtx, runner, k8sClient, &status, reconcile.New())

			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
		})

		It("removes plugin status when MNNVL mode is None", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeNone, testGPULabel))
			status := dromosv1alpha1.PluginStatus{}

			err := plugin.Status(testCtx, runner, k8sClient, &status, reconcile.New())

			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
		})

		It("sets status to Creating when ComputeDomain is not found", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			status := dromosv1alpha1.PluginStatus{}

			err := plugin.Status(testCtx, runner, k8sClient, &status, reconcile.New())
			Expect(err).NotTo(HaveOccurred())

			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseCreating))
			Expect(status.Conditions).To(HaveLen(1))

			condition := status.Conditions[0]
			Expect(condition.Type).To(Equal(ComputeDomainReadyConditionName))
			Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			Expect(condition.Reason).To(Equal("ComputeDomainNotFound"))
			Expect(condition.Message).To(ContainSubstring("not found"))
		})

		It("sets status to Running when ComputeDomain exists and is ready", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			status := dromosv1alpha1.PluginStatus{}

			computeDomain := &nvapi.ComputeDomain{
				ObjectMeta: metav1.ObjectMeta{
					Name:      composeComputeDomainName(workloadRunnerInitials, testWorkloadName),
					Namespace: testNamespace,
				},
				Spec: nvapi.ComputeDomainSpec{
					NumNodes: 2,
				},
			}
			err := k8sClient.Create(testCtx, computeDomain)
			Expect(err).NotTo(HaveOccurred())

			err = plugin.Status(testCtx, runner, k8sClient, &status, reconcile.New())
			Expect(err).NotTo(HaveOccurred())

			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
			Expect(status.Conditions).To(HaveLen(1))

			condition := status.Conditions[0]
			Expect(condition.Type).To(Equal(ComputeDomainReadyConditionName))
			Expect(condition.Status).To(Equal(metav1.ConditionTrue))
			Expect(condition.Reason).To(Equal("ComputeDomainReady"))
			Expect(condition.Message).To(ContainSubstring("is ready"))
		})

		It("updates existing plugin status when ComputeDomain state changes", func() {
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			status := &dromosv1alpha1.PluginStatus{}

			err := plugin.Status(testCtx, runner, k8sClient, status, reconcile.New())
			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseCreating))

			computeDomain := &nvapi.ComputeDomain{
				ObjectMeta: metav1.ObjectMeta{
					Name:      composeComputeDomainName(workloadRunnerInitials, testWorkloadName),
					Namespace: testNamespace,
				},
				Spec: nvapi.ComputeDomainSpec{
					NumNodes: 3,
				},
			}
			err = k8sClient.Create(testCtx, computeDomain)
			Expect(err).NotTo(HaveOccurred())

			err = plugin.Status(testCtx, runner, k8sClient, status, reconcile.New())
			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))

			condition := status.Conditions[0]
			Expect(condition.Status).To(Equal(metav1.ConditionTrue))
		})
	})
})

func verifyEmptyPodAffinity(podSpec *corev1.PodSpec) {
	if podSpec.Affinity == nil {
		return
	}
	if podSpec.Affinity.PodAffinity == nil {
		return
	}
	Expect(podSpec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(BeEmpty())
	Expect(podSpec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution).To(BeEmpty())
}

func topologyKeys(terms []corev1.WeightedPodAffinityTerm) []string {
	keys := make([]string, 0, len(terms))
	for _, t := range terms {
		keys = append(keys, t.PodAffinityTerm.TopologyKey)
	}
	return keys
}

func newTestPlugin(elasticEnabled bool) *Plugin {
	return NewPlugin(elasticEnabled, testOptions()...)
}

func newTestRunner(name, namespace string, mnnvlConfig *dromosv1alpha1.MnnvlConfig) *dromosv1alpha1.WorkloadRunner {
	runner := &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uuid.New().String()),
		},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{},
	}
	if mnnvlConfig != nil {
		runner.Spec.Instructions = &dromosv1alpha1.Instructions{
			GpuOptimizations: &dromosv1alpha1.GpuOptimizations{
				Mnnvl: mnnvlConfig,
			},
		}
	}
	return runner
}

func newMnnvlConfig(mode dromosv1alpha1.MnnvlMode, labelKeys ...string) *dromosv1alpha1.MnnvlConfig {
	return &dromosv1alpha1.MnnvlConfig{
		Mode:                            mode,
		GpuNetworkAccelerationLabelKeys: labelKeys,
	}
}

func convertToUnstructured(obj client.Object) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	Expect(err).NotTo(HaveOccurred())
	u.Object = data
	return u
}

func convertFromUnstructured(u *unstructured.Unstructured, obj client.Object) {
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, obj)
	Expect(err).NotTo(HaveOccurred())
}
