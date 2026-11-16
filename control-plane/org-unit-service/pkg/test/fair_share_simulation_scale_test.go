/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

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
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/simulation"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

// Opt-in: seeding takes minutes, so it only runs with FAIR_SHARE_SIMULATION_SCALE=1.
var _ = Describe("Fair share simulation at scale", Ordered, func() {
	const (
		departmentCount = 300
		projectCount    = 1500
		calls           = 20
		capacity        = 1024.0
	)
	ctx := SetContext(TestTenant, TestUser)
	clusterId := uuid.New().String()
	pool := clustersOpenapi.NodepoolMinimal{
		Name: "fss-scale-pool", Id: "7201", ClusterId: clusterId, Phase: ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
		SchedulingCapabilities: npSchedCapabilities(false, true, true, true),
		GpuTypes:               []clustersOpenapi.NodepoolGpuTypeDetails{{TotalDevices: int32(capacity)}},
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
	ranks := []string{"High", "Medium", "MediumLow", "Low"}

	authMock := auth_mocks.IAuthorizationClient{}
	authMock.On("GetPermittedScopes", mock.Anything, mock.Anything, mock.Anything).
		Return(&authorization_api.PermittedScopesActions{Read: authorization_api.PermittedScopes{Tenants: []string{TestTenant}}}, 0, nil)
	service := services.NewOrgUnitApiService(&authMock, &clusterServiceClient, mockTenantsClient, &clusterServiceClient)
	allOn := api.FairShareSimulationSchedulingCapabilities{Quota: true, Ranks: true, AdjustableWeights: true, Limits: true}

	var departmentIds, projectIds []string
	var focus string

	resources := func(deserved float64, rank string, weight float64) []api.ResourcesNullable {
		return []api.ResourcesNullable{{
			NodePool: &api.ResourcesNodePool{Id: pool.Id, Name: pool.Name},
			Gpu:      api.ResourcesNullableGpu{Deserved: ptr.To(deserved), Limit: ptr.To(-1.0), OverQuotaWeight: ptr.To(weight)},
			Rank:     ptr.To(rank),
		}}
	}

	BeforeAll(func() {
		if os.Getenv("FAIR_SHARE_SIMULATION_SCALE") == "" {
			Skip("set FAIR_SHARE_SIMULATION_SCALE=1 to run")
		}
		start := time.Now()
		departmentIds = make([]string, departmentCount)
		parallel(departmentCount, func(i int) {
			response, err := departmentApiService.CreateDepartment(ctx, getDepartmentResourcesCreateRequest(
				fmt.Sprintf("fss-scale-dep-%d-%s", i, uuid.NewString()[:6]), clusterId, resources(float64(i%5), ranks[i%len(ranks)], float64(1+i%4))))
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated), "%+v", response.Body)
			departmentIds[i] = response.Body.(api.Department).Id
		})
		focus = departmentIds[0]
		projectIds = make([]string, projectCount)
		parallel(projectCount, func(i int) {
			response, err := projectApiService.CreateProject(ctx, getProjectResourcesCreateRequest(
				fmt.Sprintf("fss-scale-proj-%d-%s", i, uuid.NewString()[:6]), focus, clusterId, resources(float64(i%3), ranks[i%len(ranks)], float64(1+i%3))))
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated), "%+v", response.Body)
			projectIds[i] = response.Body.(api.Project).Id
		})
		GinkgoWriter.Printf("seeded %d departments and %d projects in %s\n", departmentCount, projectCount, time.Since(start).Round(time.Second))
		// A long-lived database has fresh statistics; right after seeding, Postgres plans the views badly until it analyzes.
		_, err := runaidb.DB().WithContext(ctx).Exec("ANALYZE")
		Expect(err).To(BeNil())
	})

	measure := func(name string, call func()) time.Duration {
		durations := make([]time.Duration, calls)
		for i := range durations {
			start := time.Now()
			call()
			durations[i] = time.Since(start)
		}
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		median, p95 := durations[calls/2], durations[calls*95/100]
		fmt.Printf("SCALE %-58s median %8s  p95 %8s\n", name, median.Round(100*time.Microsecond), p95.Round(100*time.Microsecond))
		return median
	}
	simulate := func(request api.FairShareSimulationRequest) api.FairShareSimulationResponse {
		response, err := service.SimulateNodePoolFairShare(ctx, pool.Id, request)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusOK), "%+v", response.Body)
		return response.Body.(api.FairShareSimulationResponse)
	}
	rowsOf := func(ids []string) []api.FairShareSimulationOrgUnit {
		rows := make([]api.FairShareSimulationOrgUnit, len(ids))
		for i, id := range ids {
			rows[i] = api.FairShareSimulationOrgUnit{Id: id}
		}
		return rows
	}
	request := func(orgUnitType string, rows []api.FairShareSimulationOrgUnit) api.FairShareSimulationRequest {
		r := api.FairShareSimulationRequest{
			NodePool: &api.FairShareSimulationNodePool{
				Capacity: &api.FairShareSimulationResources{Gpu: capacity}, SchedulingCapabilities: &allOn,
			},
			OrgUnitType: orgUnitType,
			OrgUnits:    rows,
		}
		if orgUnitType == "project" {
			r.Parent = &api.FairShareSimulationParent{Id: focus}
		}
		return r
	}

	It("times every form and its parts", func() {
		nodePoolForm := request("department", rowsOf(departmentIds))
		departmentForm500 := request("project", rowsOf(projectIds[:500]))
		departmentForm20 := request("project", rowsOf(projectIds[:20]))

		body := simulate(nodePoolForm)
		Expect(body.OrgUnits).To(HaveLen(departmentCount))
		body = simulate(departmentForm500)
		Expect(body.OrgUnits).To(HaveLen(500))

		departmentFilters, err := queryapi.NewParser(models.DepartmentNodePoolResources{}.FilterByFieldsToColumn()).Filters([]string{"nodePoolId==" + pool.Id})
		Expect(err).To(BeNil())
		projectFilters, err := queryapi.NewParser(models.ProjectNodePoolResources{}.FilterByFieldsToColumn()).
			Filters([]string{"nodePoolId==" + pool.Id, "parentId==" + focus})
		Expect(err).To(BeNil())
		measure("  part: departments query (300 rows)", func() {
			rows, err := dao.GetDepartmentNodePoolResources(ctx, TestTenant, departmentFilters, "", "", 0, 0, nil, "", models.SlimVerbosity)
			Expect(err).To(BeNil())
			Expect(*rows).To(HaveLen(departmentCount))
		})
		measure("  part: projects query, full row (1,500 rows)", func() {
			rows, err := dao.GetProjectNodePoolResources(ctx, TestTenant, projectFilters, "", "", 0, 0, nil, "", models.VerboseVerbosity)
			Expect(err).To(BeNil())
			Expect(*rows).To(HaveLen(projectCount))
		})
		measure("  part: projects query, slim (1,500 rows)", func() {
			rows, err := dao.GetProjectNodePoolResources(ctx, TestTenant, projectFilters, "", "", 0, 0, nil, "", models.SlimVerbosity)
			Expect(err).To(BeNil())
			Expect(*rows).To(HaveLen(projectCount))
		})

		measure("department form, 1,500 projects, 20 sent (first)", func() { simulate(departmentForm20) })
		measure("department form, 1,500 projects, 500 sent", func() { simulate(departmentForm500) })
		measure("department form, 1,500 projects, 20 sent (again)", func() { simulate(departmentForm20) })
		measure("node pool form, 300 departments sent", func() { simulate(nodePoolForm) })

		measure("  part again: projects query (1,500 rows)", func() {
			rows, err := dao.GetProjectNodePoolResources(ctx, TestTenant, projectFilters, "", "", 0, 0, nil, "", models.SlimVerbosity)
			Expect(err).To(BeNil())
			Expect(*rows).To(HaveLen(projectCount))
		})

		units := make([]simulation.Unit, projectCount)
		for i := range units {
			units[i] = simulation.Unit{ID: strconv.Itoa(i), Priority: i % 4, CreatedAt: time.Unix(int64(i), 0),
				Deserved: float64(i % 3), Limit: simulation.Unlimited, Weight: float64(1 + i%3), Demand: capacity / 2}
		}
		measure("  part: KAI division, 1,500 units", func() { simulation.Divide(capacity, units) })

		response500 := simulate(departmentForm500)
		measure("  part: JSON of a 500-row response", func() {
			_, err := json.Marshal(response500)
			Expect(err).To(BeNil())
		})
	})
})

func parallel(n int, do func(i int)) {
	const workers = 16
	var wg sync.WaitGroup
	next := make(chan int)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			// GinkgoRecover must run before wg.Done, so a failure is recorded before the producer is released.
			defer wg.Done()
			defer GinkgoRecover()
			for i := range next {
				do(i)
			}
		}()
	}
	// A failed assertion ends its worker; once none is left, nothing receives, so stop sending instead of blocking.
	workersDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(workersDone)
	}()
dispatch:
	for i := 0; i < n; i++ {
		select {
		case next <- i:
		case <-workersDone:
			break dispatch
		}
	}
	close(next)
	<-workersDone
}
