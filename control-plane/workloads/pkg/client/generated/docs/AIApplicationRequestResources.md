# AIApplicationRequestResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Gpu** | Pointer to [**NullableRequestResourceCores**](RequestResourceCores.md) |  | [optional] 
**GpuMemory** | Pointer to [**NullableRequestResourceQuantity**](RequestResourceQuantity.md) |  | [optional] 
**Cpu** | Pointer to [**NullableRequestResourceCores**](RequestResourceCores.md) |  | [optional] 
**CpuMemory** | Pointer to [**NullableRequestResourceQuantity**](RequestResourceQuantity.md) |  | [optional] 

## Methods

### NewAIApplicationRequestResources

`func NewAIApplicationRequestResources() *AIApplicationRequestResources`

NewAIApplicationRequestResources instantiates a new AIApplicationRequestResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationRequestResourcesWithDefaults

`func NewAIApplicationRequestResourcesWithDefaults() *AIApplicationRequestResources`

NewAIApplicationRequestResourcesWithDefaults instantiates a new AIApplicationRequestResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpu

`func (o *AIApplicationRequestResources) GetGpu() RequestResourceCores`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *AIApplicationRequestResources) GetGpuOk() (*RequestResourceCores, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *AIApplicationRequestResources) SetGpu(v RequestResourceCores)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *AIApplicationRequestResources) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### SetGpuNil

`func (o *AIApplicationRequestResources) SetGpuNil(b bool)`

 SetGpuNil sets the value for Gpu to be an explicit nil

### UnsetGpu
`func (o *AIApplicationRequestResources) UnsetGpu()`

UnsetGpu ensures that no value is present for Gpu, not even an explicit nil
### GetGpuMemory

`func (o *AIApplicationRequestResources) GetGpuMemory() RequestResourceQuantity`

GetGpuMemory returns the GpuMemory field if non-nil, zero value otherwise.

### GetGpuMemoryOk

`func (o *AIApplicationRequestResources) GetGpuMemoryOk() (*RequestResourceQuantity, bool)`

GetGpuMemoryOk returns a tuple with the GpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemory

`func (o *AIApplicationRequestResources) SetGpuMemory(v RequestResourceQuantity)`

SetGpuMemory sets GpuMemory field to given value.

### HasGpuMemory

`func (o *AIApplicationRequestResources) HasGpuMemory() bool`

HasGpuMemory returns a boolean if a field has been set.

### SetGpuMemoryNil

`func (o *AIApplicationRequestResources) SetGpuMemoryNil(b bool)`

 SetGpuMemoryNil sets the value for GpuMemory to be an explicit nil

### UnsetGpuMemory
`func (o *AIApplicationRequestResources) UnsetGpuMemory()`

UnsetGpuMemory ensures that no value is present for GpuMemory, not even an explicit nil
### GetCpu

`func (o *AIApplicationRequestResources) GetCpu() RequestResourceCores`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *AIApplicationRequestResources) GetCpuOk() (*RequestResourceCores, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *AIApplicationRequestResources) SetCpu(v RequestResourceCores)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *AIApplicationRequestResources) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### SetCpuNil

`func (o *AIApplicationRequestResources) SetCpuNil(b bool)`

 SetCpuNil sets the value for Cpu to be an explicit nil

### UnsetCpu
`func (o *AIApplicationRequestResources) UnsetCpu()`

UnsetCpu ensures that no value is present for Cpu, not even an explicit nil
### GetCpuMemory

`func (o *AIApplicationRequestResources) GetCpuMemory() RequestResourceQuantity`

GetCpuMemory returns the CpuMemory field if non-nil, zero value otherwise.

### GetCpuMemoryOk

`func (o *AIApplicationRequestResources) GetCpuMemoryOk() (*RequestResourceQuantity, bool)`

GetCpuMemoryOk returns a tuple with the CpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemory

`func (o *AIApplicationRequestResources) SetCpuMemory(v RequestResourceQuantity)`

SetCpuMemory sets CpuMemory field to given value.

### HasCpuMemory

`func (o *AIApplicationRequestResources) HasCpuMemory() bool`

HasCpuMemory returns a boolean if a field has been set.

### SetCpuMemoryNil

`func (o *AIApplicationRequestResources) SetCpuMemoryNil(b bool)`

 SetCpuMemoryNil sets the value for CpuMemory to be an explicit nil

### UnsetCpuMemory
`func (o *AIApplicationRequestResources) UnsetCpuMemory()`

UnsetCpuMemory ensures that no value is present for CpuMemory, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


