# NodePoolAssignedResourcesCreate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePool** | [**NodePoolAssignedResourcesCreateNodePool**](NodePoolAssignedResourcesCreateNodePool.md) |  | 
**Gpu** | Pointer to [**NodePoolAssignedResourcesCreateGpu**](NodePoolAssignedResourcesCreateGpu.md) |  | [optional] 
**Cpu** | Pointer to [**NodePoolAssignedResourcesCreateCpu**](NodePoolAssignedResourcesCreateCpu.md) |  | [optional] 
**Memory** | Pointer to [**NodePoolAssignedResourcesCreateMemory**](NodePoolAssignedResourcesCreateMemory.md) |  | [optional] 

## Methods

### NewNodePoolAssignedResourcesCreate

`func NewNodePoolAssignedResourcesCreate(nodePool NodePoolAssignedResourcesCreateNodePool, ) *NodePoolAssignedResourcesCreate`

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

`func (o *NodePoolAssignedResourcesCreate) GetNodePool() NodePoolAssignedResourcesCreateNodePool`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *NodePoolAssignedResourcesCreate) GetNodePoolOk() (*NodePoolAssignedResourcesCreateNodePool, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *NodePoolAssignedResourcesCreate) SetNodePool(v NodePoolAssignedResourcesCreateNodePool)`

SetNodePool sets NodePool field to given value.


### GetGpu

`func (o *NodePoolAssignedResourcesCreate) GetGpu() NodePoolAssignedResourcesCreateGpu`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *NodePoolAssignedResourcesCreate) GetGpuOk() (*NodePoolAssignedResourcesCreateGpu, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *NodePoolAssignedResourcesCreate) SetGpu(v NodePoolAssignedResourcesCreateGpu)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *NodePoolAssignedResourcesCreate) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetCpu

`func (o *NodePoolAssignedResourcesCreate) GetCpu() NodePoolAssignedResourcesCreateCpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *NodePoolAssignedResourcesCreate) GetCpuOk() (*NodePoolAssignedResourcesCreateCpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *NodePoolAssignedResourcesCreate) SetCpu(v NodePoolAssignedResourcesCreateCpu)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *NodePoolAssignedResourcesCreate) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetMemory

`func (o *NodePoolAssignedResourcesCreate) GetMemory() NodePoolAssignedResourcesCreateMemory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *NodePoolAssignedResourcesCreate) GetMemoryOk() (*NodePoolAssignedResourcesCreateMemory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *NodePoolAssignedResourcesCreate) SetMemory(v NodePoolAssignedResourcesCreateMemory)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *NodePoolAssignedResourcesCreate) HasMemory() bool`

HasMemory returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


