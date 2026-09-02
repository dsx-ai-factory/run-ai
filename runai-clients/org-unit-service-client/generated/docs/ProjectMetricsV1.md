# ProjectMetricsV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**ProjectMetricsV1Metadata**](ProjectMetricsV1Metadata.md) |  | 
**Current** | [**ProjectMetricsV1Current**](ProjectMetricsV1Current.md) |  | 
**TimeRange** | Pointer to [**TimeRangeV1**](TimeRangeV1.md) |  | [optional] 

## Methods

### NewProjectMetricsV1

`func NewProjectMetricsV1(metadata ProjectMetricsV1Metadata, current ProjectMetricsV1Current, ) *ProjectMetricsV1`

NewProjectMetricsV1 instantiates a new ProjectMetricsV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectMetricsV1WithDefaults

`func NewProjectMetricsV1WithDefaults() *ProjectMetricsV1`

NewProjectMetricsV1WithDefaults instantiates a new ProjectMetricsV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *ProjectMetricsV1) GetMetadata() ProjectMetricsV1Metadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *ProjectMetricsV1) GetMetadataOk() (*ProjectMetricsV1Metadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *ProjectMetricsV1) SetMetadata(v ProjectMetricsV1Metadata)`

SetMetadata sets Metadata field to given value.


### GetCurrent

`func (o *ProjectMetricsV1) GetCurrent() ProjectMetricsV1Current`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *ProjectMetricsV1) GetCurrentOk() (*ProjectMetricsV1Current, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *ProjectMetricsV1) SetCurrent(v ProjectMetricsV1Current)`

SetCurrent sets Current field to given value.


### GetTimeRange

`func (o *ProjectMetricsV1) GetTimeRange() TimeRangeV1`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *ProjectMetricsV1) GetTimeRangeOk() (*TimeRangeV1, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *ProjectMetricsV1) SetTimeRange(v TimeRangeV1)`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *ProjectMetricsV1) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


