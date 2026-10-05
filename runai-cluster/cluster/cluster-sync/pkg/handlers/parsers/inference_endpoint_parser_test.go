/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromos "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("InferenceEndpoint routing condition reporting", func() {
	var runner *dromos.WorkloadRunner
	var fields openapi.WorkloadSyncFields

	BeforeEach(func() {
		runner = &dromos.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name: "runner", Namespace: "project", Generation: 2,
				Annotations: map[string]string{endpoints.InferenceEndpointAnnotation: `{"schemaVersion":"v1","spec":{
                    "workload":{"selector":{"matchLabels":{"app":"model"}},"port":8000},
                    "routing":{"gatewayRef":{"name":"gateway"},"hostnames":["model.example.com"]},
                    "epp":{"profile":"gie-v1.5"}
                }}`},
			},
			Status: dromos.WorkloadRunnerStatus{Plugins: map[string]*dromos.PluginStatus{
				endpoints.InferenceEndpointPluginName: {
					Phase: dromos.PluginPhaseReady,
					Conditions: []metav1.Condition{{
						Type: endpoints.InferenceEndpointReadyCondition, Status: metav1.ConditionTrue,
						Reason: "Ready", Message: "route accepted and endpoint picker available", ObservedGeneration: 2,
					}},
				},
			}},
		}
		fields = openapi.WorkloadSyncFields{
			Phase: "Running", Urls: []string{"https://existing.example.com"},
			Conditions: []openapi.Condition{{Type: "Ready", Status: "True"}},
		}
	})

	project := func() {
		scheme := runtime.NewScheme()
		Expect(dromos.AddToScheme(scheme)).To(Succeed())
		parser := NewWorkloadParser(fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build())
		parser.overrideFromWorkloadRunner(&v1alpha1.ExternalWorkload{ObjectMeta: metav1.ObjectMeta{
			Namespace: "project", Labels: map[string]string{commonconstants.WorkloadRunnerNameLabel: "runner"},
			Annotations: map[string]string{commonconstants.EwiStatusAnnotation: "true"},
		}}, &fields)
	}

	It("reports routing readiness separately from workload health and existing URLs", func() {
		project()
		Expect(fields.Conditions).To(HaveLen(2))
		Expect(fields.Conditions[1].Type).To(Equal(endpoints.InferenceEndpointReadyCondition))
		Expect(fields.Conditions[1].Status).To(Equal("True"))
		Expect(fields.Conditions[1].GetReason()).To(Equal("Ready"))
		Expect(fields.Conditions[1].GetMessage()).To(Equal("route accepted and endpoint picker available"))
		Expect(string(fields.Phase)).To(Equal("Running"))
		Expect(fields.Urls).To(Equal([]string{"https://existing.example.com"}))
		Expect(fields.Conditions[0]).To(Equal(openapi.Condition{Type: "Ready", Status: "True"}))
	})

	It("reports a rejected route without changing the workload phase", func() {
		plugin := runner.Status.Plugins[endpoints.InferenceEndpointPluginName]
		plugin.Phase = dromos.PluginPhasePending
		plugin.Conditions[0].Status = metav1.ConditionFalse
		plugin.Conditions[0].Reason = "EndpointNotReady"
		plugin.Conditions[0].Message = "Gateway rejected route"
		project()
		Expect(fields.Conditions).To(HaveLen(2))
		Expect(fields.Conditions[1].Status).To(Equal("False"))
		Expect(fields.Conditions[1].GetMessage()).To(Equal("Gateway rejected route"))
		Expect(string(fields.Phase)).To(Equal("Running"))
	})

	It("reports unknown routing state while the runner generation is stale", func() {
		runner.Status.Plugins[endpoints.InferenceEndpointPluginName].Conditions[0].ObservedGeneration = 1
		project()
		Expect(fields.Conditions).To(HaveLen(2))
		Expect(fields.Conditions[1].Status).To(Equal("Unknown"))
		Expect(fields.Conditions[1].GetReason()).To(Equal("Pending"))
	})

	It("reports unknown routing state before plugin status appears", func() {
		runner.Status.Plugins = nil
		project()
		Expect(fields.Conditions).To(HaveLen(2))
		Expect(fields.Conditions[1].Status).To(Equal("Unknown"))
	})

	It("leaves workloads without intent unchanged even if an old plugin status remains", func() {
		runner.Annotations = nil
		project()
		Expect(fields.Conditions).To(Equal([]openapi.Condition{{Type: "Ready", Status: "True"}}))
	})
})
