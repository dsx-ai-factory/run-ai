# Department

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**Department1Metadata**](Department1Metadata.md) |  | 
**Current** | [**Department1Current**](Department1Current.md) |  | 
**TimeRange** | Pointer to [**TimeRange**](TimeRange.md) |  | [optional] 

## Methods

### NewDepartment

`func NewDepartment(metadata Department1Metadata, current Department1Current, ) *Department`

NewDepartment instantiates a new Department object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentWithDefaults

`func NewDepartmentWithDefaults() *Department`

NewDepartmentWithDefaults instantiates a new Department object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *Department) GetMetadata() Department1Metadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *Department) GetMetadataOk() (*Department1Metadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *Department) SetMetadata(v Department1Metadata)`

SetMetadata sets Metadata field to given value.


### GetCurrent

`func (o *Department) GetCurrent() Department1Current`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *Department) GetCurrentOk() (*Department1Current, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *Department) SetCurrent(v Department1Current)`

SetCurrent sets Current field to given value.


### GetTimeRange

`func (o *Department) GetTimeRange() TimeRange`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *Department) GetTimeRangeOk() (*TimeRange, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *Department) SetTimeRange(v TimeRange)`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *Department) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


