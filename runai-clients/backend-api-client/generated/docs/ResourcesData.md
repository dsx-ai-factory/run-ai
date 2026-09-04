# ResourcesData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NumberOfPendingWorkloads** | **int32** |  | 
**Gpu** | [**ResourceData**](ResourceData.md) |  | 
**Cpu** | [**ResourceData**](ResourceData.md) |  | 
**Memory** | [**ResourceData**](ResourceData.md) |  | 

## Methods

### NewResourcesData

`func NewResourcesData(numberOfPendingWorkloads int32, gpu ResourceData, cpu ResourceData, memory ResourceData, ) *ResourcesData`

NewResourcesData instantiates a new ResourcesData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesDataWithDefaults

`func NewResourcesDataWithDefaults() *ResourcesData`

NewResourcesDataWithDefaults instantiates a new ResourcesData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumberOfPendingWorkloads

`func (o *ResourcesData) GetNumberOfPendingWorkloads() int32`

GetNumberOfPendingWorkloads returns the NumberOfPendingWorkloads field if non-nil, zero value otherwise.

### GetNumberOfPendingWorkloadsOk

`func (o *ResourcesData) GetNumberOfPendingWorkloadsOk() (*int32, bool)`

GetNumberOfPendingWorkloadsOk returns a tuple with the NumberOfPendingWorkloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumberOfPendingWorkloads

`func (o *ResourcesData) SetNumberOfPendingWorkloads(v int32)`

SetNumberOfPendingWorkloads sets NumberOfPendingWorkloads field to given value.


### GetGpu

`func (o *ResourcesData) GetGpu() ResourceData`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *ResourcesData) GetGpuOk() (*ResourceData, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *ResourcesData) SetGpu(v ResourceData)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *ResourcesData) GetCpu() ResourceData`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *ResourcesData) GetCpuOk() (*ResourceData, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *ResourcesData) SetCpu(v ResourceData)`

SetCpu sets Cpu field to given value.


### GetMemory

`func (o *ResourcesData) GetMemory() ResourceData`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *ResourcesData) GetMemoryOk() (*ResourceData, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *ResourcesData) SetMemory(v ResourceData)`

SetMemory sets Memory field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


