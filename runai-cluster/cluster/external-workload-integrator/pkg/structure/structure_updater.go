/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package structure

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/metrics"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	topowner "github.com/run-ai/runai/runai-cluster/common/top-owner"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	taskHandlingMaxRetryAttempts   = 10
	rateLimiterBaseDelay           = time.Millisecond
	rateLimiterMaxDelay            = time.Second * 10
	rateLimit                      = 30
	rateBursts                     = 90
	defaultOwnedObjectEnqueueDelay = 5 * time.Second
)

type StructureUpdater struct {
	cachedClient            client.Client
	tasksQueue              workqueue.TypedRateLimitingInterface[utils.ExternalWorkloadTask]
	ownedObjectEnqueueDelay time.Duration
}

// NewStructureUpdater creates a new structure updater with its own rate-limited task queue.
func NewStructureUpdater(cachedClient client.Client) *StructureUpdater {
	return &StructureUpdater{
		cachedClient:            cachedClient,
		ownedObjectEnqueueDelay: defaultOwnedObjectEnqueueDelay,
		tasksQueue: workqueue.NewTypedRateLimitingQueueWithConfig(
			workqueue.NewTypedMaxOfRateLimiter(
				workqueue.NewTypedItemExponentialFailureRateLimiter[utils.ExternalWorkloadTask](rateLimiterBaseDelay, rateLimiterMaxDelay),
				&workqueue.TypedBucketRateLimiter[utils.ExternalWorkloadTask]{Limiter: rate.NewLimiter(rate.Limit(rateLimit), rateBursts)},
			), workqueue.TypedRateLimitingQueueConfig[utils.ExternalWorkloadTask]{
				Name:            fmt.Sprintf("%s-structure", config.Config().ServiceName),
				MetricsProvider: metrics.PrometheusMetricsProviderInstance,
			}),
	}
}

// SetOwnedObjectEnqueueDelay overrides the delay used when enqueueing structure
// recalculation for owned objects (pods). Useful in tests to avoid waiting.
func (n *StructureUpdater) SetOwnedObjectEnqueueDelay(d time.Duration) {
	n.ownedObjectEnqueueDelay = d
}

// EnqueueStructureRecalculationForOwnedObjectIfNeeded enqueues a recalculation -
// only for pods and only if the pod was added or deleted - not if updated.
func (n *StructureUpdater) EnqueueStructureRecalculationForOwnedObjectIfNeeded(
	externalWorkload *v1alpha1.ExternalWorkload,
	topOwnerGvk schema.GroupVersionKind,
	ownedObject client.Object,
	addedOrDeletedResource bool,
) {
	if !addedOrDeletedResource {
		return
	}

	if _, isPod := ownedObject.(*corev1.Pod); isPod {
		n.enqueueStructureRecalculationAfter(externalWorkload, topOwnerGvk, n.ownedObjectEnqueueDelay)
	}
}

// EnqueueStructureRecalculation enqueues a structure recalculation task for the given
// ExternalWorkload. Called from the handler on pod DELETE and EW create events.
func (n *StructureUpdater) EnqueueStructureRecalculation(
	externalWorkload *v1alpha1.ExternalWorkload, gvk schema.GroupVersionKind,
) {
	if externalWorkload == nil {
		return
	}
	key := types.NamespacedName{Name: externalWorkload.Name, Namespace: externalWorkload.Namespace}
	task := utils.NewExternalWorkloadTask(utils.AddOperation, key, gvk)
	n.tasksQueue.Add(*task)
}

// enqueueStructureRecalculationAfter enqueues a structure recalculation task with a delay.
// Used for owned-object (pod) ADD events to allow time for the pod's top-owner labels
// to be patched and synced to the informer cache before the structure task runs.
func (n *StructureUpdater) enqueueStructureRecalculationAfter(
	externalWorkload *v1alpha1.ExternalWorkload, gvk schema.GroupVersionKind, delay time.Duration,
) {
	if externalWorkload == nil {
		return
	}
	key := types.NamespacedName{Name: externalWorkload.Name, Namespace: externalWorkload.Namespace}
	task := utils.NewExternalWorkloadTask(utils.AddOperation, key, gvk)
	n.tasksQueue.AddAfter(*task, delay)
}

// ExecuteTasks loops processing items from the task queue until the queue is shut down.
func (n *StructureUpdater) ExecuteTasks(ctx context.Context) {
	for n.processNextItem(ctx) {
	}
}

func (n *StructureUpdater) processNextItem(ctx context.Context) bool {
	queue := n.tasksQueue
	task, quit := queue.Get()
	if quit {
		log.Info().Msgf("StructureUpdater: task queue shut down, stopping")
		return false
	}
	defer queue.Done(task)

	log.Debug().Msgf("StructureUpdater: processing task for <%s> key <%s>",
		task.GVK.Kind, task.Key.String())

	n.handleTask(ctx, task)

	log.Debug().Msgf("StructureUpdater: completed task for <%s> key <%s>",
		task.GVK.Kind, task.Key.String())
	return true
}

func (n *StructureUpdater) handleTask(ctx context.Context, task utils.ExternalWorkloadTask) {
	queue := n.tasksQueue

	err := n.RecalculateStructure(ctx, task.Key, task.GVK)
	if err == nil || queue.NumRequeues(task) > taskHandlingMaxRetryAttempts {
		if err != nil {
			log.Error().Msgf(
				"StructureUpdater: failed to recalculate structure for EW <%s>, gvk <%s>, after %d retries, error: %s",
				task.Key, task.GVK.String(), queue.NumRequeues(task), err.Error())
		}
		// Success or max retries exceeded — remove the task from the rate limiter's failure tracking
		queue.Forget(task)
	} else {
		log.Warn().Msgf("StructureUpdater: retrying task for EW <%s>, gvk <%s> (attempt %d), error: %s",
			task.Key, task.GVK.String(), queue.NumRequeues(task)+1, err.Error())
		// Transient failure — re-enqueue with exponential backoff for another attempt.
		queue.AddRateLimited(task)
	}
}

// RecalculateStructure fetches an ExternalWorkload and its pods, builds the element
// structure hierarchy, patches the EW status, and labels pods with element IDs.
// Returns nil (no retry) when the EW or Karta is not found, or when the structure is unchanged.
func (n *StructureUpdater) RecalculateStructure(ctx context.Context, ewKey types.NamespacedName, gvk schema.GroupVersionKind) error {
	ew := &v1alpha1.ExternalWorkload{}
	if err := n.cachedClient.Get(ctx, ewKey, ew); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed getting EW <%s>: %w", ewKey, err)
	}

	pods, err := utils.ListWorkloadPods(ctx, n.cachedClient, ew)
	if err != nil {
		return fmt.Errorf("failed listing pods for EW <%s/%s>: %w", ew.Namespace, ew.Name, err)
	}
	if len(pods) == 0 {
		if ew.Status.StructureElements == nil {
			log.Info().Msgf("StructureUpdater: no labeled pods found for new EW <%s/%s>, gvk: <%s>",
				ew.Namespace, ew.Name, gvk.String())
			return n.patchStructureStatus(ctx, ew, nil)
		}
		if len(ew.Status.StructureElements.Elements) > 0 {
			log.Info().Msgf("StructureUpdater: clearing structure for EW <%s/%s> (0 pods found), gvk: <%s>",
				ew.Namespace, ew.Name, gvk.String())
			return n.patchStructureStatus(ctx, ew, nil)
		}
		return nil
	}

	existingElementsByID := existingElementsToMap(ew.Status.StructureElements)

	if IsStructureUnchanged(pods, existingElementsByID) {
		return nil
	}

	ownerChainByPodName, err := n.fetchMissingOwnerChains(ctx, pods, existingElementsByID)
	if err != nil {
		return err
	}

	newPodAddresses, err := n.buildStructureAndPatchStatus(ctx, ew, gvk, pods, ownerChainByPodName, existingElementsByID)
	if err != nil {
		return err
	}

	n.labelPodsWithElements(ctx, pods, newPodAddresses)

	log.Info().Msgf("StructureUpdater: successfully recalculated structure for workload <%s>, gvk: <%s>",
		ewKey.String(), gvk.String())
	return nil
}

// fetchMissingOwnerChains fetches owner chains only for pods that need them —
// unlabeled pods or pods whose element labels reference IDs not in existingElementsByID.
// Pods with valid element labels already have their structure resolved and don't need
// owner chain lookups.
func (n *StructureUpdater) fetchMissingOwnerChains(ctx context.Context, pods []corev1.Pod, existingElementsByID map[string]*v1alpha1.StructureElement) (map[string]*OwnerChain, error) {
	podsNeedingChains := PodsRequiringOwnerChain(pods, existingElementsByID)
	if len(podsNeedingChains) == 0 {
		return nil, nil
	}

	ownerChainByPodName := make(map[string]*OwnerChain, len(podsNeedingChains))
	for _, pod := range podsNeedingChains {
		chain, err := n.buildOwnerChain(ctx, pod)
		if err != nil {
			return nil, fmt.Errorf("failed building owner chain for pod <%s/%s>: %w", pod.Namespace, pod.Name, err)
		}
		ownerChainByPodName[pod.Name] = &chain
	}
	return ownerChainByPodName, nil
}

// buildOwnerChain walks the OwnerReferences chain from a pod up to the top owner,
// collecting each intermediate owner into an OwnerChain.
func (n *StructureUpdater) buildOwnerChain(ctx context.Context, pod *corev1.Pod) (OwnerChain, error) {
	var chain OwnerChain
	ownerRefs := pod.GetOwnerReferences()
	if len(ownerRefs) == 0 {
		return chain, nil
	}

	namespace := pod.Namespace
	nextOwner := &ownerRefs[0]
	for nextOwner != nil {
		ownerInstance, err := topowner.GetOwnerInstance(ctx, nextOwner, namespace, n.cachedClient)
		if err != nil {
			if errors.IsForbidden(err) {
				log.Warn().Msgf("StructureUpdater: cannot read owner <%s/%s> of kind <%s> (forbidden), stopping chain walk",
					namespace, nextOwner.Name, nextOwner.Kind)
				return chain, nil
			}
			return chain, fmt.Errorf("failed getting owner <%s/%s> of kind <%s>: %w",
				namespace, nextOwner.Name, nextOwner.Kind, err)
		}

		ownerGVK := ownerInstance.GroupVersionKind()
		chain = append(chain, OwnerChainEntry{
			UID:     ownerInstance.GetUID(),
			Name:    ownerInstance.GetName(),
			Kind:    ownerGVK.Kind,
			Group:   ownerGVK.Group,
			Version: ownerGVK.Version,
		})

		nextOwners := ownerInstance.GetOwnerReferences()
		if len(nextOwners) == 0 {
			break
		}
		if len(nextOwners) > 1 {
			return chain, fmt.Errorf("%s <%s/%s> has multiple owners", ownerGVK.Kind, namespace, ownerInstance.GetName())
		}
		nextOwner = &nextOwners[0]
	}
	return chain, nil
}

// buildStructureAndPatchStatus fetches the Karta for the workload's GVK, computes the element
// structure, and patches the EW status. Returns the addresses of newly-processed pods so
// the caller can label them.
func (n *StructureUpdater) buildStructureAndPatchStatus(
	ctx context.Context,
	ew *v1alpha1.ExternalWorkload,
	gvk schema.GroupVersionKind,
	pods []corev1.Pod,
	ownerChainByPodName map[string]*OwnerChain,
	existingElementsByID map[string]*v1alpha1.StructureElement,
) ([]*PodAddress, error) {
	karta, err := utils.GetKartaOfGvk(ctx, n.cachedClient, gvk)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debug().Msgf("StructureUpdater: no Karta found for EW <%s/%s> (gvk=%s), skipping",
				ew.Namespace, ew.Name, gvk.String())
			return nil, nil
		}
		return nil, fmt.Errorf("failed getting Karta for EW <%s/%s>: %w", ew.Namespace, ew.Name, err)
	}

	newElementsByID, newPodAddresses, err := BuildElementStructure(
		ctx, string(ew.UID), karta, pods, ownerChainByPodName, existingElementsByID)
	if err != nil {
		return nil, fmt.Errorf("failed building element structure for EW <%s/%s>: %w",
			ew.Namespace, ew.Name, err)
	}

	if err = n.patchStructureStatus(ctx, ew, newElementsByID); err != nil {
		return nil, err
	}

	return newPodAddresses, nil
}

// patchStructureStatus converts the element map to a sorted slice and patches the
// EW status subresource with the new structure.
func (n *StructureUpdater) patchStructureStatus(ctx context.Context, ew *v1alpha1.ExternalWorkload, elementsByID map[string]*v1alpha1.StructureElement) error {
	se := &v1alpha1.StructureElements{
		Elements:       elementsMapToSortedSlice(elementsByID),
		LastCalculated: metav1.Now(),
	}

	patchBytes, err := structureElementsPatchBytes(se)
	if err != nil {
		return fmt.Errorf("failed marshalling structure patch for EW <%s/%s>: %w", ew.Namespace, ew.Name, err)
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	if err = n.cachedClient.Status().Patch(ctx, ew, patch); err != nil {
		return fmt.Errorf("failed patching structure status for EW <%s/%s>: %w", ew.Namespace, ew.Name, err)
	}

	return nil
}

// labelPodsWithElements patches element-level labels onto newly-processed pods.
// This is best-effort: failures are logged as warnings but don't cause a retry,
// because unlabeled pods will be re-processed on the next recalculation cycle.
func (n *StructureUpdater) labelPodsWithElements(ctx context.Context, pods []corev1.Pod, newPodAddresses []*PodAddress) {
	if len(newPodAddresses) == 0 {
		return
	}

	podByUID := make(map[string]*corev1.Pod, len(pods))
	for i := range pods {
		podByUID[string(pods[i].UID)] = &pods[i]
	}

	for _, addr := range newPodAddresses {
		pod, ok := podByUID[addr.PodID]
		if !ok {
			continue
		}

		labels := GeneratePodElementLabels(addr.ElementsPath)
		patchBytes, err := utils.AddLabelsPatchBytes(labels)
		if err != nil {
			log.Warn().Msgf("StructureUpdater: failed to marshal label patch for pod <%s/%s>, error: %s",
				pod.Namespace, pod.Name, err.Error())
			continue
		}

		patch := client.RawPatch(types.MergePatchType, patchBytes)
		if err = n.cachedClient.Patch(ctx, pod, patch); err != nil {
			log.Warn().Msgf("StructureUpdater: failed to patch element labels on pod <%s/%s>, error: %s",
				pod.Namespace, pod.Name, err.Error())
			continue
		}

		log.Info().Msgf("StructureUpdater: patched pod <%s/%s> with element labels: %v",
			pod.Namespace, pod.Name, labels)
	}
}

func structureElementsPatchBytes(se *v1alpha1.StructureElements) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"structureElements": se,
		},
	})
}

// existingElementsToMap converts the flat slice stored in EW status to a map keyed by element ID.
func existingElementsToMap(se *v1alpha1.StructureElements) map[string]*v1alpha1.StructureElement {
	if se == nil {
		return nil
	}
	m := make(map[string]*v1alpha1.StructureElement, len(se.Elements))
	for i := range se.Elements {
		m[se.Elements[i].ID] = &se.Elements[i]
	}
	return m
}

// elementsMapToSortedSlice converts element map back to flat sorted slice for EW status storage.
func elementsMapToSortedSlice(m map[string]*v1alpha1.StructureElement) []v1alpha1.StructureElement {
	s := make([]v1alpha1.StructureElement, 0, len(m))
	for _, e := range m {
		s = append(s, *e)
	}

	sortElements(s)
	return s
}

// sortElements sorts elements deterministically: by Level ascending, then by ID ascending.
func sortElements(elements []v1alpha1.StructureElement) {
	sort.Slice(elements, func(i, j int) bool {
		if elements[i].Level != elements[j].Level {
			return elements[i].Level < elements[j].Level
		}
		return elements[i].ID < elements[j].ID
	})
}
