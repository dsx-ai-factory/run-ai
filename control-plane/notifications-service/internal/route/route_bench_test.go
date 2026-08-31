/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package route_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/enrich"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/filter"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen/kevent"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/notify"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/notify/mocks"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/route"
	msinks "github.com/run-ai/runai/control-plane/notifications-service/internal/sinks/mocks"
	"github.com/stretchr/testify/mock"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/client-go/kubernetes"
	kubefake "k8s.io/client-go/kubernetes/fake"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const conf_path = "./config.yaml"

func BenchmarkFullFlow_SinkIdle100Batch1(b *testing.B) {
	configProvider := basic.NewConfigurationProvider(conf_path)
	IntegrationFullFlow_WithSinkMockIdle(b, configProvider, 100*time.Millisecond, 1)
}

func BenchmarkFullFlow_SinkIdle100Batch10(b *testing.B) {
	configProvider := basic.NewConfigurationProvider(conf_path)
	IntegrationFullFlow_WithSinkMockIdle(b, configProvider, 100*time.Millisecond, 10)
}

func BenchmarkFullFlow_SinkIdle100Batch100(b *testing.B) {
	configProvider := basic.NewConfigurationProvider(conf_path)
	IntegrationFullFlow_WithSinkMockIdle(b, configProvider, 100*time.Millisecond, 100)
}

func BenchmarkFullFlow_SinkIdle0Batch1(b *testing.B) {
	configProvider := basic.NewConfigurationProvider(conf_path)
	IntegrationFullFlow_WithSinkMockIdle(b, configProvider, 0*time.Millisecond, 1)
}

func BenchmarkFullFlow_SinkIdle0Batch10(b *testing.B) {
	configProvider := basic.NewConfigurationProvider(conf_path)
	IntegrationFullFlow_WithSinkMockIdle(b, configProvider, 0*time.Millisecond, 10)
}

func BenchmarkFullFlow_SinkIdle0Batch100(b *testing.B) {
	configProvider := basic.NewConfigurationProvider(conf_path)
	IntegrationFullFlow_WithSinkMockIdle(b, configProvider, 0*time.Millisecond, 100)
}

func IntegrationFullFlow_WithSinkMockIdle(b *testing.B, configProvider route.RouterConfigProvider, sinkIdle time.Duration, eventsChunkSize int) {
	kubeClient := newTestKubeClient(getKubeClientsObjects(b, eventsChunkSize))

	wg := &sync.WaitGroup{}
	wg.Add(b.N * eventsChunkSize)

	sinkFactoryMock := getSinkFactoryMock(b, wg, sinkIdle, eventsChunkSize)

	eventCh := make(chan *events.Event)
	handler := createTestHandler(configProvider.GetRouteProviders("", &events.Event{})[0], sinkFactoryMock)
	listener := kevent.NewKubeEventListener("k_listener", eventCh, kubeClient, &kevent.KubeEventListenerConfig{
		RelevantObjects:    []kevent.RelevantObjectConfig{},
		RelevantNamespaces: []string{},
	})
	router := route.NewNotificationsRouter(listener, eventCh, configProvider)
	router.RouteBuilder = func(rcp route.RouteConfigProvider, sinkProvider route.SinkConfigProvider) route.Route {
		return &route.HandlersRoute{
			ConfigProvider: configProvider.GetRouteProviders("", events.NewEvent())[0],
			Handler:        handler,
		}
	}

	// reset benchmark timer to ignore initialization process
	b.ResetTimer()
	router.Start()

	wg.Wait()
}

func getSinkFactoryMock(b *testing.B, wg *sync.WaitGroup, sinkIdle time.Duration, chunkSize int) *mocks.Factory {
	sinkFactoryMock := mocks.NewFactory(b)
	sinkerMock := msinks.NewSinker(b)
	sinkFactoryMock.On("CreateSink", mock.Anything, mock.Anything).Return(sinkerMock, nil)

	sinkerMock.On("Sink", mock.Anything).Times(b.N * chunkSize).Return(func(*events.Event) error {
		time.Sleep(sinkIdle)
		wg.Done()
		return nil
	})
	return sinkFactoryMock
}

func getKubeClientsObjects(b *testing.B, chunkSize int) []ctrlclient.Object {
	objs := []ctrlclient.Object{
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: "runai-test",
				Annotations: map[string]string{
					"user": "user",
				},
				CreationTimestamp: metav1.Now(),
			},
		},
	}

	// this is the actual number the handle function will be called so it's the benchmarking factor
	for i := 0; i < b.N; i++ {
		objs = addEventsChunk(chunkSize, i, objs)
	}
	return objs
}

func addEventsChunk(chunkSize, benchmarkFactor int, objs []ctrlclient.Object) []ctrlclient.Object {
	for i := 0; i < chunkSize; i++ {
		s := fmt.Sprintf("my-event-%v-%v", benchmarkFactor, i)
		e := &eventsv1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      s,
				Namespace: "relevant-namespace",
			},
			Regarding: corev1.ObjectReference{
				Kind:       "Pod",
				Namespace:  "runai-test",
				Name:       "my-pod",
				APIVersion: "v1",
			},
			Reason: "Scheduled",
			Note:   "Note",
		}
		objs = append(objs, e)
	}
	return objs
}

func createTestHandler(configProvider route.RouteConfigProvider, sf notify.Factory) events.EventHandler {
	notifyHandler := &notify.NotifyingEventHandler{
		ConfigProvider: configProvider,
		Notifier:       notify.NewEventNotifier(),
		SinkFactory:    sf,
	}

	filterHandler := filter.NewFilteringEventHandler(configProvider, notifyHandler)

	enricherFactory := enrich.NewEnricherFactory()
	return enrich.NewEnrichingEventHandler(configProvider, enricherFactory, filterHandler)
}

func newTestKubeClient(objs []ctrlclient.Object) kubernetes.Interface {
	runtimeObjects := make([]runtime.Object, len(objs))
	for i, obj := range objs {
		runtimeObjects[i] = obj.(runtime.Object)
	}
	return kubefake.NewSimpleClientset(runtimeObjects...)
}
