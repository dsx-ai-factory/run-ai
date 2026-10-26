# ScaleToZeroRetentionField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ScaleToZeroRetentionSeconds** | Pointer to **NullableInt32** | The minimum amount of time (in seconds) that the last replica will remain active after a scale-to-zero decision. Defaults to 0. Available only if minReplicas is set to 0 | [optional] 

## Methods

### NewScaleToZeroRetentionField

`func NewScaleToZeroRetentionField() *ScaleToZeroRetentionField`

NewScaleToZeroRetentionField instantiates a new ScaleToZeroRetentionField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewScaleToZeroRetentionFieldWithDefaults

`func NewScaleToZeroRetentionFieldWithDefaults() *ScaleToZeroRetentionField`

NewScaleToZeroRetentionFieldWithDefaults instantiates a new ScaleToZeroRetentionField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScaleToZeroRetentionSeconds

`func (o *ScaleToZeroRetentionField) GetScaleToZeroRetentionSeconds() int32`

GetScaleToZeroRetentionSeconds returns the ScaleToZeroRetentionSeconds field if non-nil, zero value otherwise.

### GetScaleToZeroRetentionSecondsOk

`func (o *ScaleToZeroRetentionField) GetScaleToZeroRetentionSecondsOk() (*int32, bool)`

GetScaleToZeroRetentionSecondsOk returns a tuple with the ScaleToZeroRetentionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScaleToZeroRetentionSeconds

`func (o *ScaleToZeroRetentionField) SetScaleToZeroRetentionSeconds(v int32)`

SetScaleToZeroRetentionSeconds sets ScaleToZeroRetentionSeconds field to given value.

### HasScaleToZeroRetentionSeconds

`func (o *ScaleToZeroRetentionField) HasScaleToZeroRetentionSeconds() bool`

HasScaleToZeroRetentionSeconds returns a boolean if a field has been set.

### SetScaleToZeroRetentionSecondsNil

`func (o *ScaleToZeroRetentionField) SetScaleToZeroRetentionSecondsNil(b bool)`

 SetScaleToZeroRetentionSecondsNil sets the value for ScaleToZeroRetentionSeconds to be an explicit nil

### UnsetScaleToZeroRetentionSeconds
`func (o *ScaleToZeroRetentionField) UnsetScaleToZeroRetentionSeconds()`

UnsetScaleToZeroRetentionSeconds ensures that no value is present for ScaleToZeroRetentionSeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


