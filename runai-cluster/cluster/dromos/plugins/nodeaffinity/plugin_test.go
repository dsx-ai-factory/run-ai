/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodeaffinity

import (
	"context"
	"fmt"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

const (
	testNamespace    = "runai-test"
	testWorkloadName = "test-workload"
	hostnameLabel    = "kubernetes.io/hostname"
)

var _ = Describe("NodeAffinity Plugin", func() {
	var (
		ctx    context.Context
		state  *reconcile.State
		plugin *Plugin
	)

	BeforeEach(func() {
		ctx = context.Background()
		state = reconcile.New()
		plugin = NewPlugin(nil)
	})

	Describe("WebhookValidate", func() {
		It("passes when no instruction is set", func() {
			warnings, errs := plugin.WebhookValidate(ctx, nil, newRunner(nil))

			Expect(warnings).To(BeEmpty())
			Expect(errs).To(BeEmpty())
		})

		It("passes for a hostname exclusion", func() {
			warnings, errs := plugin.WebhookValidate(ctx, nil, newRunner(exclusionAffinity("node-a", "node-b")))

			Expect(warnings).To(BeEmpty())
			Expect(errs).To(BeEmpty())
		})

		It("rejects a NotIn requirement with no values", func() {
			affinity := requiredAffinity(corev1.NodeSelectorRequirement{
				Key:      hostnameLabel,
				Operator: corev1.NodeSelectorOpNotIn,
			})

			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(affinity))

			Expect(errs.ToAggregate()).To(MatchError(ContainSubstring("must have at least one value")))
		})

		It("rejects an Exists requirement that carries values", func() {
			affinity := requiredAffinity(corev1.NodeSelectorRequirement{
				Key:      hostnameLabel,
				Operator: corev1.NodeSelectorOpExists,
				Values:   []string{"node-a"},
			})

			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(affinity))

			Expect(errs.ToAggregate()).To(MatchError(ContainSubstring("may not be specified")))
		})

		It("rejects an unsupported operator", func() {
			affinity := requiredAffinity(corev1.NodeSelectorRequirement{
				Key:      hostnameLabel,
				Operator: "Matches",
				Values:   []string{"node-a"},
			})

			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(affinity))

			Expect(errs.ToAggregate()).To(MatchError(ContainSubstring(`Unsupported value: "Matches"`)))
		})

		It("rejects an invalid label value in a required term", func() {
			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(exclusionAffinity("not a valid/label value")))

			Expect(errs).To(HaveLen(1))
			Expect(errs[0].Type).To(Equal(field.ErrorTypeInvalid))
		})

		It("admits an empty preferred term", func() {
			affinity := &corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
					Weight:     50,
					Preference: corev1.NodeSelectorTerm{},
				}},
			}
			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(affinity))

			Expect(errs).To(BeEmpty())
		})

		It("admits an invalid label value in a preferred term", func() {
			affinity := &corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
					Weight: 50,
					Preference: corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: "zone", Operator: corev1.NodeSelectorOpNotIn, Values: []string{"us east"}},
					}},
				}},
			}
			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(affinity))

			Expect(errs).To(BeEmpty())
		})

		It("rejects a Gt value that is not an integer", func() {
			affinity := &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: "gpu-count", Operator: corev1.NodeSelectorOpGt, Values: []string{"many"}},
					}}},
				},
			}
			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(affinity))

			Expect(errs).To(HaveLen(1))
			Expect(errs[0].Type).To(Equal(field.ErrorTypeInvalid))
		})

		It("rejects an empty node selector term", func() {
			affinity := &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{}},
				},
			}

			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(affinity))

			Expect(errs).To(HaveLen(1))
			Expect(errs[0].Type).To(Equal(field.ErrorTypeRequired))
		})

		// Reachable only in this shape: an instruction counts as absent when
		// required AND preferred are both empty, so a preferred term is what keeps
		// an empty required block alive long enough to be rejected.
		It("rejects an empty required block alongside a preferred term", func() {
			affinity := &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{},
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
					Weight: 10,
					Preference: corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: hostnameLabel, Operator: corev1.NodeSelectorOpExists}}},
				}},
			}

			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(affinity))

			Expect(errs).To(HaveLen(1))
			Expect(errs[0].Type).To(Equal(field.ErrorTypeRequired))
			Expect(errs[0].Field).To(HaveSuffix("nodeSelectorTerms"))
		})

		It("accepts a well-formed matchFields requirement", func() {
			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(matchFieldsAffinity(corev1.NodeSelectorRequirement{
				Key: "metadata.name", Operator: corev1.NodeSelectorOpNotIn, Values: []string{"node-a"},
			})))

			Expect(errs).To(BeEmpty())
		})

		DescribeTable("rejects a matchFields requirement the apiserver would refuse",
			func(req corev1.NodeSelectorRequirement, expectedType field.ErrorType) {
				_, errs := plugin.WebhookValidate(ctx, nil, newRunner(matchFieldsAffinity(req)))

				Expect(errs).NotTo(BeEmpty())
				Expect(errs[0].Type).To(Equal(expectedType))
			},
			Entry("a key other than metadata.name",
				corev1.NodeSelectorRequirement{Key: "spec.unschedulable", Operator: corev1.NodeSelectorOpIn, Values: []string{"true"}},
				field.ErrorTypeNotSupported),
			Entry("an operator other than In/NotIn",
				corev1.NodeSelectorRequirement{Key: "metadata.name", Operator: corev1.NodeSelectorOpExists},
				field.ErrorTypeNotSupported),
			Entry("more than one value",
				corev1.NodeSelectorRequirement{Key: "metadata.name", Operator: corev1.NodeSelectorOpNotIn, Values: []string{"node-a", "node-b"}},
				field.ErrorTypeInvalid),
			Entry("a value that is not a valid node name",
				corev1.NodeSelectorRequirement{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{"Not A Node"}},
				field.ErrorTypeInvalid),
		)

		It("rejects more node selector terms than the merge can safely expand", func() {
			terms := make([]corev1.NodeSelectorTerm, maxTerms+1)
			for i := range terms {
				terms[i] = corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{
					{Key: hostnameLabel, Operator: corev1.NodeSelectorOpNotIn, Values: []string{"node-a"}}}}
			}
			affinity := &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: terms}}

			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(affinity))

			Expect(errs).NotTo(BeEmpty())
			Expect(errs[0].Type).To(Equal(field.ErrorTypeTooMany))
		})

		It("rejects a requirement carrying more values than the cap", func() {
			values := make([]string, maxValuesPerRequirement+1)
			for i := range values {
				values[i] = fmt.Sprintf("node-%d", i)
			}

			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(exclusionAffinity(values...)))

			Expect(errs).NotTo(BeEmpty())
			Expect(errs[0].Type).To(Equal(field.ErrorTypeTooMany))
		})

		// Clearing the instruction is how a constraint is lifted off a running
		// workload; it must stay admissible.
		It("admits removing the instruction", func() {
			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(nil))
			Expect(errs).To(BeEmpty())

			runner := newRunner(exclusionAffinity("node-a"))
			runner.Spec.Instructions.NodeAffinity = nil
			_, errs = plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).To(BeEmpty())
		})

		It("rejects an out-of-range preferred weight", func() {
			affinity := &corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
					Weight: 0,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{{Key: hostnameLabel, Operator: corev1.NodeSelectorOpExists}},
					},
				}},
			}

			_, errs := plugin.WebhookValidate(ctx, nil, newRunner(affinity))

			Expect(errs).To(HaveLen(1))
			Expect(errs[0].Field).To(ContainSubstring("weight"))
		})
	})

	DescribeTable("MutateResource across the pod-definition shapes Karta supports",
		func(testFunc func(*kartav1alpha1.Karta)) {
			By("using pod template spec Karta")
			testFunc(test.NewValidJobSetKartaWithPodTemplateSpec())

			By("using pod spec Karta")
			testFunc(test.NewValidJobSetKartaWithPodSpec())

			By("using fragmented pod spec Karta")
			testFunc(test.NewValidJobSetKartaWithFragmentedPodSpec())
		},
		Entry("lands the exclusion on every replicated job", func(karta *kartav1alpha1.Karta) {
			runner := newRunner(exclusionAffinity("node-a"))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build()

			mutated := mutate(ctx, NewPlugin(nil), runner, toUnstructured(jobset), karta, reconcile.New())

			forEachPodSpec(mutated, func(podSpec corev1.PodSpec) {
				Expect(nodeSelectorTerms(podSpec)).To(HaveLen(1))
				Expect(nodeSelectorTerms(podSpec)[0].MatchExpressions).To(ContainElement(
					corev1.NodeSelectorRequirement{
						Key:      hostnameLabel,
						Operator: corev1.NodeSelectorOpNotIn,
						Values:   []string{"node-a"},
					}))
			})
		}),
		Entry("preserves the workload's own node affinity", func(karta *kartav1alpha1.Karta) {
			runner := newRunner(exclusionAffinity("node-a"))
			userRequirement := corev1.NodeSelectorRequirement{
				Key:      "gpu-type",
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"a100"},
			}
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build()
			for i := range jobset.Spec.ReplicatedJobs {
				jobset.Spec.ReplicatedJobs[i].Template.Spec.Template.Spec.Affinity = &corev1.Affinity{
					NodeAffinity: requiredAffinity(userRequirement),
				}
			}

			mutated := mutate(ctx, NewPlugin(nil), runner, toUnstructured(jobset), karta, reconcile.New())

			forEachPodSpec(mutated, func(podSpec corev1.PodSpec) {
				Expect(nodeSelectorTerms(podSpec)[0].MatchExpressions).To(ContainElement(userRequirement))
			})
		}),
		Entry("converges: a second round changes nothing", func(karta *kartav1alpha1.Karta) {
			runner := newRunner(exclusionAffinity("node-b", "node-a"))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build()

			firstRound := mutate(ctx, NewPlugin(nil), runner, toUnstructured(jobset), karta, reconcile.New())
			secondRound := mutate(ctx, NewPlugin(nil), runner, firstRound, karta, reconcile.New())

			Expect(secondRound).To(Equal(firstRound))
		}),
		// An OR is the shape a cartesian merge re-multiplies, so drive it the way
		// the framework's mutation loop does.
		Entry("converges for a multi-term instruction", func(karta *kartav1alpha1.Karta) {
			runner := newRunner(&corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key: hostnameLabel, Operator: corev1.NodeSelectorOpNotIn, Values: []string{"node-a"},
						}}},
						{MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key: "zone", Operator: corev1.NodeSelectorOpIn, Values: []string{"us-east-1"},
						}}},
					},
				},
			})
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build()

			const maxMutationRounds = 5
			current := toUnstructured(jobset)
			var previous *unstructured.Unstructured
			settled := false
			for round := 1; round <= maxMutationRounds; round++ {
				previous = current
				current = mutate(ctx, NewPlugin(nil), runner, previous, karta, reconcile.New())
				if equality.Semantic.DeepEqual(previous, current) {
					settled = true
					break
				}
			}

			Expect(settled).To(BeTrue(), "mutation did not converge within %d rounds; the child would never be applied", maxMutationRounds)
			forEachPodSpec(current, func(podSpec corev1.PodSpec) {
				Expect(nodeSelectorTerms(podSpec)).To(HaveLen(2), "the instruction's OR branches must survive")
			})
		}),
		Entry("sorts exclusion values so the child is stable", func(karta *kartav1alpha1.Karta) {
			runner := newRunner(exclusionAffinity("node-c", "node-a", "node-b"))
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build()

			mutated := mutate(ctx, NewPlugin(nil), runner, toUnstructured(jobset), karta, reconcile.New())

			forEachPodSpec(mutated, func(podSpec corev1.PodSpec) {
				Expect(nodeSelectorTerms(podSpec)[0].MatchExpressions[0].Values).
					To(Equal([]string{"node-a", "node-b", "node-c"}))
			})
		}),
	)

	Describe("MutateResource", func() {
		It("is a no-op when no instruction is set", func() {
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build()
			resource := toUnstructured(jobset)
			original := resource.DeepCopy()

			err := plugin.MutateResource(ctx, newRunner(nil), resource, test.NewValidJobSetKarta(), state)

			Expect(err).NotTo(HaveOccurred())
			Expect(resource).To(Equal(original))
		})

		It("is a no-op for an instruction with no terms", func() {
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build()
			resource := toUnstructured(jobset)
			original := resource.DeepCopy()
			runner := newRunner(&corev1.NodeAffinity{})

			err := plugin.MutateResource(ctx, runner, resource, test.NewValidJobSetKarta(), state)

			Expect(err).NotTo(HaveOccurred())
			Expect(resource).To(Equal(original))

			status := &dromosv1alpha1.PluginStatus{}
			Expect(plugin.Status(ctx, runner, nil, status, state)).To(Succeed())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
		})

		It("reports a critical compute issue when the merge would exceed the term limit", func() {
			templateTerms := make([]corev1.NodeSelectorTerm, maxTerms+1)
			for i := range templateTerms {
				templateTerms[i] = corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{
					{Key: fmt.Sprintf("key-%d", i), Operator: corev1.NodeSelectorOpExists},
				}}
			}
			instructionTerms := make([]corev1.NodeSelectorTerm, maxTerms)
			for i := range instructionTerms {
				instructionTerms[i] = corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{
					{Key: fmt.Sprintf("zone-%d", i), Operator: corev1.NodeSelectorOpExists},
				}}
			}
			runner := newRunner(&corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: instructionTerms},
			})
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build()
			for i := range jobset.Spec.ReplicatedJobs {
				jobset.Spec.ReplicatedJobs[i].Template.Spec.Template.Spec.Affinity = &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: templateTerms},
					},
				}
			}

			err := plugin.MutateResource(ctx, runner, toUnstructured(jobset), test.NewValidJobSetKarta(), state)

			Expect(err).To(MatchError(ContainSubstring("exceeds the term limit")))
			Expect(state.PluginIssues(plugin)).To(HaveLen(1))
			Expect(state.PluginIssues(plugin)[0].IsCritical).To(BeTrue())
			Expect(state.PluginIssues(plugin)[0].Group).To(Equal(issue.Compute))
		})

		It("reports a critical issue when the Karta cannot express node affinity", func() {
			karta := test.NewValidJobSetKartaWithFragmentedPodSpec()
			for i := range karta.Spec.StructureDefinition.ChildComponents {
				karta.Spec.StructureDefinition.ChildComponents[i].SpecDefinition.FragmentedPodSpecDefinition.NodeAffinityPath = ptr.To("")
			}
			jobset := test.NewJobSetBuilder(testWorkloadName, testNamespace).WithDefault().Build()

			err := plugin.MutateResource(ctx, newRunner(exclusionAffinity("node-a")), toUnstructured(jobset), karta, state)

			Expect(err).To(HaveOccurred())
			Expect(state.PluginIssues(plugin)).To(HaveLen(1))
			Expect(state.PluginIssues(plugin)[0].IsCritical).To(BeTrue())
		})
	})

	Describe("Status", func() {
		It("reports Disabled when no instruction is set", func() {
			status := &dromosv1alpha1.PluginStatus{}

			Expect(plugin.Status(ctx, newRunner(nil), nil, status, state)).To(Succeed())

			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
			Expect(status.AdditionalInfo).To(BeNil())
		})

		It("publishes the applied affinity, normalized", func() {
			status := &dromosv1alpha1.PluginStatus{}

			Expect(plugin.Status(ctx, newRunner(exclusionAffinity("node-b", "node-a")), nil, status, state)).To(Succeed())

			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
			Expect(status.AdditionalInfo).NotTo(BeNil())
			Expect(string(status.AdditionalInfo.Raw)).To(ContainSubstring(`["node-a","node-b"]`))
		})

		It("does not claim Applied when the child was never applied", func() {
			runner := newRunner(exclusionAffinity("node-a"))
			// What a non-converging mutation loop or a failed apply leaves behind.
			state.AddControllerIssue(runner, context.Canceled, true, "", issue.Internal)
			status := &dromosv1alpha1.PluginStatus{}

			Expect(plugin.Status(ctx, runner, nil, status, state)).To(Succeed())

			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhasePending))
			Expect(status.Conditions[0].Reason).To(Equal(ReasonNotApplied))
			Expect(status.AdditionalInfo).To(BeNil())
		})

		// The framework reuses the PluginStatus across reconciles, so a stale
		// affinity would keep advertising a constraint that is no longer set.
		It("clears the published affinity once the instruction is removed", func() {
			status := &dromosv1alpha1.PluginStatus{}
			Expect(plugin.Status(ctx, newRunner(exclusionAffinity("node-a")), nil, status, state)).To(Succeed())
			Expect(status.AdditionalInfo).NotTo(BeNil())

			Expect(plugin.Status(ctx, newRunner(nil), nil, status, reconcile.New())).To(Succeed())

			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
			Expect(status.AdditionalInfo).To(BeNil())
		})

		It("reports Failed when the mutation recorded a critical issue", func() {
			runner := newRunner(exclusionAffinity("node-a"))
			Expect(state.AddPluginIssue(ctx, plugin, runner, context.Canceled, true, "", issue.Compute)).To(HaveOccurred())
			status := &dromosv1alpha1.PluginStatus{}

			Expect(plugin.Status(ctx, runner, nil, status, state)).To(Succeed())

			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseFailed))
		})
	})
})

func newRunner(nodeAffinity *corev1.NodeAffinity) *dromosv1alpha1.WorkloadRunner {
	runner := &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{Name: testWorkloadName, Namespace: testNamespace},
	}
	if nodeAffinity != nil {
		runner.Spec.Instructions = &dromosv1alpha1.Instructions{
			NodeAffinity: &dromosv1alpha1.NodeAffinityConfig{NodeAffinity: nodeAffinity},
		}
	}
	return runner
}

// exclusionAffinity builds the affinity the restart flow emits: keep the workload
// off the given hostnames.
func exclusionAffinity(hostnames ...string) *corev1.NodeAffinity {
	return requiredAffinity(corev1.NodeSelectorRequirement{
		Key:      hostnameLabel,
		Operator: corev1.NodeSelectorOpNotIn,
		Values:   hostnames,
	})
}

// matchFieldsAffinity builds an instruction whose single term uses matchFields
// rather than matchExpressions.
func matchFieldsAffinity(requirements ...corev1.NodeSelectorRequirement) *corev1.NodeAffinity {
	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchFields: requirements}},
		},
	}
}

func requiredAffinity(requirements ...corev1.NodeSelectorRequirement) *corev1.NodeAffinity {
	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: requirements}},
		},
	}
}

// mutate runs the plugin over a copy of resource and returns the mutated copy.
func mutate(
	ctx context.Context,
	plugin *Plugin,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	state *reconcile.State,
) *unstructured.Unstructured {
	mutated := resource.DeepCopy()

	Expect(plugin.MutateResource(ctx, runner, mutated, karta, state)).To(Succeed())
	Expect(state.Issues()).To(BeEmpty())

	return mutated
}

func forEachPodSpec(resource *unstructured.Unstructured, assert func(corev1.PodSpec)) {
	jobset := &jobsetv1alpha2.JobSet{}
	Expect(runtime.DefaultUnstructuredConverter.FromUnstructured(resource.Object, jobset)).To(Succeed())

	Expect(jobset.Spec.ReplicatedJobs).NotTo(BeEmpty())
	for _, replicatedJob := range jobset.Spec.ReplicatedJobs {
		assert(replicatedJob.Template.Spec.Template.Spec)
	}
}

func nodeSelectorTerms(podSpec corev1.PodSpec) []corev1.NodeSelectorTerm {
	Expect(podSpec.Affinity).NotTo(BeNil())
	Expect(podSpec.Affinity.NodeAffinity).NotTo(BeNil())
	Expect(podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).NotTo(BeNil())
	return podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
}

func toUnstructured(obj client.Object) *unstructured.Unstructured {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	Expect(err).NotTo(HaveOccurred())
	return &unstructured.Unstructured{Object: data}
}
