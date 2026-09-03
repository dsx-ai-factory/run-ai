/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/gomega"

	auth_openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateAccessRuleExistsEventually(accessRuleName string, k8sClient client.Client) {
	var ar *v1alpha1.AccessRule = nil
	EventuallyWithOffset(1, func() *v1alpha1.AccessRule {
		ar = getAccessRuleFromClient(accessRuleName, k8sClient)
		return ar
	}, DefaultTimeout).ShouldNot(BeNil())
}

func validateAccessRuleDeletedEventually(accessRuleName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func() *v1alpha1.AccessRule {
		return getAccessRuleFromClient(accessRuleName, k8sClient)
	}, DefaultTimeout).Should(BeNil())
}

func validateExistingAccessRuleMatchesBackendSpec(accessRuleName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingAccessRuleWithBackendSpec(g, accessRuleName, k8sClient)
	}, 0).Should(Succeed())
}

func validateExistingAccessRuleDoesNotMatchBackendSpec(accessRuleName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingAccessRuleWithBackendSpec(g, accessRuleName, k8sClient)
	}, 0).ShouldNot(Succeed())
}

func validateExistingAccessRuleWithBackendSpec(g Gomega, accessRuleName string, k8sClient client.Client) {
	errs := validation.IsDNS1123Label(accessRuleName)
	var expectedErrs []string
	g.ExpectWithOffset(1, errs).To(BeEquivalentTo(expectedErrs), "accessRuleName RFC validation")

	clusterAccessRule := getAccessRuleFromClient(accessRuleName, k8sClient)
	g.ExpectWithOffset(1, clusterAccessRule).ToNot(BeNil())

	backendARSpec := BackendAccessRulesMap[accessRuleName]
	g.ExpectWithOffset(1, backendARSpec).ToNot(BeNil(), "Backend accessrule spec should exist")

	g.ExpectWithOffset(1, clusterAccessRule.Labels[v1alpha1.AccessRuleIdLabel]).To(Equal(strconv.FormatInt(int64(backendARSpec.Id), 10)), "accessrule id")
	if backendARSpec.PredefinedRole.Get() != nil && backendARSpec.PredefinedRole.Get().Name != nil && backendARSpec.PredefinedRole.Get().Id != nil {
		g.ExpectWithOffset(1, clusterAccessRule.Spec.RoleName).To(Equal(convertRoleNameToK8sAllowed(*backendARSpec.PredefinedRole.Get().Name)), "RoleName")
	} else {
		g.ExpectWithOffset(1, clusterAccessRule.Spec.RoleName).To(Equal(convertRoleNameToK8sAllowed(backendARSpec.RoleName)), "RoleName")
	}
	g.ExpectWithOffset(1, clusterAccessRule.Spec.SubjectId).To(Equal(backendARSpec.SubjectId), "SubjectId")
	g.ExpectWithOffset(1, string(clusterAccessRule.Spec.SubjectType)).To(Equal(string(backendARSpec.SubjectType)), "SubjectType")
	g.ExpectWithOffset(1, clusterAccessRule.Spec.ScopeName).To(Equal(backendARSpec.ScopeName), "ScopeName")

	if backendARSpec.ScopeType == auth_openapi.SCOPETYPE_TENANT {
		g.ExpectWithOffset(1, string(clusterAccessRule.Spec.ScopeType)).To(Equal(string(auth_openapi.SCOPETYPE_CLUSTER)), "ScopeType")
	} else {
		g.ExpectWithOffset(1, string(clusterAccessRule.Spec.ScopeType)).To(Equal(string(backendARSpec.ScopeType)), "ScopeType")
	}
}

func getIntFromQuery(q url.Values, key string) int {
	valStr := q.Get(key)
	if valStr == "" {
		return 0
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0
	}
	return val
}

func _getBoolFromQuery(q url.Values, key string) bool { //nolint:unused
	valStr := q.Get(key)
	if valStr == "" {
		return false
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return false
	}
	return val
}

func getTimeFromQuery(q url.Values, key string) *time.Time {
	valStr := q.Get(key)
	if valStr == "" {
		return nil
	}

	val, err := time.Parse(time.RFC3339, valStr)
	if err != nil {
		return nil
	}
	return &val
}

func filterOutOffsetLimit(accessRules []auth_openapi.AccessRuleSync, offset, limit int) []auth_openapi.AccessRuleSync {
	accessRulesToReturn := []auth_openapi.AccessRuleSync{}

	end := len(accessRules)
	if limit > 0 {
		if offset+limit <= len(accessRules) {
			end = offset + limit
		}
	}
	if offset <= end {
		accessRulesToReturn = accessRules[offset:end]
	}

	return accessRulesToReturn
}

func filterOutLastUpdatedIncludeDeleted(accessRules []string, lastUpdated *time.Time) []auth_openapi.AccessRuleSync {
	result := []auth_openapi.AccessRuleSync{}
	for _, accessRuleName := range accessRules {
		accessRule := *BackendAccessRulesMap[accessRuleName]
		if (lastUpdated != nil && !accessRule.UpdatedAt.Before(*lastUpdated)) || lastUpdated == nil {
			// strictly for test purposes! to not return objects that are far in the future
			maxUpdatedAt := time.Now().Add(time.Hour)
			if accessRule.UpdatedAt.Before(maxUpdatedAt) {
				result = append(result, accessRule)
			}
		}
	}
	return result
}

func convertRoleNameToK8sAllowed(roleName string) string {
	result := strings.ReplaceAll(roleName, " ", "-")
	return strings.ToLower(result)
}
