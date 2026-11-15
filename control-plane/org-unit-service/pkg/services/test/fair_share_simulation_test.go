/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
)

var _ = Describe("SimulateNodePoolFairShare", func() {
	It("returns 501 until the simulation is implemented", func() {
		response, err := services.OrgUnitApiService{}.SimulateNodePoolFairShare(context.Background(), "1", api.FairShareSimulationRequest{})
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusNotImplemented))
	})

	// The generated controller rejects all-zero required objects unless they're pointers, so these requests must reach
	// the service.
	DescribeTable("the controller's schema checks",
		func(body string, reachesService bool) {
			routeContext := chi.NewRouteContext()
			routeContext.URLParams.Add("nodepoolId", "1")
			request := httptest.NewRequest(http.MethodPost, "/api/v1/org-unit/node-pools/1/fair-share-simulation",
				strings.NewReader(body)).WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, routeContext))
			recorder := httptest.NewRecorder()
			api.NewOrgUnitAPIController(services.OrgUnitApiService{}).SimulateNodePoolFairShare(recorder, request)
			if reachesService {
				Expect(recorder.Code).To(Equal(http.StatusNotImplemented), recorder.Body.String())
			} else {
				Expect(recorder.Code).To(BeNumerically(">=", http.StatusBadRequest))
				Expect(recorder.Code).NotTo(Equal(http.StatusNotImplemented))
			}
		},
		Entry("zero capacity with every capability off",
			`{"nodePool":{"capacity":{"gpu":0},"schedulingCapabilities":{"quota":false,"ranks":false,"adjustableWeights":false,"limits":false}},`+
				`"orgUnitType":"department","orgUnits":[{"id":"1"}]}`, true),
		Entry("null and omitted edits",
			`{"nodePool":{"capacity":{"gpu":8},"schedulingCapabilities":{"quota":true,"ranks":true,"adjustableWeights":true,"limits":true}},`+
				`"orgUnitType":"project","parent":{"id":"2","gpu":{"limit":null}},"orgUnits":[{"id":"1","demand":null,"rank":null}]}`, true),
		Entry("a null node pool", `{"nodePool":null,"orgUnitType":"department","orgUnits":[{"id":"1"}]}`, false),
		Entry("an unknown rank",
			`{"nodePool":{"capacity":{"gpu":8},"schedulingCapabilities":{"quota":true,"ranks":true,"adjustableWeights":true,"limits":true}},`+
				`"orgUnitType":"department","orgUnits":[{"id":"1","rank":"Sometimes"}]}`, false),
		Entry("a negative capacity",
			`{"nodePool":{"capacity":{"gpu":-1},"schedulingCapabilities":{"quota":true,"ranks":true,"adjustableWeights":true,"limits":true}},`+
				`"orgUnitType":"department","orgUnits":[{"id":"1"}]}`, false),
	)
})
