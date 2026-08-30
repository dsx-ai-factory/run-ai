# ProjectGetRanking

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the project | 
**Id** | **string** | The id of the project | 
**ParentName** | **string** | The name of the parent department | 
**AllocationValue** | **NullableFloat64** | The allocation value of the project | 
**UtilizationValue** | **NullableFloat64** | The utilization value of the project | 
**MemoryUtilizationValue** | Pointer to **NullableFloat64** | The memory utilization value of the project | [optional] 
**GpuQuota** | Pointer to **NullableFloat64** | The GPU quota of the project | [optional] 
**RunningWorkloadsCount** | Pointer to **NullableInt32** | The number of currently running workloads in the project | [optional] 
**PendingWorkloadsCount** | Pointer to **NullableInt32** | The number of currently pending workloads in the project | [optional] 

## Methods

### NewProjectGetRanking

`func NewProjectGetRanking(name string, id string, parentName string, allocationValue NullableFloat64, utilizationValue NullableFloat64, ) *ProjectGetRanking`

NewProjectGetRanking instantiates a new ProjectGetRanking object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectGetRankingWithDefaults

`func NewProjectGetRankingWithDefaults() *ProjectGetRanking`

NewProjectGetRankingWithDefaults instantiates a new ProjectGetRanking object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ProjectGetRanking) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectGetRanking) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectGetRanking) SetName(v string)`

SetName sets Name field to given value.


### GetId

`func (o *ProjectGetRanking) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ProjectGetRanking) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ProjectGetRanking) SetId(v string)`

SetId sets Id field to given value.


### GetParentName

`func (o *ProjectGetRanking) GetParentName() string`

GetParentName returns the ParentName field if non-nil, zero value otherwise.

### GetParentNameOk

`func (o *ProjectGetRanking) GetParentNameOk() (*string, bool)`

GetParentNameOk returns a tuple with the ParentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentName

`func (o *ProjectGetRanking) SetParentName(v string)`

SetParentName sets ParentName field to given value.


### GetAllocationValue

`func (o *ProjectGetRanking) GetAllocationValue() float64`

GetAllocationValue returns the AllocationValue field if non-nil, zero value otherwise.

### GetAllocationValueOk

`func (o *ProjectGetRanking) GetAllocationValueOk() (*float64, bool)`

GetAllocationValueOk returns a tuple with the AllocationValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocationValue

`func (o *ProjectGetRanking) SetAllocationValue(v float64)`

SetAllocationValue sets AllocationValue field to given value.


### SetAllocationValueNil

`func (o *ProjectGetRanking) SetAllocationValueNil(b bool)`

 SetAllocationValueNil sets the value for AllocationValue to be an explicit nil

### UnsetAllocationValue
`func (o *ProjectGetRanking) UnsetAllocationValue()`

UnsetAllocationValue ensures that no value is present for AllocationValue, not even an explicit nil
### GetUtilizationValue

`func (o *ProjectGetRanking) GetUtilizationValue() float64`

GetUtilizationValue returns the UtilizationValue field if non-nil, zero value otherwise.

### GetUtilizationValueOk

`func (o *ProjectGetRanking) GetUtilizationValueOk() (*float64, bool)`

GetUtilizationValueOk returns a tuple with the UtilizationValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUtilizationValue

`func (o *ProjectGetRanking) SetUtilizationValue(v float64)`

SetUtilizationValue sets UtilizationValue field to given value.


### SetUtilizationValueNil

`func (o *ProjectGetRanking) SetUtilizationValueNil(b bool)`

 SetUtilizationValueNil sets the value for UtilizationValue to be an explicit nil

### UnsetUtilizationValue
`func (o *ProjectGetRanking) UnsetUtilizationValue()`

UnsetUtilizationValue ensures that no value is present for UtilizationValue, not even an explicit nil
### GetMemoryUtilizationValue

`func (o *ProjectGetRanking) GetMemoryUtilizationValue() float64`

GetMemoryUtilizationValue returns the MemoryUtilizationValue field if non-nil, zero value otherwise.

### GetMemoryUtilizationValueOk

`func (o *ProjectGetRanking) GetMemoryUtilizationValueOk() (*float64, bool)`

GetMemoryUtilizationValueOk returns a tuple with the MemoryUtilizationValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemoryUtilizationValue

`func (o *ProjectGetRanking) SetMemoryUtilizationValue(v float64)`

SetMemoryUtilizationValue sets MemoryUtilizationValue field to given value.

### HasMemoryUtilizationValue

`func (o *ProjectGetRanking) HasMemoryUtilizationValue() bool`

HasMemoryUtilizationValue returns a boolean if a field has been set.

### SetMemoryUtilizationValueNil

`func (o *ProjectGetRanking) SetMemoryUtilizationValueNil(b bool)`

 SetMemoryUtilizationValueNil sets the value for MemoryUtilizationValue to be an explicit nil

### UnsetMemoryUtilizationValue
`func (o *ProjectGetRanking) UnsetMemoryUtilizationValue()`

UnsetMemoryUtilizationValue ensures that no value is present for MemoryUtilizationValue, not even an explicit nil
### GetGpuQuota

`func (o *ProjectGetRanking) GetGpuQuota() float64`

GetGpuQuota returns the GpuQuota field if non-nil, zero value otherwise.

### GetGpuQuotaOk

`func (o *ProjectGetRanking) GetGpuQuotaOk() (*float64, bool)`

GetGpuQuotaOk returns a tuple with the GpuQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuQuota

`func (o *ProjectGetRanking) SetGpuQuota(v float64)`

SetGpuQuota sets GpuQuota field to given value.

### HasGpuQuota

`func (o *ProjectGetRanking) HasGpuQuota() bool`

HasGpuQuota returns a boolean if a field has been set.

### SetGpuQuotaNil

`func (o *ProjectGetRanking) SetGpuQuotaNil(b bool)`

 SetGpuQuotaNil sets the value for GpuQuota to be an explicit nil

### UnsetGpuQuota
`func (o *ProjectGetRanking) UnsetGpuQuota()`

UnsetGpuQuota ensures that no value is present for GpuQuota, not even an explicit nil
### GetRunningWorkloadsCount

`func (o *ProjectGetRanking) GetRunningWorkloadsCount() int32`

GetRunningWorkloadsCount returns the RunningWorkloadsCount field if non-nil, zero value otherwise.

### GetRunningWorkloadsCountOk

`func (o *ProjectGetRanking) GetRunningWorkloadsCountOk() (*int32, bool)`

GetRunningWorkloadsCountOk returns a tuple with the RunningWorkloadsCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunningWorkloadsCount

`func (o *ProjectGetRanking) SetRunningWorkloadsCount(v int32)`

SetRunningWorkloadsCount sets RunningWorkloadsCount field to given value.

### HasRunningWorkloadsCount

`func (o *ProjectGetRanking) HasRunningWorkloadsCount() bool`

HasRunningWorkloadsCount returns a boolean if a field has been set.

### SetRunningWorkloadsCountNil

`func (o *ProjectGetRanking) SetRunningWorkloadsCountNil(b bool)`

 SetRunningWorkloadsCountNil sets the value for RunningWorkloadsCount to be an explicit nil

### UnsetRunningWorkloadsCount
`func (o *ProjectGetRanking) UnsetRunningWorkloadsCount()`

UnsetRunningWorkloadsCount ensures that no value is present for RunningWorkloadsCount, not even an explicit nil
### GetPendingWorkloadsCount

`func (o *ProjectGetRanking) GetPendingWorkloadsCount() int32`

GetPendingWorkloadsCount returns the PendingWorkloadsCount field if non-nil, zero value otherwise.

### GetPendingWorkloadsCountOk

`func (o *ProjectGetRanking) GetPendingWorkloadsCountOk() (*int32, bool)`

GetPendingWorkloadsCountOk returns a tuple with the PendingWorkloadsCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPendingWorkloadsCount

`func (o *ProjectGetRanking) SetPendingWorkloadsCount(v int32)`

SetPendingWorkloadsCount sets PendingWorkloadsCount field to given value.

### HasPendingWorkloadsCount

`func (o *ProjectGetRanking) HasPendingWorkloadsCount() bool`

HasPendingWorkloadsCount returns a boolean if a field has been set.

### SetPendingWorkloadsCountNil

`func (o *ProjectGetRanking) SetPendingWorkloadsCountNil(b bool)`

 SetPendingWorkloadsCountNil sets the value for PendingWorkloadsCount to be an explicit nil

### UnsetPendingWorkloadsCount
`func (o *ProjectGetRanking) UnsetPendingWorkloadsCount()`

UnsetPendingWorkloadsCount ensures that no value is present for PendingWorkloadsCount, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


