package field_groups

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

// +k8s:deepcopy-gen=true
type ResourceAllocationFields struct {
	// Specifies the number of GPUs to allocate for the created workload.
	GpuDevices *flds.IntField `json:"gpuDevices,omitempty"`

	// Specifies the fraction of GPU to be allocated to the workload, between 0 and 1.
	// for backward compatibility, it also supports number of gpuDevices larger than 1, currently
	// provided using the gpuDevices field.
	Gpu *flds.NumberField `json:"gpu,omitempty"`

	// Specifies a limit on the GPUs to allocate for this workload.
	// can be specified only when Gpu is no greater than 1, and must be no less than the gpu
	GpuLimit *flds.NumberField `json:"gpuLimit,omitempty"`

	// Specifies GPU memory to allocate for the created workload. The workload
	// will receive this amount of memory. Note that the workload will not be scheduled unless the system
	// can guarantee this amount of GPU memory to the workload.
	GpuMemory *flds.QuantityField `json:"gpuMemory,omitempty"`

	// Specifies a limit on the GPU memory to allocate for this workload. should be no less than
	// the gpuMemory
	GpuMemoryLimit *flds.QuantityField `json:"gpuMemoryLimit,omitempty"`

	// Specifies CPU units to allocate for the created workload (0.5, 1, .etc). The workload will
	// receive at least this amount of CPU. Note that the workload will not be scheduled unless the system can
	// guarantee this amount of CPUs to the workload.
	Cpu *flds.NumberField `json:"cpu,omitempty"`

	// Specifies a limit on the number of CPUs consumed by the workload (0.5, 1, .etc).
	// The system guarantees that this workload will not be able to consume more than this amount of CPUs.
	CpuLimit *flds.NumberField `json:"cpuLimit,omitempty"`

	// Specifies the amount of CPU memory to allocate for this workload (1G, 20M, .etc). The workload will
	// receive at least this amount of memory. Note that the workload will not be scheduled unless the system
	// can guarantee this amount of memory to the workload
	Memory *flds.QuantityField `json:"memory,omitempty"`

	// Specifies a limit on the CPU memory to allocate for this workload (1G, 20M, .etc).
	// The system guarantees that this workload will not be able to consume more than this amount of memory.
	// The workload will receive an error when trying to allocate more memory than this limit.
	MemoryLimit *flds.QuantityField `json:"memoryLimit,omitempty"`

	// Specifies values for extended resources.
	// Extended resources are third-party devices (such as high-performance NICs, FPGAs, or InfiniBand adapters) that you want to allocate to your Job.
	// For more information see: https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/
	ExtendedResources *flds.ExtendedResourcesField `json:"extendedResources,omitempty"`
}

var ResourceAllocationRunaiRules = ResourceAllocationFields{

	Gpu: &flds.NumberField{
		Rules: &flds.NumberRules{
			Min: ptr.To("0"),
		},
	},

	GpuLimit: &flds.NumberField{
		Rules: &flds.NumberRules{
			Min: ptr.To("0"),
		},
	},

	Cpu: &flds.NumberField{
		Rules: &flds.NumberRules{
			Min: ptr.To("0"),
		},
	},

	CpuLimit: &flds.NumberField{
		Rules: &flds.NumberRules{
			Min: ptr.To("0"),
		},
	},

	ExtendedResources: &flds.ExtendedResourcesField{
		ItemRules: &flds.ExtendedResourcesItemRules{
			Name: &flds.StringRules{
				Required: ptr.To(true),
			},
			Quantity: &flds.StringRules{
				Required: ptr.To(true),
				Format:   ptr.To(workload_fields.FormatMemory),
			},
		},
	},
}
