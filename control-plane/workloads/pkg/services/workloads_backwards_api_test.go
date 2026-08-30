package services

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
)

var _ = Describe("Workloads Backwards API tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()

	AfterAll(func() {
		testing.CleanTables()
	})

	Context("SyncJobsBackwards", func() {
		It("should return 201 Created", func() {
			apiResponse, err := workloadsService.SyncJobsBackwards(ctx, testing.ClusterUuid1, api.SyncJobsBackwardsRequest{})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))
		})
	})

	Context("SyncPodsBackwards", func() {
		It("should return 200 OK with empty ids", func() {
			apiResponse, err := workloadsService.SyncPodsBackwards(ctx, testing.ClusterUuid1, api.SyncJobsBackwardsRequest{})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			response := apiResponse.Body.(api.SyncPodsBackwards200Response)
			Expect(response.Ids).To(BeEmpty())
		})
	})

	Context("SyncPodgroupsBackwards", func() {
		It("should return 200 OK with empty ids", func() {
			apiResponse, err := workloadsService.SyncPodgroupsBackwards(ctx, testing.ClusterUuid1, "", api.SyncJobsBackwardsRequest{})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			response := apiResponse.Body.(api.SyncPodsBackwards200Response)
			Expect(response.Ids).To(BeEmpty())
		})

		It("should return 200 OK with projectName parameter", func() {
			apiResponse, err := workloadsService.SyncPodgroupsBackwards(ctx, testing.ClusterUuid1, "test-project", api.SyncJobsBackwardsRequest{})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			response := apiResponse.Body.(api.SyncPodsBackwards200Response)
			Expect(response.Ids).To(BeEmpty())
		})
	})

	Context("PostSyncPodsBackwards", func() {
		It("should return 201 Created", func() {
			apiResponse, err := workloadsService.PostSyncPodsBackwards(ctx, testing.ClusterUuid1, api.SyncJobsBackwardsRequest{})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))
		})
	})

	Context("PostSyncPodgroupsBackwards", func() {
		It("should return 200 OK", func() {
			apiResponse, err := workloadsService.PostSyncPodgroupsBackwards(ctx, testing.ClusterUuid1, api.SyncJobsBackwardsRequest{})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
		})
	})
})
