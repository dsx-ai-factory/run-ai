/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package accessrule

import (
	"context"
	"fmt"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/mappers"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/rolebindings"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/rolebindingsutils"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AccessRuleReconciler reconciles an AccessRule object
type AccessRuleReconciler struct {
	cachedClient client.Client

	accessRulesProjectsMapper *mappers.AccessRulesProjectsMapper
	roleBindingsHandler       *rolebindings.RoleBindingsHandler
}

func NewAccessRuleReconciler(
	cachedClient client.Client,
	accessRulesProjectsMapper *mappers.AccessRulesProjectsMapper,
	roleBindingsHandler *rolebindings.RoleBindingsHandler) *AccessRuleReconciler {
	return &AccessRuleReconciler{
		cachedClient:              cachedClient,
		accessRulesProjectsMapper: accessRulesProjectsMapper,
		roleBindingsHandler:       roleBindingsHandler,
	}
}

// Reconcile preforms all the required actions to add the accessRule's subject to
// the required RoleBindings/ClusterRoleBindings.
func (r *AccessRuleReconciler) Reconcile(ctx context.Context, accessRule *v1alpha1.AccessRule) error {
	log.Ctx(ctx).Info().Msgf("Reconciling access rule <%s>", accessRule.Name)

	var err error
	statusMessages := []string{}

	innerStatusMessage, innerErr := r.reconcileAccessRulesSubjectInProjectsRoleBindings(ctx, accessRule)
	if innerErr == nil {
		innerStatusMessage, innerErr = r.roleBindingsHandler.HandleExtraRoleBindingsForAccessRule(ctx, accessRule)
	} else {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, innerStatusMessage)
		innerStatusMessage, innerErr = r.deleteSubjectFromAllRoleBindings(ctx, accessRule)
	}

	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, innerStatusMessage)
	}

	statusMessage := strings.Join(statusMessages, "; ")
	innerErr = r.UpdateStatus(ctx, accessRule, statusMessage)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	return err
}

// delete Unnecessary - delete accessrule's subject from RoleBindings that are not part of the scope anymore
// returns:
// - statusMessage string
// - err error
func (r *AccessRuleReconciler) deleteUnnecessaryProjectsRoleBindingsForAccessRule(ctx context.Context,
	accessRule *v1alpha1.AccessRule,
	reconciledNamespaces map[string]bool,
	affectedRoleBindings []rbacv1.RoleBinding,
) (string, error) {
	var err error
	failedNamespaces := []string{}
	for _, roleBinding := range affectedRoleBindings {
		_, found := reconciledNamespaces[roleBinding.Namespace]
		if found {
			continue
		}

		innerErr := r.roleBindingsHandler.DeleteFromExistingProjectRoleBinding(ctx, accessRule, &roleBinding)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
			failedNamespaces = append(failedNamespaces, roleBinding.Namespace)
		}
	}

	statusMessage := ""
	if len(failedNamespaces) > 0 {
		statusMessage = fmt.Sprintf(common.FailedDeletingUnnecessaryMessage,
			strings.Join(failedNamespaces, ", "))
	}
	return statusMessage, err
}

func (r *AccessRuleReconciler) reconcileAccessRulesSubjectInProjectsRoleBindings(ctx context.Context,
	accessRule *v1alpha1.AccessRule) (string, error) {
	namespacesProjectsMap, statusMessage, err := r.getProjectsNamespacesOfAccessRuleScope(ctx, accessRule)
	if err != nil {
		return statusMessage, err
	}

	roleBindings, err := r.getProjectsRoleBindingsWithAccessRulesSubject(ctx, accessRule)
	if err != nil {
		return common.ScopeObjectFailedGettingRoleBindingsMessage, err
	}

	statusMessages := []string{}

	reconciledNamespaces, innerStatusMessage, innerErr := r.handleProjectsRoleBindingsForAccessRule(ctx,
		accessRule, namespacesProjectsMap)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, innerStatusMessage)
	}

	innerStatusMessage, innerErr = r.deleteUnnecessaryProjectsRoleBindingsForAccessRule(ctx,
		accessRule, reconciledNamespaces, roleBindings)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, innerStatusMessage)
	}

	return strings.Join(statusMessages, "; "), err
}

func (r *AccessRuleReconciler) handleProjectsRoleBindingsForAccessRule(ctx context.Context,
	accessRule *v1alpha1.AccessRule,
	namespacesProjectsMap map[string]*kaiorg.Project,
) (map[string]bool, string, error) {
	namespacedRoleBindingName := rolebindingsutils.GetNamespacedRoleBindingName(accessRule.Spec.RoleName)

	statusMessage, err := r.roleBindingsHandler.ValidateClusterRoleExists(ctx,
		namespacedRoleBindingName)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed getting ClusterRole <%s> for access rule <%s>, error: %s",
			namespacedRoleBindingName, accessRule.Name, err.Error())
		return map[string]bool{}, statusMessage, err
	}

	reconciledNamespaces := map[string]bool{}
	failedProjects := []string{}
	for namespace, project := range namespacesProjectsMap {
		log.Ctx(ctx).Debug().Msgf("Handling kai project <%s> for access rule <%s>",
			project.Name, accessRule.Name)

		innerErr := r.roleBindingsHandler.HandleRoleBindingForProject(ctx, accessRule,
			namespacedRoleBindingName, namespace, project)
		if innerErr == nil {
			reconciledNamespaces[namespace] = true
		} else {
			log.Ctx(ctx).Error().Msgf("Failed handling kai project <%s> for access rule <%s>, error: %s",
				project.Name, accessRule.Name, innerErr.Error())
			err = multierror.Append(err, innerErr)
			failedProjects = append(failedProjects, project.Name)
		}
	}

	if len(failedProjects) > 0 {
		statusMessage = fmt.Sprintf(common.FailedReconcileProjectsMessage,
			strings.Join(failedProjects, ", "))
	}

	return reconciledNamespaces, statusMessage, err
}

func (r *AccessRuleReconciler) getProjectsNamespacesOfAccessRuleScope(ctx context.Context,
	accessRule *v1alpha1.AccessRule) (map[string]*kaiorg.Project, string, error) {
	allProjectsForAccessRule, statusMessage, err := r.getProjectsOfAccessRuleScope(ctx, accessRule)
	if err != nil {
		return map[string]*kaiorg.Project{}, statusMessage, err
	}

	namespacesProjectsMap, err := r.getNamespacesForProjects(ctx, allProjectsForAccessRule)
	if err != nil && len(namespacesProjectsMap) == 0 {
		return map[string]*kaiorg.Project{}, common.ScopeObjectFailedGettingNamespacesMessage, err
	}

	return namespacesProjectsMap, "", nil
}

func (r *AccessRuleReconciler) getProjectsOfAccessRuleScope(ctx context.Context, accessRule *v1alpha1.AccessRule,
) ([]kaiorg.Project, string, error) {
	var statusMessage string
	allProjectsForAccessRule, err := r.accessRulesProjectsMapper.GetProjectsOfAccessRuleScope(ctx, accessRule)
	if err == nil {
		return allProjectsForAccessRule, "", nil
	}

	if errors.IsNotFound(err) {
		statusMessage = fmt.Sprintf(common.ScopeObjectNotFoundMessage,
			common.AccessRuleScopeTypeToObj[accessRule.Spec.ScopeType], accessRule.Spec.ScopeName)
	} else {
		statusMessage = common.ScopeObjectFailedGettingMessage
	}
	return []kaiorg.Project{}, statusMessage, err
}

func (r *AccessRuleReconciler) getNamespacesForProjects(ctx context.Context,
	projects []kaiorg.Project) (map[string]*kaiorg.Project, error) {
	var err error
	result := map[string]*kaiorg.Project{}
	for i := range projects {
		project := projects[i]
		namespace, innerErr := r.getProjectNamespace(ctx, &project)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
			continue
		}
		result[namespace] = &project
	}

	return result, err
}

func (r *AccessRuleReconciler) getProjectNamespace(ctx context.Context, project *kaiorg.Project) (string, error) {
	if project.Status.Namespace != "" {
		// just to validate the namespace does exist
		namespaceObj := &corev1.Namespace{}
		err := r.cachedClient.Get(ctx, types.NamespacedName{Name: project.Status.Namespace}, namespaceObj)
		if err != nil {
			return "", err
		}
		return project.Status.Namespace, nil
	}
	return r.findProjectNamespace(ctx, project.Name)
}

func (r *AccessRuleReconciler) findProjectNamespace(ctx context.Context, projectName string) (string, error) {
	labelCondition := map[string]string{
		commonconstants.RunaiQueueLabel: projectName,
	}

	namespaces := &corev1.NamespaceList{}
	err := r.cachedClient.List(ctx, namespaces, client.MatchingLabels(labelCondition))
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed listing namespaces for project <%s>, error: %s",
			projectName, err.Error())
		return "", err
	}

	if len(namespaces.Items) == 0 {
		err = fmt.Errorf("no namespace found to have <%s> label with the name of project <%s>",
			commonconstants.RunaiQueueLabel, projectName)
		log.Ctx(ctx).Error().Msgf("Failed to determine the namespace of project: %s", err.Error())
		return "", err
	}
	if len(namespaces.Items) != 1 {
		err = fmt.Errorf("multiple namespaces found to have <%s> label with the name of project <%s> (%s, %s, ...)",
			commonconstants.RunaiQueueLabel, projectName, namespaces.Items[0].Name, namespaces.Items[1].Name)
		log.Ctx(ctx).Error().Msgf("Failed to determine the namespace of project: %s", err.Error())
		return "", err
	}

	return namespaces.Items[0].Name, nil
}

// getProjectsRoleBindingsWithAccessRulesSubject - lists all RoleBinding
// objects that this accessrule's subject was added to.
func (r *AccessRuleReconciler) getProjectsRoleBindingsWithAccessRulesSubject(ctx context.Context,
	accessRule *v1alpha1.AccessRule) ([]rbacv1.RoleBinding, error) {
	labelCondition := map[string]string{
		rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, true): "true",
	}

	roleBindingsList, err := r.listRoleBindingsWithLabelSelector(ctx, labelCondition)
	if err != nil {
		return []rbacv1.RoleBinding{}, err
	}
	return roleBindingsList.Items, nil
}

func (r *AccessRuleReconciler) listRoleBindingsWithLabelSelector(ctx context.Context, labelCondition map[string]string,
) (*rbacv1.RoleBindingList, error) {
	roleBindings := &rbacv1.RoleBindingList{}
	err := r.cachedClient.List(ctx, roleBindings, client.MatchingLabels(labelCondition))
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed listing rolebindings with labels <%v>, error: %s",
			labelCondition, err.Error())
		return nil, err
	}
	return roleBindings, nil
}
