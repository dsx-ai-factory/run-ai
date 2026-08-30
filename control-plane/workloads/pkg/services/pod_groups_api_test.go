package services

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"k8s.io/utils/ptr"
	"net/http"
)

var _ = Describe("Pod groups API tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	podGroupsService := NewPodGroupsService()
	revisionsService := NewRevisionsService()
	workloadsService := NewWorkloadsService()

	Context("Sync pod groups", func() {
		BeforeAll(func() {
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, testing.TestUuid1, SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.UpsertWorkload(ctx, testing.TestUuid3, SetupBaseWorkload(testing.TestUuid3, DefaultProjectId, testing.ClusterUuid1))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(testing.TestUuid2, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
				},
			})
		})

		AfterAll(func() {
			testing.CleanTables()
		})

		It("Sync multiple with errors", func() {
			res, err := podGroupsService.SyncPodGroups(ctx, api.PodGroupsSyncRequest{
				PodGroups: []api.PodGroupsSyncFields{
					{
						Name:       "pg-0",
						Id:         uuid.NewString(),
						WorkloadId: testing.TestUuid1,
					},
					{
						Name:       "pg-1",
						Id:         uuid.NewString(),
						WorkloadId: testing.TestUuid1,
						RevisionId: ptr.To(testing.TestUuid2),
					},
					{
						Name:       "pg-2",
						Id:         uuid.NewString(),
						WorkloadId: testing.TestUuid3,
					},
					{
						Name:       "pg-3",
						Id:         "wrong_id",
						WorkloadId: uuid.NewString(),
					},
					{
						Name: "pg-4",
						Id:   uuid.NewString(),
					},
				},
			})

			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusMultiStatus))
			convert := res.Body.(api.SyncPodGroups207Response)
			Expect(len(convert.FailedPodGroups)).To(Equal(2))
			Expect(len(convert.SuccessfullySyncedPodGroups)).To(Equal(3))

			workloadChildren, apiErr := dao.NewWorkloadsDao(runai_config.Get[*config.ServerConfig]()).DbGetChildren(ctx, uuid.MustParse(testing.TestUuid1), ptr.To([]string{"PodGroup"}))
			Expect(apiErr).To(BeNil())
			Expect(workloadChildren).To(Not(BeNil()))
			Expect(len(*workloadChildren)).To(Equal(2))

			revisionChildren, apiErr := dao.DbGetRevisionChildren(ctx, testing.TestUuid2, ptr.To([]string{"PodGroup"}))
			Expect(apiErr).To(BeNil())
			Expect(revisionChildren).To(Not(BeNil()))
			Expect(len(*revisionChildren)).To(Equal(1))
		})
	})
})
