/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cachedclient

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	v1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// kindsToDrop - if trying to get an object of this type, return a dedicated error that means we need to drop
// and not proceed with external workload handler (those objects don't need externalworkloads created for them).
var kindsToDrop = map[schema.GroupKind]bool{
	{Group: "run.ai", Kind: "TrainingWorkload"}:             true,
	{Group: "run.ai", Kind: "InteractiveWorkload"}:          true,
	{Group: "run.ai", Kind: "InferenceWorkload"}:            true,
	{Group: "run.ai", Kind: "DistributedInferenceWorkload"}: true,
	{Group: "run.ai", Kind: "DistributedWorkload"}:          true,
}

type CachedClient interface {
	// Client - Get and List would be cached, other operations would be done on the writer client
	client.Client

	SetInformers(ctx context.Context, supportedGVKs []schema.GroupVersionKind, additionalGVKs []schema.GroupVersionKind)
	StartInformers(ctx context.Context)
	WaitForCachesToSync(ctx context.Context, stopCh <-chan struct{})
	GetObjectInformers() map[schema.GroupVersionKind]CachedInformer
	SetCrdInformer(ctx context.Context, stopCh <-chan struct{}) CachedInformer
	GetUnexistingGvks() map[schema.GroupVersionKind]bool
	RemoveInformer(ctx context.Context, gvk schema.GroupVersionKind) error
	AddAdditionalGroupKind(gk schema.GroupKind)
	RemoveAdditionalGroupKind(gk schema.GroupKind)
	SetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind) (CachedInformer, error)
}

type CachedInformer interface {
	AddEventHandler(handler clientgocache.ResourceEventHandler) (clientgocache.ResourceEventHandlerRegistration, error)
	HasSynced() bool
}

type ExternalWorkloadsCachedClient struct {
	client.Client
	cachedClient cache.Cache
	scheme       *apiruntime.Scheme
	// object informers - shouldn't have duplicates of the same group/kind
	objectInformers     sync.Map // map[schema.GroupVersionKind]CachedInformer
	supportedGroupKinds sync.Map // map[schema.GroupKind]bool
	// kinds that don't have "objectInformers" set, but we do have permissions and caches for them
	additionalSupportedGroupKinds   sync.Map // map[schema.GroupVersionKind]*int
	additionalSupportedGroupKindsMu sync.Mutex

	// gvks that are supported, including duplicate versions of the same group/kind
	supportedGvks sync.Map // map[schema.GroupVersionKind]bool

	// gvks that should be supported, but do not exist in the k8s API server
	unexistingGvks sync.Map       // map[schema.GroupVersionKind]bool
	crdInformer    CachedInformer //nolint:unused // reserved for future use
}

func NewCachedClient(ctx context.Context, writerClient client.Client,
	kubeConfig *rest.Config, scheme *apiruntime.Scheme,
) CachedClient {
	cachedClient := createCache(ctx, kubeConfig, scheme)

	// kinds that don't have "objectInformers" set, but we do have permissions and caches for them.
	supportedGroupKinds := map[schema.GroupKind]bool{
		{Group: kaiv2alpha2.GroupVersion.Group, Kind: "PodGroup"}:     true,
		{Group: v1.GroupVersion.Group, Kind: "RunaiJob"}:              true,
		{Group: "run.ai", Kind: "TrainingWorkload"}:                   true,
		{Group: "run.ai", Kind: "InteractiveWorkload"}:                true,
		{Group: "run.ai", Kind: "InferenceWorkload"}:                  true,
		{Group: "run.ai", Kind: "DistributedInferenceWorkload"}:       true,
		{Group: "run.ai", Kind: "DistributedWorkload"}:                true,
		{Group: v1alpha1.GroupVersion.Group, Kind: "WorkloadRunner"}:  true,
		{Group: "", Kind: "Service"}:                                  true,
		{Group: "networking.k8s.io", Kind: "Ingress"}:                 true,
		{Group: "gateway.networking.k8s.io", Kind: "HTTPRoute"}:       true,
		{Group: "gateway.networking.k8s.io", Kind: "GRPCRoute"}:       true,
		{Group: "gateway.networking.k8s.io", Kind: "TLSRoute"}:        true,
		{Group: "gateway.networking.k8s.io", Kind: "Gateway"}:         true,
		{Group: "inference.networking.k8s.io", Kind: "InferencePool"}: true,
		{Group: "resource.k8s.io", Kind: "ResourceClaim"}:             true,
		{Group: "resource.k8s.io", Kind: "ResourceClaimTemplate"}:     true,
	}

	ewc := &ExternalWorkloadsCachedClient{
		Client:                          writerClient,
		cachedClient:                    cachedClient,
		scheme:                          scheme,
		objectInformers:                 sync.Map{},
		supportedGroupKinds:             sync.Map{},
		supportedGvks:                   sync.Map{},
		unexistingGvks:                  sync.Map{},
		additionalSupportedGroupKinds:   sync.Map{},
		additionalSupportedGroupKindsMu: sync.Mutex{},
	}

	for gk := range supportedGroupKinds {
		ewc.supportedGroupKinds.Store(gk, true)
	}
	return ewc
}

// Get is a cached operation by default.
// There's a bug in k8s cached client, where "Get" for objects we don't have the
// RBAC for are left waiting forever. So this is a wrapper to avoid those issues.
// https://github.com/kubernetes-sigs/controller-runtime/issues/550
func (c *ExternalWorkloadsCachedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object,
	opts ...client.GetOption,
) error {
	if key.Namespace == "" {
		return c.cachedClient.Get(ctx, key, obj, opts...)
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	groupKind := gvk.GroupKind()

	_, found := c.supportedGroupKinds.Load(groupKind)
	if found || groupKind.Kind == "" {
		// when we are getting an object explicitly, its gvk is not set,
		// and it's fine, we should have permissions for it.
		return c.cachedClient.Get(ctx, key, obj, opts...)
	}
	_, foundAsAdditional := c.additionalSupportedGroupKinds.Load(groupKind)
	if foundAsAdditional {
		return c.cachedClient.Get(ctx, key, obj, opts...)
	}

	// if it's an unrecognized type - used Uncached client to Get it to avoid the k8s bug
	return c.Client.Get(ctx, key, obj, opts...)
}

// List is a cached operation
func (c *ExternalWorkloadsCachedClient) List(ctx context.Context, list client.ObjectList,
	opts ...client.ListOption,
) error {
	return c.cachedClient.List(ctx, list, opts...)
}

func (c *ExternalWorkloadsCachedClient) SetInformers(ctx context.Context, supportedGVKs []schema.GroupVersionKind, additionalGVKs []schema.GroupVersionKind) {
	for _, supportedGVK := range append(supportedGVKs, additionalGVKs...) {
		c.setInformer(ctx, supportedGVK)
		c.supportedGvks.Store(supportedGVK, true)
	}
	RemoveDuplicateInformers(&c.objectInformers, &c.unexistingGvks)
}

func (c *ExternalWorkloadsCachedClient) setInformer(ctx context.Context,
	gvk schema.GroupVersionKind,
) {
	objectInformer, err := c.cachedClient.GetInformerForKind(ctx, gvk)
	if err != nil {
		if meta.IsNoMatchError(err) {
			log.Warn().Msgf("Type <%s> doesn't exist in the k8s api server, skipping informer creation...",
				gvk.String())
			c.unexistingGvks.Store(gvk, true)
			return
		}
		log.Error().Msgf("Failed to create informer for type <%s>, error: %s",
			gvk.String(), err.Error())
		return
	}

	c.objectInformers.Store(gvk, objectInformer)
	c.supportedGroupKinds.Store(gvk.GroupKind(), true)
}

func (c *ExternalWorkloadsCachedClient) RemoveInformer(ctx context.Context, gvk schema.GroupVersionKind) error {
	if _, ok := c.objectInformers.Load(gvk); ok {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gvk)
		err := c.cachedClient.RemoveInformer(ctx, obj)
		if err != nil {
			return err
		}
		c.objectInformers.Delete(gvk)
	}

	c.unexistingGvks.Delete(gvk)
	c.supportedGvks.Delete(gvk)

	gkToGvks := GetGvksMapPerGroupKindFromMap(&c.supportedGvks)
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
		c.setInformer(ctx, latestGvk)
		return nil
	}

	c.supportedGroupKinds.Delete(gvk.GroupKind())

	return nil
}

// GetUnexistingGvks - make sure to call only after a call to "SetInformers"
// returns the gvks that were requested to have informers, but do not exist in the k8s api server, hence unregistered
func (c *ExternalWorkloadsCachedClient) GetUnexistingGvks() map[schema.GroupVersionKind]bool {
	snapshot := map[schema.GroupVersionKind]bool{}
	c.unexistingGvks.Range(func(key, value interface{}) bool {
		snapshot[key.(schema.GroupVersionKind)] = value.(bool)
		return true
	})
	return snapshot
}

// SetCrdInformer - sometimes when the external-workload-integrator starts, some operators/crds
// are not yet installed (for example, RayCluster/Job/Service).
// And so during the startup, informers for those objects will not be created.
// Using the CustomResourceDefinition informer, we can identify
// the newly added crds during runtime and create informers for them.
func (c *ExternalWorkloadsCachedClient) SetCrdInformer(ctx context.Context, stopCh <-chan struct{}) CachedInformer {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	crdInformer, err := c.cachedClient.GetInformer(ctx, crd)
	if err != nil {
		log.Fatal().Msgf("Failed to create informer for type <%s>, error: %s",
			crd.GetObjectKind().GroupVersionKind().String(), err.Error())
	}

	if !clientgocache.WaitForNamedCacheSync("crds", stopCh, crdInformer.HasSynced) {
		log.Fatal().Msgf("failed waiting for crds cache to sync")
	}

	log.Info().Msgf("Created informer to watch for new crds in runtime")
	return crdInformer
}

// RemoveDuplicateInformers -
// objectInformers should be map[schema.GroupVersionKind]CachedInformer
// unexistingGvks should be map[schema.GroupVersionKind]bool
func RemoveDuplicateInformers(objectInformers *sync.Map, unexistingGvks *sync.Map) {
	// sometimes, two versions of the same group/kind exist in the cluster (e.g. v1 and v1alpha1);
	// in that case we could have duplicate informers for the same objects.
	// we need to remove the duplicate informers, otherwise we'll get duplicate events.
	// we'll keep the informer with the latest version.

	gkToGvks := GetGvksMapPerGroupKindFromMap(objectInformers)
	for gk, gvks := range gkToGvks {
		if len(gvks) <= 1 {
			continue
		}

		log.Info().Msgf("Found multiple versions for group/kind <%s>, versions: %v", gk.String(), gvks)
		latestVersion := ""
		for _, gvk := range gvks {
			if gvk.Version > latestVersion {
				latestVersion = gvk.Version
			}
		}

		// remove all other versions except the latest
		// from objectInformers and also from c.unexistingGvks - we don't want to track them anymore
		for _, gvk := range gvks {
			if gvk.Version != latestVersion {
				log.Info().Msgf("Deleting duplicate informer for gvk <%s>; will use latest version <%s>",
					gvk.String(), latestVersion)
				objectInformers.Delete(gvk)

				if unexistingGvks != nil {
					unexistingGvks.Delete(gvk)
				}
			}
		}
	}
}

func (c *ExternalWorkloadsCachedClient) StartInformers(ctx context.Context) {
	go func() {
		if err := c.cachedClient.Start(ctx); err != nil {
			log.Fatal().Msgf("Failed to start cached client, error: %s", err.Error())
		}
	}()
}

func (c *ExternalWorkloadsCachedClient) WaitForCachesToSync(ctx context.Context, stopCh <-chan struct{}) {
	if !c.cachedClient.WaitForCacheSync(ctx) {
		log.Fatal().Msgf("failed waiting for cache to sync")
	}

	objectInformers := c.GetObjectInformers()
	for gvk := range objectInformers {
		objectInformer := objectInformers[gvk]
		if !clientgocache.WaitForNamedCacheSync(gvk.String(), stopCh, objectInformer.HasSynced) {
			log.Fatal().Msgf("failed waiting for %s cache to sync", gvk.String())
		}
	}
}

func (c *ExternalWorkloadsCachedClient) GetObjectInformers() map[schema.GroupVersionKind]CachedInformer {
	snapshot := map[schema.GroupVersionKind]CachedInformer{}
	c.objectInformers.Range(func(key, value interface{}) bool {
		snapshot[key.(schema.GroupVersionKind)] = value.(CachedInformer)
		return true
	})
	return snapshot
}

func (c *ExternalWorkloadsCachedClient) AddAdditionalGroupKind(gk schema.GroupKind) {
	c.additionalSupportedGroupKindsMu.Lock()
	defer c.additionalSupportedGroupKindsMu.Unlock()

	if count, ok := c.additionalSupportedGroupKinds.Load(gk); ok {
		c.additionalSupportedGroupKinds.Store(gk, count.(int)+1)
	} else {
		c.additionalSupportedGroupKinds.Store(gk, 1)
	}
}

// SetInformerForKind returns a cached informer for the given GVK.
// It registers the GroupKind in supportedGroupKinds (enabling cached Get/List)
// but does NOT add to objectInformers (so workload event handlers are not registered on it).
func (c *ExternalWorkloadsCachedClient) SetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind) (CachedInformer, error) {
	informer, err := c.cachedClient.GetInformerForKind(ctx, gvk)
	if err != nil {
		return nil, err
	}
	c.supportedGroupKinds.Store(gvk.GroupKind(), true)
	return informer, nil
}

func (c *ExternalWorkloadsCachedClient) RemoveAdditionalGroupKind(gk schema.GroupKind) {
	c.additionalSupportedGroupKindsMu.Lock()
	defer c.additionalSupportedGroupKindsMu.Unlock()

	if count, ok := c.additionalSupportedGroupKinds.Load(gk); ok {
		if count.(int) <= 1 {
			c.additionalSupportedGroupKinds.Delete(gk)
		} else {
			c.additionalSupportedGroupKinds.Store(gk, count.(int)-1)
		}
	}
}

func createCache(ctx context.Context, kubeConfig *rest.Config, scheme *apiruntime.Scheme) cache.Cache {
	schedulerSelector := fields.Set{constants.SchedulerNameIndexerName: constants.RunaiSchedulerName}.AsSelector()
	cacheOptions := cache.Options{
		Scheme: scheme,
		ByObject: map[client.Object]cache.ByObject{
			&corev1.Pod{}: {Field: schedulerSelector},
		},
	}

	// can't add other types to the ByObject field selector - the schedulerName field is not supported

	cachedClient, err := cache.New(kubeConfig, cacheOptions)
	if err != nil {
		log.Fatal().Msgf("Failed to create cached client, error: %s", err.Error())
	}

	addIndexers(ctx, cachedClient)

	return cachedClient
}

func addIndexers(ctx context.Context, cachedClient cache.Cache) {
	err := cachedClient.IndexField(ctx, &v1alpha1.ExternalWorkload{},
		constants.ExternalWorkloadByOriginalUIDIndexerName, ExternalWorkloadByOriginalWorkloadUIDIndexer)
	if err != nil {
		log.Fatal().Msgf("Failed indexing externalWorkload field: %s, err: %s",
			constants.ExternalWorkloadByOriginalUIDIndexerName, err.Error())
	}
}

func ExternalWorkloadByOriginalWorkloadUIDIndexer(obj client.Object) []string {
	externalWorkload, ok := obj.(*v1alpha1.ExternalWorkload)
	if !ok {
		log.Error().Msgf("Failed converting object to externalWorkload in ExternalWorkloadOfObjectIndexer: %v",
			obj)
		return []string{""}
	}

	return []string{string(externalWorkload.Spec.OriginalWorkloadUID)}
}

func ShouldDropWorkload(topOwner client.Object) bool {
	gvk := topOwner.GetObjectKind().GroupVersionKind()
	_, found := kindsToDrop[gvk.GroupKind()]
	return found
}

func GetGvksMapPerGroupKind[T any](gvks map[schema.GroupVersionKind]T) map[schema.GroupKind][]schema.GroupVersionKind {
	gkToGvks := map[schema.GroupKind][]schema.GroupVersionKind{}
	for gvk := range gvks {
		gk := gvk.GroupKind()

		if _, ok := gkToGvks[gk]; !ok {
			gkToGvks[gk] = []schema.GroupVersionKind{}
		}

		gkToGvks[gk] = append(gkToGvks[gk], gvk)
	}
	return gkToGvks
}

// GetGvksMapPerGroupKindFromMap - gvks should be a map[schema.GroupVersionKind]T
func GetGvksMapPerGroupKindFromMap(gvks *sync.Map) map[schema.GroupKind][]schema.GroupVersionKind {
	if gvks == nil {
		return map[schema.GroupKind][]schema.GroupVersionKind{}
	}

	// implement using GetGvksMapPerGroupKind
	gvksMap := map[schema.GroupVersionKind]bool{}
	gvks.Range(func(key, _ interface{}) bool {
		gvksMap[key.(schema.GroupVersionKind)] = true
		return true
	})
	return GetGvksMapPerGroupKind(gvksMap)
}
