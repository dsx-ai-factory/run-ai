/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networkstatus

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconsts "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	infv1 "sigs.k8s.io/gateway-api-inference-extension/api/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// Task queue rate limiter configuration.
// The queue uses a combined rate limiter: exponential backoff for failed tasks (1ms base, 10s max)
// and a token bucket limiter (30 ops/sec with bursts up to 90) for overall throughput control.
const (
	taskHandlingMaxRetryAttempts = 10
	rateLimiterBaseDelay         = time.Millisecond
	rateLimiterMaxDelay          = time.Second * 10
	rateLimit                    = 30
	rateBursts                   = 90
)

// NetworkStatusUpdater discovers and patches ExternalWorkload network endpoints.
// It owns a rate-limited task queue — tasks are enqueued via UpdateNetworkStatusIfNeeded
// (pod ADD events) or network resource watchers (Service/Ingress/Route events)
// and processed by ExecuteTasks in a separate goroutine.
type NetworkStatusUpdater struct {
	cachedClient client.Client
	tasksQueue   workqueue.TypedRateLimitingInterface[utils.ExternalWorkloadTask]
}

// NewNetworkStatusUpdater creates a new updater with its own rate-limited task queue.
func NewNetworkStatusUpdater(cachedClient client.Client) *NetworkStatusUpdater {
	return &NetworkStatusUpdater{
		cachedClient: cachedClient,
		tasksQueue: workqueue.NewTypedRateLimitingQueueWithConfig(
			workqueue.NewTypedMaxOfRateLimiter(
				workqueue.NewTypedItemExponentialFailureRateLimiter[utils.ExternalWorkloadTask](rateLimiterBaseDelay, rateLimiterMaxDelay),
				&workqueue.TypedBucketRateLimiter[utils.ExternalWorkloadTask]{Limiter: rate.NewLimiter(rate.Limit(rateLimit), rateBursts)},
			), workqueue.TypedRateLimitingQueueConfig[utils.ExternalWorkloadTask]{Name: fmt.Sprintf("%s-network-status", config.Config().ServiceName),
				MetricsProvider: metrics.PrometheusMetricsProviderInstance}),
	}
}

// TasksQueue returns the rate-limited work queue for network status recomputation tasks.
func (n *NetworkStatusUpdater) TasksQueue() workqueue.TypedRateLimitingInterface[utils.ExternalWorkloadTask] {
	return n.tasksQueue
}

// ExecuteTasks loops processing items from the task queue until the queue is shut down.
func (n *NetworkStatusUpdater) ExecuteTasks(ctx context.Context) {
	for n.processNextItem(ctx) {
	}
}

func (n *NetworkStatusUpdater) processNextItem(ctx context.Context) bool {
	queue := n.tasksQueue
	task, quit := queue.Get()
	if quit {
		log.Info().Msgf("NetworkStatusUpdater: task queue shut down, stopping")
		return false
	}
	defer queue.Done(task)

	log.Debug().Msgf("NetworkStatusUpdater: processing task <%s> for <%s> key <%s>",
		task.GetOperation(), task.GVK.Kind, task.Key.String())

	n.handleTask(ctx, task)

	log.Debug().Msgf("NetworkStatusUpdater: completed task <%s> for <%s> key <%s>",
		task.GetOperation(), task.GVK.Kind, task.Key.String())
	return true
}

func (n *NetworkStatusUpdater) handleTask(ctx context.Context, task utils.ExternalWorkloadTask) {
	queue := n.tasksQueue

	err := n.HandleNetworkResourceChange(ctx, task.Object, task.GVK)
	if err == nil || queue.NumRequeues(task) > taskHandlingMaxRetryAttempts {
		// Success or max retries exceeded — remove the task from the rate limiter's failure tracking
		queue.Forget(task)
	} else {
		// Transient failure — re-enqueue with exponential backoff for another attempt.
		queue.AddRateLimited(task)
	}
}

// HandleNetworkResourceChange handles ADD/UPDATE/DELETE events for network resources.
// It resolves the changed resource into affected Pods, then recomputes endpoints
// for all ExternalWorkloads that own those Pods.
func (n *NetworkStatusUpdater) HandleNetworkResourceChange(ctx context.Context, obj client.Object, gvk schema.GroupVersionKind) error {
	log.Debug().Msgf("NetworkStatusUpdater: handling %s add/update/delete for <%s/%s>", gvk.Kind, obj.GetNamespace(), obj.GetName())

	ew, ok := obj.(*v1alpha1.ExternalWorkload)
	if !ok {
		log.Warn().Msgf("NetworkStatusUpdater: expected ExternalWorkload object but got %T", obj)
		return nil
	}
	pods, err := utils.ListWorkloadPods(ctx, n.cachedClient, ew)
	if err != nil {
		log.Error().Msgf("NetworkStatusUpdater: failed listing pods for ExternalWorkload <%s/%s>, error: %s",
			ew.Namespace, ew.Name, err.Error())
		return nil
	}
	return n.recomputeExternalWorkloadEndpoints(ctx, ew, pods)
}

// UpdateNetworkStatusIfNeeded is called by ExternalWorkloadHandler on pod ADD events.
// Instead of computing endpoints synchronously, it enqueues the ExternalWorkload as a task
// on the task queue. ExecuteTasks will pick it up, find the EW's
// pods, discover all endpoints, and patch the EW status.
// Skips enqueue when addedResource is false (not a new pod) or externalWorkload is nil.
func (n *NetworkStatusUpdater) UpdateNetworkStatusIfNeeded(
	_ context.Context, _ client.Object, externalWorkload *v1alpha1.ExternalWorkload, addedResource bool) error {

	if !addedResource || externalWorkload == nil {
		return nil
	}

	key := types.NamespacedName{Name: externalWorkload.Name, Namespace: externalWorkload.Namespace}
	// DeepCopy so the consumer can't mutate (or observe partial mutations of)
	// the caller's ExternalWorkload pointer while the caller is still updating it.
	task := utils.NewExternalWorkloadTask(utils.AddOperation, key, externalWorkload.GroupVersionKind()).SetObject(externalWorkload.DeepCopy())
	n.tasksQueue.Add(*task)
	return nil
}

func (n *NetworkStatusUpdater) FindAffectedEwsForObject(ctx context.Context, obj client.Object, gvk schema.GroupVersionKind) []*v1alpha1.ExternalWorkload {
	log.Debug().Msgf("NetworkStatusUpdater: handling %s add/update/delete for <%s/%s>", gvk.Kind, obj.GetNamespace(), obj.GetName())

	pods := n.resolveAffectedPods(ctx, obj, gvk)
	if len(pods) == 0 {
		return nil
	}
	return n.findAffectedExternalWorkloadsFromPods(ctx, pods)
}

// resolveAffectedPods resolves a network resource change into the list of Pods
// that may be affected. Each GVK kind has its own resolution strategy:
//   - Service: Pods matching the service's selector
//   - InferencePool: Pods matching the pool's selector
//   - Ingress/Route: Pods matching selectors of referenced services (and InferencePools for HTTPRoute)
func (n *NetworkStatusUpdater) resolveAffectedPods(ctx context.Context, obj client.Object, gvk schema.GroupVersionKind) []corev1.Pod {

	switch gvk.Kind {
	case kindService:
		svc, ok := obj.(*corev1.Service)
		if !ok {
			log.Warn().Msgf("NetworkStatusUpdater: expected Service object but got %T", obj)
			return nil
		}
		pods, err := n.FindPodsForService(ctx, svc)
		if err != nil {
			log.Error().Msgf("NetworkStatusUpdater: failed listing Pods for service <%s/%s>: %s",
				svc.Namespace, svc.Name, err.Error())
			return nil
		}
		return pods

	case kindIngress, kindGRPCRoute, kindTLSRoute:
		return n.resolvePodsFromRouteOrIngressSvc(ctx, obj, gvk)

	case kindHTTPRoute:
		pods := n.resolvePodsFromRouteOrIngressSvc(ctx, obj, gvk)
		return append(pods, n.resolvePodsFromInferencePoolOrRoute(ctx, obj, gvk)...)

	case kindInferencePool:
		return n.resolvePodsFromInferencePoolOrRoute(ctx, obj, gvk)

	case kindRoute:
		return n.resolvePodsOpenshiftRoute(ctx, obj)
	}
	return nil
}

func (n *NetworkStatusUpdater) getInferencePoolsFromObj(ctx context.Context, obj client.Object, gvk schema.GroupVersionKind) []infv1.InferencePool {
	var pools []infv1.InferencePool

	switch gvk.Kind {
	case kindInferencePool:
		pool, ok := obj.(*infv1.InferencePool)
		if !ok {
			log.Warn().Msgf("NetworkStatusUpdater: expected InferencePool object but got %T", obj)
			return nil
		}
		pools = append(pools, *pool)

	case kindHTTPRoute:
		// Also resolve pods from InferencePools referenced by the HTTPRoute's backendRefs.
		route, ok := obj.(*gwv1.HTTPRoute)
		if !ok {
			log.Warn().Msgf("NetworkStatusUpdater: expected HTTPRoute object but got %T", obj)
			return nil
		}
		poolNames := extractInferencePoolNamesFromHTTPRoute(route)
		for _, name := range poolNames {
			pool := &infv1.InferencePool{}
			if err := n.cachedClient.Get(ctx, types.NamespacedName{Name: name, Namespace: obj.GetNamespace()}, pool); err != nil {
				log.Debug().Msgf("NetworkStatusUpdater: failed to get InferencePool <%s/%s>, skipping: %s", obj.GetNamespace(), name, err.Error())
				continue
			}
			pools = append(pools, *pool)
		}
	}
	return pools
}

// resolvePodsFromInferencePoolOrRoute resolves pods affected by an InferencePool or HTTPRoute change.
// Returns pods matching the selectors of referenced InferencePools.
func (n *NetworkStatusUpdater) resolvePodsFromInferencePoolOrRoute(ctx context.Context, obj client.Object, gvk schema.GroupVersionKind) []corev1.Pod {
	inferencePools := n.getInferencePoolsFromObj(ctx, obj, gvk)
	if len(inferencePools) == 0 {
		return nil
	}

	seen := map[types.UID]bool{}
	var allPods []corev1.Pod

	for i := range inferencePools {
		pool := &inferencePools[i]
		if pool.Spec.Selector.MatchLabels == nil {
			continue
		}
		selector := make(map[string]string, len(pool.Spec.Selector.MatchLabels))
		for k, v := range pool.Spec.Selector.MatchLabels {
			selector[string(k)] = string(v)
		}
		podList := &corev1.PodList{}
		if err := n.cachedClient.List(ctx, podList, client.InNamespace(pool.Namespace), client.MatchingLabels(selector)); err != nil {
			log.Error().Msgf("NetworkStatusUpdater: failed listing Pods for InferencePool <%s/%s>: %s",
				pool.Namespace, pool.Name, err.Error())
			continue
		}
		for _, pod := range podList.Items {
			if !seen[pod.UID] {
				seen[pod.UID] = true
				allPods = append(allPods, pod)
			}
		}
	}

	return allPods
}

// resolvePodsOpenshiftRoute resolves pods targeted by an OpenShift Route via its backend services.
func (n *NetworkStatusUpdater) resolvePodsOpenshiftRoute(ctx context.Context, obj client.Object) []corev1.Pod {
	route, ok := obj.(*openshiftV1.Route)
	if !ok {
		log.Warn().Msgf("NetworkStatusUpdater: expected Route object but got %T", obj)
		return nil
	}

	serviceNames := extractServiceNamesFromOpenshiftRoute(route)
	seen := map[types.UID]bool{}
	var allPods []corev1.Pod
	for _, svcName := range serviceNames {
		svc := &corev1.Service{}
		if err := n.cachedClient.Get(ctx, types.NamespacedName{Name: svcName, Namespace: route.Namespace}, svc); err != nil {
			log.Debug().Msgf("NetworkStatusUpdater: service <%s/%s> not found, skipping: %s", route.Namespace, svcName, err.Error())
			continue
		}
		pods, err := n.FindPodsForService(ctx, svc)
		if err != nil {
			log.Error().Msgf("NetworkStatusUpdater: failed listing Pods for service <%s/%s>: %s", svc.Namespace, svc.Name, err.Error())
			continue
		}
		for _, pod := range pods {
			if !seen[pod.UID] {
				seen[pod.UID] = true
				allPods = append(allPods, pod)
			}
		}
	}
	return allPods
}

// resolvePodsFromRouteOrIngressSvc extracts services (and InferencePools for HTTPRoutes) referenced
// by a route or ingress, then returns all Pods matching those selectors.
func (n *NetworkStatusUpdater) resolvePodsFromRouteOrIngressSvc(ctx context.Context, obj client.Object, gvk schema.GroupVersionKind) []corev1.Pod {
	seen := map[types.UID]bool{}
	var allPods []corev1.Pod

	serviceNames := n.extractServiceNames(ctx, obj, gvk)
	for _, svcName := range serviceNames {
		svc := &corev1.Service{}
		if err := n.cachedClient.Get(ctx, types.NamespacedName{Name: svcName, Namespace: obj.GetNamespace()}, svc); err != nil {
			log.Debug().Msgf("NetworkStatusUpdater: service <%s/%s> not found, skipping: %s", obj.GetNamespace(), svcName, err.Error())
			continue
		}
		pods, err := n.FindPodsForService(ctx, svc)
		if err != nil {
			log.Error().Msgf("NetworkStatusUpdater: failed listing Pods for service <%s/%s>: %s",
				svc.Namespace, svc.Name, err.Error())
			continue
		}
		for _, pod := range pods {
			if !seen[pod.UID] {
				seen[pod.UID] = true
				allPods = append(allPods, pod)
			}
		}
	}

	return allPods
}

// extractServiceNames returns the service names referenced by an Ingress or Route resource.
// Returns nil for Service objects (caller should handle them directly).
func (n *NetworkStatusUpdater) extractServiceNames(_ context.Context, obj client.Object, gvk schema.GroupVersionKind) []string {
	switch gvk.Kind {
	case kindIngress:
		if ing, ok := obj.(*networkingv1.Ingress); ok {
			return extractServiceNamesFromIngress(ing)
		}
	case kindHTTPRoute:
		if route, ok := obj.(*gwv1.HTTPRoute); ok {
			return extractServiceNamesFromHTTPRoute(route)
		}
	case kindGRPCRoute:
		if route, ok := obj.(*gwv1.GRPCRoute); ok {
			return extractServiceNamesFromGRPCRoute(route)
		}
	case kindTLSRoute:
		if route, ok := obj.(*gwv1alpha2.TLSRoute); ok {
			return extractServiceNamesFromTLSRoute(route)
		}
	case kindRoute:
		if route, ok := obj.(*openshiftV1.Route); ok {
			return extractServiceNamesFromOpenshiftRoute(route)
		}
	}
	return []string{}
}

// recomputeExternalWorkloadEndpoints discovers endpoints for the EW's pods
// and patches the EW status if changed.
func (n *NetworkStatusUpdater) recomputeExternalWorkloadEndpoints(ctx context.Context, externalWorkload *v1alpha1.ExternalWorkload, pods []corev1.Pod) error {
	if err := n.reconcileDistributedInferenceInternalURL(ctx, externalWorkload); err != nil {
		log.Error().Msgf("NetworkStatusUpdater: failed reconciling internal URL for ExternalWorkload <%s/%s>: %s",
			externalWorkload.Namespace, externalWorkload.Name, err.Error())
		return err
	}
	discoveredEndpoints, err := n.DiscoverEndpointsForEwAndPods(ctx, externalWorkload, pods)
	if err != nil {
		log.Error().Msgf("NetworkStatusUpdater: failed discovering endpoints for ExternalWorkload <%s/%s>: %s",
			externalWorkload.Namespace, externalWorkload.Name, err.Error())
		return err
	}
	if err := n.ReplaceEndpointsAndPatch(ctx, externalWorkload, deduplicateEndpoints(discoveredEndpoints)); err != nil {
		if errors.IsNotFound(err) {
			log.Debug().Msgf("NetworkStatusUpdater: ExternalWorkload <%s/%s> not found, skipping patch",
				externalWorkload.Namespace, externalWorkload.Name)
			return nil
		}
		log.Error().Msgf("NetworkStatusUpdater: failed patching endpoints for EW <%s/%s>: %s",
			externalWorkload.Namespace, externalWorkload.Name, err.Error())
		return err
	}
	return nil
}

// findAffectedExternalWorkloadsFromPods finds all ExternalWorkloads for Pods.
// Only Pods whose top-owner already has an EW are included — if the EW doesn't exist yet,
// the pod's endpoints will be calculated when the pod is added to ExternalWorkload via the pod ADD flow.
func (n *NetworkStatusUpdater) findAffectedExternalWorkloadsFromPods(ctx context.Context, pods []corev1.Pod) []*v1alpha1.ExternalWorkload {
	ewById := map[string]*v1alpha1.ExternalWorkload{}
	for i := range pods {
		pod := &pods[i]
		topOwnerUid, found := pod.GetLabels()[commonconsts.ClusterTopOwnerUidLabel]
		if !found || topOwnerUid == "" {
			log.Debug().Msgf("NetworkStatusUpdater: no workload id set in pod <%s/%s>, pod will be handled once added",
				pod.Namespace, pod.Name)
			continue
		}
		entry, exists := ewById[topOwnerUid]
		if !exists || entry == nil {
			ew, err := utils.GetExternalWorkloadOfObject(ctx, topOwnerUid, pod.Namespace, n.cachedClient)
			if err != nil {
				log.Warn().Msgf("NetworkStatusUpdater: Failed to get external workload for pod <%s/%s>: %s",
					pod.Namespace, pod.Name, err.Error())
				continue
			}
			if ew == nil {
				// EW doesn't exist yet for this pod's workload — endpoints will be
				// calculated when the pod is added to the EW via the normal flow.
				continue
			}
			ewById[topOwnerUid] = ew
		}
	}
	result := make([]*v1alpha1.ExternalWorkload, 0, len(ewById))
	for _, entry := range ewById {
		result = append(result, entry)
	}
	return result
}

// deduplicateEndpoints removes duplicate endpoints by URL.
func deduplicateEndpoints(eps []v1alpha1.ExternalEndpoint) []v1alpha1.ExternalEndpoint {
	seen := make(map[string]bool, len(eps))
	result := make([]v1alpha1.ExternalEndpoint, 0, len(eps))
	for _, ep := range eps {
		if !seen[ep.Url] {
			seen[ep.Url] = true
			result = append(result, ep)
		}
	}
	return result
}

// sortEndpoints sorts a slice of endpoints by their key for deterministic comparison.
func sortEndpoints(eps []v1alpha1.ExternalEndpoint) {
	sort.Slice(eps, func(i, j int) bool {
		return eps[i].Url < eps[j].Url
	})
}

func updateNetworkStatusPatchBytes(networkStatus *v1alpha1.WorkloadNetworkStatus) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"status": map[string]interface{}{
		"networkStatus": map[string]interface{}{
			"externalEndpoints": networkStatus.ExternalEndpoints}}})
}

// ReplaceEndpointsAndPatch fully replaces the ExternalWorkload's external endpoints and patches if changed.
func (n *NetworkStatusUpdater) ReplaceEndpointsAndPatch(ctx context.Context,
	ew *v1alpha1.ExternalWorkload, endpoints []v1alpha1.ExternalEndpoint) error {
	sortEndpoints(endpoints)

	// Treat nil and empty slices as equal — reflect.DeepEqual returns false for
	// nil vs []ExternalEndpoint{} even though both mean "no endpoints", which
	// would trigger spurious status patches and a tracker resourceVersion bump
	// that races with concurrent Updates.
	if len(ew.Status.NetworkStatus.ExternalEndpoints) == 0 && len(endpoints) == 0 {
		return nil
	}
	if len(ew.Status.NetworkStatus.ExternalEndpoints) == len(endpoints) &&
		reflect.DeepEqual(ew.Status.NetworkStatus.ExternalEndpoints, endpoints) {
		return nil
	}

	ew.Status.NetworkStatus.ExternalEndpoints = endpoints
	patchBytes, err := updateNetworkStatusPatchBytes(&ew.Status.NetworkStatus)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - "+
			"replace network status of externalworkload <%s/%s>, error: %s",
			ew.Namespace, ew.Name, err.Error())
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	if err := n.cachedClient.Status().Patch(ctx, ew, patch); err != nil {
		log.Error().Msgf("Failed to patch external workload network status <%s/%s>, error: %s",
			ew.Namespace, ew.Name, err.Error())
		return err
	}
	return nil
}
