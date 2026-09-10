/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"golang.org/x/exp/maps"

	constants2 "github.com/kai-scheduler/KAI-scheduler/pkg/podgrouper/podgrouper/plugins/constants"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/networkstatus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/structure"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/supportedhandlers"
	workloadTypeConfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/workload-type-config"
	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedulingv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExternalWorkloadHandler struct {
	cachedClient           client.Client
	allowNonSupportedTypes bool

	supportedGVKsHandlers map[schema.GroupVersionKind]supportedhandlers.TopOwnerHandler

	// no need to handle if an object of this type is the top owner -
	// but need to handle if it's an owned object.
	onlyOwnedSupportedGVKs map[schema.GroupVersionKind]bool

	tasksQueue      workqueue.TypedRateLimitingInterface[utils.ExternalWorkloadTask]
	metricsExporter *metrics.MetricsExporter

	networkStatusUpdater *networkstatus.NetworkStatusUpdater
	structureUpdater     *structure.StructureUpdater
}

const (
	rateLimiterBaseDelay = time.Millisecond
	rateLimiterMaxDelay  = time.Second * 10
	rateLimit            = 30
	rateBursts           = 90
)

var (
	LabelsToCopy = []string{
		constants.ManagedByLabelKey,
		constants.DynamoGraphDeploymentNameLabelKey,
		commonconstants.ConvertedFromNativeLabel,
		clustercommonconstants.LabelWorkloadName,
	}
	AnnotationsToCopy = []string{constants.HelmReleaseAnnotationKey, commonconstants.WorkloadCliCommandAnnotation}
)

func externalWorkloadTypeFromObjectLabels(labels map[string]string, fallback string) string {
	if labels[commonconstants.ConvertedFromNativeLabel] == "true" {
		if nativeType := labels[clustercommonconstants.LabelWorkloadKind]; nativeType != "" {
			return nativeType
		}
	}

	return fallback
}

type ObjectIdentifier struct {
	Name         string
	Namespace    string
	GroupVersion schema.GroupVersion
}

func NewExternalWorkloadHandler(cachedClient client.Client, draParser *dra.DRAParser, allowNonSupportedTypes bool,
	supportedGVKsHandlers map[schema.GroupVersionKind]supportedhandlers.TopOwnerHandler,
	onlyOwnedSupportedGVKs map[schema.GroupVersionKind]bool,
	additionalGVKs []schema.GroupVersionKind,
	networkStatusUpdater *networkstatus.NetworkStatusUpdater,
	structureUpdater *structure.StructureUpdater) *ExternalWorkloadHandler {
	return &ExternalWorkloadHandler{
		cachedClient:           cachedClient,
		allowNonSupportedTypes: allowNonSupportedTypes,
		supportedGVKsHandlers:  buildSupportedGVKsHandlers(cachedClient, draParser, supportedGVKsHandlers, additionalGVKs),
		onlyOwnedSupportedGVKs: onlyOwnedSupportedGVKs,
		tasksQueue: workqueue.NewTypedRateLimitingQueueWithConfig(
			workqueue.NewTypedMaxOfRateLimiter(
				workqueue.NewTypedItemExponentialFailureRateLimiter[utils.ExternalWorkloadTask](rateLimiterBaseDelay, rateLimiterMaxDelay),
				&workqueue.TypedBucketRateLimiter[utils.ExternalWorkloadTask]{Limiter: rate.NewLimiter(rate.Limit(rateLimit), rateBursts)},
			), workqueue.TypedRateLimitingQueueConfig[utils.ExternalWorkloadTask]{Name: config.Config().ServiceName, MetricsProvider: metrics.PrometheusMetricsProviderInstance}),
		metricsExporter:      metrics.GetMetricsExporter(),
		networkStatusUpdater: networkStatusUpdater,
		structureUpdater:     structureUpdater,
	}
}

func (h *ExternalWorkloadHandler) StructureUpdater() *structure.StructureUpdater {
	return h.structureUpdater
}

func buildSupportedGVKsHandlers(
	cachedClient client.Client, draParser *dra.DRAParser,
	supportedGVKsHandlers map[schema.GroupVersionKind]supportedhandlers.TopOwnerHandler,
	additionalGVKs []schema.GroupVersionKind,
) map[schema.GroupVersionKind]supportedhandlers.TopOwnerHandler {
	result := map[schema.GroupVersionKind]supportedhandlers.TopOwnerHandler{}

	for gvk, handler := range supportedGVKsHandlers {
		result[gvk] = handler
	}

	for _, gvk := range additionalGVKs {
		handler := handlers.NewGvkWorkloadHandler(cachedClient, draParser, gvk)
		result[gvk] = handler
	}

	return result
}

func (h *ExternalWorkloadHandler) SetAllowNonSupportedTypes(allowNonSupportedTypes bool) {
	h.allowNonSupportedTypes = allowNonSupportedTypes
}

func (h *ExternalWorkloadHandler) InitMetricsExporter() {
	h.metricsExporter.Init()
}

// EventHandlerFilter filters out objects that are not in a runai-managed namespace + optional filter per-type.
// As you add more objects, make sure you add tests for them in "Test workload objects that don't pass FilterFunc".
func (h *ExternalWorkloadHandler) EventHandlerFilter(ctx context.Context, object client.Object,
	gvk schema.GroupVersionKind,
) bool {
	filterObject := object.DeepCopyObject().(client.Object)
	if !h.checkSchedulerFilter(ctx, filterObject, gvk) {
		return false
	}

	if !h.checkSecondaryFilter(ctx, filterObject, gvk) {
		return false
	}

	return true
}

// checkSchedulerFilter applies namespace- and scheduler-related filters.
func (h *ExternalWorkloadHandler) checkSchedulerFilter(ctx context.Context, object client.Object,
	gvk schema.GroupVersionKind,
) bool {
	object.GetObjectKind().SetGroupVersionKind(gvk)

	namespace, err := utils.GetNamespace(ctx, object.GetNamespace(), h.cachedClient)
	if err != nil {
		return false
	}

	if !utils.IsRunaiManagedNamespace(namespace) {
		return false
	}

	// see if the namespace has the enforce label; except for pods -
	// since the pod itself should already be created with schedulerName in spec.
	if gvk.Kind != "Pod" && utils.IsSchedulerEnforcedForNamespace(namespace) {
		return true
	}

	gvkHandler, found := h.supportedGVKsHandlers[gvk]
	if !found {
		_, found = h.onlyOwnedSupportedGVKs[gvk]

		return found
	}

	eventFilterHandler, ok := gvkHandler.(EventFilterHandler)
	if ok {
		return eventFilterHandler.Filter(object)
	}

	return true
}

// checkSecondaryFilter applies a secondary filter per-type, if such.
func (h *ExternalWorkloadHandler) checkSecondaryFilter(ctx context.Context, object client.Object,
	gvk schema.GroupVersionKind,
) bool {
	object.GetObjectKind().SetGroupVersionKind(gvk)

	gvkHandler, found := h.supportedGVKsHandlers[gvk]
	if !found {
		return true
	}

	secondaryEventFilterHandler, ok := gvkHandler.(SecondaryEventFilterHandler)
	if !ok {
		return true
	}

	return secondaryEventFilterHandler.SecondaryFilter(object)
}

// HandleWorkloadAddUpdateEvent handles the given workload object of type "gvk" when an add/update events occurs.
// If it's a top-owned object (has no owners) - calls the relevant HandleFn.
// If it has an owner - finds the top owner and handles accordingly.
// operation indicates whether this is an add or update event.
func (h *ExternalWorkloadHandler) HandleWorkloadAddUpdateEvent(ctx context.Context, object client.Object,
	gvk schema.GroupVersionKind, operation utils.TaskOperation,
) error {
	object.GetObjectKind().SetGroupVersionKind(gvk)
	log.Debug().Msgf("Handling workload %s event for %s", operation, objectLogStr(object))

	addedResource := operation == utils.AddOperation

	ownerRefs := object.GetOwnerReferences()
	if len(ownerRefs) == 0 {
		// this is the top-owned object
		return h.handleTopOwnedObject(ctx, object, addedResource)
	}

	return h.handleOwnedObject(ctx, object, addedResource)
}

func (h *ExternalWorkloadHandler) createChildId(object client.Object) *v1alpha1.ChildId {
	return &v1alpha1.ChildId{
		UID:  string(object.GetUID()),
		Type: object.GetObjectKind().GroupVersionKind().Kind,
	}
}

func (h *ExternalWorkloadHandler) getExternalWorkloadByName(ctx context.Context, namespace, name string) (*v1alpha1.ExternalWorkload, error) {
	ew := &v1alpha1.ExternalWorkload{}
	err := h.cachedClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, ew)
	if err != nil {
		return nil, err
	}
	return ew, nil
}

func getExternalWorkloadName(obj client.Object) string {
	return obj.GetAnnotations()[commonconstants.ExternalWorkloadName]
}

func (h *ExternalWorkloadHandler) HandleDeleteEvent(ctx context.Context, deleteObj client.Object, _ schema.GroupVersionKind, _ utils.TaskOperation) error {
	log.Debug().Msgf("Handling workload delete event for %s", objectLogStr(deleteObj))
	childId := h.createChildId(deleteObj)
	childIdsToDelete := []v1alpha1.ChildId{*childId}

	externWorkloadOfObj, err := h.getExternalWorkload(ctx, deleteObj)

	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil || externWorkloadOfObj == nil {
		return err
	}
	err = h.deleteChildIdsToExternalWorkloadStatus(ctx, externWorkloadOfObj, childIdsToDelete)
	if err != nil {
		return err
	}

	h.structureUpdater.EnqueueStructureRecalculationForOwnedObjectIfNeeded(externWorkloadOfObj,
		schema.GroupVersionKind{Group: externWorkloadOfObj.Spec.Group, Kind: externWorkloadOfObj.Spec.Type},
		deleteObj, true)

	return nil
}

func (h *ExternalWorkloadHandler) getExternalWorkload(ctx context.Context, childObj client.Object) (*v1alpha1.ExternalWorkload, error) {
	var externWorkloadOfObj *v1alpha1.ExternalWorkload
	var err error

	name := getExternalWorkloadName(childObj)
	if name != "" {
		externWorkloadOfObj, err = h.getExternalWorkloadByName(ctx, childObj.GetNamespace(), name)
	} else {
		ownerRefs := childObj.GetOwnerReferences()
		if len(ownerRefs) == 0 {
			externWorkloadOfObj, err = utils.GetExternalWorkloadOfObject(ctx, string(childObj.GetUID()), childObj.GetNamespace(), h.cachedClient)
		} else {
			_, topOwnerWithHandlerMeta, getOwnerErr := h.getTopResourceOwnerWithHandler(ctx, childObj)
			if getOwnerErr != nil {
				if !errors.IsNotFound(getOwnerErr) {
					log.Error().Msgf("Could not find External Workload %s %s %s\n",
						childObj.GetName(), childObj.GetNamespace(), getOwnerErr)
				}
				return nil, getOwnerErr
			}
			var topOwnerMeta client.Object
			if topOwnerWithHandlerMeta != nil {
				topOwnerMeta = topOwnerWithHandlerMeta
			} else {
				topOwnerMeta = childObj
			}
			var uid types.UID
			uid, _, err = h.getOriginalWorkloadUidName(topOwnerMeta)
			if err != nil {
				// this is not considered an error, no need to retry
				return nil, nil
			}
			externWorkloadOfObj, err = utils.GetExternalWorkloadOfObject(ctx, string(uid), topOwnerMeta.GetNamespace(), h.cachedClient)
		}
	}
	if err != nil {
		return nil, err
	}
	return externWorkloadOfObj, nil
}

// ////////////// top-owned objects section
func (h *ExternalWorkloadHandler) handleTopOwnedObject(ctx context.Context, object client.Object, addedResource bool) error {
	gvk := object.GetObjectKind().GroupVersionKind()
	_, found := h.onlyOwnedSupportedGVKs[gvk]
	if found {
		log.Debug().Msgf("Dropping top-owned object %s: it should only be handled as owned",
			objectLogStr(object))
		return nil
	}

	if !h.checkSecondaryFilter(ctx, object, gvk) {
		log.Debug().Msgf("Dropping top-owned object %s: secondary filter didn't pass",
			objectLogStr(object))
		return nil
	}

	log.Debug().Msgf("Handling top-owned object %s", objectLogStr(object))

	uid, _, err := h.getOriginalWorkloadUidName(object)
	if err != nil {
		// this is not considered an error, no need to retry
		return nil
	}
	externWorkloadOfObj, err := utils.GetExternalWorkloadOfObject(ctx, string(uid), object.GetNamespace(), h.cachedClient)
	if err != nil || externWorkloadOfObj == nil {
		if errors.IsNotFound(err) {
			return h.createExternalWorkloadForObject(ctx, object)
		}
		log.Error().Msgf("Failed to get external workload of top-owned object %s, error: %s",
			objectLogStr(object), err.Error())
		return err
	}
	_ = h.objectAdditionalHandler(ctx, object, nil, externWorkloadOfObj)

	err = h.handleExternalWorkloadOfTopOwnedObject(ctx, object, externWorkloadOfObj, addedResource)
	if err == nil && !addedResource {
		// ADD events are already enqueued by updateExternalWorkloadForObject.
		// Enqueue after update so that Spec.Urls are already persisted
		h.enqueueNetworkStatusIfNeeded(object.GetObjectKind().GroupVersionKind(), externWorkloadOfObj)
	}
	return err
}

func (h *ExternalWorkloadHandler) handleExternalWorkloadOfTopOwnedObject(ctx context.Context,
	object client.Object, externWorkloadOfObj *v1alpha1.ExternalWorkload, addedResource bool,
) error {
	return h.updateExternalWorkloadForObject(ctx, object, externWorkloadOfObj, addedResource)
}

func (h *ExternalWorkloadHandler) createExternalWorkloadForObject(ctx context.Context, object client.Object) error {
	gvk := object.GetObjectKind().GroupVersionKind()
	originalWorkloadUid, originalWorkloadName, err := h.getOriginalWorkloadUidName(object)
	if err != nil {
		// this is not considered an error, no need to retry
		return nil
	}
	ownerReferences := h.getOwnerReferences(object)

	externalWorkload := &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:            utils.GenerateExternalWorkloadName(gvk, originalWorkloadName),
			Namespace:       object.GetNamespace(),
			Annotations:     map[string]string{},
			OwnerReferences: ownerReferences,
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			OriginalWorkloadUID:  originalWorkloadUid,
			OriginalWorkloadName: originalWorkloadName,
			Type:                 externalWorkloadTypeFromObjectLabels(object.GetLabels(), gvk.Kind),
			Group:                gvk.Group,
			Version:              gvk.Version,
			CreationTimestamp:    object.GetCreationTimestamp(),
		},
	}

	h.setWorkloadCategory(ctx, object, externalWorkload)

	// Copy category and priority labels from the K8s object to the ExternalWorkload (in case they were changed by the API)
	if categoryLabel := object.GetLabels()[commonconstants.WorkloadCategoryLabel]; categoryLabel != "" {
		addToLabels(externalWorkload, commonconstants.WorkloadCategoryLabel, categoryLabel)
	}
	// Copy priority label (priorityClassName) set by the workloads-manager API
	if priorityLabel := object.GetLabels()[constants2.PriorityLabelKey]; priorityLabel != "" {
		addToLabels(externalWorkload, constants2.PriorityLabelKey, priorityLabel)
	}

	workloadId := topowner.GetRunaiWorkloadId(object.GetUID(), object.GetAnnotations())
	if workloadId != string(object.GetUID()) {
		externalWorkload.Annotations[commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload] = workloadId
	}

	setWorkloadRunnerLabel(object, externalWorkload)

	err = h.fillHierarchyInfoForTopOwner(ctx, object, &externalWorkload.Spec)
	if err != nil {
		return err
	}

	err = h.parseObjectToExternalWorkload(ctx, object, externalWorkload)
	if err != nil {
		return err
	}

	err = h.createExternalWorkloadAndUpdateStatus(ctx, externalWorkload)
	if err != nil {
		return err
	}

	// Enqueue network status update after EW creation so endpoints (NodePort, Ingress, etc.)
	// are discovered even if the Service ADD event was processed before this EW existed.
	key := types.NamespacedName{Name: externalWorkload.Name, Namespace: externalWorkload.Namespace}
	// DeepCopy so the consumer can't mutate (or observe partial mutations of)
	// the caller's ExternalWorkload pointer while the caller is still updating it.
	task := utils.NewExternalWorkloadTask(utils.AddOperation, key, externalWorkload.GroupVersionKind()).SetObject(externalWorkload.DeepCopy())
	h.networkStatusUpdater.TasksQueue().Add(*task)

	h.structureUpdater.EnqueueStructureRecalculation(externalWorkload, gvk)
	return nil
}

func (h *ExternalWorkloadHandler) getOriginalWorkloadUidName(object client.Object) (types.UID, string, error) {
	gvk := object.GetObjectKind().GroupVersionKind()
	gvkHandler, found := h.supportedGVKsHandlers[gvk]
	if found {
		nameUidHandler, ok := gvkHandler.(GetOriginalWorkloadNameUidHandler)
		if ok {
			uid, name, err := nameUidHandler.GetOriginalWorkloadNameUid(object)
			if err != nil {
				log.Error().Msgf("Failed to get original workload uid/name for top-owned object %s, error: %s",
					objectLogStr(object), err.Error())
			}
			return uid, name, err
		}
	}

	return object.GetUID(), object.GetName(), nil
}

func (h *ExternalWorkloadHandler) getOwnerReferences(object client.Object) []metav1.OwnerReference {
	gvk := object.GetObjectKind().GroupVersionKind()
	gvkHandler, found := h.supportedGVKsHandlers[gvk]
	if found {
		nameUidHandler, ok := gvkHandler.(OwnerReferencesHandler)
		if ok {
			ownerReferences := nameUidHandler.GetOwnerReferences(object)
			if ownerReferences != nil {
				return ownerReferences
			}
			// else, we fall back to the default owner reference, to be on the safe side and never have a nil owner reference.
		}
	}

	// default owner reference - the object itself
	return []metav1.OwnerReference{
		{
			APIVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
			Name:       object.GetName(),
			UID:        object.GetUID(),
		},
	}
}

func (h *ExternalWorkloadHandler) updateExternalWorkloadForObject(ctx context.Context, object client.Object,
	externalWorkloadOfObj *v1alpha1.ExternalWorkload, addedResource bool,
) error {
	gvk := object.GetObjectKind().GroupVersionKind()
	if !h.checkSecondaryFilter(ctx, object, gvk) {
		return nil
	}

	originalExternalWorkload := externalWorkloadOfObj.DeepCopy()

	externalWorkloadOfObj.OwnerReferences = h.getOwnerReferences(object)

	// Copy category and priority labels from the K8s object to the ExternalWorkload (in case they were changed by the API)
	if categoryLabel := object.GetLabels()[commonconstants.WorkloadCategoryLabel]; categoryLabel != "" {
		addToLabels(externalWorkloadOfObj, commonconstants.WorkloadCategoryLabel, categoryLabel)
	}
	// Copy priority label (priorityClassName) set by the workloads-manager API
	if priorityLabel := object.GetLabels()[constants2.PriorityLabelKey]; priorityLabel != "" {
		addToLabels(externalWorkloadOfObj, constants2.PriorityLabelKey, priorityLabel)
	}

	setWorkloadRunnerLabel(object, externalWorkloadOfObj)

	err := h.parseObjectToExternalWorkload(ctx, object, externalWorkloadOfObj)
	if err != nil {
		return err
	}

	// client.Update might override the status
	observedStatusCopy := externalWorkloadOfObj.Status.ObservedStatus.DeepCopy()

	innerErr := h.updateSpecIfNeeded(ctx, externalWorkloadOfObj, originalExternalWorkload)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	externalWorkloadOfObj.Status.ObservedStatus = *observedStatusCopy
	innerErr = h.updateObservedStatusIfNeeded(ctx, externalWorkloadOfObj,
		&originalExternalWorkload.Status.ObservedStatus)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	innerErr = h.networkStatusUpdater.UpdateNetworkStatusIfNeeded(ctx, object, externalWorkloadOfObj, addedResource)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	return err
}

func (h *ExternalWorkloadHandler) updateSpecIfNeeded(ctx context.Context,
	externalWorkloadToUpdate, originalExternalWorkload *v1alpha1.ExternalWorkload,
) error {
	specChanged := !equalExternalWorkloadSpec(&originalExternalWorkload.Spec, &externalWorkloadToUpdate.Spec)
	labelsChanged := !maps.Equal(originalExternalWorkload.Labels, externalWorkloadToUpdate.Labels)
	// Annotations carry the ewi-status handshake; without this they only persist when spec/labels also change.
	annotationsChanged := !maps.Equal(originalExternalWorkload.Annotations, externalWorkloadToUpdate.Annotations)

	if specChanged || labelsChanged || annotationsChanged {
		return h.updateExternalWorkload(ctx, externalWorkloadToUpdate)
	}
	return nil
}

func (h *ExternalWorkloadHandler) updateObservedStatusIfNeeded(ctx context.Context,
	externalWorkloadToUpdate *v1alpha1.ExternalWorkload,
	observedStatusCopy *v1alpha1.ExternalWorkloadObservedStatus,
) error {
	var err error

	if !equalExternalWorkloadObservedStatus(observedStatusCopy,
		&externalWorkloadToUpdate.Status.ObservedStatus) {
		innerErr := h.patchExternalWorkloadObservedStatusWithoutChildIds(ctx, externalWorkloadToUpdate)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}
	}

	return err
}

func (h *ExternalWorkloadHandler) GetFreshObject(ctx context.Context, key client.ObjectKey, gvk schema.GroupVersionKind) (client.Object, error) {
	gvkHandler, found := h.supportedGVKsHandlers[gvk]
	if !found {
		return nil, errors.NewNotFound(v1alpha1.Resource("GVKsHandler"), "")
	}
	return gvkHandler.GetObject(ctx, key.Name, key.Namespace)
}

func (h *ExternalWorkloadHandler) parseObjectToExternalWorkload(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload,
) error {
	gvk := object.GetObjectKind().GroupVersionKind()

	// Set Spec.Group/Version on reconcile (not just create) so existing EWs
	// get backfilled. The per-kind Parse handlers below may override
	// Spec.Type for logical naming (e.g. NIMService → "NIM"); none touch
	// Group or Version.
	externalWorkload.Spec.Group = gvk.Group
	externalWorkload.Spec.Version = gvk.Version

	gvkHandler, found := h.supportedGVKsHandlers[gvk]
	if !found {
		// shouldn't happen
		return fmt.Errorf("unsupported object <%s> (shouldn't happen)", gvk.String())
	}

	// empty the observed status before parsing, so it will be re-calculated
	externalWorkload.Status.ObservedStatus = v1alpha1.ExternalWorkloadObservedStatus{}

	err := gvkHandler.Parse(ctx, object, externalWorkload)
	if err != nil {
		log.Error().Msgf("Failed filling external workload spec for top-owned object %s, error: %s",
			objectLogStr(object), err.Error())
		return err
	}

	externalWorkload.Spec.Type = externalWorkloadTypeFromObjectLabels(object.GetLabels(), externalWorkload.Spec.Type)
	fillLabelsAndAnnotations(object.GetLabels(), object.GetAnnotations(), externalWorkload)
	fillTopology(object.GetAnnotations(), externalWorkload)
	h.applyRunaiStatus(ctx, object, externalWorkload)

	return nil
}

func fillLabelsAndAnnotations(labels map[string]string, annotations map[string]string, externalWorkload *v1alpha1.ExternalWorkload) {
	log.Debug().Msgf("Copy labels for workload <%s> in namespace <%s>", externalWorkload.Name, externalWorkload.Namespace)
	if externalWorkload.Labels == nil {
		externalWorkload.Labels = map[string]string{}
	}
	if externalWorkload.Annotations == nil {
		externalWorkload.Annotations = map[string]string{}
	}
	fillValuesToAdd(labels, LabelsToCopy, externalWorkload.Labels, "label")
	fillValuesToAdd(annotations, AnnotationsToCopy, externalWorkload.Annotations, "annotation")
}

func fillValuesToAdd(sourceMap map[string]string, keysToAdd []string, destMap map[string]string, mapName string) {
	for _, key := range keysToAdd {
		value, exists := sourceMap[key]
		if exists && value != "" {
			log.Debug().Msgf("Copy <%s> <%s>:<%s>", mapName, key, value)
			destMap[key] = value
		}
	}
}

func (h *ExternalWorkloadHandler) fillHierarchyInfoForTopOwner(ctx context.Context, object client.Object,
	externalWorkloadSpec *v1alpha1.ExternalWorkloadSpec,
) error {
	project, err := utils.GetProjectFromNamespace(ctx, object.GetNamespace(), h.cachedClient)
	if err != nil {
		return err
	}

	projectId, found := project.Labels[v2.ProjectIdLabel]
	if !found {
		return fmt.Errorf("failed to get project id of project <%s>", project.Name)
	}

	departmentId, found := project.Labels[schedulingv2alpha2.DepartmentIdLabel]
	if !found {
		return fmt.Errorf("failed to get department id of project <%s>", project.Name)
	}

	externalWorkloadSpec.ProjectName = project.Name
	externalWorkloadSpec.ProjectId = projectId
	externalWorkloadSpec.DepartmentName = project.Spec.Parent
	externalWorkloadSpec.DepartmentId = departmentId

	return nil
}

func (h *ExternalWorkloadHandler) getWorkloadCategory(ctx context.Context, object client.Object) (string, error) {
	workloadKind := object.GetObjectKind().GroupVersionKind().Kind
	workloadGroup := object.GetObjectKind().GroupVersionKind().Group
	workloadCategoryFromLabel := object.GetLabels()[commonconstants.WorkloadCategoryLabel]

	if workloadCategoryFromLabel != "" {
		return workloadCategoryFromLabel, nil
	}

	cfg, err := workloadTypeConfig.GetWorkloadConfigByGroupKind(ctx, h.cachedClient, workloadGroup, workloadKind)
	if err != nil {
		return "", err
	}

	return cfg.CategoryName, nil
}

func (h *ExternalWorkloadHandler) setWorkloadCategory(ctx context.Context, object client.Object, externalWorkload *v1alpha1.ExternalWorkload) {
	kind := object.GetObjectKind().GroupVersionKind().Kind
	category, err := h.getWorkloadCategory(ctx, object)
	if err != nil {
		log.Error().Msgf("failed workload category for type %s, error: %s",
			kind, err.Error())
	}

	if category != "" {
		addToLabels(externalWorkload, commonconstants.WorkloadCategoryLabel, category)
	}
}

func setWorkloadRunnerLabel(object client.Object, externalWorkload *v1alpha1.ExternalWorkload) {
	for _, ownerRef := range object.GetOwnerReferences() {
		if ownerRef.Kind == dromosv1alpha1.WorkloadRunnerKind &&
			ownerRef.APIVersion == v1alpha1.SchemeGroupVersion.String() {
			addToLabels(externalWorkload, commonconstants.WorkloadRunnerNameLabel, ownerRef.Name)
			return
		}
	}
}

func addToLabels(object client.Object, label, value string) {
	if object.GetLabels() == nil {
		object.SetLabels(make(map[string]string))
	}
	object.GetLabels()[label] = value
}

func IsValidTypeForChildId(object client.Object) bool {
	_, ok := object.(*corev1.Pod)

	return !ok
}

func (h *ExternalWorkloadHandler) appendChildIdIfNeeded(object client.Object, list *[]v1alpha1.ChildId) {
	if object == nil || list == nil || !IsValidTypeForChildId(object) {
		return
	}
	childId := h.createChildId(object)
	*list = append(*list, *childId)
}

// ////////////// owned objects section
func (h *ExternalWorkloadHandler) handleOwnedObject(ctx context.Context, object client.Object, addedResource bool) error {
	topOwnerMeta, topOwnerWithHandlerMeta, err := h.getTopResourceOwnerWithHandler(ctx, object)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debug().Msgf("External Workload not found %s %s - Assumed deleted. %s.\n",
				object.GetName(), object.GetNamespace(), err)
			return nil
		}
		if !h.allowNonSupportedTypes || !errors.IsForbidden(err) {
			log.Error().Msgf("Failed getting top owner of object %s, error: %s",
				objectLogStr(object), err.Error())
			return err
		}

		// in this case, topOwnerMeta would be the lastKnownTopOwner
		if topOwnerMeta == nil {
			// consider the current object as the top-owner
			log.Warn().Msgf("Owner of %s is unrecognized; proceeding with self as top owned object (error was: %s)",
				objectLogStr(object), err.Error())
			return h.handleTopOwnedObject(ctx, object, addedResource)
		}
		// continue with the found topOwnerMeta
		log.Warn().Msgf("Top owner of %s is unrecognized; proceeding with the last known top owner: %s (error was: %s)",
			objectLogStr(object), objectLogStr(topOwnerMeta), err.Error())
	}

	if cachedclient.ShouldDropWorkload(topOwnerMeta) {
		log.Debug().Msgf("Dropping owned object <%s> (kind <%s> shouldn't trigger externalworkload creation)",
			objectLogStr(object), topOwnerMeta.GetObjectKind().GroupVersionKind().String())
		return h.objectAdditionalHandler(ctx, object, topOwnerMeta, nil)
	}

	if topOwnerWithHandlerMeta == nil {
		// in this case, if the topOwnerWithHandlerMeta is nil, consider the current object as the top-owner
		log.Warn().Msgf("Top owner of %s doesn't have a handler; proceeding with self as top owned object",
			objectLogStr(object))
		return h.handleTopOwnedObject(ctx, object, addedResource)
	}

	externalWorkloadOfOwner, err := h.handleTopOwnerOfOwnedObject(ctx, object, topOwnerWithHandlerMeta)

	// Owned-object updates do not pass through the top-owner post-update enqueue path.
	// ADD events are already enqueued by UpdateNetworkStatusIfNeeded below.
	if !addedResource {
		h.enqueueNetworkStatusIfNeeded(
			topOwnerWithHandlerMeta.GetObjectKind().GroupVersionKind(),
			externalWorkloadOfOwner,
		)
	}

	additionalHandlerErr := h.objectAdditionalHandler(ctx, object, topOwnerWithHandlerMeta, externalWorkloadOfOwner)

	updateNetworkErr := h.networkStatusUpdater.UpdateNetworkStatusIfNeeded(ctx, object, externalWorkloadOfOwner, addedResource)

	h.structureUpdater.EnqueueStructureRecalculationForOwnedObjectIfNeeded(externalWorkloadOfOwner,
		topOwnerWithHandlerMeta.GetObjectKind().GroupVersionKind(), object, addedResource)

	if err == nil || errors.IsNotFound(err) {
		h.triggerTopOwnerHandlerIfNeeded(ctx, object, topOwnerWithHandlerMeta, externalWorkloadOfOwner)
	}

	if err != nil || additionalHandlerErr != nil || updateNetworkErr != nil {
		return multierror.Append(err, additionalHandlerErr, updateNetworkErr)
	}

	return additionalHandlerErr
}

func (h *ExternalWorkloadHandler) handleTopOwnerOfOwnedObject(ctx context.Context, object client.Object,
	topOwner *metav1.PartialObjectMetadata,
) (*v1alpha1.ExternalWorkload, error) {
	originalWorkloadUid, _, err := h.getOriginalWorkloadUidName(topOwner)
	if err != nil {
		// this is not considered an error, no need to retry
		return nil, nil
	}
	externalWorkloadOfOwner, err := utils.GetExternalWorkloadOfObject(ctx, string(originalWorkloadUid), topOwner.GetNamespace(), h.cachedClient)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, err
		}
		log.Error().Msgf("Failed to get external workload of top owner %s, error: %s",
			objectLogStr(object), err.Error())
		return nil, err
	}

	log.Debug().Msgf("Detected %s as child of externalworkload <%s/%s>",
		objectLogStr(object), externalWorkloadOfOwner.Namespace, externalWorkloadOfOwner.Name)
	err = h.handleExternalWorkloadOfObjectOwner(ctx, externalWorkloadOfOwner, object)
	if err != nil {
		return nil, err
	}
	return externalWorkloadOfOwner, nil
}

func (h *ExternalWorkloadHandler) handleExternalWorkloadOfObjectOwner(ctx context.Context,
	externalWorkloadOfOwner *v1alpha1.ExternalWorkload, ownedObject client.Object,
) error {
	var childIdsToAdd []v1alpha1.ChildId
	h.appendChildIdIfNeeded(ownedObject, &childIdsToAdd)

	return h.addChildIdsToExternalWorkloadStatus(ctx, externalWorkloadOfOwner, childIdsToAdd)
}

func (h *ExternalWorkloadHandler) objectAdditionalHandler(ctx context.Context, object client.Object,
	topOwner *metav1.PartialObjectMetadata, externalWorkloadOfOwner *v1alpha1.ExternalWorkload,
) error {
	gvk := object.GetObjectKind().GroupVersionKind()
	gvkHandler, found := h.supportedGVKsHandlers[gvk]
	if !found {
		return nil
	}

	additionalObjectHandler, ok := gvkHandler.(AdditionalObjectHandler)
	if !ok {
		return nil
	}

	return additionalObjectHandler.AdditionalHandler(ctx, object, topOwner, externalWorkloadOfOwner)
}

func (h *ExternalWorkloadHandler) updateNetworkStatusIfNeeded(gvkHandler supportedhandlers.TopOwnerHandler, ew *v1alpha1.ExternalWorkload) {
	networkUpdater, ok := gvkHandler.(NetworkUpdateObjectHandler)
	if ok && networkUpdater.ShouldUpdateNetworkStatus() && ew != nil {
		key := types.NamespacedName{Name: ew.Name, Namespace: ew.Namespace}
		// DeepCopy so the consumer can't mutate (or observe partial mutations of)
		// the caller's ExternalWorkload pointer while the caller is still updating it.
		task := utils.NewExternalWorkloadTask(utils.AddOperation, key, ew.GroupVersionKind()).SetObject(ew.DeepCopy())
		h.networkStatusUpdater.TasksQueue().Add(*task)
	}
}

func (h *ExternalWorkloadHandler) enqueueNetworkStatusIfNeeded(gvk schema.GroupVersionKind, ew *v1alpha1.ExternalWorkload) {
	gvkHandler, found := h.supportedGVKsHandlers[gvk]
	if !found {
		return
	}
	h.updateNetworkStatusIfNeeded(gvkHandler, ew)
}

func (h *ExternalWorkloadHandler) triggerTopOwnerHandlerIfNeeded(ctx context.Context,
	ownedObject client.Object, topOwnerMeta *metav1.PartialObjectMetadata, externalWorkloadOfOwner *v1alpha1.ExternalWorkload,
) {
	topOwnerGvk := topOwnerMeta.GetObjectKind().GroupVersionKind()
	topOwnerGvkHandler, found := h.supportedGVKsHandlers[topOwnerGvk]
	if !found {
		return
	}

	triggerTopOwner, ok := topOwnerGvkHandler.(TriggerTopOwner)
	if !ok {
		return
	}

	if !triggerTopOwner.ShouldTriggerTopOwnerHandlerAsWell(ownedObject) {
		return
	}

	topOwner, errObj := triggerTopOwner.GetObject(ctx, topOwnerMeta.GetName(), topOwnerMeta.GetNamespace())
	if errObj != nil || topOwner == nil {
		return
	}

	log.Debug().Msgf("For event of %s, triggering top owner handler as well - %s",
		objectLogStr(ownedObject), objectLogStr(topOwnerMeta))

	var err error
	if externalWorkloadOfOwner != nil {
		err = h.handleExternalWorkloadOfTopOwnedObject(ctx, topOwner, externalWorkloadOfOwner, false)
	} else {
		err = h.handleTopOwnedObject(ctx, topOwner, false)
	}
	if err != nil {
		log.Error().Msgf("Failed to handle top owned %s, error: %s",
			objectLogStr(topOwner), err.Error())
	}
}

// ////////////// API section
func (h *ExternalWorkloadHandler) createExternalWorkloadAndUpdateStatus(ctx context.Context, externalWorkload *v1alpha1.ExternalWorkload) error {
	requiredObservedStatus := externalWorkload.Status.ObservedStatus.DeepCopy()

	err := h.cachedClient.Create(ctx, externalWorkload)
	if err != nil {
		log.Error().Msgf("Failed to create external workload <%s/%s>, error: %s",
			externalWorkload.GetNamespace(), externalWorkload.GetName(), err.Error())
		return err
	}

	log.Info().Msgf("Created externalworkload <%s/%s>", externalWorkload.Namespace, externalWorkload.Name)

	apiServerObservedStatus := externalWorkload.Status.ObservedStatus.DeepCopy()
	externalWorkload.Status.ObservedStatus = *requiredObservedStatus
	return h.updateObservedStatusIfNeeded(ctx, externalWorkload, apiServerObservedStatus)
}

func (h *ExternalWorkloadHandler) updateExternalWorkload(ctx context.Context,
	externalWorkload *v1alpha1.ExternalWorkload,
) error {
	err := h.cachedClient.Update(ctx, externalWorkload)
	if err != nil {
		log.Error().Msgf("Failed to update external workload <%s/%s>, error: %s",
			externalWorkload.GetNamespace(), externalWorkload.GetName(), err.Error())
		return err
	}
	log.Debug().Msgf("Updated externalworkload <%s/%s>", externalWorkload.Namespace, externalWorkload.Name)
	return nil
}

func (h *ExternalWorkloadHandler) patchExternalWorkloadObservedStatusWithoutChildIds(ctx context.Context,
	externalWorkload *v1alpha1.ExternalWorkload,
) error {
	patchBytes, err := updateObservedStatusPatchBytes(&externalWorkload.Status.ObservedStatus)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - "+
			"update observed status of externalworkload <%s/%s>, error: %s",
			externalWorkload.Namespace, externalWorkload.Name, err.Error())
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = h.cachedClient.Status().Patch(ctx, externalWorkload, patch)
	if err != nil {
		log.Error().Msgf("Failed to patch external workload observed status <%s/%s>, error: %s",
			externalWorkload.GetNamespace(), externalWorkload.GetName(), err.Error())
		return err
	}
	log.Debug().Msgf("Patched externalworkload <%s/%s> observed status",
		externalWorkload.Namespace, externalWorkload.Name)
	return nil
}

func (h *ExternalWorkloadHandler) addChildIdsToExternalWorkloadStatus(ctx context.Context,
	externalWorkload *v1alpha1.ExternalWorkload, childIdsToAdd []v1alpha1.ChildId,
) error {
	if len(childIdsToAdd) == 0 {
		return nil
	}

	wasRemoved := utils.RemoveOldChildIdsFromStatus(externalWorkload)
	wasAdded := utils.AddChildIdsToStatusObjectIfNeeded(externalWorkload, childIdsToAdd)

	if !wasAdded && !wasRemoved {
		return nil
	}

	patchBytes, err := updateFullObservedStatusPatchBytes(&externalWorkload.Status.ObservedStatus)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - "+
			"update observed status of externalworkload <%s/%s>, error: %s",
			externalWorkload.Namespace, externalWorkload.Name, err.Error())
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = h.cachedClient.Status().Patch(ctx, externalWorkload, patch)
	if err != nil {
		return err
	}
	log.Debug().Msgf("Added child ids <%s> to externalworkload <%s/%s> observed status",
		childIdsStr(childIdsToAdd), externalWorkload.Namespace, externalWorkload.Name)
	return nil
}

func (h *ExternalWorkloadHandler) deleteChildIdsToExternalWorkloadStatus(ctx context.Context,
	externalWorkload *v1alpha1.ExternalWorkload, childIdsToDelete []v1alpha1.ChildId,
) error {
	if len(childIdsToDelete) == 0 {
		return nil
	}

	wasDeleted := utils.DeleteChildIdsFromStatus(externalWorkload, childIdsToDelete)
	if !wasDeleted {
		return nil
	}

	patchBytes, err := updateFullObservedStatusPatchBytes(&externalWorkload.Status.ObservedStatus)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - "+
			"update observed status of externalworkload <%s/%s>, error: %s",
			externalWorkload.Namespace, externalWorkload.Name, err.Error())
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = h.cachedClient.Status().Patch(ctx, externalWorkload, patch)
	if err != nil {
		return err
	}
	return nil
}

func equalExternalWorkloadSpec(left, right *v1alpha1.ExternalWorkloadSpec) bool {
	return left.ImageName == right.ImageName &&
		left.Group == right.Group &&
		left.Version == right.Version &&
		reflect.DeepEqual(left.RequestedNodePools, right.RequestedNodePools) &&
		reflect.DeepEqual(left.RequestedPods, right.RequestedPods) &&
		reflect.DeepEqual(left.PriorityClassName, right.PriorityClassName) &&
		reflect.DeepEqual(left.Preemptibility, right.Preemptibility) &&
		reflect.DeepEqual(left.Topology, right.Topology) &&
		reflect.DeepEqual(left.ComponentTopologies, right.ComponentTopologies) &&
		reflect.DeepEqual(left.NodeAffinities, right.NodeAffinities) &&
		equalRequestedResources(left.RequestedResources, right.RequestedResources) &&
		reflect.DeepEqual(left.AdditionalFields, right.AdditionalFields) &&
		reflect.DeepEqual(left.StorageRefs, right.StorageRefs) &&
		reflect.DeepEqual(left.Urls, right.Urls)
}

func equalRequestedResources(left, right corev1.ResourceRequirements) bool {
	return equalResourceList(left.Limits, right.Limits) &&
		equalResourceList(left.Requests, right.Requests)
}

func equalResourceList(left, right corev1.ResourceList) bool {
	if len(left) != len(right) {
		return false
	}

	for resourceName, leftResources := range left {
		rightResources, found := right[resourceName]
		if !found {
			return false
		}
		if leftResources.Cmp(rightResources) != 0 {
			return false
		}
	}

	return true
}

func equalExternalWorkloadObservedStatus(left, right *v1alpha1.ExternalWorkloadObservedStatus) bool {
	return left.Phase == right.Phase &&
		left.Message == right.Message &&
		left.K8sStatus == right.K8sStatus &&
		equalExternalWorkloadConditions(left.Conditions, right.Conditions)
}

func equalExternalWorkloadConditions(left, right []v1alpha1.ObservedCondition) bool {
	if len(left) != len(right) {
		return false
	}

	for i := range left {
		leftCondition := left[i]
		rightCondition := right[i]
		if leftCondition.Type != rightCondition.Type ||
			leftCondition.Status != rightCondition.Status ||
			leftCondition.Reason != rightCondition.Reason ||
			leftCondition.Message != rightCondition.Message ||
			!leftCondition.LastTransitionTime.Equal(&rightCondition.LastTransitionTime) {
			return false
		}
	}

	return true
}

func updateFullObservedStatusPatchBytes(observedStatus *v1alpha1.ExternalWorkloadObservedStatus) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"status": map[string]interface{}{
		"observedStatus": map[string]interface{}{
			"phase":                observedStatus.Phase,
			"message":              observedStatus.Message,
			"k8sStatus":            observedStatus.K8sStatus,
			"conditions":           observedStatus.Conditions,
			"childIds":             observedStatus.ChildIds,
			"lastStatusChangeTime": metav1.Now(),
		},
	}})
}

func updateObservedStatusPatchBytes(observedStatus *v1alpha1.ExternalWorkloadObservedStatus) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"status": map[string]interface{}{
		"observedStatus": map[string]interface{}{
			"phase":                observedStatus.Phase,
			"message":              observedStatus.Message,
			"k8sStatus":            observedStatus.K8sStatus,
			"conditions":           observedStatus.Conditions,
			"lastStatusChangeTime": metav1.Now(),
		},
	}})
}

func objectLogStr(object client.Object) string {
	gvk := object.GetObjectKind().GroupVersionKind()
	return fmt.Sprintf("%s/%s <%s/%s>",
		gvk.GroupVersion().String(), gvk.Kind, object.GetNamespace(), object.GetName())
}

func childIdsStr(childIds []v1alpha1.ChildId) string {
	var childIdsStrs []string
	for _, childId := range childIds {
		childIdsStrs = append(childIdsStrs, fmt.Sprintf("%s/%s", childId.Type, childId.UID))
	}
	return strings.Join(childIdsStrs, ",")
}

func fillTopology(annotations map[string]string, externalWorkload *v1alpha1.ExternalWorkload) {
	// Skip if a type-specific handler (e.g. DynamoGraphDeployment) already set the topology.
	if externalWorkload.Spec.Topology.Name != "" ||
		externalWorkload.Spec.Topology.PreferredPlacement != "" ||
		externalWorkload.Spec.Topology.RequiredPlacement != "" {
		return
	}
	if annotations == nil {
		return
	}

	topologyName := annotations[clustercommonconstants.TopologyAnnotation]
	topologyPreferred := annotations[clustercommonconstants.TopologyPreferredPlacementAnnotation]
	topologyRequired := annotations[clustercommonconstants.TopologyRequiredPlacementAnnotation]

	externalWorkload.Spec.Topology.Name = topologyName
	externalWorkload.Spec.Topology.PreferredPlacement = topologyPreferred
	externalWorkload.Spec.Topology.RequiredPlacement = topologyRequired
}
