package rolebindings

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/rolebindingsutils"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	rbacv1 "k8s.io/api/rbac/v1"
)

func (r *RoleBindingsHandler) addToRoleBindingSubjects(
	ctx context.Context, accessRule *v1alpha1.AccessRule,
	roleBinding *rbacv1.RoleBinding, projectRoleBinding bool) error {
	accessRuleAddedSubjectLabelKey := rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, projectRoleBinding)
	if _, found := roleBinding.Labels[accessRuleAddedSubjectLabelKey]; found {
		if isAccessRuleInSubjects(accessRule, roleBinding.Subjects) {
			log.Ctx(ctx).Debug().Msgf("Subject already exists in RoleBinding <%s/%s> "+
				"for accessrule <%s> - nothing to do",
				roleBinding.Namespace, roleBinding.Name, accessRule.Name)
			return nil
		}
	}

	accessRulesSubject := rolebindingsutils.GetSubjectForAccessRule(accessRule)
	newSubjects := append(roleBinding.Subjects, accessRulesSubject)
	roleBinding.Subjects = newSubjects

	if roleBinding.Labels == nil {
		roleBinding.Labels = map[string]string{}
	}
	roleBinding.Labels[accessRuleAddedSubjectLabelKey] = "true"

	err := r.cachedClient.Update(ctx, roleBinding)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed updating RoleBinding <%s> for access rule <%s> "+
			"in namespace <%s>, error: %s",
			roleBinding.Name, accessRule.Name, roleBinding.Namespace, err.Error())
		return err
	}

	log.Ctx(ctx).Debug().Msgf("Successfully added subject <%s/%s> of accessrule <%s> to RoleBinding <%s/%s>",
		accessRulesSubject.Kind, accessRulesSubject.Name, accessRule.Name, roleBinding.Namespace, roleBinding.Name)
	return nil
}

func (r *RoleBindingsHandler) addToClusterRoleBindingSubjects(
	ctx context.Context, accessRule *v1alpha1.AccessRule, clusterRoleBinding *rbacv1.ClusterRoleBinding) error {
	accessRuleAddedSubjectLabelKey := rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, false)
	if _, found := clusterRoleBinding.Labels[accessRuleAddedSubjectLabelKey]; found {
		if isAccessRuleInSubjects(accessRule, clusterRoleBinding.Subjects) {
			log.Ctx(ctx).Debug().Msgf("Subject already exists in ClusterRoleBinding <%s> "+
				"for accessrule <%s> - nothing to do",
				clusterRoleBinding.Name, accessRule.Name)
			return nil
		}
	}

	accessRulesSubject := rolebindingsutils.GetSubjectForAccessRule(accessRule)
	newSubjects := append(clusterRoleBinding.Subjects, accessRulesSubject)
	clusterRoleBinding.Subjects = newSubjects

	if clusterRoleBinding.Labels == nil {
		clusterRoleBinding.Labels = map[string]string{}
	}
	clusterRoleBinding.Labels[accessRuleAddedSubjectLabelKey] = "true"

	err := r.cachedClient.Update(ctx, clusterRoleBinding)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed updating ClusterRoleBinding <%s> for access rule <%s>, error: %s",
			clusterRoleBinding.Name, accessRule.Name, err.Error())
		return err
	}

	log.Ctx(ctx).Debug().Msgf("Successfully added subject <%s/%s> of accessrule <%s> to ClusterRoleBinding <%s>",
		accessRulesSubject.Kind, accessRulesSubject.Name, accessRule.Name, clusterRoleBinding.Name)
	return nil
}

// deleteFromRoleBindingSubjects
// returns:
// roleBinding after update
// error
func (r *RoleBindingsHandler) deleteFromRoleBindingSubjects(
	ctx context.Context, accessRule *v1alpha1.AccessRule,
	roleBinding *rbacv1.RoleBinding, projectRoleBinding bool,
) (*rbacv1.RoleBinding, error) {
	accessRuleAddedSubjectLabelKey := rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, projectRoleBinding)
	if _, found := roleBinding.Labels[accessRuleAddedSubjectLabelKey]; !found {
		log.Ctx(ctx).Debug().Msgf("Subject of accessrule <%s> already does not exist in RoleBinding <%s/%s>"+
			" - nothing to do",
			accessRule.Name, roleBinding.Namespace, roleBinding.Name)
		return roleBinding, nil
	}

	accessRulesSubject := rolebindingsutils.GetSubjectForAccessRule(accessRule)
	newSubjects := utils.DeleteFirstFromList(roleBinding.Subjects, accessRulesSubject)
	roleBinding.Subjects = newSubjects

	if roleBinding.Labels != nil {
		delete(roleBinding.Labels, rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, projectRoleBinding))
	}

	err := r.cachedClient.Update(ctx, roleBinding)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed updating RoleBinding <%s> "+
			"to remove subject of access rule <%s> in namespace <%s>, error: %s",
			roleBinding.Name, accessRule.Name, roleBinding.Namespace, err.Error())
		return nil, err
	}

	log.Ctx(ctx).Debug().Msgf("Successfully removed subject <%s/%s> of accessrule <%s> from RoleBinding <%s/%s>",
		accessRulesSubject.Kind, accessRulesSubject.Name, accessRule.Name, roleBinding.Namespace, roleBinding.Name)
	return roleBinding, nil
}

func (r *RoleBindingsHandler) deleteFromClusterRoleBindingSubjects(
	ctx context.Context, accessRule *v1alpha1.AccessRule, clusterRoleBinding *rbacv1.ClusterRoleBinding) error {
	accessRuleAddedSubjectLabelKey := rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, false)
	if _, found := clusterRoleBinding.Labels[accessRuleAddedSubjectLabelKey]; !found {
		log.Ctx(ctx).Debug().Msgf("Subject of accessrule <%s> already does not exist in ClusterRoleBinding <%s>"+
			" - nothing to do",
			accessRule.Name, clusterRoleBinding.Name)
		return nil
	}

	accessRulesSubject := rolebindingsutils.GetSubjectForAccessRule(accessRule)
	newSubjects := utils.DeleteFirstFromList(clusterRoleBinding.Subjects, accessRulesSubject)
	clusterRoleBinding.Subjects = newSubjects

	if clusterRoleBinding.Labels != nil {
		delete(clusterRoleBinding.Labels, rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, false))
	}

	err := r.cachedClient.Update(ctx, clusterRoleBinding)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed updating ClusterRoleBinding <%s> "+
			"to remove subject for access rule <%s>, error: %s",
			clusterRoleBinding.Name, accessRule.Name, err.Error())
		return err
	}

	log.Ctx(ctx).Debug().Msgf("Successfully removed subject <%s/%s>"+
		" of accessrule <%s> from ClusterRoleBinding <%s>",
		accessRulesSubject.Kind, accessRulesSubject.Name, accessRule.Name, clusterRoleBinding.Name)
	return nil
}

func isAccessRuleInSubjects(accessRule *v1alpha1.AccessRule, subjects []rbacv1.Subject) bool {
	accessRuleSubject := rolebindingsutils.GetSubjectForAccessRule(accessRule)
	for _, subject := range subjects {
		if subject.Name == accessRuleSubject.Name && subject.Kind == accessRuleSubject.Kind {
			return true
		}
	}
	return false
}
