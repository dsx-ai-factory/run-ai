# PodToleration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Key** | Pointer to **string** | Key is the taint key that the toleration applies to. Empty means match all taint keys. | [optional] 
**Operator** | Pointer to **string** | Operator represents a key&#39;s relationship to the value. | [optional] 
**Value** | Pointer to **string** | Value is the taint value the toleration matches to. | [optional] 
**Effect** | Pointer to **string** | Effect indicates the taint effect to match. Empty means match all taint effects. | [optional] 
**TolerationSeconds** | Pointer to **NullableInt32** | TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. | [optional] 

## Methods

### NewPodToleration

`func NewPodToleration() *PodToleration`

NewPodToleration instantiates a new PodToleration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodTolerationWithDefaults

`func NewPodTolerationWithDefaults() *PodToleration`

NewPodTolerationWithDefaults instantiates a new PodToleration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKey

`func (o *PodToleration) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *PodToleration) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *PodToleration) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *PodToleration) HasKey() bool`

HasKey returns a boolean if a field has been set.

### GetOperator

`func (o *PodToleration) GetOperator() string`

GetOperator returns the Operator field if non-nil, zero value otherwise.

### GetOperatorOk

`func (o *PodToleration) GetOperatorOk() (*string, bool)`

GetOperatorOk returns a tuple with the Operator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperator

`func (o *PodToleration) SetOperator(v string)`

SetOperator sets Operator field to given value.

### HasOperator

`func (o *PodToleration) HasOperator() bool`

HasOperator returns a boolean if a field has been set.

### GetValue

`func (o *PodToleration) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *PodToleration) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *PodToleration) SetValue(v string)`

SetValue sets Value field to given value.

### HasValue

`func (o *PodToleration) HasValue() bool`

HasValue returns a boolean if a field has been set.

### GetEffect

`func (o *PodToleration) GetEffect() string`

GetEffect returns the Effect field if non-nil, zero value otherwise.

### GetEffectOk

`func (o *PodToleration) GetEffectOk() (*string, bool)`

GetEffectOk returns a tuple with the Effect field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffect

`func (o *PodToleration) SetEffect(v string)`

SetEffect sets Effect field to given value.

### HasEffect

`func (o *PodToleration) HasEffect() bool`

HasEffect returns a boolean if a field has been set.

### GetTolerationSeconds

`func (o *PodToleration) GetTolerationSeconds() int32`

GetTolerationSeconds returns the TolerationSeconds field if non-nil, zero value otherwise.

### GetTolerationSecondsOk

`func (o *PodToleration) GetTolerationSecondsOk() (*int32, bool)`

GetTolerationSecondsOk returns a tuple with the TolerationSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerationSeconds

`func (o *PodToleration) SetTolerationSeconds(v int32)`

SetTolerationSeconds sets TolerationSeconds field to given value.

### HasTolerationSeconds

`func (o *PodToleration) HasTolerationSeconds() bool`

HasTolerationSeconds returns a boolean if a field has been set.

### SetTolerationSecondsNil

`func (o *PodToleration) SetTolerationSecondsNil(b bool)`

 SetTolerationSecondsNil sets the value for TolerationSeconds to be an explicit nil

### UnsetTolerationSeconds
`func (o *PodToleration) UnsetTolerationSeconds()`

UnsetTolerationSeconds ensures that no value is present for TolerationSeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


