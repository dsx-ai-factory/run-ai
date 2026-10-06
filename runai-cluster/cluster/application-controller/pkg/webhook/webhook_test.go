/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestApplicationAuthorization(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Application Authorization Suite")
}

var _ = Describe("Application authorization", func() {
	DescribeTable("requires an exact trusted identity",
		func(username string, allowed bool) {
			validator := NewServiceAccountValidator(nil, []string{
				"system:serviceaccount:runai:application-controller",
				"system:serviceaccount:runai:runai-agent",
			})
			for _, operation := range []admissionv1.Operation{admissionv1.Create, admissionv1.Update, admissionv1.Delete} {
				response := validator.Handle(context.Background(), admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
					Operation: operation, UserInfo: authenticationv1.UserInfo{Username: username},
					Resource: metav1.GroupVersionResource{Group: "run.ai", Version: "v1alpha1", Resource: "applications"},
				}})
				Expect(response.Allowed).To(Equal(allowed), string(operation))
			}
		},
		Entry("application controller", "system:serviceaccount:runai:application-controller", true),
		Entry("agent", "system:serviceaccount:runai:runai-agent", true),
		Entry("tenant service account", "system:serviceaccount:tenant:default", false),
		Entry("same name in tenant namespace", "system:serviceaccount:tenant:application-controller", false),
		Entry("controller name suffix", "system:serviceaccount:runai:application-controller-extra", false),
		Entry("tenant garbage collector", "system:serviceaccount:tenant:generic-garbage-collector", false),
		Entry("tenant namespace controller", "system:serviceaccount:tenant:namespace-controller", false),
		Entry("garbage collector suffix", "system:serviceaccount:kube-system:generic-garbage-collector-extra", false),
		Entry("namespace controller suffix", "system:serviceaccount:kube-system:namespace-controller-extra", false),
		Entry("unrelated system account", "system:serviceaccount:kube-system:coredns", false),
		Entry("empty identity", "", false),
	)
})

var _ = DescribeTableSubtree("Kubernetes Application cleanup", func(username string) {
	var request admission.Request
	var handler *ServiceAccountValidator

	BeforeEach(func() {
		handler = NewServiceAccountValidator(nil, []string{
			"system:serviceaccount:custom-runai:application-controller",
			"system:serviceaccount:custom-runai:nmc-agent",
		})
		request = admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Delete,
			UserInfo:  authenticationv1.UserInfo{Username: username},
			Kind:      metav1.GroupVersionKind{Group: "run.ai", Version: "v1alpha1", Kind: "Application"},
			Resource:  metav1.GroupVersionResource{Group: "run.ai", Version: "v1alpha1", Resource: "applications"},
			Name:      "repro-am", Namespace: "runai-repro-app",
		}}
	})

	It("allows deleting an Application", func() {
		Expect(handler.Handle(context.Background(), request).Allowed).To(BeTrue())
	})

	DescribeTable("rejects operations outside Application cleanup", func(change func(*admission.Request)) {
		change(&request)
		Expect(handler.Handle(context.Background(), request).Allowed).To(BeFalse())
	},
		Entry("create", func(req *admission.Request) { req.Operation = admissionv1.Create }),
		Entry("update", func(req *admission.Request) { req.Operation = admissionv1.Update }),
		Entry("connect", func(req *admission.Request) { req.Operation = admissionv1.Connect }),
		Entry("another API group", func(req *admission.Request) { req.Resource.Group = "other.example.com" }),
		Entry("another resource", func(req *admission.Request) { req.Resource.Resource = "projects" }),
		Entry("a subresource", func(req *admission.Request) { req.SubResource = "status" }),
	)
},
	Entry("garbage collector", "system:serviceaccount:kube-system:generic-garbage-collector"),
	Entry("namespace controller", "system:serviceaccount:kube-system:namespace-controller"),
)
