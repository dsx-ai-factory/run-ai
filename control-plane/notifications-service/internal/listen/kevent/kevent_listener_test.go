/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kevent

import (
	"context"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	relevantNamespace       = "relevant-namespace"
	relevantEventObjectKind = "relevant-object-kind"
	relevantEventObjectName = "relevant-object-name"
	relevantEventObjectNs   = relevantNamespace
	relevantEventReason     = "relevant-event-reason"
	relevantEventNote       = "relevant-event-note"
)

// initKubeListenerTest sets up a fake kube client, config, listener and starts listening.
func initKubeListenerTest(eventCh chan *events.Event) (kubeClient kubernetes.Interface, kubeListener *KubeEventListener, cancel context.CancelFunc) {
	kubeClient = fake.NewSimpleClientset()
	config := &KubeEventListenerConfig{
		RelevantNamespaces: []string{relevantNamespace},
		RelevantObjects: []RelevantObjectConfig{
			{
				Kind:    relevantEventObjectKind,
				Reasons: []string{relevantEventReason},
			},
		},
	}

	kubeListener = NewKubeEventListener("k_listener", eventCh, kubeClient, config)
	ctx := context.Background()
	ctx, cancel = context.WithCancel(ctx)
	go kubeListener.Listen(ctx)
	return
}

func teardownKubeListenerTest(cancel context.CancelFunc) {
	if cancel != nil {
		cancel()
	}
}

func TestKubeListener_RelevantEvent(t *testing.T) {
	eventCh := make(chan *events.Event, 1)
	kubeClient, kubeListener, cancel := initKubeListenerTest(eventCh)
	defer teardownKubeListenerTest(cancel)

	relevantEvent := &eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Now(),
			Namespace:         relevantNamespace,
		},
		Regarding: corev1.ObjectReference{
			Kind:      relevantEventObjectKind,
			Name:      relevantEventObjectName,
			Namespace: relevantEventObjectNs,
		},
		Reason: relevantEventReason,
		Note:   relevantEventNote,
	}

	_, err := kubeClient.EventsV1().Events(relevantEvent.Namespace).Create(context.Background(), relevantEvent, metav1.CreateOptions{})
	require.NoError(t, err)

	select {
	case ev := <-eventCh:
		// Validate important properties rather than deep-equal on the whole struct
		if ev == nil {
			t.Fatalf("received nil event")
		}
		// source
		assert.Equal(t, "kubernetes", ev.Properties["source"])
		// regarding map
		regarding, ok := ev.Properties["regarding"].(map[string]interface{})
		require.True(t, ok, "regarding should be a map")
		assert.Equal(t, relevantEventObjectKind, regarding["kind"])
		assert.Equal(t, relevantEventObjectName, regarding["name"])
		assert.Equal(t, relevantEventObjectNs, regarding["namespace"])
		// reason and note
		assert.Equal(t, relevantEventReason, ev.Properties["reason"])
		assert.Equal(t, relevantEventNote, ev.Properties["note"])
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for event on channel")
	}

	_ = kubeListener // avoid unused
}

func TestKubeListener_IrrelevantEvent(t *testing.T) {
	eventCh := make(chan *events.Event, 1)
	kubeClient, _, cancel := initKubeListenerTest(eventCh)
	defer teardownKubeListenerTest(cancel)

	irrelevantEvents := []*eventsv1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "irrelevant-namespace",
				Name:      "irrelevant-namespace-event",
			},
			Regarding: corev1.ObjectReference{
				Kind: relevantEventObjectKind,
			},
			Reason: relevantEventReason,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: relevantNamespace,
				Name:      "irrelevant-object-kind-event",
			},
			Regarding: corev1.ObjectReference{
				Kind: "irrelevant-object-kind",
			},
			Reason: relevantEventReason,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: relevantNamespace,
				Name:      "irrelevant-event-reason-event",
			},
			Regarding: corev1.ObjectReference{
				Kind: relevantEventObjectKind,
			},
			Reason: "irrelevant-event-reason",
		},
	}

	for _, irrelevantEvent := range irrelevantEvents {
		_, err := kubeClient.EventsV1().Events(irrelevantEvent.Namespace).Create(context.Background(), irrelevantEvent, metav1.CreateOptions{})
		require.NoError(t, err)
	}

	// Ensure no events are received within a short window
	select {
	case ev := <-eventCh:
		t.Fatalf("expected no event, but received: %+v", ev)
	case <-time.After(500 * time.Millisecond):
		// success - no event
	}
}

func TestEventAgeValid_Valid(t *testing.T) {
	event := &eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Now(),
		},
	}
	valid := eventAgeValid(event)
	assert.True(t, valid, "valid should be true")
}

func TestEventAgeValid_NotValid(t *testing.T) {
	event := &eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-15 * time.Second)},
		},
	}
	valid := eventAgeValid(event)
	assert.False(t, valid, "valid should be false")
}
