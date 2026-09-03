# SpecificRunAutoScalingAutoScaling

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MinReplicas** | Pointer to **NullableInt32** | The minimum number of replicas for autoscaling. Defaults to 1. Use 0 to allow scale-to-zero | [optional] 
**MaxReplicas** | Pointer to **NullableInt32** | The maximum number of replicas for autoscaling. Defaults to minReplicas, or to 1 if minReplicas is set to 0 | [optional] 
**ScaleToZeroRetentionSeconds** | Pointer to **NullableInt32** | The minimum amount of time (in seconds) that the last replica will remain active after a scale-to-zero decision. Defaults to 0. Available only if minReplicas is set to 0 | [optional] 
**ThresholdMetric** | Pointer to [**NullableAutoScalingMetric**](AutoScalingMetric.md) |  | [optional] 
**ThresholdValue** | Pointer to **NullableInt32** | The threshold value to use with the specified metric for autoscaling | [optional] 

## Methods

### NewSpecificRunAutoScalingAutoScaling

`func NewSpecificRunAutoScalingAutoScaling() *SpecificRunAutoScalingAutoScaling`

NewSpecificRunAutoScalingAutoScaling instantiates a new SpecificRunAutoScalingAutoScaling object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSpecificRunAutoScalingAutoScalingWithDefaults

`func NewSpecificRunAutoScalingAutoScalingWithDefaults() *SpecificRunAutoScalingAutoScaling`

NewSpecificRunAutoScalingAutoScalingWithDefaults instantiates a new SpecificRunAutoScalingAutoScaling object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMinReplicas

`func (o *SpecificRunAutoScalingAutoScaling) GetMinReplicas() int32`

GetMinReplicas returns the MinReplicas field if non-nil, zero value otherwise.

### GetMinReplicasOk

`func (o *SpecificRunAutoScalingAutoScaling) GetMinReplicasOk() (*int32, bool)`

GetMinReplicasOk returns a tuple with the MinReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinReplicas

`func (o *SpecificRunAutoScalingAutoScaling) SetMinReplicas(v int32)`

SetMinReplicas sets MinReplicas field to given value.

### HasMinReplicas

`func (o *SpecificRunAutoScalingAutoScaling) HasMinReplicas() bool`

HasMinReplicas returns a boolean if a field has been set.

### SetMinReplicasNil

`func (o *SpecificRunAutoScalingAutoScaling) SetMinReplicasNil(b bool)`

 SetMinReplicasNil sets the value for MinReplicas to be an explicit nil

### UnsetMinReplicas
`func (o *SpecificRunAutoScalingAutoScaling) UnsetMinReplicas()`

UnsetMinReplicas ensures that no value is present for MinReplicas, not even an explicit nil
### GetMaxReplicas

`func (o *SpecificRunAutoScalingAutoScaling) GetMaxReplicas() int32`

GetMaxReplicas returns the MaxReplicas field if non-nil, zero value otherwise.

### GetMaxReplicasOk

`func (o *SpecificRunAutoScalingAutoScaling) GetMaxReplicasOk() (*int32, bool)`

GetMaxReplicasOk returns a tuple with the MaxReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxReplicas

`func (o *SpecificRunAutoScalingAutoScaling) SetMaxReplicas(v int32)`

SetMaxReplicas sets MaxReplicas field to given value.

### HasMaxReplicas

`func (o *SpecificRunAutoScalingAutoScaling) HasMaxReplicas() bool`

HasMaxReplicas returns a boolean if a field has been set.

### SetMaxReplicasNil

`func (o *SpecificRunAutoScalingAutoScaling) SetMaxReplicasNil(b bool)`

 SetMaxReplicasNil sets the value for MaxReplicas to be an explicit nil

### UnsetMaxReplicas
`func (o *SpecificRunAutoScalingAutoScaling) UnsetMaxReplicas()`

UnsetMaxReplicas ensures that no value is present for MaxReplicas, not even an explicit nil
### GetScaleToZeroRetentionSeconds

`func (o *SpecificRunAutoScalingAutoScaling) GetScaleToZeroRetentionSeconds() int32`

GetScaleToZeroRetentionSeconds returns the ScaleToZeroRetentionSeconds field if non-nil, zero value otherwise.

### GetScaleToZeroRetentionSecondsOk

`func (o *SpecificRunAutoScalingAutoScaling) GetScaleToZeroRetentionSecondsOk() (*int32, bool)`

GetScaleToZeroRetentionSecondsOk returns a tuple with the ScaleToZeroRetentionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScaleToZeroRetentionSeconds

`func (o *SpecificRunAutoScalingAutoScaling) SetScaleToZeroRetentionSeconds(v int32)`

SetScaleToZeroRetentionSeconds sets ScaleToZeroRetentionSeconds field to given value.

### HasScaleToZeroRetentionSeconds

`func (o *SpecificRunAutoScalingAutoScaling) HasScaleToZeroRetentionSeconds() bool`

HasScaleToZeroRetentionSeconds returns a boolean if a field has been set.

### SetScaleToZeroRetentionSecondsNil

`func (o *SpecificRunAutoScalingAutoScaling) SetScaleToZeroRetentionSecondsNil(b bool)`

 SetScaleToZeroRetentionSecondsNil sets the value for ScaleToZeroRetentionSeconds to be an explicit nil

### UnsetScaleToZeroRetentionSeconds
`func (o *SpecificRunAutoScalingAutoScaling) UnsetScaleToZeroRetentionSeconds()`

UnsetScaleToZeroRetentionSeconds ensures that no value is present for ScaleToZeroRetentionSeconds, not even an explicit nil
### GetThresholdMetric

`func (o *SpecificRunAutoScalingAutoScaling) GetThresholdMetric() AutoScalingMetric`

GetThresholdMetric returns the ThresholdMetric field if non-nil, zero value otherwise.

### GetThresholdMetricOk

`func (o *SpecificRunAutoScalingAutoScaling) GetThresholdMetricOk() (*AutoScalingMetric, bool)`

GetThresholdMetricOk returns a tuple with the ThresholdMetric field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThresholdMetric

`func (o *SpecificRunAutoScalingAutoScaling) SetThresholdMetric(v AutoScalingMetric)`

SetThresholdMetric sets ThresholdMetric field to given value.

### HasThresholdMetric

`func (o *SpecificRunAutoScalingAutoScaling) HasThresholdMetric() bool`

HasThresholdMetric returns a boolean if a field has been set.

### SetThresholdMetricNil

`func (o *SpecificRunAutoScalingAutoScaling) SetThresholdMetricNil(b bool)`

 SetThresholdMetricNil sets the value for ThresholdMetric to be an explicit nil

### UnsetThresholdMetric
`func (o *SpecificRunAutoScalingAutoScaling) UnsetThresholdMetric()`

UnsetThresholdMetric ensures that no value is present for ThresholdMetric, not even an explicit nil
### GetThresholdValue

`func (o *SpecificRunAutoScalingAutoScaling) GetThresholdValue() int32`

GetThresholdValue returns the ThresholdValue field if non-nil, zero value otherwise.

### GetThresholdValueOk

`func (o *SpecificRunAutoScalingAutoScaling) GetThresholdValueOk() (*int32, bool)`

GetThresholdValueOk returns a tuple with the ThresholdValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThresholdValue

`func (o *SpecificRunAutoScalingAutoScaling) SetThresholdValue(v int32)`

SetThresholdValue sets ThresholdValue field to given value.

### HasThresholdValue

`func (o *SpecificRunAutoScalingAutoScaling) HasThresholdValue() bool`

HasThresholdValue returns a boolean if a field has been set.

### SetThresholdValueNil

`func (o *SpecificRunAutoScalingAutoScaling) SetThresholdValueNil(b bool)`

 SetThresholdValueNil sets the value for ThresholdValue to be an explicit nil

### UnsetThresholdValue
`func (o *SpecificRunAutoScalingAutoScaling) UnsetThresholdValue()`

UnsetThresholdValue ensures that no value is present for ThresholdValue, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


