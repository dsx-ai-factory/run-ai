package services

import (
	contextpkg "context"
	"net/http"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	clientsclusterservice "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/clusterservice"
	workloadsApi "github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/run-ai/runai/control-plane/workloads/mocks/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"k8s.io/utils/ptr"
)

var _ = Describe("Topology persistence on SyncWorkloads", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()

	AfterEach(func() {
		testing.CleanTables()
	})

	It("stores NULL when Topology is nil", func() {
		w := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
		w.Topology = nil

		req := workloadsApi.WorkloadsSyncRequest{ProjectId: DefaultProjectId, Workloads: []workloadsApi.WorkloadSyncFields{w}}
		resp, err := workloadsService.SyncWorkloads(ctx, []string{}, req)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))

		var row models.Workload
		res := dao_utils.DB().WithContext(contextpkg.Background()).Table("workloads").Where("id = ?", testing.TestUuid1).First(&row)
		Expect(res.Error).To(BeNil())
		Expect(row.Topology).To(BeNil())

		apiResp, _ := workloadsService.GetWorkload(ctx, testing.TestUuid1)
		Expect(apiResp.Code).To(Equal(http.StatusOK))
		wd := httpResponseBodyToWorkloadResponse(apiResp)
		Expect(wd.Topology).To(BeNil())
	})

	It("stores JSON when Topology has values", func() {
		w := SetupBaseWorkload(testing.TestUuid2, DefaultProjectId, testing.ClusterUuid1)
		w.Topology = &workloadsApi.WorkloadSyncFieldsTopology{
			Name:               "nvl-topology",
			PreferredPlacement: "rack-a",
			RequiredPlacement:  "zone-a",
		}

		req := workloadsApi.WorkloadsSyncRequest{ProjectId: DefaultProjectId, Workloads: []workloadsApi.WorkloadSyncFields{w}}
		resp, err := workloadsService.SyncWorkloads(ctx, []string{}, req)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))

		var row models.Workload
		res := dao_utils.DB().WithContext(contextpkg.Background()).Table("workloads").Where("id = ?", testing.TestUuid2).First(&row)
		Expect(res.Error).To(BeNil())
		Expect(row.Topology).ToNot(BeNil())
		Expect(row.Topology.Name).To(Equal("nvl-topology"))
		Expect(row.Topology.PreferredPlacement).To(Equal("rack-a"))
		Expect(row.Topology.RequiredPlacement).To(Equal("zone-a"))

		apiResp, _ := workloadsService.GetWorkload(ctx, testing.TestUuid2)
		Expect(apiResp.Code).To(Equal(http.StatusOK))
		wd := httpResponseBodyToWorkloadResponse(apiResp)
		Expect(wd.Topology).ToNot(BeNil())
		Expect(wd.Topology.ActualPlacement).To(BeNil())
	})

	It("stores only provided fields (partial)", func() {
		w := SetupBaseWorkload(testing.TestUuid3, DefaultProjectId, testing.ClusterUuid1)
		w.Topology = &workloadsApi.WorkloadSyncFieldsTopology{
			PreferredPlacement: "zone-b",
		}

		req := workloadsApi.WorkloadsSyncRequest{ProjectId: DefaultProjectId, Workloads: []workloadsApi.WorkloadSyncFields{w}}
		resp, err := workloadsService.SyncWorkloads(ctx, []string{}, req)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))

		var row models.Workload
		res := dao_utils.DB().WithContext(contextpkg.Background()).Table("workloads").Where("id = ?", testing.TestUuid3).First(&row)
		Expect(res.Error).To(BeNil())
		Expect(row.Topology).ToNot(BeNil())
		Expect(row.Topology.PreferredPlacement).To(Equal("zone-b"))
		// other fields should be zero-values (empty strings) when omitted
		Expect(row.Topology.Name).To(Equal(""))
		Expect(row.Topology.RequiredPlacement).To(Equal(""))

		apiResp, _ := workloadsService.GetWorkload(ctx, testing.TestUuid3)
		Expect(apiResp.Code).To(Equal(http.StatusOK))
		wd := httpResponseBodyToWorkloadResponse(apiResp)
		Expect(wd.Topology).ToNot(BeNil())
		Expect(wd.Topology.ActualPlacement).To(BeNil())
	})
})

var _ = Describe("Topology ActualPlacement on GetWorkload", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()
	podsService := NewPodsService(&bc_mocks.TenantSettingsClient{})

	topologyName := "test-net-topology"
	topologyLevels := []string{"zone", "rack", "host"}
	clusterVersion := "2.25.0"

	setupWorkloadWithTopologyPods := func(clusterId string, nodeLabels map[string][]string) string {
		// Setup cluster capability check
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(clusterId, testing.BuildDisplayedCluster(clusterId, clusterVersion, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		// Setup cluster service mock with topology and node labels
		mock := clientsclusterservice.NewSimpleMock().
			WithTopology(topologyName, &clusterApi.NetworkTopology{
				Name:   topologyName,
				Levels: topologyLevels,
			})
		for nodeName, labels := range nodeLabels {
			mock.WithNodeNetworkTopologyLabelValues(clusterId, nodeName, labels)
		}
		clusterservice.SetMock(mock)

		// Create workload with topology
		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, clusterId)
		w.Topology = &workloadsApi.WorkloadSyncFieldsTopology{
			Name: topologyName,
		}
		apiResp, _ := workloadsService.SyncWorkloads(ctx, []string{}, workloadsApi.WorkloadsSyncRequest{
			ProjectId: DefaultProjectId,
			Workloads: []workloadsApi.WorkloadSyncFields{w},
		})
		Expect(apiResp.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))

		// Sync pods, each on its own node
		var pods []workloadsApi.PodSyncFields
		for nodeName := range nodeLabels {
			pod := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, clusterId)
			pod.NodeName = ptr.To(nodeName)
			pod.Name = nodeName + "-pod"
			pods = append(pods, pod)
		}
		apiResp, _ = podsService.SyncWorkloadsPods(ctx, workloadsApi.PodsSyncRequest{
			ProjectId: DefaultProjectId,
			ClusterId: clusterId,
			Pods:      pods,
		})
		Expect(apiResp.Code).To(Equal(http.StatusNoContent))

		return workloadId
	}

	AfterEach(func() {
		testing.CleanTables()
	})

	It("returns deepest level when all pods share the same topology path", func() {
		clusterId := uuid.NewString()
		workloadId := setupWorkloadWithTopologyPods(clusterId, map[string][]string{
			"node-1": {"zone-a", "rack-1", "host-1"},
			"node-2": {"zone-a", "rack-1", "host-1"},
			"node-3": {"zone-a", "rack-1", "host-1"},
		})

		apiResp, _ := workloadsService.GetWorkload(ctx, workloadId)
		Expect(apiResp.Code).To(Equal(http.StatusOK))
		wd := httpResponseBodyToWorkloadResponse(apiResp)
		Expect(wd.Topology).ToNot(BeNil())
		Expect(wd.Topology.ActualPlacement).ToNot(BeNil())
		Expect(wd.Topology.ActualPlacement.Level).To(Equal("host"))
		Expect(wd.Topology.ActualPlacement.Value).To(Equal("host-1"))
	})

	It("returns rack level when 1 out of 3 pods differs at the lowest level", func() {
		clusterId := uuid.NewString()
		workloadId := setupWorkloadWithTopologyPods(clusterId, map[string][]string{
			"node-1": {"zone-a", "rack-1", "host-1"},
			"node-2": {"zone-a", "rack-1", "host-1"},
			"node-3": {"zone-a", "rack-1", "host-2"},
		})

		apiResp, _ := workloadsService.GetWorkload(ctx, workloadId)
		Expect(apiResp.Code).To(Equal(http.StatusOK))
		wd := httpResponseBodyToWorkloadResponse(apiResp)
		Expect(wd.Topology).ToNot(BeNil())
		Expect(wd.Topology.ActualPlacement).ToNot(BeNil())
		Expect(wd.Topology.ActualPlacement.Level).To(Equal("rack"))
		Expect(wd.Topology.ActualPlacement.Value).To(Equal("rack-1"))
	})

	It("returns zone level when 2 pods differ at the second level", func() {
		clusterId := uuid.NewString()
		workloadId := setupWorkloadWithTopologyPods(clusterId, map[string][]string{
			"node-1": {"zone-a", "rack-1", "host-1"},
			"node-2": {"zone-a", "rack-2", "host-2"},
			"node-3": {"zone-a", "rack-3", "host-3"},
		})

		apiResp, _ := workloadsService.GetWorkload(ctx, workloadId)
		Expect(apiResp.Code).To(Equal(http.StatusOK))
		wd := httpResponseBodyToWorkloadResponse(apiResp)
		Expect(wd.Topology).ToNot(BeNil())
		Expect(wd.Topology.ActualPlacement).ToNot(BeNil())
		Expect(wd.Topology.ActualPlacement.Level).To(Equal("zone"))
		Expect(wd.Topology.ActualPlacement.Value).To(Equal("zone-a"))
	})

	It("returns nil when pods have empty topology label values", func() {
		clusterId := uuid.NewString()
		workloadId := setupWorkloadWithTopologyPods(clusterId, map[string][]string{
			"node-1": {},
			"node-2": {},
		})

		apiResp, _ := workloadsService.GetWorkload(ctx, workloadId)
		Expect(apiResp.Code).To(Equal(http.StatusOK))
		wd := httpResponseBodyToWorkloadResponse(apiResp)
		Expect(wd.Topology).ToNot(BeNil())
		Expect(wd.Topology.ActualPlacement).To(BeNil())
	})

	It("returns nil when all pods differ at all levels", func() {
		clusterId := uuid.NewString()
		workloadId := setupWorkloadWithTopologyPods(clusterId, map[string][]string{
			"node-1": {"zone-a", "rack-1", "host-1"},
			"node-2": {"zone-b", "rack-2", "host-2"},
			"node-3": {"zone-c", "rack-3", "host-3"},
		})

		apiResp, _ := workloadsService.GetWorkload(ctx, workloadId)
		Expect(apiResp.Code).To(Equal(http.StatusOK))
		wd := httpResponseBodyToWorkloadResponse(apiResp)
		Expect(wd.Topology).ToNot(BeNil())
		Expect(wd.Topology.ActualPlacement).To(BeNil())
	})
})
