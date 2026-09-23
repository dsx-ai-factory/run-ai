# DepartmentMetricsV1Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**DepartmentMetricsV1Metadata**](DepartmentMetricsV1Metadata.md) |  | 
**Current** | [**DepartmentMetricsV1Current**](DepartmentMetricsV1Current.md) |  | 
**TimeRange** | Pointer to [**TimeRangeV1**](TimeRangeV1.md) |  | [optional] 

## Methods

### NewDepartmentMetricsV1Response

`func NewDepartmentMetricsV1Response(metadata DepartmentMetricsV1Metadata, current DepartmentMetricsV1Current, ) *DepartmentMetricsV1Response`

NewDepartmentMetricsV1Response instantiates a new DepartmentMetricsV1Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentMetricsV1ResponseWithDefaults

`func NewDepartmentMetricsV1ResponseWithDefaults() *DepartmentMetricsV1Response`

NewDepartmentMetricsV1ResponseWithDefaults instantiates a new DepartmentMetricsV1Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *DepartmentMetricsV1Response) GetMetadata() DepartmentMetricsV1Metadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *DepartmentMetricsV1Response) GetMetadataOk() (*DepartmentMetricsV1Metadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *DepartmentMetricsV1Response) SetMetadata(v DepartmentMetricsV1Metadata)`

SetMetadata sets Metadata field to given value.


### GetCurrent

`func (o *DepartmentMetricsV1Response) GetCurrent() DepartmentMetricsV1Current`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *DepartmentMetricsV1Response) GetCurrentOk() (*DepartmentMetricsV1Current, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *DepartmentMetricsV1Response) SetCurrent(v DepartmentMetricsV1Current)`

SetCurrent sets Current field to given value.


### GetTimeRange

`func (o *DepartmentMetricsV1Response) GetTimeRange() TimeRangeV1`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *DepartmentMetricsV1Response) GetTimeRangeOk() (*TimeRangeV1, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *DepartmentMetricsV1Response) SetTimeRange(v TimeRangeV1)`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *DepartmentMetricsV1Response) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


