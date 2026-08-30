/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package run

import (
	"context"
	"fmt"
	"sync"

	kube_client "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/google/uuid"
	me "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/multi-error"
	"gomodules.xyz/jsonpatch/v2"
	"k8s.io/apimachinery/pkg/types"

	"os"
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	routev1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/installation/util"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/consts"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	cu "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

var (
	runaiNamespaceResources = []client.ObjectList{
		&v1.ConfigMapList{},
		&appsv1.DeploymentList{},
		&monitoringv1.PodMonitorList{},
		&monitoringv1.ServiceMonitorList{},
		&rbacv1.RoleList{},
		&rbacv1.RoleBindingList{},
		&v1.ServiceList{},
		&v1.ServiceAccountList{},
		&routev1.RouteList{},
		&appsv1.DaemonSetList{},
		&batchv1.JobList{},
		&monitoringv1.PrometheusList{},
		&monitoringv1.PrometheusRuleList{},
	}

	// slice of resources this controller should watch outside of runai's namespace.
	// The map keys specify the namespace, and the valuesa are lists of resources lists.
	foreignNamespacesResources = map[string][]client.ObjectList{
		consts.OpenshiftMonitoringNamespace: {
			&v1.ConfigMapList{},
		},
		"": { // Cluster scope
			// TODO @omeryahud temprorary until I can think of a better way to do it
			// This is needed in order to update our Prometheus instance to scrape metrics
			// from new namespaces created for Projects
			&v1.NamespaceList{},
		},
	}

	successRequeue = requeueAfter(30 * time.Second)
)

// RunAIConfigReconciler reconciles a RunAIConfig object
type RunAIConfigReconciler struct {
	runaiReconciler

	// localClusterDiscoveryClient is the discovery client for the local cluster
	localClusterDiscoveryClient discovery.DiscoveryInterface

	// remoteClusterDiscoveryClient is the dynamic client for the managed cluster, if ManagedClusterURL is empty, this will be the same as localClusterClient
	remoteClusterDiscoveryClient discovery.DiscoveryInterface

	// remoteClusterClient is the client for the managed cluster, if ManagedClusterURL is empty, this will be the same as localClusterClient
	remoteClusterClient kube_client.HybridClient

	// disableRunaiConfigFinalizer is a flag to disable the runai config finalizer
	disableRunaiConfigFinalizer bool
}

//+kubebuilder:rbac:groups=run.ai,resources=runaiconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=run.ai,resources=runaiconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=run.ai,resources=runaiconfigs/finalizers,verbs=update
//+kubebuilder:rbac:namespace:runai,groups="",resources=configmaps,verbs=get;list;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get

// NewRunAIConfigReconciler creates a new RunAIConfigReconciler
func NewRunAIConfigReconciler(localClusterClient kube_client.HybridClient,
	localClusterDiscoveryClient discovery.DiscoveryInterface,
	remoteClusterClient kube_client.HybridClient,
	remoteClusterDynamicClient discovery.DiscoveryInterface,
	verbosity int,
	disableRunaiConfigFinalizer bool) *RunAIConfigReconciler {
	reconciler := &RunAIConfigReconciler{
		runaiReconciler: runaiReconciler{
			localClusterClient: localClusterClient,
			scheme:             localClusterClient.Scheme(),
			verbosity:          verbosity,
		},
		localClusterDiscoveryClient:  localClusterDiscoveryClient,
		remoteClusterClient:          remoteClusterClient,
		remoteClusterDiscoveryClient: remoteClusterDynamicClient,
		disableRunaiConfigFinalizer:  disableRunaiConfigFinalizer,
	}

	return reconciler
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the RunAIConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *RunAIConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, retErr error) {
	logger := log.FromContext(ctx).V(r.verbosity)

	multiErr := &me.MultiError{}
	retErr = multiErr

	logger.Info("=========================== Reconcile Loop Started ===========================")

	start := time.Now()
	defer func() {
		// If the multi error is empty, return nil
		if multiErr.Empty() {
			retErr = nil
		}

		logger.Info(fmt.Sprintf("============================ Reconcile Loop Ended. Duration: %s ===========================", time.Since(start)))
	}()

	// Fetch the latest RunaiConfig from the cluster
	// * Be careful when passing a RunaiConfig to functions as they might modify it,
	// * unless necessary, pass a DeepCopy()
	logger.Info("fetching runaiconfig")
	runaiConfig, err := cu.GetRunaiConfigForRequest(r.localClusterClient, ctx, req)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("[%s/%s/%s] not found, probably deleted",
				req.Namespace, utils.ResourceStructType(runaiConfig), req.Name))
			return noRequeue(), nil
		}

		return requeue(), err
	}

	clusterResource, err := cu.GetClusterResource(r.localClusterClient, ctx)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("cluster resource not found, proceeding with RunaiConfig-only finalization")
		} else {
			logger.Error(err, "failed fetching cluster resource")
			multiErr.Append(err)
		}
	}

	res = requeue()
	defer func() {
		res, retErr = r.controllerDeferFunction(ctx, runaiConfig, multiErr, res)
		if clusterResource != nil {
			res, retErr = r.controllerDeferFunction(ctx, clusterResource, multiErr, res)
		}
	}()

	logger.Info(fmt.Sprintf("reconciling %s at resourceVersion [%s], generation [%d]",
		utils.ResourceUniqueID(runaiConfig), runaiConfig.GetResourceVersion(), runaiConfig.GetGeneration()))

	runaiConfigWithDefaultsSet := runaiConfig.DeepCopy()
	// Safe-guard to prevent consumers of runaiConfigWithDefaultsSet to update it
	runaiConfigWithDefaultsSet.SetResourceVersion("")
	runaiConfigWithDefaultsSet.SetDefaultsWhereNeeded()
	// TODO

	err = checkKubeVersions(r.localClusterDiscoveryClient, r.remoteClusterDiscoveryClient)
	if err != nil {
		logger.Error(err, "kubernetes version check failed")
		cu.SetCondition(runaiConfig, runaiv1.ConditionDependenciesFulfilled,
			metav1.ConditionFalse,
			runaiv1.DependenciesNotFulfilled,
			err.Error(), runaiConfig.Generation)
		if clusterResource != nil {
			cu.SetCondition(clusterResource, runaiv1.ConditionDependenciesFulfilled,
				metav1.ConditionFalse,
				runaiv1.DependenciesNotFulfilled,
				err.Error(), clusterResource.Generation)
		}
		return noRequeue(), err
	}

	logger.Info("handling dependency checks")
	err = handleDependencyChecks(ctx, r.remoteClusterClient, runaiConfig, runaiConfigWithDefaultsSet)
	if err != nil {
		logger.Error(err, "failed checking pre-installed dependencies")
		multiErr.Append(err)
	}

	cleanupChanged, cleanupErr := r.reconcileLegacyWorkloadCleanup(ctx, runaiConfig)
	if cleanupErr != nil {
		logger.Error(cleanupErr, "legacy workload cleanup reconciliation failed")
		multiErr.Append(cleanupErr)
		return res, retErr
	}
	runaiConfigWithDefaultsSet.Status.Conditions = append(
		[]metav1.Condition(nil), runaiConfig.Status.Conditions...,
	)
	cleanupCondition := runaiConfig.Status.GetCondition(runaiv1.ConditionTypeLegacyWorkloadCleanupComplete)
	cleanupWaiting := cleanupCondition != nil && cleanupCondition.Status == metav1.ConditionFalse &&
		cleanupCondition.Reason != string(runaiv1.LegacyWorkloadCleanupBlocked) &&
		cleanupCondition.Reason != string(runaiv1.LegacyWorkloadCleanupFailed)

	// Fetch all registered operands
	logger.Info("fetching operands")
	operands := NewManagedOperands(ctx, r.localClusterClient, r.remoteClusterClient, runaiConfigWithDefaultsSet)

	if r.disableRunaiConfigFinalizer {
		logger.Info("finalizer is disabled, removing finalizer if exists", "disableRunaiConfigFinalizer", r.disableRunaiConfigFinalizer)
	} else {
		logger.Info("finalizer is enabled, adding finalizer if missing", "disableRunaiConfigFinalizer", r.disableRunaiConfigFinalizer)
	}

	// Add or remove finalizer based on the disableRunaiConfigFinalizer flag
	if r.disableRunaiConfigFinalizer && cu.HasFinalizer(runaiConfig, runaiv1.Finalizer) {
		cu.RemoveFinalizer(runaiConfig, runaiv1.Finalizer)
		err := cu.PatchRunaiConfigFinalizers(ctx, r.localClusterClient, runaiConfig)
		if err != nil {
			multiErr.Append(err)
		} else {
			return noRequeue(), nil
		}
	} else if !r.disableRunaiConfigFinalizer && !cu.HasFinalizer(runaiConfig, runaiv1.Finalizer) {
		cu.AddFinalizer(runaiConfig, runaiv1.Finalizer)
		err := cu.PatchRunaiConfigFinalizers(ctx, r.localClusterClient, runaiConfig)
		if err != nil {
			multiErr.Append(err)
		} else {
			return noRequeue(), nil
		}
	}

	shouldFinalize := cu.ShouldFinalizeRunaiConfig(runaiConfig)
	if shouldFinalize {
		logger.Info("should finalize", "shouldFinalize", shouldFinalize)
	} else {
		logger.Info("should not finalize", "shouldFinalize", shouldFinalize)
	}

	logger.Info("reconciling local cluster")
	localClusterOperandStatuses, localClusterFinalized, err := r.reconcileLocalCluster(ctx,
		runaiConfigWithDefaultsSet, operands.GetLocalClusterOperands(), shouldFinalize)
	if err != nil {
		logger.Error(err, "local cluster reconciliation error")
		multiErr.Append(fmt.Errorf("local cluster reconciliation error: %v", err))
	}

	if localClusterFinalized {
		logger.Info("local cluster was finalized", "localClusterFinalized", localClusterFinalized)
	}

	setOwnerReferences := r.localClusterClient == r.remoteClusterClient
	if setOwnerReferences {
		logger.Info("setting owner references, remote and local clusters are the same", "setOwnerReferences", setOwnerReferences)
	} else {
		logger.Info("not setting owner references", "setOwnerReferences", setOwnerReferences)
	}

	logger.Info("reconciling remote cluster")

	remoteClusterOperandStatuses, remoteClusterFinalized, err := r.reconcileRemoteCluster(ctx,
		runaiConfigWithDefaultsSet, operands.GetRemoteClusterOperands(), shouldFinalize)
	if err != nil {
		logger.Error(err, "remote cluster reconciliation error")
		multiErr.Append(fmt.Errorf("remote cluster reconciliation error: %v", err))
	}

	if remoteClusterFinalized {
		logger.Info("remote cluster was finalized", "remoteClusterFinalized", remoteClusterFinalized)
	}

	if shouldFinalize {
		if localClusterFinalized && remoteClusterFinalized {
			cu.RemoveFinalizer(runaiConfig, runaiv1.Finalizer)
			err := cu.PatchRunaiConfigFinalizers(ctx, r.localClusterClient, runaiConfig)
			if err != nil {
				logger.Error(err, "failed removing finalizer", "namespace", runaiConfig.Namespace, "name", runaiConfig.Name)
				multiErr.Append(err)

				res = requeue()
				return
			}

			return noRequeue(), nil
		} else {
			logger.Error(nil, "not all clusters are finalized, requeuing")
			logger.Info("local cluster finalized", "localClusterFinalized", localClusterFinalized)
			logger.Info("remote cluster finalized", "remoteClusterFinalized", remoteClusterFinalized)
			multiErr.Append(fmt.Errorf("not all clusters are finalized: localClusterFinalized: %v, remoteClusterFinalized: %v", localClusterFinalized, remoteClusterFinalized))

			res = requeue()
			return
		}
	}

	if multiErr.Empty() {
		cu.SetCondition(runaiConfig, runaiv1.ConditionTypeDeployed,
			metav1.ConditionTrue, runaiv1.Deployed, "Resources Deployed", runaiConfig.Generation)
		if clusterResource != nil {
			cu.SetCondition(clusterResource, runaiv1.ConditionTypeDeployed,
				metav1.ConditionTrue, runaiv1.Deployed, "Resources Deployed", clusterResource.Generation)
		}
	} else {
		cu.SetCondition(runaiConfig, runaiv1.ConditionTypeDeployed,
			metav1.ConditionFalse, runaiv1.NotDeployed, multiErr.Error(), runaiConfig.Generation)
		if clusterResource != nil {
			cu.SetCondition(clusterResource, runaiv1.ConditionTypeDeployed,
				metav1.ConditionFalse, runaiv1.NotDeployed, multiErr.Error(), clusterResource.Generation)
		}
	}

	logger.Info("setting operand statuses")
	operandStatusMap := make(map[string]runaiv1.TimedOperandStatus)
	for name, status := range localClusterOperandStatuses {
		operandStatusMap[name] = status
	}

	for name, status := range remoteClusterOperandStatuses {
		operandStatusMap[name] = status
	}

	runaiConfig.Status.OperandStatuses = operandStatusMap

	ready, message := calculateAggregatedAvailableStatusReason(operandStatusMap)
	if !ready {
		logger.Error(nil, "not all operands are ready", "message", message)
		multiErr.Append(fmt.Errorf("%s", message))
		multiErr.Append(fmt.Errorf("reconciling due to unavailable condition"))

		cu.SetCondition(runaiConfig, runaiv1.ConditionTypeAvailable,
			metav1.ConditionFalse, runaiv1.NotAvailable, multiErr.Error(), runaiConfig.Generation)
		if clusterResource != nil {
			cu.SetCondition(clusterResource, runaiv1.ConditionTypeAvailable,
				metav1.ConditionFalse, runaiv1.NotAvailable, multiErr.Error(), clusterResource.Generation)
		}
	} else {
		cu.SetCondition(runaiConfig, runaiv1.ConditionTypeAvailable,
			metav1.ConditionTrue, runaiv1.Available, "System Available", runaiConfig.Generation)
		if clusterResource != nil {
			cu.SetCondition(clusterResource, runaiv1.ConditionTypeAvailable,
				metav1.ConditionTrue, runaiv1.Available, "System Available", clusterResource.Generation)
		}

		runaiConfig.Status.RemoteConfig = runaiConfig.Spec.Internal.RemoteConfig
		res = successRequeue
	}

	logger.Info("Setting metrics")
	setMetrics(runaiConfigWithDefaultsSet, operandStatusMap)

	if !multiErr.Empty() {
		res = requeue()
		return
	}
	// Non-creating cleanup transitions do not block the normal target-version
	// operand reconcile. Persist their state with the normal status update.
	if cleanupChanged || cleanupErr != nil {
		res = requeue()
	} else if cleanupWaiting {
		res = requeueAfter(5 * time.Second)
	}

	return
}

func (r *RunAIConfigReconciler) reconcileRemoteCluster(ctx context.Context, runaiConfig *runaiv1.RunaiConfig,
	ops []operands.Operand, finalize bool) (map[string]runaiv1.TimedOperandStatus, bool, error) {
	return reconcileCluster(ctx, r.remoteClusterClient, r.scheme, runaiConfig, ops, false, finalize)
}

func (r *RunAIConfigReconciler) reconcileLocalCluster(ctx context.Context, runaiConfig *runaiv1.RunaiConfig,
	ops []operands.Operand, finalize bool) (map[string]runaiv1.TimedOperandStatus, bool, error) {
	return reconcileCluster(ctx, r.localClusterClient, r.scheme, runaiConfig, ops, true, finalize)
}

func addWatchesToBuilder(builder *ctrl.Builder, owns []client.Object, watches []client.Object) {
	for _, resource := range owns {
		builder.Owns(resource)
	}

	for _, resource := range watches {
		builder.Watches(resource, handler.EnqueueRequestsFromMapFunc(mapResourceToRunaiConfigEvent))
	}
}

func ownedResources(isOpenshift bool) []client.Object {
	resources := []client.Object{
		&v1.ConfigMap{},
		&appsv1.Deployment{},
		&monitoringv1.PodMonitor{},
		&monitoringv1.ServiceMonitor{},
		&rbacv1.Role{},
		&rbacv1.RoleBinding{},
		&v1.Service{},
		&v1.ServiceAccount{},
		&appsv1.DaemonSet{},
		&batchv1.Job{},
		&monitoringv1.Prometheus{},
		&monitoringv1.PrometheusRule{},
	}

	if isOpenshift {
		resources = append(resources, &routev1.Route{})
	}

	return resources
}

func watchedResources(vapSupported bool) []client.Object {
	resources := []client.Object{
		&v1.ConfigMap{},
		&v1.Namespace{},
	}

	if vapSupported {
		resources = append(resources,
			&admissionv1.ValidatingAdmissionPolicy{},
			&admissionv1.ValidatingAdmissionPolicyBinding{})
	}

	return resources
}

// SetupWithManager sets up the controller with the Manager.
func (r *RunAIConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(&runaiv1.RunaiConfig{})

	isOpenshift, err := util.IsOpenshift(context.Background(), r.localClusterClient.Uncached())
	if err != nil {
		return fmt.Errorf("runai config reconciler setup: failed to determine if running on openshift or not: %w", err)
	}

	vapSupported, err := isVAPSupported(r.localClusterDiscoveryClient)
	if err != nil {
		return fmt.Errorf("runai config reconciler setup: failed to check ValidatingAdmissionPolicy support: %w", err)
	}

	addWatchesToBuilder(builder, ownedResources(isOpenshift), watchedResources(vapSupported))

	builder.WithEventFilter(&reconciliationPredicate{})
	builder.WithOptions(controller.Options{
		RateLimiter:             workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](time.Second, 10*time.Second),
		MaxConcurrentReconciles: 1,
	})

	return builder.Complete(r)
}

func watchOperandResources(ctx context.Context, builder *ctrl.Builder, mgr ctrl.Manager,
	watchFunction func(builder *ctrl.Builder, obj client.Object)) error {
	setupClient := kube_client.New(mgr.GetClient(),
		mgr.GetAPIReader(),
		mgr.GetConfig())
	// Slice of resources that exist for the k8s flavor the controller is installed on
	registeredResourcesWithinRunaiConfigNamespace, err :=
		filterOutUnregisteredResources(setupClient, runaiNamespaceResources)
	if err != nil {
		return err
	}

	for _, resource := range registeredResourcesWithinRunaiConfigNamespace {
		watchFunction(builder, resource)
	}

	setupLog := log.FromContext(ctx)
	setupLog.Info("setting up watchers for foreign namespaces")

	for namespace, resources := range foreignNamespacesResources {
		filteredResources, err := filterOutUnregisteredResourcesLists(setupClient, resources)
		if err != nil {
			return err
		}

		for _, resource := range filteredResources {
			err = watchResourcesInForeignNamespace(namespace, resource, mgr.GetScheme(), mgr.GetConfig(), mgr.GetCache(), ctx)
			if err != nil {
				setupLog.Error(err,
					fmt.Sprintf("failed setting up watchers in foreign namespace: %s", namespace))
				os.Exit(1)
			}

			setupLog.Info(fmt.Sprintf("watcher started for %s in the %s namespace",
				reflect.TypeOf(resource).Elem().Name(), namespace))
		}
	}

	return nil
}

func filterOutUnregisteredResourcesLists(k8sClient client.Client,
	resources []client.ObjectList) ([]client.ObjectList, error) {
	registeredResourceLists := []client.ObjectList{}

	for _, list := range resources {
		err := k8sClient.List(context.Background(), list, &client.ListOptions{
			Namespace: runaiv1.DefaultNamespace,
		})
		// Do not watch resources that are not registered with the API server.
		// Ignoring the error here will allow watching all other resources that are
		// registered with the API server.
		// For resources that are not registered with the API server, but should still
		// be watched (dependencies, for example cert-manager),
		// a reconciliation error will occur that would let us know about it
		if err == nil || errors.IsNotFound(err) || errors.IsUnauthorized(err) {
			registeredResourceLists = append(registeredResourceLists, list)
		} else if !meta.IsNoMatchError(err) {
			return nil, err
		}
	}

	return registeredResourceLists, nil
}

func filterOutUnregisteredResources(k8sClient client.Client,
	resources []client.ObjectList) ([]client.Object, error) {
	registeredResourcesLists, err := filterOutUnregisteredResourcesLists(k8sClient, resources)
	if err != nil {
		return nil, err
	}

	registeredResources := []client.Object{}

	for _, list := range registeredResourcesLists {
		itemsField, exists := reflect.TypeOf(list).Elem().FieldByName("Items")
		if !exists {
			return nil, fmt.Errorf("resource list of type %s has no 'Items' field",
				reflect.TypeOf(list).Elem().Name())
		}

		objectType := itemsField.Type.Elem()

		if objectType.Kind() == reflect.Ptr {
			objectType = objectType.Elem()
		}

		objectFromList, ok := reflect.New(objectType).Interface().(client.Object)
		if !ok {
			return nil, fmt.Errorf("could not infer resource kind to watch from list %s",
				reflect.TypeOf(list).Elem().Name())
		}

		registeredResources = append(registeredResources, objectFromList)
	}

	return registeredResources, nil
}

func watchResourcesInForeignNamespace(namespace string, objList client.ObjectList,
	scheme *runtime.Scheme, config *rest.Config, clientCache cache.Cache, ctx context.Context) error {
	setupLog := log.FromContext(ctx, "namespace", namespace)
	watchClient, err := client.NewWithWatch(config, client.Options{
		Scheme: scheme,
		Cache: &client.CacheOptions{
			Reader: clientCache,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to set up sibling namespace watch client: %v", err)
	}

	if namespace != "" { // Only check for namespace existence if a namespace is specified
		err = watchClient.Get(ctx, types.NamespacedName{Name: namespace}, &v1.Namespace{})
		if err != nil {
			if errors.IsNotFound(err) {
				setupLog.Info("namespace does not exist, forfeiting watch", "namespace", namespace)
				return nil
			}

			return err
		}
	}

	w, err := watchClient.Watch(ctx, objList, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		if meta.IsNoMatchError(err) {
			setupLog.Info(fmt.Sprintf("kind %s is not registered with the api server, skipping it",
				reflect.TypeOf(objList).Elem().Name()), "error", err)
			return nil
		} else {
			return fmt.Errorf("failed setting up watch %v", err)
		}
	}

	go handleResource(ctx, watchClient, w, namespace, reflect.TypeOf(objList).Elem().Name())

	return nil
}

func handleResource(ctx context.Context, k8s client.WithWatch, w watch.Interface, namespace, resourceKind string) {
	for event := range w.ResultChan() {
		logger := log.FromContext(ctx, "namespace", namespace, "kind", resourceKind)
		if event.Object == nil {
			// If event.Object is nil, we can assume that for some reason the
			// result channel was closed, and the watcher has died.
			// This is safe to assume because a zero value was received from it,
			// so we panic to restart our watchers when we recover
			logger.Error(fmt.Errorf("watcher has died, exiting in order to restart and recover"), "")
			os.Exit(1)
		}

		shouldTriggerReconciliation := cu.IsOwnerOf(metav1.GroupKind{
			Group: runaiv1.GroupVersion.Group,
			Kind:  runaiv1.RunaiConfigKind,
		}, runaiv1.DefaultNamespace, runaiv1.DefaultName, event.Object.(metav1.Object).GetAnnotations())

		if !shouldTriggerReconciliation {
			continue
		}

		runaiConfigs := &runaiv1.RunaiConfigList{}

		for {
			err := k8s.List(ctx, runaiConfigs)
			if err == nil {
				break
			}

			logger.Error(err, "failed listing RunaiConfigs in order to reconcile them")
			continue
		}

		for i := 0; i < len(runaiConfigs.Items); {
			// Move on to the next RunaiConfig only if reconciliation was triggered for the current one
			err := triggerReconciliationForRunaiConfigUsingConfigMap(ctx, k8s, &runaiConfigs.Items[i])
			if err == nil {
				logger.Info("triggered reconciliation for RunaiConfig",
					"name", runaiConfigs.Items[i].Name,
					"namespace", runaiConfigs.Items[i].Namespace,
					"resourceName", event.Object.(metav1.Object).GetName(),
					"resourceNamespace", event.Object.(metav1.Object).GetNamespace())
				i++
			} else {
				logger.Error(err, "failed triggering reconciliation for RunaiConfig",
					"name", runaiConfigs.Items[i].Name,
					"namespace", runaiConfigs.Items[i].Namespace,
					"resourceName", event.Object.(metav1.Object).GetName(),
					"resourceNamespace", event.Object.(metav1.Object).GetNamespace())
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func triggerReconciliationForRunaiConfigUsingConfigMap(ctx context.Context, k8s client.Client,
	runaiConfig *runaiv1.RunaiConfig) error {
	reconciliationTriggerConfigMap := &v1.ConfigMap{}
	err := k8s.Get(ctx, types.NamespacedName{
		Namespace: runaiConfig.Namespace,
		Name: fmt.Sprintf("%s-%s",
			runaiConfig.Name, runaiv1.ReconciliationTriggerConfigmapNameSuffix),
	}, reconciliationTriggerConfigMap)
	if err != nil {
		return err
	}

	reconciliationTriggerConfigMap.Annotations = map[string]string{
		runaiv1.ReconciliationTriggerRequestKey: uuid.New().String(),
	}

	patch := jsonpatch.NewOperation("add", "/metadata/annotations",
		reconciliationTriggerConfigMap.Annotations)

	return k8s.Patch(ctx, reconciliationTriggerConfigMap, client.RawPatch(types.JSONPatchType,
		[]byte(fmt.Sprintf("[%s]", patch.Json()))))
}

func reconcileCluster(ctx context.Context, k8s kube_client.HybridClient, scheme *runtime.Scheme,
	runaiConfig *runaiv1.RunaiConfig,
	ops []operands.Operand,
	ownResources bool, finalize bool) (map[string]runaiv1.TimedOperandStatus, bool, error) {
	merr := me.New()
	logger := log.FromContext(ctx)

	resources, err := ResourcesForOperands(ctx, ops)
	if err != nil {
		merr.Append(fmt.Errorf("resource gathering error: %v", err))
	}

	if finalize {
		err := cu.Finalize(ctx, k8s.Uncached(), append(resources.CreateOrUpdate, resources.Delete...), logger)
		if err != nil {
			merr.Append(fmt.Errorf("finalizer handling error: %v", err))
			return nil, false, merr.ErrorOrNil()
		}

		// Finalizer was completed, everything should be removed, nothing else to do
		return nil, true, nil
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		clusterResource := &runaiv1.Cluster{}
		err = k8s.Get(ctx, types.NamespacedName{
			Name: runaiv1.ClusterName,
		}, clusterResource)
		if err != nil {
			merr.Append(fmt.Errorf("could not get cluster resource: %v", err))
		}

		err = cu.CreateOrUpdateResources(k8s, ctx, resources.CreateOrUpdate, runaiConfig, clusterResource, ownResources)
		if err != nil {
			merr.Append(fmt.Errorf("could not create or update resources: %v", err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err = cu.DeleteResources(k8s, ctx, resources.Delete)
		if err != nil {
			merr.Append(fmt.Errorf("could not delete resources: %v", err))
		}
	}()

	operandStatusMap, err := calculateOperandStatusMap(ctx, runaiConfig, ops)
	if err != nil {
		merr.Append(fmt.Errorf("operand status calculation error: %v", err))
	}

	wg.Wait()

	return operandStatusMap, false, merr.ErrorOrNil()
}

func handleDependencyChecks(ctx context.Context, k8s kube_client.HybridClient,
	clusterRunaiconfig *runaiv1.RunaiConfig, runaiConfigWithDefaults *runaiv1.RunaiConfig) error {
	logger := log.FromContext(ctx)

	if runaiConfigWithDefaults.ShouldSkipDependencyChecks() {
		logger.Info("Dependency checks skipped")
		cu.SetCondition(clusterRunaiconfig, runaiv1.ConditionDependenciesFulfilled,
			metav1.ConditionTrue,
			runaiv1.DependenciesFulfilled,
			"Dependency checks skipped",
			clusterRunaiconfig.Generation)
		return nil
	} else {
		err := checkDependencies(ctx, k8s, runaiConfigWithDefaults)
		if err != nil {
			logger.Error(err, "failed checking pre-installed dependencies")
			cu.SetCondition(clusterRunaiconfig, runaiv1.ConditionDependenciesFulfilled,
				metav1.ConditionFalse, runaiv1.DependenciesNotFulfilled, err.Error(), clusterRunaiconfig.Generation)
			return err
		} else {
			logger.Info("dependencies are fulfilled")
			cu.SetCondition(clusterRunaiconfig, runaiv1.ConditionDependenciesFulfilled,
				metav1.ConditionTrue, runaiv1.DependenciesFulfilled, "Dependencies are fulfilled", clusterRunaiconfig.Generation)
			return nil
		}
	}
}
