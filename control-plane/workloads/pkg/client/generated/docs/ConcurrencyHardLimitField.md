# ConcurrencyHardLimitField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConcurrencyHardLimit** | Pointer to **NullableInt32** | The maximum number of requests allowed to flow to a single replica at any time. 0 means no limit | [optional] 

## Methods

### NewConcurrencyHardLimitField

`func NewConcurrencyHardLimitField() *ConcurrencyHardLimitField`

NewConcurrencyHardLimitField instantiates a new ConcurrencyHardLimitField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConcurrencyHardLimitFieldWithDefaults

`func NewConcurrencyHardLimitFieldWithDefaults() *ConcurrencyHardLimitField`

NewConcurrencyHardLimitFieldWithDefaults instantiates a new ConcurrencyHardLimitField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConcurrencyHardLimit

`func (o *ConcurrencyHardLimitField) GetConcurrencyHardLimit() int32`

GetConcurrencyHardLimit returns the ConcurrencyHardLimit field if non-nil, zero value otherwise.

### GetConcurrencyHardLimitOk

`func (o *ConcurrencyHardLimitField) GetConcurrencyHardLimitOk() (*int32, bool)`

GetConcurrencyHardLimitOk returns a tuple with the ConcurrencyHardLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConcurrencyHardLimit

`func (o *ConcurrencyHardLimitField) SetConcurrencyHardLimit(v int32)`

SetConcurrencyHardLimit sets ConcurrencyHardLimit field to given value.

### HasConcurrencyHardLimit

`func (o *ConcurrencyHardLimitField) HasConcurrencyHardLimit() bool`

HasConcurrencyHardLimit returns a boolean if a field has been set.

### SetConcurrencyHardLimitNil

`func (o *ConcurrencyHardLimitField) SetConcurrencyHardLimitNil(b bool)`

 SetConcurrencyHardLimitNil sets the value for ConcurrencyHardLimit to be an explicit nil

### UnsetConcurrencyHardLimit
`func (o *ConcurrencyHardLimitField) UnsetConcurrencyHardLimit()`

UnsetConcurrencyHardLimit ensures that no value is present for ConcurrencyHardLimit, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


