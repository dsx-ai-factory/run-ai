/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package accessrule

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (c *AccessRuleController) MapProjectToAccessRuleEvent(_ context.Context, object client.Object) []reconcile.Request {
	project, ok := object.(*kaiorg.Project)
	if !ok {
		log.Panic().Msgf("Cannot convert object to *kaiorg.Project: %v", object)
		return []reconcile.Request{}
	}

	accessRules, err := c.accessRulesProjectsMapper.GetAllAccessRulesForProject(context.Background(), project)
	if err != nil {
		log.Error().Msgf("Failed getting all access rules to reconcile for kai project <%s>, error: %s",
			project.Name, err.Error())
		return []reconcile.Request{}
	}

	requests := []reconcile.Request{}
	for _, accessRule := range accessRules {
		log.Debug().Msgf("For kai project <%s>, triggering reconcile of access rule <%s>",
			project.Name, accessRule.Name)

		accessRuleNamespacedName := types.NamespacedName{
			Name: accessRule.Name,
		}
		requests = append(requests, reconcile.Request{
			NamespacedName: accessRuleNamespacedName})
	}
	return requests
}

func (c *AccessRuleController) MapRoleBindingToAccessRuleEvent(_ context.Context, object client.Object) []reconcile.Request {
	roleBinding, ok := object.(*rbacv1.RoleBinding)
	if !ok {
		log.Panic().Msgf("Cannot convert object to *rbacv1.RoleBinding: %v", object)
		return []reconcile.Request{}
	}

	if roleBinding.Labels == nil {
		return []reconcile.Request{}
	}

	return getAccessRuleRequestsFromRoleBindingLabels("RoleBinding",
		fmt.Sprintf("%s/%s", roleBinding.Namespace, roleBinding.Name),
		roleBinding.Labels,
		common.AccessRuleAddedSubjectToProjectsRBPrefix, common.AccessRuleAddedSubjectPrefix)
}

func (c *AccessRuleController) MapClusterRoleBindingToAccessRuleEvent(_ context.Context, object client.Object) []reconcile.Request {
	clusterRoleBinding, ok := object.(*rbacv1.ClusterRoleBinding)
	if !ok {
		log.Panic().Msgf("Cannot convert object to *rbacv1.ClusterRoleBinding: %v", object)
		return []reconcile.Request{}
	}

	if clusterRoleBinding.Labels == nil {
		return []reconcile.Request{}
	}

	return getAccessRuleRequestsFromRoleBindingLabels("ClusterRoleBinding",
		clusterRoleBinding.Name,
		clusterRoleBinding.Labels,
		common.AccessRuleAddedSubjectPrefix)
}

func getAccessRuleRequestsFromRoleBindingLabels(roleBindingType string, roleBindingName string,
	labels map[string]string, prefixes ...string) []reconcile.Request {
	requests := []reconcile.Request{}
	for labelKey := range labels {
		existingPrefix := ""
		for _, prefix := range prefixes {
			if strings.HasPrefix(labelKey, prefix) {
				existingPrefix = prefix
				break
			}
		}
		if existingPrefix == "" {
			continue
		}

		accessRuleName := labelKey[len(existingPrefix):]
		log.Debug().Msgf("For role %s <%s>, triggering reconcile of access rule <%s>",
			roleBindingType, roleBindingName, accessRuleName)

		accessRuleNamespacedName := types.NamespacedName{
			Name: accessRuleName,
		}
		requests = append(requests, reconcile.Request{
			NamespacedName: accessRuleNamespacedName})
	}
	return requests
}

func ControllerPredicateFuncs() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return filterUpdateEvents(e.ObjectOld, e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return true
		},
	}
}

func filterUpdateEvents(objectOld, objectNew client.Object) bool {
	switch objectNew.(type) {
	case *kaiorg.Project:
		return filterProjectUpdatesForController(objectOld.(*kaiorg.Project), objectNew.(*kaiorg.Project))
	case *v1alpha1.AccessRule:
		return objectNew.GetResourceVersion() != objectOld.GetResourceVersion()
	case *rbacv1.RoleBinding:
		return true
	}
	return false
}

func filterProjectUpdatesForController(projectOld, projectNew *kaiorg.Project) bool {
	// we are only interested in Reconcile if the project has changed departments or namespace
	if projectOld.Spec.Parent != projectNew.Spec.Parent {
		return true
	}
	return projectOld.Status.Namespace != projectNew.Status.Namespace
}
