package notifications

import (
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func servingPortToConnection(src *composer.ServingPortInfo, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	environments, ok := target.(*[]workloadsApi.Environment)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("servingPortToConnection: target type %T is not *[]workloadsApi.Environment", target))
	}

	environment := EnvironmentByReplica(environments, src.ReplicaType)
	if environment == nil {
		return nil
	}

	environment.Connections = append(environment.Connections, workloadsApi.Connection{
		Name:              string(openapi.INTERNALCONNECTIONTYPE_SERVING_PORT),
		ToolType:          string(openapi.INTERNALTOOLTYPE_CUSTOM),
		ConnectionType:    string(openapi.INTERNALCONNECTIONTYPE_SERVING_PORT),
		Url:               nil,
		ContainerPort:     src.PortInfo.Container.Get(),
		AuthorizationType: src.AuthInfo.AuthorizationType,
		AuthorizedUsers:   src.AuthInfo.AuthorizedUsers,
		AuthorizedGroups:  src.AuthInfo.AuthorizedGroups,
		CorrelationName:   &src.ItemId,
	})

	return nil
}

func nonZeroFloatToResourcesCore(src *float64, target any, _ any) *workloadsApi.Error {
	if src != nil {
		targetPtr, ok := target.(*workloadsApi.NullableRequestResourceCores)
		if !ok {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("nonZeroFloatToResourcesCore: target type %T is not *workloadsApi.NullableRequestResourceCores", target))
		}
		if targetPtr.Get() == nil {
			requestCore := workloadsApi.RequestResourceCores{
				Request: *workloadsApi.NewNullableFloat64(src),
			}
			targetPtr.Set(&requestCore)
		} else {
			targetPtr.Get().SetRequest(*src)
		}
	}

	return nil
}

func nonEmptyStringToResourcesQuantity(src *string, target any, _ any) *workloadsApi.Error {
	if src != nil {
		targetPtr, ok := target.(*workloadsApi.NullableRequestResourceQuantity)
		if !ok {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("nonEmptyStringToResourcesQuantity: target type %T is not *workloadsApi.NullableRequestResourceQuantity", target))
		}
		if targetPtr.Get() == nil {
			requestQuantity := workloadsApi.RequestResourceQuantity{
				Request: *workloadsApi.NewNullableString(src),
			}
			targetPtr.Set(&requestQuantity)
		} else {
			targetPtr.Get().SetRequest(*src)
		}
	}

	return nil
}

func gpuRequestedResourcesSetter(src *composer.GpuSizingInfo, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*workloadsApi.NullableWorkloadRequestResources)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("gpuRequestedResourcesSetter: target type %T is not *workloadsApi.NullableWorkloadRequestResources", target))
	}

	if targetPtr.Get() == nil {
		targetPtr.Set(&workloadsApi.WorkloadRequestResources{})
	}

	requestedResources := targetPtr.Get()

	if src.PortionRequest != nil {
		requestedResources.GpuRequestType.Set(convutils.Ref(workloadsApi.GPUREQUESTTYPE_PORTION))
	} else if src.MemoryRequest != nil {
		requestedResources.GpuRequestType.Set(convutils.Ref(workloadsApi.GPUREQUESTTYPE_MEMORY))
	}

	gpu := workloadsApi.NewRequestResourceCores()

	// The total workload GPU request is (per-device fraction) * (number of devices).
	// Falls back to just NumDevices when no fractional portion is provided (whole-GPU
	// requests), and leaves out request when NumDevices is unset.
	if src.NumDevices != nil {
		numDevices := float64(*src.NumDevices)
		if src.PortionRequest != nil && *src.PortionRequest > 0 && *src.PortionRequest <= 1 {
			gpu.SetRequest(*src.PortionRequest * numDevices)
		} else {
			gpu.SetRequest(numDevices)
		}
	}

	if src.PortionLimit != nil {
		limit := *src.PortionLimit
		if src.NumDevices != nil && limit > 0 && limit <= 1 {
			limit = limit * float64(*src.NumDevices)
		}
		gpu.SetLimit(limit)
	}

	requestedResources.Gpu.Set(gpu)

	if src.MemoryRequest != nil || src.MemoryLimit != nil {
		gpuMemory := workloadsApi.NewRequestResourceQuantity()

		if src.MemoryRequest != nil {
			gpuMemory.SetRequest(*src.MemoryRequest)
		}

		if src.MemoryLimit != nil {
			gpuMemory.SetLimit(*src.MemoryLimit)
		}

		requestedResources.GpuMemory.Set(gpuMemory)
	}

	return nil
}

func requestedPodsSetter(src *workloadsApi.RequestedPods, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(**workloadsApi.RequestedPods)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("requestedPodsSetter: target type %T is not **workloadsApi.RequestedPods", target))
	}
	if *targetPtr == nil {
		*targetPtr = src
		return nil
	}

	existingNumber := convutils.Deref((*targetPtr).Number.Get(), 0)
	updatedNumber := convutils.Deref(src.Number.Get(), 0)
	(*targetPtr).Number.Set(convutils.Ref(max(existingNumber, updatedNumber)))

	return nil
}
