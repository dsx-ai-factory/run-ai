# ResourcesPerNodePool

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Gpu** | Pointer to [**ResourcesPerNodePoolGpu**](ResourcesPerNodePoolGpu.md) |  | [optional] 
**Cpu** | Pointer to [**NullableResourcesCpu**](ResourcesCpu.md) |  | [optional] 
**Memory** | Pointer to [**NullableResourcesMemory**](ResourcesMemory.md) |  | [optional] 

## Methods

### NewResourcesPerNodePool

`func NewResourcesPerNodePool() *ResourcesPerNodePool`

NewResourcesPerNodePool instantiates a new ResourcesPerNodePool object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesPerNodePoolWithDefaults

`func NewResourcesPerNodePoolWithDefaults() *ResourcesPerNodePool`

NewResourcesPerNodePoolWithDefaults instantiates a new ResourcesPerNodePool object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpu

`func (o *ResourcesPerNodePool) GetGpu() ResourcesPerNodePoolGpu`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *ResourcesPerNodePool) GetGpuOk() (*ResourcesPerNodePoolGpu, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *ResourcesPerNodePool) SetGpu(v ResourcesPerNodePoolGpu)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *ResourcesPerNodePool) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetCpu

`func (o *ResourcesPerNodePool) GetCpu() ResourcesCpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *ResourcesPerNodePool) GetCpuOk() (*ResourcesCpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *ResourcesPerNodePool) SetCpu(v ResourcesCpu)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *ResourcesPerNodePool) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### SetCpuNil

`func (o *ResourcesPerNodePool) SetCpuNil(b bool)`

 SetCpuNil sets the value for Cpu to be an explicit nil

### UnsetCpu
`func (o *ResourcesPerNodePool) UnsetCpu()`

UnsetCpu ensures that no value is present for Cpu, not even an explicit nil
### GetMemory

`func (o *ResourcesPerNodePool) GetMemory() ResourcesMemory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *ResourcesPerNodePool) GetMemoryOk() (*ResourcesMemory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *ResourcesPerNodePool) SetMemory(v ResourcesMemory)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *ResourcesPerNodePool) HasMemory() bool`

HasMemory returns a boolean if a field has been set.

### SetMemoryNil

`func (o *ResourcesPerNodePool) SetMemoryNil(b bool)`

 SetMemoryNil sets the value for Memory to be an explicit nil

### UnsetMemory
`func (o *ResourcesPerNodePool) UnsetMemory()`

UnsetMemory ensures that no value is present for Memory, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


