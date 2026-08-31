/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aggregator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestHandlePhaseAndPhaseMessageUsesSharedCalculator(t *testing.T) {
	readyPod := corev1.Pod{Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
		Conditions: []corev1.PodCondition{{
			Type: corev1.PodReady, Status: corev1.ConditionTrue,
		}},
	}}

	tests := []struct {
		name    string
		service *knv1.Service
		pods    []corev1.Pod
		phase   string
		message string
	}{
		{
			name:    "missing service preserves the legacy failure result",
			phase:   Failed,
			message: "Timed out waiting for the workload to become ready.",
		},
		{
			name: "ready service remains ready in the legacy API",
			service: &knv1.Service{Status: knv1.ServiceStatus{
				ConfigurationStatusFields: knv1.ConfigurationStatusFields{
					LatestCreatedRevisionName: "rev-1",
					LatestReadyRevisionName:   "rev-1",
				},
			}},
			pods:  []corev1.Pod{readyPod},
			phase: Ready,
		},
		{
			name: "failed service keeps the legacy failure message",
			service: &knv1.Service{Status: knv1.ServiceStatus{
				ConfigurationStatusFields: knv1.ConfigurationStatusFields{
					LatestCreatedRevisionName: "rev-1",
				},
				Status: duckv1.Status{Conditions: duckv1.Conditions{{
					Status: corev1.ConditionFalse, Message: "unable to fetch image",
				}}},
			}},
			phase:   Failed,
			message: "Failed to pull container image.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := newTestCalculator()
			sc.knativeService = tt.service
			sc.latestCreatedRevisionPods = tt.pods

			sc.handlePhaseAndPhaseMessage(context.Background())

			require.Equal(t, tt.phase, sc.workload.Status.InferenceStatus.Phase)
			require.Equal(t, tt.message, sc.workload.Status.InferenceStatus.PhaseMessage)
		})
	}
}

func TestHandlePhaseAndPhaseMessageUsesRetainedRevisionNames(t *testing.T) {
	sc := newTestCalculator()
	sc.knativeService = &knv1.Service{}
	sc.workload.Status.LatestCreatedRevisionName = ptrTo("rev-1")
	sc.workload.Status.LatestReadyRevisionName = ptrTo("rev-1")

	sc.handlePhaseAndPhaseMessage(context.Background())

	require.Equal(t, Ready, sc.workload.Status.InferenceStatus.Phase)
	require.Empty(t, sc.workload.Status.InferenceStatus.PhaseMessage)
}

func TestHandlePhaseAndPhaseMessagePreservesBuildErrorWithoutService(t *testing.T) {
	sc := newTestCalculator()
	sc.buildError = ptrTo("admission denied")

	sc.handlePhaseAndPhaseMessage(context.Background())

	require.Equal(t, Failed, sc.workload.Status.InferenceStatus.Phase)
	require.Equal(t, "Failed to create workload: admission denied", sc.workload.Status.InferenceStatus.PhaseMessage)
}
