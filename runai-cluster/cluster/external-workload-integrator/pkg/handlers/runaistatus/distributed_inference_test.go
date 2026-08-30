package runaistatus

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mnnvltypes "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("ComputeDistributedInferenceStatus", func() {
	Describe("legacy phase parity", func() {
		DescribeTable("calculates the expected phase and message",
			func(input DistributedInferenceInput, phase v1alpha1.ExternalWorkloadPhase, message string) {
				result := ComputeDistributedInferenceStatus(input)

				Expect(result.Phase).To(Equal(phase))
				Expect(result.Message).To(Equal(message))
			},
			Entry("missing LWS is creating",
				DistributedInferenceInput{}, v1alpha1.ExWCreatingPhase, ""),
			Entry("missing LWS with a build failure is failed",
				DistributedInferenceInput{BuildError: ptr.To("LeaderWorkerSet not found")},
				v1alpha1.ExWFailedPhase, ""),
			Entry("zero replicas is running",
				DistributedInferenceInput{Exists: true, Replicas: 0, GroupSize: 1},
				v1alpha1.ExWRunningPhase, "Replicas: 0"),
			Entry("missing resources degrade the workload and are sorted in the message",
				DistributedInferenceInput{Exists: true, Replicas: 0, GroupSize: 1,
					MissingResources: []string{"gpu", "cpu"}},
				v1alpha1.ExWDegradedPhase, "Replicas: 0; Missing resources: cpu, gpu"),
			Entry("one ready and one failed group is degraded",
				DistributedInferenceInput{Exists: true, Replicas: 2, GroupSize: 1, Pods: []corev1.Pod{
					distributedInferenceGroupPod("0", "0", corev1.PodRunning, true, ""),
					distributedInferenceGroupPod("1", "0", corev1.PodFailed, false, ""),
				}}, v1alpha1.ExWDegradedPhase, "Replicas: 2; Ready: 1; Failed: 1"),
			Entry("missing group is initializing beside ready group",
				DistributedInferenceInput{Exists: true, Replicas: 2, GroupSize: 1, Pods: []corev1.Pod{
					distributedInferenceGroupPod("0", "0", corev1.PodRunning, true, ""),
				}}, v1alpha1.ExWInitializingPhase, "Replicas: 2; Ready: 1; Pending: 1"),
			Entry("ready leader keeps group ready while worker is running",
				DistributedInferenceInput{Exists: true, Replicas: 1, GroupSize: 2, Pods: []corev1.Pod{
					distributedInferenceGroupPod("0", "0", corev1.PodRunning, true, ""),
					distributedInferenceGroupPod("0", "1", corev1.PodRunning, false, ""),
				}}, v1alpha1.ExWRunningPhase, "Replicas: 1; Ready: 1"),
		)
	})

	Context("when MNNVL is enabled but not ready", func() {
		It("degrades the workload and includes a user-facing message", func() {
			result := ComputeDistributedInferenceStatus(DistributedInferenceInput{
				Exists: true, Replicas: 0, GroupSize: 1, MNNVLEnabled: true,
			})

			Expect(result.Phase).To(Equal(v1alpha1.ExWDegradedPhase))
			Expect(result.Message).To(Equal(
				"Replicas: 0; " + mnnvltypes.UserFacingFeatureName + " is not ready"))
		})
	})
})

func distributedInferenceGroupPod(group, worker string, phase corev1.PodPhase, ready bool, waitingReason string) corev1.Pod {
	pod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
		"leaderworkerset.sigs.k8s.io/group-index":  group,
		"leaderworkerset.sigs.k8s.io/worker-index": worker,
	}}, Status: corev1.PodStatus{Phase: phase}}
	if ready {
		pod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}
	}
	if waitingReason != "" {
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{{State: corev1.ContainerState{
			Waiting: &corev1.ContainerStateWaiting{Reason: waitingReason},
		}}}
	}
	return pod
}
