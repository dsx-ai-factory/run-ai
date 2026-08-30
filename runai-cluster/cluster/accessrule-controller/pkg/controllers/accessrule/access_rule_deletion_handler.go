package accessrule

import (
	"context"
	"fmt"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"

	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

// HandleDeletion of accessrule.
// returns:
// - statusMessage string
// - err error
func (r *AccessRuleReconciler) HandleDeletion(ctx context.Context, accessRule *v1alpha1.AccessRule) (string, error) {
	return r.deleteSubjectFromAllRoleBindings(ctx, accessRule)
}

func (r *AccessRuleReconciler) deleteSubjectFromAllRoleBindings(
	ctx context.Context, accessRule *v1alpha1.AccessRule) (string, error) {
	var err error
	statusMessages := []string{}

	innerStatusMessage, innerErr := r.deleteFromProjectsRoleBindingsForAccessRule(ctx, accessRule)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, innerStatusMessage)
	}

	innerStatusMessage, innerErr = r.roleBindingsHandler.DeleteFromExtraRoleBindingsForAccessRule(ctx, accessRule)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, innerStatusMessage)
	}

	return strings.Join(statusMessages, "; "), err
}

// deleteFromProjectsRoleBindingsForAccessRule - deletes accessrule's subject from the RoleBindings
// in the namespaces of the accessrule's scope.
// returns:
// - statusMessage string
// - err error
func (r *AccessRuleReconciler) deleteFromProjectsRoleBindingsForAccessRule(ctx context.Context,
	accessRule *v1alpha1.AccessRule,
) (string, error) {
	roleBindings, err := r.getProjectsRoleBindingsWithAccessRulesSubject(ctx, accessRule)
	if err != nil {
		return common.ScopeObjectFailedGettingRoleBindingsMessage, err
	}

	failedNamespaces := []string{}
	for i := range roleBindings {
		roleBinding := roleBindings[i]
		log.Ctx(ctx).Debug().Msgf("Handling permissions removal for access rule <%s> in namespace <%s>",
			accessRule.Name, roleBinding.Namespace)

		innerErr := r.roleBindingsHandler.DeleteFromExistingProjectRoleBinding(ctx, accessRule, &roleBinding)
		if innerErr != nil {
			log.Ctx(ctx).Error().Msgf("Failed handling permissions removal from access rule's <%s> rolebinding "+
				"in namespace <%s>, error: %s",
				accessRule.Name, roleBinding.Namespace, innerErr.Error())
			err = multierror.Append(err, innerErr)
			failedNamespaces = append(failedNamespaces, roleBinding.Namespace)
		}
	}

	statusMessage := ""
	if len(failedNamespaces) > 0 {
		statusMessage = fmt.Sprintf(common.FailedHandlingDeletionInNamespacesMessage,
			strings.Join(failedNamespaces, ", "))
	}

	return statusMessage, err
}
