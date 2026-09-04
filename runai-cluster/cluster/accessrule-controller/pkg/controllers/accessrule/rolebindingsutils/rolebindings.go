/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rolebindingsutils

import (
	"fmt"
	"strings"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	RbacGroup       = "rbac.authorization.k8s.io"
	ClusterRoleKind = "ClusterRole"
	UserKind        = "User"
	GroupKind       = "Group"

	AuthenticatedUsersGroup = "system:authenticated"

	allRolesRunaiPrefix                       = "runai"
	namespacedRoleNameSuffix                  = "ns"
	namespaceRunaiRoleNameSuffix              = "runai"
	clusterScopedNamespaceRunaiRoleNameSuffix = "runai-cs"
	labelSuffixIfNeeded                       = "a"
)

var (
	TrueRef          = true
	RoleNamesMapping = map[string]string{
		"application-administrator": "editor",
		"cloud-support":             "viewer",
	}
)

func getRoleBindingObject(
	roleBindingName, namespace, roleName string,
	subjects []rbacv1.Subject,
	ownerReferences []metav1.OwnerReference,
	labels map[string]string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            roleBindingName,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerReferences,
		},
		Subjects: subjects,
		RoleRef: rbacv1.RoleRef{
			APIGroup: RbacGroup,
			Kind:     ClusterRoleKind,
			Name:     roleName,
		},
	}
}

// GetRoleBindingObjectForAccessRuleInProject builds the RoleBinding for an access rule in a project.
// The owner reference carries the kai.resources Project GVK so garbage collection tracks the real
// object.
func GetRoleBindingObjectForAccessRuleInProject(
	roleBindingName, namespace string,
	accessRule *v1alpha1.AccessRule,
	project *kaiorg.Project,
	labels map[string]string) *rbacv1.RoleBinding {
	return getRoleBindingObject(
		roleBindingName, namespace, roleBindingName,
		getSubjectsForAccessRule(accessRule),
		getProjectOwnerReferences(project),
		labels)
}

func getProjectOwnerReferences(project *kaiorg.Project) []metav1.OwnerReference {
	return []metav1.OwnerReference{{
		APIVersion: kaiorg.GroupVersion.Identifier(),
		Kind:       "Project",
		Name:       project.Name,
		UID:        project.UID,
		Controller: &TrueRef,
	}}
}

func getSubjectsForAccessRule(accessRule *v1alpha1.AccessRule) []rbacv1.Subject {
	return []rbacv1.Subject{GetSubjectForAccessRule(accessRule)}
}

func GetSubjectForAccessRule(accessRule *v1alpha1.AccessRule) rbacv1.Subject {
	subject := rbacv1.Subject{
		APIGroup: RbacGroup,
		Name:     subjectNameForAccessRule(accessRule),
		Kind:     GetSubjectKind(accessRule.Spec.SubjectType),
	}

	return subject
}

// subjectNameForAccessRule resolves the RBAC subject Name. The all-authenticated-users type
// carries no subjectId and binds to the built-in system:authenticated group; every other type
// uses its subjectId verbatim.
func subjectNameForAccessRule(accessRule *v1alpha1.AccessRule) string {
	if accessRule.Spec.SubjectType == v1alpha1.AllAuthenticatedUsersSubject {
		return AuthenticatedUsersGroup
	}
	return accessRule.Spec.SubjectId
}

func GetSubjectKind(subjectType v1alpha1.AccessRuleSubjectType) string {
	if subjectType == v1alpha1.UserSubject ||
		subjectType == v1alpha1.AppSubject ||
		subjectType == v1alpha1.ServiceAccountSubject {
		return UserKind
	}
	// group and all-authenticated-users both bind as a Kubernetes Group
	return GroupKind
}

func GetClusterRoleBindingName(roleName string) string {
	result := strings.ReplaceAll(roleName, " ", "-")
	result = strings.ToLower(result)
	result = getRoleNameForRoleBinding(result)
	return fmt.Sprintf("%s-%s", allRolesRunaiPrefix, result)
}

func GetNamespacedRoleBindingName(roleName string) string {
	result := GetClusterRoleBindingName(roleName)
	return fmt.Sprintf("%s-%s", result, namespacedRoleNameSuffix)
}

func GetRunaiNamespaceRoleBindingName(roleName string) string {
	result := GetClusterRoleBindingName(roleName)
	return fmt.Sprintf("%s-%s", result, namespaceRunaiRoleNameSuffix)
}

func GetClusterScopedRunaiNamespaceRoleBindingName(roleName string) string {
	result := GetClusterRoleBindingName(roleName)
	return fmt.Sprintf("%s-%s", result, clusterScopedNamespaceRunaiRoleNameSuffix)
}

// GetAccessRuleAddedSubjectLabelKey returns the label key that the affected RoleBindings should be labeled with.
// the label is different for projects' RoleBindings and non-projects'.
func GetAccessRuleAddedSubjectLabelKey(accessRule *v1alpha1.AccessRule, projectRoleBinding bool) string {
	var result string

	if projectRoleBinding {
		result = fmt.Sprintf("%s%s", common.AccessRuleAddedSubjectToProjectsRBPrefix, accessRule.Name)
	} else {
		result = fmt.Sprintf("%s%s", common.AccessRuleAddedSubjectPrefix, accessRule.Name)
	}

	if len(result) > validation.LabelValueMaxLength-1 {
		result = result[:validation.LabelValueMaxLength-1]
	}
	if result[len(result)-1] == '-' {
		result = fmt.Sprintf("%s%s", result, labelSuffixIfNeeded)
	}
	return result
}

func getRoleNameForRoleBinding(roleName string) string {
	if mappedToName, exists := RoleNamesMapping[roleName]; exists {
		return mappedToName
	}

	return roleName
}
