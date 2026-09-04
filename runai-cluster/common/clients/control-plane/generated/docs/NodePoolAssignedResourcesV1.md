# NodePoolAssignedResourcesV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **float32** | The id of the assigned resources. Required in PUT when updating the assigned resources. | 
**NodePool** | [**NodePoolAssignedResourcesV1ResponseNodePool**](NodePoolAssignedResourcesV1ResponseNodePool.md) |  | 
**Gpu** | Pointer to [**NodePoolAssignedResourcesV1Gpu**](NodePoolAssignedResourcesV1Gpu.md) |  | [optional] 
**Cpu** | Pointer to [**NodePoolAssignedResourcesV1Cpu**](NodePoolAssignedResourcesV1Cpu.md) |  | [optional] 
**Memory** | Pointer to [**NodePoolAssignedResourcesV1Memory**](NodePoolAssignedResourcesV1Memory.md) |  | [optional] 

## Methods

### NewNodePoolAssignedResourcesV1

`func NewNodePoolAssignedResourcesV1(id float32, nodePool NodePoolAssignedResourcesV1ResponseNodePool, ) *NodePoolAssignedResourcesV1`

NewNodePoolAssignedResourcesV1 instantiates a new NodePoolAssignedResourcesV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolAssignedResourcesV1WithDefaults

`func NewNodePoolAssignedResourcesV1WithDefaults() *NodePoolAssignedResourcesV1`

NewNodePoolAssignedResourcesV1WithDefaults instantiates a new NodePoolAssignedResourcesV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *NodePoolAssignedResourcesV1) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NodePoolAssignedResourcesV1) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NodePoolAssignedResourcesV1) SetId(v float32)`

SetId sets Id field to given value.


### GetNodePool

`func (o *NodePoolAssignedResourcesV1) GetNodePool() NodePoolAssignedResourcesV1ResponseNodePool`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *NodePoolAssignedResourcesV1) GetNodePoolOk() (*NodePoolAssignedResourcesV1ResponseNodePool, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *NodePoolAssignedResourcesV1) SetNodePool(v NodePoolAssignedResourcesV1ResponseNodePool)`

SetNodePool sets NodePool field to given value.


### GetGpu

`func (o *NodePoolAssignedResourcesV1) GetGpu() NodePoolAssignedResourcesV1Gpu`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *NodePoolAssignedResourcesV1) GetGpuOk() (*NodePoolAssignedResourcesV1Gpu, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *NodePoolAssignedResourcesV1) SetGpu(v NodePoolAssignedResourcesV1Gpu)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *NodePoolAssignedResourcesV1) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetCpu

`func (o *NodePoolAssignedResourcesV1) GetCpu() NodePoolAssignedResourcesV1Cpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *NodePoolAssignedResourcesV1) GetCpuOk() (*NodePoolAssignedResourcesV1Cpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *NodePoolAssignedResourcesV1) SetCpu(v NodePoolAssignedResourcesV1Cpu)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *NodePoolAssignedResourcesV1) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetMemory

`func (o *NodePoolAssignedResourcesV1) GetMemory() NodePoolAssignedResourcesV1Memory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *NodePoolAssignedResourcesV1) GetMemoryOk() (*NodePoolAssignedResourcesV1Memory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *NodePoolAssignedResourcesV1) SetMemory(v NodePoolAssignedResourcesV1Memory)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *NodePoolAssignedResourcesV1) HasMemory() bool`

HasMemory returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


