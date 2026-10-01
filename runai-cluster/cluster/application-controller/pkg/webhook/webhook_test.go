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
				}})
				Expect(response.Allowed).To(Equal(allowed), string(operation))
			}
		},
		Entry("application controller", "system:serviceaccount:runai:application-controller", true),
		Entry("agent", "system:serviceaccount:runai:runai-agent", true),
		Entry("tenant service account", "system:serviceaccount:tenant:default", false),
		Entry("same name in tenant namespace", "system:serviceaccount:tenant:application-controller", false),
		Entry("controller name suffix", "system:serviceaccount:runai:application-controller-extra", false),
	)
})
