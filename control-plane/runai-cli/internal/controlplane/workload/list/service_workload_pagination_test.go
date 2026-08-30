package list

import (
	"fmt"
	"time"

	apilist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/structs"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cpworkloadmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane/workload"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var _ = Describe("Workload Paginator", func() {
	var (
		ctrl     *gomock.Controller
		mockpage *cpworkloadmocks.MockPaginator
		service  *Pager
		fb       *filter.Builder
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockpage = cpworkloadmocks.NewMockPaginator(ctrl)
		service = NewPager(mockpage)
		fb = filter.NewBuilder()
	})

	Context("Pager", func() {
		It("filter is nil", func() {
			response, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 10}, nil, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("filter is nil"))
			Expect(next).To(BeNil())
			Expect(response).To(BeNil())
		})

		It("page size is not positive", func() {
			response, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{}, fb, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("page size must be a positive number"))
			Expect(next).To(BeNil())
			Expect(response).To(BeNil())
		})

		It("page is nil", func() {
			mockpage.EXPECT().List(gomock.Any(), int32(11), int32(0), fb.StringList(), false).Return(nil, nil)
			response, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 10}, fb, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("page is nil"))
			Expect(next).To(BeNil())
			Expect(response).To(BeNil())
		})

		It("singlePage return error", func() {
			mockpage.EXPECT().List(gomock.Any(), int32(11), int32(0), fb.StringList(), false).Return(&openapi.GetWorkloads200Response{}, fmt.Errorf("bla"))
			response, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 10}, fb, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("bla"))
			Expect(next).To(BeNil())
			Expect(response).To(BeNil())
		})

		It("singlePage without next", func() {
			response := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{{Id: "1", CreatedAt: time.Unix(10, 0)}}}
			mockpage.EXPECT().List(gomock.Any(), int32(11), int32(0), fb.StringList(), false).Return(response, nil)
			workloadList, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 10}, fb, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(next).To(BeNil())
			Expect(workloadList).ToNot(BeNil())
			Expect(workloadList).To(HaveLen(1))
		})

		It("singlePage with next, but no create timestamp", func() {
			response := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{{Id: "1"}, {Id: "2"}}}
			mockpage.EXPECT().List(gomock.Any(), int32(2), int32(0), fb.StringList(), false).Return(response, nil)
			workloadList, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 1}, fb, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("next token cannot be with 0 timestamp"))
			Expect(next).To(BeNil())
			Expect(workloadList).To(HaveLen(0))
		})

		It("singlePage with next, but no create timestamp", func() {
			response := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "1"},
				{Id: "2"},
				{Id: "3", CreatedAt: time.Unix(10, 0)},
			}}
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList(), false).Return(response, nil)
			workloadList, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 2}, fb, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(next).ToNot(BeNil())
			Expect(next.String()).To(HavePrefix("1970-01-01T")) // unit + 10 sec
			Expect(next.String()).To(ContainSubstring(":00:10"))
			Expect(workloadList).To(HaveLen(2))
		})
	})

	Context("multiPages", func() {
		It("multiPages: response is nil", func() {
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList(), false).Return(nil, nil)
			response, next, err := service.multiPages(transaction.NewTxn(), &apilist.PageOptions{PageSize: 2, Offset: 0}, fb, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("page is nil"))
			Expect(next).To(BeNil())
			Expect(response).To(BeNil())
		})

		It("multiPages: responses has the same timestamp", func() {
			response1 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "1"},
				{Id: "2"},
				{Id: "3", CreatedAt: time.Unix(10, 0)},
			}}
			response2 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "3"},
				{Id: "4"},
				{Id: "5", CreatedAt: time.Unix(10, 0)},
			}}
			next, err := structs.NextTokenFromTimestamp(time.Unix(10, 0))
			Expect(err).ToNot(HaveOccurred())
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(2), fb.StringList(), false).Return(response1, nil)
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(2), fb.WithCreatedAt(next.GetTimestamp()).StringList(), false).Return(response2, nil)
			// if we fail to catch this, the error will be the same call matched twice - should not reach it.
			response, next, err := service.multiPages(transaction.NewTxn(), &apilist.PageOptions{PageSize: 2, Offset: 2}, fb.WithCreatedAt(time.Time{}), false)
			Expect(err).To(HaveOccurred())
			// local uses UTC +2 and the make test docker uses UTC, so there is a different value to the timestamp
			Expect(err.Error()).To(HavePrefix("next token timestamp (1970-01-01T"))
			Expect(err.Error()).To(ContainSubstring("must be after current timestamp (1970-01-01"))
			Expect(err.Error()).To(ContainSubstring(":00:10"))
			Expect(next).To(BeNil())
			Expect(response).To(BeNil()) // error will return empty response
		})

		It("multiPages: new response has sooner timestamp", func() {
			response1 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "1"},
				{Id: "2"},
				{Id: "3", CreatedAt: time.Unix(10, 0)},
			}}
			response2 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "3"},
				{Id: "4"},
				{Id: "5", CreatedAt: time.Unix(10, 0)},
			}}
			next, err := structs.NextTokenFromTimestamp(time.Unix(10, 0))
			Expect(err).ToNot(HaveOccurred())
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(2), fb.StringList(), false).Return(response1, nil)
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(2), fb.WithCreatedAt(next.GetTimestamp()).StringList(), false).Return(response2, nil)
			// if we fail to catch this, the error will be the same call matched twice - should not reach it.
			response, next, err := service.multiPages(transaction.NewTxn(), &apilist.PageOptions{PageSize: 2, Offset: 2}, fb.WithCreatedAt(time.Time{}), false)
			Expect(err).To(HaveOccurred())
			// local uses UTC +2 and the make test docker uses UTC, so there is a different value to the timestamp
			Expect(err.Error()).To(HavePrefix("next token timestamp (1970-01-01T"))
			Expect(err.Error()).To(ContainSubstring("must be after current timestamp (1970-01-01"))
			Expect(next).To(BeNil())
			Expect(response).To(BeNil()) // error will return empty response
		})

		It("multiPages: reached max max items", func() {
			response1 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "1"},
				{Id: "2"},
				{Id: "3", CreatedAt: time.Unix(10, 0)},
			}}
			response2 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "3"},
				{Id: "4", CreatedAt: time.Unix(11, 0)},
				{Id: "5", CreatedAt: time.Unix(13, 0)},
			}}
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList(), false).Return(response1, nil)

			next, err := structs.NextTokenFromTimestamp(time.Unix(10, 0))
			Expect(err).ToNot(HaveOccurred())
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.WithCreatedAt(next.GetTimestamp()).StringList(), false).Return(response2, nil)

			// if we fail to catch this, the error will be the same call matched twice - should not reach it.
			workloadList, next, err := service.multiPages(transaction.NewTxn(), &apilist.PageOptions{MaxItems: 4, PageSize: 2}, fb.WithCreatedAt(time.Time{}), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(next).ToNot(BeNil())
			// local uses UTC +2 and the make test docker uses UTC, so there is a different value to the timestamp
			Expect(next.String()).To(HavePrefix("1970-01-01T"))
			Expect(next.String()).To(ContainSubstring(":13")) // equal to last of response3 (unit test container uses UTC so time is shift it tests)
			Expect(workloadList).To(HaveLen(4))
		})

		It("multiPages: fetched all", func() {
			response1 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "1"},
				{Id: "2"},
				{Id: "3", CreatedAt: time.Unix(10, 0)},
			}}
			response2 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "3"},
				{Id: "4"},
				{Id: "5", CreatedAt: time.Unix(11, 0)},
			}}
			response3 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "5"},
				{Id: "6"},
			}}
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList(), false).Return(response1, nil)

			next, err := structs.NextTokenFromTimestamp(time.Unix(10, 0))
			Expect(err).ToNot(HaveOccurred())
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.WithCreatedAt(next.GetTimestamp()).StringList(), false).Return(response2, nil)

			next, err = structs.NextTokenFromTimestamp(time.Unix(11, 0))
			Expect(err).ToNot(HaveOccurred())
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.WithCreatedAt(next.GetTimestamp()).StringList(), false).Return(response3, nil)
			// if we fail to catch this, the error will be the same call matched twice - should not reach it.
			workloadList, next, err := service.multiPages(transaction.NewTxn(), &apilist.PageOptions{MaxItems: 6, PageSize: 2}, fb.WithCreatedAt(time.Time{}), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(next).To(BeNil())
			Expect(workloadList).To(HaveLen(6))
		})
	})

	Context("All", func() {
		It("multiPages: fetched all", func() {
			response1 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "1"},
				{Id: "2"},
				{Id: "3", CreatedAt: time.Unix(10, 0)},
			}}
			response2 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "3"},
				{Id: "4"},
				{Id: "5", CreatedAt: time.Unix(11, 0)},
			}}
			response3 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "5"},
			}}
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList(), false).Return(response1, nil)

			next, err := structs.NextTokenFromTimestamp(time.Unix(10, 0))
			Expect(err).ToNot(HaveOccurred())
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.WithCreatedAt(next.GetTimestamp()).StringList(), false).Return(response2, nil)

			next, err = structs.NextTokenFromTimestamp(time.Unix(11, 0))
			Expect(err).ToNot(HaveOccurred())
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.WithCreatedAt(next.GetTimestamp()).StringList(), false).Return(response3, nil)
			// if we fail to catch this, the error will be the same call matched twice - should not reach it.
			response, next, err := service.multiPages(transaction.NewTxn(), &apilist.PageOptions{PageSize: 2}, fb.WithCreatedAt(time.Time{}), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(next).To(BeNil())
			Expect(response).To(HaveLen(5))
		})

		It("All: error on third response", func() {
			response1 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "1"},
				{Id: "2"},
				{Id: "3", CreatedAt: time.Unix(10, 0)},
			}}
			response2 := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				{Id: "3"},
				{Id: "4"},
				{Id: "5", CreatedAt: time.Unix(11, 0)},
			}}
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList(), false).Return(response1, nil)

			next, err := structs.NextTokenFromTimestamp(time.Unix(10, 0))
			Expect(err).ToNot(HaveOccurred())
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.WithCreatedAt(next.GetTimestamp()).StringList(), false).Return(response2, nil)

			next, err = structs.NextTokenFromTimestamp(time.Unix(11, 0))
			Expect(err).ToNot(HaveOccurred())
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.WithCreatedAt(next.GetTimestamp()).StringList(), false).Return(nil, fmt.Errorf("bla"))
			// if we fail to catch this, the error will be the same call matched twice - should not reach it.
			response, next, err := service.multiPages(transaction.NewTxn(), &apilist.PageOptions{PageSize: 2}, fb.WithCreatedAt(time.Time{}), false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("bla"))
			Expect(next).To(BeNil())
			Expect(response).To(BeNil())
		})
	})
})
