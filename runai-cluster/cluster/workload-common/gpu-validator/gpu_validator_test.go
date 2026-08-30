package gpu_validator

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	fields "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

const (
	NoError = ""

	FractionalNotSupported  = "fractional GPU is not supported in this cluster, please use the 'gpuDevicesRequest' field"
	MigNotSupported         = "Multi-Instance GPUs(MIG) are not supported in this cluster"
	GpuLimitNotSupported    = "using explicit 'gpuLimit' is only available with fractional GPU, which is not supported in this cluster"
	MemoryNotSupported      = "using explicit 'gpuMemory' is only available with fractional GPU, which is not supported in this cluster"
	MemoryLimitNotSupported = "using explicit 'gpuMemoryLimit' is only available with fractional GPU, which is not supported in this cluster"

	MixedAttributesError          = "only one of the fields: gpu (fractional), gpuMemory or migProfile should be provided. it is not permitted to specify both"
	MissingAttributesError        = "either gpu, gpuMemory or migProfile should be provided"
	FractionalConflictsGpuDevices = "GPU is only supported with gpuDevices > 0"

	MigOrMemoryConflictsGpuDevices  = "gpuMemory or migProfile can be provided only when requesting single GPU"
	NonFractionConflictsGpuDevices  = "when the value of 'gpuDevices' 1 and above, the value of 'gpu' must reside between 0 and 1"
	NonFractionConflictsZeroDevices = "when the value of 'gpuDevices' is 0, 'gpu' should not be specified"
	NonFractionConflictsGpuLimit    = "gpuLimit can only be provided with fractional GPU (0 < gpu < 1)"
	GpuLimitGreaterThanOne          = "gpuLimit must be no greater than 1"
	LimitVsRequestError             = "the value of gpuLimit must be no less than gpu"
	MemLimitVsRequestError          = "the value of gpuMemoryLimit must be no less than gpuMemory"
)

type gpuTestEntry struct {
	description    string
	resourceFields field_groups.ResourceAllocationFields
	migProfile     *fields.StringField
	NvidiaError    string
}

type gpuSpecFields struct {
	gpuDevices     *int
	gpu            *float64
	gpuLimit       *float64
	gpuMemory      string
	gpuMemoryLimit string
	migProfile     *string
}

const (
	moreThanOneInteger    = 4
	smallFractionalNumber = 0.1
	largeFractionalNumber = 0.9
	wholeWithFractional   = 1.5
	someMemorySize        = "1G"
	smallMemorySize       = "100M"
	largeMemorySize       = "200M"
)

var someMigProfile = ptr.To("profile1")

func createGinkgoTestEntryFromGpuTestEntries(tests []gpuTestEntry) []TableEntry {
	var testEntries []TableEntry
	for _, entry := range tests {
		testEntries = append(testEntries, Entry(entry.description, &entry.resourceFields, entry.migProfile, entry.NvidiaError))
	}
	return testEntries
}

func createGPUSpecTestEntries() []gpuTestEntry {
	tests := []gpuTestEntry{
		//********************
		// nothing specified *
		//********************
		{
			description:    "nothing is given",
			resourceFields: createResourceFields(gpuSpecFields{}),
			migProfile:     &fields.StringField{},
			NvidiaError:    NoError,
		},

		//***********************
		// single field - valid *
		//***********************
		{
			description: "only non-fractional gpu is given",
			resourceFields: createResourceFields(gpuSpecFields{
				gpu: ptr.To[float64](moreThanOneInteger),
			}),
			NvidiaError: NoError,
		},
		{
			description: "only gpu devices > 1 is given",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuDevices: ptr.To(moreThanOneInteger),
			}),
			NvidiaError: NoError,
		},
		{
			description: "only fractional gpu is given",
			resourceFields: createResourceFields(gpuSpecFields{
				gpu: ptr.To[float64](smallFractionalNumber),
			}),
			NvidiaError: NoError,
		},
		{
			description: "only gpuMemory is given",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuMemory: someMemorySize,
			}),
			NvidiaError: NoError,
		},

		//***************
		// only limits  *
		//***************
		{
			description: "only gpu limit / fractional",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuLimit: ptr.To[float64](largeFractionalNumber),
			}),
			NvidiaError: NoError,
		},
		{
			description: "only gpu limit / non-fractional",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuLimit: ptr.To[float64](moreThanOneInteger),
			}),
			NvidiaError: GpuLimitGreaterThanOne,
		},
		{
			description: "only memory limit",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuMemoryLimit: someMemorySize,
			}),
			NvidiaError: NoError,
		},

		//***************
		// limits + gpu *
		//***************
		{
			description: "gpu limit / fractional - request = limit",
			resourceFields: createResourceFields(gpuSpecFields{
				gpu:      ptr.To[float64](largeFractionalNumber),
				gpuLimit: ptr.To[float64](largeFractionalNumber),
			}),
			NvidiaError: NoError,
		},
		{
			description: "gpu limit / fractional - request < limit",
			resourceFields: createResourceFields(gpuSpecFields{
				gpu:      ptr.To[float64](smallFractionalNumber),
				gpuLimit: ptr.To[float64](largeFractionalNumber),
			}),
			NvidiaError: NoError,
		},
		{
			description: "gpu limit / fractional - request > limit",
			resourceFields: createResourceFields(gpuSpecFields{
				gpu:      ptr.To[float64](largeFractionalNumber),
				gpuLimit: ptr.To[float64](smallFractionalNumber),
			}),
			NvidiaError: LimitVsRequestError,
		},

		//******************
		// limits + memory *
		//******************
		{
			description: "memory + memory limit",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuMemory:      smallMemorySize,
				gpuMemoryLimit: largeMemorySize,
			}),
			NvidiaError: NoError,
		},
		{
			description: "memory = memory limit",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuMemory:      someMemorySize,
				gpuMemoryLimit: someMemorySize,
			}),
			NvidiaError: NoError,
		},
		{
			description: "memory > memoryLimit",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuMemory:      largeMemorySize,
				gpuMemoryLimit: smallMemorySize,
				gpuLimit:       ptr.To[float64](smallFractionalNumber), // no problem with this, it is ignored
			}),
			NvidiaError: MemLimitVsRequestError,
		},

		//******************
		// gpu devices = 0 *
		//******************
		{
			description: "zero gpu devices + fractional gpu",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuDevices: ptr.To(0),
				gpu:        ptr.To[float64](smallFractionalNumber),
			}),
			NvidiaError: NonFractionConflictsZeroDevices,
		},
		{
			description: "zero gpu devices + non-fractional gpu",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuDevices: ptr.To(0),
				gpu:        ptr.To[float64](moreThanOneInteger),
			}),
			NvidiaError: NonFractionConflictsZeroDevices,
		},

		//******************
		// gpu devices = 1 *
		//******************
		{
			description: "1 gpuDevice + non-fractional gpu",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuDevices: ptr.To(1),
				gpu:        ptr.To[float64](moreThanOneInteger),
			}),
			NvidiaError: NonFractionConflictsGpuDevices,
		},
		{
			description: "1 gpuDevice + fractional",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuDevices: ptr.To(1),
				gpu:        ptr.To[float64](smallFractionalNumber),
			}),
			NvidiaError: NoError,
		},
		{
			description: "1 gpuDevice + memory",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuDevices: ptr.To(1),
				gpuMemory:  someMemorySize,
			}),
			NvidiaError: NoError,
		},

		//******************
		// backward no gpu device with gpu whole number fractional gpu
		// ******************
		{
			description: "no gpuDevice + whole gpu + fractional gpu",
			resourceFields: createResourceFields(gpuSpecFields{
				gpu: ptr.To[float64](wholeWithFractional),
			}),
			NvidiaError: WholeGPUWithFractionalGPUError,
		},
		{
			description: "no gpuDevice + whole gpu",
			resourceFields: createResourceFields(gpuSpecFields{
				gpu: ptr.To[float64](1.0),
			}),
			NvidiaError: NoError,
		},

		//******************
		// gpu devices > 1 *
		//******************
		{
			description: "gpu devices > 1 + fractional gpu",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuDevices: ptr.To(moreThanOneInteger),
				gpu:        ptr.To[float64](smallFractionalNumber),
			}),
			NvidiaError: NoError,
		},
		{
			description: "gpu devices > 1 + non-fractional gpu",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuDevices: ptr.To(moreThanOneInteger),
				gpu:        ptr.To[float64](moreThanOneInteger),
			}),
			NvidiaError: NonFractionConflictsGpuDevices,
		},
		{
			description: "gpuDevice > 1 + memory",
			resourceFields: createResourceFields(gpuSpecFields{
				gpuDevices: ptr.To(moreThanOneInteger),
				gpuMemory:  someMemorySize,
			}),
			NvidiaError: NoError,
		},
	}
	return tests
}

var _ = Describe("GPU Validator", func() {

	var (
		validator      GpuValidator
		ctx            context.Context
		gpuTestEntries []gpuTestEntry
	)
	gpuTestEntries = createGPUSpecTestEntries()

	Context("nvidia cluster", func() {
		var nvidiaTestEntries []TableEntry

		nvidiaTestEntries = createGinkgoTestEntryFromGpuTestEntries(gpuTestEntries)

		BeforeEach(func() {
			validator = NewGpuValidator()
		})

		DescribeTable("Validate", func(resourceFields *field_groups.ResourceAllocationFields, expectedError string) {
			checkValidator(ctx, validator, resourceFields, expectedError)
		}, nvidiaTestEntries)
	})
})

func createResourceFields(specFields gpuSpecFields) field_groups.ResourceAllocationFields {
	var migProfile *fields.StringField
	if specFields.migProfile != nil {
		migProfile = &fields.StringField{Value: specFields.migProfile}
	}
	_ = migProfile // This is handled separately in the test

	return field_groups.ResourceAllocationFields{
		GpuDevices:     fields.NewOptionalIntField(specFields.gpuDevices),
		Gpu:            fields.NewOptionalNumberField(specFields.gpu),
		GpuLimit:       fields.NewOptionalNumberField(specFields.gpuLimit),
		GpuMemory:      fields.NewOptionalQuantityField(specFields.gpuMemory, ""),
		GpuMemoryLimit: fields.NewOptionalQuantityField(specFields.gpuMemoryLimit, ""),
	}
}

func checkValidator(ctx context.Context, validator GpuValidator, resourceFields *field_groups.ResourceAllocationFields, expectedError string) {
	err := validator.Validate(ctx, resourceFields)
	if expectedError == NoError {
		Expect(err).To(Succeed())
	} else {
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(expectedError))
	}
}
