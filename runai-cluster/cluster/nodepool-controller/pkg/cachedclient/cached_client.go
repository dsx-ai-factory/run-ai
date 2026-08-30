package cachedclient

import (
	"context"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CachedClient interface {
	// Client - Get and List would be cached, other operations would be done on the writer client
	client.Client

	StartInformers(ctx context.Context, stopCh <-chan struct{})
	GetNodeInformer() CachedInformer
}

type CachedInformer interface {
	AddEventHandler(handler clientgocache.ResourceEventHandler) (clientgocache.ResourceEventHandlerRegistration, error)
	HasSynced() bool
}

type NodePoolControllerCachedClient struct {
	client.Client
	cachedClient cache.Cache
	scheme       *apiruntime.Scheme
	nodeInformer CachedInformer
}

func NewCachedClient(ctx context.Context, writerClient client.Client,
	kubeConfig *rest.Config, scheme *apiruntime.Scheme) CachedClient {
	cachedClient := createCache(kubeConfig, scheme)

	node := &corev1.Node{}
	nodeInformer, err := cachedClient.GetInformer(ctx, node)
	if err != nil {
		log.Fatal().Msgf("Failed to create node informer, error: %s", err.Error())
	}

	return &NodePoolControllerCachedClient{
		Client:       writerClient,
		cachedClient: cachedClient,
		scheme:       scheme,
		nodeInformer: nodeInformer,
	}
}

// Get is a cached operation
func (c *NodePoolControllerCachedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object,
	opts ...client.GetOption) error {
	return c.cachedClient.Get(ctx, key, obj, opts...)
}

// List is a cached operation
func (c *NodePoolControllerCachedClient) List(ctx context.Context, list client.ObjectList,
	opts ...client.ListOption) error {
	return c.cachedClient.List(ctx, list, opts...)
}

func (c *NodePoolControllerCachedClient) StartInformers(ctx context.Context, stopCh <-chan struct{}) {
	go func() {
		if err := c.cachedClient.Start(ctx); err != nil {
			log.Fatal().Msgf("Failed to start cache client, error: %s", err.Error())
		}
	}()

	if !c.cachedClient.WaitForCacheSync(ctx) {
		log.Fatal().Msgf("Failed waiting for cache to sync")
	}

	if !clientgocache.WaitForNamedCacheSync("node", stopCh, c.nodeInformer.HasSynced) {
		log.Fatal().Msgf("failed waiting for node cache to sync")
	}
}

func (c *NodePoolControllerCachedClient) GetNodeInformer() CachedInformer {
	return c.nodeInformer
}

func createCache(kubeConfig *rest.Config, scheme *apiruntime.Scheme) cache.Cache {
	cacheOptions := cache.Options{
		Scheme: scheme,
	}
	cachedClient, err := cache.New(kubeConfig, cacheOptions)
	if err != nil {
		log.Fatal().Msgf("Failed to create cached client, error: %s", err.Error())
	}

	return cachedClient
}
