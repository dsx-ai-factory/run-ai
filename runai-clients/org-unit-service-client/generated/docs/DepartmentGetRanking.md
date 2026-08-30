# DepartmentGetRanking

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the department | 
**Id** | **string** | The id of the department | 
**AllocationValue** | **NullableFloat64** | The metric value for the department. For allocation/utilization metrics this is the average GPU count; for gpuAllocationRatio it is a ratio (0–1+, e.g. 0.75 means 75% allocated). | 
**UtilizationValue** | **NullableFloat64** | The utilization value of the department | 
**MemoryUtilizationValue** | Pointer to **NullableFloat64** | The memory utilization value of the department | [optional] 
**GpuQuota** | Pointer to **NullableFloat64** | The GPU quota of the department | [optional] 
**RunningWorkloadsCount** | Pointer to **NullableInt32** | The number of currently running workloads in the department (aggregated from all projects) | [optional] 
**PendingWorkloadsCount** | Pointer to **NullableInt32** | The number of currently pending workloads in the department (aggregated from all projects) | [optional] 

## Methods

### NewDepartmentGetRanking

`func NewDepartmentGetRanking(name string, id string, allocationValue NullableFloat64, utilizationValue NullableFloat64, ) *DepartmentGetRanking`

NewDepartmentGetRanking instantiates a new DepartmentGetRanking object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentGetRankingWithDefaults

`func NewDepartmentGetRankingWithDefaults() *DepartmentGetRanking`

NewDepartmentGetRankingWithDefaults instantiates a new DepartmentGetRanking object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *DepartmentGetRanking) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentGetRanking) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentGetRanking) SetName(v string)`

SetName sets Name field to given value.


### GetId

`func (o *DepartmentGetRanking) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DepartmentGetRanking) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DepartmentGetRanking) SetId(v string)`

SetId sets Id field to given value.


### GetAllocationValue

`func (o *DepartmentGetRanking) GetAllocationValue() float64`

GetAllocationValue returns the AllocationValue field if non-nil, zero value otherwise.

### GetAllocationValueOk

`func (o *DepartmentGetRanking) GetAllocationValueOk() (*float64, bool)`

GetAllocationValueOk returns a tuple with the AllocationValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocationValue

`func (o *DepartmentGetRanking) SetAllocationValue(v float64)`

SetAllocationValue sets AllocationValue field to given value.


### SetAllocationValueNil

`func (o *DepartmentGetRanking) SetAllocationValueNil(b bool)`

 SetAllocationValueNil sets the value for AllocationValue to be an explicit nil

### UnsetAllocationValue
`func (o *DepartmentGetRanking) UnsetAllocationValue()`

UnsetAllocationValue ensures that no value is present for AllocationValue, not even an explicit nil
### GetUtilizationValue

`func (o *DepartmentGetRanking) GetUtilizationValue() float64`

GetUtilizationValue returns the UtilizationValue field if non-nil, zero value otherwise.

### GetUtilizationValueOk

`func (o *DepartmentGetRanking) GetUtilizationValueOk() (*float64, bool)`

GetUtilizationValueOk returns a tuple with the UtilizationValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUtilizationValue

`func (o *DepartmentGetRanking) SetUtilizationValue(v float64)`

SetUtilizationValue sets UtilizationValue field to given value.


### SetUtilizationValueNil

`func (o *DepartmentGetRanking) SetUtilizationValueNil(b bool)`

 SetUtilizationValueNil sets the value for UtilizationValue to be an explicit nil

### UnsetUtilizationValue
`func (o *DepartmentGetRanking) UnsetUtilizationValue()`

UnsetUtilizationValue ensures that no value is present for UtilizationValue, not even an explicit nil
### GetMemoryUtilizationValue

`func (o *DepartmentGetRanking) GetMemoryUtilizationValue() float64`

GetMemoryUtilizationValue returns the MemoryUtilizationValue field if non-nil, zero value otherwise.

### GetMemoryUtilizationValueOk

`func (o *DepartmentGetRanking) GetMemoryUtilizationValueOk() (*float64, bool)`

GetMemoryUtilizationValueOk returns a tuple with the MemoryUtilizationValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemoryUtilizationValue

`func (o *DepartmentGetRanking) SetMemoryUtilizationValue(v float64)`

SetMemoryUtilizationValue sets MemoryUtilizationValue field to given value.

### HasMemoryUtilizationValue

`func (o *DepartmentGetRanking) HasMemoryUtilizationValue() bool`

HasMemoryUtilizationValue returns a boolean if a field has been set.

### SetMemoryUtilizationValueNil

`func (o *DepartmentGetRanking) SetMemoryUtilizationValueNil(b bool)`

 SetMemoryUtilizationValueNil sets the value for MemoryUtilizationValue to be an explicit nil

### UnsetMemoryUtilizationValue
`func (o *DepartmentGetRanking) UnsetMemoryUtilizationValue()`

UnsetMemoryUtilizationValue ensures that no value is present for MemoryUtilizationValue, not even an explicit nil
### GetGpuQuota

`func (o *DepartmentGetRanking) GetGpuQuota() float64`

GetGpuQuota returns the GpuQuota field if non-nil, zero value otherwise.

### GetGpuQuotaOk

`func (o *DepartmentGetRanking) GetGpuQuotaOk() (*float64, bool)`

GetGpuQuotaOk returns a tuple with the GpuQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuQuota

`func (o *DepartmentGetRanking) SetGpuQuota(v float64)`

SetGpuQuota sets GpuQuota field to given value.

### HasGpuQuota

`func (o *DepartmentGetRanking) HasGpuQuota() bool`

HasGpuQuota returns a boolean if a field has been set.

### SetGpuQuotaNil

`func (o *DepartmentGetRanking) SetGpuQuotaNil(b bool)`

 SetGpuQuotaNil sets the value for GpuQuota to be an explicit nil

### UnsetGpuQuota
`func (o *DepartmentGetRanking) UnsetGpuQuota()`

UnsetGpuQuota ensures that no value is present for GpuQuota, not even an explicit nil
### GetRunningWorkloadsCount

`func (o *DepartmentGetRanking) GetRunningWorkloadsCount() int32`

GetRunningWorkloadsCount returns the RunningWorkloadsCount field if non-nil, zero value otherwise.

### GetRunningWorkloadsCountOk

`func (o *DepartmentGetRanking) GetRunningWorkloadsCountOk() (*int32, bool)`

GetRunningWorkloadsCountOk returns a tuple with the RunningWorkloadsCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunningWorkloadsCount

`func (o *DepartmentGetRanking) SetRunningWorkloadsCount(v int32)`

SetRunningWorkloadsCount sets RunningWorkloadsCount field to given value.

### HasRunningWorkloadsCount

`func (o *DepartmentGetRanking) HasRunningWorkloadsCount() bool`

HasRunningWorkloadsCount returns a boolean if a field has been set.

### SetRunningWorkloadsCountNil

`func (o *DepartmentGetRanking) SetRunningWorkloadsCountNil(b bool)`

 SetRunningWorkloadsCountNil sets the value for RunningWorkloadsCount to be an explicit nil

### UnsetRunningWorkloadsCount
`func (o *DepartmentGetRanking) UnsetRunningWorkloadsCount()`

UnsetRunningWorkloadsCount ensures that no value is present for RunningWorkloadsCount, not even an explicit nil
### GetPendingWorkloadsCount

`func (o *DepartmentGetRanking) GetPendingWorkloadsCount() int32`

GetPendingWorkloadsCount returns the PendingWorkloadsCount field if non-nil, zero value otherwise.

### GetPendingWorkloadsCountOk

`func (o *DepartmentGetRanking) GetPendingWorkloadsCountOk() (*int32, bool)`

GetPendingWorkloadsCountOk returns a tuple with the PendingWorkloadsCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPendingWorkloadsCount

`func (o *DepartmentGetRanking) SetPendingWorkloadsCount(v int32)`

SetPendingWorkloadsCount sets PendingWorkloadsCount field to given value.

### HasPendingWorkloadsCount

`func (o *DepartmentGetRanking) HasPendingWorkloadsCount() bool`

HasPendingWorkloadsCount returns a boolean if a field has been set.

### SetPendingWorkloadsCountNil

`func (o *DepartmentGetRanking) SetPendingWorkloadsCountNil(b bool)`

 SetPendingWorkloadsCountNil sets the value for PendingWorkloadsCount to be an explicit nil

### UnsetPendingWorkloadsCount
`func (o *DepartmentGetRanking) UnsetPendingWorkloadsCount()`

UnsetPendingWorkloadsCount ensures that no value is present for PendingWorkloadsCount, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


