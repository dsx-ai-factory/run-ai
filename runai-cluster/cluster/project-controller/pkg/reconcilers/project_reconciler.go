package reconcilers

import (
	"context"
	"encoding/json"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	multierror "github.com/hashicorp/go-multierror"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/handlers"
	deletionhandlers "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/handlers/deletion"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ProjectReconciler reconciles Project objects and creates associated Scheduler resources
type ProjectReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	reconcilers      []handlers.ProjectResourceHandler
	statusReconciler handlers.ProjectStatusResourceHandler
	deletionHandlers []deletionhandlers.ProjectResourceDeletionHandler
	deletionBlockers []deletionhandlers.ProjectResourceDeletionHandler
	config           *config.ProjectReconcilerConfig

	// projectEvents is a channel held locally by the controller to pass reconcile events between the reconcilers,
	// instead of going out via the api.
	projectEvents chan event.GenericEvent
}

func NewProjectReconciler(
	cachedClient client.Client,
	uncachedReader client.Reader,
	scheme *runtime.Scheme,
	projectEvents chan event.GenericEvent,
	config *config.ProjectReconcilerConfig) *ProjectReconciler {
	logger := ctrl.Log.WithName("reconcilers").WithName(common.LogProjectTag)

	reconcilers := []handlers.ProjectResourceHandler{
		handlers.NewQueueResourceHandler(cachedClient),
	}
	deletionHandlers := []deletionhandlers.ProjectResourceDeletionHandler{
		deletionhandlers.NewQueueDeletionHandler(cachedClient),
	}
	// Deletion blockers are config-driven: they are loaded once at startup from the
	// project-delete-blockers ConfigMap (each group becomes a generic ConfigurableBlocker).
	// With no namespace configured / no ConfigMap (the open-source default) there are no
	// blockers, so deletion is never blocked.
	groups := loadDeletionBlockers(uncachedReader, config, logger)
	deletionBlockers := make([]deletionhandlers.ProjectResourceDeletionHandler, 0, len(groups))
	for _, group := range groups {
		deletionBlockers = append(deletionBlockers, deletionhandlers.NewConfigurableBlocker(cachedClient, group))
	}

	if config.LimitRange {
		reconcilers = append(reconcilers, handlers.NewLimitRangeResourceHandler(cachedClient))
		deletionHandlers = append(deletionHandlers, deletionhandlers.NewLimitRangeDeletionHandler(cachedClient))
	}

	if config.CreateRoleBindings { // this must run as second reconciler
		reconcilers = append([]handlers.ProjectResourceHandler{
			handlers.NewRoleBindingsResourceHandler(cachedClient, config.RoleBindingsCm, config.RoleBindingsCmNamespace, !config.CreateNamespaces, config.IsOpenshift)},
			reconcilers...)
	}

	// this must run as first reconciler
	reconcilers = append([]handlers.ProjectResourceHandler{
		handlers.NewNamespaceResourceHandler(cachedClient, config.IsOpenshift, config.CreateNamespaces)}, reconcilers...)

	// always keep the project status handler as the last handler
	statusReconciler := handlers.NewProjectStatusHandler(cachedClient)
	return &ProjectReconciler{
		Client:           cachedClient,
		Log:              logger,
		Scheme:           scheme,
		projectEvents:    projectEvents,
		reconcilers:      reconcilers,
		statusReconciler: statusReconciler,
		deletionHandlers: deletionHandlers,
		deletionBlockers: deletionBlockers,
		config:           config,
	}
}

// loadDeletionBlockers reads the project-delete-blockers ConfigMap (in the configured
// namespace) once at startup and turns each group into a generic ConfigurableBlocker.
// It reads through apiReader (the manager's uncached client) because the controller's
// cache is not yet started at construction time. With no namespace configured, no
// ConfigMap present, or a malformed ConfigMap, it returns no blockers (deletion is not blocked)
func loadDeletionBlockers(apiReader client.Reader, cfg *config.ProjectReconcilerConfig, logger logr.Logger) []deletionhandlers.BlockerGroup {
	if cfg.GvkDeleteBlockersNamespace == "" {
		return nil
	}

	var cm corev1.ConfigMap
	err := apiReader.Get(context.Background(), client.ObjectKey{
		Namespace: cfg.GvkDeleteBlockersNamespace,
		Name:      common.GvkDeleteBlockersConfigMapName,
	}, &cm)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("project-delete-blockers ConfigMap not found, project-deletion blockers are disabled",
				common.LogNamespaceTag, cfg.GvkDeleteBlockersNamespace)
			return nil
		} else {
			logger.Error(err, "Failed reading project-delete-blockers ConfigMap, project-deletion blockers are disabled",
				common.LogNamespaceTag, cfg.GvkDeleteBlockersNamespace)
			return nil
		}
	}

	groups, err := deletionhandlers.BlockerGroupsFromConfigMapData(cm.Data)
	if err != nil {
		logger.Error(err, "Failed parsing project-delete-blockers ConfigMap, project-deletion blockers are disabled",
			common.LogNamespaceTag, cfg.GvkDeleteBlockersNamespace)
		return nil
	}
	logger.Info("Loaded project-deletion blockers from ConfigMap",
		common.LogNamespaceTag, cfg.GvkDeleteBlockersNamespace, "blockers", groups)
	return groups
}

// Reconcile - This reconciler responds to events of the Project resource by modifying the secondary resources derived
// from the Project resource: (Queue, Namespace, LimitRange and image pull secret).
// The second functionality is to respond to changes in these 'owned' resources and to run the reconciliation cycle
// on the owning Project resource to keep this logic idempotent.
func (reconciler *ProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	project := &kaiv1alpha1.Project{}
	if result, abort, err := reconciler.retrieveProject(req, project); err != nil || abort {
		return result, err
	}

	if project.DeletionTimestamp.IsZero() {
		// This project is not being deleted - check if it already has this controller as a finalizer and add it if not.
		if err := reconciler.AddControllerAsFinalizerIfNeeded(ctx, project); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		// This project is being deleted - check if the finalizer is set. If it is we need to delete the project,
		// if not it can be ignored. (because of delete-jitter)
		return reconciler.Finalize(ctx, project)
	}
	return reconciler.reconcileProject(ctx, project)
}

func (reconciler *ProjectReconciler) retrieveProject(
	req ctrl.Request, project *kaiv1alpha1.Project,
) (result ctrl.Result, abort bool, err error) {
	if getErr := reconciler.GetProject(req, project); getErr != nil {
		if errors.IsNotFound(getErr) {
			reconciler.Log.Info("Can't locate project in cluster, "+
				"this might also mean it was purposefully deleted",
				common.LogProjectTag, req.Name)
			// Nothing really more to do.
			abort = true
			result = ctrl.Result{Requeue: false}
		} else {
			reconciler.Log.Error(getErr, "Could not extract Project from incoming request: ",
				"Request", req)
			err = getErr
			abort = true
			result = ctrl.Result{Requeue: true}
		}
	}
	return result, abort, err
}

func (reconciler *ProjectReconciler) reconcileProject(ctx context.Context,
	project *kaiv1alpha1.Project) (ctrl.Result, error) {
	reconciler.Log.Info("Reconciling Project", common.LogProjectTag, project.Name)

	var err error
	conditionsChanged := false
	for _, subReconciler := range reconciler.reconcilers {
		projectConditions, innerErr := subReconciler.HandleResource(*project)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
			reconciler.Log.Error(innerErr, fmt.Sprintf("Error running %T reconciler, "+
				"this is maybe due to former reconciler error", subReconciler))
		}
		conditionsChanged = handlers.UpdateProjectConditions(&project.Status, projectConditions) || conditionsChanged
	}

	innerErr := reconciler.statusReconciler.ReconcileStatus(ctx, project, conditionsChanged)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	return ctrl.Result{}, err
}

func (reconciler *ProjectReconciler) AddControllerAsFinalizerIfNeeded(ctx context.Context, project *kaiv1alpha1.Project) (err error) {
	if common.ContainsString(config.FinalizerName(), project.Finalizers) {
		// Controller already present in finalizers list
		return nil
	}
	reconciler.Log.Info("Adding controller to finalizer list in Project",
		common.LogProjectTag, project.Name)
	project.Finalizers = append(project.Finalizers, config.FinalizerName())
	if err = reconciler.updateFinalizers(ctx, project); err != nil {
		reconciler.Log.Error(err, "Error adding controller to finalizer list in Project",
			common.LogProjectTag, project.Name)
	}
	return err
}

func (reconciler *ProjectReconciler) Finalize(ctx context.Context, project *kaiv1alpha1.Project) (ctrl.Result, error) {
	if !common.ContainsString(config.FinalizerName(), project.Finalizers) {
		reconciler.Log.Info("Got a finalization event for Project, but this controller is "+
			"not in its finalizers list. Skipping finalization.",
			common.LogProjectTag, project.Name)
		return ctrl.Result{}, nil
	}
	if common.IsManuallyOverridden(project) {
		reconciler.Log.Info("Project is marked as manually overridden, it will not be finalized",
			common.LogProjectTag, project.Name)
		return ctrl.Result{}, nil
	}

	if err := reconciler.finalizeProject(ctx, project); err != nil {
		if !common.IsForceDelete(project) {
			// Already logged by sub-finalizer
			return ctrl.Result{}, fmt.Errorf(
				"errors encountered during deletion of secondary resources of Project '%s'; error: %w",
				project.Name, err)
		}
	}
	reconciler.addMissingQueueSpecIfNeeded(project)

	if err := reconciler.deleteFinalizer(ctx, project); err != nil {
		return ctrl.Result{}, err
	}

	// Reaching here means finalization of all secondary resources went ok.
	reconciler.Log.Info("Done deleting all secondary resources for Project",
		common.LogProjectTag, project.Name)
	return ctrl.Result{}, nil
}

func (reconciler *ProjectReconciler) addMissingQueueSpecIfNeeded(project *kaiv1alpha1.Project) {
	if project.Spec.Queues == nil {
		reconciler.Log.Info("Warning - adding empty queues to project obj to avoid invalid object",
			common.LogProjectTag, project.Name)
		project.Spec.Queues = []kaiv1alpha1.QueueConfig{}
	}
}

func (reconciler *ProjectReconciler) finalizeProject(ctx context.Context, project *kaiv1alpha1.Project) error {
	reconciler.Log.Info("Deleting Project and secondary resources",
		common.LogProjectTag, project.Name)

	var err error
	conditionsChanged := false
	if project.Spec.DeletionType != nil &&
		*project.Spec.DeletionType == kaiv1alpha1.Blocking {
		for _, deletionBlocker := range reconciler.deletionBlockers {
			projectConditions, innerErr := deletionBlocker.OnDelete(project)
			if innerErr != nil {
				err = multierror.Append(err, innerErr)
			}

			conditionsChanged = handlers.UpdateProjectConditions(&project.Status, projectConditions) || conditionsChanged
		}
	}
	if err != nil {
		_ = reconciler.statusReconciler.ReconcileStatusOnDeletion(ctx, project, conditionsChanged)
		return err
	}

	for _, deletionHandler := range reconciler.deletionHandlers {
		projectConditions, innerErr := deletionHandler.OnDelete(project)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}

		conditionsChanged = handlers.UpdateProjectConditions(&project.Status, projectConditions) || conditionsChanged
	}

	// even if returned error - we don't want to fail this function if the finalization succeeded
	_ = reconciler.statusReconciler.ReconcileStatusOnDeletion(ctx, project, conditionsChanged)

	return err
}

func (reconciler *ProjectReconciler) deleteFinalizer(ctx context.Context, project *kaiv1alpha1.Project) error {
	reconciler.Log.Info("Removing finalizer <%v> from project's <%v> finalizers list",
		config.FinalizerName(), project.Name)
	project.Finalizers = common.DeleteTerm(config.FinalizerName(), project.Finalizers)
	if err := reconciler.updateFinalizers(ctx, project); err != nil {
		reconciler.Log.Error(err, "Error removing controller from finalizer list in Project",
			common.LogProjectTag, project.Name)
		return err
	}
	return nil
}

func (reconciler *ProjectReconciler) updateFinalizers(ctx context.Context, project *kaiv1alpha1.Project) (err error) {
	patchBytes, err := updateFinalizersPatchBytes(project.Finalizers)
	if err != nil {
		reconciler.Log.Error(err, "Failed to json.Marshal patch - update finalizers of project",
			common.LogProjectTag, project.Name)
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = reconciler.Patch(ctx, project, patch)
	if err != nil {
		return err
	}
	return nil
}

func updateFinalizersPatchBytes(finalizers []string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{
		"finalizers": finalizers}})
}

func (reconciler *ProjectReconciler) GetProject(req ctrl.Request, project *kaiv1alpha1.Project) error {
	return reconciler.Get(context.Background(), req.NamespacedName, project)
}

func (reconciler *ProjectReconciler) SetupWithManager(mgr ctrl.Manager, config *config.ProjectReconcilerConfig) error {
	result := ctrl.NewControllerManagedBy(mgr).
		For(&kaiv1alpha1.Project{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		// TODO: this single line (WatchesRawSource...) should be removed after 2.27 and after new project crd upgrade
		WatchesRawSource(source.Channel[client.Object](reconciler.projectEvents, &handler.TypedEnqueueRequestForObject[client.Object]{})).
		Watches(&corev1.Namespace{}, handler.EnqueueRequestsFromMapFunc(reconciler.MapNamespaceToProjectEvent)).
		Watches(&rbacv1.RoleBinding{}, handler.EnqueueRequestsFromMapFunc(reconciler.MapRoleBindingToProjectEvent)).
		Owns(&kaiv2.Queue{})

	if config.CreateRoleBindings && config.RoleBindingsCm != "" {
		result = result.Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(reconciler.MapRoleBindingsConfigMapToProjectEvents),
		)
	}

	if config.LimitRange {
		result = result.Owns(&corev1.LimitRange{})
	}

	return result.
		WithEventFilter(ProjectEventFilter).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](rateLimiterBaseDelay, rateLimiterMaxDelay)}).
		Complete(reconciler)
}
