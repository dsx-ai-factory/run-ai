# Toleration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | The name of the toleration. (mandatory) | [optional] 
**Operator** | Pointer to [**NullableTolerationOperator**](TolerationOperator.md) |  | [optional] 
**Key** | Pointer to **NullableString** | The taint key that the toleration applies to. (mandatory) | [optional] 
**Value** | Pointer to **NullableString** | The taint value the toleration matches to. Mandatory if operator is Exists, forbidden otherwise. | [optional] 
**Effect** | Pointer to [**NullableTolerationEffect**](TolerationEffect.md) |  | [optional] 
**Seconds** | Pointer to **NullableInt32** | The period of time the toleration tolerates the taint. Valid only if effect is NoExecute. taint. | [optional] 

## Methods

### NewToleration

`func NewToleration() *Toleration`

NewToleration instantiates a new Toleration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTolerationWithDefaults

`func NewTolerationWithDefaults() *Toleration`

NewTolerationWithDefaults instantiates a new Toleration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Toleration) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Toleration) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Toleration) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *Toleration) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *Toleration) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *Toleration) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetOperator

`func (o *Toleration) GetOperator() TolerationOperator`

GetOperator returns the Operator field if non-nil, zero value otherwise.

### GetOperatorOk

`func (o *Toleration) GetOperatorOk() (*TolerationOperator, bool)`

GetOperatorOk returns a tuple with the Operator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperator

`func (o *Toleration) SetOperator(v TolerationOperator)`

SetOperator sets Operator field to given value.

### HasOperator

`func (o *Toleration) HasOperator() bool`

HasOperator returns a boolean if a field has been set.

### SetOperatorNil

`func (o *Toleration) SetOperatorNil(b bool)`

 SetOperatorNil sets the value for Operator to be an explicit nil

### UnsetOperator
`func (o *Toleration) UnsetOperator()`

UnsetOperator ensures that no value is present for Operator, not even an explicit nil
### GetKey

`func (o *Toleration) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *Toleration) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *Toleration) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *Toleration) HasKey() bool`

HasKey returns a boolean if a field has been set.

### SetKeyNil

`func (o *Toleration) SetKeyNil(b bool)`

 SetKeyNil sets the value for Key to be an explicit nil

### UnsetKey
`func (o *Toleration) UnsetKey()`

UnsetKey ensures that no value is present for Key, not even an explicit nil
### GetValue

`func (o *Toleration) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *Toleration) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *Toleration) SetValue(v string)`

SetValue sets Value field to given value.

### HasValue

`func (o *Toleration) HasValue() bool`

HasValue returns a boolean if a field has been set.

### SetValueNil

`func (o *Toleration) SetValueNil(b bool)`

 SetValueNil sets the value for Value to be an explicit nil

### UnsetValue
`func (o *Toleration) UnsetValue()`

UnsetValue ensures that no value is present for Value, not even an explicit nil
### GetEffect

`func (o *Toleration) GetEffect() TolerationEffect`

GetEffect returns the Effect field if non-nil, zero value otherwise.

### GetEffectOk

`func (o *Toleration) GetEffectOk() (*TolerationEffect, bool)`

GetEffectOk returns a tuple with the Effect field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffect

`func (o *Toleration) SetEffect(v TolerationEffect)`

SetEffect sets Effect field to given value.

### HasEffect

`func (o *Toleration) HasEffect() bool`

HasEffect returns a boolean if a field has been set.

### SetEffectNil

`func (o *Toleration) SetEffectNil(b bool)`

 SetEffectNil sets the value for Effect to be an explicit nil

### UnsetEffect
`func (o *Toleration) UnsetEffect()`

UnsetEffect ensures that no value is present for Effect, not even an explicit nil
### GetSeconds

`func (o *Toleration) GetSeconds() int32`

GetSeconds returns the Seconds field if non-nil, zero value otherwise.

### GetSecondsOk

`func (o *Toleration) GetSecondsOk() (*int32, bool)`

GetSecondsOk returns a tuple with the Seconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeconds

`func (o *Toleration) SetSeconds(v int32)`

SetSeconds sets Seconds field to given value.

### HasSeconds

`func (o *Toleration) HasSeconds() bool`

HasSeconds returns a boolean if a field has been set.

### SetSecondsNil

`func (o *Toleration) SetSecondsNil(b bool)`

 SetSecondsNil sets the value for Seconds to be an explicit nil

### UnsetSeconds
`func (o *Toleration) UnsetSeconds()`

UnsetSeconds ensures that no value is present for Seconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


