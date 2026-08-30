package datasource

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cpmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

func TestDatasource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Datasource Suite")
}

var _ = Describe("GetId", func() {
	var (
		mockCtrl   *gomock.Controller
		mockLister *cpmocks.MockDatasourceDescribeLister
		txn        *transaction.Txn
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockLister = cpmocks.NewMockDatasourceDescribeLister(mockCtrl)
		SetClient(mockLister)
		txn = transaction.NewTxn()
	})

	AfterEach(func() {
		mockCtrl.Finish()
		SetClient(nil)
	})

	Describe("Datasource found by name", func() {
		It("should return datasource ID when found", func() {
			ds := newTestDatasource("my-datasource", "ds-uuid-123")
			listResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{ds},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			id, err := GetId(txn, "http://localhost", "token", "my-datasource", "pvc", "cluster-1", "123")

			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal("ds-uuid-123"))
		})

		It("should work with type filter", func() {
			ds := newTestDatasource("s3-datasource", "ds-uuid-456")
			listResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{ds},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			id, err := GetId(txn, "http://localhost", "token", "s3-datasource", "s3", "cluster-1", "123")

			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal("ds-uuid-456"))
		})
	})

	Describe("Datasource not found", func() {
		It("should return error when datasource not found", func() {
			listResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{},
			}
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			id, err := GetId(txn, "http://localhost", "token", "nonexistent", "pvc", "cluster-1", "123")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(id).To(BeEmpty())
		})
	})

	Describe("API error", func() {
		It("should return error when API fails", func() {
			mockLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))

			id, err := GetId(txn, "http://localhost", "token", "my-datasource", "pvc", "cluster-1", "123")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to search for datasource"))
			Expect(id).To(BeEmpty())
		})
	})
})

func newTestDatasource(name, id string) openapi.DatasourceListResponseEntry {
	return openapi.DatasourceListResponseEntry{
		Meta: openapi.AssetMeta{
			Id:   id,
			Name: name,
		},
	}
}
