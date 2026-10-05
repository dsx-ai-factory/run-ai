/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"bytes"
	"cmp"
	"context"
	"slices"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	networkingPluginName    = "networking-plugin"
	externalAccessReadyType = "ExternalAccessReady"
)

func AddWorkloadRunnerHandlerToCache(cachedClient cachedclient.CachedClient,
	externalRLC *cachedclient.RateLimitedController) {
	wlrInformer := cachedClient.GetInformer(cachedclient.WorkloadRunnerType)
	if wlrInformer == nil {
		zap.S().Warn("WorkloadRunner informer not available, skipping WLR handler registration")
		return
	}

	_, err := wlrInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj any, newObj any) {
			oldWlr := oldObj.(*dromosv1alpha1.WorkloadRunner)
			newWlr := newObj.(*dromosv1alpha1.WorkloadRunner)

			if !wlrSyncFieldsChanged(oldWlr, newWlr) {
				return
			}

			triggerExternalWorkloadsForWLR(cachedClient, externalRLC, newWlr)
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to workload runner informer, error: %s", err.Error())
	}

	zap.S().Debug("WorkloadRunner handling has started")
}

func wlrSyncFieldsChanged(oldWlr, newWlr *dromosv1alpha1.WorkloadRunner) bool {
	oldIntent, oldConfigured := oldWlr.Annotations[endpoints.InferenceEndpointAnnotation]
	newIntent, newConfigured := newWlr.Annotations[endpoints.InferenceEndpointAnnotation]
	if oldConfigured != newConfigured || oldIntent != newIntent ||
		(newConfigured && oldWlr.Generation != newWlr.Generation) {
		return true
	}

	if !bytes.Equal(oldWlr.Spec.Template.Raw, newWlr.Spec.Template.Raw) {
		return true
	}

	if oldWlr.Status.Phase != newWlr.Status.Phase {
		return true
	}

	if len(oldWlr.Status.Plugins) != len(newWlr.Status.Plugins) {
		return true
	}

	if !sortedEqual(oldWlr.Status.Issues, newWlr.Status.Issues, compareIssues) {
		return true
	}

	for name, oldPlugin := range oldWlr.Status.Plugins {
		newPlugin, exists := newWlr.Status.Plugins[name]
		if !exists {
			return true
		}
		if (oldPlugin == nil) != (newPlugin == nil) {
			return true
		}
		if oldPlugin != nil {
			if !sortedEqual(oldPlugin.Issues, newPlugin.Issues, compareIssues) {
				return true
			}
			if name == networkingPluginName &&
				(externalAccessConditionChanged(oldPlugin, newPlugin) || pluginDetailsChanged(oldPlugin, newPlugin)) {
				return true
			}
			if name == endpoints.InferenceEndpointPluginName &&
				(oldPlugin.Phase != newPlugin.Phase || inferenceEndpointConditionChanged(oldPlugin, newPlugin) || pluginDetailsChanged(oldPlugin, newPlugin)) {
				return true
			}
		}
	}

	return false
}

func inferenceEndpointConditionChanged(old, new *dromosv1alpha1.PluginStatus) bool {
	oldCond := findCondition(old.Conditions, endpoints.InferenceEndpointReadyCondition)
	newCond := findCondition(new.Conditions, endpoints.InferenceEndpointReadyCondition)
	if oldCond == nil || newCond == nil {
		return oldCond != newCond
	}
	// Transition timestamps alone do not change the routing state reported to users.
	return oldCond.Status != newCond.Status || oldCond.Reason != newCond.Reason ||
		oldCond.Message != newCond.Message || oldCond.ObservedGeneration != newCond.ObservedGeneration
}

func externalAccessConditionChanged(old, new *dromosv1alpha1.PluginStatus) bool {
	oldCond := findCondition(old.Conditions, externalAccessReadyType)
	newCond := findCondition(new.Conditions, externalAccessReadyType)
	oldStatus, oldMsg := conditionFields(oldCond)
	newStatus, newMsg := conditionFields(newCond)
	return oldStatus != newStatus || oldMsg != newMsg
}

func pluginDetailsChanged(old, new *dromosv1alpha1.PluginStatus) bool {
	return !bytes.Equal(pluginDetailsRaw(old), pluginDetailsRaw(new))
}

func pluginDetailsRaw(plugin *dromosv1alpha1.PluginStatus) []byte {
	if plugin == nil || plugin.AdditionalInfo == nil {
		return nil
	}
	return plugin.AdditionalInfo.Raw
}

func findCondition(conditions []metav1.Condition, condType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}

func conditionFields(c *metav1.Condition) (metav1.ConditionStatus, string) {
	if c == nil {
		return "", ""
	}
	return c.Status, c.Message
}

func sortedEqual[T comparable](a, b []T, cmpFunc func(T, T) int) bool {
	if len(a) != len(b) {
		return false
	}
	return slices.Equal(
		slices.SortedFunc(slices.Values(a), cmpFunc),
		slices.SortedFunc(slices.Values(b), cmpFunc),
	)
}

func compareIssues(a, b dromosv1alpha1.Issue) int {
	if c := cmp.Compare(a.Group, b.Group); c != 0 {
		return c
	}
	if c := cmp.Compare(a.Kind, b.Kind); c != 0 {
		return c
	}
	if c := cmp.Compare(a.Name, b.Name); c != 0 {
		return c
	}
	return cmp.Compare(a.Description, b.Description)
}

func triggerExternalWorkloadsForWLR(cachedClient cachedclient.CachedClient,
	externalRLC *cachedclient.RateLimitedController, wlr *dromosv1alpha1.WorkloadRunner) {
	ewList := &v1alpha1.ExternalWorkloadList{}
	err := cachedClient.List(context.Background(), ewList,
		client.InNamespace(wlr.Namespace),
		client.MatchingLabels{commonconstants.WorkloadRunnerNameLabel: wlr.Name})
	if err != nil {
		zap.S().Errorf("Failed to list ExternalWorkloads for WorkloadRunner <%s/%s>, error: %s",
			wlr.Namespace, wlr.Name, err.Error())
		return
	}

	for i := range ewList.Items {
		externalRLC.AddRateLimited(objectKeyFromObject(&ewList.Items[i]))
	}
}
