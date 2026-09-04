/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mappers_test

// Unit coverage for the AccessRulesProjectsMapper against kai.resources org units.

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/mappers"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// newKaiFakeClient builds a fake client with the kai.resources Project parent index the manager
// installs (see AccessRuleController.indexFields), plus the shared AccessRule indexes.
func newKaiFakeClient(objs ...client.Object) client.Client {
	scheme := scheme_init.Scheme()
	Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
	return fakeclient.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		WithIndex(&kaiorg.Project{}, common.ProjectsParentNameField, accessrule.ProjectsParentIndexer).
		WithIndex(&v1alpha1.AccessRule{}, common.AccessRuleProjectNameField, accessrule.AccessRulesProjectIndexer).
		WithIndex(&v1alpha1.AccessRule{}, common.AccessRuleDepartmentNameField, accessrule.AccessRulesDepartmentIndexer).
		WithIndex(&v1alpha1.AccessRule{}, common.AccessRuleClusterField, accessrule.AccessRulesClusterIndexer).
		Build()
}

func kaiProject(name, parent string) *kaiorg.Project {
	return &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       kaiorg.ProjectSpec{Parent: parent},
	}
}

func kaiDepartment(name string) *kaiorg.Department {
	return &kaiorg.Department{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

func kaiProjectNames(projects []kaiorg.Project) []string {
	names := make([]string, 0, len(projects))
	for _, p := range projects {
		names = append(names, p.Name)
	}
	return names
}

var _ = Describe("AccessRulesProjectsMapper", func() {
	ctx := context.Background()

	Describe("GetProjectsOfAccessRuleScope", func() {
		It("project scope - returns the single named project", func() {
			c := newKaiFakeClient(kaiProject("p1", depA), kaiProject("p2", depB))
			mapper := mappers.NewAccessRulesProjectsMapper(c)

			projects, err := mapper.GetProjectsOfAccessRuleScope(ctx, accessRule("ar", v1alpha1.ProjectScope, "p1"))
			Expect(err).NotTo(HaveOccurred())
			Expect(kaiProjectNames(projects)).To(ConsistOf("p1"))
		})

		It("project scope - errors when the project does not exist", func() {
			c := newKaiFakeClient()
			mapper := mappers.NewAccessRulesProjectsMapper(c)

			_, err := mapper.GetProjectsOfAccessRuleScope(ctx, accessRule("ar", v1alpha1.ProjectScope, "missing"))
			Expect(err).To(HaveOccurred())
		})

		It("department scope - returns only projects whose parent is that department", func() {
			c := newKaiFakeClient(
				kaiDepartment(depA),
				kaiProject("p1", depA), kaiProject("p2", depA), kaiProject("p3", depB))
			mapper := mappers.NewAccessRulesProjectsMapper(c)

			projects, err := mapper.GetProjectsOfAccessRuleScope(ctx, accessRule("ar", v1alpha1.DepartmentScope, depA))
			Expect(err).NotTo(HaveOccurred())
			Expect(kaiProjectNames(projects)).To(ConsistOf("p1", "p2"))
		})

		It("department scope - errors when the kai department does not exist", func() {
			c := newKaiFakeClient(kaiProject("p1", depA))
			mapper := mappers.NewAccessRulesProjectsMapper(c)

			_, err := mapper.GetProjectsOfAccessRuleScope(ctx, accessRule("ar", v1alpha1.DepartmentScope, depA))
			Expect(err).To(HaveOccurred())
		})

		It("cluster scope - returns all projects", func() {
			c := newKaiFakeClient(kaiProject("p1", depA), kaiProject("p2", depB))
			mapper := mappers.NewAccessRulesProjectsMapper(c)

			projects, err := mapper.GetProjectsOfAccessRuleScope(ctx, accessRule("ar", v1alpha1.ClusterScope, ""))
			Expect(err).NotTo(HaveOccurred())
			Expect(kaiProjectNames(projects)).To(ConsistOf("p1", "p2"))
		})

		It("errors on an unrecognized scope type", func() {
			c := newKaiFakeClient()
			mapper := mappers.NewAccessRulesProjectsMapper(c)

			_, err := mapper.GetProjectsOfAccessRuleScope(ctx, accessRule("ar", v1alpha1.ScopeType("bogus"), "x"))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetAllAccessRulesForProject", func() {
		It("returns project-scoped, department-scoped and cluster-scoped rules that apply to the project", func() {
			c := newKaiFakeClient(
				accessRule("ar-proj-p1", v1alpha1.ProjectScope, "p1"),
				accessRule("ar-proj-p2", v1alpha1.ProjectScope, "p2"),
				accessRule("ar-dep-a", v1alpha1.DepartmentScope, depA),
				accessRule("ar-dep-b", v1alpha1.DepartmentScope, depB),
				accessRule("ar-cluster", v1alpha1.ClusterScope, ""),
			)
			mapper := mappers.NewAccessRulesProjectsMapper(c)

			rules, err := mapper.GetAllAccessRulesForProject(ctx, kaiProject("p1", depA))
			Expect(err).NotTo(HaveOccurred())
			Expect(accessRuleNames(rules)).To(ConsistOf("ar-proj-p1", "ar-dep-a", "ar-cluster"))
		})

		It("returns only the cluster-scoped rule when nothing else applies", func() {
			c := newKaiFakeClient(
				accessRule("ar-proj-p2", v1alpha1.ProjectScope, "p2"),
				accessRule("ar-dep-b", v1alpha1.DepartmentScope, depB),
				accessRule("ar-cluster", v1alpha1.ClusterScope, ""),
			)
			mapper := mappers.NewAccessRulesProjectsMapper(c)

			rules, err := mapper.GetAllAccessRulesForProject(ctx, kaiProject("p1", depA))
			Expect(err).NotTo(HaveOccurred())
			Expect(accessRuleNames(rules)).To(ConsistOf("ar-cluster"))
		})
	})
})
