# ResourcesSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePool** | Pointer to [**NullableResourcesNodePool**](ResourcesNodePool.md) |  | [optional] 
**Gpu** | [**ResourcesSyncGpu**](ResourcesSyncGpu.md) |  | 
**Cpu** | Pointer to [**NullableResourcesCpu**](ResourcesCpu.md) |  | [optional] 
**Memory** | Pointer to [**NullableResourcesSyncMemory**](ResourcesSyncMemory.md) |  | [optional] 
**Rank** | Pointer to **int32** |  | [optional] 
**Priority** | Pointer to **int32** |  | [optional] 

## Methods

### NewResourcesSync

`func NewResourcesSync(gpu ResourcesSyncGpu, ) *ResourcesSync`

NewResourcesSync instantiates a new ResourcesSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesSyncWithDefaults

`func NewResourcesSyncWithDefaults() *ResourcesSync`

NewResourcesSyncWithDefaults instantiates a new ResourcesSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePool

`func (o *ResourcesSync) GetNodePool() ResourcesNodePool`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *ResourcesSync) GetNodePoolOk() (*ResourcesNodePool, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *ResourcesSync) SetNodePool(v ResourcesNodePool)`

SetNodePool sets NodePool field to given value.

### HasNodePool

`func (o *ResourcesSync) HasNodePool() bool`

HasNodePool returns a boolean if a field has been set.

### SetNodePoolNil

`func (o *ResourcesSync) SetNodePoolNil(b bool)`

 SetNodePoolNil sets the value for NodePool to be an explicit nil

### UnsetNodePool
`func (o *ResourcesSync) UnsetNodePool()`

UnsetNodePool ensures that no value is present for NodePool, not even an explicit nil
### GetGpu

`func (o *ResourcesSync) GetGpu() ResourcesSyncGpu`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *ResourcesSync) GetGpuOk() (*ResourcesSyncGpu, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *ResourcesSync) SetGpu(v ResourcesSyncGpu)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *ResourcesSync) GetCpu() ResourcesCpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *ResourcesSync) GetCpuOk() (*ResourcesCpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *ResourcesSync) SetCpu(v ResourcesCpu)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *ResourcesSync) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### SetCpuNil

`func (o *ResourcesSync) SetCpuNil(b bool)`

 SetCpuNil sets the value for Cpu to be an explicit nil

### UnsetCpu
`func (o *ResourcesSync) UnsetCpu()`

UnsetCpu ensures that no value is present for Cpu, not even an explicit nil
### GetMemory

`func (o *ResourcesSync) GetMemory() ResourcesSyncMemory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *ResourcesSync) GetMemoryOk() (*ResourcesSyncMemory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *ResourcesSync) SetMemory(v ResourcesSyncMemory)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *ResourcesSync) HasMemory() bool`

HasMemory returns a boolean if a field has been set.

### SetMemoryNil

`func (o *ResourcesSync) SetMemoryNil(b bool)`

 SetMemoryNil sets the value for Memory to be an explicit nil

### UnsetMemory
`func (o *ResourcesSync) UnsetMemory()`

UnsetMemory ensures that no value is present for Memory, not even an explicit nil
### GetRank

`func (o *ResourcesSync) GetRank() int32`

GetRank returns the Rank field if non-nil, zero value otherwise.

### GetRankOk

`func (o *ResourcesSync) GetRankOk() (*int32, bool)`

GetRankOk returns a tuple with the Rank field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRank

`func (o *ResourcesSync) SetRank(v int32)`

SetRank sets Rank field to given value.

### HasRank

`func (o *ResourcesSync) HasRank() bool`

HasRank returns a boolean if a field has been set.

### GetPriority

`func (o *ResourcesSync) GetPriority() int32`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *ResourcesSync) GetPriorityOk() (*int32, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *ResourcesSync) SetPriority(v int32)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *ResourcesSync) HasPriority() bool`

HasPriority returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


