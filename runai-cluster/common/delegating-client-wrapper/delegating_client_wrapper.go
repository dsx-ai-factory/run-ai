/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package delegating_client_wrapper

import (
	"context"

	supportedworkloads "github.com/run-ai/runai/runai-cluster/common/supported-workloads"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DelegatingClientWrapper struct {
	client.Client
	uncachedClient      client.Reader
	supportedGroupKinds map[schema.GroupKind]bool
}

func NewDelegatingClientWrapper(cache cache.Cache, config *rest.Config, options client.Options,
	uncachedObjects ...client.Object) (client.Client, error) {
	uncachedClient, err := client.New(config, options)
	if err != nil {
		return nil, err
	}

	delegatingClient, err := client.NewDelegatingClient(client.NewDelegatingClientInput{
		CacheReader:     cache,
		Client:          uncachedClient,
		UncachedObjects: uncachedObjects,
	})

	delegatingClientWrapper := &DelegatingClientWrapper{
		Client:              delegatingClient,
		uncachedClient:      uncachedClient,
		supportedGroupKinds: getInitialSupportedGroupKinds(),
	}

	return delegatingClientWrapper, nil
}

// Get is a cached operation by default.
// There's a bug in k8s cached client, where "Get" for objects we don't have the
// RBAC for are left waiting forever. So this is a wrapper to avoid those issues.
// https://github.com/kubernetes-sigs/controller-runtime/issues/550
func (c *DelegatingClientWrapper) Get(ctx context.Context, key client.ObjectKey, obj client.Object,
	opts ...client.GetOption) error {
	if key.Namespace == "" {
		return c.Client.Get(ctx, key, obj, opts...)
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	groupKind := gvk.GroupKind()

	_, found := c.supportedGroupKinds[groupKind]
	if found || groupKind.Kind == "" {
		// when we are getting an object explicitly, its gvk is not set,
		// and it's fine, we should have permissions for it.
		return c.Client.Get(ctx, key, obj, opts...)
	}

	// if it's an unrecognized type - use Uncached client to Get it to avoid the k8s bug
	return c.uncachedClient.Get(ctx, key, obj, opts...)
}

func getInitialSupportedGroupKinds() map[schema.GroupKind]bool {
	allTypesThatNeedSupport := supportedworkloads.GetAllTypesThatNeedSupport()

	// kinds that are not part of "GetAllTypesThatNeedSupport", but we do have permissions and caches for them.
	supportedGroupKinds := map[schema.GroupKind]bool{
		{Group: "scheduling.run.ai", Kind: "PodGroup"}: true,
		{Group: "run.ai", Kind: "RunaiJob"}:            true,
		{Group: "run.ai", Kind: "TrainingWorkload"}:    true,
		{Group: "run.ai", Kind: "InteractiveWorkload"}: true,
		{Group: "run.ai", Kind: "DistributedWorkload"}: true,
		//TODO: inference?
	}

	for gvk := range allTypesThatNeedSupport {
		supportedGroupKinds[schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}] = true
	}

	return supportedGroupKinds
}
