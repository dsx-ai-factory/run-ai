/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/networkstatus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/structure"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/supportedhandlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var statusWatcherGVK = schema.GroupVersionKind{Group: "serving.kserve.io", Version: "v1beta1", Kind: "InferenceService"}

func watcherKarta() *kartav1alpha1.Karta {
	return &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{Name: "kserve", UID: "existing-karta",
			Labels: v1alpha1.GetKartaLabels("test-workload-type", statusWatcherGVK.Group, statusWatcherGVK.Version, statusWatcherGVK.Kind, "old")},
		Spec: kartav1alpha1.KartaSpec{StructureDefinition: kartav1alpha1.StructureDefinition{
			RootComponent: kartav1alpha1.ComponentDefinition{Name: "service",
				Kind: &kartav1alpha1.GroupVersionKind{Group: statusWatcherGVK.Group, Version: statusWatcherGVK.Version, Kind: statusWatcherGVK.Kind},
				StatusDefinition: &kartav1alpha1.StatusDefinition{
					PhaseDefinition: &kartav1alpha1.PhaseDefinition{Path: ".status.phase"},
					StatusMappings:  kartav1alpha1.StatusMappings{Running: []kartav1alpha1.StatusMatcher{{ByPhase: "old-running"}}},
				},
			},
		}},
	}
}

func updatedWatcherKarta() *kartav1alpha1.Karta {
	k := watcherKarta()
	k.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings = kartav1alpha1.StatusMappings{
		Running: []kartav1alpha1.StatusMatcher{{ByPhase: "Ready"}},
		Failed:  []kartav1alpha1.StatusMatcher{{ByPhase: "Broken"}},
	}
	return k
}

func watcherWorkload(namespace, name string) *v1alpha1.ExternalWorkload {
	return &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{Name: "ew-" + name, Namespace: namespace, UID: types.UID("ew-" + name)},
		Spec: v1alpha1.ExternalWorkloadSpec{Group: statusWatcherGVK.Group, Version: statusWatcherGVK.Version,
			Type: statusWatcherGVK.Kind, OriginalWorkloadName: name, OriginalWorkloadUID: types.UID(name)},
	}
}

func watcherClient(t *testing.T, objects ...client.Object) client.WithWatch {
	t.Helper()
	g := NewWithT(t)
	scheme := runtime.NewScheme()
	g.Expect(kartav1alpha1.AddToScheme(scheme)).To(Succeed())
	g.Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	g.Expect(corev1.AddToScheme(scheme)).To(Succeed())
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).
		WithStatusSubresource(&v1alpha1.ExternalWorkload{}).
		WithIndex(&v1alpha1.ExternalWorkload{}, constants.ExternalWorkloadByOriginalUIDIndexerName, cachedclient.ExternalWorkloadByOriginalWorkloadUIDIndexer).Build()
}

func watcherHandler(t *testing.T, c client.Client) *ExternalWorkloadHandler {
	t.Helper()
	h := NewExternalWorkloadHandler(c, nil, false, map[schema.GroupVersionKind]supportedhandlers.TopOwnerHandler{}, nil,
		[]schema.GroupVersionKind{statusWatcherGVK}, networkstatus.NewNetworkStatusUpdater(c), structure.NewStructureUpdater(c))
	t.Cleanup(h.tasksQueue.ShutDown)
	t.Cleanup(h.networkStatusUpdater.TasksQueue().ShutDown)
	return h
}

func TestKartaStatusWatcherFiltersEvents(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*kartav1alpha1.Karta)
		want   int
	}{
		{"mapping change", func(k *kartav1alpha1.Karta) { k.Spec = updatedWatcherKarta().Spec }, 1},
		{"labels only", func(k *kartav1alpha1.Karta) { k.Labels["updated-at"] = "new" }, 0},
		{"status only", func(k *kartav1alpha1.Karta) {
			k.Status.Conditions = []metav1.Condition{{Type: "RunaiReady", Status: metav1.ConditionTrue}}
		}, 0},
		{"deleting", func(k *kartav1alpha1.Karta) {
			k.Spec = updatedWatcherKarta().Spec
			k.DeletionTimestamp = &metav1.Time{Time: time.Now()}
		}, 0},
		{"replacement identity", func(k *kartav1alpha1.Karta) { k.Spec = updatedWatcherKarta().Spec; k.UID = "replacement" }, 0},
		{"different resource type", func(k *kartav1alpha1.Karta) {
			k.Spec = updatedWatcherKarta().Spec
			k.Spec.StructureDefinition.RootComponent.Kind.Kind = "Other"
		}, 0},
		{"removed status definition", func(k *kartav1alpha1.Karta) { k.Spec.StructureDefinition.RootComponent.StatusDefinition = nil }, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			q := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[utils.ExternalWorkloadTask]())
			t.Cleanup(q.ShutDown)
			h := &ExternalWorkloadHandler{tasksQueue: q}
			old, next := watcherKarta(), watcherKarta()
			tc.change(next)
			kartaStatusEventHandler(h).OnUpdate(old, next)
			g.Expect(q.Len()).To(Equal(tc.want))
		})
	}
}

func TestKartaStatusRefreshSelectsMatchingWorkloads(t *testing.T) {
	g := NewWithT(t)
	matching := watcherWorkload("team-a", "healthy")
	legacy := watcherWorkload("team-b", "legacy")
	legacy.Spec.Version = ""
	otherGroup := watcherWorkload("team-a", "other-group")
	otherGroup.Spec.Group = "other.example"
	otherVersion := watcherWorkload("team-a", "other-version")
	otherVersion.Spec.Version = "v2"
	otherKind := watcherWorkload("team-a", "other-kind")
	otherKind.Spec.Type = "Other"
	alias := watcherWorkload("team-b", "alias")
	alias.Spec.Type = "CustomDisplayName"
	alias.Spec.OriginalWorkloadKind = statusWatcherGVK.Kind
	deleting := watcherWorkload("team-a", "deleting")
	deleting.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	deleting.Finalizers = []string{"test-finalizer"}
	c := watcherClient(t, updatedWatcherKarta(), matching, legacy, alias, otherGroup, otherVersion, otherKind, deleting)
	h := watcherHandler(t, c)
	kartaStatusEventHandler(h).OnUpdate(watcherKarta(), updatedWatcherKarta())
	g.Expect(processNextItem(context.Background(), h)).To(BeTrue())
	g.Expect(h.tasksQueue.Len()).To(Equal(3))
	names := []string{}
	for h.tasksQueue.Len() > 0 {
		task, _ := h.tasksQueue.Get()
		names = append(names, task.Key.Namespace+"/"+task.Key.Name)
		g.Expect(task.GetOperation()).To(Equal(string(utils.UpdateOperation)))
		g.Expect(task.GVK).To(Equal(statusWatcherGVK))
		g.Expect(task.ExternalWorkloadName).To(Equal("ew-" + task.Key.Name))
		h.tasksQueue.Done(task)
	}
	g.Expect(names).To(ConsistOf("team-a/healthy", "team-b/legacy", "team-b/alias"))
}

func TestKartaStatusRefreshSkipsUnavailableKarta(t *testing.T) {
	for _, state := range []string{"missing", "deleting", "unregistered resource"} {
		t.Run(state, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			karta := updatedWatcherKarta()
			ew := watcherWorkload("team-a", "model")
			objects := []client.Object{ew}
			if state == "deleting" {
				karta.DeletionTimestamp = &metav1.Time{Time: time.Now()}
				karta.Finalizers = []string{"karta.run.ai/finalize"}
			}
			if state != "missing" {
				objects = append(objects, karta)
			}
			c := watcherClient(t, objects...)
			h := watcherHandler(t, c)
			if state == "unregistered resource" {
				delete(h.supportedGVKsHandlers, statusWatcherGVK)
			}
			// The Karta may become unavailable after its update was enqueued.
			kartaStatusEventHandler(h).OnUpdate(watcherKarta(), updatedWatcherKarta())
			g.Expect(processNextItem(ctx, h)).To(BeTrue())
			g.Expect(h.tasksQueue.Len()).To(BeZero())
			stored := &v1alpha1.ExternalWorkload{}
			g.Expect(c.Get(ctx, client.ObjectKeyFromObject(ew), stored)).To(Succeed())
			g.Expect(stored.UID).To(Equal(ew.UID))
			g.Expect(stored.DeletionTimestamp.IsZero()).To(BeTrue())
		})
	}
}

func TestKartaStatusRefreshUpdatesIdleWorkload(t *testing.T) {
	for _, tc := range []struct {
		phase    string
		expected v1alpha1.ExternalWorkloadPhase
	}{
		{"Ready", v1alpha1.ExWRunningPhase}, {"Broken", v1alpha1.ExWFailedPhase},
	} {
		t.Run(tc.phase, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			source := &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": statusWatcherGVK.GroupVersion().String(), "kind": statusWatcherGVK.Kind,
				"metadata": map[string]any{"name": "model", "namespace": "team-a", "uid": "model"},
				"status":   map[string]any{"phase": tc.phase},
			}}
			ew := watcherWorkload("team-a", "model")
			ew.Status.ObservedStatus.Phase = v1alpha1.ExWUnknownPhase
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "team-a"}}
			c := watcherClient(t, watcherKarta(), source, ew, ns)
			h := watcherHandler(t, c)
			before := source.DeepCopy()
			g.Expect(c.Get(ctx, client.ObjectKeyFromObject(source), before)).To(Succeed())
			updated := &kartav1alpha1.Karta{}
			g.Expect(c.Get(ctx, client.ObjectKey{Name: "kserve"}, updated)).To(Succeed())
			updated.Spec = updatedWatcherKarta().Spec
			g.Expect(c.Update(ctx, updated)).To(Succeed())
			// The only event is the mapping update; the source workload stays idle.
			kartaStatusEventHandler(h).OnUpdate(watcherKarta(), updated)
			g.Expect(h.tasksQueue.Len()).To(Equal(1))
			g.Expect(processNextItem(ctx, h)).To(BeTrue())
			g.Expect(h.tasksQueue.Len()).To(Equal(1))
			g.Expect(processNextItem(ctx, h)).To(BeTrue())
			stored := &v1alpha1.ExternalWorkload{}
			g.Expect(c.Get(ctx, client.ObjectKeyFromObject(ew), stored)).To(Succeed())
			g.Expect(stored.Status.ObservedStatus.Phase).To(Equal(tc.expected))
			g.Expect(stored.UID).To(Equal(ew.UID))
			after := source.DeepCopy()
			g.Expect(c.Get(ctx, client.ObjectKeyFromObject(source), after)).To(Succeed())
			g.Expect(after).To(Equal(before))
		})
	}
}

func TestKartaStatusRefreshRetriesTransientListFailure(t *testing.T) {
	g := NewWithT(t)
	c := watcherClient(t, updatedWatcherKarta(), watcherWorkload("team-a", "model"))
	failed := false
	failing := interceptor.NewClient(c, interceptor.Funcs{List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
		if _, ok := list.(*v1alpha1.ExternalWorkloadList); ok && !failed {
			failed = true
			return fmt.Errorf("temporary list failure")
		}
		return c.List(ctx, list, opts...)
	}})
	h := watcherHandler(t, failing)
	kartaStatusEventHandler(h).OnUpdate(watcherKarta(), updatedWatcherKarta())
	g.Expect(processNextItem(context.Background(), h)).To(BeTrue())
	g.Eventually(h.tasksQueue.Len, time.Second).Should(Equal(1))
	g.Expect(processNextItem(context.Background(), h)).To(BeTrue())
	g.Expect(h.tasksQueue.Len()).To(Equal(1))
	task, _ := h.tasksQueue.Get()
	g.Expect(task.GetOperation()).To(Equal(string(utils.UpdateOperation)))
	g.Expect(task.Key.Name).To(Equal("model"))
	h.tasksQueue.Done(task)
}
