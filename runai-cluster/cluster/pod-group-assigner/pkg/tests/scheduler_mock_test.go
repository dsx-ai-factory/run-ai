package tests

import (
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/podgroup/assigner"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TestSchedulerMock struct {
}

var SchedulerMock = &TestSchedulerMock{}

func (sm *TestSchedulerMock) UnschedulableOnNodePool(pg types.NamespacedName, assignmentParams *assigner.FullNodePoolAssignmentParams, k8sClient client.Client) {
	pgFromClient := getPodGroupFromClient(pg, k8sClient)
	pgFromClient.Status.SchedulingConditions = append(pgFromClient.Status.SchedulingConditions,
		getUnschedulableOnNodePoolCondition(assignmentParams.NodePoolName))
	expectUpdateResourceStatus(k8sClient, pgFromClient)

	EventuallyWithOffset(1, func() bool {
		pgFromClient = getPodGroupFromClient(pg, k8sClient)
		for _, condition := range pgFromClient.Status.SchedulingConditions {
			if condition.Type == kaiv2alpha2.UnschedulableOnNodePool &&
				condition.NodePool == assignmentParams.NodePoolName {
				return true
			}
		}

		return false
	}, timeout, interval).Should(BeTrue())

	if assignmentParams.MarkUnschedulable {
		markAsUnschedulable(pg, k8sClient)
	}
}

func (sm *TestSchedulerMock) UnschedulableOnNodePoolNoValidate(pg types.NamespacedName, assignmentParams *assigner.FullNodePoolAssignmentParams, k8sClient client.Client) {
	pgFromClient := getPodGroupFromClient(pg, k8sClient)
	pgFromClient.Status.SchedulingConditions = append(pgFromClient.Status.SchedulingConditions,
		getUnschedulableOnNodePoolCondition(assignmentParams.NodePoolName))
	expectUpdateResourceStatus(k8sClient, pgFromClient)

	if assignmentParams.MarkUnschedulable {
		markAsUnschedulable(pg, k8sClient)
	}
}

func (sm *TestSchedulerMock) Running(pg types.NamespacedName, k8sClient client.Client) {
	pods := getPodGroupPods(pg)
	Expect(pods).ToNot(BeNil())

	pod := pods.Items[len(pods.Items)-1]
	pod.Status.Phase = corev1.PodRunning
	expectUpdateResourceStatus(k8sClient, &pod)

	Eventually(func() bool {
		pods = getPodGroupPods(pg)
		pod = pods.Items[len(pods.Items)-1]

		return pod.Status.Phase == corev1.PodRunning
	}, timeout, interval).Should(BeTrue())
}

func markAsUnschedulable(pg types.NamespacedName, k8sClient client.Client) {
	var pods *corev1.PodList

	Eventually(func() bool {
		pods = getPodGroupPods(pg)
		return pods != nil && len(pods.Items) > 0
	}, timeout, interval).Should(BeTrue())

	condition := &corev1.PodCondition{
		Type:   corev1.PodScheduled,
		Status: corev1.ConditionFalse,
		Reason: corev1.PodReasonUnschedulable,
	}

	for i := range pods.Items {
		pod := pods.Items[i]
		if podutil.UpdatePodCondition(&pod.Status, condition) {
			Eventually(func() bool {
				apiPod := &corev1.Pod{}

				err := k8sClient.Get(apiCtx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, apiPod)
				if err != nil {
					return false
				}

				_ = podutil.UpdatePodCondition(&apiPod.Status, condition)
				err = k8sClient.Status().Update(apiCtx, apiPod)

				return err == nil
			}, timeout, interval).Should(BeTrue())
		}
	}

	Eventually(func() bool {
		pods = getPodGroupPods(pg)
		if pods == nil || len(pods.Items) == 0 {
			return false
		}

		for _, pod := range pods.Items {
			foundConditionForThisPod := false

			for _, pCondition := range pod.Status.Conditions {
				if pCondition.Type == corev1.PodScheduled ||
					pCondition.Status == corev1.ConditionFalse ||
					pCondition.Reason == corev1.PodReasonUnschedulable {
					foundConditionForThisPod = true
					break
				}
			}

			if !foundConditionForThisPod {
				return false
			}
		}

		return true
	}, timeout, interval).Should(BeTrue())
}

func getUnschedulableOnNodePoolCondition(nodePool string) kaiv2alpha2.SchedulingCondition {
	return kaiv2alpha2.SchedulingCondition{
		Type:     kaiv2alpha2.UnschedulableOnNodePool,
		NodePool: nodePool,
	}
}
