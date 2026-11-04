/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"fmt"
	"net/http"

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
// and A2 (quota 0), B has B1 (quota 1). Every weight is 1. At full demand A gets its quota plus all 14 left over GPUs,
// and splits its 18 as A1 2+8, A2 0+8; B gets its quota 2, which B1 takes as 1+1.
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
	fullAccess := authorization_api.PermittedScopes{Tenants: []string{TestTenant}}

	simulate := func(s api.OrgUnitAPIServicer, request api.FairShareSimulationRequest, orgUnitType, parentId string) map[string]api.FairShareSimulationResult {
		response, err := s.SimulateNodePoolFairShare(ctx, pool.Id, request, orgUnitType, parentId)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusOK), "%+v", response.Body)
		body := response.Body.(api.FairShareSimulationResponse)
		Expect(body.Capacity.Gpu).To(Equal(20.0))
		results := map[string]api.FairShareSimulationResult{}
		for _, r := range body.OrgUnits {
			results[r.Type+"/"+r.Id] = r
		}
		return results
	}
	expectShare := func(results map[string]api.FairShareSimulationResult, key string, fairShare, guaranteed, unmet float64) {
		Expect(results).To(HaveKey(key))
		r := results[key]
		Expect(r.FairShare.Gpu).To(Equal(fairShare), key)
		Expect(r.Guaranteed.Gpu).To(Equal(guaranteed), key)
		Expect(r.OverQuota.Gpu).To(Equal(fairShare-guaranteed), key)
		Expect(r.Unmet.Gpu).To(Equal(unmet), key)
	}
	expectStatus := func(s api.OrgUnitAPIServicer, request api.FairShareSimulationRequest, status int) {
		response, err := s.SimulateNodePoolFairShare(ctx, pool.Id, request, "", "")
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(status), "%+v", response.Body)
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
			_, _ = projectApiService.DeleteProject(ctx, id, "", false)
		}
		for _, id := range []string{a, b} {
			_, _ = departmentApiService.DeleteDepartment(ctx, id)
		}
	})

	It("splits the node pool at full demand", func() {
		results := simulate(serviceFor(fullAccess), api.FairShareSimulationRequest{}, "", "")

		expectShare(results, "department/"+a, 18, 4, 2)
		expectShare(results, "department/"+b, 2, 2, 18)
		expectShare(results, "project/"+a1, 10, 2, 10)
		expectShare(results, "project/"+a2, 8, 0, 12)
		expectShare(results, "project/"+b1, 2, 1, 18)
		Expect(results["project/"+a1].ParentId).To(Equal(ptr.To(a)))
	})

	It("gives what a department's projects don't ask for to its siblings", func() {
		results := simulate(serviceFor(fullAccess), api.FairShareSimulationRequest{
			OrgUnits: []api.FairShareSimulationOrgUnit{{Id: b1, Type: "project", Demand: &api.QuotaStatusResource{Gpu: 0}}},
		}, "", "")

		expectShare(results, "department/"+a, 20, 4, 0)
		expectShare(results, "department/"+b, 0, 0, 0)
		expectShare(results, "project/"+a1, 11, 2, 9)
		expectShare(results, "project/"+b1, 0, 0, 0)
	})

	It("uses changed settings in place of the stored ones", func() {
		results := simulate(serviceFor(fullAccess), api.FairShareSimulationRequest{
			OrgUnits: []api.FairShareSimulationOrgUnit{{
				Id: a2, Type: "project", Rank: "MediumLow",
				Resources: &api.ResourcesPerNodePool{Gpu: api.ResourcesPerNodePoolGpu{Deserved: ptr.To(6.0), Limit: ptr.To(-1.0), OverQuotaWeight: ptr.To(1.0)}},
			}},
		}, "", "")

		expectShare(results, "project/"+a1, 7, 2, 13)
		expectShare(results, "project/"+a2, 11, 6, 9)
	})

	It("applies an unsaved scheduling mode to every org unit", func() {
		results := simulate(serviceFor(fullAccess), api.FairShareSimulationRequest{
			NodePool: &api.FairShareSimulationRequestNodePool{SchedulingCapabilities: &api.FairShareSimulationSchedulingCapabilities{
				Quota: false, Ranks: false, AdjustableWeights: true,
			}},
		}, "", "")

		expectShare(results, "department/"+a, 10, 0, 10)
		expectShare(results, "department/"+b, 10, 0, 10)
		expectShare(results, "project/"+a1, 5, 0, 15)
		expectShare(results, "project/"+b1, 10, 0, 10)
	})

	It("narrows the returned org units without changing their shares", func() {
		results := simulate(serviceFor(fullAccess), api.FairShareSimulationRequest{}, "project", a)

		Expect(results).To(HaveLen(2))
		expectShare(results, "project/"+a1, 10, 2, 10)
		expectShare(results, "project/"+a2, 8, 0, 12)
	})

	It("computes department rows from project counts the same as from every project", func() {
		results := simulate(serviceFor(fullAccess), api.FairShareSimulationRequest{}, "department", "")

		Expect(results).To(HaveLen(2))
		expectShare(results, "department/"+a, 18, 4, 2)
		expectShare(results, "department/"+b, 2, 2, 18)
	})

	It("adds a changed project's demand to its department when only department rows are returned", func() {
		withoutB1 := simulate(serviceFor(fullAccess), api.FairShareSimulationRequest{
			OrgUnits: []api.FairShareSimulationOrgUnit{{Id: b1, Type: "project", Demand: &api.QuotaStatusResource{Gpu: 0}}},
		}, "department", "")
		expectShare(withoutB1, "department/"+a, 20, 4, 0)
		expectShare(withoutB1, "department/"+b, 0, 0, 0)

		withoutA1 := simulate(serviceFor(fullAccess), api.FairShareSimulationRequest{
			OrgUnits: []api.FairShareSimulationOrgUnit{{Id: a1, Type: "project", Demand: &api.QuotaStatusResource{Gpu: 0}}},
		}, "department", "")
		expectShare(withoutA1, "department/"+a, 18, 4, 2)
		expectShare(withoutA1, "department/"+b, 2, 2, 18)
	})

	It("splits one department's share with a changed project of another department", func() {
		results := simulate(serviceFor(fullAccess), api.FairShareSimulationRequest{
			OrgUnits: []api.FairShareSimulationOrgUnit{{Id: b1, Type: "project", Demand: &api.QuotaStatusResource{Gpu: 0}}},
		}, "project", a)

		Expect(results).To(HaveLen(2))
		expectShare(results, "project/"+a1, 11, 2, 9)
		expectShare(results, "project/"+a2, 9, 0, 11)
	})

	It("returns only org units the caller can read, computed over all of them", func() {
		s := serviceFor(authorization_api.PermittedScopes{Departments: []string{b}, Projects: []string{b1}})
		results := simulate(s, api.FairShareSimulationRequest{}, "", "")

		Expect(results).To(HaveLen(2))
		expectShare(results, "department/"+b, 2, 2, 18)
		expectShare(results, "project/"+b1, 2, 1, 18)

		expectStatus(s, api.FairShareSimulationRequest{
			OrgUnits: []api.FairShareSimulationOrgUnit{{Id: a1, Type: "project", Demand: &api.QuotaStatusResource{Gpu: 1}}},
		}, http.StatusBadRequest)
	})

	DescribeTable("rejects invalid changes",
		func(change func() api.FairShareSimulationOrgUnit, capabilities *api.FairShareSimulationSchedulingCapabilities) {
			request := api.FairShareSimulationRequest{OrgUnits: []api.FairShareSimulationOrgUnit{change()}}
			if capabilities != nil {
				request.NodePool = &api.FairShareSimulationRequestNodePool{SchedulingCapabilities: capabilities}
			}
			expectStatus(serviceFor(fullAccess), request, http.StatusBadRequest)
		},
		Entry("demand on a department", func() api.FairShareSimulationOrgUnit {
			return api.FairShareSimulationOrgUnit{Id: a, Type: "department", Demand: &api.QuotaStatusResource{Gpu: 3}}
		}, nil),
		Entry("negative demand", func() api.FairShareSimulationOrgUnit {
			return api.FairShareSimulationOrgUnit{Id: a1, Type: "project", Demand: &api.QuotaStatusResource{Gpu: -1}}
		}, nil),
		Entry("an unknown org unit", func() api.FairShareSimulationOrgUnit {
			return api.FairShareSimulationOrgUnit{Id: "999999", Type: "project", Demand: &api.QuotaStatusResource{Gpu: 1}}
		}, nil),
		Entry("rank without resources", func() api.FairShareSimulationOrgUnit {
			return api.FairShareSimulationOrgUnit{Id: a1, Type: "project", Rank: "High"}
		}, nil),
		Entry("quota while quota is off", func() api.FairShareSimulationOrgUnit {
			return api.FairShareSimulationOrgUnit{Id: a1, Type: "project", Rank: "MediumLow",
				Resources: &api.ResourcesPerNodePool{Gpu: api.ResourcesPerNodePoolGpu{Deserved: ptr.To(3.0), Limit: ptr.To(-1.0), OverQuotaWeight: ptr.To(1.0)}}}
		}, &api.FairShareSimulationSchedulingCapabilities{Quota: false, Ranks: true, AdjustableWeights: true}),
	)

	It("rejects the same org unit twice", func() {
		change := api.FairShareSimulationOrgUnit{Id: a1, Type: "project", Demand: &api.QuotaStatusResource{Gpu: 1}}
		expectStatus(serviceFor(fullAccess), api.FairShareSimulationRequest{OrgUnits: []api.FairShareSimulationOrgUnit{change, change}},
			http.StatusBadRequest)
	})

	It("rejects an unknown org unit type filter", func() {
		response, err := serviceFor(fullAccess).SimulateNodePoolFairShare(ctx, pool.Id, api.FairShareSimulationRequest{}, "tenant", "")
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusBadRequest), fmt.Sprintf("%+v", response.Body))
	})
})
