/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package fakecachedclient

import (
	"context"
	v1alpha2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	diwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type FakeCachedClient struct {
	client.Client
	objectInformers        map[cachedclient.SupportedType]*FakeCachedInformer
	rateLimitedControllers map[string]*cachedclient.RateLimitedController
}

func NewFakeCachedClient(fakeClientBuilder *fake.ClientBuilder) cachedclient.CachedClient {
	fakeClientBuilder = fakeClientBuilder.
		WithStatusSubresource(&twv2.TrainingWorkload{}, &diwv2.DistributedWorkload{}, &iwv2.InteractiveWorkload{}, &infwv2.InferenceWorkload{}, &v1alpha2.DistributedInferenceWorkload{}, &v1alpha1.ExternalWorkload{}).
		WithIndex(&corev1.Pod{}, constants.PodByPodGroupIndexerName, cachedclient.PodByPodGroupIndexer).
		WithIndex(&corev1.Pod{}, constants.NonRunaiPodByNodeNameIndexer, cachedclient.NonRunaiPodByNodeNameIndexer).
		WithIndex(&corev1.Pod{}, constants.PodBySchedulerNameIndexer, cachedclient.PodBySchedulerNameIndexer).
		WithIndex(&twv2.TrainingWorkload{}, constants.RunaiWorkloadByUsageIndexer, cachedclient.RunaiWorkloadByUsageIndexer).
		WithIndex(&iwv2.InteractiveWorkload{}, constants.RunaiWorkloadByUsageIndexer, cachedclient.RunaiWorkloadByUsageIndexer).
		WithIndex(&infwv2.InferenceWorkload{}, constants.RunaiWorkloadByUsageIndexer, cachedclient.RunaiWorkloadByUsageIndexer).
		WithIndex(&v1alpha2.DistributedInferenceWorkload{}, constants.RunaiWorkloadByUsageIndexer, cachedclient.RunaiWorkloadByUsageIndexer).
		WithIndex(&diwv2.DistributedWorkload{}, constants.RunaiWorkloadByUsageIndexer, cachedclient.RunaiWorkloadByUsageIndexer)
	watchClient := fakeClientBuilder.Build()

	objectInformers := map[cachedclient.SupportedType]*FakeCachedInformer{}
	for _, supportedType := range cachedclient.AllSupportedTypes {
		informer := NewFakeCachedInformer(watchClient)
		informer.SetObjectType(supportedType)

		objectInformers[supportedType] = informer
	}

	return &FakeCachedClient{
		Client:                 watchClient,
		objectInformers:        objectInformers,
		rateLimitedControllers: map[string]*cachedclient.RateLimitedController{},
	}
}

func (c *FakeCachedClient) Start(ctx context.Context, stopCh <-chan struct{}) {
	for _, objectInformer := range c.objectInformers {
		go objectInformer.Start(ctx, stopCh)
	}

	for gvk := range c.objectInformers {
		objectInformer := c.objectInformers[gvk]
		if !clientgocache.WaitForNamedCacheSync(gvk.String(), stopCh, objectInformer.HasSynced) {
			zap.S().Fatalf("failed waiting for %s cache to sync", gvk.String())
		}
	}

	for _, rateLimitedController := range c.rateLimitedControllers {
		rateLimitedController.Run()
	}
}

func (c *FakeCachedClient) GetInformer(objectType cachedclient.SupportedType) cachedclient.CachedInformer {
	objectInformer, found := c.objectInformers[objectType]
	if found && objectInformer != nil {
		return objectInformer
	}
	return nil
}

func (c *FakeCachedClient) NewRateLimitedController(resourceName cachedclient.RateLimiterName, handler cachedclient.Handler) *cachedclient.RateLimitedController {
	rateLimitedController := cachedclient.NewRateLimitedController(resourceName, handler)
	c.rateLimitedControllers[string(resourceName)] = rateLimitedController
	return rateLimitedController
}

func (c *FakeCachedClient) Shutdown() {
	for _, rateLimitedController := range c.rateLimitedControllers {
		rateLimitedController.ShutDown()
	}
}
