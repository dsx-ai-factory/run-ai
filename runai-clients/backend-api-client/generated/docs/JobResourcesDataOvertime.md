# JobResourcesDataOvertime

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Gpu** | [**JobResourceData**](JobResourceData.md) |  | 
**GpuMemory** | [**JobResourceData**](JobResourceData.md) |  | 
**Cpu** | [**JobResourceData**](JobResourceData.md) |  | 
**CpuMemory** | [**JobResourceData**](JobResourceData.md) |  | 
**Advanced** | Pointer to [**JobAdvancedData**](JobAdvancedData.md) |  | [optional] 
**Timestamp** | **time.Time** |  | 

## Methods

### NewJobResourcesDataOvertime

`func NewJobResourcesDataOvertime(gpu JobResourceData, gpuMemory JobResourceData, cpu JobResourceData, cpuMemory JobResourceData, timestamp time.Time, ) *JobResourcesDataOvertime`

NewJobResourcesDataOvertime instantiates a new JobResourcesDataOvertime object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJobResourcesDataOvertimeWithDefaults

`func NewJobResourcesDataOvertimeWithDefaults() *JobResourcesDataOvertime`

NewJobResourcesDataOvertimeWithDefaults instantiates a new JobResourcesDataOvertime object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpu

`func (o *JobResourcesDataOvertime) GetGpu() JobResourceData`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *JobResourcesDataOvertime) GetGpuOk() (*JobResourceData, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *JobResourcesDataOvertime) SetGpu(v JobResourceData)`

SetGpu sets Gpu field to given value.


### GetGpuMemory

`func (o *JobResourcesDataOvertime) GetGpuMemory() JobResourceData`

GetGpuMemory returns the GpuMemory field if non-nil, zero value otherwise.

### GetGpuMemoryOk

`func (o *JobResourcesDataOvertime) GetGpuMemoryOk() (*JobResourceData, bool)`

GetGpuMemoryOk returns a tuple with the GpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemory

`func (o *JobResourcesDataOvertime) SetGpuMemory(v JobResourceData)`

SetGpuMemory sets GpuMemory field to given value.


### GetCpu

`func (o *JobResourcesDataOvertime) GetCpu() JobResourceData`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *JobResourcesDataOvertime) GetCpuOk() (*JobResourceData, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *JobResourcesDataOvertime) SetCpu(v JobResourceData)`

SetCpu sets Cpu field to given value.


### GetCpuMemory

`func (o *JobResourcesDataOvertime) GetCpuMemory() JobResourceData`

GetCpuMemory returns the CpuMemory field if non-nil, zero value otherwise.

### GetCpuMemoryOk

`func (o *JobResourcesDataOvertime) GetCpuMemoryOk() (*JobResourceData, bool)`

GetCpuMemoryOk returns a tuple with the CpuMemory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuMemory

`func (o *JobResourcesDataOvertime) SetCpuMemory(v JobResourceData)`

SetCpuMemory sets CpuMemory field to given value.


### GetAdvanced

`func (o *JobResourcesDataOvertime) GetAdvanced() JobAdvancedData`

GetAdvanced returns the Advanced field if non-nil, zero value otherwise.

### GetAdvancedOk

`func (o *JobResourcesDataOvertime) GetAdvancedOk() (*JobAdvancedData, bool)`

GetAdvancedOk returns a tuple with the Advanced field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdvanced

`func (o *JobResourcesDataOvertime) SetAdvanced(v JobAdvancedData)`

SetAdvanced sets Advanced field to given value.

### HasAdvanced

`func (o *JobResourcesDataOvertime) HasAdvanced() bool`

HasAdvanced returns a boolean if a field has been set.

### GetTimestamp

`func (o *JobResourcesDataOvertime) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *JobResourcesDataOvertime) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *JobResourcesDataOvertime) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


