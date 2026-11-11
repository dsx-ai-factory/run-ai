/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
)

var _ = Describe("SimulateNodePoolFairShare", func() {
	It("returns 501 until the simulation is implemented", func() {
		response, err := services.OrgUnitApiService{}.SimulateNodePoolFairShare(context.Background(), "1", "department",
			api.FairShareSimulationRequest{}, "", "", "", 0)
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusNotImplemented))
	})
})
