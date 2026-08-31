/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaistatus

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	statuscontract "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/contract"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestRunaiStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RunaiStatus Suite")
}

func int32p(v int32) *int32 { return &v }

func jobWith(conds []batchv1.JobCondition, active int32, ready *int32, parallelism *int32) *batchv1.Job {
	return &batchv1.Job{
		Spec:   batchv1.JobSpec{Parallelism: parallelism},
		Status: batchv1.JobStatus{Conditions: conds, Active: active, Ready: ready},
	}
}

func cond(t batchv1.JobConditionType, msg string) batchv1.JobCondition {
	return batchv1.JobCondition{Type: t, Status: corev1.ConditionTrue, Message: msg}
}

func readyPod() corev1.Pod {
	return corev1.Pod{Status: corev1.PodStatus{
		Phase:      corev1.PodRunning,
		Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}},
	}}
}

func runningNotReadyPod() corev1.Pod {
	return corev1.Pod{Status: corev1.PodStatus{
		Phase:      corev1.PodRunning,
		Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionFalse}},
	}}
}

// scheduledPendingPod is bound to a node but not yet running (image pull, container creating, etc.).
func scheduledPendingPod() corev1.Pod {
	return corev1.Pod{Status: corev1.PodStatus{
		Phase:      corev1.PodPending,
		Conditions: []corev1.PodCondition{{Type: corev1.PodScheduled, Status: corev1.ConditionTrue}},
	}}
}

func pendingPod() corev1.Pod {
	return corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodPending}}
}

func wlrWith(phase dromosv1alpha1.WorkloadRunnerPhase, groups ...string) *dromosv1alpha1.WorkloadRunner {
	issues := make([]dromosv1alpha1.Issue, 0, len(groups))
	for _, g := range groups {
		issues = append(issues, dromosv1alpha1.Issue{Group: g, Description: "x", IsCritical: true})
	}
	return &dromosv1alpha1.WorkloadRunner{Status: dromosv1alpha1.WorkloadRunnerStatus{Phase: phase, Issues: issues}}
}

func wlrWithIssues(phase dromosv1alpha1.WorkloadRunnerPhase, issues ...dromosv1alpha1.Issue) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{Status: dromosv1alpha1.WorkloadRunnerStatus{Phase: phase, Issues: issues}}
}

func issue(group string, critical bool) dromosv1alpha1.Issue {
	return dromosv1alpha1.Issue{Group: group, Description: "x", IsCritical: critical}
}

var _ = Describe("ComputeStatus", func() {
	DescribeTable("derives the Run:ai phase and message from job, pods and wlr",
		func(job *batchv1.Job, pods []corev1.Pod, wlr *dromosv1alpha1.WorkloadRunner,
			wantPhase v1alpha1.ExternalWorkloadPhase, wantMessage string) {
			gotPhase, gotMsg := ComputeStatus(job, pods, wlr)
			Expect(gotPhase).To(Equal(wantPhase))
			Expect(gotMsg).To(Equal(wantMessage))
		},
		Entry("nil job -> Unknown",
			nil, nil, nil,
			v1alpha1.ExWUnknownPhase, ""),
		Entry("suspended -> Stopped",
			jobWith([]batchv1.JobCondition{cond(batchv1.JobSuspended, "")}, 0, nil, nil), nil, nil,
			v1alpha1.ExWStoppedPhase, ""),
		Entry("complete -> Completed",
			jobWith([]batchv1.JobCondition{cond(batchv1.JobComplete, "")}, 0, nil, nil), nil, nil,
			v1alpha1.ExWCompletedPhase, ""),
		Entry("failed, no wlr, no cond msg -> generic",
			jobWith([]batchv1.JobCondition{cond(batchv1.JobFailed, "")}, 0, nil, nil), nil, nil,
			v1alpha1.ExWFailedPhase, statuscontract.MsgGenericJobFailed),
		Entry("failed with cond message -> cond message",
			jobWith([]batchv1.JobCondition{cond(batchv1.JobFailed, "container exited 1")}, 0, nil, nil), nil, nil,
			v1alpha1.ExWFailedPhase, "container exited 1"),
		Entry("job failed with message + wlr issue -> job message wins (v1 precedence)",
			jobWith([]batchv1.JobCondition{cond(batchv1.JobFailed, "OOMKilled: exit 137")}, 0, nil, nil), nil,
			wlrWith(dromosv1alpha1.WorkloadRunnerPhaseFailed, "Compute"),
			v1alpha1.ExWFailedPhase, "OOMKilled: exit 137"),
		Entry("wlr failed -> Failed phase (message asserted in v1-parity test)",
			jobWith(nil, 1, nil, nil), nil,
			wlrWith(dromosv1alpha1.WorkloadRunnerPhaseFailed, "Compute"),
			v1alpha1.ExWFailedPhase, "Workload failed due to a Compute issue"),
		Entry("ready pod -> Running",
			jobWith(nil, 1, int32p(1), nil), []corev1.Pod{readyPod()}, nil,
			v1alpha1.ExWRunningPhase, ""),
		Entry("single-pod initializing -> Initializing single message",
			jobWith(nil, 1, int32p(0), int32p(1)), []corev1.Pod{runningNotReadyPod()}, nil,
			v1alpha1.ExWInitializingPhase, statuscontract.MsgInitializingSinglePod),
		Entry("multi-pod scheduled-pending -> Initializing multi message",
			jobWith(nil, 3, int32p(0), int32p(3)), []corev1.Pod{scheduledPendingPod()}, nil,
			v1alpha1.ExWInitializingPhase, "Workload is initializing, minimum ready pods not reached, waiting for 1 pods to be ready"),
		Entry("scheduled pending (e.g. ImagePullBackOff) -> Initializing, not Pending",
			jobWith(nil, 1, int32p(0), int32p(1)), []corev1.Pod{scheduledPendingPod()}, nil,
			v1alpha1.ExWInitializingPhase, statuscontract.MsgInitializingSinglePod),
		Entry("active but pods pending -> Pending",
			jobWith(nil, 1, int32p(0), nil), []corev1.Pod{pendingPod()}, nil,
			v1alpha1.ExWPendingPhase, ""),
		Entry("no pods, not active -> Creating",
			jobWith(nil, 0, nil, nil), nil, nil,
			v1alpha1.ExWCreatingPhase, ""),
		Entry("wlr degraded with no issues -> generic degraded message",
			jobWith(nil, 1, int32p(1), nil), nil,
			wlrWith(dromosv1alpha1.WorkloadRunnerPhaseDegraded),
			v1alpha1.ExWDegradedPhase, statuscontract.MsgGenericDegraded),
		// only critical issues affect the phase message (exact strings asserted in the
		// v1-parity table below; these cases pin the non-critical filtering behavior).
		Entry("wlr failed with critical Compute + non-critical Network -> only Compute named",
			jobWith(nil, 1, nil, nil), nil,
			wlrWithIssues(dromosv1alpha1.WorkloadRunnerPhaseFailed, issue("Compute", true), issue("Network", false)),
			v1alpha1.ExWFailedPhase, "Workload failed due to a Compute issue"),
		Entry("wlr failed with only non-critical issue -> generic fallback",
			jobWith(nil, 1, nil, nil), nil,
			wlrWithIssues(dromosv1alpha1.WorkloadRunnerPhaseFailed, issue("Network", false)),
			v1alpha1.ExWFailedPhase, statuscontract.MsgGenericJobFailed),
		Entry("wlr degraded with critical Storage + non-critical Network -> only Storage named",
			jobWith(nil, 1, int32p(1), nil), nil,
			wlrWithIssues(dromosv1alpha1.WorkloadRunnerPhaseDegraded, issue("Storage", true), issue("Network", false)),
			v1alpha1.ExWDegradedPhase, "Workload is in degraded state due to a Storage issue"),
		Entry("wlr degraded with only non-critical issues -> generic fallback",
			jobWith(nil, 1, int32p(1), nil), nil,
			wlrWithIssues(dromosv1alpha1.WorkloadRunnerPhaseDegraded, issue("Network", false), issue("Storage", false)),
			v1alpha1.ExWDegradedPhase, statuscontract.MsgGenericDegraded),
	)
})

// v1 parity oracle: the want strings below are copied VERBATIM from v1's status calculator test
// (workload-controller/pkg/status/aggregator/workload_status_calculator_test.go). Only edit a string
// here to track a change in v1 — never to make our code pass. In v2 the Failed/Degraded decision comes
// from the WorkloadRunner phase; this table pins the message we render from its issues.
var _ = Describe("ComputeStatus issue messages (v1 parity)", func() {
	DescribeTable("renders the wlr issue message exactly as v1 does",
		func(wlrPhase dromosv1alpha1.WorkloadRunnerPhase, groups []string,
			wantPhase v1alpha1.ExternalWorkloadPhase, want string) {
			// Active Job so only the wlr phase drives Failed/Degraded (job itself not terminal).
			gotPhase, gotMsg := ComputeStatus(jobWith(nil, 1, nil, nil), nil, wlrWith(wlrPhase, groups...))
			Expect(gotPhase).To(Equal(wantPhase))
			Expect(gotMsg).To(Equal(want))
		},
		Entry("failed, single Network", dromosv1alpha1.WorkloadRunnerPhaseFailed, []string{"Network"},
			v1alpha1.ExWFailedPhase, "Workload failed due to a Network issue"),
		Entry("failed, single Internal", dromosv1alpha1.WorkloadRunnerPhaseFailed, []string{"Internal"},
			v1alpha1.ExWFailedPhase, "Workload failed due to a Internal issue"),
		Entry("failed, Compute+Network", dromosv1alpha1.WorkloadRunnerPhaseFailed, []string{"Network", "Compute"},
			v1alpha1.ExWFailedPhase, "Workload failed due to Compute, Network issues"),
		Entry("failed, Network+Storage", dromosv1alpha1.WorkloadRunnerPhaseFailed, []string{"Storage", "Network"},
			v1alpha1.ExWFailedPhase, "Workload failed due to Network, Storage issues"),
		Entry("failed, Internal+Network+Storage", dromosv1alpha1.WorkloadRunnerPhaseFailed, []string{"Storage", "Internal", "Network"},
			v1alpha1.ExWFailedPhase, "Workload failed due to Internal, Network, Storage issues"),
		Entry("degraded, single Network", dromosv1alpha1.WorkloadRunnerPhaseDegraded, []string{"Network"},
			v1alpha1.ExWDegradedPhase, "Workload is in degraded state due to a Network issue"),
		Entry("degraded, Network+Storage", dromosv1alpha1.WorkloadRunnerPhaseDegraded, []string{"Storage", "Network"},
			v1alpha1.ExWDegradedPhase, "Workload is in degraded state due to Network, Storage issues"),
		Entry("degraded, Internal+Network+Storage", dromosv1alpha1.WorkloadRunnerPhaseDegraded, []string{"Storage", "Internal", "Network"},
			v1alpha1.ExWDegradedPhase, "Workload is in degraded state due to Internal, Network, Storage issues"),
	)
})
