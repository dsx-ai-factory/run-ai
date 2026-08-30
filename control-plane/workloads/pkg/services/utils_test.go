package services

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

var _ = Describe("ResourceRef Conversion Utilities", func() {
	Describe("workloadResourceRefApiToModel", func() {
		It("should map uid, name, and kind to the model", func() {
			apiVal := api.ResourceRef{
				Uid:  "uid-1",
				Name: "data",
				Kind: api.ResourceRefKind(runv1.StorageRefKindPVC),
			}

			model := workloadResourceRefApiToModel(apiVal)

			Expect(model.UID).To(Equal("uid-1"))
			Expect(model.Name).To(Equal("data"))
			Expect(model.Kind).To(Equal(string(runv1.StorageRefKindPVC)))
		})
	})

	Describe("ConvertSliceTypes over resource refs", func() {
		It("should convert every api ref into a model ref", func() {
			apiRefs := []api.ResourceRef{
				{Uid: "uid-1", Name: "data", Kind: api.ResourceRefKind(runv1.StorageRefKindPVC)},
				{Uid: "uid-2", Name: "regcred", Kind: api.ResourceRefKind(runv1.StorageRefKindSecret)},
			}

			result := convutils.ConvertSliceTypes(apiRefs, workloadResourceRefApiToModel)

			Expect(result).To(Equal([]models.WorkloadResourceRef{
				{UID: "uid-1", Name: "data", Kind: string(runv1.StorageRefKindPVC)},
				{UID: "uid-2", Name: "regcred", Kind: string(runv1.StorageRefKindSecret)},
			}))
		})
	})
})
