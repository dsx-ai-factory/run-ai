/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package restartrules

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/restart"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestRestartRules(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Restart Rules Suite")
}

var _ = Describe("ManagedBy authorization", func() {
	DescribeTable("requires an exact trusted identity",
		func(username string, allowed bool) {
			plugin := NewPlugin([]string{"system:serviceaccount:runai:anyworkload-controller"})
			ctx := admission.NewContextWithRequest(context.Background(), admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
				UserInfo: authenticationv1.UserInfo{Username: username},
			}})
			for _, operation := range []string{"create", "handoff", "managed update", "handback"} {
				var existing *dromosv1alpha1.WorkloadRunner
				desired := &dromosv1alpha1.WorkloadRunner{
					Spec: dromosv1alpha1.WorkloadRunnerSpec{ManagedBy: restart.RestartControllerManagedBy},
				}
				desired.Labels = map[string]string{restart.RequestRefLabel: "restart-request"}
				switch operation {
				case "handoff":
					existing = &dromosv1alpha1.WorkloadRunner{}
				case "managed update":
					existing = desired.DeepCopy()
				case "handback":
					existing = desired.DeepCopy()
					desired.Spec.ManagedBy = ""
					desired.Labels = nil
				}
				_, errs := plugin.WebhookValidate(ctx, existing, desired)
				if allowed {
					Expect(errs).To(BeEmpty(), operation)
				} else {
					Expect(errs).To(HaveLen(1), operation)
					Expect(errs[0].Type).To(Equal(field.ErrorTypeForbidden), operation)
					Expect(errs[0].Field).To(Equal("spec.managedBy"), operation)
				}
			}
		},
		Entry("restart controller", "system:serviceaccount:runai:anyworkload-controller", true),
		Entry("tenant service account", "system:serviceaccount:tenant:default", false),
		Entry("same name in tenant namespace", "system:serviceaccount:tenant:anyworkload-controller", false),
		Entry("controller name suffix", "system:serviceaccount:runai:anyworkload-controller-extra", false),
	)
})
