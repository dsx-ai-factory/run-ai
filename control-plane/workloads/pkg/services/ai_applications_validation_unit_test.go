/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("validateApplicationSpec chart sources", func() {
	const validUUID = "3fa85f64-5717-4562-b3fc-2c963f66afa6"

	DescribeTable("accepts exactly one chart source and rejects the rest",
		func(chartRef api.AiApplicationSpecChartRef, wantError bool) {
			err := validateApplicationSpec(&api.AiApplicationSpec{ChartRef: chartRef})
			if wantError {
				Expect(err).NotTo(BeNil())
				Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
				return
			}
			Expect(err).To(BeNil())
		},
		Entry("chartId only", api.AiApplicationSpecChartRef{ChartId: validUUID}, false),
		Entry("chartURL only", api.AiApplicationSpecChartRef{ChartURL: "https://example.com/chart.tgz"}, false),
		Entry("repo triplet", api.AiApplicationSpecChartRef{Repo: "https://charts.example.com", ChartName: "redis", ChartVersion: "1.2.3"}, false),
		Entry("chartId and chartURL together", api.AiApplicationSpecChartRef{ChartId: validUUID, ChartURL: "https://example.com/chart.tgz"}, true),
		Entry("chartId and repo together", api.AiApplicationSpecChartRef{ChartId: validUUID, Repo: "https://charts.example.com", ChartName: "redis", ChartVersion: "1.2.3"}, true),
		Entry("chartId that is not a uuid", api.AiApplicationSpecChartRef{ChartId: "not-a-uuid"}, true),
		Entry("no source at all", api.AiApplicationSpecChartRef{}, true),
		Entry("partial repo triplet", api.AiApplicationSpecChartRef{Repo: "https://charts.example.com", ChartName: "redis"}, true),
	)
})
