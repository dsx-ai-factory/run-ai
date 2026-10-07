/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package accessrule

// Unit coverage for the kai.resources project paths of the event filter and project mapper.

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("ControllerPredicateFuncs (project updates)", func() {
	funcs := ControllerPredicateFuncs()

	It("admits a kai Project update only when parent or namespace changed", func() {
		base := &kaiorg.Project{
			ObjectMeta: metav1.ObjectMeta{Name: "p1"},
			Spec:       kaiorg.ProjectSpec{Parent: "dep-a"},
		}
		base.Status.Namespace = "ns-a"

		unchanged := base.DeepCopy()
		Expect(funcs.Update(event.UpdateEvent{ObjectOld: base, ObjectNew: unchanged})).To(BeFalse(),
			"a kai project update that changes neither parent nor namespace must be filtered out")

		parentChanged := base.DeepCopy()
		parentChanged.Spec.Parent = "dep-b"
		Expect(funcs.Update(event.UpdateEvent{ObjectOld: base, ObjectNew: parentChanged})).To(BeTrue())

		nsChanged := base.DeepCopy()
		nsChanged.Status.Namespace = "ns-b"
		Expect(funcs.Update(event.UpdateEvent{ObjectOld: base, ObjectNew: nsChanged})).To(BeTrue())
	})
})

var _ = Describe("MapProjectToAccessRuleEvent", func() {
	newControllerWithRules := func(rules ...client.Object) *AccessRuleController {
		scheme := scheme_init.Scheme()
		Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
		fakeClient := fakeclient.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(rules...).
			WithIndex(&v1alpha1.AccessRule{}, common.AccessRuleProjectNameField, AccessRulesProjectIndexer).
			WithIndex(&v1alpha1.AccessRule{}, common.AccessRuleDepartmentNameField, AccessRulesDepartmentIndexer).
			WithIndex(&v1alpha1.AccessRule{}, common.AccessRuleClusterField, AccessRulesClusterIndexer).
			Build()
		return NewAccessRuleController(fakeClient)
	}

	ar := func(name string, scopeType v1alpha1.ScopeType, scopeName string) *v1alpha1.AccessRule {
		return &v1alpha1.AccessRule{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec:       v1alpha1.AccessRuleSpec{ScopeType: scopeType, ScopeName: scopeName, RoleName: "role-a"},
		}
	}

	It("triggers reconcile of project-, department- and cluster-scoped rules that apply to the kai project", func() {
		c := newControllerWithRules(
			ar("ar-proj-p1", v1alpha1.ProjectScope, "p1"),
			ar("ar-proj-p2", v1alpha1.ProjectScope, "p2"),
			ar("ar-dep-a", v1alpha1.DepartmentScope, "dep-a"),
			ar("ar-dep-b", v1alpha1.DepartmentScope, "dep-b"),
			ar("ar-cluster", v1alpha1.ClusterScope, ""),
		)

		proj := &kaiorg.Project{
			ObjectMeta: metav1.ObjectMeta{Name: "p1"},
			Spec:       kaiorg.ProjectSpec{Parent: "dep-a"},
		}

		reqs := c.MapProjectToAccessRuleEvent(context.Background(), proj)
		names := make([]string, 0, len(reqs))
		for _, r := range reqs {
			names = append(names, r.NamespacedName.Name)
		}
		Expect(names).To(ConsistOf("ar-proj-p1", "ar-dep-a", "ar-cluster"))
	})
})
