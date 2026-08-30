# AssignedResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **float32** |  | [optional] 
**Gpu** | Pointer to [**Resource**](Resource.md) | GPU number assigned | [optional] 
**Cpu** | Pointer to [**Resource**](Resource.md) | CPU Millicores assigned. Supported only if &#39;CPU Resources Quota&#39; feature flag is enabled. | [optional] 
**Memory** | Pointer to [**Resource**](Resource.md) | CPU Memory Mib assigned. Supported only if &#39;CPU Resources Quota&#39; feature flag is enabled. | [optional] 

## Methods

### NewAssignedResources

`func NewAssignedResources() *AssignedResources`

NewAssignedResources instantiates a new AssignedResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssignedResourcesWithDefaults

`func NewAssignedResourcesWithDefaults() *AssignedResources`

NewAssignedResourcesWithDefaults instantiates a new AssignedResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AssignedResources) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AssignedResources) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AssignedResources) SetId(v float32)`

SetId sets Id field to given value.

### HasId

`func (o *AssignedResources) HasId() bool`

HasId returns a boolean if a field has been set.

### GetGpu

`func (o *AssignedResources) GetGpu() Resource`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *AssignedResources) GetGpuOk() (*Resource, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *AssignedResources) SetGpu(v Resource)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *AssignedResources) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetCpu

`func (o *AssignedResources) GetCpu() Resource`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *AssignedResources) GetCpuOk() (*Resource, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *AssignedResources) SetCpu(v Resource)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *AssignedResources) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetMemory

`func (o *AssignedResources) GetMemory() Resource`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *AssignedResources) GetMemoryOk() (*Resource, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *AssignedResources) SetMemory(v Resource)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *AssignedResources) HasMemory() bool`

HasMemory returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


