# DepartmentsDataInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**DepartmentMetadata**](DepartmentMetadata.md) |  | 
**Current** | [**DepartmentCurrent**](DepartmentCurrent.md) |  | 
**TimeRange** | Pointer to **interface{}** |  | [optional] 

## Methods

### NewDepartmentsDataInner

`func NewDepartmentsDataInner(metadata DepartmentMetadata, current DepartmentCurrent, ) *DepartmentsDataInner`

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

`func (o *DepartmentsDataInner) GetMetadata() DepartmentMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *DepartmentsDataInner) GetMetadataOk() (*DepartmentMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *DepartmentsDataInner) SetMetadata(v DepartmentMetadata)`

SetMetadata sets Metadata field to given value.


### GetCurrent

`func (o *DepartmentsDataInner) GetCurrent() DepartmentCurrent`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *DepartmentsDataInner) GetCurrentOk() (*DepartmentCurrent, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *DepartmentsDataInner) SetCurrent(v DepartmentCurrent)`

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


