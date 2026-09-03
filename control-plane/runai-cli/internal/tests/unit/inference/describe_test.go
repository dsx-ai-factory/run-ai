/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_test

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Inference Describe Test", Ordered, func() {
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
		describeMocker, err = clients.NewDescribeMocker(mockCtrl, clients.InferenceFixture)
		Expect(err).NotTo(HaveOccurred())
		Expect(describeMocker).NotTo(BeNil())
	})

	AfterAll(func() {
		mockCtrl.Finish()
	})

	Context("Describe Inference Workload - Table Output", func() {
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
			err = clients.CompareWithExpectedOutput(clients.InferenceFixture, "table", actualOutput)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Describe Inference Workload - JSON Output", func() {
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
			err = clients.CompareWithExpectedOutput(clients.InferenceFixture, "json", actualOutput)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
