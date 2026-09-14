/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	trainingWorkloadType             = "Training"
	workspaceWorkloadType            = "Workspace"
	inferenceWorkloadType            = "Inference"
	distributedInferenceWorkloadType = "DistributedInference"
	kubeflowGroup                    = "kubeflow.org"
)

var workloadTypeGVKs = map[string]schema.GroupVersionKind{
	trainingWorkloadType:             batchv1.SchemeGroupVersion.WithKind("Job"),
	workspaceWorkloadType:            batchv1.SchemeGroupVersion.WithKind("Job"),
	inferenceWorkloadType:            {Group: "serving.knative.dev", Version: "v1", Kind: "Service"},
	distributedInferenceWorkloadType: {Group: "leaderworkerset.x-k8s.io", Version: "v1", Kind: "LeaderWorkerSet"},
}

// setupWorkloadRunnerWatcher re-enqueues a wlr's child Job on status change (reference-only; never a top owner).
func setupWorkloadRunnerWatcher(ctx context.Context, cachedClient cachedclient.CachedClient,
	handler *ExternalWorkloadHandler) {
	gvk := dromosv1alpha1.SchemeGroupVersion.WithKind(dromosv1alpha1.WorkloadRunnerKind)
	informer, err := cachedClient.SetInformerForKind(ctx, gvk)
	if err != nil {
		if meta.IsNoMatchError(err) {
			log.Warn().Msgf("WorkloadRunnerWatcher: GVK <%s> not available in cluster (CRD not installed), skipping", gvk.String())
			return
		}
		log.Error().Msgf("WorkloadRunnerWatcher: failed to get informer for GVK <%s>: %s", gvk.String(), err.Error())
		return
	}

	_, err = informer.AddEventHandler(clientgocache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			if wlr, ok := obj.(*dromosv1alpha1.WorkloadRunner); ok {
				enqueueWorkloadsForWorkloadRunner(ctx, cachedClient, handler, wlr)
			}
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldWlr, ok1 := oldObj.(*dromosv1alpha1.WorkloadRunner)
			newWlr, ok2 := newObj.(*dromosv1alpha1.WorkloadRunner)
			if !ok1 || !ok2 || !wlrStatusRelevantChanged(oldWlr, newWlr) {
				return
			}
			enqueueWorkloadsForWorkloadRunner(ctx, cachedClient, handler, newWlr)
		},
	})
	if err != nil {
		log.Error().Msgf("WorkloadRunnerWatcher: failed to add event handler for GVK <%s>: %s", gvk.String(), err.Error())
		return
	}
	log.Info().Msgf("WorkloadRunnerWatcher: added event handler for GVK <%s>", gvk.String())
}

func enqueueWorkloadsForWorkloadRunner(ctx context.Context, cachedClient client.Client,
	handler *ExternalWorkloadHandler, wlr *dromosv1alpha1.WorkloadRunner) {
	ewList := &v1alpha1.ExternalWorkloadList{}
	if err := cachedClient.List(ctx, ewList,
		client.InNamespace(wlr.Namespace),
		client.MatchingLabels{commonconstants.WorkloadRunnerNameLabel: wlr.Name}); err != nil {
		log.Error().Msgf("WorkloadRunnerWatcher: failed listing ExternalWorkloads for WorkloadRunner <%s/%s>: %s",
			wlr.Namespace, wlr.Name, err.Error())
		return
	}

	for i := range ewList.Items {
		ew := &ewList.Items[i]
		if ew.Spec.OriginalWorkloadName == "" {
			continue
		}
		gvk, supported := workloadTypeGVKs[ew.Spec.Type]
		if !supported {
			if ew.Spec.Group != kubeflowGroup {
				continue
			}
			var ok bool
			gvk, ok = resolveChildGVK(ew)
			if !ok {
				continue
			}
		}

		key := types.NamespacedName{Name: ew.Spec.OriginalWorkloadName, Namespace: ew.Namespace}
		task := utils.NewExternalWorkloadTask(utils.UpdateOperation, key, gvk).SetExternalWorkloadName(ew.Name)
		handler.tasksQueue.Add(*task)
	}
}

// wlrStatusRelevantChanged reports whether status used by a Run:ai phase calculator changed.
func wlrStatusRelevantChanged(oldWlr, newWlr *dromosv1alpha1.WorkloadRunner) bool {
	if oldWlr.Status.Phase != newWlr.Status.Phase {
		return true
	}
	return statusInputSignature(oldWlr) != statusInputSignature(newWlr)
}

func statusInputSignature(wlr *dromosv1alpha1.WorkloadRunner) string {
	parts := make([]string, 0)
	addIssues := func(owner string, issues []dromosv1alpha1.Issue) {
		for _, currentIssue := range issues {
			parts = append(parts, fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s\x00%t",
				owner, currentIssue.Kind, currentIssue.Name, currentIssue.Description,
				currentIssue.Group, currentIssue.IsCritical))
		}
	}
	addIssues("controller", wlr.Status.Issues)
	for name, plugin := range wlr.Status.Plugins {
		if plugin == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("plugin-phase\x00%s\x00%s", name, plugin.Phase))
		addIssues("plugin:"+name, plugin.Issues)
	}
	sort.Strings(parts)
	return strings.Join(parts, "\x01")
}
