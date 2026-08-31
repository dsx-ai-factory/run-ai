/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package fakecachedclient

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var podControllerKinds = map[string]bool{
	"Pod":         true,
	"Deployment":  true,
	"ReplicaSet":  true,
	"StatefulSet": true,
	"Job":         true,
	"CronJob":     true,
}

type FakeCachedInformer struct {
	watchClient client.WithWatch
	objectList  client.ObjectList
	handlers    []clientgocache.ResourceEventHandler
	hasSynced   bool
}

func NewFakeCachedInformer(watchClient client.WithWatch) *FakeCachedInformer {
	return &FakeCachedInformer{
		watchClient: watchClient,
		hasSynced:   false,
	}
}

func (c *FakeCachedInformer) AddEventHandler(handler clientgocache.ResourceEventHandler) (
	clientgocache.ResourceEventHandlerRegistration, error,
) {
	c.handlers = append(c.handlers, handler)
	return nil, nil
}

func (c *FakeCachedInformer) onAdd(object any) {
	for _, handler := range c.handlers {
		handler.OnAdd(object, true)
	}
}

func (c *FakeCachedInformer) onDelete(object any) {
	for _, handler := range c.handlers {
		handler.OnDelete(object)
	}
}

func (c *FakeCachedInformer) onUpdate(object any) {
	for _, handler := range c.handlers {
		handler.OnUpdate(object, object)
	}
}

func (c *FakeCachedInformer) HasSynced() bool {
	return c.hasSynced
}

func (c *FakeCachedInformer) SetObjectType(gvk schema.GroupVersionKind) {
	unstructuredObjectList := &unstructured.UnstructuredList{}
	unstructuredObjectList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind + "List",
	})
	c.objectList = unstructuredObjectList
}

func (c *FakeCachedInformer) Start(stopCh <-chan struct{}) {
	opts := &client.ListOptions{
		FieldSelector: fields.Set{constants.SchedulerNameIndexerName: constants.RunaiSchedulerName}.AsSelector(),
	}

	var w watch.Interface
	var err error

	if _, found := podControllerKinds[c.objectList.GetObjectKind().GroupVersionKind().Kind]; found {
		w, err = c.watchClient.Watch(context.Background(), c.objectList, opts)
	} else {
		w, err = c.watchClient.Watch(context.Background(), c.objectList)
	}

	if err != nil {
		log.Fatal().Msgf("unable to watch <%s> changes: %+v",
			c.objectList.GetObjectKind().GroupVersionKind().String(), err)
	}

	c.hasSynced = true
loop:
	for {
		select {
		case event, ok := <-w.ResultChan():
			object := event.Object
			if !ok {
				go c.Start(stopCh)
				log.Info().Msgf("channel of the %s controller has closed, starting a new one",
					c.objectList.GetObjectKind().GroupVersionKind().String())
				return
			}
			switch event.Type {
			case watch.Added:
				c.onAdd(object)
			case watch.Deleted:
				c.onDelete(object)
			case watch.Modified:
				c.onUpdate(object)
			default:
			}
		case <-stopCh:
			w.Stop()
			break loop
		}
	}
}
