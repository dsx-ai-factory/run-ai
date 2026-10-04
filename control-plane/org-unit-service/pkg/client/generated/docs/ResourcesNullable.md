# ResourcesNullable

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePool** | Pointer to [**NullableResourcesNodePool**](ResourcesNodePool.md) |  | [optional] 
**Gpu** | Pointer to [**ResourcesNullableGpu**](ResourcesNullableGpu.md) |  | [optional] 
**Cpu** | Pointer to [**NullableResourcesCpu**](ResourcesCpu.md) |  | [optional] 
**Memory** | Pointer to [**NullableResourcesMemory**](ResourcesMemory.md) |  | [optional] 
**Rank** | Pointer to **NullableString** | The scheduling rank of this org-unit within the node pool. Determines the scheduling order of workloads between org-units. Workloads in higher-ranked org-units are scheduled before workloads in lower-ranked org-units. For example, all workloads in rank \&quot;High\&quot; will be scheduled before all workloads in rank \&quot;Medium\&quot;. | [optional] [default to "MediumLow"]

## Methods

### NewResourcesNullable

`func NewResourcesNullable() *ResourcesNullable`

NewResourcesNullable instantiates a new ResourcesNullable object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesNullableWithDefaults

`func NewResourcesNullableWithDefaults() *ResourcesNullable`

NewResourcesNullableWithDefaults instantiates a new ResourcesNullable object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePool

`func (o *ResourcesNullable) GetNodePool() ResourcesNodePool`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *ResourcesNullable) GetNodePoolOk() (*ResourcesNodePool, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *ResourcesNullable) SetNodePool(v ResourcesNodePool)`

SetNodePool sets NodePool field to given value.

### HasNodePool

`func (o *ResourcesNullable) HasNodePool() bool`

HasNodePool returns a boolean if a field has been set.

### SetNodePoolNil

`func (o *ResourcesNullable) SetNodePoolNil(b bool)`

 SetNodePoolNil sets the value for NodePool to be an explicit nil

### UnsetNodePool
`func (o *ResourcesNullable) UnsetNodePool()`

UnsetNodePool ensures that no value is present for NodePool, not even an explicit nil
### GetGpu

`func (o *ResourcesNullable) GetGpu() ResourcesNullableGpu`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *ResourcesNullable) GetGpuOk() (*ResourcesNullableGpu, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *ResourcesNullable) SetGpu(v ResourcesNullableGpu)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *ResourcesNullable) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetCpu

`func (o *ResourcesNullable) GetCpu() ResourcesCpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *ResourcesNullable) GetCpuOk() (*ResourcesCpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *ResourcesNullable) SetCpu(v ResourcesCpu)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *ResourcesNullable) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### SetCpuNil

`func (o *ResourcesNullable) SetCpuNil(b bool)`

 SetCpuNil sets the value for Cpu to be an explicit nil

### UnsetCpu
`func (o *ResourcesNullable) UnsetCpu()`

UnsetCpu ensures that no value is present for Cpu, not even an explicit nil
### GetMemory

`func (o *ResourcesNullable) GetMemory() ResourcesMemory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *ResourcesNullable) GetMemoryOk() (*ResourcesMemory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *ResourcesNullable) SetMemory(v ResourcesMemory)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *ResourcesNullable) HasMemory() bool`

HasMemory returns a boolean if a field has been set.

### SetMemoryNil

`func (o *ResourcesNullable) SetMemoryNil(b bool)`

 SetMemoryNil sets the value for Memory to be an explicit nil

### UnsetMemory
`func (o *ResourcesNullable) UnsetMemory()`

UnsetMemory ensures that no value is present for Memory, not even an explicit nil
### GetRank

`func (o *ResourcesNullable) GetRank() string`

GetRank returns the Rank field if non-nil, zero value otherwise.

### GetRankOk

`func (o *ResourcesNullable) GetRankOk() (*string, bool)`

GetRankOk returns a tuple with the Rank field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRank

`func (o *ResourcesNullable) SetRank(v string)`

SetRank sets Rank field to given value.

### HasRank

`func (o *ResourcesNullable) HasRank() bool`

HasRank returns a boolean if a field has been set.

### SetRankNil

`func (o *ResourcesNullable) SetRankNil(b bool)`

 SetRankNil sets the value for Rank to be an explicit nil

### UnsetRank
`func (o *ResourcesNullable) UnsetRank()`

UnsetRank ensures that no value is present for Rank, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


