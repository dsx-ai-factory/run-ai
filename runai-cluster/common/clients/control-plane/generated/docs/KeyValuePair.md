# KeyValuePair

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Key** | Pointer to **NullableString** | The key of the secret | [optional] 
**Value** | Pointer to **NullableString** | The value of the secret. | [optional] 

## Methods

### NewKeyValuePair

`func NewKeyValuePair() *KeyValuePair`

NewKeyValuePair instantiates a new KeyValuePair object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewKeyValuePairWithDefaults

`func NewKeyValuePairWithDefaults() *KeyValuePair`

NewKeyValuePairWithDefaults instantiates a new KeyValuePair object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKey

`func (o *KeyValuePair) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *KeyValuePair) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *KeyValuePair) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *KeyValuePair) HasKey() bool`

HasKey returns a boolean if a field has been set.

### SetKeyNil

`func (o *KeyValuePair) SetKeyNil(b bool)`

 SetKeyNil sets the value for Key to be an explicit nil

### UnsetKey
`func (o *KeyValuePair) UnsetKey()`

UnsetKey ensures that no value is present for Key, not even an explicit nil
### GetValue

`func (o *KeyValuePair) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *KeyValuePair) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *KeyValuePair) SetValue(v string)`

SetValue sets Value field to given value.

### HasValue

`func (o *KeyValuePair) HasValue() bool`

HasValue returns a boolean if a field has been set.

### SetValueNil

`func (o *KeyValuePair) SetValueNil(b bool)`

 SetValueNil sets the value for Value to be an explicit nil

### UnsetValue
`func (o *KeyValuePair) UnsetValue()`

UnsetValue ensures that no value is present for Value, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


