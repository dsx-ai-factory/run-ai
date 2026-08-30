package distributedinference

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/controller"
	resourcemanager "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/resources"
	distributedstatus "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/distributed"
	issue2 "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/issue"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/issue"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	workloadConstants "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerUtil "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

var (
	log = logger.Log.WithName("reconcile.dinfw")
)

type Controller struct {
	workloadType workloadConstants.WorkloadType
	k8sClient    client.Client
	isOpenshift  bool

	resourceManager resourcemanager.K8sResourceManagerInterface
	builder         builders.DistributedInferenceBuilderInterface
	statusUpdater   distributedstatus.DistributedStatusUpdaterInterface
}

// SetupWithManager sets up the controller with the Manager
func (c *Controller) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconcile int) error {
	manager := ctrl.NewControllerManagedBy(mgr).
		For(&dinfw.DistributedInferenceWorkload{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Secret{}).
		// Watch LeaderWorkerSet only for status changes
		Watches(
			&lwsv1.LeaderWorkerSet{},
			handler.EnqueueRequestForOwner(mgr.GetScheme(), mgr.GetRESTMapper(), &dinfw.DistributedInferenceWorkload{}),
			builder.WithPredicates(controller.LwsStatusChangesOnlyPredicate()),
		).
		// Watch for changes to Pods and enqueue InferenceWorkload that reference it
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(controller.DistributedInferencePodsFilter()),
		).
		WithOptions(controllerUtil.Options{MaxConcurrentReconciles: maxConcurrentReconcile,
			RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
				constants.RateLimiterBaseDelay, constants.RateLimiterMaxDelay),
		}).
		WithEventFilter(controller.InferencePredicate())

	// Watch owned routing resources based on the active routing mode(s).
	// Both can be enabled simultaneously during migration from Ingress to Gateway API.
	if workloadconfig.IngressEnabled() && c.isOpenshift {
		manager.Owns(&openshiftV1.Route{})
	}
	if workloadconfig.IngressEnabled() && !c.isOpenshift {
		manager.Owns(&netv1.Ingress{})
	}
	if workloadconfig.GatewayEnabled() {
		manager.Owns(&gatewayv1.HTTPRoute{})
	}

	return manager.Complete(c)
}

func NewController(k8sClient client.Client, isOpenshift bool, elasticEnabled bool) *Controller {
	c := Controller{
		workloadType: controller.TypeDistributedInferenceWorkload,
		k8sClient:    k8sClient,
		isOpenshift:  isOpenshift,
	}

	c.resourceManager = resourcemanager.NewResourceManager(c.k8sClient, &issue2.InferenceResourceGrouper{})
	c.builder = builders.NewDistributedInferenceBuilder(c.k8sClient, isOpenshift, elasticEnabled)
	c.statusUpdater = distributedstatus.NewDistributedStatusUpdater(c.k8sClient)

	return &c
}

func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger.IntoContext(ctx, log)

	resourceId := fmt.Sprintf("%s:%s", req.Namespace, req.Name)
	log.Info("reconcile started", "id", resourceId)

	// obtain the resource being reconciled
	workload := &dinfw.DistributedInferenceWorkload{}

	err := c.k8sClient.Get(ctx, req.NamespacedName, workload)
	if err != nil {
		// there was an error to get the resource from k8s
		statusError := err.(*errors.StatusError)
		if statusError == nil || statusError.Status().Code != 404 {
			log.Error(err, "failed to load the workload", "id", resourceId)
		} else {
			log.Info("workload has been deleted", "id", resourceId)
		}

		// assume that the resource has been deleted with so just return normally
		return ctrl.Result{}, nil
	}

	// copy the original workload to use for status updates
	originalWorkload := workload.DeepCopy()

	// if the workload is marked for deletion, return normally
	if !workload.DeletionTimestamp.IsZero() {
		log.Info("workload is marked for deletion", "id", resourceId)
		return ctrl.Result{}, nil
	}

	// inject the project name to the context
	ctx, err = workloadproject.SetProjectInContextFromNamespace(ctx, c.k8sClient, workload.Namespace, workload.Name, workload.Spec.Usage)
	if err != nil {
		log.Error(err, "failed to inject project name to context", "id", resourceId)
		return c.reconcileFailed(ctx, err, false, workload, originalWorkload, nil)
	}

	// set workload labels and annotations
	err = c.setWorkloadLabelsAndAnnotations(ctx, workload)
	if err != nil {
		log.Error(err, "failed to set workload labels and annotations", "id", resourceId)
		return c.reconcileFailed(ctx, err, false, workload, originalWorkload, nil)
	}

	// build and upsert lws
	resources, issues, isConflict := c.buildAndUpsertResources(ctx, workload)
	if len(issues) > 0 {
		// merge the issues into a single error
		mergedErr := issues[0].GetError()
		for _, iss := range issues[1:] {
			mergedErr = fmt.Errorf("%w; %w", mergedErr, iss.GetError())
		}
		log.Error(mergedErr, "failed to build and upsert resources", "id", resourceId)
		return c.reconcileFailed(ctx, mergedErr, !isConflict, workload, originalWorkload, nil)
	}

	cleanupLabels := map[string]string{constants.LabelRelease: workload.Name}
	if err := resourcemanager.CleanupDisabledNetworkingResources(ctx, c.k8sClient, workload.Namespace, cleanupLabels, c.isOpenshift); err != nil {
		return c.reconcileFailed(ctx, err, true, workload, originalWorkload, nil)
	}

	_, err = c.reconcileSucceeded(ctx, workload, originalWorkload, resources)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("reconcile completed successfully", "id", resourceId)

	return ctrl.Result{}, nil
}

func getBackOffForUpdateLws() *backoff.ExponentialBackOff {
	backOffForUpdateLws := backoff.NewExponentialBackOff()
	backOffForUpdateLws.InitialInterval = 50 * time.Millisecond // sets the initial interval between retries.
	backOffForUpdateLws.MaxInterval = 300 * time.Millisecond    // sets the maximum interval between retries.
	backOffForUpdateLws.MaxElapsedTime = time.Second            // sets the maximum total time for retries.
	return backOffForUpdateLws
}

// upsertLws creates or updates the lws with retry mechanism in case of conflict
// it returns the issues and a boolean indicating if the error is a conflict error
func (c *Controller) upsertLws(ctx context.Context, workload *dinfw.DistributedInferenceWorkload, lws client.Object) ([]*issue.Issue, bool) {
	log := logger.FromContext(ctx)

	var issues []*issue.Issue
	var isConflict bool
	attempt := 0
	ticker := backoff.NewTicker(getBackOffForUpdateLws())
	defer ticker.Stop()

	for range ticker.C {
		issues, isConflict = c.resourceManager.UpsertResources(ctx, workload.ObjectMeta, []client.Object{lws},
			&workload.Status.OwnedResources, nil, resourcemanager.ResourceOptions{})
		if len(issues) == 0 {
			return nil, false
		}
		if !isConflict {
			return issues, false
		}

		attempt += 1
		log.Info("conflict while updating lws, retrying", "attempt", attempt, "workload", workload.Name)
	}

	// if we reached here, it means that the lws was not created successfully
	// if the error is a conflict error, we want to retry and create it in the next reconcile
	return issues, errors.IsConflict(issues[0].GetError())
}

// buildAndUpsertResources builds the resources and upserts them.
// it returns the issues and a boolean indicating if the reconciler should requeue the request
func (c *Controller) buildAndUpsertResources(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) ([]client.Object, []*issue.Issue, bool) {
	// build all resources using the distributed inference builder.
	// it is important to build the lws first, as we want to have it available even if one of the sub resources fails to build or upsert.
	// otherwise, status calculation may result in wrong phase, for example - stuck on Creating if due to a pvc error, the lws was not built.

	// save a copy of the resources to be upserted, because upsert resets the
	// object kind to be empty, but we rely on it for status calculation.
	previewResources := make([]client.Object, 0)

	// first build and upsert the lws
	log.V(4).Info("building lws", "workload", workload.Name)
	lws, err := c.builder.BuildLws(ctx, workload, workload)
	if err != nil {
		return nil, []*issue.Issue{issue2.NewIssue(workload, err)}, false
	}

	// copy before upserting
	previewResources = append(previewResources, lws.DeepCopy())

	log.V(4).Info("upserting lws", "workload", workload.Name)
	issues, isConflict := c.upsertLws(ctx, workload, lws)
	if len(issues) > 0 {
		return nil, issues, isConflict
	}

	// then, build and upsert the sub resources
	log.V(4).Info("building sub resources", "workload", workload.Name)
	objects, err := c.builder.BuildSubResources(ctx, workload, workload)
	if err != nil {
		return nil, []*issue.Issue{issue2.NewIssue(workload, err)}, false
	}

	// copy before upserting
	for _, obj := range objects {
		previewResources = append(previewResources, obj.DeepCopyObject().(client.Object))
	}

	log.V(4).Info("upserting workload resources", "workload", workload.Name)
	issues, isConflict = c.resourceManager.UpsertResources(ctx, workload.ObjectMeta, objects, &workload.Status.OwnedResources,
		nil, resourcemanager.ResourceOptions{})
	if len(issues) > 0 {
		return nil, issues, isConflict
	}

	return previewResources, nil, false
}

func (c *Controller) reconcileSucceeded(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
	originalWorkload *dinfw.DistributedInferenceWorkload, resources []client.Object) (ctrl.Result, error) {
	err := c.statusUpdater.Update(ctx, workload, originalWorkload, resources, nil)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (c *Controller) reconcileFailed(ctx context.Context, err error, isBuildError bool,
	workload *dinfw.DistributedInferenceWorkload, originalWorkload *dinfw.DistributedInferenceWorkload,
	previewResources []client.Object) (ctrl.Result, error) {
	log := logger.FromContext(ctx)
	log.Error(err, "reconcile failed with error")

	if previewResources == nil {
		var previewErr error
		previewResources, previewErr = c.builder.GetPreview(ctx, workload)
		if previewErr != nil {
			log.Error(previewErr, "failed to get preview resources", "id", workload.Name)
			// return the original error, as previewErr is just a consequence of it
			return ctrl.Result{}, err
		}
	}

	var buildErr *error
	if isBuildError {
		buildErr = &err
	}

	errUpdate := c.statusUpdater.Update(ctx, workload, originalWorkload, previewResources, buildErr)
	if errUpdate != nil {
		return ctrl.Result{}, errUpdate
	}

	// if the error is a build error, we do not want to requeue the request
	if isBuildError {
		err = nil
	}

	return ctrl.Result{}, err
}

func (c *Controller) setWorkloadLabelsAndAnnotations(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) error {
	log := logger.FromContext(ctx)
	log.V(4).Info("setting workload labels and annotations", "workload", workload.Name)

	clusterVersion := os.Getenv(workloadConstants.EnvVarClusterVersion)
	if clusterVersion == "" {
		return fmt.Errorf("cluster version is not set")
	}

	if workload.Annotations == nil {
		workload.Annotations = map[string]string{}
	}

	if workload.Labels == nil {
		workload.Labels = map[string]string{}
	}

	workload.Labels[constants.LabelWorkloadKind] = workload.Kind
	workload.Labels[constants.LabelWorkloadName] = workload.Name
	workload.Labels[constants.LabelRelease] = workload.Name
	workload.Labels[constants.RunaiProjectLabel] = ctx.Value(workloadproject.ContextProjectName).(string)
	workload.Labels[workloadConstants.CreatedAtClusterVersionLabel] = clusterVersion
	if value := workload.Spec.Preemptibility.Deref(); value != "" {
		workload.Labels[constants.LabelPreemptibilityName] = value
	}

	workload.Annotations[constants.AnnotationWorkloadApiVersion] = workload.APIVersion

	err := controller.PatchWorkloadLabelsAndAnnotations(ctx, c.k8sClient, workload, log)
	if err != nil {
		return err
	}

	// Add NIM workload label if using NIM image in leader spec
	leaderContainerFields := workload.LeaderSpec.ContainerFields
	if leaderContainerFields.Image != nil {
		controller.AddNimWorkloadLabel(leaderContainerFields.Labels, leaderContainerFields.Image.Deref())
	}

	return nil
}
