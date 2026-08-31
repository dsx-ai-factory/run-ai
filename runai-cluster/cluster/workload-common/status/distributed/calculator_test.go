/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed_test

import (
	"github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	statuscontract "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/contract"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/distributed"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Calculate", func() {
	now := metav1.Now()

	makePod := func(phase corev1.PodPhase, ready bool) corev1.Pod {
		pod := corev1.Pod{Status: corev1.PodStatus{Phase: phase}}
		if phase == corev1.PodRunning {
			// RunaiPhaseCalculator needs an explicit PodReady condition to distinguish
			// Initializing (PodReady=False) from Pending (no conditions at all).
			status := corev1.ConditionFalse
			if ready {
				status = corev1.ConditionTrue
			}
			pod.Status.Conditions = []corev1.PodCondition{
				{Type: corev1.PodReady, Status: status},
			}
		}
		return pod
	}

	condition := func(t v1.JobConditionType, status corev1.ConditionStatus) v1.JobCondition {
		return v1.JobCondition{Type: t, Status: status}
	}

	// Shorthand for a non-nil JobStatus with no conditions.
	emptyStatus := func() *v1.JobStatus { return &v1.JobStatus{} }

	DescribeTable("phase precedence",
		func(in distributed.Input, wantPhase statuscontract.Phase, wantMsg string) {
			r := distributed.Calculate(in)
			Expect(r.Phase).To(Equal(wantPhase), "phase")
			Expect(r.Message).To(Equal(wantMsg), "message")
		},

		Entry("nil JobStatus → Creating",
			distributed.Input{},
			statuscontract.Creating, ""),

		Entry("empty JobStatus, no pods → Creating",
			distributed.Input{JobStatus: emptyStatus()},
			statuscontract.Creating, ""),

		Entry("Terminating beats everything",
			distributed.Input{
				JobStatus:   &v1.JobStatus{Conditions: []v1.JobCondition{condition(v1.JobSucceeded, corev1.ConditionTrue)}},
				Terminating: true,
			},
			statuscontract.Terminating, ""),

		Entry("Failed condition → Failed with job message",
			distributed.Input{
				JobStatus: &v1.JobStatus{Conditions: []v1.JobCondition{
					{Type: v1.JobFailed, Status: corev1.ConditionTrue, Message: "OOMKilled"},
				}},
			},
			statuscontract.Failed, "OOMKilled"),

		Entry("Failed condition with no message → generic failed",
			distributed.Input{
				JobStatus: &v1.JobStatus{Conditions: []v1.JobCondition{
					condition(v1.JobFailed, corev1.ConditionTrue),
				}},
			},
			statuscontract.Failed, "job reached failed state"),

		Entry("Succeeded via CompletionTime → Completed",
			distributed.Input{
				JobStatus: &v1.JobStatus{CompletionTime: &now},
			},
			statuscontract.Completed, ""),

		Entry("CompletionTime set but also Failed → Failed wins",
			distributed.Input{
				JobStatus: &v1.JobStatus{
					CompletionTime: &now,
					Conditions:     []v1.JobCondition{condition(v1.JobFailed, corev1.ConditionTrue)},
				},
			},
			statuscontract.Failed, "job reached failed state"),

		Entry("suspend intent + no Suspended condition → Suspending",
			distributed.Input{
				JobStatus: emptyStatus(),
				Suspend:   true,
			},
			statuscontract.Suspending, ""),

		Entry("suspend=true + Suspended condition → Suspended with readyPods==0",
			distributed.Input{
				JobStatus: &v1.JobStatus{Conditions: []v1.JobCondition{
					{Type: v1.JobSuspended, Status: corev1.ConditionTrue, Message: "user suspended"},
				}},
				Suspend: true,
			},
			statuscontract.Suspended, "user suspended"),

		Entry("suspend=true + Suspended condition + ready pod → Suspending",
			distributed.Input{
				JobStatus: &v1.JobStatus{Conditions: []v1.JobCondition{
					{Type: v1.JobSuspended, Status: corev1.ConditionTrue, Message: "user suspended"},
				}},
				Pods:    []corev1.Pod{makePod(corev1.PodRunning, true)},
				Suspend: true,
			},
			statuscontract.Suspending, ""),

		Entry("suspend cleared + Suspended condition still set → Resuming",
			distributed.Input{
				JobStatus: &v1.JobStatus{Conditions: []v1.JobCondition{
					condition(v1.JobSuspended, corev1.ConditionTrue),
				}},
				Suspend: false,
			},
			statuscontract.Resuming, ""),

		Entry("suspend cleared + Suspended condition + ready pod → Running",
			distributed.Input{
				JobStatus: &v1.JobStatus{Conditions: []v1.JobCondition{
					condition(v1.JobSuspended, corev1.ConditionTrue),
				}},
				Pods:    []corev1.Pod{makePod(corev1.PodRunning, true)},
				Suspend: false,
			},
			statuscontract.Running, ""),

		Entry("one ready pod → Running",
			distributed.Input{
				JobStatus: emptyStatus(),
				Pods:      []corev1.Pod{makePod(corev1.PodRunning, true)},
			},
			statuscontract.Running, ""),

		Entry("running + critical non-compute issue → Degraded",
			distributed.Input{
				JobStatus:      emptyStatus(),
				Pods:           []corev1.Pod{makePod(corev1.PodRunning, true)},
				CriticalIssues: []distributed.CriticalIssue{{Group: "networking", Compute: false}},
			},
			statuscontract.Degraded, "Workload is in degraded state due to a networking issue"),

		Entry("running + critical compute issue → Failed (compute beats Degraded)",
			distributed.Input{
				JobStatus:      emptyStatus(),
				Pods:           []corev1.Pod{makePod(corev1.PodRunning, true)},
				CriticalIssues: []distributed.CriticalIssue{{Group: "compute", Compute: true}},
			},
			statuscontract.Failed, "Workload failed due to a compute issue"),

		Entry("initializing pod, no ready → Initializing (single replica)",
			distributed.Input{
				JobStatus:    emptyStatus(),
				Pods:         []corev1.Pod{makePod(corev1.PodRunning, false)},
				MultiReplica: false,
			},
			statuscontract.Initializing, "Workload is initializing, the pod is still not ready"),

		Entry("initializing pod, multi-replica → Initializing with count message",
			distributed.Input{
				JobStatus:    emptyStatus(),
				Pods:         []corev1.Pod{makePod(corev1.PodRunning, false), makePod(corev1.PodRunning, false)},
				MultiReplica: true,
			},
			statuscontract.Initializing, "Workload is initializing, minimum ready pods not reached, waiting for 2 pods to be ready"),

		Entry("pending pod → Pending",
			distributed.Input{
				JobStatus: emptyStatus(),
				Pods:      []corev1.Pod{makePod(corev1.PodPending, false)},
			},
			statuscontract.Pending, ""),

		Entry("no ready pods + critical non-compute issue → Failed (not running)",
			distributed.Input{
				JobStatus:      emptyStatus(),
				CriticalIssues: []distributed.CriticalIssue{{Group: "networking", Compute: false}},
			},
			statuscontract.Failed, "Workload failed due to a networking issue"),

		Entry("multiple critical groups → plural message",
			distributed.Input{
				JobStatus: emptyStatus(),
				CriticalIssues: []distributed.CriticalIssue{
					{Group: "networking", Compute: false},
					{Group: "compute", Compute: true},
				},
			},
			statuscontract.Failed, "Workload failed due to compute, networking issues"),
	)

	Describe("pod counts", func() {
		It("sums ready + running + pending + initializing correctly", func() {
			pods := []corev1.Pod{
				makePod(corev1.PodRunning, true),  // ready
				makePod(corev1.PodRunning, false), // initializing (running, not ready)
				makePod(corev1.PodPending, false), // pending
			}
			r := distributed.Calculate(distributed.Input{JobStatus: emptyStatus(), Pods: pods})
			Expect(r.ReadyPods).To(Equal(int32(1)))
			Expect(r.RunningPods).To(Equal(int32(2))) // both PodRunning pods
			Expect(r.InitializingPods).To(Equal(int32(1)))
			Expect(r.PendingPods).To(Equal(int32(1)))
		})

		It("derives SucceededPods and FailedPods from ReplicaStatuses", func() {
			r := distributed.Calculate(distributed.Input{
				JobStatus: &v1.JobStatus{
					ReplicaStatuses: map[v1.ReplicaType]*v1.ReplicaStatus{
						"Worker": {Succeeded: 2, Failed: 1},
						"Master": {Succeeded: 1},
					},
				},
			})
			Expect(r.SucceededPods).To(Equal(int32(3)))
			Expect(r.FailedPods).To(Equal(int32(1)))
		})
	})

	Describe("sticky completionTime", func() {
		It("uses PriorCompletionTime when JobStatus.CompletionTime is nil", func() {
			prior := metav1.Now()
			r := distributed.Calculate(distributed.Input{
				JobStatus:           emptyStatus(),
				PriorCompletionTime: &prior,
			})
			Expect(r.CompletionTime).To(Equal(&prior))
		})

		It("uses JobStatus.CompletionTime when both are set", func() {
			prior := metav1.Now()
			fresh := metav1.Now()
			r := distributed.Calculate(distributed.Input{
				JobStatus:           &v1.JobStatus{CompletionTime: &fresh},
				PriorCompletionTime: &prior,
			})
			Expect(r.CompletionTime).To(Equal(&fresh))
		})

		It("returns nil when neither is set", func() {
			r := distributed.Calculate(distributed.Input{JobStatus: emptyStatus()})
			Expect(r.CompletionTime).To(BeNil())
		})
	})
})
