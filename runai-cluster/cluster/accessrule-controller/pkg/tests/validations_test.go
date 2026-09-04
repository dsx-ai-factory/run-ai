/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/rolebindingsutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// projectOwnsRoleBinding checks that the role binding is owned by the kai.resources project
// (the owner reference carries that CRD's GVK).
func projectOwnsRoleBinding(projectName string, roleBinding *rbacv1.RoleBinding, k8sClient client.Client) bool {
	kaiProject := getKaiProjectFromClient(projectName, k8sClient)
	if kaiProject == nil {
		return false
	}
	for _, ref := range roleBinding.OwnerReferences {
		if ref.APIVersion == kaiorg.GroupVersion.Identifier() && ref.Kind == "Project" &&
			ref.Name == kaiProject.Name && ref.UID == kaiProject.UID {
			return true
		}
	}
	return false
}

func validateRoleBindingAccessRuleSubject(g Gomega, roleBindingName, roleBindingNamespace string,
	accessRule *v1alpha1.AccessRule, k8sClient client.Client) {
	roleBinding := getRoleBindingFromClient(roleBindingName, roleBindingNamespace, k8sClient)
	g.ExpectWithOffset(2, roleBinding).ToNot(BeNil(),
		"expected role binding <%s/%s> to exist", roleBindingName, roleBindingNamespace)

	g.ExpectWithOffset(2, roleBinding.Subjects).ToNot(BeEmpty(),
		"expected role binding's <%s/%s> subjects to not be empty",
		roleBindingName, roleBindingNamespace)
	g.ExpectWithOffset(2, roleBinding.Subjects).To(ContainElement(getExpectedSubjectForAccessRule(accessRule)),
		"expected role binding's <%s/%s> Subjects to contain expected <%v>",
		roleBindingName, roleBindingNamespace, getExpectedSubjectForAccessRule(accessRule))

	g.ExpectWithOffset(2, roleBinding.RoleRef).To(Equal(getExpectedRoleRef(roleBindingName)),
		"expected role binding's <%s/%s> RoleRef to equal expected <%v>",
		roleBindingName, roleBindingNamespace, getExpectedRoleRef(roleBindingName))

	if roleBindingNamespace != "runai" {
		projectName, found := testNamespacesToProject[roleBindingNamespace]
		g.Expect(found).To(BeTrue(), "expected to have project name for namespace: <%s>", roleBindingNamespace)

		g.ExpectWithOffset(2, roleBinding.OwnerReferences).ToNot(BeEmpty(),
			"expected role binding's <%s/%s> OwnerReferences list to not be empty",
			roleBindingName, roleBindingNamespace)
		g.ExpectWithOffset(2, projectOwnsRoleBinding(projectName, roleBinding, k8sClient)).To(BeTrue(),
			"expected role binding <%s/%s> to have project <%s> as owner",
			roleBindingName, roleBindingNamespace, accessRule.Name)
	}

	g.ExpectWithOffset(2, roleBinding.Labels).ToNot(BeEmpty(),
		"expected role binding's <%s/%s> labels to not be empty",
		roleBindingName, roleBindingNamespace)
	expectedLabel := rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, roleBindingNamespace != "runai")
	g.ExpectWithOffset(2, roleBinding.Labels[expectedLabel]).To(Equal("true"),
		"expected role binding <%s/%s> to have label <%s>=<true>",
		roleBindingName, roleBindingNamespace, expectedLabel)
}

func expectRoleBindingToNotHaveAccessRulesSubject(g Gomega, roleBindingName, roleBindingNamespace string,
	accessRule *v1alpha1.AccessRule, k8sClient client.Client) {
	roleBinding := getRoleBindingFromClient(roleBindingName, roleBindingNamespace, k8sClient)
	if roleBinding == nil {
		return
	}

	g.ExpectWithOffset(2, roleBinding.Subjects).ToNot(ContainElement(getExpectedSubjectForAccessRule(accessRule)),
		"expected role binding's <%s/%s> Subjects to not contain expected <%v>",
		roleBindingName, roleBindingNamespace, getExpectedSubjectForAccessRule(accessRule))

	expectedLabel := rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, roleBindingNamespace != "runai")
	g.ExpectWithOffset(2, roleBinding.Labels).ToNot(HaveKey(expectedLabel),
		"expected role binding <%s/%s> to not have label key <%s>",
		roleBindingName, roleBindingNamespace, expectedLabel)
}

func validateClusterRoleBindingAccessRuleSubject(g Gomega, clusterRoleBindingName string,
	accessRule *v1alpha1.AccessRule, k8sClient client.Client) {
	clusterRoleBinding := getClusterRoleBindingFromClient(clusterRoleBindingName, k8sClient)
	g.ExpectWithOffset(2, clusterRoleBinding).ToNot(BeNil(),
		"expected cluster role binding <%s> to exist", clusterRoleBindingName)

	g.ExpectWithOffset(2, clusterRoleBinding.Subjects).ToNot(BeEmpty(),
		"expected cluster role binding's <%s> subjects to not be empty",
		clusterRoleBindingName)
	g.ExpectWithOffset(2, clusterRoleBinding.Subjects).To(ContainElement(getExpectedSubjectForAccessRule(accessRule)),
		"expected cluster role binding's <%s> Subjects to contain expected <%v>",
		clusterRoleBindingName, getExpectedSubjectForAccessRule(accessRule))

	g.ExpectWithOffset(2, clusterRoleBinding.RoleRef).To(Equal(getExpectedRoleRef(clusterRoleBindingName)),
		"expected cluster role binding's <%s> RoleRef to equal expected <%v>",
		clusterRoleBindingName, getExpectedRoleRef(clusterRoleBindingName))

	g.ExpectWithOffset(2, clusterRoleBinding.Labels).ToNot(BeEmpty(),
		"expected cluster role binding's <%s> labels to not be empty",
		clusterRoleBindingName)
	expectedLabel := rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, false)
	g.ExpectWithOffset(2, clusterRoleBinding.Labels[expectedLabel]).To(Equal("true"),
		"expected cluster role binding <%s> to have label <%s>=<true>",
		clusterRoleBindingName, expectedLabel)
}

func expectClusterRoleBindingToNotHaveAccessRulesSubject(g Gomega, clusterRoleBindingName string,
	accessRule *v1alpha1.AccessRule, k8sClient client.Client) {
	clusterRoleBinding := getClusterRoleBindingFromClient(clusterRoleBindingName, k8sClient)
	if clusterRoleBinding == nil {
		return
	}

	g.ExpectWithOffset(2, clusterRoleBinding.Subjects).ToNot(ContainElement(getExpectedSubjectForAccessRule(accessRule)),
		"expected cluster role binding's <%s> Subjects to not contain expected <%v>",
		clusterRoleBindingName, getExpectedSubjectForAccessRule(accessRule))

	expectedLabel := rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, false)
	g.ExpectWithOffset(2, clusterRoleBinding.Labels).ToNot(HaveKey(expectedLabel),
		"expected clsuter role binding <%s> to not have label key <%s>",
		clusterRoleBindingName, expectedLabel)
}

func validateRoleBindingsForAccessRule(g Gomega,
	accessRule *v1alpha1.AccessRule,
	namespaces []string,
	k8sClient client.Client) {
	for _, namespace := range namespaces {
		roleBindingName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
		validateRoleBindingAccessRuleSubject(g, roleBindingName, namespace, accessRule, k8sClient)
	}

	runaiRoleBindingName := rolebindingsutils.GetRunaiNamespaceRoleBindingName(accessRule.Spec.RoleName)
	validateRoleBindingAccessRuleSubject(g, runaiRoleBindingName, "runai", accessRule, k8sClient)

	clusterRoleBindingName := rolebindingsutils.GetClusterRoleBindingName(accessRule.Spec.RoleName)
	validateClusterRoleBindingAccessRuleSubject(g, clusterRoleBindingName, accessRule, k8sClient)

	_, found := testClusterRolesWithClusterScopedRunaiNamespaceRoleBinding[accessRule.Spec.RoleName]
	if found {
		runaiClusterScopedRoleBindingName := rolebindingsutils.GetClusterScopedRunaiNamespaceRoleBindingName(accessRule.Spec.RoleName)
		if accessRule.Spec.ScopeType == v1alpha1.ClusterScope {
			validateRoleBindingAccessRuleSubject(g, runaiClusterScopedRoleBindingName, "runai", accessRule, k8sClient)
		} else {
			expectRoleBindingToNotHaveAccessRulesSubject(g, runaiClusterScopedRoleBindingName, "runai", accessRule, k8sClient)
		}
	}
}

func validateRoleBindingsForAccessRuleDontHaveAccessRuleSubject(g Gomega,
	accessRule *v1alpha1.AccessRule,
	namespaces []string,
	validateClusterScopedAsWell bool,
	k8sClient client.Client) {
	for _, namespace := range namespaces {
		roleBindingName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)
		expectRoleBindingToNotHaveAccessRulesSubject(g, roleBindingName, namespace, accessRule, k8sClient)
	}

	if !validateClusterScopedAsWell {
		return
	}

	runaiRoleBindingName := rolebindingsutils.GetRunaiNamespaceRoleBindingName(accessRule.Spec.RoleName)
	expectRoleBindingToNotHaveAccessRulesSubject(g, runaiRoleBindingName, "runai", accessRule, k8sClient)

	clusterRoleBindingName := rolebindingsutils.GetClusterRoleBindingName(accessRule.Spec.RoleName)
	expectClusterRoleBindingToNotHaveAccessRulesSubject(g, clusterRoleBindingName, accessRule, k8sClient)

	_, found := testClusterRolesWithClusterScopedRunaiNamespaceRoleBinding[accessRule.Spec.RoleName]
	if found {
		runaiClusterScopedRoleBindingName := rolebindingsutils.GetClusterScopedRunaiNamespaceRoleBindingName(accessRule.Spec.RoleName)
		expectRoleBindingToNotHaveAccessRulesSubject(g, runaiClusterScopedRoleBindingName, "runai", accessRule, k8sClient)
	}
}

// The helpers below reconcile all access rules (the manual stand-in for the
// manager's watch fan-out) and then assert synchronously. The reconcile happens
// once, outside the assertion, so the deterministic fake-client state is checked
// directly rather than polled.

func eventuallyValidateRoleBindingsForAccessRule(accessRule *v1alpha1.AccessRule,
	namespaces []string,
	k8sClient client.Client) {
	reconcileAllAccessRules()
	validateRoleBindingsForAccessRule(Default, accessRule, namespaces, k8sClient)
}

func consistentlyValidateRoleBinding(accessRule *v1alpha1.AccessRule,
	namespaces []string,
	k8sClient client.Client) {
	// Reconcile once more and assert the bindings are unchanged: with no manager
	// there are no spurious events, so a single idempotent sweep is a faithful,
	// deterministic substitute for the original Consistently poll.
	reconcileAllAccessRules()
	validateRoleBindingsForAccessRule(Default, accessRule, namespaces, k8sClient)
}

func eventuallyValidateRoleBindingsInNamespacesDontHaveAccessRuleSubject(accessRule *v1alpha1.AccessRule,
	namespaces []string,
	k8sClient client.Client) {
	reconcileAllAccessRules()
	validateRoleBindingsForAccessRuleDontHaveAccessRuleSubject(Default, accessRule, namespaces, false, k8sClient)
}

func eventuallyValidateAllRoleBindingsDontHaveAccessRuleSubject(accessRule *v1alpha1.AccessRule,
	namespaces []string,
	k8sClient client.Client) {
	reconcileAllAccessRules()
	validateRoleBindingsForAccessRuleDontHaveAccessRuleSubject(Default, accessRule, namespaces, true, k8sClient)
}

func eventuallyExpectRoleBindingDeleted(roleBinding types.NamespacedName, msg string, k8sClient client.Client) {
	reconcileAllAccessRules()
	By(fmt.Sprintf("Checking RoleBinding <%s/%s> is deleted", roleBinding.Namespace, roleBinding.Name))
	roleBindingObj := getRoleBindingFromClient(roleBinding.Name, roleBinding.Namespace, k8sClient)
	ExpectWithOffset(1, roleBindingObj).To(BeNil(),
		msg,
		roleBinding.Namespace, roleBinding.Name)
}

func validateAccessRuleStatus(accessRuleName string, expectedPhase v1alpha1.AccessRulePhase,
	statusMessage string, k8sClient client.Client) {
	reconcileAllAccessRules()
	accessRule := getAccessRuleFromClient(accessRuleName, k8sClient)
	ExpectWithOffset(1, accessRule).ToNot(BeNil(), "expected access rule <%s> to exist", accessRuleName)
	ExpectWithOffset(1, accessRule.Status.Message).To(Equal(statusMessage))
	ExpectWithOffset(1, accessRule.Status.Phase).To(Equal(expectedPhase))
}

func validateAccessRuleReady(accessRuleName string, k8sClient client.Client) {
	validateAccessRuleStatus(accessRuleName, v1alpha1.AccessRuleReady, "", k8sClient)
}

func validateAccessRuleNotReady(accessRuleName string, statusMessage string, k8sClient client.Client) {
	validateAccessRuleStatus(accessRuleName, v1alpha1.AccessRuleNotReady, statusMessage, k8sClient)
}

func eventuallyValidateAccessRuleNotReady(accessRuleName string, k8sClient client.Client, statusMessageOptions ...string) {
	reconcileAllAccessRules()
	accessRule := getAccessRuleFromClient(accessRuleName, k8sClient)
	ExpectWithOffset(1, accessRule).ToNot(BeNil())
	ExpectWithOffset(1, accessRule.Status.Phase).To(Equal(v1alpha1.AccessRuleNotReady))
	ExpectWithOffset(1, accessRule.Status.Message).To(BeElementOf(statusMessageOptions))
}
