/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/rolebindingsutils"
	utils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	rbacGroup = "rbac.authorization.k8s.io"

	testProjectA  = "team-a"
	testProjectA2 = "team-a-2"
	testProjectB  = "team-b"

	testDefaultDepartment = "default"
	testDepartmentA       = "dep-a"

	testClusterRoleA = "role-a"
	testClusterRoleB = "role-b"

	userA  = "user-a"
	groupA = "group-a"
	appA   = "app-a"
)

var (
	testNamespaceA  = fmt.Sprintf("runai-%s", testProjectA)
	testNamespaceA2 = fmt.Sprintf("%s-ns", testProjectA2) // to test non "runai-" prefixed ns
	testNamespaceB  = fmt.Sprintf("runai-%s", testProjectB)

	// mapping each project to its namespace
	testProjectsToNamespace = map[string]string{
		testProjectA:  testNamespaceA,
		testProjectA2: testNamespaceA2,
		testProjectB:  testNamespaceB}
	// mapping each namespace to its project
	testNamespacesToProject = map[string]string{
		testNamespaceA:  testProjectA,
		testNamespaceA2: testProjectA2,
		testNamespaceB:  testProjectB}
	testDepartments                                            = []string{testDefaultDepartment, testDepartmentA}
	testClusterRoles                                           = []string{testClusterRoleA, testClusterRoleB}
	testClusterRolesWithClusterScopedRunaiNamespaceRoleBinding = map[string]bool{
		testClusterRoleA: true,
	}
	// mapping each project to its department
	testProjectsToDepartment = map[string]string{
		testProjectA:  testDepartmentA,
		testProjectA2: testDepartmentA,
		testProjectB:  testDefaultDepartment,
	}
)

var _ = Describe("Access Rule Controller Tests", Ordered, func() {
	var (
		accessRule          *v1alpha1.AccessRule
		testAccessRule      *v1alpha1.AccessRule
		roleBindingsCreated []types.NamespacedName

		dontValidateRoleBindingDeleted bool
	)

	BeforeAll(func() {
		controllerSetup()

		afterAllCleanup()

		createNamespaceIfNeeded(runaiv1.DefaultNamespace, "", k8sClient)
		for testProject, testNamespace := range testProjectsToNamespace {
			createNamespaceIfNeeded(testNamespace, testProject, k8sClient)
		}
		for _, testDepartment := range testDepartments {
			createDepartment(testDepartment, k8sClient)
		}
		for _, testClusterRole := range testClusterRoles {
			createNeededClusterRolesAndBindingsForTestRole(testClusterRole)
		}
		for testProject, testDepartment := range testProjectsToDepartment {
			testNamespace := testProjectsToNamespace[testProject]
			createProject(testProject, testDepartment, testNamespace, k8sClient)
		}
	})

	AfterAll(func() {
		afterAllCleanup()
	})

	AfterEach(func() {
		if accessRule != nil {
			deleteAccessRule(accessRule.Name, k8sClient)
			accessRule = nil
		}
		if testAccessRule != nil {
			deleteAccessRule(testAccessRule.Name, k8sClient)
			testAccessRule = nil
		}

		for _, roleBinding := range roleBindingsCreated {
			if !dontValidateRoleBindingDeleted {
				eventuallyExpectRoleBindingDeleted(roleBinding,
					"Expected Role Binding <%s/%s> to be deleted after owner access rule is deleted", k8sClient)
			} else {
				deleteRoleBinding(roleBinding.Name, roleBinding.Namespace, k8sClient)
			}
		}
		roleBindingsCreated = []types.NamespacedName{}

		dontValidateRoleBindingDeleted = false
	})

	Describe("Sanity - create role bindings for each Scope type", Ordered, func() {
		It("Sanity for project-scoped access role", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   testProjectA,
				ScopeType:   v1alpha1.ProjectScope,
				RoleName:    testClusterRoleA,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA))

			validateAccessRuleReady(accessRuleName, k8sClient)
		})
		It("Sanity for department-scoped access role", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   groupA,
				SubjectType: v1alpha1.GroupSubject,
				ScopeName:   testDepartmentA,
				ScopeType:   v1alpha1.DepartmentScope,
				RoleName:    testClusterRoleB,
			}, k8sClient)

			// RoleBinding should be created in namespaces of both projects that belong to department
			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA, testNamespaceA2}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA))
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA2))

			validateAccessRuleReady(accessRuleName, k8sClient)
		})
		It("Sanity for cluster-scoped access role", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeType:   v1alpha1.ClusterScope,
				RoleName:    testClusterRoleB,
			}, k8sClient)

			// RoleBinding should be created in namespaces of all projects
			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA, testNamespaceA2, testNamespaceB}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA))
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA2))
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceB))

			validateAccessRuleReady(accessRuleName, k8sClient)
		})
	})

	Describe("Update access role and see RoleBindings updated/created/deleted", Ordered, func() {
		It("Update Scope Name for project-scoped access role", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   testProjectA,
				ScopeType:   v1alpha1.ProjectScope,
				RoleName:    testClusterRoleA,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA))

			validateAccessRuleReady(accessRuleName, k8sClient)
			accessRule = getAccessRuleFromClient(accessRuleName, k8sClient)
			accessRule.Spec.ScopeName = testProjectA2
			utils.ExpectUpdateResource(k8sClient, accessRule)

			eventuallyExpectRoleBindingDeleted(nsName(rbName, testNamespaceA),
				"Expected Role Binding <%s/%s> to be deleted after access rule points to different project",
				k8sClient)
			roleBindingsCreated = []types.NamespacedName{}

			// RoleBinding should now exist in the project the access rule points to
			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA2}, k8sClient)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA2))

			validateAccessRuleReady(accessRuleName, k8sClient)
		})
		It("Update Scope Name for department-scoped access role", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   appA,
				SubjectType: v1alpha1.AppSubject,
				ScopeName:   testDepartmentA,
				ScopeType:   v1alpha1.DepartmentScope,
				RoleName:    testClusterRoleB,
			}, k8sClient)

			// RoleBinding should be created in namespaces of both projects that belong to department
			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA, testNamespaceA2}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA))
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA2))

			Eventually(func(g Gomega) {
				accessRule = getAccessRuleFromClient(accessRuleName, k8sClient)
				g.Expect(accessRule.Status.Phase).To(Equal(v1alpha1.AccessRuleReady))
			}, timeout, interval).Should(Succeed())

			accessRule.Spec.ScopeName = testDefaultDepartment
			utils.ExpectUpdateResource(k8sClient, accessRule)

			for _, roleBinding := range roleBindingsCreated {
				eventuallyExpectRoleBindingDeleted(roleBinding,
					"Expected Role Binding <%s/%s> to be deleted after access rule points to different department",
					k8sClient)
			}
			roleBindingsCreated = []types.NamespacedName{}

			// RoleBinding should now exist in the project belonging to department testDefaultDepartment
			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceB}, k8sClient)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceB))

			validateAccessRuleReady(accessRuleName, k8sClient)
		})
		It("Update Scope Type from department to cluster", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   testDepartmentA,
				ScopeType:   v1alpha1.DepartmentScope,
				RoleName:    testClusterRoleA,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA, testNamespaceA2}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA))
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA2))

			accessRule = getAccessRuleFromClient(accessRuleName, k8sClient)
			Expect(accessRule.Status.Phase).To(Equal(v1alpha1.AccessRuleReady))
			accessRule.Spec.ScopeType = v1alpha1.ClusterScope
			accessRule.Spec.ScopeName = ""
			utils.ExpectUpdateResource(k8sClient, accessRule)

			// RoleBinding should now exist in all projects
			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA, testNamespaceA2, testNamespaceB}, k8sClient)

			validateAccessRuleReady(accessRuleName, k8sClient)
		})
		It("Update Scope Type from project to department", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   testProjectA,
				ScopeType:   v1alpha1.ProjectScope,
				RoleName:    testClusterRoleA,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA))

			accessRule = getAccessRuleFromClient(accessRuleName, k8sClient)
			Expect(accessRule.Status.Phase).To(Equal(v1alpha1.AccessRuleReady))
			accessRule.Spec.ScopeType = v1alpha1.DepartmentScope
			accessRule.Spec.ScopeName = testDefaultDepartment
			utils.ExpectUpdateResource(k8sClient, accessRule)

			eventuallyExpectRoleBindingDeleted(nsName(rbName, testNamespaceA),
				"Expected Role Binding <%s/%s> to be deleted after access rule points to different department",
				k8sClient)
			roleBindingsCreated = []types.NamespacedName{}

			// RoleBinding should now exist in the project belonging to department testDefaultDepartment
			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceB}, k8sClient)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceB))

			validateAccessRuleReady(accessRuleName, k8sClient)
		})
		It("Update Scope Type from cluster to project", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   groupA,
				SubjectType: v1alpha1.GroupSubject,
				ScopeType:   v1alpha1.ClusterScope,
				RoleName:    testClusterRoleA,
			}, k8sClient)

			// RoleBinding should be created in namespaces of all projects
			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA, testNamespaceA2, testNamespaceB}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA))
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA2))
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceB))

			clusterScopedRunaiNamespaceRoleBindingName := rolebindingsutils.GetClusterScopedRunaiNamespaceRoleBindingName(accessRule.Spec.RoleName)
			Eventually(func(g Gomega) {
				roleBinding := getRoleBindingFromClient(clusterScopedRunaiNamespaceRoleBindingName, "runai", k8sClient)
				g.Expect(roleBinding).ToNot(BeNil())
				g.Expect(roleBinding.Subjects).To(ContainElement(getExpectedSubjectForAccessRule(accessRule)),
					"expected role binding's <%s/%s> Subjects to contain expected <%v>",
					clusterScopedRunaiNamespaceRoleBindingName, "runai", getExpectedSubjectForAccessRule(accessRule))
			}, timeout, interval).Should(Succeed())

			Eventually(func(g Gomega) {
				accessRule = getAccessRuleFromClient(accessRuleName, k8sClient)
				g.Expect(accessRule.Status.Phase).To(Equal(v1alpha1.AccessRuleReady))
				accessRule.Spec.ScopeName = testProjectB
				accessRule.Spec.ScopeType = v1alpha1.ProjectScope
				err := k8sClient.Update(apiCtx, accessRule)
				g.Expect(err).ToNot(HaveOccurred())
			}, timeout, interval).Should(Succeed())

			eventuallyExpectRoleBindingDeleted(nsName(rbName, testNamespaceA),
				"Expected Role Binding <%s/%s> to be deleted after access rule points to specific project",
				k8sClient)
			eventuallyExpectRoleBindingDeleted(nsName(rbName, testNamespaceA2),
				"Expected Role Binding <%s/%s> to be deleted after access rule points to specific project",
				k8sClient)
			// cluster-scoped runai namespace rolebinding should have empty subjects, now that the scope is not cluster-scoped
			Eventually(func(g Gomega) {
				roleBinding := getRoleBindingFromClient(clusterScopedRunaiNamespaceRoleBindingName, "runai", k8sClient)
				g.Expect(roleBinding).ToNot(BeNil())
				g.Expect(roleBinding.Subjects).To(BeEmpty(),
					"expected role binding's <%s/%s> Subjects to be empty",
					clusterScopedRunaiNamespaceRoleBindingName, "runai")
			}, timeout, interval).Should(Succeed())

			// RoleBinding should now exist only in the project the access rule points to
			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceB}, k8sClient)
			consistentlyValidateRoleBinding(accessRule, []string{testNamespaceB}, k8sClient)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceB))

			validateAccessRuleReady(accessRuleName, k8sClient)
		})
		It("Update Scope to un-existing project - see old RoleBindings deleted", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   testDepartmentA,
				ScopeType:   v1alpha1.DepartmentScope,
				RoleName:    testClusterRoleA,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA, testNamespaceA2}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA))
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA2))

			accessRule = getAccessRuleFromClient(accessRuleName, k8sClient)
			Expect(accessRule.Status.Phase).To(Equal(v1alpha1.AccessRuleReady))
			accessRule.Spec.ScopeType = v1alpha1.ProjectScope
			accessRule.Spec.ScopeName = "un-existing-proj"
			utils.ExpectUpdateResource(k8sClient, accessRule)

			for _, roleBinding := range roleBindingsCreated {
				eventuallyExpectRoleBindingDeleted(roleBinding,
					"Expected Role Binding <%s/%s> to be deleted after access rule points to unexisting project",
					k8sClient)
			}
			roleBindingsCreated = []types.NamespacedName{}

			eventuallyValidateAllRoleBindingsDontHaveAccessRuleSubject(accessRule,
				[]string{testNamespaceA, testNamespaceA2}, k8sClient)

			eventuallyValidateAccessRuleNotReady(accessRuleName, k8sClient,
				fmt.Sprintf(common.ScopeObjectNotFoundMessage, "project", accessRule.Spec.ScopeName))
		})
		It("Update Scope to un-existing department - see old RoleBindings deleted", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   testProjectA,
				ScopeType:   v1alpha1.ProjectScope,
				RoleName:    testClusterRoleA,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA))

			accessRule = getAccessRuleFromClient(accessRuleName, k8sClient)
			Expect(accessRule.Status.Phase).To(Equal(v1alpha1.AccessRuleReady))
			accessRule.Spec.ScopeType = v1alpha1.DepartmentScope
			accessRule.Spec.ScopeName = "un-existing-dep"
			utils.ExpectUpdateResource(k8sClient, accessRule)

			eventuallyExpectRoleBindingDeleted(nsName(rbName, testNamespaceA),
				"Expected Role Binding <%s/%s> to be deleted after access rule points to unexisting department",
				k8sClient)
			roleBindingsCreated = []types.NamespacedName{}

			eventuallyValidateAllRoleBindingsDontHaveAccessRuleSubject(accessRule,
				[]string{testNamespaceA}, k8sClient)

			validateAccessRuleNotReady(accessRuleName,
				fmt.Sprintf(common.ScopeObjectNotFoundMessage, "department", accessRule.Spec.ScopeName),
				k8sClient)
		})
		It("Update Scope Name for project-scoped access role - with another accessrule existing - see old RoleBindings not deleted", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   testProjectA,
				ScopeType:   v1alpha1.ProjectScope,
				RoleName:    testClusterRoleA,
			}, k8sClient)
			testAccessRuleName := generateAccessRuleName()
			testAccessRule = createAccessRule(testAccessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   testDepartmentA,
				ScopeType:   v1alpha1.DepartmentScope,
				RoleName:    testClusterRoleA,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA}, k8sClient)
			eventuallyValidateRoleBindingsForAccessRule(testAccessRule, []string{testNamespaceA, testNamespaceA2}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA))

			Eventually(func(g Gomega) {
				accessRule = getAccessRuleFromClient(accessRuleName, k8sClient)
				g.Expect(accessRule.Status.Phase).To(Equal(v1alpha1.AccessRuleReady))
			}, timeout, interval).Should(Succeed())
			accessRule.Spec.ScopeName = testProjectA2
			utils.ExpectUpdateResource(k8sClient, accessRule)

			roleBinding := getRoleBindingFromClient(rbName, testNamespaceA, k8sClient)
			Expect(roleBinding).ToNot(BeNil(),
				"RoleBinding <%s/%s> should exist - testAccessRule points to same role with same subject",
				testNamespaceA, rbName)

			// RoleBinding should now exist in the project the access rule points to
			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA2}, k8sClient)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA2))

			validateAccessRuleReady(accessRuleName, k8sClient)
		})
	})

	Describe("Update RoleBindings and see them reconciled", Ordered, func() {
		It("Update Labels of RoleBinding", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   groupA,
				SubjectType: v1alpha1.GroupSubject,
				ScopeName:   testDepartmentA,
				ScopeType:   v1alpha1.DepartmentScope,
				RoleName:    testClusterRoleB,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceA))

			eventuallyUpdateRoleBindingLabels(k8sClient, rbName, testNamespaceA, nil)
			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA}, k8sClient)

			labelKey := rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, true)

			roleBinding := getRoleBindingFromClient(rbName, testNamespaceA, k8sClient)
			delete(roleBinding.Labels, labelKey)
			if roleBinding.Labels == nil {
				roleBinding.Labels = map[string]string{}
			}
			roleBinding.Labels["some-other-label-key"] = "some-other-label-value"
			eventuallyUpdateRoleBindingLabels(k8sClient, rbName, testNamespaceA, roleBinding.Labels)
			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA}, k8sClient)
			roleBinding = getRoleBindingFromClient(rbName, testNamespaceA, k8sClient)
			Expect(roleBinding.Labels[labelKey]).To(Equal("true"))
			Expect(roleBinding.Labels["some-other-label-key"]).To(Equal("some-other-label-value"))

			validateAccessRuleReady(accessRuleName, k8sClient)

			dontValidateRoleBindingDeleted = true
		})
		It("Update Subjects of RoleBinding", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeType:   v1alpha1.ClusterScope,
				RoleName:    testClusterRoleB,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceB}, k8sClient)
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingsCreated = append(roleBindingsCreated, nsName(rbName, testNamespaceB))

			eventuallyUpdateRoleBindingSubjects(k8sClient, rbName, testNamespaceB, []rbacv1.Subject{})

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceB}, k8sClient)

			eventuallyUpdateRoleBindingSubjects(k8sClient, rbName, testNamespaceB, []rbacv1.Subject{
				{
					APIGroup: rolebindingsutils.RbacGroup,
					Name:     "some-fake-user",
					Kind:     rolebindingsutils.UserKind,
				},
				{
					APIGroup: rolebindingsutils.RbacGroup,
					Name:     "some-fake-group",
					Kind:     rolebindingsutils.GroupKind,
				},
			})

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceB}, k8sClient)

			validateAccessRuleReady(accessRuleName, k8sClient)

			dontValidateRoleBindingDeleted = true
		})
		// RoleRef of RoleBinding cannot be updated - no need to test
	})

	Describe("Update/Create/Delete Project and see RoleBindings reconciled if needed", Ordered, func() {
		var (
			proj1Name         string
			proj1Namespace    string
			testProjName      string
			testProjNamespace string

			accessRuleDepartmentA       *v1alpha1.AccessRule
			accessRuleDefaultDepartment *v1alpha1.AccessRule
		)

		BeforeAll(func() {
			proj1Namespace = generateNamespaceName()
			proj1Name = generateProjectName()
			testProjNamespace = generateNamespaceName()
			testProjName = generateProjectName()

			testNamespacesToProject[testProjNamespace] = testProjName
			testNamespacesToProject[proj1Namespace] = proj1Name

			createNamespaceIfNeeded(testProjNamespace, testProjName, k8sClient)
			createNamespaceIfNeeded(proj1Namespace, proj1Name, k8sClient)
			createProject(proj1Name, testDepartmentA, proj1Namespace, k8sClient)

			accessRuleName := generateAccessRuleName()
			accessRuleDepartmentA = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   appA,
				SubjectType: v1alpha1.AppSubject,
				ScopeName:   testDepartmentA,
				ScopeType:   v1alpha1.DepartmentScope,
				RoleName:    testClusterRoleB,
			}, k8sClient)
			accessRuleName = generateAccessRuleName()
			accessRuleDefaultDepartment = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   testDefaultDepartment,
				ScopeType:   v1alpha1.DepartmentScope,
				RoleName:    testClusterRoleA,
			}, k8sClient)
		})

		AfterAll(func() {
			deleteAccessRule(accessRuleDepartmentA.Name, k8sClient)
			deleteAccessRule(accessRuleDefaultDepartment.Name, k8sClient)

			deleteProject(proj1Name, k8sClient)
			deleteProject(testProjName, k8sClient)
			deleteNamespace(proj1Namespace, k8sClient)
			deleteNamespace(testProjNamespace, k8sClient)
		})

		It("Update Project - un relevant Spec, nothing to update", func() {
			// RoleBinding should be created in the namespace of our project that belongs to department A
			eventuallyValidateRoleBindingsForAccessRule(accessRuleDepartmentA, []string{proj1Namespace}, k8sClient)

			// update un relevant spec value
			project := getProjectFromClient(proj1Name, k8sClient)
			project.Spec.TrainingJobMaxIdleDurationSecs = 8000
			utils.ExpectUpdateResource(k8sClient, project)

			// nothing should change
			consistentlyValidateRoleBinding(accessRuleDepartmentA, []string{proj1Namespace}, k8sClient)

			validateAccessRuleReady(accessRuleDepartmentA.Name, k8sClient)
		})
		It("Update Project's department - see access rule of department reconciled and RoleBindings deleted/created", func() {
			// RoleBinding for accessRuleDepartmentA should be created in the namespace of our project that belongs to department A
			eventuallyValidateRoleBindingsForAccessRule(accessRuleDepartmentA, []string{proj1Namespace}, k8sClient)
			validateAccessRuleReady(accessRuleDepartmentA.Name, k8sClient)

			project := getProjectFromClient(proj1Name, k8sClient)
			project.Spec.Department = testDefaultDepartment
			utils.ExpectUpdateResource(k8sClient, project)

			// mirror the department change onto the kai.resources twin (Spec.Parent) so the
			// reconciler observes it.
			kaiProject := getKaiProjectFromClient(proj1Name, k8sClient)
			kaiProject.Spec.Parent = testDefaultDepartment
			utils.ExpectUpdateResource(k8sClient, kaiProject)

			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRuleDepartmentA.Spec.RoleName)
			eventuallyExpectRoleBindingDeleted(nsName(rbName, proj1Namespace),
				"Expected Role Binding <%s/%s> to be deleted after project's department changed",
				k8sClient)

			// RoleBinding for accessRuleDefaultDepartment should now be created in the namespace of our project
			eventuallyValidateRoleBindingsForAccessRule(accessRuleDefaultDepartment, []string{proj1Namespace}, k8sClient)

			validateAccessRuleReady(accessRuleDefaultDepartment.Name, k8sClient)
		})
		It("Create/Delete project under department with access rule - see RoleBinding created/deleted", func() {
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRuleDepartmentA.Spec.RoleName)
			roleBindingObj := getRoleBindingFromClient(rbName, testProjNamespace, k8sClient)
			Expect(roleBindingObj).To(BeNil(),
				"RoleBinding <%s/%s> in namespace of un-existing project shouldn't exist",
				testProjNamespace, rbName)

			createProject(testProjName, testDepartmentA, testProjNamespace, k8sClient)
			// now that we created the project - the RoleBinding should exist
			eventuallyValidateRoleBindingsForAccessRule(accessRuleDepartmentA, []string{testProjNamespace}, k8sClient)

			validateAccessRuleReady(accessRuleDepartmentA.Name, k8sClient)

			deleteProject(testProjName, k8sClient)
			eventuallyExpectRoleBindingDeleted(nsName(rbName, testProjNamespace),
				"RoleBinding <%s/%s> should be deleted after project is deleted",
				k8sClient)
			eventuallyValidateRoleBindingsInNamespacesDontHaveAccessRuleSubject(accessRuleDepartmentA,
				[]string{testProjNamespace}, k8sClient)

			validateAccessRuleReady(accessRuleDepartmentA.Name, k8sClient)
		})
	})

	Describe("Reconcile for faulty projects (no namespace) - access rule status NotReady", Ordered, func() {
		var (
			proj1Name         string
			proj1Namespace    string
			proj2Name         string
			proj2Namespace    string
			testProjName      string
			testProjNamespace string

			dep3Name  string
			proj3Name string
		)

		BeforeAll(func() {
			proj1Namespace = generateNamespaceName()
			proj1Name = generateProjectName()
			proj2Namespace = generateNamespaceName()
			proj2Name = generateProjectName()
			testProjNamespace = generateNamespaceName()
			testProjName = generateProjectName()

			dep3Name = fmt.Sprintf("dep-%s", rand.String(5))
			proj3Name = generateProjectName()

			testNamespacesToProject[testProjNamespace] = testProjName

			createDepartment(dep3Name, k8sClient)

			// creating namespace but without the runai/queue label
			createNamespaceIfNeeded(proj1Namespace, "", k8sClient)
			createProject(proj1Name, testDepartmentA, "", k8sClient)

			// not creating the namespace - only project
			createProject(proj2Name, testDepartmentA, proj2Namespace, k8sClient)
			createProject(proj3Name, dep3Name, "", k8sClient)
		})

		AfterEach(func() {
			deleteProject(testProjName, k8sClient)
			deleteNamespace(testProjNamespace, k8sClient)
		})

		AfterAll(func() {
			deleteProject(proj1Name, k8sClient)
			deleteProject(proj2Name, k8sClient)
			deleteProject(proj3Name, k8sClient)
			deleteDepartment(dep3Name, k8sClient)
			deleteNamespace(proj1Namespace, k8sClient)
		})

		It("Access Rule for faulty project - NotReady", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   proj1Name,
				ScopeType:   v1alpha1.ProjectScope,
				RoleName:    testClusterRoleA,
			}, k8sClient)

			eventuallyValidateAccessRuleNotReady(accessRuleName, k8sClient,
				common.ScopeObjectFailedGettingNamespacesMessage)

			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingObj := getRoleBindingFromClient(rbName, proj1Namespace, k8sClient)
			Expect(roleBindingObj).To(BeNil(),
				"RoleBinding <%s/%s> shouldn't exist in namespace of faulty project %s",
				proj1Namespace, rbName, proj1Name)
		})
		It("Access Rule for department with faulty projects - Ready since some namespaces are ok", func() {
			// create namespace - will add runai/queue label later
			createNamespaceIfNeeded(testProjNamespace, "", k8sClient)
			createProject(testProjName, testDepartmentA, "", k8sClient)

			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   groupA,
				SubjectType: v1alpha1.GroupSubject,
				ScopeName:   testDepartmentA,
				ScopeType:   v1alpha1.DepartmentScope,
				RoleName:    testClusterRoleB,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testNamespaceA, testNamespaceA2},
				k8sClient)
			validateAccessRuleReady(accessRuleName, k8sClient)

			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingObj := getRoleBindingFromClient(rbName, proj1Namespace, k8sClient)
			Expect(roleBindingObj).To(BeNil(),
				"RoleBinding <%s/%s> shouldn't exist in namespace of faulty project %s",
				proj1Namespace, rbName, proj1Name)
			roleBindingObj = getRoleBindingFromClient(rbName, proj2Namespace, k8sClient)
			Expect(roleBindingObj).To(BeNil(),
				"RoleBinding <%s/%s> shouldn't exist in un-exisintg namespace of faulty project %s",
				proj2Namespace, rbName, proj2Name)
			roleBindingObj = getRoleBindingFromClient(rbName, testProjNamespace, k8sClient)
			Expect(roleBindingObj).To(BeNil(),
				"RoleBinding <%s/%s> shouldn't exist in un-exisintg namespace of faulty project %s",
				testProjNamespace, rbName, testProjName)

			// now add the missing project label
			namespaceObj := getNamespaceFromClient(testProjNamespace, k8sClient)
			Expect(namespaceObj).ToNot(BeNil())
			namespaceObj.Labels = map[string]string{"runai/queue": testProjName}
			utils.ExpectUpdateResource(k8sClient, namespaceObj)
			projObj := getProjectFromClient(testProjName, k8sClient)
			Expect(projObj).ToNot(BeNil())
			projObj.Status.Namespace = testProjNamespace
			Expect(k8sClient.Status().Update(apiCtx, projObj)).To(Succeed())

			// mirror the status namespace onto the kai.resources twin.
			kaiProjObj := getKaiProjectFromClient(testProjName, k8sClient)
			Expect(kaiProjObj).ToNot(BeNil())
			kaiProjObj.Status.Namespace = testProjNamespace
			Expect(k8sClient.Status().Update(apiCtx, kaiProjObj)).To(Succeed())

			// validate a RoleBinding was created in this namespace
			eventuallyValidateRoleBindingsForAccessRule(accessRule,
				[]string{testNamespaceA, testNamespaceA2, testProjNamespace}, k8sClient)
		})
		It("Access Rule for department with faulty projects - NotReady since all namespaces are faulty", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   groupA,
				SubjectType: v1alpha1.GroupSubject,
				ScopeName:   dep3Name,
				ScopeType:   v1alpha1.DepartmentScope,
				RoleName:    testClusterRoleB,
			}, k8sClient)

			eventuallyValidateAccessRuleNotReady(accessRuleName, k8sClient,
				common.ScopeObjectFailedGettingNamespacesMessage)
		})
		It("Access Rule for cluster with faulty projects - Ready", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   appA,
				SubjectType: v1alpha1.AppSubject,
				ScopeType:   v1alpha1.ClusterScope,
				RoleName:    testClusterRoleB,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule,
				[]string{testNamespaceA, testNamespaceA2, testNamespaceB}, k8sClient)

			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingObj := getRoleBindingFromClient(rbName, proj1Namespace, k8sClient)
			Expect(roleBindingObj).To(BeNil(),
				"RoleBinding <%s/%s> shouldn't exist in namespace of faulty project %s",
				proj1Namespace, rbName, proj1Name)
			roleBindingObj = getRoleBindingFromClient(rbName, proj2Namespace, k8sClient)
			Expect(roleBindingObj).To(BeNil(),
				"RoleBinding <%s/%s> shouldn't exist in un-exisintg namespace of faulty project %s",
				proj2Namespace, rbName, proj2Name)
		})
		It("Access Rule with service-account subject type creates User RoleBinding", func() {
			accessRuleName := generateAccessRuleName()
			serviceAccountId := "test-service-account-id"
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   serviceAccountId,
				SubjectType: v1alpha1.ServiceAccountSubject,
				ScopeType:   v1alpha1.ClusterScope,
				RoleName:    testClusterRoleB,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule,
				[]string{testNamespaceA, testNamespaceA2, testNamespaceB}, k8sClient)

			// Verify that service-account is treated as User kind in RoleBinding
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBinding := getRoleBindingFromClient(rbName, testNamespaceA, k8sClient)
			Expect(roleBinding).ToNot(BeNil(), "RoleBinding should exist for service-account subject")

			// Find the subject in the RoleBinding
			var foundSubject *rbacv1.Subject
			for i := range roleBinding.Subjects {
				if roleBinding.Subjects[i].Name == serviceAccountId {
					foundSubject = &roleBinding.Subjects[i]
					break
				}
			}
			Expect(foundSubject).ToNot(BeNil(), "Service account subject should be in RoleBinding")
			Expect(foundSubject.Kind).To(Equal("User"), "Service account should be mapped to User kind")
			Expect(foundSubject.APIGroup).To(Equal("rbac.authorization.k8s.io"))

			validateAccessRuleReady(accessRuleName, k8sClient)
		})
		It("Access Rule for valid project - that then becomes faulty - see RoleBinding deleted", func() {
			// creating namespace normally - will remove runai/queue label later
			createNamespaceIfNeeded(testProjNamespace, testProjName, k8sClient)
			createProject(testProjName, testDefaultDepartment, "", k8sClient)

			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   groupA,
				SubjectType: v1alpha1.GroupSubject,
				ScopeName:   testProjName,
				ScopeType:   v1alpha1.ProjectScope,
				RoleName:    testClusterRoleB,
			}, k8sClient)

			eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{testProjNamespace}, k8sClient)
			validateAccessRuleReady(accessRuleName, k8sClient)

			// now remove the runai/queue label from the namespace - which will turn the project faulty
			namespaceObj := getNamespaceFromClient(testProjNamespace, k8sClient)
			Expect(namespaceObj).ToNot(BeNil())
			namespaceObj.Labels = map[string]string{}
			utils.ExpectUpdateResource(k8sClient, namespaceObj)

			// now update the accessrule just to trigger the "Reconcile" loop -
			// important: we only do this for testing purposes to test a faulty project (.handleProject returns error);
			// we shouldn't actually watch namespaces in our controller since we don't expect this behaviour..
			accessRule = getAccessRuleFromClient(accessRuleName, k8sClient)
			accessRule.Labels = map[string]string{"some-label": "some-value"}
			utils.ExpectUpdateResource(k8sClient, accessRule)

			eventuallyValidateAccessRuleNotReady(accessRuleName, k8sClient,
				common.ScopeObjectFailedGettingNamespacesMessage)

			// validate the previously created rolebinding is now deleted
			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingObj := getRoleBindingFromClient(rbName, testProjNamespace, k8sClient)
			Expect(roleBindingObj).To(BeNil(),
				"RoleBinding <%s/%s> shouldn't exist in namespace of faulty project %s",
				testProjNamespace, rbName, testProjName)
		})
	})

	Describe("Access rule status NotReady", Ordered, func() {
		It("Role of AccessRule doesn't exist", func() {
			accessRuleName := generateAccessRuleName()
			accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
				SubjectId:   userA,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   testProjectA,
				ScopeType:   v1alpha1.ProjectScope,
				RoleName:    "un-existing-role",
			}, k8sClient)

			eventuallyValidateAccessRuleNotReady(accessRuleName, k8sClient,
				fmt.Sprintf(common.ClusterRoleNotFoundMessage,
					rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)))

			rbName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
			roleBindingObj := getRoleBindingFromClient(rbName, testNamespaceA, k8sClient)
			Expect(roleBindingObj).To(BeNil(),
				"RoleBinding <%s/%s> shouldn't exist in namespace of faulty project %s",
				testNamespaceA, rbName, testProjectA)
		})
	})
})

var _ = Describe("Access Rule Controller Tests - more", Ordered, func() {
	var (
		accessRule *v1alpha1.AccessRule
	)

	BeforeAll(func() {
		controllerSetup()

		createNamespaceIfNeeded(runaiv1.DefaultNamespace, "", k8sClient)

		createDepartment(testDepartmentA, k8sClient)
	})

	AfterAll(func() {
		afterAllCleanup()
	})

	AfterEach(func() {
		if accessRule != nil {
			deleteAccessRule(accessRule.Name, k8sClient)
			accessRule = nil
		}
	})

	Describe("No projects exist at all - validate runai Cluster/RoleBindings exist", Ordered, func() {
		var (
			proj1Name      string
			proj1Namespace string
		)

		BeforeAll(func() {
			proj1Namespace = generateNamespaceName()
			proj1Name = generateProjectName()
			testNamespacesToProject[proj1Namespace] = proj1Name
		})

		AfterEach(func() {
			deleteProject(proj1Name, k8sClient)
			deleteNamespace(proj1Namespace, k8sClient)
		})

		Describe("Expected runai cluster/RoleBindings exist", func() {
			BeforeAll(func() {
				for _, testClusterRole := range testClusterRoles {
					createNeededClusterRolesAndBindingsForTestRole(testClusterRole)
				}
			})

			AfterAll(func() {
				for _, testClusterRole := range testClusterRoles {
					deleteCreatedClusterRolesAndBindingsForTestRole(testClusterRole)
				}
			})

			It("Department-scoped access role with no projects + project created", func() {
				accessRuleName := generateAccessRuleName()
				accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
					SubjectId:   groupA,
					SubjectType: v1alpha1.GroupSubject,
					ScopeName:   testDepartmentA,
					ScopeType:   v1alpha1.DepartmentScope,
					RoleName:    testClusterRoleB,
				}, k8sClient)

				// subject should be added only to runai Cluster/RoleBinding, cause no projects exist
				eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{}, k8sClient)

				clusterRoleBindingName := rolebindingsutils.GetClusterRoleBindingName(accessRule.Spec.RoleName)
				clusterRoleBinding := getClusterRoleBindingFromClient(clusterRoleBindingName, k8sClient)
				Expect(clusterRoleBinding.Subjects).To(ContainElement(getExpectedSubjectForAccessRule(accessRule)),
					"expected cluster role binding's <%s> Subjects to contain expected <%v>",
					clusterRoleBindingName, getExpectedSubjectForAccessRule(accessRule))

				validateAccessRuleReady(accessRuleName, k8sClient)

				// now create the project
				createNamespaceIfNeeded(proj1Namespace, proj1Name, k8sClient)
				createProject(proj1Name, testDepartmentA, proj1Namespace, k8sClient)

				eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{proj1Namespace}, k8sClient)
				validateAccessRuleReady(accessRuleName, k8sClient)
			})
			It("Cluster-scoped access role with no projects + project created", func() {
				accessRuleName := generateAccessRuleName()
				accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
					SubjectId:   userA,
					SubjectType: v1alpha1.UserSubject,
					ScopeType:   v1alpha1.ClusterScope,
					RoleName:    testClusterRoleA,
				}, k8sClient)

				// subject should be added only to runai Cluster/RoleBinding, cause no projects exist
				eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{}, k8sClient)

				runaiNsRoleBindingName := rolebindingsutils.GetRunaiNamespaceRoleBindingName(accessRule.Spec.RoleName)
				roleBinding := getRoleBindingFromClient(runaiNsRoleBindingName, "runai", k8sClient)
				Expect(roleBinding.Subjects).To(ContainElement(getExpectedSubjectForAccessRule(accessRule)),
					"expected role binding's <%s/%s> Subjects to contain expected <%v>",
					"runai", runaiNsRoleBindingName, getExpectedSubjectForAccessRule(accessRule))

				validateAccessRuleReady(accessRuleName, k8sClient)

				// now create the project
				createNamespaceIfNeeded(proj1Namespace, proj1Name, k8sClient)
				createProject(proj1Name, testDepartmentA, proj1Namespace, k8sClient)

				eventuallyValidateRoleBindingsForAccessRule(accessRule, []string{proj1Namespace}, k8sClient)
				validateAccessRuleReady(accessRuleName, k8sClient)
			})
		})

		Describe("Expected runai cluster/RoleBindings don't exist", func() {
			AfterEach(func() {
				if accessRule != nil {
					deleteAccessRule(accessRule.Name, k8sClient)
					accessRule = nil
				}

				for _, testClusterRole := range testClusterRoles {
					deleteCreatedClusterRolesAndBindingsForTestRole(testClusterRole)
				}
			})
			Describe("cluster-ClusterRole doesn't exist", func() {
				BeforeEach(func() {
					createNeededClusterRolesAndBindingsForTestRoleInner(testClusterRoleA,
						false, true,
						true,
						true, true)
				})
				It("Cluster-scoped access role - failed handling", func() {
					accessRuleName := generateAccessRuleName()
					accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
						SubjectId:   groupA,
						SubjectType: v1alpha1.GroupSubject,
						ScopeType:   v1alpha1.ClusterScope,
						RoleName:    testClusterRoleA,
					}, k8sClient)

					eventuallyValidateAccessRuleNotReady(accessRuleName, k8sClient,
						fmt.Sprintf(common.ClusterRoleNotFoundMessage,
							rolebindingsutils.GetClusterRoleBindingName(accessRule.Spec.RoleName)))
				})
			})
			Describe("cluster-ClusterRoleBinding doesn't exist", func() {
				BeforeEach(func() {
					createNeededClusterRolesAndBindingsForTestRoleInner(testClusterRoleA,
						true, false,
						true,
						true, true)
				})
				It("Cluster-scoped access role - failed handling", func() {
					accessRuleName := generateAccessRuleName()
					accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
						SubjectId:   groupA,
						SubjectType: v1alpha1.GroupSubject,
						ScopeType:   v1alpha1.ClusterScope,
						RoleName:    testClusterRoleA,
					}, k8sClient)

					eventuallyValidateAccessRuleNotReady(accessRuleName, k8sClient,
						common.FailedReconcileClusterRoleBindingMessage)
				})
			})
			Describe("namespaced-Role doesn't exist", func() {
				BeforeEach(func() {
					createNeededClusterRolesAndBindingsForTestRoleInner(testClusterRoleB,
						true, true,
						false,
						true, true)
				})
				It("cluster-scoped access role - failed handling", func() {
					accessRuleName := generateAccessRuleName()
					accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
						SubjectId:   userA,
						SubjectType: v1alpha1.UserSubject,
						ScopeType:   v1alpha1.ClusterScope,
						RoleName:    testClusterRoleB,
					}, k8sClient)

					eventuallyValidateAccessRuleNotReady(accessRuleName, k8sClient,
						fmt.Sprintf(common.ClusterRoleNotFoundMessage,
							rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)))
				})
			})
			Describe("runai-namespaced-Role doesn't exist", func() {
				BeforeEach(func() {
					createNeededClusterRolesAndBindingsForTestRoleInner(testClusterRoleB,
						true, true,
						true,
						false, true)
				})
				It("cluster-scoped access role - failed handling", func() {
					accessRuleName := generateAccessRuleName()
					accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
						SubjectId:   userA,
						SubjectType: v1alpha1.UserSubject,
						ScopeType:   v1alpha1.ClusterScope,
						RoleName:    testClusterRoleB,
					}, k8sClient)

					eventuallyValidateAccessRuleNotReady(accessRuleName, k8sClient,
						fmt.Sprintf(common.ClusterRoleNotFoundMessage,
							rolebindingsutils.GetRunaiNamespaceRoleBindingName(accessRule.Spec.RoleName)))
				})
			})
			Describe("runai-namespaced-RoleBinding doesn't exist", func() {
				BeforeEach(func() {
					createNeededClusterRolesAndBindingsForTestRoleInner(testClusterRoleA,
						true, true,
						true,
						true, false)
				})
				It("Cluster-scoped access role - failed handling", func() {
					accessRuleName := generateAccessRuleName()
					accessRule = createAccessRule(accessRuleName, &v1alpha1.AccessRuleSpec{
						SubjectId:   groupA,
						SubjectType: v1alpha1.GroupSubject,
						ScopeType:   v1alpha1.ClusterScope,
						RoleName:    testClusterRoleA,
					}, k8sClient)

					eventuallyValidateAccessRuleNotReady(accessRuleName, k8sClient,
						common.FailedReconcileRunaiMessage)
				})
			})
		})
	})
})

func afterAllCleanup() {
	for testProject := range testProjectsToNamespace {
		deleteProject(testProject, k8sClient)
	}
	for _, testClusterRole := range testClusterRoles {
		deleteCreatedClusterRolesAndBindingsForTestRole(testClusterRole)
	}
	for _, testDepartment := range testDepartments {
		deleteDepartment(testDepartment, k8sClient)
	}

	accessRules := listAccessRules(k8sClient)
	for _, accessRule := range accessRules.Items {
		deleteAccessRule(accessRule.Name, k8sClient)
	}

	for _, testNamespace := range testProjectsToNamespace {
		deleteNamespace(testNamespace, k8sClient)
	}
	deleteNamespace(runaiv1.DefaultNamespace, k8sClient)
}

func createNeededClusterRolesAndBindingsForTestRole(roleName string) {
	createNeededClusterRolesAndBindingsForTestRoleInner(roleName,
		true, true,
		true,
		true, true)
}

func createNeededClusterRolesAndBindingsForTestRoleInner(roleName string,
	createClusterClusterRole bool,
	createClusterClusterRoleBinding bool,
	createNamespacedRole bool,
	createRunaiNamespaceRole bool,
	createRunaiNamespaceRoleBinding bool,
) {
	clusterRoleName := rolebindingsutils.GetClusterRoleBindingName(roleName)
	if createClusterClusterRole {
		createClusterRole(clusterRoleName, k8sClient)
	}
	if createClusterClusterRoleBinding {
		createClusterRoleBinding(clusterRoleName, k8sClient)
	}

	clusterRoleName = rolebindingsutils.GetNamespacedRoleBindingName(roleName)
	if createNamespacedRole {
		createClusterRole(clusterRoleName, k8sClient)
	}

	clusterRoleName = rolebindingsutils.GetRunaiNamespaceRoleBindingName(roleName)
	if createRunaiNamespaceRole {
		createClusterRole(clusterRoleName, k8sClient)
	}
	if createRunaiNamespaceRoleBinding {
		createRoleBinding(clusterRoleName, "runai", k8sClient)
	}

	if _, found := testClusterRolesWithClusterScopedRunaiNamespaceRoleBinding[roleName]; found {
		clusterRoleName = rolebindingsutils.GetClusterScopedRunaiNamespaceRoleBindingName(roleName)
		createClusterRole(clusterRoleName, k8sClient)
		createRoleBinding(clusterRoleName, "runai", k8sClient)
	}
}

func deleteCreatedClusterRolesAndBindingsForTestRole(roleName string) {
	clusterRoleName := rolebindingsutils.GetClusterRoleBindingName(roleName)
	deleteClusterRoleBinding(clusterRoleName, k8sClient)
	deleteClusterRole(clusterRoleName, k8sClient)

	clusterRoleName = rolebindingsutils.GetNamespacedRoleBindingName(roleName)
	deleteClusterRole(clusterRoleName, k8sClient)

	clusterRoleName = rolebindingsutils.GetRunaiNamespaceRoleBindingName(roleName)
	deleteRoleBinding(clusterRoleName, "runai", k8sClient)
	deleteClusterRole(clusterRoleName, k8sClient)

	if _, found := testClusterRolesWithClusterScopedRunaiNamespaceRoleBinding[roleName]; found {
		clusterRoleName = rolebindingsutils.GetClusterScopedRunaiNamespaceRoleBindingName(roleName)
		deleteRoleBinding(clusterRoleName, "runai", k8sClient)
		deleteClusterRole(clusterRoleName, k8sClient)
	}
}

func nsName(name, namespace string) types.NamespacedName {
	return types.NamespacedName{Name: name, Namespace: namespace}
}
