# ProjectsMetricsV1ResponseDataInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**ProjectMetricsV1Metadata**](ProjectMetricsV1Metadata.md) |  | 
**Current** | [**ProjectMetricsV1Current**](ProjectMetricsV1Current.md) |  | 
**TimeRange** | Pointer to **interface{}** |  | [optional] 

## Methods

### NewProjectsMetricsV1ResponseDataInner

`func NewProjectsMetricsV1ResponseDataInner(metadata ProjectMetricsV1Metadata, current ProjectMetricsV1Current, ) *ProjectsMetricsV1ResponseDataInner`

NewProjectsMetricsV1ResponseDataInner instantiates a new ProjectsMetricsV1ResponseDataInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectsMetricsV1ResponseDataInnerWithDefaults

`func NewProjectsMetricsV1ResponseDataInnerWithDefaults() *ProjectsMetricsV1ResponseDataInner`

NewProjectsMetricsV1ResponseDataInnerWithDefaults instantiates a new ProjectsMetricsV1ResponseDataInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *ProjectsMetricsV1ResponseDataInner) GetMetadata() ProjectMetricsV1Metadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *ProjectsMetricsV1ResponseDataInner) GetMetadataOk() (*ProjectMetricsV1Metadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *ProjectsMetricsV1ResponseDataInner) SetMetadata(v ProjectMetricsV1Metadata)`

SetMetadata sets Metadata field to given value.


### GetCurrent

`func (o *ProjectsMetricsV1ResponseDataInner) GetCurrent() ProjectMetricsV1Current`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *ProjectsMetricsV1ResponseDataInner) GetCurrentOk() (*ProjectMetricsV1Current, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *ProjectsMetricsV1ResponseDataInner) SetCurrent(v ProjectMetricsV1Current)`

SetCurrent sets Current field to given value.


### GetTimeRange

`func (o *ProjectsMetricsV1ResponseDataInner) GetTimeRange() interface{}`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *ProjectsMetricsV1ResponseDataInner) GetTimeRangeOk() (*interface{}, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *ProjectsMetricsV1ResponseDataInner) SetTimeRange(v interface{})`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *ProjectsMetricsV1ResponseDataInner) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.

### SetTimeRangeNil

`func (o *ProjectsMetricsV1ResponseDataInner) SetTimeRangeNil(b bool)`

 SetTimeRangeNil sets the value for TimeRange to be an explicit nil

### UnsetTimeRange
`func (o *ProjectsMetricsV1ResponseDataInner) UnsetTimeRange()`

UnsetTimeRange ensures that no value is present for TimeRange, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


