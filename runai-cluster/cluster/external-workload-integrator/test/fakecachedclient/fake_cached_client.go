/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package fakecachedclient

import (
	"context"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FakeCachedClient struct {
	client.WithWatch
	fakeClientWithFieldSelector *FakeClientWithFieldSelector

	// objectInformers - values hold the actual object of FakeCachedInformer
	objectInformers map[schema.GroupVersionKind]*FakeCachedInformer
	// objectInformersInterfaces - same values but their type is the interface CachedInformer
	objectInformersInterfaces map[schema.GroupVersionKind]cachedclient.CachedInformer
	stopCh                    chan struct{}
	withRemoveDuplicates      bool

	supportedGvks                 map[schema.GroupVersionKind]bool
	additionalSupportedGroupKinds sync.Map // map[schema.GroupVersionKind]bool
}

func NewFakeCachedClient(watchClient client.WithWatch, stopCh chan struct{}, withRemoveDuplicates bool) *FakeCachedClient {
	fakeClientWithFieldSelector := &FakeClientWithFieldSelector{WithWatch: watchClient}

	return &FakeCachedClient{
		WithWatch:                     fakeClientWithFieldSelector,
		fakeClientWithFieldSelector:   fakeClientWithFieldSelector,
		objectInformers:               map[schema.GroupVersionKind]*FakeCachedInformer{},
		objectInformersInterfaces:     map[schema.GroupVersionKind]cachedclient.CachedInformer{},
		stopCh:                        stopCh,
		withRemoveDuplicates:          withRemoveDuplicates,
		supportedGvks:                 map[schema.GroupVersionKind]bool{},
		additionalSupportedGroupKinds: sync.Map{},
	}
}

func (c *FakeCachedClient) SetInformers(_ context.Context, supportedGVKs []schema.GroupVersionKind, additionalGVKs []schema.GroupVersionKind) {
	objectInformers := map[schema.GroupVersionKind]*FakeCachedInformer{}
	objectInformersI := map[schema.GroupVersionKind]cachedclient.CachedInformer{}
	for _, supportedGVK := range append(supportedGVKs, additionalGVKs...) {
		informer := NewFakeCachedInformer(c.fakeClientWithFieldSelector)
		informer.SetObjectType(supportedGVK)
		objectInformers[supportedGVK] = informer
		objectInformersI[supportedGVK] = informer
		c.supportedGvks[supportedGVK] = true
	}

	if c.withRemoveDuplicates {
		objectInformersIMap := &sync.Map{}
		for gvk, informer := range objectInformersI {
			objectInformersIMap.Store(gvk, informer)
		}
		cachedclient.RemoveDuplicateInformers(objectInformersIMap, nil)
		objectInformersI = map[schema.GroupVersionKind]cachedclient.CachedInformer{}
		objectInformersIMap.Range(func(key, value interface{}) bool {
			objectInformersI[key.(schema.GroupVersionKind)] = value.(cachedclient.CachedInformer)
			return true
		})

		// remove duplicates from objectInformers as well
		for gvk := range objectInformers {
			if _, found := objectInformersI[gvk]; !found {
				delete(objectInformers, gvk)
			}
		}
	}

	c.objectInformers = objectInformers
	c.objectInformersInterfaces = objectInformersI

	c.fakeClientWithFieldSelector.SetInformers(objectInformers)
}

func (c *FakeCachedClient) RemoveInformer(ctx context.Context, gvk schema.GroupVersionKind) error {
	delete(c.objectInformers, gvk)
	delete(c.objectInformersInterfaces, gvk)
	delete(c.supportedGvks, gvk)

	gkToGvks := cachedclient.GetGvksMapPerGroupKind(c.supportedGvks)
	if gvks, ok := gkToGvks[gvk.GroupKind()]; ok {
		// if there are still other versions of this group/kind after we removed the current gvk informer,
		// then: 1. don't delete it from supportedGroupKinds
		//       2. need to start an informer for one of the remaining versions (pick the latest)
		latestGvk := schema.GroupVersionKind{}
		for _, gvkOfGv := range gvks {
			if gvkOfGv.Version > latestGvk.Version {
				latestGvk = gvkOfGv
			}
		}
		informer := NewFakeCachedInformer(c.fakeClientWithFieldSelector)
		informer.SetObjectType(latestGvk)
		c.objectInformers[latestGvk] = informer
		c.objectInformersInterfaces[latestGvk] = informer
	}
	c.fakeClientWithFieldSelector.SetInformers(c.objectInformers)

	return nil
}

func (c *FakeCachedClient) GetUnexistingGvks() map[schema.GroupVersionKind]bool {
	return map[schema.GroupVersionKind]bool{}
}

func (c *FakeCachedClient) SetCrdInformer(ctx context.Context, stopCh <-chan struct{}) cachedclient.CachedInformer {
	informer := NewFakeCachedInformer(c.fakeClientWithFieldSelector)
	informer.SetObjectType(schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Version: "v1",
		Kind:    "CustomResourceDefinition",
	})
	return informer
}

func (c *FakeCachedClient) StartInformers(_ context.Context) {
	for _, objectInformer := range c.objectInformers {
		go objectInformer.Start(c.stopCh)
	}
}

func (c *FakeCachedClient) WaitForCachesToSync(_ context.Context, stopCh <-chan struct{}) {
	for gvk := range c.objectInformers {
		objectInformer := c.objectInformers[gvk]
		if !clientgocache.WaitForNamedCacheSync(gvk.String(), stopCh, objectInformer.HasSynced) {
			log.Fatal().Msgf("failed waiting for %s cache to sync", gvk.String())
		}
	}
}

func (c *FakeCachedClient) GetObjectInformers() map[schema.GroupVersionKind]cachedclient.CachedInformer {
	return c.objectInformersInterfaces
}

func (c *FakeCachedClient) AddAdditionalGroupKind(gk schema.GroupKind) {
	if count, ok := c.additionalSupportedGroupKinds.Load(gk); ok {
		c.additionalSupportedGroupKinds.Store(gk, count.(int)+1)
	} else {
		c.additionalSupportedGroupKinds.Store(gk, 1)
	}
}

func (c *FakeCachedClient) GetAdditionalGroupKinds() map[schema.GroupKind]int {
	result := map[schema.GroupKind]int{}
	c.additionalSupportedGroupKinds.Range(func(key, value interface{}) bool {
		result[key.(schema.GroupKind)] = value.(int)
		return true
	})
	return result
}

func (c *FakeCachedClient) SetInformerForKind(_ context.Context, gvk schema.GroupVersionKind) (cachedclient.CachedInformer, error) {
	informer := NewFakeCachedInformer(c.fakeClientWithFieldSelector)
	informer.SetObjectType(gvk)
	c.objectInformers[gvk] = informer
	c.objectInformersInterfaces[gvk] = informer
	c.fakeClientWithFieldSelector.SetInformers(c.objectInformers)
	return informer, nil
}

func (c *FakeCachedClient) RemoveAdditionalGroupKind(gk schema.GroupKind) {
	if count, ok := c.additionalSupportedGroupKinds.Load(gk); ok {
		if count.(int) <= 1 {
			c.additionalSupportedGroupKinds.Delete(gk)
		} else {
			c.additionalSupportedGroupKinds.Store(gk, count.(int)-1)
		}
	}
}

type FakeClientWithFieldSelector struct {
	client.WithWatch
	objectInformers map[schema.GroupVersionKind]*FakeCachedInformer
}

func (f *FakeClientWithFieldSelector) SetInformers(objectInformers map[schema.GroupVersionKind]*FakeCachedInformer) {
	f.objectInformers = objectInformers
}

// Get - if needed, fake a "Forbidden" error
func (f *FakeClientWithFieldSelector) Get(ctx context.Context, key client.ObjectKey,
	obj client.Object, opts ...client.GetOption,
) error {
	if key.Namespace == "" {
		return f.WithWatch.Get(ctx, key, obj, opts...)
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	_, found := f.objectInformers[gvk]
	if found || gvk.Kind == "PodGroup" || gvk.Kind == "" || gvk.Kind == "RunaiJob" || gvk.Kind == "TrainingWorkload" ||
		gvk.Kind == "InteractiveWorkload" || gvk.Kind == "DistributedWorkload" || gvk.Kind == "DaemonSet" {
		return f.WithWatch.Get(ctx, key, obj, opts...)
	}
	return errors.NewForbidden(schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind}, key.Name,
		errors.NewUnauthorized(""))
}

// List - if needed, list with field selector "manually"
func (f *FakeClientWithFieldSelector) List(
	ctx context.Context, list client.ObjectList, opts ...client.ListOption,
) error {
	listOpts := client.ListOptions{}
	listOpts.ApplyOptions(opts)
	if listOpts.FieldSelector == nil || listOpts.FieldSelector.Empty() ||
		!strings.Contains(listOpts.FieldSelector.String(), constants.ExternalWorkloadByOriginalUIDIndexerName) {
		return f.WithWatch.List(ctx, list, opts...)
	}
	fieldSelector := listOpts.FieldSelector.Requirements()[0]

	allExternalWorkloadsList := &v1alpha1.ExternalWorkloadList{}
	listOpts.FieldSelector = nil
	err := f.WithWatch.List(ctx, allExternalWorkloadsList, &listOpts)
	if err != nil {
		return err
	}

	result := []runtime.Object{}
	for index := range allExternalWorkloadsList.Items {
		externalWorkload := &allExternalWorkloadsList.Items[index]
		resultKeys := cachedclient.ExternalWorkloadByOriginalWorkloadUIDIndexer(externalWorkload)
		if len(resultKeys) == 0 {
			continue
		}

		for _, key := range resultKeys {
			if fieldSelector.Value == key {
				result = append(result, externalWorkload)
			}
		}
	}
	return apimeta.SetList(list, result)
}
