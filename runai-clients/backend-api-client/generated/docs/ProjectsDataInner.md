# ProjectsDataInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**ProjectMetadata**](ProjectMetadata.md) |  | 
**Current** | [**ProjectCurrent**](ProjectCurrent.md) |  | 
**TimeRange** | Pointer to **interface{}** |  | [optional] 

## Methods

### NewProjectsDataInner

`func NewProjectsDataInner(metadata ProjectMetadata, current ProjectCurrent, ) *ProjectsDataInner`

NewProjectsDataInner instantiates a new ProjectsDataInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectsDataInnerWithDefaults

`func NewProjectsDataInnerWithDefaults() *ProjectsDataInner`

NewProjectsDataInnerWithDefaults instantiates a new ProjectsDataInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *ProjectsDataInner) GetMetadata() ProjectMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *ProjectsDataInner) GetMetadataOk() (*ProjectMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *ProjectsDataInner) SetMetadata(v ProjectMetadata)`

SetMetadata sets Metadata field to given value.


### GetCurrent

`func (o *ProjectsDataInner) GetCurrent() ProjectCurrent`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *ProjectsDataInner) GetCurrentOk() (*ProjectCurrent, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *ProjectsDataInner) SetCurrent(v ProjectCurrent)`

SetCurrent sets Current field to given value.


### GetTimeRange

`func (o *ProjectsDataInner) GetTimeRange() interface{}`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *ProjectsDataInner) GetTimeRangeOk() (*interface{}, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *ProjectsDataInner) SetTimeRange(v interface{})`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *ProjectsDataInner) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.

### SetTimeRangeNil

`func (o *ProjectsDataInner) SetTimeRangeNil(b bool)`

 SetTimeRangeNil sets the value for TimeRange to be an explicit nil

### UnsetTimeRange
`func (o *ProjectsDataInner) UnsetTimeRange()`

UnsetTimeRange ensures that no value is present for TimeRange, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


