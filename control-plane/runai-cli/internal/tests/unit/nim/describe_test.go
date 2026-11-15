/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nim_test

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	nimclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NIM Describe Test", Ordered, func() {
	var (
		mockCtrl       *gomock.Controller
		describeMocker *clients.DescribeMocker
		ctx            context.Context
	)

	const (
		podLimit   int32 = 10
		eventLimit int32 = 50
	)

	BeforeAll(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctx = context.Background()

		var err error
		describeMocker, err = clients.NewDescribeMocker(mockCtrl, clients.NimFixture)
		Expect(err).NotTo(HaveOccurred())
		Expect(describeMocker).NotTo(BeNil())
	})

	AfterAll(func() {
		mockCtrl.Finish()
	})

	It("describes the new GPU fields from the NIM response", func() {
		fixtures := describeMocker.GetFixtures()
		original := fixtures.NimSpec.Spec.Compute
		defer func() { fixtures.NimSpec.Spec.Compute = original }()
		compute := nimclient.NewNimServiceSpecAllOfCompute()
		compute.SetGpuMemoryPortionRequest(0.25)
		compute.SetGpuMemoryPortionLimit(0.5)
		compute.SetGpuComputeMode(nimclient.GPUCOMPUTEMODE_SM_SHARING)
		fixtures.NimSpec.Spec.SetCompute(*compute)
		result, err := describeMocker.CreateDescribeService().GetWorkload(ctx, &api.Workload{ID: fixtures.WorkloadDetailed.Id}, podLimit, eventLimit)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Nim.GpuPortionRequest).NotTo(BeNil())
		Expect(*result.Nim.GpuPortionRequest).To(Equal(0.25))
		Expect(*result.Nim.GpuPortionLimit).To(Equal(0.5))
		output, err := clients.CaptureDescribeTableOutput(result)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(MatchRegexp(`GPU Compute Mode:\s+sm-sharing`))
	})

	Context("Describe NIM Workload - Table Output", func() {
		It("should match expected table output", func() {
			fixtures := describeMocker.GetFixtures()
			Expect(fixtures.WorkloadDetailed).NotTo(BeNil())

			reqWorkload := &api.Workload{
				ID: fixtures.WorkloadDetailed.Id,
			}

			service := describeMocker.CreateDescribeService()
			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())

			// Get actual output
			actualOutput, err := clients.CaptureDescribeTableOutput(result)
			Expect(err).NotTo(HaveOccurred())

			// Compare with expected output
			err = clients.CompareWithExpectedOutput(clients.NimFixture, "table", actualOutput)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Describe NIM Workload - JSON Output", func() {
		It("should match expected JSON output", func() {
			fixtures := describeMocker.GetFixtures()
			Expect(fixtures.WorkloadDetailed).NotTo(BeNil())

			reqWorkload := &api.Workload{
				ID: fixtures.WorkloadDetailed.Id,
			}

			service := describeMocker.CreateDescribeService()
			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())

			// Get actual output
			actualOutput, err := clients.CaptureDescribeJSONOutput(result)
			Expect(err).NotTo(HaveOccurred())

			// Compare with expected output
			err = clients.CompareWithExpectedOutput(clients.NimFixture, "json", actualOutput)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
