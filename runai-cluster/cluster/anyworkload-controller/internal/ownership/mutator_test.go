/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ownership_test

import (
	"context"
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/run-ai/runai/runai-cluster/cluster/anyworkload-controller/internal/ownership"
	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

var _ = Describe("Legacy ownership update mutation", func() {
	DescribeTable("restores the stored owner for every legacy workload kind",
		func(gvk schema.GroupVersionKind) {
			oldRaw := legacyWorkloadJSON(gvk, "alice")
			newRaw := legacyWorkloadJSON(gvk, "bob")
			request := admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Update,
				Kind: metav1.GroupVersionKind{
					Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind,
				},
				OldObject: runtime.RawExtension{Raw: oldRaw},
				Object:    runtime.RawExtension{Raw: newRaw},
			}}

			response := ownership.NewUpdateMutator().Handle(context.Background(), request)

			Expect(response.Allowed).To(BeTrue())
			Expect(response.Patches).To(ContainElement(SatisfyAll(
				HaveField("Operation", Equal("replace")),
				HaveField("Path", Equal("/spec/userId/value")),
				HaveField("Value", Equal("alice")),
			)))
		},
		Entry("TrainingWorkload", schema.GroupVersionKind{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkload"}),
		Entry("InteractiveWorkload", schema.GroupVersionKind{Group: "run.ai", Version: "v2alpha1", Kind: "InteractiveWorkload"}),
		Entry("DistributedWorkload", schema.GroupVersionKind{Group: "run.ai", Version: "v2alpha1", Kind: "DistributedWorkload"}),
		Entry("InferenceWorkload", schema.GroupVersionKind{Group: "run.ai", Version: "v2alpha1", Kind: "InferenceWorkload"}),
		Entry("DistributedInferenceWorkload", schema.GroupVersionKind{Group: "run.ai", Version: "v1alpha1", Kind: "DistributedInferenceWorkload"}),
	)

	It("restores an owner that was cleared", func() {
		gvk := schema.GroupVersionKind{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkload"}
		request := admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Update,
			OldObject: runtime.RawExtension{Raw: legacyWorkloadJSON(gvk, "alice")},
			Object: runtime.RawExtension{Raw: []byte(`{
				"apiVersion":"run.ai/v2alpha1","kind":"TrainingWorkload",
				"metadata":{"name":"workload","namespace":"runai-project"},
				"spec":{"usage":"Submit"}
			}`)},
		}}

		response := ownership.NewUpdateMutator().Handle(context.Background(), request)

		Expect(response.Allowed).To(BeTrue())
		Expect(response.Patches).To(ContainElement(SatisfyAll(
			HaveField("Operation", Equal("add")),
			HaveField("Path", Equal("/spec/userId")),
		)))
	})

	It("does not support create-time ownership mutation", func() {
		gvk := schema.GroupVersionKind{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkload"}
		request := admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Create,
			Object:    runtime.RawExtension{Raw: legacyWorkloadJSON(gvk, "bob")},
		}}

		response := ownership.NewUpdateMutator().Handle(context.Background(), request)

		Expect(response.Allowed).To(BeFalse())
		Expect(response.Result.Message).To(ContainSubstring("unsupported operation"))
	})

	It("removes a newly added owner when the stored workload had no userId", func() {
		gvk := schema.GroupVersionKind{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkload"}
		request := admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Update,
			OldObject: runtime.RawExtension{Raw: []byte(`{
				"apiVersion":"run.ai/v2alpha1","kind":"TrainingWorkload",
				"metadata":{"name":"workload","namespace":"runai-project"},
				"spec":{"usage":"Submit"}
			}`)},
			Object: runtime.RawExtension{Raw: legacyWorkloadJSON(gvk, "bob")},
		}}

		response := ownership.NewUpdateMutator().Handle(context.Background(), request)

		Expect(response.Allowed).To(BeTrue())
		Expect(response.Patches).To(ContainElement(SatisfyAll(
			HaveField("Operation", Equal("remove")),
			HaveField("Path", Equal("/spec/userId")),
		)))
	})
})

var _ = Describe("WorkloadRunner ownership update mutation", func() {
	It("restores the stored owner when an update changes it", func() {
		request := workloadRunnerValidationRequest(admissionv1.Update, "alice", "bob", "alice")

		response := ownership.NewUpdateMutator().Handle(context.Background(), request)

		Expect(response.Allowed).To(BeTrue())
		Expect(response.Patches).To(ContainElement(SatisfyAll(
			HaveField("Operation", Equal("replace")),
			HaveField("Path", Equal("/metadata/annotations/"+escapeJSONPointer(clusterconstants.WorkloadOwnerUIDAnnotation))),
			HaveField("Value", Equal("alice")),
		)))
	})

	It("restores the stored owner when an update clears it", func() {
		request := workloadRunnerValidationRequest(admissionv1.Update, "alice", "", "alice")

		response := ownership.NewUpdateMutator().Handle(context.Background(), request)

		Expect(response.Allowed).To(BeTrue())
		Expect(response.Patches).To(ContainElement(SatisfyAll(
			HaveField("Operation", Equal("add")),
			HaveField("Path", Equal("/metadata/annotations/"+escapeJSONPointer(clusterconstants.WorkloadOwnerUIDAnnotation))),
			HaveField("Value", Equal("alice")),
		)))
	})

	It("leaves an agent ownership backfill for validation", func() {
		request := workloadRunnerValidationRequest(admissionv1.Update, "", "alice", "agent-uid")

		response := ownership.NewUpdateMutator().Handle(context.Background(), request)

		Expect(response.Allowed).To(BeTrue())
		Expect(response.Patches).To(BeEmpty())
	})
})

func escapeJSONPointer(value string) string {
	value = strings.ReplaceAll(value, "~", "~0")
	return strings.ReplaceAll(value, "/", "~1")
}

func legacyWorkloadJSON(gvk schema.GroupVersionKind, owner string) []byte {
	object := map[string]any{
		"apiVersion": gvk.GroupVersion().String(),
		"kind":       gvk.Kind,
		"metadata": map[string]any{
			"name":      "workload",
			"namespace": "runai-project",
		},
		"spec": map[string]any{
			"usage":  "Submit",
			"userId": map[string]any{"value": owner},
		},
	}
	raw, err := json.Marshal(object)
	Expect(err).NotTo(HaveOccurred())
	return raw
}
