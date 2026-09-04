/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package seed_cluster

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
	workload_scale "github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/workload-scale"
	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
	"time"
)

type K8sController struct {
	workloadSampler *WorkloadSampler
	projectsPrefix  string
	client          client.Client
}

func NewK8sController(workloadSampler *WorkloadSampler, projectsPrefix string, client client.Client) *K8sController {
	return &K8sController{
		workloadSampler: workloadSampler,
		projectsPrefix:  fmt.Sprintf("runai-%s", projectsPrefix),
		client:          client,
	}
}
func (r *K8sController) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, retErr error) {
	pod, err := getPodForRequest(ctx, r.client, req)
	if err != nil {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 2 * time.Minute,
		}, nil
	}
	podSampleData := getPodsMetrics(pod)
	r.workloadSampler.Pods = append(r.workloadSampler.Pods, podSampleData)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *K8sController) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Watches(&corev1.Pod{}, &handler.EnqueueRequestForObject{})
	builder.WithEventFilter(predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldPod, okOld := e.ObjectOld.(*corev1.Pod)
			newPod, okNew := e.ObjectNew.(*corev1.Pod)
			if !okOld || !okNew {
				return false
			}
			return strings.HasPrefix(newPod.Namespace, r.projectsPrefix) && oldPod.Status.Phase != corev1.PodSucceeded && newPod.Status.Phase == corev1.PodSucceeded
		},
		CreateFunc: func(e event.CreateEvent) bool {
			pod, ok := e.Object.(*corev1.Pod)
			return ok && strings.HasPrefix(pod.Namespace, r.projectsPrefix) && pod.Status.Phase == corev1.PodSucceeded
		},
	})
	return builder.Complete(r)
}

func getPodForRequest(ctx context.Context, client client.Client, req ctrl.Request) (*corev1.Pod, error) {
	pod := &corev1.Pod{}
	err := client.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		return nil, err
	}
	return pod, nil
}

func getPodsMetrics(pod *corev1.Pod) *workload_scale.PodK8sData {
	workloadName := pod.Labels["workloadName"]
	var podBoundConditionTime *time.Time
	for _, condition := range pod.Status.Conditions {
		if condition.Type == clustercommonconstants.PodBoundCondition {
			podBoundConditionTime = &condition.LastTransitionTime.Time
		}
	}
	return &workload_scale.PodK8sData{
		WorkloadName:  workloadName,
		CreationTime:  scale.GetTimeInUtc(&pod.CreationTimestamp.Time),
		ScheduledTime: scale.GetTimeInUtc(podBoundConditionTime),
	}
}
