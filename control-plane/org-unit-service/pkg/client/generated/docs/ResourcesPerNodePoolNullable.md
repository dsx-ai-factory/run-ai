# ResourcesPerNodePoolNullable

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Gpu** | Pointer to [**NullableResourcesPerNodePoolNullableGpu**](ResourcesPerNodePoolNullableGpu.md) |  | [optional] 
**Cpu** | Pointer to [**NullableResourcesCpu**](ResourcesCpu.md) |  | [optional] 
**Memory** | Pointer to [**NullableResourcesMemory**](ResourcesMemory.md) |  | [optional] 

## Methods

### NewResourcesPerNodePoolNullable

`func NewResourcesPerNodePoolNullable() *ResourcesPerNodePoolNullable`

NewResourcesPerNodePoolNullable instantiates a new ResourcesPerNodePoolNullable object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesPerNodePoolNullableWithDefaults

`func NewResourcesPerNodePoolNullableWithDefaults() *ResourcesPerNodePoolNullable`

NewResourcesPerNodePoolNullableWithDefaults instantiates a new ResourcesPerNodePoolNullable object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpu

`func (o *ResourcesPerNodePoolNullable) GetGpu() ResourcesPerNodePoolNullableGpu`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *ResourcesPerNodePoolNullable) GetGpuOk() (*ResourcesPerNodePoolNullableGpu, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *ResourcesPerNodePoolNullable) SetGpu(v ResourcesPerNodePoolNullableGpu)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *ResourcesPerNodePoolNullable) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### SetGpuNil

`func (o *ResourcesPerNodePoolNullable) SetGpuNil(b bool)`

 SetGpuNil sets the value for Gpu to be an explicit nil

### UnsetGpu
`func (o *ResourcesPerNodePoolNullable) UnsetGpu()`

UnsetGpu ensures that no value is present for Gpu, not even an explicit nil
### GetCpu

`func (o *ResourcesPerNodePoolNullable) GetCpu() ResourcesCpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *ResourcesPerNodePoolNullable) GetCpuOk() (*ResourcesCpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *ResourcesPerNodePoolNullable) SetCpu(v ResourcesCpu)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *ResourcesPerNodePoolNullable) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### SetCpuNil

`func (o *ResourcesPerNodePoolNullable) SetCpuNil(b bool)`

 SetCpuNil sets the value for Cpu to be an explicit nil

### UnsetCpu
`func (o *ResourcesPerNodePoolNullable) UnsetCpu()`

UnsetCpu ensures that no value is present for Cpu, not even an explicit nil
### GetMemory

`func (o *ResourcesPerNodePoolNullable) GetMemory() ResourcesMemory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *ResourcesPerNodePoolNullable) GetMemoryOk() (*ResourcesMemory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *ResourcesPerNodePoolNullable) SetMemory(v ResourcesMemory)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *ResourcesPerNodePoolNullable) HasMemory() bool`

HasMemory returns a boolean if a field has been set.

### SetMemoryNil

`func (o *ResourcesPerNodePoolNullable) SetMemoryNil(b bool)`

 SetMemoryNil sets the value for Memory to be an explicit nil

### UnsetMemory
`func (o *ResourcesPerNodePoolNullable) UnsetMemory()`

UnsetMemory ensures that no value is present for Memory, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


