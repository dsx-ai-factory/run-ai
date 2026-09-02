# NodePoolAssignedResourcesCreate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePool** | [**NodePoolResources**](NodePoolResources.md) | The node pool which the assigned resources refer to. | 
**Gpu** | Pointer to [**Resource**](Resource.md) | Number of GPUs assigned in the node pool. | [optional] 
**Cpu** | Pointer to [**Resource**](Resource.md) | Number of CPU Millicores assigned in the node pool. Supported only if the &#39;CPU Resources Quota&#39; feature flag is enabled. | [optional] 
**Memory** | Pointer to [**Resource**](Resource.md) | Amount of CPU Memory Mib assigned in the node pool. Supported only if the &#39;CPU Resources Quota&#39; feature flag is enabled. | [optional] 

## Methods

### NewNodePoolAssignedResourcesCreate

`func NewNodePoolAssignedResourcesCreate(nodePool NodePoolResources, ) *NodePoolAssignedResourcesCreate`

NewNodePoolAssignedResourcesCreate instantiates a new NodePoolAssignedResourcesCreate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolAssignedResourcesCreateWithDefaults

`func NewNodePoolAssignedResourcesCreateWithDefaults() *NodePoolAssignedResourcesCreate`

NewNodePoolAssignedResourcesCreateWithDefaults instantiates a new NodePoolAssignedResourcesCreate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePool

`func (o *NodePoolAssignedResourcesCreate) GetNodePool() NodePoolResources`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *NodePoolAssignedResourcesCreate) GetNodePoolOk() (*NodePoolResources, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *NodePoolAssignedResourcesCreate) SetNodePool(v NodePoolResources)`

SetNodePool sets NodePool field to given value.


### GetGpu

`func (o *NodePoolAssignedResourcesCreate) GetGpu() Resource`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *NodePoolAssignedResourcesCreate) GetGpuOk() (*Resource, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *NodePoolAssignedResourcesCreate) SetGpu(v Resource)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *NodePoolAssignedResourcesCreate) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetCpu

`func (o *NodePoolAssignedResourcesCreate) GetCpu() Resource`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *NodePoolAssignedResourcesCreate) GetCpuOk() (*Resource, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *NodePoolAssignedResourcesCreate) SetCpu(v Resource)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *NodePoolAssignedResourcesCreate) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetMemory

`func (o *NodePoolAssignedResourcesCreate) GetMemory() Resource`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *NodePoolAssignedResourcesCreate) GetMemoryOk() (*Resource, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *NodePoolAssignedResourcesCreate) SetMemory(v Resource)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *NodePoolAssignedResourcesCreate) HasMemory() bool`

HasMemory returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


