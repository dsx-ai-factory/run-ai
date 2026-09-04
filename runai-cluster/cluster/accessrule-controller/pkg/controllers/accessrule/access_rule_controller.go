/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package accessrule

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/mappers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/rolebindings"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

const (
	rateLimiterBaseDelay = time.Second
	rateLimiterMaxDelay  = time.Minute * 10
)

// AccessRuleController controls an AccessRule object
type AccessRuleController struct {
	cachedClient client.Client

	accessRuleReconciler      *AccessRuleReconciler
	accessRulesProjectsMapper *mappers.AccessRulesProjectsMapper
}

func NewAccessRuleController(cachedClient client.Client) *AccessRuleController {
	accessRulesProjectsMapper := mappers.NewAccessRulesProjectsMapper(cachedClient)
	roleBindingsHandler := rolebindings.NewRoleBindingsHandler(cachedClient)
	accessRuleReconciler := NewAccessRuleReconciler(cachedClient,
		accessRulesProjectsMapper, roleBindingsHandler)
	return &AccessRuleController{
		cachedClient:              cachedClient,
		accessRuleReconciler:      accessRuleReconciler,
		accessRulesProjectsMapper: accessRulesProjectsMapper,
	}
}

//+kubebuilder:rbac:groups=run.ai,resources=accessrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=run.ai,resources=accessrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=run.ai,resources=accessrules/finalizers,verbs=update

//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:groups=run.ai,resources=projects,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (c *AccessRuleController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)

	accessRule := &v1alpha1.AccessRule{}
	if shouldContinue, err := c.getAccessRuleForRequest(ctx, req, accessRule); err != nil {
		return ctrl.Result{Requeue: true}, err
	} else if !shouldContinue {
		return ctrl.Result{Requeue: false}, nil
	}

	var err error

	if accessRule.DeletionTimestamp.IsZero() {
		if err = c.addControllerAsFinalizerIfNeeded(ctx, accessRule); err != nil {
			return ctrl.Result{}, err
		}

		err = c.accessRuleReconciler.Reconcile(ctx, accessRule)
	} else {
		err = c.finalize(ctx, accessRule)
	}

	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (c *AccessRuleController) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := c.indexFields(ctx, mgr)
	if err != nil {
		return err
	}

	// watches Projects as well - converting each to Reconcile on all relevant AccessRules;
	// but with event filter - that filters in only Projects that are Created, Deleted, or their Department is Updated.
	b := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AccessRule{}).
		Watches(&rbacv1.RoleBinding{}, handler.EnqueueRequestsFromMapFunc(c.MapRoleBindingToAccessRuleEvent)).
		Watches(&rbacv1.ClusterRoleBinding{}, handler.EnqueueRequestsFromMapFunc(c.MapClusterRoleBindingToAccessRuleEvent))

	b = b.Watches(&kaiorg.Project{}, handler.EnqueueRequestsFromMapFunc(c.MapProjectToAccessRuleEvent))

	return b.
		WithEventFilter(ControllerPredicateFuncs()).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](rateLimiterBaseDelay, rateLimiterMaxDelay)}).
		Complete(c)
}

func (c *AccessRuleController) getAccessRuleForRequest(ctx context.Context, req ctrl.Request, accessRule *v1alpha1.AccessRule) (shouldContinue bool, bool error) {
	err := c.cachedClient.Get(ctx, req.NamespacedName, accessRule)
	if err == nil {
		return true, nil
	}

	if errors.IsNotFound(err) {
		log.Ctx(ctx).Warn().Msgf("Can't locate access rule <%s> in cluster, this might also mean it was purposefully deleted",
			req.Name)
		// Nothing really more to do.
		return false, nil
	}

	log.Ctx(ctx).Error().Msgf("Could not extract access rule <%s> from incoming request, error: %s",
		req.Name, err.Error())
	return false, err
}

func (c *AccessRuleController) indexFields(ctx context.Context, mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(
		ctx, &kaiorg.Project{},
		common.ProjectsParentNameField, ProjectsParentIndexer)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed indexing project field: %s, error: %s",
			common.ProjectsParentNameField, err.Error())
		return err
	}

	err = mgr.GetFieldIndexer().IndexField(
		ctx, &v1alpha1.AccessRule{},
		common.AccessRuleProjectNameField, AccessRulesProjectIndexer)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed indexing access rule field: %s, error: %s",
			common.AccessRuleProjectNameField, err.Error())
		return err
	}

	err = mgr.GetFieldIndexer().IndexField(
		ctx, &v1alpha1.AccessRule{},
		common.AccessRuleDepartmentNameField, AccessRulesDepartmentIndexer)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed indexing access rule field: %s, error: %s",
			common.AccessRuleDepartmentNameField, err.Error())
		return err
	}

	err = mgr.GetFieldIndexer().IndexField(
		ctx, &v1alpha1.AccessRule{},
		common.AccessRuleClusterField, AccessRulesClusterIndexer)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed indexing access rule field: %s, error: %s",
			common.AccessRuleClusterField, err.Error())
		return err
	}

	return nil
}

// ProjectsParentIndexer indexes the kai Project by its Spec.Parent (the department linkage).
func ProjectsParentIndexer(object client.Object) (indexedKeys []string) {
	project, ok := object.(*kaiorg.Project)
	if !ok {
		log.Error().Msgf("KaiProjectsParentIndexer: Cannot convert object to *kaiorg.Project: %v", object)
		return indexedKeys
	}

	return []string{project.Spec.Parent}
}

func AccessRulesProjectIndexer(object client.Object) (indexedKeys []string) {
	accessRule, ok := object.(*v1alpha1.AccessRule)
	if !ok {
		log.Error().Msgf("AccessRulesProjectIndexer: Cannot convert object to *v1alpha1.AccessRule: %v", object)
		return indexedKeys
	}

	// this will filter out access rules that are not project scoped
	indexedKey := ""
	if accessRule.Spec.ScopeType == v1alpha1.ProjectScope {
		indexedKey = accessRule.Spec.ScopeName
	}
	return []string{indexedKey}
}

func AccessRulesDepartmentIndexer(object client.Object) (indexedKeys []string) {
	accessRule, ok := object.(*v1alpha1.AccessRule)
	if !ok {
		log.Error().Msgf("AccessRulesDepartmentIndexer: Cannot convert object to *v1alpha1.AccessRule: %v", object)
		return indexedKeys
	}

	// this will filter out access rules that are not department scoped
	indexedKey := ""
	if accessRule.Spec.ScopeType == v1alpha1.DepartmentScope {
		indexedKey = accessRule.Spec.ScopeName
	}
	return []string{indexedKey}
}

func AccessRulesClusterIndexer(object client.Object) (indexedKeys []string) {
	accessRule, ok := object.(*v1alpha1.AccessRule)
	if !ok {
		log.Error().Msgf("AccessRulesClusterIndexer: Cannot convert object to *v1alpha1.AccessRule: %v", object)
		return indexedKeys
	}

	// this will filter out access rules that are not cluster scoped
	indexedKey := ""
	if accessRule.Spec.ScopeType == v1alpha1.ClusterScope {
		indexedKey = common.ClusterNameFieldSelectorValue
	}
	return []string{indexedKey}
}
