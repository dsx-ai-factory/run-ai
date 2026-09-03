# Toleration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Key** | Pointer to **string** | Key is the taint key that the toleration applies to. Empty means match all taint keys. | [optional] 
**Operator** | Pointer to **string** | Operator represents a key&#39;s relationship to the value. | [optional] 
**Value** | Pointer to **string** | Value is the taint value the toleration matches to. | [optional] 
**Effect** | Pointer to **string** | Effect indicates the taint effect to match. Empty means match all taint effects. | [optional] 
**TolerationSeconds** | Pointer to **NullableInt32** | TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. | [optional] 

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

### GetOperator

`func (o *Toleration) GetOperator() string`

GetOperator returns the Operator field if non-nil, zero value otherwise.

### GetOperatorOk

`func (o *Toleration) GetOperatorOk() (*string, bool)`

GetOperatorOk returns a tuple with the Operator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperator

`func (o *Toleration) SetOperator(v string)`

SetOperator sets Operator field to given value.

### HasOperator

`func (o *Toleration) HasOperator() bool`

HasOperator returns a boolean if a field has been set.

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

### GetEffect

`func (o *Toleration) GetEffect() string`

GetEffect returns the Effect field if non-nil, zero value otherwise.

### GetEffectOk

`func (o *Toleration) GetEffectOk() (*string, bool)`

GetEffectOk returns a tuple with the Effect field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffect

`func (o *Toleration) SetEffect(v string)`

SetEffect sets Effect field to given value.

### HasEffect

`func (o *Toleration) HasEffect() bool`

HasEffect returns a boolean if a field has been set.

### GetTolerationSeconds

`func (o *Toleration) GetTolerationSeconds() int32`

GetTolerationSeconds returns the TolerationSeconds field if non-nil, zero value otherwise.

### GetTolerationSecondsOk

`func (o *Toleration) GetTolerationSecondsOk() (*int32, bool)`

GetTolerationSecondsOk returns a tuple with the TolerationSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerationSeconds

`func (o *Toleration) SetTolerationSeconds(v int32)`

SetTolerationSeconds sets TolerationSeconds field to given value.

### HasTolerationSeconds

`func (o *Toleration) HasTolerationSeconds() bool`

HasTolerationSeconds returns a boolean if a field has been set.

### SetTolerationSecondsNil

`func (o *Toleration) SetTolerationSecondsNil(b bool)`

 SetTolerationSecondsNil sets the value for TolerationSeconds to be an explicit nil

### UnsetTolerationSeconds
`func (o *Toleration) UnsetTolerationSeconds()`

UnsetTolerationSeconds ensures that no value is present for TolerationSeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


