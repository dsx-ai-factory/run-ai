/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runaiPodPhase "github.com/run-ai/runai/runai-cluster/cluster/common/utils/pod"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("TestRunAIPodPhaseAndMessage", func() {
	var (
		runPodPhaseHandler runaiPodPhase.RunaiPodPhaseHandler
	)

	BeforeEach(func() {
		runPodPhaseHandler = runaiPodPhase.NewRunaiPodPhaseHandler(&runaiPodPhase.RunaiPhaseCalculator{})
	})

	Context("Pod to RunAI phase", func() {
		It("Phase and reason check", func() {
			pod := generatePod(v1.PodPending, "pending reason", nil)
			runaiPod := runPodPhaseHandler.Calculate(pod)
			Expect(runaiPod.GetPhase()).To(Equal(runaiPodPhase.RunaiPhasePending))
			Expect(runaiPod.GetPhaseMessage()).To(Equal("pending reason"))
			Expect(runaiPodPhase.GetPhaseStr(runaiPodPhase.RunaiPhasePending)).To(Equal(runaiPodPhase.RunaiPhasePendingStr))

			conditions := []v1.PodCondition{{
				Type:    v1.PodScheduled,
				Status:  v1.ConditionFalse,
				Reason:  "schedule issue",
				Message: "schedule message",
			}}

			pod = generatePod(v1.PodPending, "pending reason", &conditions)
			runaiPod = runPodPhaseHandler.Calculate(pod)
			Expect(runaiPod.GetPhase()).To(Equal(runaiPodPhase.RunaiPhasePending))
			Expect(runaiPod.GetPhaseMessage()).To(Equal("pending reason"))
			Expect(runaiPodPhase.GetPhaseStr(runaiPodPhase.RunaiPhasePending)).To(Equal(runaiPodPhase.RunaiPhasePendingStr))

			conditions = []v1.PodCondition{{
				Type:    v1.PodScheduled,
				Status:  v1.ConditionTrue,
				Reason:  "reason",
				Message: "message",
			}}

			pod = generatePod(v1.PodPending, "pending reason", &conditions)
			runaiPod = runPodPhaseHandler.Calculate(pod)
			Expect(runaiPod.GetPhase()).To(Equal(runaiPodPhase.RunaiPhaseInitializing))
			Expect(runaiPod.GetPhaseMessage()).To(Equal("pending reason"))
			Expect(runaiPodPhase.GetPhaseStr(runaiPodPhase.RunaiPhaseInitializing)).To(Equal(runaiPodPhase.RunaiPhaseInitializingStr))

			pod = generatePod(v1.PodFailed, "failed reason", nil)
			runaiPod = runPodPhaseHandler.Calculate(pod)
			Expect(runaiPod.GetPhase()).To(Equal(runaiPodPhase.RunaiPhaseFailed))
			Expect(runaiPod.GetPhaseMessage()).To(Equal("failed reason"))
			Expect(runaiPodPhase.GetPhaseStr(runaiPodPhase.RunaiPhaseFailed)).To(Equal(runaiPodPhase.RunaiPhaseFailedStr))

			pod = generatePod(v1.PodSucceeded, "success reason", nil)
			runaiPod = runPodPhaseHandler.Calculate(pod)
			Expect(runaiPod.GetPhase()).To(Equal(runaiPodPhase.RunaiPhaseSuccess))
			Expect(runaiPod.GetPhaseMessage()).To(Equal("success reason"))
			Expect(runaiPodPhase.GetPhaseStr(runaiPodPhase.RunaiPhaseSuccess)).To(Equal(runaiPodPhase.RunaiPhaseSuccessStr))
		})

		It("Phase running no condition", func() {
			pod := generatePod(v1.PodRunning, "running reason", nil)
			runaiPod := runPodPhaseHandler.Calculate(pod)
			Expect(runaiPod.GetPhase()).To(Equal(runaiPodPhase.RunaiPhasePending))
			Expect(runaiPod.GetPhaseMessage()).To(Equal(""))
		})
		It("Phase running ready condition true", func() {
			conditions := []v1.PodCondition{{
				Type:    v1.PodReady,
				Status:  v1.ConditionTrue,
				Reason:  "reason",
				Message: "message",
			}}
			pod := generatePod(v1.PodRunning, "running reason", &conditions)
			runaiPod := runPodPhaseHandler.Calculate(pod)
			Expect(runaiPod.GetPhase()).To(Equal(runaiPodPhase.RunaiPhaseReady))
			Expect(runaiPod.GetPhaseMessage()).To(Equal(""))
		})
		It("Phase running ready condition false", func() {
			conditions := []v1.PodCondition{{
				Type:    v1.PodReady,
				Status:  v1.ConditionFalse,
				Reason:  "ReasonPodReady",
				Message: "MessagePodReady",
			}}
			pod := generatePod(v1.PodRunning, "running reason", &conditions)
			runaiPod := runPodPhaseHandler.Calculate(pod)
			Expect(runaiPod.GetPhase()).To(Equal(runaiPodPhase.RunaiPhaseInitializing))
			Expect(runaiPod.GetPhaseMessage()).To(Equal("ReasonPodReady"))
		})
		It("Phase running ready condition false", func() {
			conditions := []v1.PodCondition{{
				Type:    v1.PodReady,
				Status:  v1.ConditionFalse,
				Reason:  "ReasonPodReady",
				Message: "MessagePodReady",
			}}
			pod := generatePod(v1.PodRunning, "running reason", &conditions)
			runaiPod := runPodPhaseHandler.Calculate(pod)
			Expect(runaiPod.GetPhase()).To(Equal(runaiPodPhase.RunaiPhaseInitializing))
			Expect(runaiPod.GetPhaseMessage()).To(Equal("ReasonPodReady"))
		})
		It("Phase running init condition false", func() {
			conditions := []v1.PodCondition{
				{
					Type:    v1.PodScheduled,
					Status:  v1.ConditionFalse,
					Reason:  "ReasonPodScheduled",
					Message: "MessagePodScheduled",
				},
			}
			pod := generatePod(v1.PodRunning, "running reason", &conditions)
			runaiPod := runPodPhaseHandler.Calculate(pod)
			Expect(runaiPod.GetPhase()).To(Equal(runaiPodPhase.RunaiPhaseInitializing))
			Expect(runaiPod.GetPhaseMessage()).To(Equal("ReasonPodScheduled"))
		})
		It("Phase running container ready condition true", func() {
			conditions := []v1.PodCondition{
				{
					Type:    v1.ContainersReady,
					Status:  v1.ConditionTrue,
					Reason:  "ReasonPodContainersReady",
					Message: "MessagePoContainersReady",
				},
			}
			pod := generatePod(v1.PodRunning, "running reason", &conditions)
			runaiPod := runPodPhaseHandler.Calculate(pod)
			Expect(runaiPod.GetPhase()).To(Equal(runaiPodPhase.RunaiPhaseInitializing))
			Expect(runaiPod.GetPhaseMessage()).To(Equal("ReasonPodContainersReady"))
		})
		It("Phase running container ready condition false", func() {
			conditions := []v1.PodCondition{
				{
					Type:    v1.ContainersReady,
					Status:  v1.ConditionFalse,
					Reason:  "ReasonPodContainersReady",
					Message: "MessagePoContainersReady",
				},
			}
			pod := generatePod(v1.PodRunning, "running reason", &conditions)
			runaiPod := runPodPhaseHandler.Calculate(pod)
			Expect(runaiPod.GetPhase()).To(Equal(runaiPodPhase.RunaiPhaseInitializing))
			Expect(runaiPod.GetPhaseMessage()).To(Equal("ReasonPodContainersReady"))
		})
	})
})
