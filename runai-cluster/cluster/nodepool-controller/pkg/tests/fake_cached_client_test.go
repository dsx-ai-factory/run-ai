package tests

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/cachedclient"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FakeNodeCachedClient struct {
	client.Client
	fakeNodeInformer *FakeCachedInformer
}

type FakeCachedInformer struct {
	watchClient client.WithWatch
	hasSynced   bool
	nodeHandler clientgocache.ResourceEventHandler

	nodeMap map[string]*corev1.Node
}

func NewFakeCachedInformer(watchClient client.WithWatch) *FakeCachedInformer {
	return &FakeCachedInformer{
		watchClient: watchClient,
		hasSynced:   false,
		nodeMap:     map[string]*corev1.Node{},
	}
}

func (c *FakeCachedInformer) AddEventHandler(handler clientgocache.ResourceEventHandler) (clientgocache.ResourceEventHandlerRegistration, error) {
	c.nodeHandler = handler
	return nil, nil
}

func (c *FakeCachedInformer) HasSynced() bool {
	return c.hasSynced
}

func (c *FakeCachedInformer) Start(ctx context.Context, stopCh <-chan struct{}) {
	w, err := c.watchClient.Watch(ctx, &corev1.NodeList{})

	if err != nil {
		zap.S().Fatalf("unable to watch changes for nodes, error:: %+v", err)
	}

	c.hasSynced = true

loop:
	for {
		select {
		case event, ok := <-w.ResultChan():
			object := event.Object
			if !ok {
				go c.Start(ctx, stopCh)
				zap.S().Infof("channel of the node controller has closed, starting a new one")
				return
			}
			node := object.(*corev1.Node)
			switch event.Type {
			case watch.Added:
				c.nodeHandler.OnAdd(object, false)
				c.nodeMap[node.Name] = node
			case watch.Deleted:
				c.nodeHandler.OnDelete(object)
			case watch.Modified:
				oldNode, found := c.nodeMap[node.Name]
				if !found {
					oldNode = node
				}
				c.nodeHandler.OnUpdate(oldNode, object)
				c.nodeMap[node.Name] = node
			default:
			}
		case <-stopCh:
			w.Stop()
			break loop
		}
	}
}

func NewFakeCachedClient(fakeClient client.WithWatch) *FakeNodeCachedClient {
	return &FakeNodeCachedClient{
		Client:           fakeClient,
		fakeNodeInformer: NewFakeCachedInformer(fakeClient),
	}
}

func (c *FakeNodeCachedClient) StartInformers(ctx context.Context, stopCh <-chan struct{}) {
	go c.fakeNodeInformer.Start(ctx, stopCh)

	if !clientgocache.WaitForNamedCacheSync("nodes", stopCh, c.fakeNodeInformer.HasSynced) {
		zap.S().Fatalf("failed waiting for nodes cache to sync")
	}
}

func (c *FakeNodeCachedClient) GetNodeInformer() cachedclient.CachedInformer {
	return c.fakeNodeInformer
}
