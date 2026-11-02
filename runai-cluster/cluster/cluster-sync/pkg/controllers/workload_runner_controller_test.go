/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controllers Suite")
}

const (
	inferenceEndpoints        = `{"schemaVersion":"v1","endpoints":[{"services":[{"ref":{"name":"model","port":8000}}],"externalAccess":{"httpRoute":{},"inference":{"epp":{"profile":"gie-v1.5"}}}}]}`
	endpointsWithoutInference = `{"schemaVersion":"v1","endpoints":[{"services":[{"ref":{"name":"model","port":8000}}],"externalAccess":{"httpRoute":{}}}]}`
)

var _ = Describe("wlrSyncFieldsChanged", func() {
	It("syncs inference block removal before stale plugin status is replaced", func() {
		oldWlr := wlrWithPluginDetails(networkingPluginName, nil)
		oldWlr.Annotations = map[string]string{endpoints.Annotation: inferenceEndpoints}
		newWlr := oldWlr.DeepCopy()
		newWlr.Annotations = map[string]string{endpoints.Annotation: endpointsWithoutInference}
		Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(BeTrue())
	})

	It("syncs an endpoints change while inference is requested", func() {
		oldWlr := wlrWithPluginDetails(networkingPluginName, nil)
		oldWlr.Annotations = map[string]string{endpoints.Annotation: inferenceEndpoints}
		newWlr := oldWlr.DeepCopy()
		newWlr.Annotations[endpoints.Annotation] = strings.Replace(inferenceEndpoints, "gie-v1.5", "llm-d-v0.10", 1)
		Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(BeTrue())
	})

	It("syncs a new runner generation to invalidate old routing readiness", func() {
		oldWlr := wlrWithPluginDetails(networkingPluginName, nil)
		oldWlr.Annotations = map[string]string{endpoints.Annotation: inferenceEndpoints}
		newWlr := oldWlr.DeepCopy()
		newWlr.Generation++
		Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(BeTrue())
	})

	It("does not sync a generation change for endpoints without inference", func() {
		oldWlr := wlrWithPluginDetails(networkingPluginName, nil)
		oldWlr.Annotations = map[string]string{endpoints.Annotation: endpointsWithoutInference}
		newWlr := oldWlr.DeepCopy()
		newWlr.Generation++
		Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(BeFalse())
	})

	DescribeTable("detects inference routing updates while the workload phase stays unchanged",
		func(change func(*dromosv1alpha1.PluginStatus), expected bool) {
			oldWlr := wlrWithPluginDetails(networkingPluginName, []byte(`{"schemaVersion":"v1","endpoints":[{"phase":"Initializing"}]}`))
			oldWlr.Status.Phase = dromosv1alpha1.WorkloadRunnerPhaseRunning
			oldWlr.Status.Plugins[networkingPluginName].Phase = dromosv1alpha1.PluginPhaseInitializing
			oldWlr.Status.Plugins[networkingPluginName].Conditions = []metav1.Condition{{
				Type: endpoints.InferenceEndpointReadyCondition, Status: metav1.ConditionFalse,
				Reason: "EndpointNotReady", Message: "waiting for route", ObservedGeneration: 1,
				LastTransitionTime: metav1.NewTime(time.Unix(1, 0)),
			}}
			newWlr := oldWlr.DeepCopy()
			change(newWlr.Status.Plugins[networkingPluginName])
			Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(Equal(expected))
		},
		Entry("readiness changes", func(p *dromosv1alpha1.PluginStatus) { p.Conditions[0].Status = metav1.ConditionTrue }, true),
		Entry("reason changes", func(p *dromosv1alpha1.PluginStatus) { p.Conditions[0].Reason = "ReconcileFailed" }, true),
		Entry("message changes", func(p *dromosv1alpha1.PluginStatus) { p.Conditions[0].Message = "Gateway rejected route" }, true),
		Entry("observed generation changes", func(p *dromosv1alpha1.PluginStatus) { p.Conditions[0].ObservedGeneration++ }, true),
		Entry("condition is removed", func(p *dromosv1alpha1.PluginStatus) { p.Conditions = nil }, true),
		Entry("networking details change", func(p *dromosv1alpha1.PluginStatus) {
			p.AdditionalInfo.Raw = []byte(`{"schemaVersion":"v1","endpoints":[{"phase":"Ready"}]}`)
		}, true),
		Entry("networking details are removed", func(p *dromosv1alpha1.PluginStatus) { p.AdditionalInfo = nil }, true),
		Entry("transition timestamp alone changes", func(p *dromosv1alpha1.PluginStatus) {
			p.Conditions[0].LastTransitionTime = metav1.NewTime(time.Unix(2, 0))
		}, false),
		Entry("unrelated condition changes", func(p *dromosv1alpha1.PluginStatus) {
			p.Conditions = append(p.Conditions, metav1.Condition{Type: "Other", Status: metav1.ConditionTrue})
		}, false),
		Entry("no status change", func(_ *dromosv1alpha1.PluginStatus) {}, false),
	)

	It("returns true when networking plugin details change", func() {
		oldWlr := wlrWithPluginDetails(networkingPluginName, []byte(`{"schemaVersion":"v1","endpoints":[]}`))
		newWlr := wlrWithPluginDetails(networkingPluginName, []byte(`{"schemaVersion":"v1","endpoints":[{"phase":"Ready"}]}`))

		Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(BeTrue())
	})

	It("returns false when non-networking plugin details change", func() {
		oldWlr := wlrWithPluginDetails("storage-plugin", []byte(`{"phase":"Pending"}`))
		newWlr := wlrWithPluginDetails("storage-plugin", []byte(`{"phase":"Ready"}`))

		Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(BeFalse())
	})

	It("returns true when the workload template changes", func() {
		oldWlr := wlrWithTemplate([]byte(`{"spec":{"replicas":1}}`))
		newWlr := wlrWithTemplate([]byte(`{"spec":{"replicas":2}}`))

		Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(BeTrue())
	})

	It("returns false when the workload template is unchanged", func() {
		oldWlr := wlrWithTemplate([]byte(`{"spec":{"replicas":1}}`))
		newWlr := wlrWithTemplate([]byte(`{"spec":{"replicas":1}}`))

		Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(BeFalse())
	})
})

func wlrWithPluginDetails(pluginName string, details []byte) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		Status: dromosv1alpha1.WorkloadRunnerStatus{
			Plugins: map[string]*dromosv1alpha1.PluginStatus{
				pluginName: {
					AdditionalInfo: &runtime.RawExtension{Raw: details},
				},
			},
		},
	}
}

func wlrWithTemplate(template []byte) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Template: runtime.RawExtension{Raw: template},
		},
	}
}
