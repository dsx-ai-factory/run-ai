package runaistatus

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
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
				actualPhase, actualMessage := ComputeInferenceStatus(input)

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
				phase, message := ComputeInferenceStatus(InferenceInput{Service: service, Pods: pods})
				Expect(phase).To(Equal(v1alpha1.ExWPendingPhase))
				Expect(message).To(BeEmpty())
			}
		})
	})
})
