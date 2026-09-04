# DepartmentsDataInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**Department1Metadata**](Department1Metadata.md) |  | 
**Current** | [**Department1Current**](Department1Current.md) |  | 
**TimeRange** | Pointer to **interface{}** |  | [optional] 

## Methods

### NewDepartmentsDataInner

`func NewDepartmentsDataInner(metadata Department1Metadata, current Department1Current, ) *DepartmentsDataInner`

NewDepartmentsDataInner instantiates a new DepartmentsDataInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentsDataInnerWithDefaults

`func NewDepartmentsDataInnerWithDefaults() *DepartmentsDataInner`

NewDepartmentsDataInnerWithDefaults instantiates a new DepartmentsDataInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *DepartmentsDataInner) GetMetadata() Department1Metadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *DepartmentsDataInner) GetMetadataOk() (*Department1Metadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *DepartmentsDataInner) SetMetadata(v Department1Metadata)`

SetMetadata sets Metadata field to given value.


### GetCurrent

`func (o *DepartmentsDataInner) GetCurrent() Department1Current`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *DepartmentsDataInner) GetCurrentOk() (*Department1Current, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *DepartmentsDataInner) SetCurrent(v Department1Current)`

SetCurrent sets Current field to given value.


### GetTimeRange

`func (o *DepartmentsDataInner) GetTimeRange() interface{}`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *DepartmentsDataInner) GetTimeRangeOk() (*interface{}, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *DepartmentsDataInner) SetTimeRange(v interface{})`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *DepartmentsDataInner) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.

### SetTimeRangeNil

`func (o *DepartmentsDataInner) SetTimeRangeNil(b bool)`

 SetTimeRangeNil sets the value for TimeRange to be an explicit nil

### UnsetTimeRange
`func (o *DepartmentsDataInner) UnsetTimeRange()`

UnsetTimeRange ensures that no value is present for TimeRange, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


