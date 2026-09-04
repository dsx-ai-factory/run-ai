# Resources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePool** | Pointer to [**NullableResourcesNodePool**](ResourcesNodePool.md) |  | [optional] 
**Gpu** | [**ResourcesGpu**](ResourcesGpu.md) |  | 
**Cpu** | Pointer to [**NullableResourcesCpu**](ResourcesCpu.md) |  | [optional] 
**Memory** | Pointer to [**NullableResourcesMemory**](ResourcesMemory.md) |  | [optional] 

## Methods

### NewResources

`func NewResources(gpu ResourcesGpu, ) *Resources`

NewResources instantiates a new Resources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesWithDefaults

`func NewResourcesWithDefaults() *Resources`

NewResourcesWithDefaults instantiates a new Resources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePool

`func (o *Resources) GetNodePool() ResourcesNodePool`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *Resources) GetNodePoolOk() (*ResourcesNodePool, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *Resources) SetNodePool(v ResourcesNodePool)`

SetNodePool sets NodePool field to given value.

### HasNodePool

`func (o *Resources) HasNodePool() bool`

HasNodePool returns a boolean if a field has been set.

### SetNodePoolNil

`func (o *Resources) SetNodePoolNil(b bool)`

 SetNodePoolNil sets the value for NodePool to be an explicit nil

### UnsetNodePool
`func (o *Resources) UnsetNodePool()`

UnsetNodePool ensures that no value is present for NodePool, not even an explicit nil
### GetGpu

`func (o *Resources) GetGpu() ResourcesGpu`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *Resources) GetGpuOk() (*ResourcesGpu, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *Resources) SetGpu(v ResourcesGpu)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *Resources) GetCpu() ResourcesCpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *Resources) GetCpuOk() (*ResourcesCpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *Resources) SetCpu(v ResourcesCpu)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *Resources) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### SetCpuNil

`func (o *Resources) SetCpuNil(b bool)`

 SetCpuNil sets the value for Cpu to be an explicit nil

### UnsetCpu
`func (o *Resources) UnsetCpu()`

UnsetCpu ensures that no value is present for Cpu, not even an explicit nil
### GetMemory

`func (o *Resources) GetMemory() ResourcesMemory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *Resources) GetMemoryOk() (*ResourcesMemory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *Resources) SetMemory(v ResourcesMemory)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *Resources) HasMemory() bool`

HasMemory returns a boolean if a field has been set.

### SetMemoryNil

`func (o *Resources) SetMemoryNil(b bool)`

 SetMemoryNil sets the value for Memory to be an explicit nil

### UnsetMemory
`func (o *Resources) UnsetMemory()`

UnsetMemory ensures that no value is present for Memory, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


