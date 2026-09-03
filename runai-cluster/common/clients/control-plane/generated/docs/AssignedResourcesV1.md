# AssignedResourcesV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **float32** |  | [optional] 
**Gpu** | Pointer to [**AssignedResourcesV1Gpu**](AssignedResourcesV1Gpu.md) |  | [optional] 
**Cpu** | Pointer to [**AssignedResourcesV1Cpu**](AssignedResourcesV1Cpu.md) |  | [optional] 
**Memory** | Pointer to [**AssignedResourcesV1Memory**](AssignedResourcesV1Memory.md) |  | [optional] 

## Methods

### NewAssignedResourcesV1

`func NewAssignedResourcesV1() *AssignedResourcesV1`

NewAssignedResourcesV1 instantiates a new AssignedResourcesV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssignedResourcesV1WithDefaults

`func NewAssignedResourcesV1WithDefaults() *AssignedResourcesV1`

NewAssignedResourcesV1WithDefaults instantiates a new AssignedResourcesV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AssignedResourcesV1) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AssignedResourcesV1) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AssignedResourcesV1) SetId(v float32)`

SetId sets Id field to given value.

### HasId

`func (o *AssignedResourcesV1) HasId() bool`

HasId returns a boolean if a field has been set.

### GetGpu

`func (o *AssignedResourcesV1) GetGpu() AssignedResourcesV1Gpu`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *AssignedResourcesV1) GetGpuOk() (*AssignedResourcesV1Gpu, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *AssignedResourcesV1) SetGpu(v AssignedResourcesV1Gpu)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *AssignedResourcesV1) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetCpu

`func (o *AssignedResourcesV1) GetCpu() AssignedResourcesV1Cpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *AssignedResourcesV1) GetCpuOk() (*AssignedResourcesV1Cpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *AssignedResourcesV1) SetCpu(v AssignedResourcesV1Cpu)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *AssignedResourcesV1) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetMemory

`func (o *AssignedResourcesV1) GetMemory() AssignedResourcesV1Memory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *AssignedResourcesV1) GetMemoryOk() (*AssignedResourcesV1Memory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *AssignedResourcesV1) SetMemory(v AssignedResourcesV1Memory)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *AssignedResourcesV1) HasMemory() bool`

HasMemory returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


