# ResourcesDataOvertimeV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NumberOfPendingWorkloads** | **int32** |  | 
**Gpu** | [**MetricsResourceDataV1**](MetricsResourceDataV1.md) |  | 
**Cpu** | [**MetricsResourceDataV1**](MetricsResourceDataV1.md) |  | 
**Memory** | [**MetricsResourceDataV1**](MetricsResourceDataV1.md) |  | 
**Timestamp** | **time.Time** |  | 

## Methods

### NewResourcesDataOvertimeV1

`func NewResourcesDataOvertimeV1(numberOfPendingWorkloads int32, gpu MetricsResourceDataV1, cpu MetricsResourceDataV1, memory MetricsResourceDataV1, timestamp time.Time, ) *ResourcesDataOvertimeV1`

NewResourcesDataOvertimeV1 instantiates a new ResourcesDataOvertimeV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesDataOvertimeV1WithDefaults

`func NewResourcesDataOvertimeV1WithDefaults() *ResourcesDataOvertimeV1`

NewResourcesDataOvertimeV1WithDefaults instantiates a new ResourcesDataOvertimeV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumberOfPendingWorkloads

`func (o *ResourcesDataOvertimeV1) GetNumberOfPendingWorkloads() int32`

GetNumberOfPendingWorkloads returns the NumberOfPendingWorkloads field if non-nil, zero value otherwise.

### GetNumberOfPendingWorkloadsOk

`func (o *ResourcesDataOvertimeV1) GetNumberOfPendingWorkloadsOk() (*int32, bool)`

GetNumberOfPendingWorkloadsOk returns a tuple with the NumberOfPendingWorkloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumberOfPendingWorkloads

`func (o *ResourcesDataOvertimeV1) SetNumberOfPendingWorkloads(v int32)`

SetNumberOfPendingWorkloads sets NumberOfPendingWorkloads field to given value.


### GetGpu

`func (o *ResourcesDataOvertimeV1) GetGpu() MetricsResourceDataV1`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *ResourcesDataOvertimeV1) GetGpuOk() (*MetricsResourceDataV1, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *ResourcesDataOvertimeV1) SetGpu(v MetricsResourceDataV1)`

SetGpu sets Gpu field to given value.


### GetCpu

`func (o *ResourcesDataOvertimeV1) GetCpu() MetricsResourceDataV1`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *ResourcesDataOvertimeV1) GetCpuOk() (*MetricsResourceDataV1, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *ResourcesDataOvertimeV1) SetCpu(v MetricsResourceDataV1)`

SetCpu sets Cpu field to given value.


### GetMemory

`func (o *ResourcesDataOvertimeV1) GetMemory() MetricsResourceDataV1`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *ResourcesDataOvertimeV1) GetMemoryOk() (*MetricsResourceDataV1, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *ResourcesDataOvertimeV1) SetMemory(v MetricsResourceDataV1)`

SetMemory sets Memory field to given value.


### GetTimestamp

`func (o *ResourcesDataOvertimeV1) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *ResourcesDataOvertimeV1) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *ResourcesDataOvertimeV1) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


