package services_crud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	clientsclusterservice "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/clusterservice"

	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/org_unit"

	bc_mocks "github.com/run-ai/runai/control-plane/workloads/mocks/pkg/clients/backend"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"

	"github.com/run-ai/runai/control-plane/workloads/pkg/services"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_integration"
	"github.com/run-ai/runai/control-plane/workloads/pkg/testing_crud"

	policies "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

func httpResponseBodyToListUpdatesResponse(httpResponse api.ImplResponse) api.WorkloadsSyncResponse {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	response := &api.WorkloadsSyncResponse{}
	err = json.Unmarshal(bytes, response)
	Expect(err).To(BeNil())
	return *response
}

var _ = Describe("sync workloads", func() {
	var ctx context.Context
	var randoms *testutils.RandomValues
	var differentCluster *testutils.RandomValues

	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}

	workspacesService := NewWorkspacesService(mockTenantSettingsClient)
	trainingsService := NewTrainingsService(mockTenantSettingsClient)

	BeforeEach(func() {
		testing.CleanTables()
		CleanupCrudTable()
		RestartCrudUpdateCounter()

		randoms = testutils.NewRandomValues()
		differentCluster = randoms.CloneWithSpecificCluster(uuid.New().String())
		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

		authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms).With(differentCluster)
		authMock.AllowedProjectsToContext(&ctx)

		orgunit.SetMock(authMock.OrgUnitMock())

		mockClient := checkerClusterClient.NewSimpleMockClient()

		mockClient.WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionNativeSubmission, nil))
		mockClient.WithCluster(differentCluster.ClusterId, testing.BuildDisplayedCluster(differentCluster.ClusterId, testing.TestClusterVersionNativeSubmission, nil))

		checkerClusterClient.SetMock(mockClient)

		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		policyMockUpdateTime := time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))

		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))
	})

	AfterEach(func() {
		testing.CleanTables()
		CleanupCrudTable()
		RestartCrudUpdateCounter()
	})

	Context("list updates - only updates with DesiredPhase=Deleted are to be deleted", func() {
		It("cleanup obsolete updates and returns the correct number of workload updates", func() {
			// creating 5 workloads
			for i := 0; i < 5; i++ {
				request := api.WorkspaceCreationRequest{
					Name:      fmt.Sprintf("ws-%d", i),
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.WorkspaceSpecSpec{
						Image: convutils.Ref("ubuntu"),
					},
				}

				createResponse, _ := workspacesService.CreateWorkspace(ctx, request)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))

				// deleting only the even indexes meaning 3 would change their desiredstatus to deleted
				if i%2 == 0 {
					createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
					deleteResponse, _ := workspacesService.DeleteWorkspace(ctx, createPayload.WorkloadId)
					Expect(deleteResponse.Code).To(Equal(http.StatusAccepted))
				}
			}

			syncConfig := config.WorkloadsSyncConfig{
				BulkSize: 40,
			}

			// First request - asking for afterUpdate=0 meaning we will receive all 5 updates
			listHttpRes, listErr := ListUpdates(ctx, Workspace, 0, randoms.ClusterId, &syncConfig, nil)
			Expect(listErr).To(BeNil())
			resp := httpResponseBodyToListUpdatesResponse(listHttpRes)

			Expect(len(resp.Workloads)).To(Equal(5))

			// Second request - asking for afterUpdate=resp.ResponseInfo.HighestUpdate+1 meaning all updates with desiredStats= Deleted would now be considered obsolete
			highestUpdateCount := resp.ResponseInfo.HighestUpdate
			secondReqlistHttpRes, secondReqlistErr := ListUpdates(ctx, Workspace, highestUpdateCount+1, randoms.ClusterId, &syncConfig, nil)
			Expect(secondReqlistErr).To(BeNil())

			secondResp := httpResponseBodyToListUpdatesResponse(secondReqlistHttpRes)
			Expect(secondResp.RequestInfo.AfterUpdate).To(Equal(int64(highestUpdateCount + 1)))
			Expect(len(secondResp.Workloads)).To(Equal(0))

			// Third request - verifying the obsolete records have been deleted by requesting afterUpdate=0 again (all records with desiredState=Deleted would be deleted (only 2 remaining))
			thirdListHttpRes, thirdListErr := ListUpdates(ctx, Workspace, 0, randoms.ClusterId, &syncConfig, nil)
			Expect(thirdListErr).To(BeNil())

			thirdResp := httpResponseBodyToListUpdatesResponse(thirdListHttpRes)
			Expect(thirdResp.RequestInfo.AfterUpdate).To(Equal(int64(0)))

			Expect(len(thirdResp.Workloads)).To(Equal(2))
			Expect(thirdResp.Workloads[0].Workload.DesiredPhase).ToNot(Equal(api.WORKLOADDESIREDPHASE_DELETED))
			Expect(thirdResp.Workloads[1].Workload.DesiredPhase).ToNot(Equal(api.WORKLOADDESIREDPHASE_DELETED))
		})
	})

	Context("list updates - simple case single cluster workspace only", func() {
		It("cleanup obsolete updates and returns the correct number of workload updates", func() {
			// creating 5 workloads
			for i := 0; i < 5; i++ {
				request := api.WorkspaceCreationRequest{
					Name:      fmt.Sprintf("ws-%d", i),
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.WorkspaceSpecSpec{
						Image: convutils.Ref("ubuntu"),
					},
				}

				createResponse, _ := workspacesService.CreateWorkspace(ctx, request)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))

				// initiating delete for the workload (only workload updates with DesiredPhase of deleted would be considered obsolete)
				createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
				deleteResponse, _ := workspacesService.DeleteWorkspace(ctx, createPayload.WorkloadId)
				Expect(deleteResponse.Code).To(Equal(http.StatusAccepted))
			}

			syncConfig := config.WorkloadsSyncConfig{
				BulkSize: 40,
			}

			// First request - asking for afterUpdate=0 meaning we will receive all 5 updates
			listHttpRes, listErr := ListUpdates(ctx, Workspace, 0, randoms.ClusterId, &syncConfig, nil)
			Expect(listErr).To(BeNil())
			resp := httpResponseBodyToListUpdatesResponse(listHttpRes)

			Expect(len(resp.Workloads)).To(Equal(5))

			// Second request - asking for afterUpdate=resp.ResponseInfo.HighestUpdate-1 meaning all updates with updateCount below the most recent avaiable update would now be considered obsolete and will be deleted from the table
			highestUpdateCount := resp.ResponseInfo.HighestUpdate
			secondReqlistHttpRes, secondReqlistErr := ListUpdates(ctx, Workspace, highestUpdateCount-1, randoms.ClusterId, &syncConfig, nil)
			Expect(secondReqlistErr).To(BeNil())

			secondResp := httpResponseBodyToListUpdatesResponse(secondReqlistHttpRes)
			Expect(secondResp.RequestInfo.AfterUpdate).To(Equal(int64(highestUpdateCount - 1)))

			Expect(len(secondResp.Workloads)).To(Equal(1))
			Expect(secondResp.Workloads[0].UpdateCount).To(Equal(int64(highestUpdateCount)))

			// Third request - verifying the obsolete records have been deleted by requesting afterUpdate=0 again (all records except for the last 2 should be deleted)
			thirdListHttpRes, thirdListErr := ListUpdates(ctx, Workspace, 0, randoms.ClusterId, &syncConfig, nil)
			Expect(thirdListErr).To(BeNil())

			thirdResp := httpResponseBodyToListUpdatesResponse(thirdListHttpRes)
			Expect(thirdResp.RequestInfo.AfterUpdate).To(Equal(int64(0)))

			Expect(len(thirdResp.Workloads)).To(Equal(1))
			Expect(thirdResp.Workloads[0].UpdateCount).To(Equal(int64(thirdResp.ResponseInfo.HighestUpdate)))
		})
	})

	Context("list updates - different workload types", func() {
		It("cleanup obsolete updates and returns the correct number of workload updates for specific workload type", func() {
			// creating 2 workspaces and 2 trainings on the same cluster
			for i := 0; i < 2; i++ {
				wsRequest := api.WorkspaceCreationRequest{
					Name:      fmt.Sprintf("ws-%d", i),
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.WorkspaceSpecSpec{
						Image: convutils.Ref("ubuntu"),
					},
				}
				wsCreateResponse, _ := workspacesService.CreateWorkspace(ctx, wsRequest)
				Expect(wsCreateResponse.Code).To(Equal(http.StatusAccepted))

				// initiating delete for the workload (only workload updates with DesiredPhase of deleted would be considered obsolete)
				createWsPayload := wsCreateResponse.Body.(workloadsApi.SupersetWorkload)
				deleteWsResponse, _ := workspacesService.DeleteWorkspace(ctx, createWsPayload.WorkloadId)
				Expect(deleteWsResponse.Code).To(Equal(http.StatusAccepted))

				trRequest := api.TrainingCreationRequest{
					Name:      fmt.Sprintf("tr-%d", i),
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.TrainingSpecSpec{
						Image: convutils.Ref("ubuntu"),
					},
				}
				trCreateResponse, _ := trainingsService.CreateTraining(ctx, trRequest)
				Expect(trCreateResponse.Code).To(Equal(http.StatusAccepted))

				// initiating delete for the workload (only workload updates with DesiredPhase of deleted would be considered obsolete)
				createTrPayload := trCreateResponse.Body.(workloadsApi.SupersetWorkload)
				deleteTrResponse, _ := trainingsService.DeleteTraining(ctx, createTrPayload.WorkloadId)
				Expect(deleteTrResponse.Code).To(Equal(http.StatusAccepted))
			}

			syncConfig := config.WorkloadsSyncConfig{
				BulkSize: 40,
			}

			// First request - asking for afterUpdate=0 for type workspace meaning we will receive 2 updates
			listHttpRes, listErr := ListUpdates(ctx, Workspace, 0, randoms.ClusterId, &syncConfig, nil)
			Expect(listErr).To(BeNil())

			resp := httpResponseBodyToListUpdatesResponse(listHttpRes)
			Expect(len(resp.Workloads)).To(Equal(2))

			// Second request - asking for afterUpdate=resp.ResponseInfo.HighestUpdate + 1 meaning all existing updates for workspace in the table are now considered obsolete and will be deleted
			highestUpdateCount := resp.ResponseInfo.HighestUpdate
			secondReqlistHttpRes, secondReqlistErr := ListUpdates(ctx, Workspace, highestUpdateCount+1, randoms.ClusterId, &syncConfig, nil)
			Expect(secondReqlistErr).To(BeNil())

			secondResp := httpResponseBodyToListUpdatesResponse(secondReqlistHttpRes)
			Expect(secondResp.RequestInfo.AfterUpdate).To(Equal(int64(highestUpdateCount + 1)))
			Expect(len(secondResp.Workloads)).To(Equal(0))

			// Third request - verifying the obsolete records have been deleted for type workspace
			thirdListHttpRes, thirdListErr := ListUpdates(ctx, Workspace, 0, randoms.ClusterId, &syncConfig, nil)
			Expect(thirdListErr).To(BeNil())

			thirdResp := httpResponseBodyToListUpdatesResponse(thirdListHttpRes)
			Expect(thirdResp.RequestInfo.AfterUpdate).To(Equal(int64(0)))
			Expect(len(thirdResp.Workloads)).To(Equal(0))

			// checking that training updates are unaffected
			trainingReqlistHttpRes, trainingReqlistErr := ListUpdates(ctx, Training, 0, randoms.ClusterId, &syncConfig, nil)
			Expect(trainingReqlistErr).To(BeNil())

			trainingResp := httpResponseBodyToListUpdatesResponse(trainingReqlistHttpRes)
			Expect(trainingResp.RequestInfo.AfterUpdate).To(Equal(int64(0)))
			Expect(len(trainingResp.Workloads)).To(Equal(2))
		})
	})
	Context("list updates - multi cluster", func() {
		It("cleanup obselete updates and returns the correct number of workload updates for specific cluster", func() {
			randomSets := [2]*testutils.RandomValues{randoms, differentCluster}

			// creating 3 trainings for each of the clusters
			for _, randSet := range randomSets {
				for i := 0; i < 3; i++ {
					trRequest := api.TrainingCreationRequest{
						Name:      fmt.Sprintf("tr-%d", i),
						ProjectId: randSet.ProjectIdStr,
						ClusterId: randSet.ClusterId,
						Spec: api.TrainingSpecSpec{
							Image: convutils.Ref("ubuntu"),
						},
					}

					org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randSet.ProjectIdStr, GetProjectResponse(*randSet)))
					clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randSet.ClusterId))

					trCreateResponse, _ := trainingsService.CreateTraining(ctx, trRequest)
					Expect(trCreateResponse.Code).To(Equal(http.StatusAccepted))

					// initiating delete for the workload (only workload updates with DesiredPhase of deleted would be considered obsolete)
					createTrPayload := trCreateResponse.Body.(workloadsApi.SupersetWorkload)
					deleteTrResponse, _ := trainingsService.DeleteTraining(ctx, createTrPayload.WorkloadId)
					Expect(deleteTrResponse.Code).To(Equal(http.StatusAccepted))
				}
			}

			syncConfig := config.WorkloadsSyncConfig{
				BulkSize: 40,
			}

			// First request - asking for afterUpdate=0 for the first cluster we will receive 3 updates
			listHttpRes, listErr := ListUpdates(ctx, Training, 0, randoms.ClusterId, &syncConfig, nil)
			Expect(listErr).To(BeNil())

			resp := httpResponseBodyToListUpdatesResponse(listHttpRes)
			Expect(len(resp.Workloads)).To(Equal(3))

			// Second request - asking for afterUpdate=resp.ResponseInfo.HighestUpdate + 1 for the first cluster meaning all existing training updates for the specific cluster are now considered obsolete and will be deleted
			highestUpdateCount := resp.ResponseInfo.HighestUpdate
			secondReqlistHttpRes, secondReqlistErr := ListUpdates(ctx, Training, highestUpdateCount+1, randoms.ClusterId, &syncConfig, nil)
			Expect(secondReqlistErr).To(BeNil())

			secondResp := httpResponseBodyToListUpdatesResponse(secondReqlistHttpRes)
			Expect(secondResp.RequestInfo.AfterUpdate).To(Equal(int64(highestUpdateCount + 1)))
			Expect(len(secondResp.Workloads)).To(Equal(0))

			// Third request - verifying the obsolete records have been deleted in the first cluster
			thirdListHttpRes, thirdListErr := ListUpdates(ctx, Training, 0, randoms.ClusterId, &syncConfig, nil)
			Expect(thirdListErr).To(BeNil())

			thirdResp := httpResponseBodyToListUpdatesResponse(thirdListHttpRes)
			Expect(thirdResp.RequestInfo.AfterUpdate).To(Equal(int64(0)))
			Expect(len(thirdResp.Workloads)).To(Equal(0))

			// checking records of the second cluster training updates are unaffected
			secondClusterHttpRes, secondClusterListErr := ListUpdates(ctx, Training, 0, differentCluster.ClusterId, &syncConfig, nil)
			Expect(secondClusterListErr).To(BeNil())

			secondClusterResp := httpResponseBodyToListUpdatesResponse(secondClusterHttpRes)
			Expect(secondClusterResp.RequestInfo.AfterUpdate).To(Equal(int64(0)))
			Expect(len(secondClusterResp.Workloads)).To(Equal(3))
		})
	})

	Context("list updates - workloads that failed to be created in the cluster", func() {
		It("delete workloads from CP with phase Failed or Deleting if they do not exist in the cluster", func() {
			workloadsService := services.NewWorkloadsService()

			var workloadIds []string

			// create 3 workloads
			for i := 0; i < 3; i++ {
				request := api.WorkspaceCreationRequest{
					Name:      fmt.Sprintf("ws-%d", i),
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.WorkspaceSpecSpec{
						Image: convutils.Ref("ubuntu"),
					},
				}
				createResponse, _ := workspacesService.CreateWorkspace(ctx, request)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))

				workloadId := createResponse.Body.(workloadsApi.SupersetWorkload).WorkloadId
				workloadIds = append(workloadIds, workloadId)

				// set workload status to failed depending on the index
				if i%2 == 0 {
					// Use a future time to ensure the phase update is always applied (bypasses the "phase_updated_at < updateTime" check)
					phaseChangeErr := services_integration.UpdateWorkloadFailure(ctx, workloadId, "something went wrong", time.Now().Add(1*time.Minute))
					Expect(phaseChangeErr).To(BeNil())
				}

			}

			listReq, err := workloadsService.GetWorkloads(ctx, false, 0, 50, "", "", []string{}, "")
			Expect(err).To(BeNil())
			Expect(listReq.Code).To(Equal(http.StatusOK))
			Expect(len(listReq.Body.(api.GetWorkloads200Response).Workloads)).To(Equal(3))

			// mock the cluster to return an error when trying to get the failed workloads
			syncConfig := config.WorkloadsSyncConfig{
				BulkSize: 40,
			}

			syncReport := api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:                workloadIds[0],
						UpdateCount:       0,
						UpdateAttemptTime: time.Now(),
						Message:           convutils.Ref("InteractiveWorkload \"ws-0\" not found"),
					},
					{
						Id:                workloadIds[2],
						UpdateCount:       0,
						UpdateAttemptTime: time.Now(),
						Message:           convutils.Ref("InteractiveWorkload \"ws-2\" not found"),
					},
				},
			}

			_, listErr := ListUpdates(ctx, Workspace, 0, randoms.ClusterId, &syncConfig, &syncReport)
			Expect(listErr).To(BeNil())

			// checking the workload was deleted from the CP
			listReqPostClusterError, err := workloadsService.GetWorkloads(ctx, false, 0, 50, "", "", []string{}, "")

			Expect(err).To(BeNil())
			Expect(listReqPostClusterError.Code).To(Equal(http.StatusOK))

			Expect(len(listReqPostClusterError.Body.(api.GetWorkloads200Response).Workloads)).To(Equal(1))
			Expect(listReqPostClusterError.Body.(api.GetWorkloads200Response).Workloads[0].Id).To(Equal(workloadIds[1]))
		})
	})
})
