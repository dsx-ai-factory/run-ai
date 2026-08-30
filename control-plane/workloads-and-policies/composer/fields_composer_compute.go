package composer

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/jsonutils"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	MemoryFormatError = "compute.%s does not comply with k8s resource units rules. for further details: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers"
)

var computeJsonCracker = jsonutils.NewJsonCracker(workloadsApi.WorkspaceSpecSpecAllOfCompute{})
var GpuPortionField = computeJsonCracker.NameOf(&computeJsonCracker.Struct().GpuPortionRequest)
var GpuMemoryField = computeJsonCracker.NameOf(&computeJsonCracker.Struct().GpuMemoryRequest)
var GpuRequestTypeField = computeJsonCracker.NameOf(&computeJsonCracker.Struct().GpuRequestType)
var GpuDeviceRequestField = computeJsonCracker.NameOf(&computeJsonCracker.Struct().GpuDevicesRequest)

type FractionalValueInfo struct {
	ValueGiven bool
	FieldName  string
}

func (fc *FieldsComposer) ComposeCpuCoreRequest(src *float64,
	target any, setter FnFloatSetter[float64], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src != nil {
			if *src < 0 {
				return workloadsApi.NewError(http.StatusBadRequest, "compute.cpuCoreRequest must be no less than zero")
			}

			return setter(src, target, cookie)
		}

		return nil
	}
}

func (fc *FieldsComposer) ComposeCpuCoreLimit(src *float64, cpuCoreRequest *float64,
	target any, setter FnFloatSetter[float64], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		limit := convutils.Deref(src, 0)
		request := convutils.Deref(cpuCoreRequest, 0)

		if limit < 0 {
			return workloadsApi.NewError(http.StatusBadRequest, "compute.cpuCoreLimit must be no less than zero")
		}

		if limit < request {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("compute.cpuCoreLimit (%.1f) should be no less than cpuCoreRequest (%.1f)", limit, request))
		}

		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeCpuMemoryRequest(src *string,
	target any, setter FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		quantity, parseErr := resource.ParseQuantity(*src)
		if parseErr != nil {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf(MemoryFormatError, "cpuMemoryRequest"))
		}

		if quantity.Value() < 0 {
			return workloadsApi.NewError(http.StatusBadRequest, "compute.cpuMemoryRequest must be no less than zero")
		}

		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeCpuMemoryLimit(src *string, cpuMemoryRequest *string,
	target any, setter FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		cpuMemoryLimitQuantity, parseErr := resource.ParseQuantity(*src)
		if parseErr != nil {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf(MemoryFormatError, "cpuMemoryLimit"))
		}

		if cpuMemoryLimitQuantity.Value() < 0 {
			return workloadsApi.NewError(http.StatusBadRequest, "compute.cpuMemoryLimit must be no less than zero")
		}

		if cpuMemoryRequest != nil {
			cpuMemoryRequestQuantity, parseErr := resource.ParseQuantity(*cpuMemoryRequest)
			if parseErr == nil && cpuMemoryLimitQuantity.Value() < cpuMemoryRequestQuantity.Value() {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("compute.cpuMemoryLimit (%v) should be no less than cpuMemoryRequest (%v)", cpuMemoryLimitQuantity.Value(), cpuMemoryRequestQuantity.Value()))
			}
		}

		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeGpuRequestsAndLimits(_ context.Context, devicesReq *int32, requestType *workloadsApi.GpuRequestType,
	portionRequest *float64, portionLimit *float64, memoryRequest *string, memoryLimit *string, setter FnGpuSizingInfo,
	target any, cookie any, sizingInfo *GpuSizingInfo,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		var gpuSizing GpuSizingInfo

		// gpu quantity can be fractional (gpuPortionRequest specified) or multiple non-fractional gpus if not specified.
		// we load it from the workload or, if not specified, from the default
		fractionalRequestType := convutils.Deref(requestType, "")

		//  note if value for any of the fractional fields is given. if such default is given, it determines the
		//  gpu request type of the workload
		gpuPortionMode := portionRequest != nil
		gpuMemoryMode := memoryRequest != nil

		if gpuMemoryMode && gpuPortionMode {
			return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuPortionRequest and gpuMemoryRequest are mutually exclusive")
		}

		if !fc.Options.IgnoreMissingValues {
			if fractionalRequestType == workloadsApi.GPUREQUESTTYPE_MEMORY && memoryRequest == nil {
				return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuMemoryRequest is mandatory when gpuRequestType is memory")
			}
			if fractionalRequestType == workloadsApi.GPUREQUESTTYPE_PORTION && portionRequest == nil {
				return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuPortionRequest is mandatory when gpuRequestType is portion")
			}
		}

		gpuSizing.NumDevices = devicesReq

		if convutils.Deref(gpuSizing.NumDevices, 0) < 0 {
			return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuDevicesRequest must be no less than zero")
		}

		var parseErr error
		var gpuMemoryRequestQuantity resource.Quantity

		if gpuMemoryMode {
			gpuMemoryRequestQuantity, parseErr = resource.ParseQuantity(*memoryRequest)
			if parseErr != nil {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf(MemoryFormatError, "gpuMemoryRequest"))
			}
			if gpuMemoryRequestQuantity.Value() <= 0 {
				return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuMemoryRequest must be greater than zero")
			}

			gpuSizing.MemoryRequest = memoryRequest

			gpuMemoryRequestQuantity, parseErr = resource.ParseQuantity(*memoryRequest)
			if parseErr != nil {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf(MemoryFormatError, "gpuMemoryRequest"))
			}

			if gpuMemoryRequestQuantity.Value() <= 0 {
				return workloadsApi.NewError(http.StatusBadRequest, "compute: gpuMemoryRequest must be greater than zero")
			}
		}

		if memoryLimit != nil {
			gpuSizing.MemoryLimit = memoryLimit

			gpuMemoryLimitQuantity, err := resource.ParseQuantity(*memoryLimit)
			if err != nil {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf(MemoryFormatError, "gpuMemoryLimit"))
			}

			if gpuMemoryLimitQuantity.Value() <= 0 {
				return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuMemoryLimit must be greater than zero")
			}

			if memoryRequest != nil && gpuMemoryLimitQuantity.Value() < gpuMemoryRequestQuantity.Value() {
				return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuMemoryLimit should be no less than gpuMemoryRequest")
			}
		}

		if gpuPortionMode {
			gpuSizing.PortionRequest = portionRequest

			if gpuSizing.PortionRequest != nil && convutils.Deref(gpuSizing.PortionRequest, 0.0) <= 0 {
				return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuPortionRequest must be greater than zero")
			}

			// Reject whole numbers with fractional parts (e.g. 2.5) - must be either 0 < portion <= 1 or an integer.
			// NOTE: Ideally gpuPortionRequest should only be 0-1, but integer values > 1 (e.g. 2, 3) are still
			// accepted for backward compatibility with the legacy `gpu` field which was overloaded to hold either
			// a fraction (0-1) or a device count (2, 3, 4...). The historic gpu field is no longer needed as all
			// supported clusters now support the new set of fields, but we keep it here for the time being and
			// remove it as part of future code cleanup.
			if gpuSizing.PortionRequest != nil {
				portion := *gpuSizing.PortionRequest
				if portion > 1 && portion != float64(int64(portion)) {
					return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuPortionRequest must be either a fraction (0 to 1) or a whole number")
				}
			}
		}

		if portionLimit != nil {
			gpuSizing.PortionLimit = portionLimit

			if *gpuSizing.PortionLimit <= 0 {
				return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuPortionLimit must be greater than zero")
			}

			// Reject whole numbers with fractional parts (e.g. 2.5). See comment above for gpuPortionRequest
			// regarding the historic gpu field and legacy support for integer values > 1.
			if *gpuSizing.PortionLimit > 1 && *gpuSizing.PortionLimit != float64(int64(*gpuSizing.PortionLimit)) {
				return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuPortionLimit must be either a fraction (0 to 1) or a whole number")
			}

			if portionRequest != nil {
				if *portionLimit < *portionRequest {
					return workloadsApi.NewError(http.StatusBadRequest, "compute.gpuPortionLimit should be no less than gpuPortionRequest")
				}
			}
		}

		*sizingInfo = gpuSizing

		return setter(&gpuSizing, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeExtendedResources(itemPrefix string, src []workloadsApi.ExtendedResource,
	target any, setter FnMapToMapSetter[workloadsApi.ExtendedResource], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		extendedResourcesItems := make(map[string]*workloadsApi.ExtendedResource)

		for idx, extendedResource := range src {
			extendedKey := fc.ItemizedNumbering.Compose(itemPrefix, extendedResource.Resource.Get(), nil, false)

			if !convutils.BoolIsTrue(extendedResource.Exclude.Get()) {
				if convutils.HasValue(extendedResource.Quantity) {
					_, err := resource.ParseQuantity(*extendedResource.Quantity.Get())
					if err != nil {
						return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("malformed quantity: %v", err.Error()))
					}
				} else if !fc.Options.IgnoreMissingValues {
					return workloadsApi.NewError(http.StatusBadRequest, "compute.extendedResources: quantity is not provided")
				}
			}

			extendedResourcesItems[extendedKey] = &src[idx]
		}

		return setter(extendedResourcesItems, target, cookie)
	}
}

const bytesToMB = 1000 * 1000

func (fc *FieldsComposer) ComposeMemoryLimitEnv(src *GpuSizingInfo, target any, setter FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src.MemoryLimit != nil && src.MemoryRequest != nil {
			return setter(src.MemoryLimit, target, cookie)
		}
		return nil
	}
}

func (fc *FieldsComposer) ComposePortionLimitEnv(src *GpuSizingInfo, target any, setter FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src.PortionLimit != nil && src.PortionRequest != nil {
			value := strconv.FormatFloat(*src.PortionLimit, 'f', -1, 64)
			return setter(&value, target, cookie)
		}
		return nil
	}
}

func (fc *FieldsComposer) ComposeMemoryReqAnnotation(src *GpuSizingInfo, target any, setter FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src.MemoryRequest != nil {
			gpuMemoryQuantity, err := resource.ParseQuantity(*src.MemoryRequest)
			if err != nil {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf(MemoryFormatError, "gpuMemoryLimit"))
			}

			quantityStr := strconv.Itoa(int(gpuMemoryQuantity.Value() / bytesToMB))

			return setter(&quantityStr, target, cookie)
		}
		return nil
	}
}

func (fc *FieldsComposer) ComposePortionReqAnnotation(src *GpuSizingInfo, target any, setter FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		// the gpu-fraction annotation is only valid for a fractional (sub-1) allocation.
		// a portion of 1 is a whole GPU and must not emit a fractional annotation.
		if src.IsFractionalPortion() {
			quantityStr := strconv.FormatFloat(*src.PortionRequest, 'f', -1, 64)
			return setter(&quantityStr, target, cookie)
		}
		return nil
	}
}

func (fc *FieldsComposer) ComposeNumDevicesAnnotation(src *GpuSizingInfo, target any, setter FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		// the gpu-fraction-num-devices annotation accompanies the gpu-fraction annotation and
		// is therefore only emitted for fractional allocations.
		if src.IsFractionalPortion() {
			gpuDevices := 1
			if src.NumDevices != nil && *src.NumDevices > 0 {
				gpuDevices = int(*src.NumDevices)
			}

			devicesStr := fmt.Sprintf("%d", gpuDevices)

			return setter(&devicesStr, target, cookie)
		}
		return nil
	}
}
