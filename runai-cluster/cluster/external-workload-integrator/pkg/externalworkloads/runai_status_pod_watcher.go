/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"
	"reflect"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	knativeServicePodLabel  = "serving.knative.dev/service"
	knativeRevisionPodLabel = "serving.knative.dev/revision"
	leaderWorkerSetPodLabel = "leaderworkerset.sigs.k8s.io/name"
)

func setupRunaiStatusPodWatcher(ctx context.Context, cachedClient cachedclient.CachedClient,
	handler *ExternalWorkloadHandler,
) {
	podGVK := corev1.SchemeGroupVersion.WithKind("Pod")
	informer, found := cachedClient.GetObjectInformers()[podGVK]
	if !found {
		var err error
		informer, err = cachedClient.SetInformerForKind(ctx, podGVK)
		if err != nil {
			log.Error().Msgf("RunaiStatusPodWatcher: failed to get Pod informer: %s", err.Error())
			return
		}
	}

	_, err := informer.AddEventHandler(runaiStatusPodEventHandler(ctx, handler))
	if err != nil {
		log.Error().Msgf("RunaiStatusPodWatcher: failed to add Pod event handler: %s", err.Error())
		return
	}
	log.Info().Msg("RunaiStatusPodWatcher: added Pod event handler")
}

func runaiStatusPodEventHandler(ctx context.Context, handler *ExternalWorkloadHandler) clientgocache.ResourceEventHandler {
	return clientgocache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			pod := podFromInformerEvent(obj)
			return pod != nil && (pod.Labels[knativeServicePodLabel] != "" ||
				pod.Labels[leaderWorkerSetPodLabel] != "" ||
				pod.Labels[kubeflowv1.JobNameLabel] != "")
		},
		Handler: clientgocache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				if pod := podFromInformerEvent(obj); pod != nil {
					enqueueRunaiStatusForPod(ctx, handler, pod)
				}
			},
			UpdateFunc: func(oldObj, newObj any) {
				oldPod, oldOK := oldObj.(*corev1.Pod)
				newPod, newOK := newObj.(*corev1.Pod)
				if oldOK && newOK && runaiStatusPodInputsChanged(oldPod, newPod) {
					enqueueRunaiStatusForPod(ctx, handler, newPod)
				}
			},
			DeleteFunc: func(obj any) {
				if pod := podFromInformerEvent(obj); pod != nil {
					enqueueRunaiStatusForPod(ctx, handler, pod)
				}
			},
		},
	}
}

func podFromInformerEvent(obj any) *corev1.Pod {
	switch typed := obj.(type) {
	case *corev1.Pod:
		return typed
	case clientgocache.DeletedFinalStateUnknown:
		pod, _ := typed.Obj.(*corev1.Pod)
		return pod
	default:
		return nil
	}
}

func runaiStatusPodInputsChanged(oldPod, newPod *corev1.Pod) bool {
	if !reflect.DeepEqual(oldPod.Status, newPod.Status) ||
		!reflect.DeepEqual(oldPod.DeletionTimestamp, newPod.DeletionTimestamp) {
		return true
	}
	return oldPod.Annotations[clustercommonconstants.PodStatusAnnotation] !=
		newPod.Annotations[clustercommonconstants.PodStatusAnnotation]
}

func enqueueRunaiStatusForPod(ctx context.Context, handler *ExternalWorkloadHandler, pod *corev1.Pod) {
	for _, workload := range []struct {
		label string
		gvk   schema.GroupVersionKind
	}{
		{knativeServicePodLabel, workloadTypeGVKs[inferenceWorkloadType]},
		{leaderWorkerSetPodLabel, workloadTypeGVKs[distributedInferenceWorkloadType]},
	} {
		name := pod.Labels[workload.label]
		if name == "" {
			continue
		}
		key := types.NamespacedName{Name: name, Namespace: pod.Namespace}
		task := utils.NewExternalWorkloadTask(utils.UpdateOperation, key, workload.gvk)
		handler.tasksQueue.Add(*task)
	}
	enqueueKubeflowRunaiStatusForPod(ctx, handler, pod)
}

func enqueueKubeflowRunaiStatusForPod(ctx context.Context, handler *ExternalWorkloadHandler, pod *corev1.Pod) {
	jobName := pod.Labels[kubeflowv1.JobNameLabel]
	workloadRunnerName := pod.Labels[clustercommonconstants.LabelWorkloadName]
	if jobName == "" || workloadRunnerName == "" || handler.cachedClient == nil {
		return
	}

	ewList := &v1alpha1.ExternalWorkloadList{}
	if err := handler.cachedClient.List(ctx, ewList,
		client.InNamespace(pod.Namespace),
		client.MatchingLabels{
			commonconstants.WorkloadRunnerNameLabel:  workloadRunnerName,
			commonconstants.ConvertedFromNativeLabel: "true",
		}); err != nil {
		log.Warn().Msgf("RunaiStatusPodWatcher: failed listing ExternalWorkloads for Kubeflow pod <%s/%s>: %s",
			pod.Namespace, pod.Name, err.Error())
		return
	}

	for i := range ewList.Items {
		ew := &ewList.Items[i]
		if ew.Spec.OriginalWorkloadName != jobName || ew.Spec.Group != kubeflowGroup {
			continue
		}
		gvk, ok := gvkFromExternalWorkloadOwner(ew)
		if !ok {
			continue
		}
		key := types.NamespacedName{Name: jobName, Namespace: pod.Namespace}
		task := utils.NewExternalWorkloadTask(utils.UpdateOperation, key, gvk)
		handler.tasksQueue.Add(*task)
	}
}

func gvkFromExternalWorkloadOwner(ew *v1alpha1.ExternalWorkload) (schema.GroupVersionKind, bool) {
	if ew.Spec.Group == "" || ew.Spec.Version == "" {
		return schema.GroupVersionKind{}, false
	}
	for _, ownerRef := range ew.OwnerReferences {
		if ownerRef.Kind == "" {
			continue
		}
		ownerGV, err := schema.ParseGroupVersion(ownerRef.APIVersion)
		if err != nil || ownerGV.Group != ew.Spec.Group || ownerGV.Version != ew.Spec.Version {
			continue
		}
		return ownerGV.WithKind(ownerRef.Kind), true
	}
	return schema.GroupVersionKind{}, false
}
