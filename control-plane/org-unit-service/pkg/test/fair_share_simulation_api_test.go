/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	auth_mocks "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
)

// A 20 GPU node pool with department A (quota 4, rank High) and B (quota 2, rank Low). A has projects A1 (quota 2)
// and A2 (quota 0); B has B1 (quota 1). Every weight is 1 and no limit is set.
var _ = Describe("Fair share simulation API", Ordered, func() {
	ctx := SetContext(TestTenant, TestUser)
	clusterId := uuid.New().String()
	pool := clustersOpenapi.NodepoolMinimal{
		Name: "fss-pool", Id: "7101", ClusterId: clusterId, Phase: ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
		SchedulingCapabilities: npSchedCapabilities(false, true, true, true),
		GpuTypes:               []clustersOpenapi.NodepoolGpuTypeDetails{{TotalDevices: 20}},
	}

	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		Version: *clustersOpenapi.NewNullableString(ptr.To("2.24")),
	}, nil)
	clusterServiceClient.On("GetClusterByUuid", mock.Anything, mock.Anything).Return(&clustersOpenapi.DisplayedCluster{
		Version: *clustersOpenapi.NewNullableString(ptr.To("2.24")),
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId).
		Return(&clustersOpenapi.GetNodepoolsMinimal200Response{Nodepools: []clustersOpenapi.NodepoolMinimal{pool}}, nil)
	clusterServiceClient.On("GetNodepoolMinimal", mock.Anything, pool.Id).Return(&pool, nil)
	clusterServiceClient.On("GetNodepoolMinimal", mock.Anything, "999997").Return(nil, errors.New("connection refused"))
	clusterServiceClient.On("GetNodepoolMinimal", mock.Anything, mock.Anything).Return(nil, errors.New("404 Not Found"))
	departmentApiService := departments.NewDepartmentApiService(&clusterServiceClient, nil)
	projectApiService := projects.NewProjectApiService(services.NewProjectDeletionService(&clusterServiceClient, nil, nil), &clusterServiceClient, nil)

	var a, b, a1, a2, b1 string

	resources := func(deserved float64, rank string) []api.ResourcesNullable {
		return []api.ResourcesNullable{{
			NodePool: &api.ResourcesNodePool{Id: pool.Id, Name: pool.Name},
			Gpu:      api.ResourcesNullableGpu{Deserved: ptr.To(deserved), Limit: ptr.To(-1.0), OverQuotaWeight: ptr.To(1.0)},
			Rank:     ptr.To(rank),
		}}
	}
	createDepartment := func(deserved float64, rank string) string {
		response, err := departmentApiService.CreateDepartment(ctx,
			getDepartmentResourcesCreateRequest("fss-dep-"+uuid.NewString()[:8], clusterId, resources(deserved, rank)))
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusCreated), "%+v", response.Body)
		return response.Body.(api.Department).Id
	}
	createProject := func(parentId string, deserved float64) string {
		response, err := projectApiService.CreateProject(ctx,
			getProjectResourcesCreateRequest("fss-proj-"+uuid.NewString()[:8], parentId, clusterId, resources(deserved, "MediumLow")))
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusCreated), "%+v", response.Body)
		return response.Body.(api.Project).Id
	}

	serviceFor := func(read authorization_api.PermittedScopes) api.OrgUnitAPIServicer {
		authMock := auth_mocks.IAuthorizationClient{}
		authMock.On("GetPermittedScopes", mock.Anything, mock.Anything, mock.Anything).
			Return(&authorization_api.PermittedScopesActions{Read: read}, 0, nil)
		return services.NewOrgUnitApiService(&authMock, &clusterServiceClient, mockTenantsClient, &clusterServiceClient)
	}
	fullAccess := serviceFor(authorization_api.PermittedScopes{Tenants: []string{TestTenant}})
	allOn := api.FairShareSimulationSchedulingCapabilities{Quota: true, Ranks: true, AdjustableWeights: true, Limits: true}

	type scenario struct {
		capacity     float64
		capabilities api.FairShareSimulationSchedulingCapabilities
		orgUnitType  string
		parent       *api.FairShareSimulationParent
		orgUnits     []api.FairShareSimulationOrgUnit
	}
	departmentsScenario := func(orgUnits ...api.FairShareSimulationOrgUnit) scenario {
		return scenario{capacity: 20, capabilities: allOn, orgUnitType: "department", orgUnits: orgUnits}
	}
	projectsScenario := func(parentId string, orgUnits ...api.FairShareSimulationOrgUnit) scenario {
		return scenario{capacity: 20, capabilities: allOn, orgUnitType: "project",
			parent: &api.FairShareSimulationParent{Id: parentId}, orgUnits: orgUnits}
	}
	row := func(id string) api.FairShareSimulationOrgUnit { return api.FairShareSimulationOrgUnit{Id: id} }
	withDemand := func(id string, demand float64) api.FairShareSimulationOrgUnit {
		return api.FairShareSimulationOrgUnit{Id: id, Demand: &api.FairShareSimulationResources{Gpu: demand}}
	}

	call := func(s api.OrgUnitAPIServicer, sc scenario) api.ImplResponse {
		response, err := s.SimulateNodePoolFairShare(ctx, pool.Id, api.FairShareSimulationRequest{
			NodePool: &api.FairShareSimulationNodePool{
				Capacity: &api.FairShareSimulationResources{Gpu: sc.capacity}, SchedulingCapabilities: &sc.capabilities,
			},
			OrgUnitType: sc.orgUnitType,
			Parent:      sc.parent,
			OrgUnits:    sc.orgUnits,
		})
		Expect(err).To(BeNil())
		return response
	}
	simulate := func(s api.OrgUnitAPIServicer, sc scenario) api.FairShareSimulationResponse {
		response := call(s, sc)
		Expect(response.Code).To(Equal(http.StatusOK), "%+v", response.Body)
		return response.Body.(api.FairShareSimulationResponse)
	}
	type expected struct{ demand, fairShare, guaranteed, unmet, weightShare float64 }
	expectResult := func(r api.FairShareSimulationResult, id string, e expected) {
		Expect(r.Id).To(Equal(id))
		Expect(r.Demand.Gpu).To(Equal(e.demand), "demand of %s", id)
		Expect(r.FairShare.Gpu).To(Equal(e.fairShare), "fair share of %s", id)
		Expect(r.Guaranteed.Gpu).To(Equal(e.guaranteed), "guaranteed of %s", id)
		Expect(r.OverQuota.Gpu).To(Equal(e.fairShare-e.guaranteed), "over quota of %s", id)
		Expect(r.Unmet.Gpu).To(Equal(e.unmet), "unmet of %s", id)
		Expect(r.WeightShareInRank).To(BeNumerically("~", e.weightShare, 1e-9), "weight share of %s", id)
	}
	expectBadRequest := func(s api.OrgUnitAPIServicer, sc scenario) {
		Expect(call(s, sc).Code).To(Equal(http.StatusBadRequest))
	}

	BeforeAll(func() {
		a = createDepartment(4, "High")
		b = createDepartment(2, "Low")
		a1 = createProject(a, 2)
		a2 = createProject(a, 0)
		b1 = createProject(b, 1)
	})

	AfterAll(func() {
		for _, id := range []string{a1, a2, b1} {
			if id == "" {
				continue
			}
			response, err := projectApiService.DeleteProject(ctx, id, "", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusNoContent), "%+v", response.Body)
		}
		for _, id := range []string{a, b} {
			if id == "" {
				continue
			}
			response, err := departmentApiService.DeleteDepartment(ctx, id)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusNoContent), "%+v", response.Body)
		}
	})

	Context("departments", func() {
		It("gives each department its quota, then what's left by rank, at half the capacity", func() {
			body := simulate(fullAccess, departmentsScenario(row(a), row(b)))
			expectResult(body.OrgUnits[0], a, expected{demand: 10, fairShare: 10, guaranteed: 4, unmet: 0, weightShare: 1})
			expectResult(body.OrgUnits[1], b, expected{demand: 10, fairShare: 10, guaranteed: 2, unmet: 0, weightShare: 1})
		})

		It("gives the higher rank everything above quota when demand is the whole pool", func() {
			body := simulate(fullAccess, departmentsScenario(withDemand(b, 20), withDemand(a, 20)))
			expectResult(body.OrgUnits[0], b, expected{demand: 20, fairShare: 2, guaranteed: 2, unmet: 18, weightShare: 1})
			expectResult(body.OrgUnits[1], a, expected{demand: 20, fairShare: 18, guaranteed: 4, unmet: 2, weightShare: 1})
		})

		It("uses a sent demand and unsaved rank and weight", func() {
			high := withDemand(b, 20)
			high.Rank = ptr.To(api.FairShareSimulationRank("High"))
			high.Gpu = &api.FairShareSimulationGpuSettings{OverQuotaWeight: ptr.To(3.0)}
			body := simulate(fullAccess, departmentsScenario(withDemand(a, 20), high))
			Expect(body.OrgUnits[0].FairShare.Gpu + body.OrgUnits[1].FairShare.Gpu).To(Equal(20.0))
			Expect(body.OrgUnits[0].WeightShareInRank).To(Equal(0.25))
			Expect(body.OrgUnits[1].WeightShareInRank).To(Equal(0.75))
			Expect(body.OrgUnits[1].FairShare.Gpu).To(BeNumerically(">", body.OrgUnits[0].FairShare.Gpu))

			body = simulate(fullAccess, departmentsScenario(withDemand(a, 5), withDemand(b, 20)))
			expectResult(body.OrgUnits[0], a, expected{demand: 5, fairShare: 5, guaranteed: 4, unmet: 0, weightShare: 1})
			expectResult(body.OrgUnits[1], b, expected{demand: 20, fairShare: 15, guaranteed: 2, unmet: 5, weightShare: 1})
		})

		It("still grants quotas when the node pool has no GPUs", func() {
			sc := departmentsScenario(withDemand(a, 20), withDemand(b, 20))
			sc.capacity = 0
			body := simulate(fullAccess, sc)
			Expect(body.OrgUnits[0].FairShare.Gpu).To(Equal(4.0))
			Expect(body.OrgUnits[1].FairShare.Gpu).To(Equal(2.0))
		})

		It("resets every department's quota and rank when the scheduling mode turns them off", func() {
			sc := departmentsScenario(withDemand(a, 20), withDemand(b, 20))
			sc.capabilities = api.FairShareSimulationSchedulingCapabilities{AdjustableWeights: true}
			body := simulate(fullAccess, sc)
			expectResult(body.OrgUnits[0], a, expected{demand: 20, fairShare: 10, guaranteed: 0, unmet: 10, weightShare: 0.5})
			expectResult(body.OrgUnits[1], b, expected{demand: 20, fairShare: 10, guaranteed: 0, unmet: 10, weightShare: 0.5})
		})

		It("returns only the requested departments, and counts the others at half the capacity", func() {
			body := simulate(fullAccess, departmentsScenario(withDemand(b, 20)))
			Expect(body.OrgUnits).To(HaveLen(1))
			expectResult(body.OrgUnits[0], b, expected{demand: 20, fairShare: 10, guaranteed: 2, unmet: 10, weightShare: 1})
		})

		It("uses each department's quota as its weight while adjustable weights is off, ignoring a sent weight", func() {
			weighted := withDemand(a, 20)
			weighted.Gpu = &api.FairShareSimulationGpuSettings{OverQuotaWeight: ptr.To(9.0)}
			sc := departmentsScenario(weighted, withDemand(b, 20))
			sc.capabilities = api.FairShareSimulationSchedulingCapabilities{Quota: true, Limits: true}
			body := simulate(fullAccess, sc)
			Expect(body.OrgUnits[0].WeightShareInRank).To(BeNumerically("~", 4.0/6, 1e-9))
			Expect(body.OrgUnits[1].WeightShareInRank).To(BeNumerically("~", 2.0/6, 1e-9))
			Expect(body.OrgUnits[0].FairShare.Gpu + body.OrgUnits[1].FairShare.Gpu).To(Equal(20.0))
			Expect(body.OrgUnits[0].FairShare.Gpu).To(BeNumerically(">", body.OrgUnits[1].FairShare.Gpu))
		})

		It("guarantees the whole demand under an unlimited quota", func() {
			unlimited := api.FairShareSimulationOrgUnit{Id: a, Gpu: &api.FairShareSimulationGpuSettings{Deserved: ptr.To(-1.0)}}
			body := simulate(fullAccess, departmentsScenario(unlimited, row(b)))
			expectResult(body.OrgUnits[0], a, expected{demand: 10, fairShare: 10, guaranteed: 10, unmet: 0, weightShare: 1})
		})

		It("caps the demand and the fair share at an unsaved limit", func() {
			limited := api.FairShareSimulationOrgUnit{Id: a, Gpu: &api.FairShareSimulationGpuSettings{Limit: ptr.To(5.0)}}
			body := simulate(fullAccess, departmentsScenario(limited, row(b)))
			expectResult(body.OrgUnits[0], a, expected{demand: 5, fairShare: 5, guaranteed: 4, unmet: 0, weightShare: 1})
		})

		It("refuses a department the caller can't read the same way as an unknown one", func() {
			onlyB := serviceFor(authorization_api.PermittedScopes{Departments: []string{b}})
			Expect(simulate(onlyB, departmentsScenario(row(b))).OrgUnits[0].FairShare.Gpu).To(Equal(10.0))
			hidden := call(onlyB, departmentsScenario(row(a)))
			unknown := call(onlyB, departmentsScenario(row("999999")))
			Expect(hidden.Code).To(Equal(http.StatusBadRequest))
			Expect(unknown.Code).To(Equal(http.StatusBadRequest))
			Expect(hidden.Body.(api.Error).Message).To(Equal("department " + a + " not found in the node pool"))
			Expect(unknown.Body.(api.Error).Message).To(Equal("department 999999 not found in the node pool"))
		})
	})

	Context("projects of a parent", func() {
		It("lets the parent take the node pool while the other departments are idle", func() {
			body := simulate(fullAccess, projectsScenario(a, row(a1), row(a2)))
			expectResult(body.OrgUnits[0], a1, expected{demand: 10, fairShare: 10, guaranteed: 2, unmet: 0, weightShare: 0.5})
			expectResult(body.OrgUnits[1], a2, expected{demand: 10, fairShare: 10, guaranteed: 0, unmet: 0, weightShare: 0.5})
		})

		It("counts the projects that weren't sent at half the capacity", func() {
			body := simulate(fullAccess, projectsScenario(a, withDemand(a1, 15)))
			Expect(body.OrgUnits).To(HaveLen(1))
			expectResult(body.OrgUnits[0], a1, expected{demand: 15, fairShare: 11, guaranteed: 2, unmet: 4, weightShare: 0.5})
		})

		It("serves a caller who can read the project but not its department", func() {
			onlyA1 := serviceFor(authorization_api.PermittedScopes{Projects: []string{a1}})
			body := simulate(onlyA1, projectsScenario(a, row(a1)))
			expectResult(body.OrgUnits[0], a1, expected{demand: 10, fairShare: 10, guaranteed: 2, unmet: 0, weightShare: 0.5})
			expectBadRequest(onlyA1, projectsScenario(a, row(a2)))
		})

		It("refuses projects of another department, and an unknown parent like an unknown project", func() {
			expectBadRequest(fullAccess, projectsScenario(a, row(b1)))
			unknownParent := call(fullAccess, projectsScenario("999999", row(a1)))
			Expect(unknownParent.Code).To(Equal(http.StatusBadRequest))
			Expect(unknownParent.Body.(api.Error).Message).To(Equal("project " + a1 + " not found in the node pool"))
		})
	})

	Context("authorization", func() {
		readOn := func(departments, projects authorization_api.PermittedScopes) api.OrgUnitAPIServicer {
			authMock := auth_mocks.IAuthorizationClient{}
			authMock.On("GetPermittedScopes", mock.Anything, authorization_api.RESOURCETYPE_DEPARTMENT, mock.Anything).
				Return(&authorization_api.PermittedScopesActions{Read: departments}, 0, nil)
			authMock.On("GetPermittedScopes", mock.Anything, authorization_api.RESOURCETYPE_PROJECT, mock.Anything).
				Return(&authorization_api.PermittedScopesActions{Read: projects}, 0, nil)
			return services.NewOrgUnitApiService(&authMock, &clusterServiceClient, mockTenantsClient, &clusterServiceClient)
		}

		It("answers 503 when the caller's permissions can't be fetched", func() {
			authMock := auth_mocks.IAuthorizationClient{}
			authMock.On("GetPermittedScopes", mock.Anything, mock.Anything, mock.Anything).
				Return(nil, 0, errors.New("authorization is down"))
			unavailable := services.NewOrgUnitApiService(&authMock, &clusterServiceClient, mockTenantsClient, &clusterServiceClient)
			Expect(call(unavailable, departmentsScenario(row(a))).Code).To(Equal(http.StatusServiceUnavailable))
		})

		It("requires read on departments to simulate departments", func() {
			projectsOnly := readOn(authorization_api.PermittedScopes{}, authorization_api.PermittedScopes{Projects: []string{a1}})
			Expect(call(projectsOnly, departmentsScenario(row(a))).Code).To(Equal(http.StatusForbidden))
		})

		It("requires read on projects to simulate projects", func() {
			departmentsOnly := readOn(authorization_api.PermittedScopes{Tenants: []string{TestTenant}}, authorization_api.PermittedScopes{})
			Expect(call(departmentsOnly, projectsScenario(a, row(a1))).Code).To(Equal(http.StatusForbidden))
			Expect(call(departmentsOnly, departmentsScenario(row(a))).Code).To(Equal(http.StatusOK))
		})
	})

	Context("validation", func() {
		It("refuses an invalid scenario", func() {
			departmentsWithParent := departmentsScenario(row(a))
			departmentsWithParent.parent = &api.FairShareSimulationParent{Id: a}
			expectBadRequest(fullAccess, departmentsWithParent)

			projectsWithoutParent := projectsScenario(a, row(a1))
			projectsWithoutParent.parent = nil
			expectBadRequest(fullAccess, projectsWithoutParent)

			limitsWithoutQuota := departmentsScenario(row(a))
			limitsWithoutQuota.capabilities = api.FairShareSimulationSchedulingCapabilities{Limits: true}
			expectBadRequest(fullAccess, limitsWithoutQuota)

			negativeCapacity := departmentsScenario(row(a))
			negativeCapacity.capacity = -1
			expectBadRequest(fullAccess, negativeCapacity)

			expectBadRequest(fullAccess, departmentsScenario(row(a), row(a)))
			expectBadRequest(fullAccess, departmentsScenario())
		})

		It("refuses values the scheduling mode turns off", func() {
			rankWithRanksOff := departmentsScenario(api.FairShareSimulationOrgUnit{Id: a, Rank: ptr.To(api.FairShareSimulationRank("Low"))})
			rankWithRanksOff.capabilities.Ranks = false
			expectBadRequest(fullAccess, rankWithRanksOff)

			limitWithLimitsOff := departmentsScenario(api.FairShareSimulationOrgUnit{Id: a, Gpu: &api.FairShareSimulationGpuSettings{Limit: ptr.To(8.0)}})
			limitWithLimitsOff.capabilities.Limits = false
			expectBadRequest(fullAccess, limitWithLimitsOff)

			expectBadRequest(fullAccess, departmentsScenario(api.FairShareSimulationOrgUnit{Id: a, Rank: ptr.To(api.FairShareSimulationRank("Sometimes"))}))
			expectBadRequest(fullAccess, departmentsScenario(withDemand(a, -1)))
			expectBadRequest(fullAccess, departmentsScenario(api.FairShareSimulationOrgUnit{Id: a,
				Gpu: &api.FairShareSimulationGpuSettings{Deserved: ptr.To(-0.5)}}))
			expectBadRequest(fullAccess, departmentsScenario(api.FairShareSimulationOrgUnit{Id: a,
				Gpu: &api.FairShareSimulationGpuSettings{Deserved: ptr.To(6.0), Limit: ptr.To(5.0)}}))
		})

		It("lets zero values and omitted edits through the controller", func() {
			post := func(body string) *httptest.ResponseRecorder {
				routeContext := chi.NewRouteContext()
				routeContext.URLParams.Add("nodepoolId", pool.Id)
				request := httptest.NewRequest(http.MethodPost, "/api/v1/org-unit/node-pools/"+pool.Id+"/fair-share-simulation",
					strings.NewReader(body)).WithContext(context.WithValue(ctx, chi.RouteCtxKey, routeContext))
				recorder := httptest.NewRecorder()
				api.NewOrgUnitAPIController(fullAccess).SimulateNodePoolFairShare(recorder, request)
				return recorder
			}
			allOff := `{"capacity":{"gpu":0},"schedulingCapabilities":{"quota":false,"ranks":false,"adjustableWeights":false,"limits":false}}`

			response := post(`{"nodePool":` + allOff + `,"orgUnitType":"department","orgUnits":[{"id":"` + a + `","demand":{"gpu":5}}]}`)
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
			Expect(response.Body.String()).To(ContainSubstring(`"fairShare":{"gpu":0}`))

			response = post(`{"nodePool":{"capacity":{"gpu":20},"schedulingCapabilities":{"quota":true,"ranks":true,"adjustableWeights":true,"limits":true}},` +
				`"orgUnitType":"department","orgUnits":[{"id":"` + a + `","demand":null,"gpu":{"limit":null}}]}`)
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
			Expect(response.Body.String()).To(ContainSubstring(`"demand":{"gpu":10}`))

			Expect(post(`{"nodePool":null,"orgUnitType":"department","orgUnits":[{"id":"` + a + `"}]}`).Code).
				To(Equal(http.StatusBadRequest))
			Expect(post(`{"nodePool":` + allOff + `,"orgUnitType":"department","orgUnits":[{"id":"` + a + `","rank":"Sometimes"}]}`).Code).
				To(Equal(http.StatusBadRequest))
		})

		It("answers 404 for a node pool that doesn't exist", func() {
			response, err := fullAccess.SimulateNodePoolFairShare(ctx, "999998", api.FairShareSimulationRequest{
				NodePool: &api.FairShareSimulationNodePool{
					Capacity: &api.FairShareSimulationResources{}, SchedulingCapabilities: &allOn,
				},
				OrgUnitType: "department", OrgUnits: []api.FairShareSimulationOrgUnit{row(a)},
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})

		It("answers 503 when cluster-service can't say whether the node pool exists", func() {
			response, err := fullAccess.SimulateNodePoolFairShare(ctx, "999997", api.FairShareSimulationRequest{
				NodePool: &api.FairShareSimulationNodePool{
					Capacity: &api.FairShareSimulationResources{}, SchedulingCapabilities: &allOn,
				},
				OrgUnitType: "department", OrgUnits: []api.FairShareSimulationOrgUnit{row(a)},
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusServiceUnavailable))
		})

		It("refuses ids that would change the filter", func() {
			response, err := fullAccess.SimulateNodePoolFairShare(ctx, pool.Id+"||7201", api.FairShareSimulationRequest{
				NodePool: &api.FairShareSimulationNodePool{
					Capacity: &api.FairShareSimulationResources{}, SchedulingCapabilities: &allOn,
				},
				OrgUnitType: "department", OrgUnits: []api.FairShareSimulationOrgUnit{row(a)},
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))

			parent := call(fullAccess, projectsScenario(a+"||"+b, row(a1)))
			Expect(parent.Code).To(Equal(http.StatusBadRequest))
			Expect(parent.Body.(api.Error).Message).To(Equal("project " + a1 + " not found in the node pool"))
		})
	})
	Context("a department with a stored limit", Ordered, func() {
		var c, c1, c2 string

		BeforeAll(func() {
			response, err := departmentApiService.CreateDepartment(ctx, getDepartmentResourcesCreateRequest(
				"fss-dep-"+uuid.NewString()[:8], clusterId, []api.ResourcesNullable{{
					NodePool: &api.ResourcesNodePool{Id: pool.Id, Name: pool.Name},
					Gpu:      api.ResourcesNullableGpu{Deserved: ptr.To(0.0), Limit: ptr.To(6.0), OverQuotaWeight: ptr.To(1.0)},
					Rank:     ptr.To("High"),
				}}))
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated), "%+v", response.Body)
			c = response.Body.(api.Department).Id
			c1 = createProject(c, 0)
			c2 = createProject(c, 0)
		})

		AfterAll(func() {
			for _, id := range []string{c1, c2} {
				if id == "" {
					continue
				}
				response, err := projectApiService.DeleteProject(ctx, id, "", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusNoContent), "%+v", response.Body)
			}
			if c != "" {
				response, err := departmentApiService.DeleteDepartment(ctx, c)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusNoContent), "%+v", response.Body)
			}
		})

		It("caps a department that wasn't sent at its stored limit", func() {
			body := simulate(fullAccess, departmentsScenario(withDemand(a, 20)))
			Expect(body.OrgUnits[0].FairShare.Gpu).To(Equal(12.0))
		})

		It("caps the parent's projects at the parent's stored limit", func() {
			body := simulate(fullAccess, projectsScenario(c, row(c1), row(c2)))
			Expect(body.OrgUnits[0].FairShare.Gpu + body.OrgUnits[1].FairShare.Gpu).To(Equal(6.0))
		})
	})
})
