/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package events

import (
	"fmt"
	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("Events", func() {
	const (
		eventType   = corev1.EventTypeNormal
		eventReason = "Reason"

		msg    = "This is event"
		format = "This is event with %s and %s"
		arg1   = "arg1"
		arg2   = "arg2"
	)

	var (
		fakeRecorder     *record.FakeRecorder
		workloadObj      client.Object
		workloadRecorder Recorder
	)

	BeforeEach(func() {
		fakeRecorder = record.NewFakeRecorder(200)
		workloadObj = &twv2.TrainingWorkload{}
		workloadRecorder = NewWorkloadRecorder(workloadObj, fakeRecorder)
	})

	Context("Event", func() {
		It("sent event is correct", func() {
			workloadRecorder.Event(eventType, eventReason, msg)

			Eventually(fakeRecorder.Events).WithTimeout(3 * time.Second).
				Should(Receive(ContainSubstring(msg)))
		})
	})

	Context("Eventf", func() {
		It("sent event is correct - with args", func() {
			workloadRecorder.Eventf(eventType, eventReason, format, arg1, arg2)

			Eventually(fakeRecorder.Events).WithTimeout(3 * time.Second).
				Should(Receive(ContainSubstring(fmt.Sprintf(format, arg1, arg2))))
		})
		It("sent event is correct - no args", func() {
			workloadRecorder.Eventf(eventType, eventReason, msg)

			Eventually(fakeRecorder.Events).WithTimeout(3 * time.Second).
				Should(Receive(ContainSubstring(msg)))
		})
	})
})
