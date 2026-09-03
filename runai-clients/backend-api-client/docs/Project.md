# Project

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**ProjectMetadata**](ProjectMetadata.md) |  | 
**Current** | [**ProjectCurrent**](ProjectCurrent.md) |  | 
**TimeRange** | Pointer to [**TimeRange**](TimeRange.md) |  | [optional] 

## Methods

### NewProject

`func NewProject(metadata ProjectMetadata, current ProjectCurrent, ) *Project`

NewProject instantiates a new Project object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectWithDefaults

`func NewProjectWithDefaults() *Project`

NewProjectWithDefaults instantiates a new Project object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *Project) GetMetadata() ProjectMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *Project) GetMetadataOk() (*ProjectMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *Project) SetMetadata(v ProjectMetadata)`

SetMetadata sets Metadata field to given value.


### GetCurrent

`func (o *Project) GetCurrent() ProjectCurrent`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *Project) GetCurrentOk() (*ProjectCurrent, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *Project) SetCurrent(v ProjectCurrent)`

SetCurrent sets Current field to given value.


### GetTimeRange

`func (o *Project) GetTimeRange() TimeRange`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *Project) GetTimeRangeOk() (*TimeRange, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *Project) SetTimeRange(v TimeRange)`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *Project) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


