/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/metrics"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var elementIdLabelKeys = []string{
	fmt.Sprintf("%s1-id", commonconstants.ElementLabelPrefix),
	fmt.Sprintf("%s2-id", commonconstants.ElementLabelPrefix),
	fmt.Sprintf("%s3-id", commonconstants.ElementLabelPrefix),
}

type PodController struct {
	client client.Client
}

func NewPodController(client client.Client) *PodController {
	return &PodController{
		client: client,
	}
}

func (c *PodController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(PodsEventFilter()).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
			utils.RateLimiterBaseDelay, utils.RateLimiterMaxDelay)}).
		Complete(c)
}

func (c *PodController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)
	log.Debug().Msgf("Reconciling pod %s/%s", req.Name, req.Namespace)

	pod := &corev1.Pod{}
	isObjectFound, err := utils.GetObjectForRequest(ctx, req, "pod", pod, c.client)
	if err != nil {
		return ctrl.Result{}, err
	} else if !isObjectFound {
		metrics.ResetPodMetrics(req.Name, req.Namespace)
		return ctrl.Result{Requeue: false}, nil
	}

	err = metrics.SetPodMetrics(ctx, pod, c.client)
	if err != nil {
		log.Error().Msgf("Failed to set metrics for pod %s/%s, error: %s", pod.Namespace, pod.Name, err.Error())
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func PodsEventFilter() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isRunaiPod(e.Object.(*corev1.Pod))
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return filterUpdateEvents(e.ObjectOld.(*corev1.Pod), e.ObjectNew.(*corev1.Pod))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isRunaiPod(e.Object.(*corev1.Pod))
		},
	}
}

func filterUpdateEvents(old, new *corev1.Pod) bool {
	return isRunaiPod(new) && (old.Status.Phase != new.Status.Phase ||
		old.Annotations[commonconstants.PodStatusAnnotation] != new.Annotations[commonconstants.PodStatusAnnotation] ||
		old.Spec.NodeName != new.Spec.NodeName ||
		old.Labels[v1alpha1.NodePoolNameLabel] != new.Labels[v1alpha1.NodePoolNameLabel] ||
		hasElementIdLabelAdded(old.Labels, new.Labels))
}

func isRunaiPod(pod *corev1.Pod) bool {
	return pod.Spec.SchedulerName == commonconstants.RunaiScheduler
}

func hasElementIdLabelAdded(oldLabels, newLabels map[string]string) bool {
	for _, labelKey := range elementIdLabelKeys {
		if oldLabels[labelKey] != newLabels[labelKey] {
			return true
		}
	}
	return false
}
