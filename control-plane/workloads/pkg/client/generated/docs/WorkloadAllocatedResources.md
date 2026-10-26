# WorkloadAllocatedResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Gpu** | Pointer to **NullableFloat64** | Required if and only if gpuRequestType is portion. States the number of GPUs allocated for the created workload. The default is no allocated GPUs. | [optional] 
**GpuMemory** | Pointer to **NullableString** |  | [optional] 
**Cpu** | Pointer to **NullableFloat64** | States the amount of CPU cores used by the workload running. | [optional] 
**CpuMemory** | Pointer to **NullableString** |  | [optional] 
**ExtendedResources** | Pointer to [**[]WorkloadsExtendedResource**](WorkloadsExtendedResource.md) | Set of extended resources with their quantity | [optional] 

## Methods

### NewWorkloadAllocatedResources

`func NewWorkloadAllocatedResources() *WorkloadAllocatedResources`

NewWorkloadAllocatedResources instantiates a new WorkloadAllocatedResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadAllocatedResourcesWithDefaults

`func NewWorkloadAllocatedResourcesWithDefaults() *WorkloadAllocatedResources`

NewWorkloadAllocatedResourcesWithDefaults instantiates a new WorkloadAllocatedResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpu

`func (o *WorkloadAllocatedResources) GetGpu() float64`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *WorkloadAllocatedResources) GetGpuOk() (*float64, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *WorkloadAllocatedResources) SetGpu(v float64)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *WorkloadAllocatedResources) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### SetGpuNil

`func (o *WorkloadAllocatedResources) SetGpuNil(b bool)`

 SetGpuNil sets the value for Gpu to be an explicit nil

### UnsetGpu
`func (o *WorkloadAllocatedResources) UnsetGpu()`

UnsetGpu ensures that no value is present for Gpu, not even an explicit nil
### GetGpuMemory

`func (o *WorkloadAllocatedResources) GetGpuMemory() string`

GetGpuMemory returns the GpuMemory field if non-nil, zero value otherwise.

### GetGpuMemoryOk

`func (o *WorkloadAllocatedResources) GetGpuMemoryOk() (*string, bool)`

GetGpuMemoryOk returns a tuple with the GpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemory

`func (o *WorkloadAllocatedResources) SetGpuMemory(v string)`

SetGpuMemory sets GpuMemory field to given value.

### HasGpuMemory

`func (o *WorkloadAllocatedResources) HasGpuMemory() bool`

HasGpuMemory returns a boolean if a field has been set.

### SetGpuMemoryNil

`func (o *WorkloadAllocatedResources) SetGpuMemoryNil(b bool)`

 SetGpuMemoryNil sets the value for GpuMemory to be an explicit nil

### UnsetGpuMemory
`func (o *WorkloadAllocatedResources) UnsetGpuMemory()`

UnsetGpuMemory ensures that no value is present for GpuMemory, not even an explicit nil
### GetCpu

`func (o *WorkloadAllocatedResources) GetCpu() float64`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *WorkloadAllocatedResources) GetCpuOk() (*float64, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *WorkloadAllocatedResources) SetCpu(v float64)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *WorkloadAllocatedResources) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### SetCpuNil

`func (o *WorkloadAllocatedResources) SetCpuNil(b bool)`

 SetCpuNil sets the value for Cpu to be an explicit nil

### UnsetCpu
`func (o *WorkloadAllocatedResources) UnsetCpu()`

UnsetCpu ensures that no value is present for Cpu, not even an explicit nil
### GetCpuMemory

`func (o *WorkloadAllocatedResources) GetCpuMemory() string`

GetCpuMemory returns the CpuMemory field if non-nil, zero value otherwise.

### GetCpuMemoryOk

`func (o *WorkloadAllocatedResources) GetCpuMemoryOk() (*string, bool)`

GetCpuMemoryOk returns a tuple with the CpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemory

`func (o *WorkloadAllocatedResources) SetCpuMemory(v string)`

SetCpuMemory sets CpuMemory field to given value.

### HasCpuMemory

`func (o *WorkloadAllocatedResources) HasCpuMemory() bool`

HasCpuMemory returns a boolean if a field has been set.

### SetCpuMemoryNil

`func (o *WorkloadAllocatedResources) SetCpuMemoryNil(b bool)`

 SetCpuMemoryNil sets the value for CpuMemory to be an explicit nil

### UnsetCpuMemory
`func (o *WorkloadAllocatedResources) UnsetCpuMemory()`

UnsetCpuMemory ensures that no value is present for CpuMemory, not even an explicit nil
### GetExtendedResources

`func (o *WorkloadAllocatedResources) GetExtendedResources() []WorkloadsExtendedResource`

GetExtendedResources returns the ExtendedResources field if non-nil, zero value otherwise.

### GetExtendedResourcesOk

`func (o *WorkloadAllocatedResources) GetExtendedResourcesOk() (*[]WorkloadsExtendedResource, bool)`

GetExtendedResourcesOk returns a tuple with the ExtendedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExtendedResources

`func (o *WorkloadAllocatedResources) SetExtendedResources(v []WorkloadsExtendedResource)`

SetExtendedResources sets ExtendedResources field to given value.

### HasExtendedResources

`func (o *WorkloadAllocatedResources) HasExtendedResources() bool`

HasExtendedResources returns a boolean if a field has been set.

### SetExtendedResourcesNil

`func (o *WorkloadAllocatedResources) SetExtendedResourcesNil(b bool)`

 SetExtendedResourcesNil sets the value for ExtendedResources to be an explicit nil

### UnsetExtendedResources
`func (o *WorkloadAllocatedResources) UnsetExtendedResources()`

UnsetExtendedResources ensures that no value is present for ExtendedResources, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


