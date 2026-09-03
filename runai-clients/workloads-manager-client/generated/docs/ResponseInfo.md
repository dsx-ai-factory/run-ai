# ResponseInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LowestUpdate** | Pointer to **int64** | the value of an update counter of a data in the system. | [optional] 
**HighestUpdate** | Pointer to **int64** | the value of an update counter of a data in the system. | [optional] 

## Methods

### NewResponseInfo

`func NewResponseInfo() *ResponseInfo`

NewResponseInfo instantiates a new ResponseInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResponseInfoWithDefaults

`func NewResponseInfoWithDefaults() *ResponseInfo`

NewResponseInfoWithDefaults instantiates a new ResponseInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLowestUpdate

`func (o *ResponseInfo) GetLowestUpdate() int64`

GetLowestUpdate returns the LowestUpdate field if non-nil, zero value otherwise.

### GetLowestUpdateOk

`func (o *ResponseInfo) GetLowestUpdateOk() (*int64, bool)`

GetLowestUpdateOk returns a tuple with the LowestUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLowestUpdate

`func (o *ResponseInfo) SetLowestUpdate(v int64)`

SetLowestUpdate sets LowestUpdate field to given value.

### HasLowestUpdate

`func (o *ResponseInfo) HasLowestUpdate() bool`

HasLowestUpdate returns a boolean if a field has been set.

### GetHighestUpdate

`func (o *ResponseInfo) GetHighestUpdate() int64`

GetHighestUpdate returns the HighestUpdate field if non-nil, zero value otherwise.

### GetHighestUpdateOk

`func (o *ResponseInfo) GetHighestUpdateOk() (*int64, bool)`

GetHighestUpdateOk returns a tuple with the HighestUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHighestUpdate

`func (o *ResponseInfo) SetHighestUpdate(v int64)`

SetHighestUpdate sets HighestUpdate field to given value.

### HasHighestUpdate

`func (o *ResponseInfo) HasHighestUpdate() bool`

HasHighestUpdate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


