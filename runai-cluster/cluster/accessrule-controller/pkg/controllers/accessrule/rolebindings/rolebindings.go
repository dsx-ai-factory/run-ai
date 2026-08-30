package rolebindings

import (
	"context"
	"fmt"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/rolebindingsutils"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	systemAdminAccessToLogsRoleBindingName = "runai-system-administrator-logs"
	systemAdministratorRoleName            = "system-administrator"
)

var namespacesForAccessSystemAdmin = []string{"runai-backend", "runai-reservation", "training-operator", "knative", "gpu-operator", "nim-operator"}

type RoleBindingsHandler struct {
	cachedClient client.Client
}

func NewRoleBindingsHandler(cachedClient client.Client) *RoleBindingsHandler {
	return &RoleBindingsHandler{
		cachedClient: cachedClient,
	}
}

func (r *RoleBindingsHandler) HandleRoleBindingForProject(
	ctx context.Context,
	accessRule *v1alpha1.AccessRule,
	roleBindingName string,
	namespace string,
	project *kaiorg.Project) error {
	return r.createOrUpdateRoleBindingForAccessRuleIfNeeded(ctx, accessRule, roleBindingName, namespace,
		false, project)
}

func (r *RoleBindingsHandler) HandleExtraRoleBindingsForAccessRule(
	ctx context.Context, accessRule *v1alpha1.AccessRule) (string, error) {
	var err error
	statusMessages := []string{}

	// Role and RoleBinding/ClusterRoleBinding names are always equal here

	statusMessage, innerErr := r.handleRunaiNamespaceRoleBindingForAccessRule(ctx, accessRule)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, statusMessage)
	}

	statusMessage, innerErr = r.handleClusterRoleBindingForAccessRule(ctx, accessRule)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, statusMessage)
	}

	statusMessage, innerErr = r.handleClusterScopedRunaiNamespaceRoleBindingForAccessRule(ctx, accessRule)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, statusMessage)
	}

	statusMessage, innerErr = r.handleExtraRoleBindingsForSystemAdmin(ctx, accessRule)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, statusMessage)
	}

	return strings.Join(statusMessages, ", "), err
}

func (r *RoleBindingsHandler) ValidateClusterRoleExists(ctx context.Context, clusterRoleName string,
) (string, error) {
	var statusMessage string

	_, err := r.getClusterRole(ctx, clusterRoleName)
	if err == nil {
		return "", nil
	}

	if k8s_errors.IsNotFound(err) {
		statusMessage = fmt.Sprintf(common.ClusterRoleNotFoundMessage,
			clusterRoleName)
	} else {
		statusMessage = fmt.Sprintf(common.ClusterRoleFailedGettingMessage,
			clusterRoleName)
	}
	return statusMessage, err
}

func (r *RoleBindingsHandler) handleRunaiNamespaceRoleBindingForAccessRule(ctx context.Context, accessRule *v1alpha1.AccessRule,
) (string, error) {
	runaiNamespaceRoleBindingName := rolebindingsutils.GetRunaiNamespaceRoleBindingName(accessRule.Spec.RoleName)

	statusMessage, err := r.ValidateClusterRoleExists(ctx, runaiNamespaceRoleBindingName)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed getting ClusterRole <%s> for access rule <%s>, error: %s",
			runaiNamespaceRoleBindingName, accessRule.Name, err.Error())
		return statusMessage, err
	}

	err = r.handleExistingRoleBinding(ctx, accessRule, runaiNamespaceRoleBindingName, runv1.DefaultNamespace)
	if err != nil {
		return common.FailedReconcileRunaiMessage, err
	}

	return "", nil
}

func (r *RoleBindingsHandler) handleClusterRoleBindingForAccessRule(ctx context.Context, accessRule *v1alpha1.AccessRule,
) (string, error) {
	clusterRoleBindingName := rolebindingsutils.GetClusterRoleBindingName(accessRule.Spec.RoleName)

	statusMessage, err := r.ValidateClusterRoleExists(ctx, clusterRoleBindingName)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed getting ClusterRole <%s> for access rule <%s>, error: %s",
			clusterRoleBindingName, accessRule.Name, err.Error())
		return statusMessage, err
	}

	err = r.handleExistingClusterRoleBinding(ctx, accessRule, clusterRoleBindingName)
	if err != nil {
		return common.FailedReconcileClusterRoleBindingMessage, err
	}

	return "", nil
}

func (r *RoleBindingsHandler) handleClusterScopedRunaiNamespaceRoleBindingForAccessRule(
	ctx context.Context, accessRule *v1alpha1.AccessRule) (string, error) {
	if accessRule.Spec.ScopeType != v1alpha1.ClusterScope {
		// this roleBinding is only relevant if the accessrule is cluster-scoped
		err := r.deleteFromClusterScopedRunaiNamespaceRoleBinding(ctx, accessRule)
		if err != nil {
			return common.FailedDeleteClusterScopedRunaiMessage, err
		}
		return "", nil
	}

	clusterScopedRunaiNamespaceRoleBindingName :=
		rolebindingsutils.GetClusterScopedRunaiNamespaceRoleBindingName(accessRule.Spec.RoleName)

	statusMessage, err := r.ValidateClusterRoleExists(ctx, clusterScopedRunaiNamespaceRoleBindingName)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			return "", nil
		}
		return statusMessage, err
	}

	err = r.handleRoleBindingIfExists(ctx, accessRule, clusterScopedRunaiNamespaceRoleBindingName,
		runv1.DefaultNamespace)
	if err != nil {
		return common.FailedReconcileClusterScopedRunaiMessage, err
	}

	return "", nil
}

func (r *RoleBindingsHandler) handleExtraRoleBindingsForSystemAdmin(ctx context.Context, accessRule *v1alpha1.AccessRule) (string, error) {
	if accessRule.Spec.RoleName != systemAdministratorRoleName {
		return "", nil
	}

	if statusMessage, err := r.ValidateClusterRoleExists(ctx, systemAdminAccessToLogsRoleBindingName); err != nil {
		if k8s_errors.IsNotFound(err) {
			return "", nil
		}
		return statusMessage, err
	}

	var err error
	var errorsStr []string
	for _, namespace := range namespacesForAccessSystemAdmin {
		log.Ctx(ctx).Debug().Msgf("RoleBindingsHandler.HandleExtraRoleBindingsForAccessRule going to create <%s> RoleBinding in %s namespace", systemAdminAccessToLogsRoleBindingName, namespace)
		innerErr := r.handleRoleBindingIfExists(ctx, accessRule, systemAdminAccessToLogsRoleBindingName, namespace)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
			errorsStr = append(errorsStr, fmt.Sprintf(common.FailedReconcileSystemAdminNamespaceMessage, namespace))
		}
		log.Ctx(ctx).Debug().Msgf("RoleBindingsHandler.HandleExtraRoleBindingsForAccessRule created <%s> RoleBinding in %s namespace with status: %t with err: %v", systemAdminAccessToLogsRoleBindingName, namespace, innerErr == nil, innerErr)
	}

	return strings.Join(errorsStr, "\n"), err
}
