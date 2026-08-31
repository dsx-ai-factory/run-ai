/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composite_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders/composite"
	buildersMock "github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders/mocks"
	"github.com/stretchr/testify/mock"
)

const (
	timeout  = time.Second * 5
	interval = time.Millisecond * 500
)

func TestBuilders(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Builders Suite")
}

var _ = Describe("Builders", Ordered, func() {
	Context("Parallel construction", Ordered, func() {
		It("Happy flow", func() {
			builder1 := getSuccessfulBuilderMock(1)
			builder2 := getSuccessfulBuilderMock(2)
			compositeBuilder := composite.NewCompositeBuilder([]builders.Interface{builder1, builder2}, composite.Parallel)
			err := compositeBuilder.Construct(context.Background(), getTenantInfo(), nil)
			Expect(err).To(BeNil())
		})

		It("One builder completes with err", func() {
			builder1 := getSuccessfulBuilderMock(1)
			builder2 := getErrBuilderMock(2, "builder2 failed")
			compositeBuilder := composite.NewCompositeBuilder([]builders.Interface{builder1, builder2}, composite.Parallel)
			err := compositeBuilder.Construct(context.Background(), getTenantInfo(), nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("builder2 failed"))
		})
	})
	Context("Serial construction", Ordered, func() {
		It("Happy flow", func() {
			builder1 := getSuccessfulBuilderMock(1)
			builder2 := getSuccessfulBuilderMock(2)
			compositeBuilder := composite.NewCompositeBuilder([]builders.Interface{builder1, builder2}, composite.Serial)

			err := compositeBuilder.Construct(context.Background(), getTenantInfo(), nil)
			Expect(err).To(BeNil())
		})

		It("One builder completes with err", func() {
			builder1 := getSuccessfulBuilderMock(1)
			builder2 := getErrBuilderMock(2, "builder2 failed")
			compositeBuilder := composite.NewCompositeBuilder([]builders.Interface{builder1, builder2}, composite.Serial)

			err := compositeBuilder.Construct(context.Background(), getTenantInfo(), nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("builder2 failed"))
		})
	})
})

func getTenantInfo() *builders.TenantInfo {
	return &builders.TenantInfo{
		Name: "name",
		Id:   1,
	}
}

func getSuccessfulBuilderMock(id int) builders.Interface {
	builder := buildersMock.NewInterface(GinkgoT())
	builder.On("Construct", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	builder.On("Name").Return(fmt.Sprintf("builder%d", id))
	return builder
}

func getErrBuilderMock(id int, errString string) builders.Interface {
	builder := buildersMock.NewInterface(GinkgoT())
	builder.On("Construct", mock.Anything, mock.Anything, mock.Anything).Return(errors.New(errString))
	builder.On("Name").Return(fmt.Sprintf("builder%d", id))
	return builder
}
