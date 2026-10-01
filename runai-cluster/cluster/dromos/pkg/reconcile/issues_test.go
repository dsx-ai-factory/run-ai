/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reconcile_test

import (
	"context"
	"errors"
	"testing"
	"time"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIssues(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Issues Suite")
}

var _ = Describe("Deferred plugin retries", func() {
	DescribeTable("preserves the earliest positive retry requested by plugins",
		func(delays []time.Duration, expected time.Duration) {
			state := reconcile.New()
			for _, delay := range delays {
				state.RequestRequeueAfter(delay)
			}
			Expect(state.RequeueAfter()).To(Equal(expected))
		},
		Entry("no retry requested", nil, time.Duration(0)),
		Entry("shorter retry takes priority", []time.Duration{time.Minute, 30 * time.Second}, 30*time.Second),
		Entry("later requests cannot postpone recovery", []time.Duration{30 * time.Second, time.Minute}, 30*time.Second),
		Entry("nonpositive requests cannot cancel recovery", []time.Duration{time.Minute, 0, -time.Second}, time.Minute),
	)
})

var _ = Describe("SetControllerIssues", func() {
	var (
		ctx    context.Context
		runner *dromosv1alpha1.WorkloadRunner
		state  *reconcile.State
	)

	BeforeEach(func() {
		ctx = context.Background()
		runner = &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runner",
				Namespace: "test-namespace",
			},
		}
		state = reconcile.New()
	})

	Context("when runner or state is nil", func() {
		It("handles nil runner gracefully", func() {
			reconcile.SetControllerIssues(ctx, nil, state)
			// Should not panic
		})

		It("handles nil state gracefully", func() {
			reconcile.SetControllerIssues(ctx, runner, nil)
			// Should not panic
		})
	})

	Context("when there are no issues", func() {
		It("clears controller issues", func() {
			runner.Status.Issues = []dromosv1alpha1.Issue{
				{
					Kind:        "Pod",
					Name:        "test-pod",
					Description: "old issue",
					Group:       issue.Compute.String(),
				},
			}

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Issues).To(BeNil())
		})

		It("clears plugin issues", func() {
			runner.Status.Plugins = map[string]*dromosv1alpha1.PluginStatus{
				"plugin1": {
					Phase: dromosv1alpha1.PluginPhaseReady,
					Issues: []dromosv1alpha1.Issue{
						{
							Kind:        "Pod",
							Name:        "test-pod",
							Description: "old issue",
							Group:       issue.Compute.String(),
						},
					},
				},
			}

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Plugins["plugin1"].Issues).To(BeNil())
		})
	})

	Context("when there are controller issues", func() {
		It("sets controller issues with correct fields", func() {
			pod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			}
			state.AddControllerIssue(pod, errors.New("test error"), true, "Failed to create pod", issue.Compute)

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Issues).To(HaveLen(1))
			controllerIssue := runner.Status.Issues[0]
			Expect(controllerIssue.Kind).To(Equal("Pod"))
			Expect(controllerIssue.Name).To(Equal("test-pod"))
			Expect(controllerIssue.Description).To(ContainSubstring("Failed to create pod"))
			Expect(controllerIssue.Group).To(Equal(issue.Compute.String()))
		})

		It("sets multiple controller issues", func() {
			pod1 := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-1",
					Namespace: "test-namespace",
				},
			}
			pod2 := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-2",
					Namespace: "test-namespace",
				},
			}

			state.AddControllerIssue(pod1, errors.New("error 1"), true, "Issue 1", issue.Compute)
			state.AddControllerIssue(pod2, errors.New("error 2"), false, "Issue 2", issue.Network)

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Issues).To(HaveLen(2))
			Expect(runner.Status.Issues[0].Name).To(Equal("test-pod-1"))
			Expect(runner.Status.Issues[0].Group).To(Equal(issue.Compute.String()))
			Expect(runner.Status.Issues[1].Name).To(Equal("test-pod-2"))
			Expect(runner.Status.Issues[1].Group).To(Equal(issue.Network.String()))
		})

		It("replaces old controller issues with new ones", func() {
			runner.Status.Issues = []dromosv1alpha1.Issue{
				{
					Kind:        "Pod",
					Name:        "old-pod",
					Description: "old issue",
					Group:       issue.Compute.String(),
				},
			}

			pod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "new-pod",
					Namespace: "test-namespace",
				},
			}
			state.AddControllerIssue(pod, errors.New("new error"), true, "New issue", issue.Storage)

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Issues).To(HaveLen(1))
			Expect(runner.Status.Issues[0].Name).To(Equal("new-pod"))
			Expect(runner.Status.Issues[0].Group).To(Equal(issue.Storage.String()))
		})
	})

	Context("when there are plugin issues", func() {
		var testReporter *testIssueReporter

		BeforeEach(func() {
			testReporter = &testIssueReporter{name: "test-plugin"}
		})

		It("sets plugin issues with correct fields", func() {
			pod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			}
			_ = state.AddPluginIssue(ctx, testReporter, pod, errors.New("plugin error"), true, "Plugin failed", issue.Network)

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Plugins).To(HaveKey("test-plugin"))
			pluginStatus := runner.Status.Plugins["test-plugin"]
			Expect(pluginStatus.Issues).To(HaveLen(1))

			pluginIssue := pluginStatus.Issues[0]
			Expect(pluginIssue.Kind).To(Equal("Pod"))
			Expect(pluginIssue.Name).To(Equal("test-pod"))
			Expect(pluginIssue.Description).To(ContainSubstring("Plugin failed"))
			Expect(pluginIssue.Group).To(Equal(issue.Network.String()))
		})

		It("sets issues for multiple plugins", func() {
			testReporter1 := &testIssueReporter{name: "plugin1"}
			testReporter2 := &testIssueReporter{name: "plugin2"}

			pod1 := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "test-namespace",
				},
			}
			pod2 := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-2",
					Namespace: "test-namespace",
				},
			}

			_ = state.AddPluginIssue(ctx, testReporter1, pod1, errors.New("error 1"), true, "Plugin 1 issue", issue.Compute)
			_ = state.AddPluginIssue(ctx, testReporter2, pod2, errors.New("error 2"), true, "Plugin 2 issue", issue.Storage)

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Plugins).To(HaveKey("plugin1"))
			Expect(runner.Status.Plugins).To(HaveKey("plugin2"))

			Expect(runner.Status.Plugins["plugin1"].Issues).To(HaveLen(1))
			Expect(runner.Status.Plugins["plugin1"].Issues[0].Name).To(Equal("pod-1"))
			Expect(runner.Status.Plugins["plugin1"].Issues[0].Group).To(Equal(issue.Compute.String()))

			Expect(runner.Status.Plugins["plugin2"].Issues).To(HaveLen(1))
			Expect(runner.Status.Plugins["plugin2"].Issues[0].Name).To(Equal("pod-2"))
			Expect(runner.Status.Plugins["plugin2"].Issues[0].Group).To(Equal(issue.Storage.String()))
		})

		It("preserves plugin phase when updating issues", func() {
			runner.Status.Plugins = map[string]*dromosv1alpha1.PluginStatus{
				"test-plugin": {
					Phase: dromosv1alpha1.PluginPhaseReady,
				},
			}

			pod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			}
			_ = state.AddPluginIssue(ctx, testReporter, pod, errors.New("error"), true, "New issue", issue.Compute)

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Plugins["test-plugin"].Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
			Expect(runner.Status.Plugins["test-plugin"].Issues).To(HaveLen(1))
		})

		It("creates new plugin status when plugin doesn't exist", func() {
			pod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			}
			_ = state.AddPluginIssue(ctx, testReporter, pod, errors.New("error"), true, "New plugin issue", issue.Internal)

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Plugins).To(HaveKey("test-plugin"))
			pluginStatus := runner.Status.Plugins["test-plugin"]
			Expect(pluginStatus.Phase).To(Equal(dromosv1alpha1.PluginPhaseUnknown))
			Expect(pluginStatus.Issues).To(HaveLen(1))
		})

		It("clears issues for plugins that had issues but now don't", func() {
			runner.Status.Plugins = map[string]*dromosv1alpha1.PluginStatus{
				"plugin1": {
					Phase: dromosv1alpha1.PluginPhaseReady,
					Issues: []dromosv1alpha1.Issue{
						{
							Kind:        "Pod",
							Name:        "old-pod",
							Description: "old issue",
							Group:       issue.Compute.String(),
						},
					},
				},
			}

			testReporter2 := &testIssueReporter{name: "plugin2"}
			pod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "new-pod",
					Namespace: "test-namespace",
				},
			}
			_ = state.AddPluginIssue(ctx, testReporter2, pod, errors.New("error"), true, "Plugin 2 issue", issue.Storage)

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Plugins["plugin1"].Issues).To(BeNil())
			Expect(runner.Status.Plugins["plugin2"].Issues).To(HaveLen(1))
		})
	})

	Context("when there are both controller and plugin issues", func() {
		It("sets both types of issues correctly", func() {
			testReporter := &testIssueReporter{name: "test-plugin"}

			controllerPod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "controller-pod",
					Namespace: "test-namespace",
				},
			}
			pluginPod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "plugin-pod",
					Namespace: "test-namespace",
				},
			}

			state.AddControllerIssue(controllerPod, errors.New("controller error"), true, "Controller issue", issue.Compute)
			_ = state.AddPluginIssue(ctx, testReporter, pluginPod, errors.New("plugin error"), true, "Plugin issue", issue.Network)

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Issues).To(HaveLen(1))
			Expect(runner.Status.Issues[0].Name).To(Equal("controller-pod"))

			Expect(runner.Status.Plugins).To(HaveKey("test-plugin"))
			Expect(runner.Status.Plugins["test-plugin"].Issues).To(HaveLen(1))
			Expect(runner.Status.Plugins["test-plugin"].Issues[0].Name).To(Equal("plugin-pod"))
		})
	})

	Context("when plugins map is nil", func() {
		It("initializes plugins map when adding plugin issues", func() {
			runner.Status.Plugins = nil
			testReporter := &testIssueReporter{name: "test-plugin"}

			pod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			}
			_ = state.AddPluginIssue(ctx, testReporter, pod, errors.New("error"), true, "Plugin issue", issue.Compute)

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Plugins).NotTo(BeNil())
			Expect(runner.Status.Plugins).To(HaveKey("test-plugin"))
		})
	})

	Context("IsCritical propagation", func() {
		makePod := func(name string) *corev1.Pod {
			return &corev1.Pod{
				TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "test-namespace"},
			}
		}

		It("propagates IsCritical onto controller issues", func() {
			state.AddControllerIssue(makePod("crit-pod"), errors.New("err"), true, "critical", issue.Compute)
			state.AddControllerIssue(makePod("noncrit-pod"), errors.New("err"), false, "non-critical", issue.Network)

			reconcile.SetControllerIssues(ctx, runner, state)

			byName := map[string]bool{}
			for _, iss := range runner.Status.Issues {
				byName[iss.Name] = iss.IsCritical
			}
			Expect(byName).To(HaveKeyWithValue("crit-pod", true))
			Expect(byName).To(HaveKeyWithValue("noncrit-pod", false))
		})

		It("propagates IsCritical onto plugin issues", func() {
			reporter := &testIssueReporter{name: "test-plugin"}
			_ = state.AddPluginIssue(ctx, reporter, makePod("crit-pod"), errors.New("err"), true, "critical", issue.Compute)

			reconcile.SetControllerIssues(ctx, runner, state)

			Expect(runner.Status.Plugins["test-plugin"].Issues).To(HaveLen(1))
			Expect(runner.Status.Plugins["test-plugin"].Issues[0].IsCritical).To(BeTrue())
		})
	})

	Context("issue group mapping", func() {
		It("correctly maps all issue groups", func() {
			groups := []issue.Group{
				issue.Compute,
				issue.Network,
				issue.Storage,
				issue.Internal,
			}

			for i, group := range groups {
				state = reconcile.New()
				pod := &corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-namespace",
					},
				}
				state.AddControllerIssue(pod, errors.New("test error"), true, "Test issue", group)

				runner = &dromosv1alpha1.WorkloadRunner{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runner",
						Namespace: "test-namespace",
					},
				}
				reconcile.SetControllerIssues(ctx, runner, state)

				Expect(runner.Status.Issues).To(HaveLen(1), "Failed for group %d: %s", i, group)
				Expect(runner.Status.Issues[0].Group).To(Equal(group.String()), "Failed for group %d: %s", i, group)
			}
		})
	})
})

var _ = Describe("HasCriticalControllerIssues", func() {
	var state *reconcile.State

	BeforeEach(func() {
		state = reconcile.New()
	})

	newPod := func(name string) *corev1.Pod {
		return &corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "test-namespace",
			},
		}
	}

	It("returns false when there are no issues", func() {
		Expect(state.HasCriticalControllerIssues()).To(BeFalse())
	})

	It("returns false when there are only non-critical issues", func() {
		state.AddControllerIssue(newPod("pod-1"), errors.New("err"), false, "non-critical", issue.Compute)
		state.AddControllerIssue(newPod("pod-2"), errors.New("err"), false, "also non-critical", issue.Network)

		Expect(state.HasCriticalControllerIssues()).To(BeFalse())
	})

	It("returns true when at least one issue is critical", func() {
		state.AddControllerIssue(newPod("pod-1"), errors.New("err"), false, "non-critical", issue.Compute)
		state.AddControllerIssue(newPod("pod-2"), errors.New("err"), true, "critical", issue.Compute)

		Expect(state.HasCriticalControllerIssues()).To(BeTrue())
	})

	It("ignores plugin issues", func() {
		ctx := context.Background()
		reporter := &testIssueReporter{name: "plugin"}
		_ = state.AddPluginIssue(ctx, reporter, newPod("pod-1"), errors.New("err"), true, "critical plugin issue", issue.Compute)

		Expect(state.HasCriticalControllerIssues()).To(BeFalse())
	})
})

// testIssueReporter implements the reconcile.Reporter interface for testing
type testIssueReporter struct {
	name string
}

func (m *testIssueReporter) Name() string {
	return m.name
}
