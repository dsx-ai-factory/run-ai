# Projects

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Data** | [**[]ProjectsDataInner**](ProjectsDataInner.md) |  | 
**TimeRange** | Pointer to [**TimeRange**](TimeRange.md) |  | [optional] 

## Methods

### NewProjects

`func NewProjects(data []ProjectsDataInner, ) *Projects`

NewProjects instantiates a new Projects object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectsWithDefaults

`func NewProjectsWithDefaults() *Projects`

NewProjectsWithDefaults instantiates a new Projects object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetData

`func (o *Projects) GetData() []ProjectsDataInner`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *Projects) GetDataOk() (*[]ProjectsDataInner, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *Projects) SetData(v []ProjectsDataInner)`

SetData sets Data field to given value.


### GetTimeRange

`func (o *Projects) GetTimeRange() TimeRange`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *Projects) GetTimeRangeOk() (*TimeRange, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *Projects) SetTimeRange(v TimeRange)`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *Projects) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


