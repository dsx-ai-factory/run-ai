package tests

import (
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/common"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/podgroup/assigner"

	nodepoolutils "github.com/run-ai/runai/runai-cluster/common/node-pool-utils/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func expectNodePoolAssignment(g Gomega, offset int, pg types.NamespacedName, assignmentParams *assigner.FullNodePoolAssignmentParams,
	checkLastEvent bool, validateNoUnschedulableOnNodePoolCondition bool, k8sClient client.Client) {
	assignmentParams.Queue = getPodGroupQueue(testProject, assignmentParams.NodePoolName)

	expectNodePoolAssignmentInner(g, offset, pg, assignmentParams, checkLastEvent, validateNoUnschedulableOnNodePoolCondition, k8sClient)
}

func expectNodePoolAssignmentInner(g Gomega, offset int, pg types.NamespacedName, assignmentParams *assigner.FullNodePoolAssignmentParams,
	checkLastEvent bool, validateNoUnschedulableOnNodePoolCondition bool, k8sClient client.Client) {
	pgFromClient := getPodGroupFromClient(pg, k8sClient)
	g.ExpectWithOffset(offset, pgFromClient).ToNot(BeNil())
	g.ExpectWithOffset(offset, pgFromClient.Spec.Queue).To(Equal(assignmentParams.Queue),
		"expected pod group <%v> to have Queue <%s>",
		pg, assignmentParams.Queue)
	g.ExpectWithOffset(offset, common.GetMarkUnschedulableValue(pgFromClient.Spec.MarkUnschedulable)).To(Equal(assignmentParams.MarkUnschedulable),
		"expected pod group <%v> to have MarkUnschedulable <%t>",
		pg, assignmentParams.MarkUnschedulable)
	g.ExpectWithOffset(offset, common.GetSchedulingBackoffValue(pgFromClient.Spec.SchedulingBackoff)).To(Equal(assignmentParams.SchedulingBackoff),
		"expected pod group <%v> to have SchedulingBackoff <%d>",
		pg, assignmentParams.SchedulingBackoff)

	if assignmentParams.NodePoolName != testUnexistingNodePool {
		g.ExpectWithOffset(offset, pgFromClient.Labels).To(HaveKeyWithValue(RunaiQueueLabel, assignmentParams.Queue),
			"expected pod group <%v> to have queue label of: <%s>",
			pg, assignmentParams.Queue)
	}

	foundNodePoolName := nodepoolutils.GetNodePoolNameFromLabels(pgFromClient.Labels, config.Config().NodePoolLabelKey, config.Config().DefaultNodepoolName)
	g.ExpectWithOffset(offset, foundNodePoolName).To(Equal(assignmentParams.NodePoolName),
		"expected pod group <%v> to have nodepool <%v> label",
		pg, assignmentParams.NodePoolName)

	if assignmentParams.NodePoolName != testDefaultNodePoolName {
		g.ExpectWithOffset(offset, pgFromClient.Labels).To(HaveKeyWithValue(testNodePoolLabelKey, assignmentParams.NodePoolName),
			"expected pod group <%v> to have NodePool label of: <%s>",
			pg, assignmentParams.NodePoolName)
	} else {
		g.ExpectWithOffset(offset, pgFromClient.Labels).ToNot(HaveKey(testNodePoolLabelKey),
			"expected pod group <%v> to not have NodePool label (default node pool)",
			pg)
	}

	if validateNoUnschedulableOnNodePoolCondition {
		for _, c := range pgFromClient.Status.SchedulingConditions {
			g.ExpectWithOffset(offset, c.Type).ToNot(Equal(kaiv2alpha2.UnschedulableOnNodePool),
				"expected pod group <%v> to not have UnschedulableOnNodePool SchedulingCondition", pg)
		}

		return
	}

	if !checkLastEvent {
		return
	}

	conditions := pgFromClient.Status.SchedulingConditions
	if len(conditions) == 0 {
		return
	}

	// checking that last event is NOT same nodePool
	lastCondition := conditions[len(conditions)-1]
	g.ExpectWithOffset(offset, lastCondition.NodePool).ToNot(Equal(assignmentParams.NodePoolName),
		"expected pod group's <%v> last scheduling condition to have a different nodepool then the current assignment",
		pg)
}

// eventuallyExpectNodePoolAssignment reconciles once (the manual stand-in for
// the watch event the manager would have delivered) and then asserts the
// resulting assignment synchronously. The reconcile must happen exactly once
// outside the assertion: re-running it (as a polling loop would) could advance
// the round-robin an extra step.
func eventuallyExpectNodePoolAssignment(pg types.NamespacedName, assignmentParams *assigner.FullNodePoolAssignmentParams, k8sClient client.Client) {
	reconcile(pg)
	expectNodePoolAssignment(Default, 2, pg, assignmentParams, true, false, k8sClient)
}

func eventuallyExpectNodePoolAssignmentWithQueueName(pg types.NamespacedName, queueName string, assignmentParams *assigner.FullNodePoolAssignmentParams, k8sClient client.Client) {
	reconcile(pg)
	assignmentParams.Queue = queueName
	expectNodePoolAssignmentInner(Default, 2, pg, assignmentParams, true, false, k8sClient)
}

func eventuallyExpectNodePoolAssignmentNoLastEventCheck(pg types.NamespacedName, assignmentParams *assigner.FullNodePoolAssignmentParams, k8sClient client.Client) {
	ExpectWithOffset(1, assignmentParams.SchedulingBackoff).To(BeEquivalentTo(-1),
		"Don't use eventuallyExpectNodePoolAssignmentNoLastEventCheck func if SchedulingBackoff is not -1!")

	reconcile(pg)
	expectNodePoolAssignment(Default, 2, pg, assignmentParams, false, false, k8sClient)
}

func eventuallyExpectNodePoolAssignmentRemoveUnschedulableOnNodePoolSchedulingCondition(pg types.NamespacedName, assignmentParams *assigner.FullNodePoolAssignmentParams, k8sClient client.Client) {
	ExpectWithOffset(1, assignmentParams.RemoveUnschedulableOnNodePoolSchedulingCondition).To(BeEquivalentTo(true),
		"Don't use eventuallyExpectNodePoolAssignmentRemoveUnschedulableOnNodePoolSchedulingCondition func if RemoveUnschedulableOnNodePoolSchedulingCondition is not true!")
	ExpectWithOffset(1, assignmentParams.SchedulingBackoff).ToNot(BeEquivalentTo(-1),
		"Don't use eventuallyExpectNodePoolAssignmentNoLastEventCheck func if SchedulingBackoff is -1!")

	reconcile(pg)
	expectNodePoolAssignment(Default, 2, pg, assignmentParams, false, true, k8sClient)
}

// consistentlyExpectNodePoolAssignment proves the assignment is stable: it
// reconciles once more and asserts the assignment did not change. With no
// manager there are no spurious events, so a single idempotent reconcile is a
// faithful, deterministic substitute for the original Consistently poll.
func consistentlyExpectNodePoolAssignment(pg types.NamespacedName, assignmentParams *assigner.FullNodePoolAssignmentParams, k8sClient client.Client) {
	reconcile(pg)
	expectNodePoolAssignment(Default, 2, pg, assignmentParams, false, false, k8sClient)
}

func expectUnschedulableOnNodePool(pg types.NamespacedName, assignmentParams *assigner.FullNodePoolAssignmentParams, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		pgFromClient := getPodGroupFromClient(pg, k8sClient)
		conditions := pgFromClient.Status.SchedulingConditions
		g.ExpectWithOffset(1, len(conditions) > 0).To(BeTrue(), "expected to have at least 1 Scheduling Condition")

		lastCondition := conditions[len(conditions)-1]
		g.ExpectWithOffset(1,
			lastCondition).To(Equal(getUnschedulableOnNodePoolCondition(assignmentParams.NodePoolName)),
			"expected pod group <%v> to have UnschedulableOnNodePool SchedulingCondition, on: <%s>",
			pg, assignmentParams.NodePoolName)
	}, timeout, interval).Should(Succeed())

	if assignmentParams.MarkUnschedulable {
		expectUnschedulable(2, pg, k8sClient)
	}
}

// expectNoUnschedulableOnNodePool asserts the controller removed the
// UnschedulableOnNodePool condition the scheduler mock just wrote. The removal
// is the controller's job (it happens when only one node pool is available), so
// we reconcile once first - the manual stand-in for the watch event the mock's
// status update would have triggered in the envtest-based suite.
func expectNoUnschedulableOnNodePool(pg types.NamespacedName, assignmentParams *assigner.FullNodePoolAssignmentParams, k8sClient client.Client) {
	reconcile(pg)

	pgFromClient := getPodGroupFromClient(pg, k8sClient)
	ExpectWithOffset(1, pgFromClient).NotTo(BeNil(), "expected pod group <%v> to exist", pg)
	conditions := pgFromClient.Status.SchedulingConditions

	unschedConditions := []kaiv2alpha2.SchedulingCondition{}

	for _, c := range conditions {
		if c.Type == kaiv2alpha2.UnschedulableOnNodePool {
			unschedConditions = append(unschedConditions, c)
		}
	}

	ExpectWithOffset(1, len(unschedConditions) == 0).To(BeTrue(), "expected to have 0 UnschedulableOnNodePool Scheduling Conditions")

	if assignmentParams.MarkUnschedulable {
		expectUnschedulable(2, pg, k8sClient)
	}
}

func expectUnschedulable(offset int, pg types.NamespacedName, k8sClient client.Client) {
	EventuallyWithOffset(offset, func(g Gomega) {
		pods := getPodGroupPods(pg)
		g.ExpectWithOffset(offset, pods).ToNot(BeNil())
		g.ExpectWithOffset(offset, len(pods.Items) > 0).To(BeTrue())

		for _, pod := range pods.Items {
			conditions := pod.Status.Conditions
			g.ExpectWithOffset(offset, len(conditions) > 0).To(BeTrue())

			foundConditionForThisPod := false

			for _, pCondition := range pod.Status.Conditions {
				if pCondition.Type == corev1.PodScheduled ||
					pCondition.Status == corev1.ConditionFalse ||
					pCondition.Reason == corev1.PodReasonUnschedulable {
					foundConditionForThisPod = true
					break
				}
			}

			g.ExpectWithOffset(offset, foundConditionForThisPod).To(BeTrue(),
				"expected podgroup/pod <%v/%s> to have <%v/%v/%v> condition",
				pg, pod.Name, corev1.PodScheduled, corev1.ConditionFalse, corev1.PodReasonUnschedulable)
		}
	}, timeout*4, interval).Should(Succeed())
}
