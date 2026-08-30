package gpu_validator

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

var _ = Describe("Resource Utils", func() {
	Context("HasRequestedGpus", func() {
		DescribeTable("via explicit resources",
			func(gpuDevices *int, gpu, gpuLimit *float64, gpuMemory, gpuMemoryLimit string, expected bool) {
				resourceFields := getResourceAllocationFields(gpuDevices, gpu, gpuLimit, gpuMemory, gpuMemoryLimit)
				actual := HasRequestedGpus(&resourceFields)
				Expect(actual).To(Equal(expected))
			},
			Entry("devices", ptr.To(1), nil, nil, "", "", true),
			Entry("gpu", nil, ptr.To(0.5), nil, "", "", true),
			Entry("gpu limit", nil, nil, ptr.To(0.5), "", "", true),
			Entry("gpu memory", nil, nil, nil, "0.5", "", true),
			Entry("gpu memory limit", nil, nil, nil, "", "0.5", true),
			Entry("none", nil, nil, nil, "", "", false),
		)

		DescribeTable("via ExtendedResources",
			func(resourceNamePrefix, quantity string, expected bool) {
				resourceFields := field_groups.ResourceAllocationFields{
					ExtendedResources: getExtendedResourcesWithGpu(resourceNamePrefix, quantity),
				}
				actual := HasRequestedGpus(&resourceFields)
				Expect(actual).To(Equal(expected))
			},
			Entry("uses gpu", NvidiaGpuExtendedResourceNamePrefixes[0], "1", true),
			Entry("zero quantity", NvidiaGpuExtendedResourceNamePrefixes[0], "0", false),
			Entry("not using", "madeupsomething", "1", false),
		)

		It("should return false for nil ResourceAllocationFields", func() {
			actual := HasRequestedGpus(nil)
			Expect(actual).To(BeFalse())
		})
	})
})

func getResourceAllocationFields(gpuDevices *int, gpu *float64, gpuLimit *float64, gpuMemory string, gpuMemoryLimit string) field_groups.ResourceAllocationFields {
	return field_groups.ResourceAllocationFields{
		GpuDevices:     field_types.NewOptionalIntField(gpuDevices),
		Gpu:            field_types.NewOptionalNumberField(gpu),
		GpuLimit:       field_types.NewOptionalNumberField(gpuLimit),
		GpuMemory:      field_types.NewOptionalQuantityField(gpuMemory, ""),
		GpuMemoryLimit: field_types.NewOptionalQuantityField(gpuMemoryLimit, ""),
	}
}

func getExtendedResourcesWithGpu(prefix, quantity string) *field_types.ExtendedResourcesField {
	extendedResourcesMap := make(map[string]field_types.ExtendedResourcesItemAndRules)
	extendedResourcesMap["ext-resc"] = field_types.ExtendedResourcesItemAndRules{
		Rules: nil,
		Value: &field_types.ExtendedResourcesItemValue{
			Name:     ptr.To(prefix),
			Quantity: ptr.To(quantity),
		},
		Deleted: nil,
	}

	return field_types.NewExtendedResourcesField(extendedResourcesMap)
}
