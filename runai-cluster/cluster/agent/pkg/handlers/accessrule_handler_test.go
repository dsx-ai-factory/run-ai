/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"time"

	_ "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/tests/helper"

	auth_openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// newAccessRulesHandlerWithObjects builds an AccessRulesHandler backed by a fake client already
// holding the given AccessRule objects.
func newAccessRulesHandlerWithObjects(objs ...client.Object) *AccessRulesHandler {
	scheme := runtime.NewScheme()
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	Expect(corev1.AddToScheme(scheme)).To(Succeed())

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()

	gvk, err := getGvkForObject(&v1alpha1.AccessRule{}, scheme)
	Expect(err).ToNot(HaveOccurred())

	return &AccessRulesHandler{
		CommonHandler: &CommonHandler{
			runaiClient:       fakeClient,
			cachedClient:      fakeClient,
			gvk:               gvk,
			objectInspector:   AccessRuleObjectInspector(true),
			deletionPredicate: func(object client.Object) bool { return true },
		},
	}
}

func accessRuleObject(name, accessRuleId string) *v1alpha1.AccessRule {
	return &v1alpha1.AccessRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{v1alpha1.AccessRuleIdLabel: accessRuleId},
		},
	}
}

// departmentAdminRule models the customer's access rule 11 from RUN-44340: a user holding the
// predefined "Department administrator" role, scoped to a department.
func departmentAdminRule(roleName string) auth_openapi.AccessRuleSync {
	return auth_openapi.AccessRuleSync{
		Id:          11,
		ScopeId:     "4500002",
		ScopeType:   auth_openapi.SCOPETYPE_DEPARTMENT,
		ScopeName:   "dept",
		SubjectId:   "someuser",
		SubjectType: auth_openapi.SUBJECTTYPEEXTENDED_USER,
		RoleId:      5,
		RoleName:    roleName,
		UpdatedAt:   time.Now(),
		DeletedAt:   *auth_openapi.NewNullableTime(nil),
	}
}

func accessRuleExists(h *AccessRulesHandler, name string) bool {
	err := h.cachedClient.Get(context.Background(),
		types.NamespacedName{Name: name}, &v1alpha1.AccessRule{})
	return err == nil
}

func allAccessRules(h *AccessRulesHandler) []v1alpha1.AccessRule {
	list := &v1alpha1.AccessRuleList{}
	Expect(h.cachedClient.List(context.Background(), list)).To(Succeed())
	return list.Items
}

var _ = Describe("AccessRulesHandler", func() {
	Context("a role rename changes the generated object name", func() {
		It("generates a different name for the same access rule id", func() {
			// This is the premise of RUN-44340: the generated name is not a stable identity.
			before := departmentAdminRule("department-administrator")
			after := departmentAdminRule("department-administrator-(legacy)")

			Expect(GenerateAccessRuleName(&before)).ToNot(Equal(GenerateAccessRuleName(&after)))
			Expect(accessRuleIdLabelValue(&before)).To(Equal(accessRuleIdLabelValue(&after)))
		})
	})

	Context("extractAccessRules", func() {
		It("collects the access rule id for deleted rules, not the generated name", func() {
			h := newAccessRulesHandlerWithObjects()

			deletedAt := time.Now()
			deletedRule := departmentAdminRule("department-administrator-(legacy)")
			deletedRule.DeletedAt = *auth_openapi.NewNullableTime(&deletedAt)

			result, err := h.extractAccessRules([]auth_openapi.AccessRuleSync{deletedRule})

			Expect(err).ToNot(HaveOccurred())
			Expect(result.accessRuleIdsToDelete).To(Equal([]string{"11"}))
			Expect(result.clusterAccessRules).To(BeEmpty())
		})

		It("collects rules that are deleting, not only ones already deleted", func() {
			h := newAccessRulesHandlerWithObjects()

			deletingRule := departmentAdminRule("department-administrator")
			deletingRule.IsDeleting = ptrTo(true)

			result, err := h.extractAccessRules([]auth_openapi.AccessRuleSync{deletingRule})

			Expect(err).ToNot(HaveOccurred())
			Expect(result.accessRuleIdsToDelete).To(Equal([]string{"11"}))
		})

		It("still converts live rules into cluster objects", func() {
			h := newAccessRulesHandlerWithObjects()

			liveRule := departmentAdminRule("department-administrator")

			result, err := h.extractAccessRules([]auth_openapi.AccessRuleSync{liveRule})

			Expect(err).ToNot(HaveOccurred())
			Expect(result.accessRuleIdsToDelete).To(BeEmpty())
			Expect(result.clusterAccessRules).To(HaveLen(1))
			Expect(result.clusterAccessRules[0].GetLabels()).To(
				HaveKeyWithValue(v1alpha1.AccessRuleIdLabel, "11"))
		})
	})

	Context("deleteAccessRulesByIds", func() {
		It("deletes the object left behind by a role rename", func() {
			// The object on the cluster was created before the role was renamed, so its name is the
			// one generated from the OLD role name. The backend now reports the rule under the new
			// role name; a name-keyed delete would regenerate the new name and miss this object.
			beforeRename := departmentAdminRule("department-administrator")
			afterRename := departmentAdminRule("department-administrator-(legacy)")

			orphanName := GenerateAccessRuleName(&beforeRename)
			h := newAccessRulesHandlerWithObjects(accessRuleObject(orphanName, "11"))

			Expect(accessRuleExists(h, orphanName)).To(BeTrue())

			err := h.deleteAccessRulesByIds([]string{accessRuleIdLabelValue(&afterRename)})

			Expect(err).ToNot(HaveOccurred())
			Expect(accessRuleExists(h, orphanName)).To(BeFalse())
		})

		It("reaches an orphan that the name-keyed delete cannot", func() {
			// Demonstrates the bug the label-keyed delete fixes. DeleteObjects is the name-keyed
			// path the handler used before: given the rule as the backend now reports it, it
			// regenerates the post-rename name and never touches the object that is actually there.
			beforeRename := departmentAdminRule("department-administrator")
			afterRename := departmentAdminRule("department-administrator-(legacy)")

			orphanName := GenerateAccessRuleName(&beforeRename)
			h := newAccessRulesHandlerWithObjects(accessRuleObject(orphanName, "11"))

			_, err := h.DeleteObjects([]types.NamespacedName{
				{Name: GenerateAccessRuleName(&afterRename)},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(accessRuleExists(h, orphanName)).To(BeTrue(), "name-keyed delete misses the orphan")

			err = h.deleteAccessRulesByIds([]string{accessRuleIdLabelValue(&afterRename)})
			Expect(err).ToNot(HaveOccurred())
			Expect(accessRuleExists(h, orphanName)).To(BeFalse(), "label-keyed delete reaches it")
		})

		It("deletes every object sharing one access rule id", func() {
			// After an upgrade both the original and the duplicate exist for the same id; deleting
			// the access rule in the control plane must remove both.
			beforeRename := departmentAdminRule("department-administrator")
			afterRename := departmentAdminRule("department-administrator-(legacy)")

			originalName := GenerateAccessRuleName(&beforeRename)
			duplicateName := GenerateAccessRuleName(&afterRename)

			h := newAccessRulesHandlerWithObjects(
				accessRuleObject(originalName, "11"),
				accessRuleObject(duplicateName, "11"),
			)

			err := h.deleteAccessRulesByIds([]string{"11"})

			Expect(err).ToNot(HaveOccurred())
			Expect(accessRuleExists(h, originalName)).To(BeFalse())
			Expect(accessRuleExists(h, duplicateName)).To(BeFalse())
		})

		It("leaves objects belonging to other access rules alone", func() {
			rule := departmentAdminRule("department-administrator")
			targetName := GenerateAccessRuleName(&rule)

			h := newAccessRulesHandlerWithObjects(
				accessRuleObject(targetName, "11"),
				accessRuleObject("ar-12-department-4500002-viewer-otheruser", "12"),
			)

			err := h.deleteAccessRulesByIds([]string{"11"})

			Expect(err).ToNot(HaveOccurred())
			Expect(accessRuleExists(h, targetName)).To(BeFalse())
			Expect(accessRuleExists(h, "ar-12-department-4500002-viewer-otheruser")).To(BeTrue())
		})

		It("reports NotFound per id when nothing carries the label, without erroring", func() {
			h := newAccessRulesHandlerWithObjects()

			err := h.deleteAccessRulesByIds([]string{"11"})

			// An access rule the cluster never had is not a sync failure - the same way the
			// name-keyed delete treated IsNotFound as success.
			Expect(err).ToNot(HaveOccurred())
			Expect(allAccessRules(h)).To(BeEmpty())
		})
	})
})

func ptrTo[T any](v T) *T {
	return &v
}
