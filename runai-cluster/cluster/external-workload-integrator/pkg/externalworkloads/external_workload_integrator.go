/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/networkstatus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/structure"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/supportedhandlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	supportedworkloads "github.com/dsx-ai-factory/run-ai/runai-cluster/common/supported-workloads"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const TaskHandlingMaxRetryAttempts = 10

type EventFilterHandler interface {
	// Filter - some types can be filtered out before it gets to the event handling. this filter is doing
	// scheduler-related filtering, for example, if the object is not scheduled by the runai-scheduler, it should
	// be filtered out. non-pod objects are filtered in by default if they are in a namespace where runai-scheduler is
	// enforced. thus, in such cases, this filter is not even applied.
	Filter(object client.Object) bool
}

type SecondaryEventFilterHandler interface {
	// SecondaryFilter - some types can be filtered out before it gets to the event handling. this filter is doing
	// a secondary filtering, after the primary filter that always exists and is doing scheduler-related filtering.
	// for example, we can filter out objects that have no pods owned by them.
	SecondaryFilter(object client.Object) bool
}

type GetOriginalWorkloadNameUidHandler interface {
	// GetOriginalWorkloadNameUid - some types require special name and uid setting for the external-workload,
	// which is different from the ones of the object
	GetOriginalWorkloadNameUid(object client.Object) (types.UID, string, error)
}

type AdditionalObjectHandler interface {
	// AdditionalHandler - some types require additional handling of the object itself - regardless of the external-workload
	AdditionalHandler(ctx context.Context, object client.Object, topOwner *metav1.PartialObjectMetadata,
		externalWorkloadOfOwner *v1alpha1.ExternalWorkload) error
}

type NetworkUpdateObjectHandler interface {
	// ShouldUpdateNetworkStatus - some types require network-status recalculation
	ShouldUpdateNetworkStatus() bool
}

type TriggerTopOwner interface {
	// ShouldTriggerTopOwnerHandlerAsWell - some types should trigger the top owner handler as well
	// for example - for pod events, if the top owner is a deployment, we want to trigger the deployment handler as well.
	ShouldTriggerTopOwnerHandlerAsWell(triggeredOwnedObject client.Object) bool

	GetObject(ctx context.Context, name, namespace string) (client.Object, error)
}

type OwnerReferencesHandler interface {
	// GetOwnerReferences - some types require multiple owner references, not just the single external resource from which
	// the external workload is created.
	GetOwnerReferences(object client.Object) []metav1.OwnerReference
}

func StartExternalWorkloadIntegrator(ctx context.Context, stopCh <-chan struct{},
	cachedClient cachedclient.CachedClient, draParser *dra.DRAParser, allowNonSupportedTypes bool, additionalGVKs []schema.GroupVersionKind,
) *ExternalWorkloadHandler {
	supportedGVKsHandlers := supportedhandlers.GetSupportedGVKsHandlers(cachedClient, draParser)
	onlyOwnedSupportedGVKs := map[schema.GroupVersionKind]bool{
		{Group: "serving.knative.dev", Version: "v1", Kind: "Revision"}:      true,
		{Group: "serving.knative.dev", Version: "v1", Kind: "Configuration"}: true,
	}

	err := supportedworkloads.ValidateSupportedGVKs(utils.ConvertGVKMap(supportedGVKsHandlers), false)
	if err != nil {
		log.Fatal().Msgf("ExternalWorkloadIntegrator supported GVKs error: %s", err.Error())
	}

	cachedClient.SetInformers(ctx, append(utils.GetGVKKeys(supportedGVKsHandlers),
		utils.GetGVKKeys(onlyOwnedSupportedGVKs)...), additionalGVKs)

	networkUpdater := networkstatus.NewNetworkStatusUpdater(cachedClient)
	structureUpdater := structure.NewStructureUpdater(cachedClient)

	externalWorkloadHandler := NewExternalWorkloadHandler(cachedClient, draParser, allowNonSupportedTypes,
		supportedGVKsHandlers, onlyOwnedSupportedGVKs, additionalGVKs, networkUpdater, structureUpdater)

	addEventHandlers(ctx, cachedClient.GetObjectInformers(), externalWorkloadHandler)

	setupNetworkResourceWatchers(ctx, cachedClient, networkUpdater)

	// Watch WorkloadRunners (reference-only) so wlr-only changes refresh the computed status.
	setupWorkloadRunnerWatcher(ctx, cachedClient, externalWorkloadHandler)
	// The Run:ai Inference and DistributedInference calculators consume pod status directly.
	setupRunaiStatusPodWatcher(ctx, cachedClient, externalWorkloadHandler)

	go cachedClient.StartInformers(ctx)

	externalWorkloadHandler.InitMetricsExporter()

	log.Info().Msgf("External Workload Integrator has started watching 3-rd party workload objects")
	cachedClient.WaitForCachesToSync(ctx, stopCh)

	go executeTasks(ctx, externalWorkloadHandler)
	go networkUpdater.ExecuteTasks(ctx)
	go structureUpdater.ExecuteTasks(ctx)

	return externalWorkloadHandler
}

func (h *ExternalWorkloadHandler) handleTask(ctx context.Context, obj client.Object, task utils.ExternalWorkloadTask, handler utils.HandleFunc, operation utils.TaskOperation) {
	err := handler(ctx, obj, task.GVK, operation)
	// If succeeded or Already used all attempts - Forget()
	if err == nil || h.tasksQueue.NumRequeues(task) > TaskHandlingMaxRetryAttempts {
		h.tasksQueue.Forget(task)
	} else {
		h.metricsExporter.AddErrorCounter(err)
		// Add back to Queue
		h.tasksQueue.AddRateLimited(task)
	}
}

func processNextItem(ctx context.Context, ewh *ExternalWorkloadHandler) bool {
	// Wait until there is a new item in the working queue
	task, quit := ewh.tasksQueue.Get()
	if quit {
		log.Info().Msgf("External Workload Integrator Gracefully shutting down")
		return false
	}
	defer ewh.tasksQueue.Done(task)

	log.Debug().Msgf("Start Task <%s> for <%s>", task.GetOperation(), task.GVK.String())
	switch task.GetOperation() {
	case string(utils.AddOperation):
		obj, err := ewh.GetFreshObject(ctx, task.Key, task.GVK)
		if err != nil || utils.IsObjectNilOrEmpty(obj) {
			obj = task.Object
		}
		ewh.handleTask(ctx, obj, task, ewh.HandleWorkloadAddUpdateEvent, utils.AddOperation)
	case string(utils.DeleteOperation):
		ewh.handleTask(ctx, task.Object, task, ewh.HandleDeleteEvent, utils.DeleteOperation)
	case string(utils.UpdateOperation):
		obj, err := ewh.GetFreshObject(ctx, task.Key, task.GVK)
		if err != nil || utils.IsObjectNilOrEmpty(obj) {
			ewh.handleMissingChild(ctx, task)
			return true
		}
		ewh.handleTask(ctx, obj, task, ewh.HandleWorkloadAddUpdateEvent, utils.UpdateOperation)
	}
	log.Debug().Msgf("Complete Task <%s> for <%s>", task.GetOperation(), task.GVK.String())
	return true
}

func executeTasks(ctx context.Context, externalWorkloadHandler *ExternalWorkloadHandler) {
	for processNextItem(ctx, externalWorkloadHandler) {
	}
}

func addEventHandlers(ctx context.Context, objectInformers map[schema.GroupVersionKind]cachedclient.CachedInformer,
	externalWorkloadHandler *ExternalWorkloadHandler,
) {
	for gvk, objectInformer := range objectInformers {
		addEventHandlerForObjectType(ctx, gvk, objectInformer, externalWorkloadHandler)
	}
}

func isAllowedForDeletion(object any) bool {
	_, ok := object.(*corev1.Pod)
	return ok
}

func addEventHandlerForObjectType(ctx context.Context, gvk schema.GroupVersionKind,
	objectInformer cachedclient.CachedInformer, externalWorkloadHandler *ExternalWorkloadHandler,
) {
	_, err := objectInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			return externalWorkloadHandler.EventHandlerFilter(ctx, obj.(client.Object), gvk)
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(addedObj any) {
				obj := addedObj.(client.Object)
				key := types.NamespacedName{
					Name:      obj.GetName(),
					Namespace: obj.GetNamespace(),
				}
				objCopy := obj.DeepCopyObject().(client.Object)
				task := utils.NewExternalWorkloadTask(utils.AddOperation, key, gvk).SetObject(objCopy)
				externalWorkloadHandler.tasksQueue.Add(*task)
			},
			UpdateFunc: func(oldObj any, newObj any) {
				obj := newObj.(client.Object)
				key := types.NamespacedName{
					Name:      obj.GetName(),
					Namespace: obj.GetNamespace(),
				}
				task := utils.NewExternalWorkloadTask(utils.UpdateOperation, key, gvk)
				externalWorkloadHandler.tasksQueue.Add(*task)
			},
			DeleteFunc: func(deleted any) {
				obj := deleted.(client.Object)
				// Part of the hotfix - handling deletion only for Pods
				if !isAllowedForDeletion(obj) {
					return
				}
				key := types.NamespacedName{
					Name:      obj.GetName(),
					Namespace: obj.GetNamespace(),
				}
				task := utils.NewExternalWorkloadTask(utils.DeleteOperation, key, gvk).SetObject(obj)
				externalWorkloadHandler.tasksQueue.Add(*task)
			},
		},
	})

	if err != nil {
		log.Fatal().Msgf("Failed to add event handler for gvk <%s>, error: %s",
			gvk.String(), err.Error())
	} else {
		log.Info().Msgf("Added event handler for gvk <%s>", gvk.String())
	}
}
