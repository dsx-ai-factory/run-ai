# TolerationRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Operator** | Pointer to [**NullableTolerationOperatorRules**](TolerationOperatorRules.md) |  | [optional] 
**Key** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Value** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Effect** | Pointer to [**NullableTolerationEffectRules**](TolerationEffectRules.md) |  | [optional] 
**Seconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Exclude** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 

## Methods

### NewTolerationRules

`func NewTolerationRules() *TolerationRules`

NewTolerationRules instantiates a new TolerationRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTolerationRulesWithDefaults

`func NewTolerationRulesWithDefaults() *TolerationRules`

NewTolerationRulesWithDefaults instantiates a new TolerationRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOperator

`func (o *TolerationRules) GetOperator() TolerationOperatorRules`

GetOperator returns the Operator field if non-nil, zero value otherwise.

### GetOperatorOk

`func (o *TolerationRules) GetOperatorOk() (*TolerationOperatorRules, bool)`

GetOperatorOk returns a tuple with the Operator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperator

`func (o *TolerationRules) SetOperator(v TolerationOperatorRules)`

SetOperator sets Operator field to given value.

### HasOperator

`func (o *TolerationRules) HasOperator() bool`

HasOperator returns a boolean if a field has been set.

### SetOperatorNil

`func (o *TolerationRules) SetOperatorNil(b bool)`

 SetOperatorNil sets the value for Operator to be an explicit nil

### UnsetOperator
`func (o *TolerationRules) UnsetOperator()`

UnsetOperator ensures that no value is present for Operator, not even an explicit nil
### GetKey

`func (o *TolerationRules) GetKey() StringRules`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *TolerationRules) GetKeyOk() (*StringRules, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *TolerationRules) SetKey(v StringRules)`

SetKey sets Key field to given value.

### HasKey

`func (o *TolerationRules) HasKey() bool`

HasKey returns a boolean if a field has been set.

### SetKeyNil

`func (o *TolerationRules) SetKeyNil(b bool)`

 SetKeyNil sets the value for Key to be an explicit nil

### UnsetKey
`func (o *TolerationRules) UnsetKey()`

UnsetKey ensures that no value is present for Key, not even an explicit nil
### GetValue

`func (o *TolerationRules) GetValue() StringRules`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *TolerationRules) GetValueOk() (*StringRules, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *TolerationRules) SetValue(v StringRules)`

SetValue sets Value field to given value.

### HasValue

`func (o *TolerationRules) HasValue() bool`

HasValue returns a boolean if a field has been set.

### SetValueNil

`func (o *TolerationRules) SetValueNil(b bool)`

 SetValueNil sets the value for Value to be an explicit nil

### UnsetValue
`func (o *TolerationRules) UnsetValue()`

UnsetValue ensures that no value is present for Value, not even an explicit nil
### GetEffect

`func (o *TolerationRules) GetEffect() TolerationEffectRules`

GetEffect returns the Effect field if non-nil, zero value otherwise.

### GetEffectOk

`func (o *TolerationRules) GetEffectOk() (*TolerationEffectRules, bool)`

GetEffectOk returns a tuple with the Effect field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffect

`func (o *TolerationRules) SetEffect(v TolerationEffectRules)`

SetEffect sets Effect field to given value.

### HasEffect

`func (o *TolerationRules) HasEffect() bool`

HasEffect returns a boolean if a field has been set.

### SetEffectNil

`func (o *TolerationRules) SetEffectNil(b bool)`

 SetEffectNil sets the value for Effect to be an explicit nil

### UnsetEffect
`func (o *TolerationRules) UnsetEffect()`

UnsetEffect ensures that no value is present for Effect, not even an explicit nil
### GetSeconds

`func (o *TolerationRules) GetSeconds() IntegerRules`

GetSeconds returns the Seconds field if non-nil, zero value otherwise.

### GetSecondsOk

`func (o *TolerationRules) GetSecondsOk() (*IntegerRules, bool)`

GetSecondsOk returns a tuple with the Seconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeconds

`func (o *TolerationRules) SetSeconds(v IntegerRules)`

SetSeconds sets Seconds field to given value.

### HasSeconds

`func (o *TolerationRules) HasSeconds() bool`

HasSeconds returns a boolean if a field has been set.

### SetSecondsNil

`func (o *TolerationRules) SetSecondsNil(b bool)`

 SetSecondsNil sets the value for Seconds to be an explicit nil

### UnsetSeconds
`func (o *TolerationRules) UnsetSeconds()`

UnsetSeconds ensures that no value is present for Seconds, not even an explicit nil
### GetExclude

`func (o *TolerationRules) GetExclude() BooleanRules`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *TolerationRules) GetExcludeOk() (*BooleanRules, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *TolerationRules) SetExclude(v BooleanRules)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *TolerationRules) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *TolerationRules) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *TolerationRules) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


