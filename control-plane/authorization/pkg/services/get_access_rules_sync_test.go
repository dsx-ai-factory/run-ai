/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
)

var _ = Describe("GetAccessRulesSync version-aware transformation tests", Ordered, func() {
	Describe("remapSyncSubjectForClusterSupport", func() {
		It("keeps service-account untouched regardless of cluster version", func() {
			for _, supportsAllAuthenticatedUsers := range []bool{true, false} {
				rule := remapSyncSubjectForClusterSupport(
					api.AccessRuleSync{SubjectId: "test-service-account", SubjectType: api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT},
					supportsAllAuthenticatedUsers,
				)
				Expect(rule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT))
				Expect(rule.SubjectId).To(Equal("test-service-account"))
			}
		})

		It("remaps all-authenticated-users to group system:authenticated when unsupported", func() {
			rule := remapSyncSubjectForClusterSupport(
				api.AccessRuleSync{SubjectId: "all-authenticated-users", SubjectType: api.SUBJECTTYPEEXTENDED_ALL_AUTHENTICATED_USERS},
				false,
			)
			Expect(rule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_GROUP))
			Expect(rule.SubjectId).To(Equal(authenticatedUsersK8sGroup))
		})

		It("keeps all-authenticated-users native when the cluster supports it", func() {
			rule := remapSyncSubjectForClusterSupport(
				api.AccessRuleSync{SubjectId: "all-authenticated-users", SubjectType: api.SUBJECTTYPEEXTENDED_ALL_AUTHENTICATED_USERS},
				true,
			)
			Expect(rule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_ALL_AUTHENTICATED_USERS))
			Expect(rule.SubjectId).To(Equal("all-authenticated-users"))
		})
	})

	BeforeAll(func() {
		TestsCleanupAndInit()
	})

	AfterAll(func() {
		TestsCleanupAndInit()
	})
})
