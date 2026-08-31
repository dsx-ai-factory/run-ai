/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	timeout  = time.Second * 12
	interval = time.Millisecond * 500

	accessRuleNamePrefix = "ar-"
	projectNamePrefix    = "pr-"
	namespaceNamePrefix  = "ns-"
)

func generateProjectName() string {
	return fmt.Sprintf("%s%s", projectNamePrefix, rand.String(5))
}

func generateNamespaceName() string {
	return fmt.Sprintf("%s%s", namespaceNamePrefix, rand.String(5))
}

func generateAccessRuleName() string {
	return fmt.Sprintf("%s%s", accessRuleNamePrefix, rand.String(5))
}

func getExpectedSubjectForAccessRule(accessRule *v1alpha1.AccessRule) rbacv1.Subject {
	subject := rbacv1.Subject{
		APIGroup: rbacGroup,
		Name:     accessRule.Spec.SubjectId,
	}

	if accessRule.Spec.SubjectType == v1alpha1.UserSubject ||
		accessRule.Spec.SubjectType == v1alpha1.AppSubject ||
		accessRule.Spec.SubjectType == v1alpha1.ServiceAccountSubject {
		subject.Kind = "User"
	} else if accessRule.Spec.SubjectType == v1alpha1.GroupSubject {
		// we only support types - so if not user, app, or service-account, it's a group
		subject.Kind = "Group"
	}

	return subject
}

func getExpectedRoleRef(clusterRoleName string) rbacv1.RoleRef {
	return rbacv1.RoleRef{
		APIGroup: rbacGroup,
		Kind:     "ClusterRole",
		Name:     clusterRoleName,
	}
}
