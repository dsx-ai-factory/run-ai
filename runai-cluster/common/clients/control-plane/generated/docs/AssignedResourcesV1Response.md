# AssignedResourcesV1Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **float32** |  | [optional] 
**Gpu** | [**AssignedResourcesV1ResponseGpu**](AssignedResourcesV1ResponseGpu.md) |  | 
**Cpu** | [**AssignedResourcesV1ResponseCpu**](AssignedResourcesV1ResponseCpu.md) |  | 
**Memory** | [**AssignedResourcesV1ResponseMemory**](AssignedResourcesV1ResponseMemory.md) |  | 

## Methods

### NewAssignedResourcesV1Response

`func NewAssignedResourcesV1Response(gpu AssignedResourcesV1ResponseGpu, cpu AssignedResourcesV1ResponseCpu, memory AssignedResourcesV1ResponseMemory, ) *AssignedResourcesV1Response`

NewAssignedResourcesV1Response instantiates a new AssignedResourcesV1Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssignedResourcesV1ResponseWithDefaults

`func NewAssignedResourcesV1ResponseWithDefaults() *AssignedResourcesV1Response`

NewAssignedResourcesV1ResponseWithDefaults instantiates a new AssignedResourcesV1Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AssignedResourcesV1Response) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AssignedResourcesV1Response) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AssignedResourcesV1Response) SetId(v float32)`

SetId sets Id field to given value.

### HasId

`func (o *AssignedResourcesV1Response) HasId() bool`

HasId returns a boolean if a field has been set.

### GetGpu

`func (o *AssignedResourcesV1Response) GetGpu() AssignedResourcesV1ResponseGpu`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *AssignedResourcesV1Response) GetGpuOk() (*AssignedResourcesV1ResponseGpu, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *AssignedResourcesV1Response) SetGpu(v AssignedResourcesV1ResponseGpu)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *AssignedResourcesV1Response) GetCpu() AssignedResourcesV1ResponseCpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *AssignedResourcesV1Response) GetCpuOk() (*AssignedResourcesV1ResponseCpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *AssignedResourcesV1Response) SetCpu(v AssignedResourcesV1ResponseCpu)`

SetCpu sets Cpu field to given value.


### GetMemory

`func (o *AssignedResourcesV1Response) GetMemory() AssignedResourcesV1ResponseMemory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *AssignedResourcesV1Response) GetMemoryOk() (*AssignedResourcesV1ResponseMemory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *AssignedResourcesV1Response) SetMemory(v AssignedResourcesV1ResponseMemory)`

SetMemory sets Memory field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


