// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rightsizing

import (
	"context"
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRightSizing(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Recommendation suite")
}

var _ = Describe("recommendation inventory errors", func() {
	DescribeTable("preserves the inventory response status",
		func(status int) {
			upstream := api.NewError(status, "Cluster inventory lookup failed", nil)
			inventory := inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, *api.Error) {
				return nil, upstream
			})
			response, err := New(inventory, nil, 90).Recommend(context.Background(), validRequest())
			Expect(err).To(HaveOccurred())
			Expect(err.Code).To(Equal(int32(status)))
			Expect(response.Results).To(BeEmpty())
		},
		Entry("missing cluster", http.StatusNotFound),
		Entry("authentication failure", http.StatusUnauthorized),
		Entry("authorization failure", http.StatusForbidden),
		Entry("dependency failure", http.StatusServiceUnavailable),
	)
})
