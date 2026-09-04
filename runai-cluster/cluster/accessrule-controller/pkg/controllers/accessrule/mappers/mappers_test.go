/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mappers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMappers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AccessRule Mappers Unit Tests")
}

const (
	depA     = "dep-a"
	depB     = "dep-b"
	roleName = "role-a"
)

func accessRule(name string, scopeType v1alpha1.ScopeType, scopeName string) *v1alpha1.AccessRule {
	return &v1alpha1.AccessRule{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1alpha1.AccessRuleSpec{
			ScopeType: scopeType,
			ScopeName: scopeName,
			RoleName:  roleName,
		},
	}
}

func accessRuleNames(rules []v1alpha1.AccessRule) []string {
	names := make([]string, 0, len(rules))
	for _, rule := range rules {
		names = append(names, rule.Name)
	}
	return names
}
