package services

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/dao"
	testing "github.com/run-ai/runai/control-plane/authorization/pkg/test"
)

var _ = Describe("GetAccessRulesSync version-aware transformation tests", Ordered, func() {
	tenantId := fmt.Sprintf("%d", testing.TestTenant)
	testAdminsGroups := []string{dao.SuperAdmins}
	ctx := setContext(tenantId, "test-user@run.ai", testAdminsGroups)

	Describe("Helper function isServiceAccountSupportedInCluster", func() {
		It("should return false for empty cluster ID", func() {
			supported, err := isServiceAccountSupportedInCluster(ctx, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(supported).To(BeFalse())
		})

		It("should return error for non-existent cluster", func() {
			nonExistentClusterId := "99999999-9999-9999-9999-999999999999"
			supported, err := isServiceAccountSupportedInCluster(ctx, nonExistentClusterId)
			Expect(err).To(HaveOccurred())
			Expect(supported).To(BeFalse())
		})
	})

	Describe("remapSyncSubjectForClusterSupport", func() {
		It("remaps service-account to app when the cluster doesn't support it", func() {
			rule := remapSyncSubjectForClusterSupport(
				api.AccessRuleSync{SubjectId: "test-service-account", SubjectType: api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT},
				false, true,
			)
			Expect(rule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_APP))
		})

		It("keeps service-account when the cluster supports it", func() {
			rule := remapSyncSubjectForClusterSupport(
				api.AccessRuleSync{SubjectId: "test-service-account", SubjectType: api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT},
				true, true,
			)
			Expect(rule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT))
		})

		It("leaves app untouched regardless of cluster version", func() {
			rule := remapSyncSubjectForClusterSupport(
				api.AccessRuleSync{SubjectId: "test-app", SubjectType: api.SUBJECTTYPEEXTENDED_APP},
				false, false,
			)
			Expect(rule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_APP))
		})

		It("remaps all-authenticated-users to group system:authenticated when unsupported", func() {
			rule := remapSyncSubjectForClusterSupport(
				api.AccessRuleSync{SubjectId: "all-authenticated-users", SubjectType: api.SUBJECTTYPEEXTENDED_ALL_AUTHENTICATED_USERS},
				true, false,
			)
			Expect(rule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_GROUP))
			Expect(rule.SubjectId).To(Equal(authenticatedUsersK8sGroup))
		})

		It("keeps all-authenticated-users native when the cluster supports it", func() {
			rule := remapSyncSubjectForClusterSupport(
				api.AccessRuleSync{SubjectId: "all-authenticated-users", SubjectType: api.SUBJECTTYPEEXTENDED_ALL_AUTHENTICATED_USERS},
				true, true,
			)
			Expect(rule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_ALL_AUTHENTICATED_USERS))
			Expect(rule.SubjectId).To(Equal("all-authenticated-users"))
		})
	})

	Describe("shouldMapServiceAccountToApp function", func() {
		It("should return false when clusterId is empty (UI call)", func() {
			result := shouldMapServiceAccountToApp(ctx, "")
			Expect(result).To(BeFalse())
		})

		It("should return true for non-existent/old cluster", func() {
			result := shouldMapServiceAccountToApp(ctx, "non-existent-cluster")
			Expect(result).To(BeTrue())
		})
	})

	Describe("GetAccessRules API backwards compatibility", func() {
		It("should map service-account to app for old cluster (non-existent/unknown)", func() {
			rule := api.AccessRule{
				SubjectId:   "test-service-account-id",
				SubjectType: api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT,
			}

			if shouldMapServiceAccountToApp(ctx, "old-cluster-id") && rule.SubjectType == api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT {
				rule.SubjectType = api.SUBJECTTYPEEXTENDED_APP
			}

			Expect(rule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_APP))
		})

		It("should NOT map service-account for UI calls (empty clusterId)", func() {
			rule := api.AccessRule{
				SubjectId:   "test-service-account-id",
				SubjectType: api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT,
			}

			if shouldMapServiceAccountToApp(ctx, "") && rule.SubjectType == api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT {
				rule.SubjectType = api.SUBJECTTYPEEXTENDED_APP
			}

			Expect(rule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_SERVICE_ACCOUNT))
		})
	})

	BeforeAll(func() {
		TestsCleanupAndInit()
	})

	AfterAll(func() {
		TestsCleanupAndInit()
	})
})
