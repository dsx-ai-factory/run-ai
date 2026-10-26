# ProjectAggregatedResourcesV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **float32** | The id of the assigned resources. Required in PUT when updating the assigned resources. | 
**NodePool** | [**ProjectAggregatedResourcesV1NodePool**](ProjectAggregatedResourcesV1NodePool.md) |  | 
**Gpu** | Pointer to [**ProjectAggregatedResourcesV1Gpu**](ProjectAggregatedResourcesV1Gpu.md) |  | [optional] 
**Cpu** | Pointer to [**ProjectAggregatedResourcesV1Cpu**](ProjectAggregatedResourcesV1Cpu.md) |  | [optional] 
**Memory** | Pointer to [**ProjectAggregatedResourcesV1Memory**](ProjectAggregatedResourcesV1Memory.md) |  | [optional] 

## Methods

### NewProjectAggregatedResourcesV1

`func NewProjectAggregatedResourcesV1(id float32, nodePool ProjectAggregatedResourcesV1NodePool, ) *ProjectAggregatedResourcesV1`

NewProjectAggregatedResourcesV1 instantiates a new ProjectAggregatedResourcesV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectAggregatedResourcesV1WithDefaults

`func NewProjectAggregatedResourcesV1WithDefaults() *ProjectAggregatedResourcesV1`

NewProjectAggregatedResourcesV1WithDefaults instantiates a new ProjectAggregatedResourcesV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ProjectAggregatedResourcesV1) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ProjectAggregatedResourcesV1) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ProjectAggregatedResourcesV1) SetId(v float32)`

SetId sets Id field to given value.


### GetNodePool

`func (o *ProjectAggregatedResourcesV1) GetNodePool() ProjectAggregatedResourcesV1NodePool`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *ProjectAggregatedResourcesV1) GetNodePoolOk() (*ProjectAggregatedResourcesV1NodePool, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *ProjectAggregatedResourcesV1) SetNodePool(v ProjectAggregatedResourcesV1NodePool)`

SetNodePool sets NodePool field to given value.


### GetGpu

`func (o *ProjectAggregatedResourcesV1) GetGpu() ProjectAggregatedResourcesV1Gpu`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *ProjectAggregatedResourcesV1) GetGpuOk() (*ProjectAggregatedResourcesV1Gpu, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *ProjectAggregatedResourcesV1) SetGpu(v ProjectAggregatedResourcesV1Gpu)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *ProjectAggregatedResourcesV1) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetCpu

`func (o *ProjectAggregatedResourcesV1) GetCpu() ProjectAggregatedResourcesV1Cpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *ProjectAggregatedResourcesV1) GetCpuOk() (*ProjectAggregatedResourcesV1Cpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *ProjectAggregatedResourcesV1) SetCpu(v ProjectAggregatedResourcesV1Cpu)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *ProjectAggregatedResourcesV1) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetMemory

`func (o *ProjectAggregatedResourcesV1) GetMemory() ProjectAggregatedResourcesV1Memory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *ProjectAggregatedResourcesV1) GetMemoryOk() (*ProjectAggregatedResourcesV1Memory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *ProjectAggregatedResourcesV1) SetMemory(v ProjectAggregatedResourcesV1Memory)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *ProjectAggregatedResourcesV1) HasMemory() bool`

HasMemory returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


