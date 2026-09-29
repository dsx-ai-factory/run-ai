# AIApplicationAllocatedResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Gpu** | Pointer to **NullableFloat64** | States the number of GPUs allocated for the created workloads. | [optional] 
**GpuMemory** | Pointer to **NullableString** |  | [optional] 
**Cpu** | Pointer to **NullableFloat64** | States the amount of CPU cores used by the workloads running. | [optional] 
**CpuMemory** | Pointer to **NullableString** |  | [optional] 

## Methods

### NewAIApplicationAllocatedResources

`func NewAIApplicationAllocatedResources() *AIApplicationAllocatedResources`

NewAIApplicationAllocatedResources instantiates a new AIApplicationAllocatedResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationAllocatedResourcesWithDefaults

`func NewAIApplicationAllocatedResourcesWithDefaults() *AIApplicationAllocatedResources`

NewAIApplicationAllocatedResourcesWithDefaults instantiates a new AIApplicationAllocatedResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpu

`func (o *AIApplicationAllocatedResources) GetGpu() float64`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *AIApplicationAllocatedResources) GetGpuOk() (*float64, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *AIApplicationAllocatedResources) SetGpu(v float64)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *AIApplicationAllocatedResources) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### SetGpuNil

`func (o *AIApplicationAllocatedResources) SetGpuNil(b bool)`

 SetGpuNil sets the value for Gpu to be an explicit nil

### UnsetGpu
`func (o *AIApplicationAllocatedResources) UnsetGpu()`

UnsetGpu ensures that no value is present for Gpu, not even an explicit nil
### GetGpuMemory

`func (o *AIApplicationAllocatedResources) GetGpuMemory() string`

GetGpuMemory returns the GpuMemory field if non-nil, zero value otherwise.

### GetGpuMemoryOk

`func (o *AIApplicationAllocatedResources) GetGpuMemoryOk() (*string, bool)`

GetGpuMemoryOk returns a tuple with the GpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemory

`func (o *AIApplicationAllocatedResources) SetGpuMemory(v string)`

SetGpuMemory sets GpuMemory field to given value.

### HasGpuMemory

`func (o *AIApplicationAllocatedResources) HasGpuMemory() bool`

HasGpuMemory returns a boolean if a field has been set.

### SetGpuMemoryNil

`func (o *AIApplicationAllocatedResources) SetGpuMemoryNil(b bool)`

 SetGpuMemoryNil sets the value for GpuMemory to be an explicit nil

### UnsetGpuMemory
`func (o *AIApplicationAllocatedResources) UnsetGpuMemory()`

UnsetGpuMemory ensures that no value is present for GpuMemory, not even an explicit nil
### GetCpu

`func (o *AIApplicationAllocatedResources) GetCpu() float64`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *AIApplicationAllocatedResources) GetCpuOk() (*float64, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *AIApplicationAllocatedResources) SetCpu(v float64)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *AIApplicationAllocatedResources) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### SetCpuNil

`func (o *AIApplicationAllocatedResources) SetCpuNil(b bool)`

 SetCpuNil sets the value for Cpu to be an explicit nil

### UnsetCpu
`func (o *AIApplicationAllocatedResources) UnsetCpu()`

UnsetCpu ensures that no value is present for Cpu, not even an explicit nil
### GetCpuMemory

`func (o *AIApplicationAllocatedResources) GetCpuMemory() string`

GetCpuMemory returns the CpuMemory field if non-nil, zero value otherwise.

### GetCpuMemoryOk

`func (o *AIApplicationAllocatedResources) GetCpuMemoryOk() (*string, bool)`

GetCpuMemoryOk returns a tuple with the CpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemory

`func (o *AIApplicationAllocatedResources) SetCpuMemory(v string)`

SetCpuMemory sets CpuMemory field to given value.

### HasCpuMemory

`func (o *AIApplicationAllocatedResources) HasCpuMemory() bool`

HasCpuMemory returns a boolean if a field has been set.

### SetCpuMemoryNil

`func (o *AIApplicationAllocatedResources) SetCpuMemoryNil(b bool)`

 SetCpuMemoryNil sets the value for CpuMemory to be an explicit nil

### UnsetCpuMemory
`func (o *AIApplicationAllocatedResources) UnsetCpuMemory()`

UnsetCpuMemory ensures that no value is present for CpuMemory, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


