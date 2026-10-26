# NodePoolAssignedResourcesV1Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **float32** | The id of the assigned resources. Required in PUT when updating the assigned resources. | 
**NodePool** | [**NodePoolAssignedResourcesV1ResponseNodePool**](NodePoolAssignedResourcesV1ResponseNodePool.md) |  | 
**Gpu** | [**ResourceV1Response**](ResourceV1Response.md) | Number of GPUs assigned in the node pool. | 
**Cpu** | [**ResourceV1Response**](ResourceV1Response.md) | Number of CPU Millicores assigned in the node pool. Supported only if the &#39;CPU Resources Quota&#39; feature flag is enabled. | 
**Memory** | [**ResourceV1Response**](ResourceV1Response.md) | Amount of CPU Memory Mib assigned in the node pool. Supported only if the &#39;CPU Resources Quota&#39; feature flag is enabled. | 

## Methods

### NewNodePoolAssignedResourcesV1Response

`func NewNodePoolAssignedResourcesV1Response(id float32, nodePool NodePoolAssignedResourcesV1ResponseNodePool, gpu ResourceV1Response, cpu ResourceV1Response, memory ResourceV1Response, ) *NodePoolAssignedResourcesV1Response`

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

`func (o *NodePoolAssignedResourcesV1Response) GetGpu() ResourceV1Response`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *NodePoolAssignedResourcesV1Response) GetGpuOk() (*ResourceV1Response, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *NodePoolAssignedResourcesV1Response) SetGpu(v ResourceV1Response)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *NodePoolAssignedResourcesV1Response) GetCpu() ResourceV1Response`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *NodePoolAssignedResourcesV1Response) GetCpuOk() (*ResourceV1Response, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *NodePoolAssignedResourcesV1Response) SetCpu(v ResourceV1Response)`

SetCpu sets Cpu field to given value.


### GetMemory

`func (o *NodePoolAssignedResourcesV1Response) GetMemory() ResourceV1Response`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *NodePoolAssignedResourcesV1Response) GetMemoryOk() (*ResourceV1Response, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *NodePoolAssignedResourcesV1Response) SetMemory(v ResourceV1Response)`

SetMemory sets Memory field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


