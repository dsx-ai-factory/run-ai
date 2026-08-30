package legacyprojectpermissions

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

var _ = Describe("constructPermissionsObjectFromAccessRules", func() {
	Context("with an all-authenticated-users rule alongside a user rule", func() {
		It("maps the sentinel to the system:authenticated group without leaking it into users", func() {
			rules := []v1alpha1.AccessRule{
				{Spec: v1alpha1.AccessRuleSpec{SubjectType: v1alpha1.AllAuthenticatedUsersSubject, SubjectId: "all-authenticated-users"}},
				{Spec: v1alpha1.AccessRuleSpec{SubjectType: v1alpha1.UserSubject, SubjectId: "alice@run.ai"}},
			}

			perms := constructPermissionsObjectFromAccessRules(rules)

			Expect(perms.Groups).To(ConsistOf("system:authenticated"))
			Expect(perms.Users).To(ContainElement("alice@run.ai"))
			Expect(perms.Users).NotTo(ContainElement("all-authenticated-users"))
		})
	})
})
