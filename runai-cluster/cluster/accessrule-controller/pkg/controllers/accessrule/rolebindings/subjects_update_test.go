/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rolebindings

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/rolebindingsutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestRolebindings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rolebindings Suite")
}

func allAuthRule() *v1alpha1.AccessRule {
	return &v1alpha1.AccessRule{
		Spec: v1alpha1.AccessRuleSpec{SubjectType: v1alpha1.AllAuthenticatedUsersSubject, SubjectId: "all-authenticated-users"},
	}
}

var _ = Describe("isAccessRuleInSubjects", func() {
	Context("all-authenticated-users rule", func() {
		rule := allAuthRule()

		It("matches on the resolved system:authenticated group subject", func() {
			present := []rbacv1.Subject{{APIGroup: rolebindingsutils.RbacGroup, Kind: rolebindingsutils.GroupKind, Name: "system:authenticated"}}
			Expect(isAccessRuleInSubjects(rule, present)).To(BeTrue())
		})

		It("does not falsely match an unrelated user from the raw subjectId", func() {
			absent := []rbacv1.Subject{{APIGroup: rolebindingsutils.RbacGroup, Kind: rolebindingsutils.UserKind, Name: "alice@run.ai"}}
			Expect(isAccessRuleInSubjects(rule, absent)).To(BeFalse())
		})
	})
})
