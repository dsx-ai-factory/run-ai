# AutoScalingCommonFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MinReplicas** | Pointer to **NullableInt32** | The minimum number of replicas for autoscaling. Defaults to 1. Use 0 to allow scale-to-zero | [optional] 
**MaxReplicas** | Pointer to **NullableInt32** | The maximum number of replicas for autoscaling. Defaults to minReplicas, or to 1 if minReplicas is set to 0 | [optional] 
**ScaleToZeroRetentionSeconds** | Pointer to **NullableInt32** | The minimum amount of time (in seconds) that the last replica will remain active after a scale-to-zero decision. Defaults to 0. Available only if minReplicas is set to 0 | [optional] 

## Methods

### NewAutoScalingCommonFields

`func NewAutoScalingCommonFields() *AutoScalingCommonFields`

NewAutoScalingCommonFields instantiates a new AutoScalingCommonFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAutoScalingCommonFieldsWithDefaults

`func NewAutoScalingCommonFieldsWithDefaults() *AutoScalingCommonFields`

NewAutoScalingCommonFieldsWithDefaults instantiates a new AutoScalingCommonFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMinReplicas

`func (o *AutoScalingCommonFields) GetMinReplicas() int32`

GetMinReplicas returns the MinReplicas field if non-nil, zero value otherwise.

### GetMinReplicasOk

`func (o *AutoScalingCommonFields) GetMinReplicasOk() (*int32, bool)`

GetMinReplicasOk returns a tuple with the MinReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinReplicas

`func (o *AutoScalingCommonFields) SetMinReplicas(v int32)`

SetMinReplicas sets MinReplicas field to given value.

### HasMinReplicas

`func (o *AutoScalingCommonFields) HasMinReplicas() bool`

HasMinReplicas returns a boolean if a field has been set.

### SetMinReplicasNil

`func (o *AutoScalingCommonFields) SetMinReplicasNil(b bool)`

 SetMinReplicasNil sets the value for MinReplicas to be an explicit nil

### UnsetMinReplicas
`func (o *AutoScalingCommonFields) UnsetMinReplicas()`

UnsetMinReplicas ensures that no value is present for MinReplicas, not even an explicit nil
### GetMaxReplicas

`func (o *AutoScalingCommonFields) GetMaxReplicas() int32`

GetMaxReplicas returns the MaxReplicas field if non-nil, zero value otherwise.

### GetMaxReplicasOk

`func (o *AutoScalingCommonFields) GetMaxReplicasOk() (*int32, bool)`

GetMaxReplicasOk returns a tuple with the MaxReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxReplicas

`func (o *AutoScalingCommonFields) SetMaxReplicas(v int32)`

SetMaxReplicas sets MaxReplicas field to given value.

### HasMaxReplicas

`func (o *AutoScalingCommonFields) HasMaxReplicas() bool`

HasMaxReplicas returns a boolean if a field has been set.

### SetMaxReplicasNil

`func (o *AutoScalingCommonFields) SetMaxReplicasNil(b bool)`

 SetMaxReplicasNil sets the value for MaxReplicas to be an explicit nil

### UnsetMaxReplicas
`func (o *AutoScalingCommonFields) UnsetMaxReplicas()`

UnsetMaxReplicas ensures that no value is present for MaxReplicas, not even an explicit nil
### GetScaleToZeroRetentionSeconds

`func (o *AutoScalingCommonFields) GetScaleToZeroRetentionSeconds() int32`

GetScaleToZeroRetentionSeconds returns the ScaleToZeroRetentionSeconds field if non-nil, zero value otherwise.

### GetScaleToZeroRetentionSecondsOk

`func (o *AutoScalingCommonFields) GetScaleToZeroRetentionSecondsOk() (*int32, bool)`

GetScaleToZeroRetentionSecondsOk returns a tuple with the ScaleToZeroRetentionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScaleToZeroRetentionSeconds

`func (o *AutoScalingCommonFields) SetScaleToZeroRetentionSeconds(v int32)`

SetScaleToZeroRetentionSeconds sets ScaleToZeroRetentionSeconds field to given value.

### HasScaleToZeroRetentionSeconds

`func (o *AutoScalingCommonFields) HasScaleToZeroRetentionSeconds() bool`

HasScaleToZeroRetentionSeconds returns a boolean if a field has been set.

### SetScaleToZeroRetentionSecondsNil

`func (o *AutoScalingCommonFields) SetScaleToZeroRetentionSecondsNil(b bool)`

 SetScaleToZeroRetentionSecondsNil sets the value for ScaleToZeroRetentionSeconds to be an explicit nil

### UnsetScaleToZeroRetentionSeconds
`func (o *AutoScalingCommonFields) UnsetScaleToZeroRetentionSeconds()`

UnsetScaleToZeroRetentionSeconds ensures that no value is present for ScaleToZeroRetentionSeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


