# NodePoolAssignedResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **float32** | The id of the assigned resources. Required in PUT when updating the assigned resources. | 
**NodePool** | [**NodePoolAssignedResourcesCreateNodePool**](NodePoolAssignedResourcesCreateNodePool.md) |  | 
**Gpu** | Pointer to [**NodePoolAssignedResourcesCreateGpu**](NodePoolAssignedResourcesCreateGpu.md) |  | [optional] 
**Cpu** | Pointer to [**NodePoolAssignedResourcesCreateCpu**](NodePoolAssignedResourcesCreateCpu.md) |  | [optional] 
**Memory** | Pointer to [**NodePoolAssignedResourcesCreateMemory**](NodePoolAssignedResourcesCreateMemory.md) |  | [optional] 

## Methods

### NewNodePoolAssignedResources

`func NewNodePoolAssignedResources(id float32, nodePool NodePoolAssignedResourcesCreateNodePool, ) *NodePoolAssignedResources`

NewNodePoolAssignedResources instantiates a new NodePoolAssignedResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolAssignedResourcesWithDefaults

`func NewNodePoolAssignedResourcesWithDefaults() *NodePoolAssignedResources`

NewNodePoolAssignedResourcesWithDefaults instantiates a new NodePoolAssignedResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *NodePoolAssignedResources) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NodePoolAssignedResources) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NodePoolAssignedResources) SetId(v float32)`

SetId sets Id field to given value.


### GetNodePool

`func (o *NodePoolAssignedResources) GetNodePool() NodePoolAssignedResourcesCreateNodePool`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *NodePoolAssignedResources) GetNodePoolOk() (*NodePoolAssignedResourcesCreateNodePool, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *NodePoolAssignedResources) SetNodePool(v NodePoolAssignedResourcesCreateNodePool)`

SetNodePool sets NodePool field to given value.


### GetGpu

`func (o *NodePoolAssignedResources) GetGpu() NodePoolAssignedResourcesCreateGpu`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *NodePoolAssignedResources) GetGpuOk() (*NodePoolAssignedResourcesCreateGpu, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *NodePoolAssignedResources) SetGpu(v NodePoolAssignedResourcesCreateGpu)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *NodePoolAssignedResources) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetCpu

`func (o *NodePoolAssignedResources) GetCpu() NodePoolAssignedResourcesCreateCpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *NodePoolAssignedResources) GetCpuOk() (*NodePoolAssignedResourcesCreateCpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *NodePoolAssignedResources) SetCpu(v NodePoolAssignedResourcesCreateCpu)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *NodePoolAssignedResources) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetMemory

`func (o *NodePoolAssignedResources) GetMemory() NodePoolAssignedResourcesCreateMemory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *NodePoolAssignedResources) GetMemoryOk() (*NodePoolAssignedResourcesCreateMemory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *NodePoolAssignedResources) SetMemory(v NodePoolAssignedResourcesCreateMemory)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *NodePoolAssignedResources) HasMemory() bool`

HasMemory returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


