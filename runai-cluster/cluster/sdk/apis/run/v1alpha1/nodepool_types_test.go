package v1alpha1

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("SetNodePoolCondition", func() {
	oldTime := metav1.NewTime(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))

	It("adds a new true condition and stamps a transition time", func() {
		np := &NodePool{}
		np.SetNodePoolCondition(NodePoolCondition{Type: NodeTopologyMismatch, Status: corev1.ConditionTrue})

		Expect(np.Status.Conditions).To(HaveLen(1))
		Expect(np.Status.Conditions[0].LastTransitionTime.IsZero()).To(BeFalse())
	})

	It("does not add a new false condition", func() {
		np := &NodePool{}
		np.SetNodePoolCondition(NodePoolCondition{Type: NodeTopologyMismatch, Status: corev1.ConditionFalse})

		Expect(np.Status.Conditions).To(BeEmpty())
	})

	It("preserves the transition time when the status is unchanged", func() {
		np := &NodePool{Status: NodePoolStatus{Conditions: []NodePoolCondition{
			{Type: NodeTopologyMismatch, Status: corev1.ConditionTrue, LastTransitionTime: oldTime, Message: "old"},
		}}}
		np.SetNodePoolCondition(NodePoolCondition{Type: NodeTopologyMismatch, Status: corev1.ConditionTrue, Message: "new"})

		got := np.Status.Conditions[0]
		Expect(got.LastTransitionTime.Time.Equal(oldTime.Time)).To(BeTrue())
		Expect(got.Message).To(Equal("new"))
	})

	It("updates the transition time when the status changes", func() {
		np := &NodePool{Status: NodePoolStatus{Conditions: []NodePoolCondition{
			{Type: NodeTopologyMismatch, Status: corev1.ConditionTrue, LastTransitionTime: oldTime},
		}}}
		np.SetNodePoolCondition(NodePoolCondition{Type: NodeTopologyMismatch, Status: corev1.ConditionFalse})

		got := np.Status.Conditions[0]
		Expect(got.Status).To(Equal(corev1.ConditionFalse))
		Expect(got.LastTransitionTime.Time.Equal(oldTime.Time)).To(BeFalse())
	})
})
