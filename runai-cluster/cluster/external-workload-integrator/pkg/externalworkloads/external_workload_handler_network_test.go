/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/networkstatus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/structure"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/supportedhandlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const postUpdateURL = "http://post-update.example"

type networkUpdatingTopOwnerHandler struct{}

func (networkUpdatingTopOwnerHandler) Parse(
	_ context.Context,
	_ client.Object,
	externalWorkload *v1alpha1.ExternalWorkload,
) error {
	externalWorkload.Spec.Urls = []string{postUpdateURL}
	return nil
}

func (networkUpdatingTopOwnerHandler) GetObject(
	_ context.Context,
	_, _ string,
) (client.Object, error) {
	return nil, nil
}

func (networkUpdatingTopOwnerHandler) ShouldUpdateNetworkStatus() bool {
	return true
}

var _ = Describe("top-owner network status enqueue", func() {
	DescribeTable("enqueues one task carrying the post-update ExternalWorkload",
		func(addedResource bool) {
			ctx := context.Background()
			lwsGVK := schema.GroupVersionKind{
				Group: "leaderworkerset.x-k8s.io", Version: "v1", Kind: "LeaderWorkerSet",
			}
			lwsUID := types.UID("lws-uid")

			lws := &unstructured.Unstructured{}
			lws.SetGroupVersionKind(lwsGVK)
			lws.SetName("lws")
			lws.SetNamespace("ns")
			lws.SetUID(lwsUID)

			externalWorkload := &v1alpha1.ExternalWorkload{
				ObjectMeta: metav1.ObjectMeta{Name: "ew", Namespace: "ns"},
				Spec: v1alpha1.ExternalWorkloadSpec{
					OriginalWorkloadUID: lwsUID,
					Type:                "test",
					Urls:                []string{"http://pre-update.example"},
				},
			}

			scheme := runtime.NewScheme()
			Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
			cachedClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(externalWorkload).
				WithStatusSubresource(&v1alpha1.ExternalWorkload{}).
				WithIndex(
					&v1alpha1.ExternalWorkload{},
					constants.ExternalWorkloadByOriginalUIDIndexerName,
					cachedclient.ExternalWorkloadByOriginalWorkloadUIDIndexer,
				).
				Build()

			networkStatusUpdater := networkstatus.NewNetworkStatusUpdater(cachedClient)
			handler := &ExternalWorkloadHandler{
				cachedClient: cachedClient,
				supportedGVKsHandlers: map[schema.GroupVersionKind]supportedhandlers.TopOwnerHandler{
					lwsGVK: networkUpdatingTopOwnerHandler{},
				},
				onlyOwnedSupportedGVKs: map[schema.GroupVersionKind]bool{},
				networkStatusUpdater:   networkStatusUpdater,
			}

			Expect(handler.handleTopOwnedObject(ctx, lws, addedResource)).To(Succeed())
			Expect(networkStatusUpdater.TasksQueue().Len()).To(Equal(1))

			task, shutdown := networkStatusUpdater.TasksQueue().Get()
			Expect(shutdown).To(BeFalse())
			queuedExternalWorkload, ok := task.Object.(*v1alpha1.ExternalWorkload)
			Expect(ok).To(BeTrue())
			Expect(queuedExternalWorkload.Spec.Urls).To(Equal([]string{postUpdateURL}))
			networkStatusUpdater.TasksQueue().Done(task)
		},
		Entry("for an update event", false),
		Entry("for an add event with an existing ExternalWorkload", true),
	)

	DescribeTable("enqueues one task for an LWS-owned Pod event",
		func(addedResource bool) {
			ctx := context.Background()
			lwsGVK := schema.GroupVersionKind{
				Group: "leaderworkerset.x-k8s.io", Version: "v1", Kind: "LeaderWorkerSet",
			}
			lwsUID := types.UID("lws-uid")

			lws := &unstructured.Unstructured{}
			lws.SetGroupVersionKind(lwsGVK)
			lws.SetName("lws")
			lws.SetNamespace("ns")
			lws.SetUID(lwsUID)

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "ns",
					OwnerReferences: []metav1.OwnerReference{{
						APIVersion: lwsGVK.GroupVersion().String(),
						Kind:       lwsGVK.Kind,
						Name:       lws.GetName(),
						UID:        lwsUID,
					}},
				},
			}

			externalWorkload := &v1alpha1.ExternalWorkload{
				ObjectMeta: metav1.ObjectMeta{Name: "ew", Namespace: "ns"},
				Spec: v1alpha1.ExternalWorkloadSpec{
					OriginalWorkloadUID: lwsUID,
					Type:                "test",
				},
			}

			scheme := runtime.NewScheme()
			Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
			cachedClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(externalWorkload, lws).
				WithStatusSubresource(&v1alpha1.ExternalWorkload{}).
				WithIndex(
					&v1alpha1.ExternalWorkload{},
					constants.ExternalWorkloadByOriginalUIDIndexerName,
					cachedclient.ExternalWorkloadByOriginalWorkloadUIDIndexer,
				).
				Build()

			networkStatusUpdater := networkstatus.NewNetworkStatusUpdater(cachedClient)
			handler := &ExternalWorkloadHandler{
				cachedClient: cachedClient,
				supportedGVKsHandlers: map[schema.GroupVersionKind]supportedhandlers.TopOwnerHandler{
					lwsGVK: networkUpdatingTopOwnerHandler{},
				},
				onlyOwnedSupportedGVKs: map[schema.GroupVersionKind]bool{},
				networkStatusUpdater:   networkStatusUpdater,
				structureUpdater:       structure.NewStructureUpdater(cachedClient),
			}

			Expect(handler.handleOwnedObject(ctx, pod, addedResource)).To(Succeed())
			Expect(networkStatusUpdater.TasksQueue().Len()).To(Equal(1))

			task, shutdown := networkStatusUpdater.TasksQueue().Get()
			Expect(shutdown).To(BeFalse())
			Expect(task.Object).To(BeAssignableToTypeOf(&v1alpha1.ExternalWorkload{}))
			networkStatusUpdater.TasksQueue().Done(task)
		},
		Entry("for an update event via the top-owner handler", false),
		Entry("for an add event via the existing generic enqueue", true),
	)
})
