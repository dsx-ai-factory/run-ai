# OvertimeRangeData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AverageGpuAllocation** | Pointer to **float64** | The average gpu devices allocation of the project/department. | [optional] 
**AverageGpuUtilization** | Pointer to **float64** | The average gpu utilization percentage of the project/department. | [optional] 
**AverageGpuMemoryUtilization** | Pointer to **NullableFloat64** | The average gpu memory utilization percentage of the project/department. Only available for cluster versions &gt;&#x3D; 2.22. | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewOvertimeRangeData

`func NewOvertimeRangeData() *OvertimeRangeData`

NewOvertimeRangeData instantiates a new OvertimeRangeData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOvertimeRangeDataWithDefaults

`func NewOvertimeRangeDataWithDefaults() *OvertimeRangeData`

NewOvertimeRangeDataWithDefaults instantiates a new OvertimeRangeData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAverageGpuAllocation

`func (o *OvertimeRangeData) GetAverageGpuAllocation() float64`

GetAverageGpuAllocation returns the AverageGpuAllocation field if non-nil, zero value otherwise.

### GetAverageGpuAllocationOk

`func (o *OvertimeRangeData) GetAverageGpuAllocationOk() (*float64, bool)`

GetAverageGpuAllocationOk returns a tuple with the AverageGpuAllocation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAverageGpuAllocation

`func (o *OvertimeRangeData) SetAverageGpuAllocation(v float64)`

SetAverageGpuAllocation sets AverageGpuAllocation field to given value.

### HasAverageGpuAllocation

`func (o *OvertimeRangeData) HasAverageGpuAllocation() bool`

HasAverageGpuAllocation returns a boolean if a field has been set.

### GetAverageGpuUtilization

`func (o *OvertimeRangeData) GetAverageGpuUtilization() float64`

GetAverageGpuUtilization returns the AverageGpuUtilization field if non-nil, zero value otherwise.

### GetAverageGpuUtilizationOk

`func (o *OvertimeRangeData) GetAverageGpuUtilizationOk() (*float64, bool)`

GetAverageGpuUtilizationOk returns a tuple with the AverageGpuUtilization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAverageGpuUtilization

`func (o *OvertimeRangeData) SetAverageGpuUtilization(v float64)`

SetAverageGpuUtilization sets AverageGpuUtilization field to given value.

### HasAverageGpuUtilization

`func (o *OvertimeRangeData) HasAverageGpuUtilization() bool`

HasAverageGpuUtilization returns a boolean if a field has been set.

### GetAverageGpuMemoryUtilization

`func (o *OvertimeRangeData) GetAverageGpuMemoryUtilization() float64`

GetAverageGpuMemoryUtilization returns the AverageGpuMemoryUtilization field if non-nil, zero value otherwise.

### GetAverageGpuMemoryUtilizationOk

`func (o *OvertimeRangeData) GetAverageGpuMemoryUtilizationOk() (*float64, bool)`

GetAverageGpuMemoryUtilizationOk returns a tuple with the AverageGpuMemoryUtilization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAverageGpuMemoryUtilization

`func (o *OvertimeRangeData) SetAverageGpuMemoryUtilization(v float64)`

SetAverageGpuMemoryUtilization sets AverageGpuMemoryUtilization field to given value.

### HasAverageGpuMemoryUtilization

`func (o *OvertimeRangeData) HasAverageGpuMemoryUtilization() bool`

HasAverageGpuMemoryUtilization returns a boolean if a field has been set.

### SetAverageGpuMemoryUtilizationNil

`func (o *OvertimeRangeData) SetAverageGpuMemoryUtilizationNil(b bool)`

 SetAverageGpuMemoryUtilizationNil sets the value for AverageGpuMemoryUtilization to be an explicit nil

### UnsetAverageGpuMemoryUtilization
`func (o *OvertimeRangeData) UnsetAverageGpuMemoryUtilization()`

UnsetAverageGpuMemoryUtilization ensures that no value is present for AverageGpuMemoryUtilization, not even an explicit nil
### GetUpdatedAt

`func (o *OvertimeRangeData) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *OvertimeRangeData) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *OvertimeRangeData) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *OvertimeRangeData) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


