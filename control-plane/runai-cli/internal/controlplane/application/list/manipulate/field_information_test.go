package manipulate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

var _ = Describe("Extract Functions", func() {
	var (
		app *openapi.AIApplication
		row *AIApplicationAlignedRow
	)

	BeforeEach(func() {
		app = &openapi.AIApplication{}
		row = &AIApplicationAlignedRow{}
	})

	Context("NameField extract", func() {
		It("should extract application name", func() {
			app.Name = ptr.To("my-app")
			NameField.extractFunc(app, row)
			Expect(row.Name).To(Equal("my-app"))
		})

		It("should set empty name when not set", func() {
			NameField.extractFunc(app, row)
			Expect(row.Name).To(Equal(""))
		})
	})

	Context("TypeField extract", func() {
		It("should extract application type", func() {
			app.Type = ptr.To("inference")
			TypeField.extractFunc(app, row)
			Expect(row.Type).To(Equal("inference"))
		})

		It("should set empty type when not set", func() {
			TypeField.extractFunc(app, row)
			Expect(row.Type).To(Equal(""))
		})
	})

	Context("StatusField extract", func() {
		It("should extract action status", func() {
			statusStr := "Running"
			app.ActionStatus.Set(&statusStr)
			StatusField.extractFunc(app, row)
			Expect(row.Status).To(Equal("Running"))
		})

		It("should set empty status when not set", func() {
			StatusField.extractFunc(app, row)
			Expect(row.Status).To(Equal(""))
		})
	})

	Context("ProjectField extract", func() {
		It("should extract project name", func() {
			app.ProjectName = ptr.To("team-alpha")
			ProjectField.extractFunc(app, row)
			Expect(row.Project).To(Equal("team-alpha"))
		})

		It("should set empty project when not set", func() {
			ProjectField.extractFunc(app, row)
			Expect(row.Project).To(Equal(""))
		})
	})

	Context("GPUComputeReqField extract", func() {
		It("should extract GPU compute request as formatted float", func() {
			gpuCores := openapi.RequestResourceCores{}
			gpuCores.SetRequest(3.5)
			req := openapi.AIApplicationRequestResources{}
			req.Gpu.Set(&gpuCores)
			app.RequestedResources.Set(&req)
			GPUComputeReqField.extractFunc(app, row)
			Expect(row.GPUComputeReq).To(Equal("3.50"))
		})

		It("should leave GPUComputeReq empty when resources not set", func() {
			GPUComputeReqField.extractFunc(app, row)
			Expect(row.GPUComputeReq).To(Equal(""))
		})

		It("should leave GPUComputeReq empty when GPU request not set", func() {
			gpuCores := openapi.RequestResourceCores{}
			gpuCores.SetLimit(3.5)
			req := openapi.AIApplicationRequestResources{}
			req.Gpu.Set(&gpuCores)
			app.RequestedResources.Set(&req)
			GPUComputeReqField.extractFunc(app, row)
			Expect(row.GPUComputeReq).To(Equal(""))
		})
	})

	Context("GPUComputeAllocField extract", func() {
		It("should extract GPU compute alloc as formatted float", func() {
			gpuVal := 2.5
			alloc := openapi.AIApplicationAllocatedResources{}
			alloc.Gpu.Set(&gpuVal)
			app.AllocatedResources.Set(&alloc)
			GPUComputeAllocField.extractFunc(app, row)
			Expect(row.GPUComputeAlloc).To(Equal("2.50"))
		})

		It("should leave GPUComputeAlloc empty when resources not set", func() {
			GPUComputeAllocField.extractFunc(app, row)
			Expect(row.GPUComputeAlloc).To(Equal(""))
		})

		It("should leave GPUComputeAlloc empty when GPU not set in alloc", func() {
			alloc := openapi.AIApplicationAllocatedResources{}
			app.AllocatedResources.Set(&alloc)
			GPUComputeAllocField.extractFunc(app, row)
			Expect(row.GPUComputeAlloc).To(Equal(""))
		})
	})

	Context("GPUMemoryReqField extract", func() {
		It("should extract GPU memory request", func() {
			requestStr := "4Gi"
			gpuMem := openapi.RequestResourceQuantity{}
			gpuMem.Request.Set(&requestStr)
			req := openapi.AIApplicationRequestResources{}
			req.GpuMemory.Set(&gpuMem)
			app.RequestedResources.Set(&req)
			GPUMemoryReqField.extractFunc(app, row)
			Expect(row.GPUMemoryReq).To(Equal("4Gi"))
		})

		It("should leave GPUMemoryReq empty when resources not set", func() {
			GPUMemoryReqField.extractFunc(app, row)
			Expect(row.GPUMemoryReq).To(Equal(""))
		})
	})

	Context("GPUMemoryAllocField extract", func() {
		It("should extract GPU memory alloc", func() {
			gpuMemStr := "8Gi"
			alloc := openapi.AIApplicationAllocatedResources{}
			alloc.GpuMemory.Set(&gpuMemStr)
			app.AllocatedResources.Set(&alloc)
			GPUMemoryAllocField.extractFunc(app, row)
			Expect(row.GPUMemoryAlloc).To(Equal("8Gi"))
		})

		It("should leave GPUMemoryAlloc empty when resources not set", func() {
			GPUMemoryAllocField.extractFunc(app, row)
			Expect(row.GPUMemoryAlloc).To(Equal(""))
		})
	})

	Context("CPUComputeAllocField extract", func() {
		It("should extract CPU compute alloc as formatted float", func() {
			cpuVal := 4.0
			alloc := openapi.AIApplicationAllocatedResources{}
			alloc.Cpu.Set(&cpuVal)
			app.AllocatedResources.Set(&alloc)
			CPUComputeAllocField.extractFunc(app, row)
			Expect(row.CPUComputeAlloc).To(Equal("4.00"))
		})

		It("should leave CPUComputeAlloc empty when resources not set", func() {
			CPUComputeAllocField.extractFunc(app, row)
			Expect(row.CPUComputeAlloc).To(Equal(""))
		})
	})

	Context("CPUComputeReqField extract", func() {
		It("should extract CPU compute request as formatted float", func() {
			cpuCores := openapi.RequestResourceCores{}
			cpuCores.SetRequest(2.0)
			req := openapi.AIApplicationRequestResources{}
			req.Cpu.Set(&cpuCores)
			app.RequestedResources.Set(&req)
			CPUComputeReqField.extractFunc(app, row)
			Expect(row.CPUComputeReq).To(Equal("2.00"))
		})

		It("should leave CPUComputeReq empty when resources not set", func() {
			CPUComputeReqField.extractFunc(app, row)
			Expect(row.CPUComputeReq).To(Equal(""))
		})
	})

	Context("CPUMemoryReqField extract", func() {
		It("should extract CPU memory request", func() {
			requestStr := "2Gi"
			cpuMem := openapi.RequestResourceQuantity{}
			cpuMem.Request.Set(&requestStr)
			req := openapi.AIApplicationRequestResources{}
			req.CpuMemory.Set(&cpuMem)
			app.RequestedResources.Set(&req)
			CPUMemoryReqField.extractFunc(app, row)
			Expect(row.CPUMemoryReq).To(Equal("2Gi"))
		})

		It("should leave CPUMemoryReq empty when resources not set", func() {
			CPUMemoryReqField.extractFunc(app, row)
			Expect(row.CPUMemoryReq).To(Equal(""))
		})
	})

	Context("CPUMemoryAllocField extract", func() {
		It("should extract CPU memory alloc", func() {
			cpuMemStr := "16Gi"
			alloc := openapi.AIApplicationAllocatedResources{}
			alloc.CpuMemory.Set(&cpuMemStr)
			app.AllocatedResources.Set(&alloc)
			CPUMemoryAllocField.extractFunc(app, row)
			Expect(row.CPUMemoryAlloc).To(Equal("16Gi"))
		})

		It("should leave CPUMemoryAlloc empty when resources not set", func() {
			CPUMemoryAllocField.extractFunc(app, row)
			Expect(row.CPUMemoryAlloc).To(Equal(""))
		})
	})
})

var _ = Describe("Format Functions", func() {
	var row *AIApplicationAlignedRow

	BeforeEach(func() {
		row = &AIApplicationAlignedRow{}
	})

	Context("NameField format", func() {
		It("should return name when set", func() {
			row.Name = "my-app"
			Expect(NameField.formatFunc(row)).To(Equal("my-app"))
		})

		It("should return empty string when name is empty", func() {
			Expect(NameField.formatFunc(row)).To(Equal(""))
		})
	})

	Context("TypeField format", func() {
		It("should return type when set", func() {
			row.Type = "inference"
			Expect(TypeField.formatFunc(row)).To(Equal("inference"))
		})

		It("should return '-' when type is empty", func() {
			Expect(TypeField.formatFunc(row)).To(Equal("-"))
		})
	})

	Context("StatusField format", func() {
		It("should return status when set", func() {
			row.Status = "Running"
			Expect(StatusField.formatFunc(row)).To(Equal("Running"))
		})

		It("should return '-' when status is empty", func() {
			Expect(StatusField.formatFunc(row)).To(Equal("-"))
		})
	})

	Context("ProjectField format", func() {
		It("should return project name when set", func() {
			row.Project = "team-alpha"
			Expect(ProjectField.formatFunc(row)).To(Equal("team-alpha"))
		})

		It("should return '-' when project is empty", func() {
			Expect(ProjectField.formatFunc(row)).To(Equal("-"))
		})
	})

	Context("GPUComputeReqField format", func() {
		It("should return GPU compute req when set", func() {
			row.GPUComputeReq = "2.50"
			Expect(GPUComputeReqField.formatFunc(row)).To(Equal("2.50"))
		})

		It("should return '-' when GPU compute req is empty", func() {
			Expect(GPUComputeReqField.formatFunc(row)).To(Equal("-"))
		})
	})

	Context("GPUComputeAllocField format", func() {
		It("should return GPU compute alloc when set", func() {
			row.GPUComputeAlloc = "3.00"
			Expect(GPUComputeAllocField.formatFunc(row)).To(Equal("3.00"))
		})

		It("should return '-' when GPU compute alloc is empty", func() {
			Expect(GPUComputeAllocField.formatFunc(row)).To(Equal("-"))
		})
	})

	Context("GPUMemoryReqField format", func() {
		It("should return GPU memory req when set", func() {
			row.GPUMemoryReq = "4Gi"
			Expect(GPUMemoryReqField.formatFunc(row)).To(Equal("4Gi"))
		})

		It("should return '-' when GPU memory req is empty", func() {
			Expect(GPUMemoryReqField.formatFunc(row)).To(Equal("-"))
		})
	})

	Context("GPUMemoryAllocField format", func() {
		It("should return GPU memory alloc when set", func() {
			row.GPUMemoryAlloc = "8Gi"
			Expect(GPUMemoryAllocField.formatFunc(row)).To(Equal("8Gi"))
		})

		It("should return '-' when GPU memory alloc is empty", func() {
			Expect(GPUMemoryAllocField.formatFunc(row)).To(Equal("-"))
		})
	})

	Context("CPUComputeReqField format", func() {
		It("should return CPU compute req when set", func() {
			row.CPUComputeReq = "1.00"
			Expect(CPUComputeReqField.formatFunc(row)).To(Equal("1.00"))
		})

		It("should return '-' when CPU compute req is empty", func() {
			Expect(CPUComputeReqField.formatFunc(row)).To(Equal("-"))
		})
	})

	Context("CPUComputeAllocField format", func() {
		It("should return CPU compute alloc when set", func() {
			row.CPUComputeAlloc = "2.00"
			Expect(CPUComputeAllocField.formatFunc(row)).To(Equal("2.00"))
		})

		It("should return '-' when CPU compute alloc is empty", func() {
			Expect(CPUComputeAllocField.formatFunc(row)).To(Equal("-"))
		})
	})

	Context("CPUMemoryReqField format", func() {
		It("should return CPU memory req when set", func() {
			row.CPUMemoryReq = "2Gi"
			Expect(CPUMemoryReqField.formatFunc(row)).To(Equal("2Gi"))
		})

		It("should return '-' when CPU memory req is empty", func() {
			Expect(CPUMemoryReqField.formatFunc(row)).To(Equal("-"))
		})
	})

	Context("CPUMemoryAllocField format", func() {
		It("should return CPU memory alloc when set", func() {
			row.CPUMemoryAlloc = "16Gi"
			Expect(CPUMemoryAllocField.formatFunc(row)).To(Equal("16Gi"))
		})

		It("should return '-' when CPU memory alloc is empty", func() {
			Expect(CPUMemoryAllocField.formatFunc(row)).To(Equal("-"))
		})
	})
})
