/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rolebindingsutils

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestRolebindingsUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RolebindingsUtils Suite")
}

func accessRule(subjectType v1alpha1.AccessRuleSubjectType, subjectId string) *v1alpha1.AccessRule {
	return &v1alpha1.AccessRule{Spec: v1alpha1.AccessRuleSpec{SubjectType: subjectType, SubjectId: subjectId}}
}

var _ = Describe("subjectNameForAccessRule", func() {
	It("resolves all-authenticated-users to the authenticated users group", func() {
		Expect(subjectNameForAccessRule(accessRule(v1alpha1.AllAuthenticatedUsersSubject, "all-authenticated-users"))).
			To(Equal(AuthenticatedUsersGroup))
	})

	It("uses the subjectId for a user subject", func() {
		Expect(subjectNameForAccessRule(accessRule(v1alpha1.UserSubject, "alice@run.ai"))).
			To(Equal("alice@run.ai"))
	})
})

var _ = Describe("GetSubjectForAccessRule", func() {
	It("builds the system:authenticated group subject for all-authenticated-users", func() {
		subject := GetSubjectForAccessRule(accessRule(v1alpha1.AllAuthenticatedUsersSubject, "all-authenticated-users"))
		Expect(subject).To(Equal(rbacv1.Subject{APIGroup: RbacGroup, Kind: GroupKind, Name: "system:authenticated"}))
	})
})

var _ = Describe("GetSubjectKind", func() {
	It("maps all-authenticated-users to the group kind", func() {
		Expect(GetSubjectKind(v1alpha1.AllAuthenticatedUsersSubject)).To(Equal(GroupKind))
	})

	It("maps group to the group kind", func() {
		Expect(GetSubjectKind(v1alpha1.GroupSubject)).To(Equal(GroupKind))
	})

	It("maps user to the user kind", func() {
		Expect(GetSubjectKind(v1alpha1.UserSubject)).To(Equal(UserKind))
	})

	It("maps service account to the user kind", func() {
		Expect(GetSubjectKind(v1alpha1.ServiceAccountSubject)).To(Equal(UserKind))
	})
})
