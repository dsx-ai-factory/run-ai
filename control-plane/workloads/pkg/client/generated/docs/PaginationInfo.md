# PaginationInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Limit** | Pointer to **float32** | the number of entries included in this bulk of entries | [optional] 
**Offset** | Pointer to **float32** | the first entry that was requested to be included in this bulk | [optional] 
**NextOffset** | Pointer to **float32** | the offset of the next bulk of entries | [optional] 

## Methods

### NewPaginationInfo

`func NewPaginationInfo() *PaginationInfo`

NewPaginationInfo instantiates a new PaginationInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPaginationInfoWithDefaults

`func NewPaginationInfoWithDefaults() *PaginationInfo`

NewPaginationInfoWithDefaults instantiates a new PaginationInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLimit

`func (o *PaginationInfo) GetLimit() float32`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *PaginationInfo) GetLimitOk() (*float32, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *PaginationInfo) SetLimit(v float32)`

SetLimit sets Limit field to given value.

### HasLimit

`func (o *PaginationInfo) HasLimit() bool`

HasLimit returns a boolean if a field has been set.

### GetOffset

`func (o *PaginationInfo) GetOffset() float32`

GetOffset returns the Offset field if non-nil, zero value otherwise.

### GetOffsetOk

`func (o *PaginationInfo) GetOffsetOk() (*float32, bool)`

GetOffsetOk returns a tuple with the Offset field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOffset

`func (o *PaginationInfo) SetOffset(v float32)`

SetOffset sets Offset field to given value.

### HasOffset

`func (o *PaginationInfo) HasOffset() bool`

HasOffset returns a boolean if a field has been set.

### GetNextOffset

`func (o *PaginationInfo) GetNextOffset() float32`

GetNextOffset returns the NextOffset field if non-nil, zero value otherwise.

### GetNextOffsetOk

`func (o *PaginationInfo) GetNextOffsetOk() (*float32, bool)`

GetNextOffsetOk returns a tuple with the NextOffset field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNextOffset

`func (o *PaginationInfo) SetNextOffset(v float32)`

SetNextOffset sets NextOffset field to given value.

### HasNextOffset

`func (o *PaginationInfo) HasNextOffset() bool`

HasNextOffset returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


