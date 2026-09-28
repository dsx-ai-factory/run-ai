# DepartmentsMetricsV1ResponseDataInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**DepartmentMetricsV1Metadata**](DepartmentMetricsV1Metadata.md) |  | 
**Current** | [**DepartmentMetricsV1Current**](DepartmentMetricsV1Current.md) |  | 
**TimeRange** | Pointer to **interface{}** |  | [optional] 

## Methods

### NewDepartmentsMetricsV1ResponseDataInner

`func NewDepartmentsMetricsV1ResponseDataInner(metadata DepartmentMetricsV1Metadata, current DepartmentMetricsV1Current, ) *DepartmentsMetricsV1ResponseDataInner`

NewDepartmentsMetricsV1ResponseDataInner instantiates a new DepartmentsMetricsV1ResponseDataInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentsMetricsV1ResponseDataInnerWithDefaults

`func NewDepartmentsMetricsV1ResponseDataInnerWithDefaults() *DepartmentsMetricsV1ResponseDataInner`

NewDepartmentsMetricsV1ResponseDataInnerWithDefaults instantiates a new DepartmentsMetricsV1ResponseDataInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *DepartmentsMetricsV1ResponseDataInner) GetMetadata() DepartmentMetricsV1Metadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *DepartmentsMetricsV1ResponseDataInner) GetMetadataOk() (*DepartmentMetricsV1Metadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *DepartmentsMetricsV1ResponseDataInner) SetMetadata(v DepartmentMetricsV1Metadata)`

SetMetadata sets Metadata field to given value.


### GetCurrent

`func (o *DepartmentsMetricsV1ResponseDataInner) GetCurrent() DepartmentMetricsV1Current`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *DepartmentsMetricsV1ResponseDataInner) GetCurrentOk() (*DepartmentMetricsV1Current, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *DepartmentsMetricsV1ResponseDataInner) SetCurrent(v DepartmentMetricsV1Current)`

SetCurrent sets Current field to given value.


### GetTimeRange

`func (o *DepartmentsMetricsV1ResponseDataInner) GetTimeRange() interface{}`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *DepartmentsMetricsV1ResponseDataInner) GetTimeRangeOk() (*interface{}, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *DepartmentsMetricsV1ResponseDataInner) SetTimeRange(v interface{})`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *DepartmentsMetricsV1ResponseDataInner) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.

### SetTimeRangeNil

`func (o *DepartmentsMetricsV1ResponseDataInner) SetTimeRangeNil(b bool)`

 SetTimeRangeNil sets the value for TimeRange to be an explicit nil

### UnsetTimeRange
`func (o *DepartmentsMetricsV1ResponseDataInner) UnsetTimeRange()`

UnsetTimeRange ensures that no value is present for TimeRange, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


