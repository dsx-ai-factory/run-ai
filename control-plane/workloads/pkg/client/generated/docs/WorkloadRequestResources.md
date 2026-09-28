# WorkloadRequestResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GpuRequestType** | Pointer to [**NullableGpuRequestType**](GpuRequestType.md) |  | [optional] 
**Gpu** | Pointer to [**NullableRequestResourceCores**](RequestResourceCores.md) |  | [optional] 
**GpuMemory** | Pointer to [**NullableRequestResourceQuantity**](RequestResourceQuantity.md) |  | [optional] 
**Cpu** | Pointer to [**NullableRequestResourceCores**](RequestResourceCores.md) |  | [optional] 
**CpuMemory** | Pointer to [**NullableRequestResourceQuantity**](RequestResourceQuantity.md) |  | [optional] 
**ExtendedResources** | Pointer to [**[]WorkloadsExtendedResource**](WorkloadsExtendedResource.md) | Set of extended resources with their quantity | [optional] 

## Methods

### NewWorkloadRequestResources

`func NewWorkloadRequestResources() *WorkloadRequestResources`

NewWorkloadRequestResources instantiates a new WorkloadRequestResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadRequestResourcesWithDefaults

`func NewWorkloadRequestResourcesWithDefaults() *WorkloadRequestResources`

NewWorkloadRequestResourcesWithDefaults instantiates a new WorkloadRequestResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpuRequestType

`func (o *WorkloadRequestResources) GetGpuRequestType() GpuRequestType`

GetGpuRequestType returns the GpuRequestType field if non-nil, zero value otherwise.

### GetGpuRequestTypeOk

`func (o *WorkloadRequestResources) GetGpuRequestTypeOk() (*GpuRequestType, bool)`

GetGpuRequestTypeOk returns a tuple with the GpuRequestType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuRequestType

`func (o *WorkloadRequestResources) SetGpuRequestType(v GpuRequestType)`

SetGpuRequestType sets GpuRequestType field to given value.

### HasGpuRequestType

`func (o *WorkloadRequestResources) HasGpuRequestType() bool`

HasGpuRequestType returns a boolean if a field has been set.

### SetGpuRequestTypeNil

`func (o *WorkloadRequestResources) SetGpuRequestTypeNil(b bool)`

 SetGpuRequestTypeNil sets the value for GpuRequestType to be an explicit nil

### UnsetGpuRequestType
`func (o *WorkloadRequestResources) UnsetGpuRequestType()`

UnsetGpuRequestType ensures that no value is present for GpuRequestType, not even an explicit nil
### GetGpu

`func (o *WorkloadRequestResources) GetGpu() RequestResourceCores`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *WorkloadRequestResources) GetGpuOk() (*RequestResourceCores, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *WorkloadRequestResources) SetGpu(v RequestResourceCores)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *WorkloadRequestResources) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### SetGpuNil

`func (o *WorkloadRequestResources) SetGpuNil(b bool)`

 SetGpuNil sets the value for Gpu to be an explicit nil

### UnsetGpu
`func (o *WorkloadRequestResources) UnsetGpu()`

UnsetGpu ensures that no value is present for Gpu, not even an explicit nil
### GetGpuMemory

`func (o *WorkloadRequestResources) GetGpuMemory() RequestResourceQuantity`

GetGpuMemory returns the GpuMemory field if non-nil, zero value otherwise.

### GetGpuMemoryOk

`func (o *WorkloadRequestResources) GetGpuMemoryOk() (*RequestResourceQuantity, bool)`

GetGpuMemoryOk returns a tuple with the GpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemory

`func (o *WorkloadRequestResources) SetGpuMemory(v RequestResourceQuantity)`

SetGpuMemory sets GpuMemory field to given value.

### HasGpuMemory

`func (o *WorkloadRequestResources) HasGpuMemory() bool`

HasGpuMemory returns a boolean if a field has been set.

### SetGpuMemoryNil

`func (o *WorkloadRequestResources) SetGpuMemoryNil(b bool)`

 SetGpuMemoryNil sets the value for GpuMemory to be an explicit nil

### UnsetGpuMemory
`func (o *WorkloadRequestResources) UnsetGpuMemory()`

UnsetGpuMemory ensures that no value is present for GpuMemory, not even an explicit nil
### GetCpu

`func (o *WorkloadRequestResources) GetCpu() RequestResourceCores`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *WorkloadRequestResources) GetCpuOk() (*RequestResourceCores, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *WorkloadRequestResources) SetCpu(v RequestResourceCores)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *WorkloadRequestResources) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### SetCpuNil

`func (o *WorkloadRequestResources) SetCpuNil(b bool)`

 SetCpuNil sets the value for Cpu to be an explicit nil

### UnsetCpu
`func (o *WorkloadRequestResources) UnsetCpu()`

UnsetCpu ensures that no value is present for Cpu, not even an explicit nil
### GetCpuMemory

`func (o *WorkloadRequestResources) GetCpuMemory() RequestResourceQuantity`

GetCpuMemory returns the CpuMemory field if non-nil, zero value otherwise.

### GetCpuMemoryOk

`func (o *WorkloadRequestResources) GetCpuMemoryOk() (*RequestResourceQuantity, bool)`

GetCpuMemoryOk returns a tuple with the CpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemory

`func (o *WorkloadRequestResources) SetCpuMemory(v RequestResourceQuantity)`

SetCpuMemory sets CpuMemory field to given value.

### HasCpuMemory

`func (o *WorkloadRequestResources) HasCpuMemory() bool`

HasCpuMemory returns a boolean if a field has been set.

### SetCpuMemoryNil

`func (o *WorkloadRequestResources) SetCpuMemoryNil(b bool)`

 SetCpuMemoryNil sets the value for CpuMemory to be an explicit nil

### UnsetCpuMemory
`func (o *WorkloadRequestResources) UnsetCpuMemory()`

UnsetCpuMemory ensures that no value is present for CpuMemory, not even an explicit nil
### GetExtendedResources

`func (o *WorkloadRequestResources) GetExtendedResources() []WorkloadsExtendedResource`

GetExtendedResources returns the ExtendedResources field if non-nil, zero value otherwise.

### GetExtendedResourcesOk

`func (o *WorkloadRequestResources) GetExtendedResourcesOk() (*[]WorkloadsExtendedResource, bool)`

GetExtendedResourcesOk returns a tuple with the ExtendedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExtendedResources

`func (o *WorkloadRequestResources) SetExtendedResources(v []WorkloadsExtendedResource)`

SetExtendedResources sets ExtendedResources field to given value.

### HasExtendedResources

`func (o *WorkloadRequestResources) HasExtendedResources() bool`

HasExtendedResources returns a boolean if a field has been set.

### SetExtendedResourcesNil

`func (o *WorkloadRequestResources) SetExtendedResourcesNil(b bool)`

 SetExtendedResourcesNil sets the value for ExtendedResources to be an explicit nil

### UnsetExtendedResources
`func (o *WorkloadRequestResources) UnsetExtendedResources()`

UnsetExtendedResources ensures that no value is present for ExtendedResources, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


