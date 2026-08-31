/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package framework

import (
	"context"
	"errors"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	"k8s.io/apimachinery/pkg/runtime"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestFramework(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Framework Suite")
}

var _ = Describe("NewFramework", func() {
	It("assembles plugins in name order regardless of registry map order", func() {
		reg := NewRegistry(
			newMockResourceMutatePlugin("zebra", 0, nil),
			newMockResourceMutatePlugin("alpha", 0, nil),
			newMockResourceMutatePlugin("mango", 0, nil),
		)

		for range 10 {
			f := NewFramework(reg, nil)
			ordered := make([]string, 0, len(f.resourceMutatePlugins))
			for _, p := range f.resourceMutatePlugins {
				ordered = append(ordered, p.Name())
			}
			Expect(ordered).To(Equal([]string{"alpha", "mango", "zebra"}))
		}
	})
})

var _ = Describe("RunWatchers", func() {
	It("calls plugin and returns builder", func() {
		reg := NewRegistry(newMockWatcherPlugin(nil), newEmptyPlugin())
		f := NewFramework(reg, nil)
		bldr := &builder.Builder{}

		result, err := f.RunWatchers(bldr)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(bldr))
	})

	It("returns error when plugin fails", func() {
		expectedErr := errors.New("watcher error")
		plugin := newMockWatcherPlugin(expectedErr)
		reg := NewRegistry(plugin)
		f := NewFramework(reg, nil)
		bldr := &builder.Builder{}

		result, err := f.RunWatchers(bldr)

		Expect(err).To(MatchError(expectedErr))
		Expect(result).To(BeNil())
	})
})

var _ = Describe("RunValidators", func() {
	It("runs validator plugins successfully with no errors", func() {
		validatorPlugin := newMockValidatorPlugin("validatorPlugin", nil, nil)
		reg := NewRegistry(validatorPlugin)
		f := NewFramework(reg, nil)

		runner := createRunnerWorkload()
		warnings, fieldErrs := f.RunValidators(context.Background(), nil, runner)

		Expect(warnings).To(BeEmpty())
		Expect(fieldErrs).To(BeEmpty())
	})

	It("returns field errors when validation fails", func() {
		errMsg := "template must have namespace field set"
		fieldErr := field.Required(field.NewPath("spec", "template", "namespace"), errMsg)
		validatorPlugin := newMockValidatorPlugin("validatorPlugin", nil, field.ErrorList{fieldErr})
		reg := NewRegistry(validatorPlugin)
		f := NewFramework(reg, nil)

		runner := createRunnerWorkload()
		warnings, fieldErrs := f.RunValidators(context.Background(), nil, runner)

		Expect(warnings).To(BeEmpty())
		Expect(fieldErrs).To(ConsistOf(fieldErr))
	})

	It("returns warnings when validation produces warnings", func() {
		warningMsg := "warning message"
		validatorPlugin := newMockValidatorPlugin("validatorPlugin", admission.Warnings{warningMsg}, nil)
		reg := NewRegistry(validatorPlugin)
		f := NewFramework(reg, nil)

		runner := createRunnerWorkload()
		warnings, fieldErrs := f.RunValidators(context.Background(), nil, runner)

		Expect(warnings).To(ConsistOf(warningMsg))
		Expect(fieldErrs).To(BeEmpty())
	})

	It("aggregates errors from multiple plugins", func() {
		fieldErr1 := field.Required(field.NewPath("spec", "template", "namespace"), "namespace required")
		fieldErr2 := field.Invalid(field.NewPath("spec", "template", "name"), "", "name invalid")
		validator1 := newMockValidatorPlugin("validator1", nil, field.ErrorList{fieldErr1})
		validator2 := newMockValidatorPlugin("validator2", nil, field.ErrorList{fieldErr2})
		reg := NewRegistry(validator1, validator2)
		f := NewFramework(reg, nil)

		runner := createRunnerWorkload()
		warnings, fieldErrs := f.RunValidators(context.Background(), nil, runner)

		Expect(warnings).To(BeEmpty())
		Expect(fieldErrs).To(ConsistOf(fieldErr1, fieldErr2))
	})
})

var _ = Describe("RunMutators", func() {
	It("runs mutator plugins successfully with no errors", func() {
		mutatorPlugin := newMockMutatorPlugin("mutatorPlugin", nil, nil)
		reg := NewRegistry(mutatorPlugin)
		f := NewFramework(reg, nil)

		runner := &dromosv1alpha1.WorkloadRunner{}
		warnings, err := f.RunMutators(context.Background(), runner)

		Expect(err).NotTo(HaveOccurred())
		Expect(warnings).To(BeEmpty())
	})

	It("applies mutations to the runner", func() {
		mutatorPlugin := newMockMutatorPluginWithLabelMutation("mutatorPlugin")
		reg := NewRegistry(mutatorPlugin)
		f := NewFramework(reg, nil)

		runner := &dromosv1alpha1.WorkloadRunner{}
		warnings, err := f.RunMutators(context.Background(), runner)

		Expect(err).NotTo(HaveOccurred())
		Expect(warnings).To(BeEmpty())
		Expect(runner.Labels).NotTo(BeNil())
		Expect(runner.Labels["example"]).To(Equal("true"))
	})

	It("returns warnings when mutation produces warnings", func() {
		warningMsg := "mutation warning"
		mutatorPlugin := newMockMutatorPlugin("mutatorPlugin", admission.Warnings{"mutation warning"}, nil)
		reg := NewRegistry(mutatorPlugin)
		f := NewFramework(reg, nil)

		runner := &dromosv1alpha1.WorkloadRunner{}
		warnings, err := f.RunMutators(context.Background(), runner)

		Expect(err).NotTo(HaveOccurred())
		Expect(warnings).To(ConsistOf(warningMsg))
	})

	It("returns error when mutation fails", func() {
		expectedErr := errors.New("mutation error")
		mutatorPlugin := newMockMutatorPlugin("mutatorPlugin", nil, expectedErr)
		reg := NewRegistry(mutatorPlugin)
		f := NewFramework(reg, nil)

		runner := &dromosv1alpha1.WorkloadRunner{}
		warnings, err := f.RunMutators(context.Background(), runner)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("mutation failed in plugin mutatorPlugin"))
		Expect(warnings).To(BeNil())
	})
})

var _ = Describe("RunMutateResource", func() {
	var (
		ctx              context.Context
		runner           *dromosv1alpha1.WorkloadRunner
		attachedResource *unstructured.Unstructured
		karta            *kartav1alpha1.Karta
		state            *reconcile.State
	)

	BeforeEach(func() {
		ctx = context.Background()
		state = reconcile.New()
		runner = &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runner",
				Namespace: "test-ns",
			},
		}
		attachedResource = &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]any{
					"name":      "test-job",
					"namespace": "test-ns",
				},
			},
		}
		karta = &kartav1alpha1.Karta{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-karta",
			},
			Spec: kartav1alpha1.KartaSpec{
				StructureDefinition: kartav1alpha1.StructureDefinition{
					RootComponent: kartav1alpha1.ComponentDefinition{
						Name: "root",
						Kind: &kartav1alpha1.GroupVersionKind{
							Group:   "batch",
							Version: "v1",
							Kind:    "Job",
						},
						SpecDefinition: &kartav1alpha1.SpecDefinition{
							PodSpecPath: ptr.To(".spec.template.spec"),
						},
					},
				},
			},
		}
	})

	It("runs multiple rounds when plugins mutate, then stops when no mutations occur", func() {
		plugin := newMockResourceMutatePlugin("reinvokePlugin", 1, nil)
		reg := NewRegistry(plugin)
		f := NewFramework(reg, nil)

		err := f.RunMutateResource(ctx, runner, attachedResource, karta, state)

		Expect(err).NotTo(HaveOccurred())
		Expect(plugin.callCount).To(Equal(2))
		Expect(state.Issues()).To(HaveLen(0))
	})

	It("returns error when max rounds exceeded", func() {
		plugin := newMockResourceMutatePlugin("alwaysMutatePlugin", 11, nil)
		reg := NewRegistry(plugin)
		config := &Config{
			MaxMutationRounds: 10,
		}
		f := NewFramework(reg, config)

		err := f.RunMutateResource(ctx, runner, attachedResource, karta, state)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("exceeded max rounds"))
		Expect(plugin.callCount).To(Equal(config.MaxMutationRounds))
		Expect(state.Issues()).To(HaveLen(1))

		stateIssue := state.Issues()[0]
		Expect(stateIssue.Error()).To(Equal("[Job:test-job] MutateResource exceeded max rounds (10) - mutations did not converge"))
		Expect(stateIssue.Group).To(Equal(issue.Internal))
		Expect(stateIssue.GVK.Kind).To(Equal("Job"))
		Expect(stateIssue.Name).To(Equal("test-job"))
	})

	It("max rounds can be configured", func() {
		plugin := newMockResourceMutatePlugin("alwaysMutatePlugin", 3, nil)
		reg := NewRegistry(plugin)
		config := &Config{
			MaxMutationRounds: 2,
		}
		f := NewFramework(reg, config)
		err := f.RunMutateResource(ctx, runner, attachedResource, karta, state)

		Expect(err).To(HaveOccurred())
		Expect(state.Issues()).To(HaveLen(1))

		stateIssue := state.Issues()[0]
		Expect(stateIssue.Error()).To(Equal("[Job:test-job] MutateResource exceeded max rounds (2) - mutations did not converge"))
		Expect(stateIssue.Group).To(Equal(issue.Internal))
		Expect(stateIssue.GVK.Kind).To(Equal("Job"))
		Expect(stateIssue.Name).To(Equal("test-job"))
	})
})

var _ = Describe("RunResourcesStatus", func() {
	var (
		ctx       context.Context
		runner    *dromosv1alpha1.WorkloadRunner
		k8sClient client.Client
		state     *reconcile.State
	)

	BeforeEach(func() {
		ctx = context.Background()
		runner = &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runner",
				Namespace: "test-ns",
			},
			Status: dromosv1alpha1.WorkloadRunnerStatus{
				Plugins: make(map[string]*dromosv1alpha1.PluginStatus),
			},
		}
		k8sClient = nil
		state = reconcile.New()
	})

	It("runs status plugins successfully with no errors", func() {
		statusPlugin := newMockStatusPlugin("statusPlugin", nil)
		reg := NewRegistry(statusPlugin)
		f := NewFramework(reg, nil)

		err := f.RunResourcesStatus(ctx, runner, k8sClient, state)

		Expect(err).NotTo(HaveOccurred())
		Expect(runner.Status.Plugins).To(HaveKey("statusPlugin"))
	})

	It("returns error when multiple plugins fail", func() {
		err1 := errors.New("status error from plugin1")
		err2 := errors.New("status error from plugin2")
		status1 := newMockStatusPlugin("status1", err1)
		status2 := newMockStatusPlugin("status2", err2)
		reg := NewRegistry(status1, status2)
		f := NewFramework(reg, nil)

		err := f.RunResourcesStatus(ctx, runner, k8sClient, state)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("status failed in plugin status1"))
		Expect(err.Error()).To(ContainSubstring("status failed in plugin status2"))
		Expect(runner.Status.Plugins).To(HaveKey("status1"))
		Expect(runner.Status.Plugins).To(HaveKey("status2"))
	})

	It("passes state with issues from earlier phases to status plugins", func() {
		issueCheckPlugin := &mockStatusPluginWithIssueCheck{name: "issuePlugin"}
		reg := NewRegistry(issueCheckPlugin)
		f := NewFramework(reg, nil)

		_ = state.AddPluginIssue(ctx, issueCheckPlugin, runner, errors.New("earlier phase error"), true, "test issue", issue.Internal)

		err := f.RunResourcesStatus(ctx, runner, k8sClient, state)

		Expect(err).NotTo(HaveOccurred())
		Expect(issueCheckPlugin.receivedIssues).To(HaveLen(1))
		Expect(issueCheckPlugin.receivedIssues[0].IsCritical).To(BeTrue())
	})
})

var _ = Describe("RunPostPhase", func() {
	var (
		ctx       context.Context
		runner    *dromosv1alpha1.WorkloadRunner
		k8sClient client.Client
	)

	BeforeEach(func() {
		ctx = context.Background()
		runner = createRunnerWorkload()
		k8sClient = nil
	})

	It("calls post-phase plugins successfully with no error", func() {
		plugin := &mockPostPhasePlugin{name: "postPhase1"}
		f := NewFramework(NewRegistry(plugin), nil)

		err := f.RunPostPhase(ctx, runner, nil, nil, k8sClient)

		Expect(err).NotTo(HaveOccurred())
		Expect(plugin.called).To(BeTrue())
	})

	It("returns wrapped error including plugin name when a plugin fails", func() {
		expectedErr := errors.New("post-phase boom")
		plugin := &mockPostPhasePlugin{name: "postPhase1", err: expectedErr}
		f := NewFramework(NewRegistry(plugin), nil)

		err := f.RunPostPhase(ctx, runner, nil, nil, k8sClient)

		Expect(err).To(MatchError(expectedErr))
		Expect(err.Error()).To(ContainSubstring("HandlePostPhase failed in plugin postPhase1"))
	})
})

func createRunnerWorkload() *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "runner",
			Namespace: "runner-namespace",
		},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Template: runtime.RawExtension{
				Raw: []byte(`{"apiVersion":"v1","kind":"Pod","metadata":{"name":"pod", "namespace":"runner-namespace"}}`),
			},
		},
	}
}

type emptyPlugin struct {
	name string
}

func (e *emptyPlugin) Name() string {
	return e.name
}

func newEmptyPlugin() Plugin {
	return &emptyPlugin{
		name: "empty",
	}
}

type mockWatcherPlugin struct {
	name string
	err  error
}

func (m *mockWatcherPlugin) Name() string {
	return m.name
}

func (m *mockWatcherPlugin) Watchers(bldr *builder.Builder) (*builder.Builder, error) {
	if m.err != nil {
		return nil, m.err
	}

	return bldr, nil
}

func newMockWatcherPlugin(err error) WatcherPlugin {
	return &mockWatcherPlugin{
		name: "watcherPlugin",
		err:  err,
	}
}

type mockValidatorPlugin struct {
	name     string
	warnings admission.Warnings
	errors   field.ErrorList
}

func (m *mockValidatorPlugin) Name() string {
	return m.name
}

func (m *mockValidatorPlugin) WebhookValidate(_ context.Context, _, _ *dromosv1alpha1.WorkloadRunner) (admission.Warnings, field.ErrorList) {
	return m.warnings, m.errors
}

func newMockValidatorPlugin(name string, warnings admission.Warnings, errors field.ErrorList) ValidatorPlugin {
	return &mockValidatorPlugin{
		name:     name,
		warnings: warnings,
		errors:   errors,
	}
}

type mockMutatorPlugin struct {
	name     string
	warnings admission.Warnings
	err      error
}

func (m *mockMutatorPlugin) Name() string {
	return m.name
}

func (m *mockMutatorPlugin) WebhookMutate(_ context.Context, _ *dromosv1alpha1.WorkloadRunner) (admission.Warnings, error) {
	return m.warnings, m.err
}

func newMockMutatorPlugin(name string, warnings admission.Warnings, err error) MutatorPlugin {
	return &mockMutatorPlugin{
		name:     name,
		warnings: warnings,
		err:      err,
	}
}

type mockMutatorPluginWithLabelMutation struct {
	name string
}

func (m *mockMutatorPluginWithLabelMutation) Name() string {
	return m.name
}

func (m *mockMutatorPluginWithLabelMutation) WebhookMutate(_ context.Context, runner *dromosv1alpha1.WorkloadRunner) (admission.Warnings, error) {
	if runner.Labels == nil {
		runner.Labels = make(map[string]string)
	}
	runner.Labels["example"] = "true"
	return nil, nil
}

func newMockMutatorPluginWithLabelMutation(name string) MutatorPlugin {
	return &mockMutatorPluginWithLabelMutation{
		name: name,
	}
}

type mockResourceMutatePlugin struct {
	name             string
	mutateUntilRound int
	callCount        int
	err              error
}

func newMockResourceMutatePlugin(name string, mutateUntilRound int, err error) *mockResourceMutatePlugin {
	return &mockResourceMutatePlugin{
		name:             name,
		mutateUntilRound: mutateUntilRound,
		err:              err,
	}
}

func (m *mockResourceMutatePlugin) Name() string {
	return m.name
}

func (m *mockResourceMutatePlugin) MutateResource(_ context.Context, _ *dromosv1alpha1.WorkloadRunner, attachedResource *unstructured.Unstructured, _ *kartav1alpha1.Karta, _ *reconcile.State) error {
	if m.err != nil {
		return m.err
	}
	m.callCount++
	if m.callCount <= m.mutateUntilRound {
		annotations := attachedResource.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations["mock-annotation"] = strconv.Itoa(m.callCount)
		attachedResource.SetAnnotations(annotations)
	}
	return nil
}

type mockStatusPlugin struct {
	name string
	err  error
}

func newMockStatusPlugin(name string, err error) StatusPlugin {
	return &mockStatusPlugin{
		name: name,
		err:  err,
	}
}

func (m *mockStatusPlugin) Name() string {
	return m.name
}

func (m *mockStatusPlugin) Status(_ context.Context, _ *dromosv1alpha1.WorkloadRunner, _ client.Client, _ PluginStatusSetter, _ *reconcile.State) error {
	return m.err
}

type mockStatusPluginWithIssueCheck struct {
	name           string
	receivedIssues []issue.Issue
}

func (m *mockStatusPluginWithIssueCheck) Name() string {
	return m.name
}

func (m *mockStatusPluginWithIssueCheck) Status(_ context.Context, _ *dromosv1alpha1.WorkloadRunner, _ client.Client, _ PluginStatusSetter, state *reconcile.State) error {
	m.receivedIssues = state.PluginIssues(m)
	return nil
}

type mockPostPhasePlugin struct {
	name   string
	err    error
	called bool
}

func (m *mockPostPhasePlugin) Name() string {
	return m.name
}

func (m *mockPostPhasePlugin) HandlePostPhase(_ context.Context, _ *dromosv1alpha1.WorkloadRunner, _ *unstructured.Unstructured, _ *kartav1alpha1.Karta, _ client.Client) error {
	m.called = true
	return m.err
}
