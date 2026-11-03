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

// listCountingClient counts List calls so a test can assert that the label index - a full scan of
// the cache, since labels are never indexed - is not built when nothing was created.
type listCountingClient struct {
	client.Client
	lists int
}

func (c *listCountingClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	c.lists++
	return c.Client.List(ctx, list, opts...)
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

	Context("findRenamedAccessRules", func() {
		// reconcileAll mirrors what Handle does: create/update the desired objects, then drop what a
		// rename left behind.
		reconcileAll := func(h *AccessRulesHandler, desired []client.Object) error {
			renamed, err := h.findRenamedAccessRules(desired)
			if err != nil {
				return err
			}
			if _, _, err := h.CreateOrUpdateObjects(desired); err != nil {
				return err
			}
			return h.deleteRenamedAccessRules(renamed)
		}

		It("pairs the desired object with the one left behind by the rename", func() {
			// RUN-44340: the object on the cluster was created before the rename, so it carries a
			// name the agent no longer generates.
			beforeRename := departmentAdminRule("department-administrator")
			afterRename := departmentAdminRule("department-administrator-(legacy)")

			staleName := GenerateAccessRuleName(&beforeRename)
			desiredName := GenerateAccessRuleName(&afterRename)
			Expect(staleName).ToNot(Equal(desiredName))

			h := newAccessRulesHandlerWithObjects(accessRuleObject(staleName, "11"))
			desired, err := convertBackendToClusterAccessRule(&afterRename)
			Expect(err).ToNot(HaveOccurred())

			renamed, err := h.findRenamedAccessRules([]client.Object{desired})

			Expect(err).ToNot(HaveOccurred())
			Expect(renamed).To(HaveLen(1))
			Expect(renamed[desiredName]).ToNot(BeNil())
			Expect(renamed[desiredName].GetName()).To(Equal(staleName))
		})

		It("ends with exactly one object once the rename is reconciled", func() {
			beforeRename := departmentAdminRule("department-administrator")
			afterRename := departmentAdminRule("department-administrator-(legacy)")

			staleName := GenerateAccessRuleName(&beforeRename)
			desiredName := GenerateAccessRuleName(&afterRename)

			h := newAccessRulesHandlerWithObjects(accessRuleObject(staleName, "11"))
			desired, err := convertBackendToClusterAccessRule(&afterRename)
			Expect(err).ToNot(HaveOccurred())

			Expect(reconcileAll(h, []client.Object{desired})).To(Succeed())

			Expect(accessRuleExists(h, desiredName)).To(BeTrue(), "the current object exists")
			Expect(accessRuleExists(h, staleName)).To(BeFalse(), "the pre-rename object is gone")
			Expect(allAccessRules(h)).To(HaveLen(1), "exactly one object per access rule id")
		})

		It("finds nothing when the name did not change", func() {
			rule := departmentAdminRule("department-administrator")
			name := GenerateAccessRuleName(&rule)

			h := newAccessRulesHandlerWithObjects(accessRuleObject(name, "11"))
			desired, err := convertBackendToClusterAccessRule(&rule)
			Expect(err).ToNot(HaveOccurred())

			renamed, err := h.findRenamedAccessRules([]client.Object{desired})

			Expect(err).ToNot(HaveOccurred())
			Expect(renamed).To(BeEmpty())

			Expect(reconcileAll(h, []client.Object{desired})).To(Succeed())
			Expect(accessRuleExists(h, name)).To(BeTrue())
			Expect(allAccessRules(h)).To(HaveLen(1))
		})

		It("does not pair objects belonging to other access rules", func() {
			beforeRename := departmentAdminRule("department-administrator")
			afterRename := departmentAdminRule("department-administrator-(legacy)")

			h := newAccessRulesHandlerWithObjects(
				accessRuleObject(GenerateAccessRuleName(&beforeRename), "11"),
				accessRuleObject("ar-12-department-4500002-viewer-otheruser", "12"),
			)
			desired, err := convertBackendToClusterAccessRule(&afterRename)
			Expect(err).ToNot(HaveOccurred())

			Expect(reconcileAll(h, []client.Object{desired})).To(Succeed())

			Expect(accessRuleExists(h, "ar-12-department-4500002-viewer-otheruser")).To(BeTrue())
		})

		It("finds nothing when the access rule has no object yet", func() {
			rule := departmentAdminRule("department-administrator")
			h := newAccessRulesHandlerWithObjects()

			desired, err := convertBackendToClusterAccessRule(&rule)
			Expect(err).ToNot(HaveOccurred())

			renamed, err := h.findRenamedAccessRules([]client.Object{desired})
			Expect(err).ToNot(HaveOccurred())
			Expect(renamed).To(BeEmpty())

			Expect(reconcileAll(h, []client.Object{desired})).To(Succeed())
			Expect(accessRuleExists(h, GenerateAccessRuleName(&rule))).To(BeTrue())
			Expect(allAccessRules(h)).To(HaveLen(1))
		})
	})

	Context("findRenamedAccessRules cost", func() {
		// Listing by label is a full scan: the informer cache indexes namespace and registered field
		// selectors only, never labels. So the scan must not be paid on every sync cycle.
		newCountingHandler := func(objs ...client.Object) (*AccessRulesHandler, *listCountingClient) {
			h := newAccessRulesHandlerWithObjects(objs...)
			counting := &listCountingClient{Client: h.runaiClient}
			h.runaiClient = counting
			h.cachedClient = counting
			return h, counting
		}

		It("does not scan when every object is already there under the same name", func() {
			rule := departmentAdminRule("department-administrator")
			h, counting := newCountingHandler(accessRuleObject(GenerateAccessRuleName(&rule), "11"))

			desired, err := convertBackendToClusterAccessRule(&rule)
			Expect(err).ToNot(HaveOccurred())

			_, err = h.findRenamedAccessRules([]client.Object{desired})

			Expect(err).ToNot(HaveOccurred())
			Expect(counting.lists).To(Equal(0), "steady state must cost no scan")
		})

		It("scans at most once, however many objects are missing by name", func() {
			h, counting := newCountingHandler()

			desired := []client.Object{}
			for _, roleName := range []string{"role-a", "role-b", "role-c"} {
				rule := departmentAdminRule(roleName)
				rule.Id = int32(len(desired) + 20)
				object, err := convertBackendToClusterAccessRule(&rule)
				Expect(err).ToNot(HaveOccurred())
				desired = append(desired, object)
			}

			_, err := h.findRenamedAccessRules(desired)

			Expect(err).ToNot(HaveOccurred())
			Expect(counting.lists).To(Equal(1), "three misses must still cost one scan, not three")
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
