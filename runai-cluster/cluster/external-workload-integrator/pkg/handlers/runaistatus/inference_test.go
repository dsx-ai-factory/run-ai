/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaistatus

import (
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("ComputeInferenceStatus", func() {
	readyPod := corev1.Pod{Status: corev1.PodStatus{
		Phase: corev1.PodRunning,
		Conditions: []corev1.PodCondition{{
			Type: corev1.PodReady, Status: corev1.ConditionTrue,
		}},
	}}

	Describe("legacy phase parity", func() {
		DescribeTable("calculates the expected phase and message",
			func(input InferenceInput, phase v1alpha1.ExternalWorkloadPhase, message string) {
				actualPhase, actualMessage := ComputeInferenceStatus(input, nil)

				Expect(actualPhase).To(Equal(phase))
				Expect(actualMessage).To(Equal(message))
			},
			Entry("service not reported yet is creating",
				InferenceInput{}, v1alpha1.ExWCreatingPhase, ""),
			Entry("ready revision with ready pod is running",
				InferenceInput{
					Service: &InferenceServiceState{LatestCreatedRevision: "rev-1", LatestReadyRevision: "rev-1"},
					Pods:    []corev1.Pod{readyPod},
				}, v1alpha1.ExWRunningPhase, ""),
			Entry("failed rollout keeps older revision running",
				InferenceInput{
					Service: &InferenceServiceState{
						LatestCreatedRevision: "rev-2",
						LatestReadyRevision:   "rev-1",
						Conditions: []InferenceServiceCondition{{
							Status: corev1.ConditionFalse, Message: "unable to fetch image",
						}},
					},
				}, v1alpha1.ExWRunningPhase,
				"Workload is ready. Update failed: Failed to pull container image."),
			Entry("failure on current ready revision is degraded",
				InferenceInput{
					Service: &InferenceServiceState{
						LatestCreatedRevision: "rev-1",
						LatestReadyRevision:   "rev-1",
						Conditions: []InferenceServiceCondition{{
							Status: corev1.ConditionFalse, Message: "unable to fetch image",
						}},
					},
				}, v1alpha1.ExWDegradedPhase, "Failed to pull container image."),
		)
	})

	Context("when pending and image-pull pods are mixed", func() {
		It("is independent of pod order", func() {
			pending := corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodPending}}
			imagePull := corev1.Pod{Status: corev1.PodStatus{
				Phase: corev1.PodPending,
				ContainerStatuses: []corev1.ContainerStatus{{State: corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff"},
				}}},
			}}
			service := &InferenceServiceState{LatestCreatedRevision: "rev-1"}

			for _, pods := range [][]corev1.Pod{{pending, imagePull}, {imagePull, pending}} {
				phase, message := ComputeInferenceStatus(InferenceInput{Service: service, Pods: pods}, nil)
				Expect(phase).To(Equal(v1alpha1.ExWPendingPhase))
				Expect(message).To(BeEmpty())
			}
		})
	})
})

var _ = Describe("ComputeInferenceStatus with the snapshot plugin", func() {
	failed := &dromosv1alpha1.PluginStatus{
		Phase:  dromosv1alpha1.PluginPhaseFailed,
		Issues: []dromosv1alpha1.Issue{{Description: "the snapshot capture failed, so replicas start cold"}},
	}
	readyPod := corev1.Pod{Status: corev1.PodStatus{
		Phase:      corev1.PodRunning,
		Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}},
	}}
	running := InferenceInput{
		Service: &InferenceServiceState{LatestCreatedRevision: "rev-1", LatestReadyRevision: "rev-1"},
		Pods:    []corev1.Pod{readyPod},
	}
	// Running on the older revision, with a message of its own.
	runningAfterFailedUpdate := InferenceInput{Service: &InferenceServiceState{
		LatestCreatedRevision: "rev-2", LatestReadyRevision: "rev-1",
		Conditions: []InferenceServiceCondition{{Status: corev1.ConditionFalse, Message: "unable to fetch image"}},
	}}

	DescribeTable("shows a running workload Degraded while the snapshot plugin is Failed",
		func(input InferenceInput, snapshot *dromosv1alpha1.PluginStatus,
			wantPhase v1alpha1.ExternalWorkloadPhase, wantMessage string) {
			gotPhase, gotMessage := ComputeInferenceStatus(input, snapshot)

			Expect(gotPhase).To(Equal(wantPhase))
			Expect(gotMessage).To(Equal(wantMessage))
		},
		Entry("Failed, with the plugin's issue as the message", running, failed,
			v1alpha1.ExWDegradedPhase, "the snapshot capture failed, so replicas start cold"),
		Entry("Failed with no issue keeps the message", runningAfterFailedUpdate,
			&dromosv1alpha1.PluginStatus{Phase: dromosv1alpha1.PluginPhaseFailed},
			v1alpha1.ExWDegradedPhase, "Workload is ready. Update failed: Failed to pull container image."),
		Entry("Ready", running, &dromosv1alpha1.PluginStatus{Phase: dromosv1alpha1.PluginPhaseReady},
			v1alpha1.ExWRunningPhase, ""),
		// Not opted in, or an older cluster without the plugin.
		Entry("no snapshot plugin", running, nil, v1alpha1.ExWRunningPhase, ""),
		// Only Running hides the cold start; any other phase already says more.
		Entry("not Running", InferenceInput{}, failed, v1alpha1.ExWCreatingPhase, ""),
	)
})
