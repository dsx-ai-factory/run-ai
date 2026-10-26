/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"encoding/json"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	constants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("overrideFromWorkloadRunner ewi-status guard", func() {
	const ns = "team-a"
	const wlrName = "wlr-1"

	buildParser := func() *WorkloadParser {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
		wlr := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{Name: wlrName, Namespace: ns},
			Status:     dromosv1alpha1.WorkloadRunnerStatus{Phase: dromosv1alpha1.WorkloadRunnerPhaseFailed},
		}
		c := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(wlr).Build()
		return NewWorkloadParser(c)
	}

	newExternalWorkload := func(annotations map[string]string) *v1alpha1.ExternalWorkload {
		return &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   ns,
				Labels:      map[string]string{commonconstants.WorkloadRunnerNameLabel: wlrName},
				Annotations: annotations,
			},
		}
	}

	It("keeps the EWI phase + message when the ewi-status annotation is set", func() {
		msg := "Workload is initializing, the pod is still not ready"
		fields := openapi.WorkloadSyncFields{Phase: "Initializing", PhaseMessage: &msg}

		buildParser().overrideFromWorkloadRunner(
			newExternalWorkload(map[string]string{commonconstants.EwiStatusAnnotation: "true"}), &fields)

		Expect(string(fields.Phase)).To(Equal("Initializing"))
		Expect(*fields.PhaseMessage).To(Equal(msg))
	})

	It("overrides the phase with the WorkloadRunner phase when ewi-status is not set", func() {
		fields := openapi.WorkloadSyncFields{Phase: "Initializing"}

		buildParser().overrideFromWorkloadRunner(newExternalWorkload(nil), &fields)

		Expect(string(fields.Phase)).To(Equal("Failed"))
	})

	It("surfaces the admission rejection reason from a failed WorkloadRunner", func() {
		const rejectionReason = "admission webhook denied the request: policy disallow-host-path-runai-projects, rule no-host-path failed at /spec/volumes/0/hostPath"

		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
		wlr := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{Name: wlrName, Namespace: ns},
			Status: dromosv1alpha1.WorkloadRunnerStatus{
				Phase: dromosv1alpha1.WorkloadRunnerPhaseFailed,
				Issues: []dromosv1alpha1.Issue{{
					Group:       "Internal",
					Description: rejectionReason,
					IsCritical:  true,
				}},
			},
		}
		parser := NewWorkloadParser(fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(wlr).Build())
		creatingMessage := "Workload is being created"
		fields := openapi.WorkloadSyncFields{Phase: "Creating", PhaseMessage: &creatingMessage}

		parser.overrideFromWorkloadRunner(newExternalWorkload(nil), &fields)

		Expect(string(fields.Phase)).To(Equal("Failed"))
		Expect(fields.PhaseMessage).NotTo(BeNil())
		Expect(*fields.PhaseMessage).To(Equal(rejectionReason))
	})

	It("still maps WorkloadIssue conditions when the WorkloadRunner fetch fails", func() {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
		emptyReader := fakeclient.NewClientBuilder().WithScheme(scheme).Build() // no WLR -> Get fails
		parser := NewWorkloadParser(emptyReader)

		ew := newExternalWorkload(nil)
		ew.Status.ObservedStatus.Conditions = []v1alpha1.ObservedCondition{{
			Type:    commonconstants.WorkloadIssueConditionType,
			Status:  corev1.ConditionTrue,
			Reason:  "Compute",
			Message: "container 'main' was OOMKilled",
		}}
		fields := openapi.WorkloadSyncFields{}

		parser.overrideFromWorkloadRunner(ew, &fields)

		Expect(fields.AdditionalFields).NotTo(BeNil())
		Expect(*fields.AdditionalFields).To(ContainSubstring("container 'main' was OOMKilled"))
	})

	DescribeTable("projects user images from profiled WorkloadRunner templates",
		func(template map[string]any, expected []string) {
			raw, err := json.Marshal(template)
			Expect(err).NotTo(HaveOccurred())
			scheme := runtime.NewScheme()
			Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
			Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
			wlr := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{Name: wlrName, Namespace: ns},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: runtime.RawExtension{Raw: raw},
				},
			}
			parser := NewWorkloadParser(fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(wlr).Build())
			fields := openapi.WorkloadSyncFields{}

			parser.overrideFromWorkloadRunner(newExternalWorkload(nil), &fields)

			Expect(fields.Images).To(Equal(expected))
		},
		Entry("Knative Service", map[string]any{
			"apiVersion": "serving.knative.dev/v1",
			"kind":       "Service",
			"spec": map[string]any{"template": map[string]any{"spec": map[string]any{
				"containers": []any{map[string]any{"name": "user", "image": "model:v1"}},
			}}},
		}, []string{"model:v1"}),
		Entry("LeaderWorkerSet with distinct and duplicate images", map[string]any{
			"apiVersion": "leaderworkerset.x-k8s.io/v1",
			"kind":       "LeaderWorkerSet",
			"spec": map[string]any{"leaderWorkerTemplate": map[string]any{
				"leaderTemplate": map[string]any{"spec": map[string]any{"containers": []any{
					map[string]any{"name": "leader", "image": "leader:v1"},
				}}},
				"workerTemplate": map[string]any{"spec": map[string]any{"containers": []any{
					map[string]any{"name": "worker", "image": "worker:v1"},
					map[string]any{"name": "same-as-leader", "image": "leader:v1"},
				}}},
			}},
		}, []string{"leader:v1", "worker:v1"}),
	)

	DescribeTable("projects inference engine labels from WorkloadRunner templates",
		func(labels map[string]any, expected string) {
			raw, err := json.Marshal(map[string]any{
				"apiVersion": "serving.knative.dev/v1",
				"kind":       "Service",
				"metadata":   map[string]any{"labels": labels},
				"spec":       map[string]any{"template": map[string]any{"spec": map[string]any{}}},
			})
			Expect(err).NotTo(HaveOccurred())
			scheme := runtime.NewScheme()
			Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
			Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
			wlr := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{Name: wlrName, Namespace: ns},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: runtime.RawExtension{Raw: raw},
				},
			}
			parser := NewWorkloadParser(fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(wlr).Build())
			fields := openapi.WorkloadSyncFields{}

			parser.overrideFromWorkloadRunner(newExternalWorkload(nil), &fields)

			Expect(fields.AdditionalFields).NotTo(BeNil())
			Expect(*fields.AdditionalFields).To(MatchJSON(expected))
		},
		Entry("vLLM", map[string]any{"run.ai/vllm-workload": "true"}, `{"vllmWorkload":"true"}`),
		Entry("NIM", map[string]any{"run.ai/nim-workload": "true"}, `{"nimWorkload":"true"}`),
	)

	It("projects vllmWorkload from the profiled template without a user-supplied engine label", func() {
		raw, err := json.Marshal(map[string]any{
			"apiVersion": "serving.knative.dev/v1",
			"kind":       "Service",
			"metadata": map[string]any{"labels": map[string]any{
				constants.LabelVLLMWorkload: "true",
			}},
			"spec": map[string]any{"template": map[string]any{"spec": map[string]any{}}},
		})
		Expect(err).NotTo(HaveOccurred())
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
		wlr := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{Name: wlrName, Namespace: ns},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Template: runtime.RawExtension{Raw: raw},
			},
		}
		parser := NewWorkloadParser(fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(wlr).Build())
		externalWorkload := newExternalWorkload(nil)
		Expect(externalWorkload.Labels).NotTo(HaveKey(constants.LabelVLLMWorkload))
		fields := openapi.WorkloadSyncFields{}

		parser.overrideFromWorkloadRunner(externalWorkload, &fields)

		Expect(fields.AdditionalFields).NotTo(BeNil())
		Expect(*fields.AdditionalFields).To(MatchJSON(`{"vllmWorkload":"true"}`))
	})

	It("projects the total requested pod count from LWS replicas and group size", func() {
		raw, err := json.Marshal(map[string]any{
			"apiVersion": "leaderworkerset.x-k8s.io/v1",
			"kind":       "LeaderWorkerSet",
			"spec": map[string]any{
				"replicas": int64(2),
				"leaderWorkerTemplate": map[string]any{
					"size": int64(3),
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
		wlr := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{Name: wlrName, Namespace: ns},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Template: runtime.RawExtension{Raw: raw},
			},
		}
		parser := NewWorkloadParser(fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(wlr).Build())
		fields := openapi.WorkloadSyncFields{}

		parser.overrideFromWorkloadRunner(newExternalWorkload(nil), &fields)

		Expect(fields.RequestedPods).NotTo(BeNil())
		Expect(fields.RequestedPods.HasNumber()).To(BeTrue())
		Expect(fields.RequestedPods.GetNumber()).To(Equal(int32(6)))
	})

	It("projects native DistributedInference additional fields from the LWS template", func() {
		raw, err := json.Marshal(map[string]any{
			"apiVersion": "leaderworkerset.x-k8s.io/v1",
			"kind":       "LeaderWorkerSet",
			"spec": map[string]any{
				"replicas":      int64(2),
				"startupPolicy": "LeaderReady",
				"leaderWorkerTemplate": map[string]any{
					"size":          int64(4),
					"restartPolicy": "None",
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
		wlr := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{Name: wlrName, Namespace: ns},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Template: runtime.RawExtension{Raw: raw},
			},
		}
		parser := NewWorkloadParser(fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(wlr).Build())
		externalWorkload := newExternalWorkload(nil)
		externalWorkload.Spec.Type = distributedInferenceWorkloadType
		existing := `{"statusReplicas":2,"readyReplicas":1,"keep":"value"}`
		fields := openapi.WorkloadSyncFields{AdditionalFields: &existing}

		parser.overrideFromWorkloadRunner(externalWorkload, &fields)

		Expect(fields.AdditionalFields).NotTo(BeNil())
		Expect(*fields.AdditionalFields).To(MatchJSON(`{
			"specReplicas":2,
			"workers":3,
			"restartPolicy":"None",
			"startupPolicy":"LeaderReady",
			"statusReplicas":2,
			"readyReplicas":1,
			"keep":"value"
		}`))
	})

	It("preserves valid zero values and removes stale projected fields", func() {
		raw, err := json.Marshal(map[string]any{
			"apiVersion": "leaderworkerset.x-k8s.io/v1",
			"kind":       "LeaderWorkerSet",
			"spec": map[string]any{
				"replicas": int64(0),
				"leaderWorkerTemplate": map[string]any{
					"size": int64(1),
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		wlr := &dromosv1alpha1.WorkloadRunner{
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Template: runtime.RawExtension{Raw: raw},
			},
		}
		stale := `{"statusReplicas":1,"specReplicas":2,"workers":3,"restartPolicy":"None","startupPolicy":"LeaderReady"}`
		fields := openapi.WorkloadSyncFields{AdditionalFields: &stale}

		projectDistributedInferenceCompatibilityFields(wlr, &fields)

		Expect(fields.AdditionalFields).NotTo(BeNil())
		Expect(*fields.AdditionalFields).To(MatchJSON(`{
			"statusReplicas":1,
			"specReplicas":0,
			"workers":0
		}`))
	})

	DescribeTable("validates LWS requested pod count bounds",
		func(replicas, groupSize int64, projected bool, expected int32) {
			template := &unstructured.Unstructured{Object: map[string]any{
				"spec": map[string]any{
					"replicas": replicas,
					"leaderWorkerTemplate": map[string]any{
						"size": groupSize,
					},
				},
			}}
			fields := openapi.WorkloadSyncFields{}

			projectLWSRequestedPods(template, &fields)

			if !projected {
				Expect(fields.RequestedPods).To(BeNil())
				return
			}
			Expect(fields.RequestedPods).NotTo(BeNil())
			Expect(fields.RequestedPods.GetNumber()).To(Equal(expected))
		},
		Entry("accepts the int32 maximum", int64(1<<31-1), int64(1), true, int32(1<<31-1)),
		Entry("rejects a total above the int32 maximum", int64(1<<31), int64(1), false, int32(0)),
		Entry("rejects int64 multiplication overflow", int64(1<<63-1), int64(2), false, int32(0)),
	)

	DescribeTable("gates distributed inference URLs using EWI ready replicas",
		func(additionalFields string, expectURLs bool) {
			networkingStatus := endpoints.NewNetworkingStatus(1)
			networkingStatus.Endpoints = []endpoints.EndpointStatus{{
				Services: []endpoints.ServiceStatus{{
					Name: "serving", URLs: []endpoints.ServiceURLStatus{{URL: "https://model.example.com"}},
				}},
			}}
			networkingRaw, err := endpoints.NetworkingStatusToRawExtension(networkingStatus)
			Expect(err).NotTo(HaveOccurred())
			scheme := runtime.NewScheme()
			Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
			Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
			wlr := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{Name: wlrName, Namespace: ns},
				Status: dromosv1alpha1.WorkloadRunnerStatus{Plugins: map[string]*dromosv1alpha1.PluginStatus{
					networkingPluginName: {AdditionalInfo: networkingRaw},
				}},
			}
			parser := NewWorkloadParser(fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(wlr).Build())
			ew := newExternalWorkload(map[string]string{commonconstants.EwiStatusAnnotation: "true"})
			ew.Spec.Type = "DistributedInference"
			fields := openapi.WorkloadSyncFields{AdditionalFields: &additionalFields}

			parser.overrideFromWorkloadRunner(ew, &fields)

			if expectURLs {
				Expect(fields.Urls).To(Equal([]string{"https://model.example.com"}))
			} else {
				Expect(fields.Urls).To(BeEmpty())
			}
		},
		Entry("zero ready replicas suppresses URLs", `{"readyReplicas":0}`, false),
		Entry("one ready replica keeps networking URLs", `{"readyReplicas":1}`, true),
		Entry("malformed replica status suppresses URLs", `{`, false),
	)
})
