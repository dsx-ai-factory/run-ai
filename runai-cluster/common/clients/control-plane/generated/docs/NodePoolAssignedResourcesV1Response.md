# NodePoolAssignedResourcesV1Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **float32** | The id of the assigned resources. Required in PUT when updating the assigned resources. | 
**NodePool** | [**NodePoolAssignedResourcesV1ResponseNodePool**](NodePoolAssignedResourcesV1ResponseNodePool.md) |  | 
**Gpu** | [**NodePoolAssignedResourcesV1ResponseGpu**](NodePoolAssignedResourcesV1ResponseGpu.md) |  | 
**Cpu** | [**NodePoolAssignedResourcesV1ResponseCpu**](NodePoolAssignedResourcesV1ResponseCpu.md) |  | 
**Memory** | [**NodePoolAssignedResourcesV1ResponseMemory**](NodePoolAssignedResourcesV1ResponseMemory.md) |  | 

## Methods

### NewNodePoolAssignedResourcesV1Response

`func NewNodePoolAssignedResourcesV1Response(id float32, nodePool NodePoolAssignedResourcesV1ResponseNodePool, gpu NodePoolAssignedResourcesV1ResponseGpu, cpu NodePoolAssignedResourcesV1ResponseCpu, memory NodePoolAssignedResourcesV1ResponseMemory, ) *NodePoolAssignedResourcesV1Response`

NewNodePoolAssignedResourcesV1Response instantiates a new NodePoolAssignedResourcesV1Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolAssignedResourcesV1ResponseWithDefaults

`func NewNodePoolAssignedResourcesV1ResponseWithDefaults() *NodePoolAssignedResourcesV1Response`

NewNodePoolAssignedResourcesV1ResponseWithDefaults instantiates a new NodePoolAssignedResourcesV1Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *NodePoolAssignedResourcesV1Response) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NodePoolAssignedResourcesV1Response) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NodePoolAssignedResourcesV1Response) SetId(v float32)`

SetId sets Id field to given value.


### GetNodePool

`func (o *NodePoolAssignedResourcesV1Response) GetNodePool() NodePoolAssignedResourcesV1ResponseNodePool`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *NodePoolAssignedResourcesV1Response) GetNodePoolOk() (*NodePoolAssignedResourcesV1ResponseNodePool, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *NodePoolAssignedResourcesV1Response) SetNodePool(v NodePoolAssignedResourcesV1ResponseNodePool)`

SetNodePool sets NodePool field to given value.


### GetGpu

`func (o *NodePoolAssignedResourcesV1Response) GetGpu() NodePoolAssignedResourcesV1ResponseGpu`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *NodePoolAssignedResourcesV1Response) GetGpuOk() (*NodePoolAssignedResourcesV1ResponseGpu, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *NodePoolAssignedResourcesV1Response) SetGpu(v NodePoolAssignedResourcesV1ResponseGpu)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *NodePoolAssignedResourcesV1Response) GetCpu() NodePoolAssignedResourcesV1ResponseCpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *NodePoolAssignedResourcesV1Response) GetCpuOk() (*NodePoolAssignedResourcesV1ResponseCpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *NodePoolAssignedResourcesV1Response) SetCpu(v NodePoolAssignedResourcesV1ResponseCpu)`

SetCpu sets Cpu field to given value.


### GetMemory

`func (o *NodePoolAssignedResourcesV1Response) GetMemory() NodePoolAssignedResourcesV1ResponseMemory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *NodePoolAssignedResourcesV1Response) GetMemoryOk() (*NodePoolAssignedResourcesV1ResponseMemory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *NodePoolAssignedResourcesV1Response) SetMemory(v NodePoolAssignedResourcesV1ResponseMemory)`

SetMemory sets Memory field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


