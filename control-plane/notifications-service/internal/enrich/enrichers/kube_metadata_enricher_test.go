/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package enrichers_test

import (
	"context"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/enrich/enrichers"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestKubeMetadataEnricher_Enrich_HappyFlow(t *testing.T) {
	kubeClientBuilder := fake.NewClientBuilder()

	kubeRClient := kubeClientBuilder.WithObjects(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: "default",
				Annotations: map[string]string{
					"user": "my-user",
				},
				Labels: map[string]string{
					"app": "my-app",
				},
			},
		},
	).Build()

	event := events.NewEvent()
	event.RawEvent = &eventsv1.Event{
		Regarding: corev1.ObjectReference{
			Kind:       "Pod",
			Namespace:  "default",
			Name:       "my-pod",
			APIVersion: "v1",
		},
		Reason: "Reason",
		Note:   "Note",
	}

	err := enrichers.NewKubeMetadataEnricher(kubeRClient,
		enrichers.KubeMetadataEnricherConfig{
			Annotations: map[string]string{"user": "user"},
			Labels:      map[string]string{"app": "app"},
		}).
		Enrich(context.Background(), event)
	require.NoError(t, err)

	assert.Equal(t, "my-user", event.Properties["user"])
	assert.Equal(t, "my-app", event.Properties["app"])
}

func TestKubeMetadataEnricher_NotKubeEvent(t *testing.T) {
	kubeClientBuilder := fake.NewClientBuilder()

	kubeRClient := kubeClientBuilder.Build()

	event := events.NewEvent()
	event.RawEvent = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "default",
			Annotations: map[string]string{
				"user": "my-user",
			},
			Labels: map[string]string{
				"app": "my-app",
			},
		},
	}

	err := enrichers.NewKubeMetadataEnricher(kubeRClient,
		enrichers.KubeMetadataEnricherConfig{
			Annotations: map[string]string{"user": "user"},
			Labels:      map[string]string{"app": "app"},
		}).
		Enrich(context.Background(), event)
	require.NoError(t, err)

	// properties should not be populated for non-kube raw events
	assert.Nil(t, event.Properties["user"])
	assert.Nil(t, event.Properties["app"])
}

func TestKubeMetadataEnricher_RegardingNotFound(t *testing.T) {
	kubeClientBuilder := fake.NewClientBuilder()

	kubeRClient := kubeClientBuilder.Build()

	event := events.NewEvent()
	event.RawEvent = &eventsv1.Event{
		Regarding: corev1.ObjectReference{
			Kind:       "Pod",
			Namespace:  "default",
			Name:       "my-pod",
			APIVersion: "v1",
		},
		Reason: "Reason",
		Note:   "Note",
	}

	err := enrichers.NewKubeMetadataEnricher(kubeRClient,
		enrichers.KubeMetadataEnricherConfig{Annotations: map[string]string{"user": "user"}}).
		Enrich(context.Background(), event)

	require.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
}
