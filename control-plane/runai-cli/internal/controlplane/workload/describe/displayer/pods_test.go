/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
)

var _ = Describe("PodBuilder", func() {
	It("should build a Pod correctly", func() {
		completedAt := time.Now()
		pod := displayer.NewPodBuilder().
			Name("pod1").
			NodeName("node1").
			CompletedAt(completedAt).
			CurrentNodePool("pool1").
			K8sPhase("Running").
			AddContainer(displayer.Container{Name: "container1", Image: "image1"}).
			Build()

		Expect(pod.Name).To(Equal("pod1"))
		Expect(*pod.NodeName).To(Equal("node1"))
		Expect(*pod.CompletedAt).To(Equal(completedAt))
		Expect(*pod.CurrentNodePool).To(Equal("pool1"))
		Expect(pod.K8sPhase).To(Equal("Running"))
		Expect(pod.Containers).To(HaveLen(1))
		Expect(pod.Containers[0].Name).To(Equal("container1"))
		Expect(pod.Containers[0].Image).To(Equal("image1"))
	})

	It("should set role", func() {
		pod := displayer.NewPodBuilder().
			Name("pod1").
			Role("Leader").
			Build()

		Expect(*pod.Role).To(Equal("Leader"))
	})
})

var _ = Describe("PodsBuilder", func() {
	It("should build a collection of pods", func() {
		pod1 := displayer.NewPodBuilder().Name("pod1").Build()
		pod2 := displayer.NewPodBuilder().Name("pod2").Build()

		pods := displayer.NewPodsBuilder().
			Add(pod1).
			Add(pod2).
			Build()

		Expect(pods).To(HaveLen(2))
		Expect(pods[0].Name).To(Equal("pod1"))
		Expect(pods[1].Name).To(Equal("pod2"))
	})

	It("should set entire collection with Set method", func() {
		pod1 := displayer.NewPodBuilder().Name("pod1").Build()
		pod2 := displayer.NewPodBuilder().Name("pod2").Build()

		pods := displayer.NewPodsBuilder().
			Set([]displayer.Pod{pod1}).
			Add(pod2).
			Build()

		Expect(pods).To(HaveLen(2))
	})
})
