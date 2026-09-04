/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaistatus

import (
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testPodName = "worker-0"

func failedJob() *batchv1.Job {
	return jobWith([]batchv1.JobCondition{{Type: batchv1.JobFailed, Status: corev1.ConditionTrue}}, 0, nil, nil)
}

func podTerminated(container, reason, message string, exitCode int32) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: testPodName},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: container,
				State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
					ExitCode: exitCode, Reason: reason, Message: message,
				}},
			}},
		},
	}
}

func podWaiting(container, reason, message string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: testPodName},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:  container,
				State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: reason, Message: message}},
			}},
		},
	}
}

var _ = Describe("ComputeRuntimeIssue", func() {
	It("is not ok when the workload has not failed", func() {
		_, _, ok := ComputeRuntimeIssue(jobWith(nil, 1, int32p(1), nil), nil, nil)
		Expect(ok).To(BeFalse())
	})

	It("is not ok for a nil job", func() {
		_, _, ok := ComputeRuntimeIssue(nil, nil, nil)
		Expect(ok).To(BeFalse())
	})

	It("is not ok when the wlr already reports a critical issue", func() {
		wlr := wlrWithIssues(dromosv1alpha1.WorkloadRunnerPhaseFailed, issue("Storage", true))
		_, _, ok := ComputeRuntimeIssue(failedJob(), []corev1.Pod{podTerminated("main", "Error", "", 1)}, wlr)
		Expect(ok).To(BeFalse())
	})

	It("still synthesizes when the wlr only has non-critical issues", func() {
		wlr := wlrWithIssues(dromosv1alpha1.WorkloadRunnerPhaseFailed, issue("Network", false))
		_, desc, ok := ComputeRuntimeIssue(failedJob(), []corev1.Pod{podTerminated("main", "Error", "boom", 1)}, wlr)
		Expect(ok).To(BeTrue())
		Expect(desc).To(Equal("container 'main' failed: boom (pod: worker-0)"))
	})

	It("derives a specific Compute issue from a terminated container exit code", func() {
		group, desc, ok := ComputeRuntimeIssue(failedJob(), []corev1.Pod{podTerminated("main", "Error", "", 127)}, nil)
		Expect(ok).To(BeTrue())
		Expect(group).To(Equal("Compute"))
		Expect(desc).To(Equal("container 'main' exited with exit code 127 (pod: worker-0)"))
	})

	It("reports an OOMKilled container", func() {
		_, desc, _ := ComputeRuntimeIssue(failedJob(), []corev1.Pod{podTerminated("main", "OOMKilled", "", 137)}, nil)
		Expect(desc).To(Equal("container 'main' was OOMKilled (pod: worker-0)"))
	})

	It("reports a container waiting to start", func() {
		_, desc, _ := ComputeRuntimeIssue(failedJob(), []corev1.Pod{podWaiting("main", "RunContainerError", "exec: bad binary")}, nil)
		Expect(desc).To(Equal("container 'main' failed to start: exec: bad binary (pod: worker-0)"))
	})

	It("triggers on a wlr Failed phase even without a job failed condition", func() {
		_, desc, ok := ComputeRuntimeIssue(jobWith(nil, 0, nil, nil),
			[]corev1.Pod{podTerminated("main", "OOMKilled", "", 137)},
			wlrWith(dromosv1alpha1.WorkloadRunnerPhaseFailed))
		Expect(ok).To(BeTrue())
		Expect(desc).To(Equal("container 'main' was OOMKilled (pod: worker-0)"))
	})

	It("is ok with an empty description when no pod detail is available (caller supplies fallback)", func() {
		group, desc, ok := ComputeRuntimeIssue(failedJob(), nil, nil)
		Expect(ok).To(BeTrue())
		Expect(group).To(Equal("Compute"))
		Expect(desc).To(BeEmpty())
	})

	It("omits the pod suffix when the failing pod has no name", func() {
		unnamed := corev1.Pod{Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:  "main",
				State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"}},
			}},
		}}
		_, desc, _ := ComputeRuntimeIssue(failedJob(), []corev1.Pod{unnamed}, nil)
		Expect(desc).To(Equal("container 'main' exited with exit code 1"))
	})

	It("reports an ImagePullBackOff startup failure", func() {
		_, desc, _ := ComputeRuntimeIssue(failedJob(), []corev1.Pod{podWaiting("main", "ImagePullBackOff", "Back-off pulling image")}, nil)
		Expect(desc).To(Equal("container 'main' failed to start: Back-off pulling image (pod: worker-0)"))
	})

	It("reports ErrImagePull via the reason when no message is set", func() {
		_, desc, _ := ComputeRuntimeIssue(failedJob(), []corev1.Pod{podWaiting("main", "ErrImagePull", "")}, nil)
		Expect(desc).To(Equal("container 'main' failed to start (ErrImagePull) (pod: worker-0)"))
	})

	It("picks the same (name-sorted) pod regardless of list order", func() {
		podA := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "a-pod"},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
				Name:  "main",
				State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"}},
			}}},
		}
		podZ := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "z-pod"},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
				Name:  "main",
				State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 2, Reason: "Error"}},
			}}},
		}
		_, desc1, _ := ComputeRuntimeIssue(failedJob(), []corev1.Pod{podZ, podA}, nil)
		_, desc2, _ := ComputeRuntimeIssue(failedJob(), []corev1.Pod{podA, podZ}, nil)
		Expect(desc1).To(Equal(desc2))
		Expect(desc1).To(Equal("container 'main' exited with exit code 1 (pod: a-pod)"))
	})
})
