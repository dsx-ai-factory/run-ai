/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package fakecachedclient

import (
	"context"

	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FakeCachedInformer struct {
	watchClient client.WithWatch
	objectList  client.ObjectList
	handlers    []clientgocache.ResourceEventHandler
	hasSynced   bool
}

func NewFakeCachedInformer(watchClient client.WithWatch) *FakeCachedInformer {
	return &FakeCachedInformer{
		watchClient: watchClient,
		handlers:    make([]clientgocache.ResourceEventHandler, 0),
		hasSynced:   false,
	}
}

func (c *FakeCachedInformer) AddEventHandler(handler clientgocache.ResourceEventHandler) (clientgocache.ResourceEventHandlerRegistration, error) {
	c.handlers = append(c.handlers, handler)
	return nil, nil
}

func (c *FakeCachedInformer) HasSynced() bool {
	return c.hasSynced
}

func (c *FakeCachedInformer) SetObjectType(objectType cachedclient.SupportedType) {
	unstructuredObjectList := &unstructured.UnstructuredList{}
	unstructuredObjectList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   objectType.Group,
		Version: objectType.Version,
		Kind:    objectType.Kind + "List",
	})
	c.objectList = unstructuredObjectList
}

func (c *FakeCachedInformer) Start(ctx context.Context, stopCh <-chan struct{}) {
	w, err := c.watchClient.Watch(ctx, c.objectList)

	if err != nil {
		zap.S().Fatalf("unable to watch changes for <%s>, error:: %+v",
			c.objectList.GetObjectKind().GroupVersionKind().String(), err)
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
			switch event.Type {
			case watch.Added:
				for _, handler := range c.handlers {
					handler.OnAdd(object, false)
				}
			case watch.Deleted:
				for _, handler := range c.handlers {
					handler.OnDelete(object)
				}
			case watch.Modified:
				for _, handler := range c.handlers {
					handler.OnUpdate(object, object)
				}
			default:
			}
		case <-stopCh:
			w.Stop()
			break loop
		}
	}
}
