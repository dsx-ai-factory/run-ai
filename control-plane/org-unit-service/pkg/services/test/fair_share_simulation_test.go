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
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	auth_mocks "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
)

var _ = Describe("SimulateNodePoolFairShare", func() {
	serviceWith := func(read map[authorization_api.ResourceType]authorization_api.PermittedScopes) services.OrgUnitApiService {
		authMock := &auth_mocks.IAuthorizationClient{}
		for resourceType, scopes := range read {
			authMock.On("GetPermittedScopes", mock.Anything, resourceType, mock.Anything).
				Return(&authorization_api.PermittedScopesActions{Read: scopes}, 0, nil)
		}
		return services.OrgUnitApiService{AuthorizationClient: authMock}
	}
	readEverything := serviceWith(map[authorization_api.ResourceType]authorization_api.PermittedScopes{
		authorization_api.RESOURCETYPE_DEPARTMENT: {System: ptr.To(true)},
		authorization_api.RESOURCETYPE_PROJECT:    {System: ptr.To(true)},
	})

	It("returns 501 until the simulation is implemented", func() {
		response, err := readEverything.SimulateNodePoolFairShare(context.Background(), "1",
			api.FairShareSimulationRequest{OrgUnitType: "department"})
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusNotImplemented))
	})

	DescribeTable("requires read on the simulated org unit type",
		func(orgUnitType string, read map[authorization_api.ResourceType]authorization_api.PermittedScopes, expectedCode int) {
			response, err := serviceWith(read).SimulateNodePoolFairShare(context.Background(), "1",
				api.FairShareSimulationRequest{OrgUnitType: orgUnitType})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Code).To(Equal(expectedCode))
		},
		Entry("departments with department read", "department",
			map[authorization_api.ResourceType]authorization_api.PermittedScopes{
				authorization_api.RESOURCETYPE_DEPARTMENT: {Departments: []string{"1"}}}, http.StatusNotImplemented),
		Entry("departments with project read only", "department",
			map[authorization_api.ResourceType]authorization_api.PermittedScopes{
				authorization_api.RESOURCETYPE_DEPARTMENT: {}}, http.StatusForbidden),
		Entry("projects with project read", "project",
			map[authorization_api.ResourceType]authorization_api.PermittedScopes{
				authorization_api.RESOURCETYPE_PROJECT: {Projects: []string{"1"}}}, http.StatusNotImplemented),
		Entry("projects with department read only", "project",
			map[authorization_api.ResourceType]authorization_api.PermittedScopes{
				authorization_api.RESOURCETYPE_PROJECT: {}}, http.StatusForbidden),
	)

	// The generated controller rejects all-zero required objects unless they're pointers, so these requests must reach
	// the service.
	DescribeTable("the controller's schema checks",
		func(body string, reachesService bool) {
			routeContext := chi.NewRouteContext()
			routeContext.URLParams.Add("nodepoolId", "1")
			request := httptest.NewRequest(http.MethodPost, "/api/v1/org-unit/node-pools/1/fair-share-simulation",
				strings.NewReader(body)).WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, routeContext))
			recorder := httptest.NewRecorder()
			api.NewOrgUnitAPIController(readEverything).SimulateNodePoolFairShare(recorder, request)
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
				`"orgUnitType":"project","parent":{"id":"2"},"orgUnits":[{"id":"1","demand":null,"rank":null}]}`, true),
		Entry("a null node pool", `{"nodePool":null,"orgUnitType":"department","orgUnits":[{"id":"1"}]}`, false),
		Entry("an unknown rank",
			`{"nodePool":{"capacity":{"gpu":8},"schedulingCapabilities":{"quota":true,"ranks":true,"adjustableWeights":true,"limits":true}},`+
				`"orgUnitType":"department","orgUnits":[{"id":"1","rank":"Sometimes"}]}`, false),
		Entry("a negative capacity",
			`{"nodePool":{"capacity":{"gpu":-1},"schedulingCapabilities":{"quota":true,"ranks":true,"adjustableWeights":true,"limits":true}},`+
				`"orgUnitType":"department","orgUnits":[{"id":"1"}]}`, false),
	)
})
