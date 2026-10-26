# ScaleDownDelayField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ScaleDownDelaySeconds** | Pointer to **NullableInt32** | The minimum amount of time (in seconds) that a replica will remain active after a scale-down decision | [optional] 

## Methods

### NewScaleDownDelayField

`func NewScaleDownDelayField() *ScaleDownDelayField`

NewScaleDownDelayField instantiates a new ScaleDownDelayField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewScaleDownDelayFieldWithDefaults

`func NewScaleDownDelayFieldWithDefaults() *ScaleDownDelayField`

NewScaleDownDelayFieldWithDefaults instantiates a new ScaleDownDelayField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScaleDownDelaySeconds

`func (o *ScaleDownDelayField) GetScaleDownDelaySeconds() int32`

GetScaleDownDelaySeconds returns the ScaleDownDelaySeconds field if non-nil, zero value otherwise.

### GetScaleDownDelaySecondsOk

`func (o *ScaleDownDelayField) GetScaleDownDelaySecondsOk() (*int32, bool)`

GetScaleDownDelaySecondsOk returns a tuple with the ScaleDownDelaySeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScaleDownDelaySeconds

`func (o *ScaleDownDelayField) SetScaleDownDelaySeconds(v int32)`

SetScaleDownDelaySeconds sets ScaleDownDelaySeconds field to given value.

### HasScaleDownDelaySeconds

`func (o *ScaleDownDelayField) HasScaleDownDelaySeconds() bool`

HasScaleDownDelaySeconds returns a boolean if a field has been set.

### SetScaleDownDelaySecondsNil

`func (o *ScaleDownDelayField) SetScaleDownDelaySecondsNil(b bool)`

 SetScaleDownDelaySecondsNil sets the value for ScaleDownDelaySeconds to be an explicit nil

### UnsetScaleDownDelaySeconds
`func (o *ScaleDownDelayField) UnsetScaleDownDelaySeconds()`

UnsetScaleDownDelaySeconds ensures that no value is present for ScaleDownDelaySeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


