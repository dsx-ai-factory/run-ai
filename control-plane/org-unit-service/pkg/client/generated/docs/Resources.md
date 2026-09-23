# Resources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePool** | Pointer to [**NullableResourcesNodePool**](ResourcesNodePool.md) |  | [optional] 
**Gpu** | [**ResourcesGpu**](ResourcesGpu.md) |  | 
**Cpu** | Pointer to [**NullableResourcesCpu**](ResourcesCpu.md) |  | [optional] 
**Memory** | Pointer to [**NullableResourcesMemory**](ResourcesMemory.md) |  | [optional] 
**Rank** | Pointer to **NullableString** | The scheduling rank of this org-unit within the node pool. Determines the scheduling order of workloads between org-units. Workloads in higher-ranked org-units are scheduled before workloads in lower-ranked org-units. For example, all workloads in rank \&quot;High\&quot; will be scheduled before all workloads in rank \&quot;Medium\&quot;. | [optional] [default to "MediumLow"]
**Priority** | Pointer to **NullableString** | Deprecated. Use rank instead. The scheduling priority for workloads in this node pool. | [optional] 

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
### GetRank

`func (o *Resources) GetRank() string`

GetRank returns the Rank field if non-nil, zero value otherwise.

### GetRankOk

`func (o *Resources) GetRankOk() (*string, bool)`

GetRankOk returns a tuple with the Rank field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRank

`func (o *Resources) SetRank(v string)`

SetRank sets Rank field to given value.

### HasRank

`func (o *Resources) HasRank() bool`

HasRank returns a boolean if a field has been set.

### SetRankNil

`func (o *Resources) SetRankNil(b bool)`

 SetRankNil sets the value for Rank to be an explicit nil

### UnsetRank
`func (o *Resources) UnsetRank()`

UnsetRank ensures that no value is present for Rank, not even an explicit nil
### GetPriority

`func (o *Resources) GetPriority() string`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *Resources) GetPriorityOk() (*string, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *Resources) SetPriority(v string)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *Resources) HasPriority() bool`

HasPriority returns a boolean if a field has been set.

### SetPriorityNil

`func (o *Resources) SetPriorityNil(b bool)`

 SetPriorityNil sets the value for Priority to be an explicit nil

### UnsetPriority
`func (o *Resources) UnsetPriority()`

UnsetPriority ensures that no value is present for Priority, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


