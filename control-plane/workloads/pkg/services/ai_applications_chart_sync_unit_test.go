/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/google/uuid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("applicationSyncUpdateToSyncItem", func() {
	const chartID = "3fa85f64-5717-4562-b3fc-2c963f66afa6"

	It("resolves an uploaded chartId to a download URL without mutating the stored spec", func() {
		storedSpec := &api.AiApplicationSpec{
			ChartRef: api.AiApplicationSpecChartRef{ChartId: chartID},
		}
		update := models.ApplicationSyncUpdate{ID: uuid.New(), Name: "app", Spec: storedSpec}

		item, err := applicationSyncUpdateToSyncItem(update, testChartURLBuilder())
		Expect(err).NotTo(HaveOccurred())

		Expect(item.Spec.ChartRef.ChartId).To(BeEmpty())
		Expect(item.Spec.ChartRef.ChartURL).NotTo(BeEmpty())
		Expect(item.RequiresControlPlaneAuth).To(BeTrue())

		Expect(storedSpec.ChartRef.ChartId).To(Equal(chartID))
		Expect(storedSpec.ChartRef.ChartURL).To(BeEmpty())
	})

	It("passes a user-provided chartURL through without the control-plane auth flag", func() {
		storedSpec := &api.AiApplicationSpec{
			ChartRef: api.AiApplicationSpecChartRef{ChartURL: "https://example.com/chart.tgz"},
		}
		update := models.ApplicationSyncUpdate{ID: uuid.New(), Name: "app", Spec: storedSpec}

		item, err := applicationSyncUpdateToSyncItem(update, testChartURLBuilder())
		Expect(err).NotTo(HaveOccurred())
		Expect(item.Spec.ChartRef.ChartURL).To(Equal("https://example.com/chart.tgz"))
		Expect(item.RequiresControlPlaneAuth).To(BeFalse())
	})

	It("fails when a chartId is present but the base URL is unconfigured", func() {
		storedSpec := &api.AiApplicationSpec{
			ChartRef: api.AiApplicationSpecChartRef{ChartId: chartID},
		}
		update := models.ApplicationSyncUpdate{ID: uuid.New(), Name: "app", Spec: storedSpec}

		_, err := applicationSyncUpdateToSyncItem(update, &chartDownloadURLBuilder{})
		Expect(err).To(HaveOccurred())
	})
})
