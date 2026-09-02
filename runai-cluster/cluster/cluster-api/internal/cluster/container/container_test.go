/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package container_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/cluster/container"
	v1 "k8s.io/api/core/v1"
)

const (
	ContainerName       = "containerName"
	ContainerNotTheName = "containerNotTheName"
)

var (
	EmptyState = []v1.ContainerStatus{{
		Name:  ContainerName,
		State: v1.ContainerState{},
	}}

	WaitingState = []v1.ContainerStatus{{
		Name: ContainerName,
		State: v1.ContainerState{
			Waiting: &v1.ContainerStateWaiting{
				Reason:  "this is how we party",
				Message: "",
			},
		},
	}}

	RunningState = []v1.ContainerStatus{{
		Name: ContainerName,
		State: v1.ContainerState{
			Running: &v1.ContainerStateRunning{},
		},
	}}

	TerminatedState = []v1.ContainerStatus{{
		Name: ContainerName,
		State: v1.ContainerState{
			Terminated: &v1.ContainerStateTerminated{ExitCode: 123},
		},
	}}
)

func TestLogs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Container Suite")
}

var _ = Describe("ValidateContainer", func() {
	DescribeTable("Shared Tests", validateSameResult,
		Entry("nil status list", nil, "", api.NewNotFoundError("container does not exist")),
		Entry("empty status list", []v1.ContainerStatus{}, "", api.NewNotFoundError("container does not exist")),
		Entry("status list with empty state, not the container name", EmptyState, ContainerNotTheName, api.NewNotFoundError("container does not exist")),
		Entry("status list with empty state, with container name", EmptyState, ContainerName, api.NewNotReadyError("container is not ready")),
		Entry("status list with waiting state, with container name", WaitingState, ContainerName, api.NewNotReadyError("container is waiting, reason: this is how we party")),
	)

	Context("IsReady", func() {
		It("container status list running", func() {
			err := container.ValidateContainerIsReady(RunningState, ContainerName)
			Expect(err).ToNot(HaveOccurred())
		})

		It("container status list terminated", func() {
			err := container.ValidateContainerIsReady(TerminatedState, ContainerName)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(api.NewNotReadyError("container is terminated, exit code: 123")))
		})
	})

	Context("IsExist", func() {
		It("container status list running", func() {
			err := container.ValidateContainerExist(RunningState, ContainerName)
			Expect(err).ToNot(HaveOccurred())
		})

		It("container status list terminated", func() {
			err := container.ValidateContainerExist(TerminatedState, ContainerName)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

func validateSameResult(containerStatuses []v1.ContainerStatus, name string, expectedErr error) {
	readyErr := container.ValidateContainerIsReady(containerStatuses, name)
	existErr := container.ValidateContainerExist(containerStatuses, name)
	Expect(readyErr).To(Equal(existErr))
	Expect(readyErr).To(Equal(expectedErr))
}
