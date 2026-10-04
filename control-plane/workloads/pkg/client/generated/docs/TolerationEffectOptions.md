# TolerationEffectOptions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Options** | Pointer to [**[]TolerationEffectOptionsOptionsInner**](TolerationEffectOptionsOptionsInner.md) | Limit the permitted values for imagePullPolicy to a given set of values. | [optional] 

## Methods

### NewTolerationEffectOptions

`func NewTolerationEffectOptions() *TolerationEffectOptions`

NewTolerationEffectOptions instantiates a new TolerationEffectOptions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTolerationEffectOptionsWithDefaults

`func NewTolerationEffectOptionsWithDefaults() *TolerationEffectOptions`

NewTolerationEffectOptionsWithDefaults instantiates a new TolerationEffectOptions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOptions

`func (o *TolerationEffectOptions) GetOptions() []TolerationEffectOptionsOptionsInner`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *TolerationEffectOptions) GetOptionsOk() (*[]TolerationEffectOptionsOptionsInner, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *TolerationEffectOptions) SetOptions(v []TolerationEffectOptionsOptionsInner)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *TolerationEffectOptions) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *TolerationEffectOptions) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *TolerationEffectOptions) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


