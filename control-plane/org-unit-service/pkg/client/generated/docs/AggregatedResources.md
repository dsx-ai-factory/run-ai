# AggregatedResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePool** | [**AggregatedResourcesNodePool**](AggregatedResourcesNodePool.md) |  | 
**Gpu** | Pointer to [**DeservedResource**](DeservedResource.md) |  | [optional] 
**Cpu** | Pointer to [**DeservedResource**](DeservedResource.md) |  | [optional] 
**Memory** | Pointer to [**AggregatedResourcesMemory**](AggregatedResourcesMemory.md) |  | [optional] 

## Methods

### NewAggregatedResources

`func NewAggregatedResources(nodePool AggregatedResourcesNodePool, ) *AggregatedResources`

NewAggregatedResources instantiates a new AggregatedResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAggregatedResourcesWithDefaults

`func NewAggregatedResourcesWithDefaults() *AggregatedResources`

NewAggregatedResourcesWithDefaults instantiates a new AggregatedResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePool

`func (o *AggregatedResources) GetNodePool() AggregatedResourcesNodePool`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *AggregatedResources) GetNodePoolOk() (*AggregatedResourcesNodePool, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *AggregatedResources) SetNodePool(v AggregatedResourcesNodePool)`

SetNodePool sets NodePool field to given value.


### GetGpu

`func (o *AggregatedResources) GetGpu() DeservedResource`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *AggregatedResources) GetGpuOk() (*DeservedResource, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *AggregatedResources) SetGpu(v DeservedResource)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *AggregatedResources) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetCpu

`func (o *AggregatedResources) GetCpu() DeservedResource`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *AggregatedResources) GetCpuOk() (*DeservedResource, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *AggregatedResources) SetCpu(v DeservedResource)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *AggregatedResources) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetMemory

`func (o *AggregatedResources) GetMemory() AggregatedResourcesMemory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *AggregatedResources) GetMemoryOk() (*AggregatedResourcesMemory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *AggregatedResources) SetMemory(v AggregatedResourcesMemory)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *AggregatedResources) HasMemory() bool`

HasMemory returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


