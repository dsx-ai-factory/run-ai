/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
)

var _ = Describe("Success Messages", func() {
	const (
		testWorkloadName = "test-workload"
		inferenceStr     = "inference"
		trainingStr      = "training"
		workspaceStr     = "workspace"
		distributedStr   = "distributed"
	)

	Describe("SuccessDeleteMessage", func() {
		It("should return the correct training delete success message", func() {
			testWorkloadType := api.Training
			expectedMessage := fmt.Sprintf(consts.SuccessDelete, trainingStr, testWorkloadName, consts.FollowingWorkspace)
			actualMessage := SuccessDeleteMessage(testWorkloadName, testWorkloadType)
			Expect(actualMessage).To(Equal(expectedMessage))
		})

		It("should return the correct inference delete success message", func() {
			testWorkloadType := api.Inference
			expectedMessage := fmt.Sprintf(consts.SuccessDelete, inferenceStr, testWorkloadName, consts.FollowingInference)
			actualMessage := SuccessInferenceDeleteMessage(testWorkloadName, testWorkloadType)
			Expect(actualMessage).To(Equal(expectedMessage))
		})

		It("should return the correct workspace delete success message", func() {
			testWorkloadType := api.Workspace
			expectedMessage := fmt.Sprintf(consts.SuccessDelete, workspaceStr, testWorkloadName, consts.FollowingWorkspace)
			actualMessage := SuccessDeleteMessage(testWorkloadName, testWorkloadType)
			Expect(actualMessage).To(Equal(expectedMessage))
		})

		It("should return the correct distributed delete success message", func() {
			testWorkloadType := api.Distributed
			expectedMessage := fmt.Sprintf(consts.SuccessDelete, distributedStr, testWorkloadName, consts.FollowingWorkspace)
			actualMessage := SuccessDeleteMessage(testWorkloadName, testWorkloadType)
			Expect(actualMessage).To(Equal(expectedMessage))
		})
	})

	Describe("SuccessResumeMessage", func() {
		It("should return the correct training resume success message", func() {
			testWorkloadType := api.Training
			expectedMessage := fmt.Sprintf(consts.SuccessResume, trainingStr, testWorkloadName, consts.FollowingWorkspace)
			actualMessage := SuccessResumeMessage(testWorkloadName, testWorkloadType)
			Expect(actualMessage).To(Equal(expectedMessage))
		})

		It("should return the correct workspace resume success message", func() {
			testWorkloadType := api.Workspace
			expectedMessage := fmt.Sprintf(consts.SuccessResume, workspaceStr, testWorkloadName, consts.FollowingWorkspace)
			actualMessage := SuccessResumeMessage(testWorkloadName, testWorkloadType)
			Expect(actualMessage).To(Equal(expectedMessage))
		})

		It("should return the correct distributed resume success message", func() {
			testWorkloadType := api.Distributed
			expectedMessage := fmt.Sprintf(consts.SuccessResume, distributedStr, testWorkloadName, consts.FollowingWorkspace)
			actualMessage := SuccessResumeMessage(testWorkloadName, testWorkloadType)
			Expect(actualMessage).To(Equal(expectedMessage))
		})
	})

	Describe("SuccessSuspendMessage", func() {
		It("should return the correct training suspend success message", func() {
			testWorkloadType := api.Training
			expectedMessage := fmt.Sprintf(consts.SuccessSuspend, trainingStr, testWorkloadName, consts.FollowingWorkspace)
			actualMessage := SuccessSuspendMessage(testWorkloadName, testWorkloadType)
			Expect(actualMessage).To(Equal(expectedMessage))
		})

		It("should return the correct workspace suspend success message", func() {
			testWorkloadType := api.Workspace
			expectedMessage := fmt.Sprintf(consts.SuccessSuspend, workspaceStr, testWorkloadName, consts.FollowingWorkspace)
			actualMessage := SuccessSuspendMessage(testWorkloadName, testWorkloadType)
			Expect(actualMessage).To(Equal(expectedMessage))
		})

		It("should return the correct distributed suspend success message", func() {
			testWorkloadType := api.Distributed
			expectedMessage := fmt.Sprintf(consts.SuccessSuspend, distributedStr, testWorkloadName, consts.FollowingWorkspace)
			actualMessage := SuccessSuspendMessage(testWorkloadName, testWorkloadType)
			Expect(actualMessage).To(Equal(expectedMessage))
		})
	})
})
