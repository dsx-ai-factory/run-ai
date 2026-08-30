package describe_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cpmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/compute/describe"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

var _ = Describe("FindByName", func() {
	var (
		mockCtrl   *gomock.Controller
		mockLister *cpmocks.MockComputeDescribeLister
		ctx        context.Context
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockLister = cpmocks.NewMockComputeDescribeLister(mockCtrl)
		ctx = context.Background()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("when compute exists", func() {
		It("should find compute by name with project filter", func() {
			compute := newTestComputeForList("test-compute", "compute-id-1", "Project", 8, "80GB", "128GB")
			listResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{compute},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := describe.FindByName(ctx, mockLister, "test-compute", &projectID, &clusterID)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Name).To(Equal("test-compute"))
			Expect(result.Meta.Id).To(Equal("compute-id-1"))
		})

		It("should find compute by name with cluster filter only", func() {
			compute := newTestComputeForList("test-compute", "compute-id-1", "Cluster", 8, "80GB", "128GB")
			listResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{compute},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(listResponse, nil)

			clusterID := "cluster-123"
			result, err := describe.FindByName(ctx, mockLister, "test-compute", nil, &clusterID)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Name).To(Equal("test-compute"))
		})
	})

	Context("when compute does not exist", func() {
		It("should return error when compute not found by name", func() {
			listResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "nonexistent-compute", &projectID, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(result).To(BeNil())
		})

		It("should return error when name doesn't match any entry", func() {
			compute := newTestComputeForList("other-compute", "compute-id-1", "Project", 8, "80GB", "128GB")
			listResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{compute},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "test-compute", &projectID, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(result).To(BeNil())
		})
	})

	Context("when list fails", func() {
		It("should return error when list API fails", func() {
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(nil, errors.New("list error"))

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "test-compute", &projectID, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to search for compute"))
			Expect(result).To(BeNil())
		})
	})

	Context("when multiple computes match", func() {
		It("should return the first matching compute", func() {
			compute1 := newTestComputeForList("test-compute", "compute-id-1", "Project", 8, "80GB", "128GB")
			compute2 := newTestComputeForList("test-compute", "compute-id-2", "Cluster", 4, "40GB", "64GB")
			listResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{compute1, compute2},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "test-compute", &projectID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Id).To(Equal("compute-id-1"))
		})
	})

	Context("edge cases", func() {
		It("should return error when name is empty and no matches", func() {
			listResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "", &projectID, nil)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})
})

// Helper function to create test compute assets
func newTestComputeForList(name, id, scope string, gpuDevices int32, gpuMemory, cpuMemory string) openapi.ComputeAsset {
	spec := openapi.ComputeAssetSpec{}

	if gpuDevices > 0 {
		spec.SetGpuDevicesRequest(gpuDevices)
	}

	if gpuMemory != "" {
		spec.SetGpuMemoryRequest(gpuMemory)
	}

	if cpuMemory != "" {
		spec.SetCpuMemoryRequest(cpuMemory)
	}

	return openapi.ComputeAsset{
		Meta: openapi.AssetMeta{
			Id:    id,
			Name:  name,
			Scope: openapi.Scope(scope),
		},
		Spec: spec,
	}
}
