# DepartmentMetricsV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**DepartmentMetricsV1Metadata**](DepartmentMetricsV1Metadata.md) |  | 
**Current** | [**DepartmentMetricsV1Current**](DepartmentMetricsV1Current.md) |  | 
**TimeRange** | Pointer to [**TimeRangeV1**](TimeRangeV1.md) |  | [optional] 

## Methods

### NewDepartmentMetricsV1

`func NewDepartmentMetricsV1(metadata DepartmentMetricsV1Metadata, current DepartmentMetricsV1Current, ) *DepartmentMetricsV1`

NewDepartmentMetricsV1 instantiates a new DepartmentMetricsV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentMetricsV1WithDefaults

`func NewDepartmentMetricsV1WithDefaults() *DepartmentMetricsV1`

NewDepartmentMetricsV1WithDefaults instantiates a new DepartmentMetricsV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *DepartmentMetricsV1) GetMetadata() DepartmentMetricsV1Metadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *DepartmentMetricsV1) GetMetadataOk() (*DepartmentMetricsV1Metadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *DepartmentMetricsV1) SetMetadata(v DepartmentMetricsV1Metadata)`

SetMetadata sets Metadata field to given value.


### GetCurrent

`func (o *DepartmentMetricsV1) GetCurrent() DepartmentMetricsV1Current`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *DepartmentMetricsV1) GetCurrentOk() (*DepartmentMetricsV1Current, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *DepartmentMetricsV1) SetCurrent(v DepartmentMetricsV1Current)`

SetCurrent sets Current field to given value.


### GetTimeRange

`func (o *DepartmentMetricsV1) GetTimeRange() TimeRangeV1`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *DepartmentMetricsV1) GetTimeRangeOk() (*TimeRangeV1, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *DepartmentMetricsV1) SetTimeRange(v TimeRangeV1)`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *DepartmentMetricsV1) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


