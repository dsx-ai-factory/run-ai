/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("calcPhase", func() {
	var (
		newController = func() *WorkloadRunnerController { return &WorkloadRunnerController{} }
		newKarta      = func() *kartav1alpha1.Karta { return &kartav1alpha1.Karta{} }
		newResource   = func() *unstructured.Unstructured { return &unstructured.Unstructured{} }
		newPlugins    = func(phases map[string]dromosv1alpha1.PluginPhase) map[string]*dromosv1alpha1.PluginStatus {
			if phases == nil {
				return nil
			}
			out := make(map[string]*dromosv1alpha1.PluginStatus, len(phases))
			for name, p := range phases {
				out[name] = &dromosv1alpha1.PluginStatus{Phase: p}
			}
			return out
		}
	)

	It("returns Unknown when karta is nil", func() {
		c := newController()
		resource := newResource()
		resourcePhase := dromosv1alpha1.WorkloadRunnerPhaseRunning

		got := c.calcPhase(nil, resource, resourcePhase, nil, false)

		Expect(got).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseUnknown))
	})

	It("returns Unknown when resource is nil", func() {
		c := newController()
		karta := newKarta()
		resourcePhase := dromosv1alpha1.WorkloadRunnerPhaseRunning

		got := c.calcPhase(karta, nil, resourcePhase, nil, false)

		Expect(got).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseUnknown))
	})

	DescribeTable("resource phase and plugin phases",
		func(resourcePhase, expected dromosv1alpha1.WorkloadRunnerPhase, pluginPhases map[string]dromosv1alpha1.PluginPhase, controllerHasCriticalIssues bool) {
			c := newController()
			karta := newKarta()
			resource := newResource()
			plugins := newPlugins(pluginPhases)

			got := c.calcPhase(karta, resource, resourcePhase, plugins, controllerHasCriticalIssues)

			Expect(got).To(Equal(expected))
		},
		Entry("resource Failed dominates", dromosv1alpha1.WorkloadRunnerPhaseFailed, dromosv1alpha1.WorkloadRunnerPhaseFailed, map[string]dromosv1alpha1.PluginPhase{"p": dromosv1alpha1.PluginPhaseInitializing}, false),
		Entry("resource Completed dominates", dromosv1alpha1.WorkloadRunnerPhaseCompleted, dromosv1alpha1.WorkloadRunnerPhaseCompleted, map[string]dromosv1alpha1.PluginPhase{"p": dromosv1alpha1.PluginPhaseFailed}, false),
		Entry("resource Suspended", dromosv1alpha1.WorkloadRunnerPhaseSuspended, dromosv1alpha1.WorkloadRunnerPhaseSuspended,
			map[string]dromosv1alpha1.PluginPhase{"a": dromosv1alpha1.PluginPhaseReady, "b": dromosv1alpha1.PluginPhaseFailed, "c": dromosv1alpha1.PluginPhaseInitializing}, false),
		Entry("resource Stopped", dromosv1alpha1.WorkloadRunnerPhaseStopped, dromosv1alpha1.WorkloadRunnerPhaseStopped,
			map[string]dromosv1alpha1.PluginPhase{"a": dromosv1alpha1.PluginPhaseReady, "b": dromosv1alpha1.PluginPhaseFailed, "c": dromosv1alpha1.PluginPhaseInitializing}, false),
		Entry("plugins nil yields resource phase", dromosv1alpha1.WorkloadRunnerPhaseRunning, dromosv1alpha1.WorkloadRunnerPhaseRunning, nil, false),
		Entry("plugins empty yields resource phase", dromosv1alpha1.WorkloadRunnerPhaseRunning, dromosv1alpha1.WorkloadRunnerPhaseRunning, map[string]dromosv1alpha1.PluginPhase{}, false),
		Entry("all plugins Ready yields resource phase", dromosv1alpha1.WorkloadRunnerPhaseInitializing, dromosv1alpha1.WorkloadRunnerPhaseInitializing,
			map[string]dromosv1alpha1.PluginPhase{"a": dromosv1alpha1.PluginPhaseReady, "b": dromosv1alpha1.PluginPhaseReady}, false),
		Entry("all plugins Disabled yields resource phase", dromosv1alpha1.WorkloadRunnerPhaseInitializing, dromosv1alpha1.WorkloadRunnerPhaseInitializing, map[string]dromosv1alpha1.PluginPhase{"a": dromosv1alpha1.PluginPhaseDisabled}, false),
		Entry("resource Running and plugin Failed yields Degraded", dromosv1alpha1.WorkloadRunnerPhaseRunning, dromosv1alpha1.WorkloadRunnerPhaseDegraded, map[string]dromosv1alpha1.PluginPhase{"p": dromosv1alpha1.PluginPhaseFailed}, false),
		Entry("resource Creating and plugin Failed yields Degraded", dromosv1alpha1.WorkloadRunnerPhaseCreating, dromosv1alpha1.WorkloadRunnerPhaseDegraded, map[string]dromosv1alpha1.PluginPhase{"p": dromosv1alpha1.PluginPhaseFailed}, false),
		Entry("resource Running and mixed Ready/Failed yields Degraded", dromosv1alpha1.WorkloadRunnerPhaseRunning, dromosv1alpha1.WorkloadRunnerPhaseDegraded,
			map[string]dromosv1alpha1.PluginPhase{"a": dromosv1alpha1.PluginPhaseReady, "b": dromosv1alpha1.PluginPhaseFailed}, false),
		Entry("resource Running and plugin Creating yields Initializing", dromosv1alpha1.WorkloadRunnerPhaseRunning, dromosv1alpha1.WorkloadRunnerPhaseInitializing, map[string]dromosv1alpha1.PluginPhase{"p": dromosv1alpha1.PluginPhaseCreating}, false),
		Entry("resource Initializing and plugin Creating yields Initializing", dromosv1alpha1.WorkloadRunnerPhaseInitializing,
			dromosv1alpha1.WorkloadRunnerPhaseInitializing, map[string]dromosv1alpha1.PluginPhase{"p": dromosv1alpha1.PluginPhaseCreating}, false),
		Entry("resource Pending and plugin Pending yields Pending", dromosv1alpha1.WorkloadRunnerPhasePending, dromosv1alpha1.WorkloadRunnerPhasePending, map[string]dromosv1alpha1.PluginPhase{"p": dromosv1alpha1.PluginPhasePending}, false),

		// Critical controller issues
		Entry("resource Running + critical issues + no plugins yields Degraded", dromosv1alpha1.WorkloadRunnerPhaseRunning, dromosv1alpha1.WorkloadRunnerPhaseDegraded, nil, true),
		Entry("resource Creating + critical issues yields Degraded", dromosv1alpha1.WorkloadRunnerPhaseCreating, dromosv1alpha1.WorkloadRunnerPhaseDegraded, nil, true),
		Entry("resource Failed + critical issues yields Failed (terminal takes precedence)", dromosv1alpha1.WorkloadRunnerPhaseFailed, dromosv1alpha1.WorkloadRunnerPhaseFailed, nil, true),
		Entry("resource Completed + critical issues yields Completed (terminal takes precedence)", dromosv1alpha1.WorkloadRunnerPhaseCompleted, dromosv1alpha1.WorkloadRunnerPhaseCompleted, nil, true),
	)
})
