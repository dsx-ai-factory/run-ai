# QueueResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Gpu** | **float64** | The number of GPUs (allocated/requested/...). | 
**Cpu** | **float64** | CPU in Cores (allocated/requested/...). | 
**Memory** | **float64** | Memory in Bytes (allocated/requested/...). | 

## Methods

### NewQueueResources

`func NewQueueResources(gpu float64, cpu float64, memory float64, ) *QueueResources`

NewQueueResources instantiates a new QueueResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewQueueResourcesWithDefaults

`func NewQueueResourcesWithDefaults() *QueueResources`

NewQueueResourcesWithDefaults instantiates a new QueueResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpu

`func (o *QueueResources) GetGpu() float64`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *QueueResources) GetGpuOk() (*float64, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *QueueResources) SetGpu(v float64)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *QueueResources) GetCpu() float64`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *QueueResources) GetCpuOk() (*float64, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *QueueResources) SetCpu(v float64)`

SetCpu sets Cpu field to given value.


### GetMemory

`func (o *QueueResources) GetMemory() float64`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *QueueResources) GetMemoryOk() (*float64, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *QueueResources) SetMemory(v float64)`

SetMemory sets Memory field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


