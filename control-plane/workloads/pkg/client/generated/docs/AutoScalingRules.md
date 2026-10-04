# AutoScalingRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MetricThresholdPercentage** | Pointer to [**NullableNumberRules**](NumberRules.md) |  | [optional] 
**MinReplicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**MaxReplicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**InitialReplicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**ActivationReplicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Metric** | Pointer to [**NullableAutoScalingMetricRules**](AutoScalingMetricRules.md) |  | [optional] 
**MetricThreshold** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**ConcurrencyHardLimit** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**ScaleToZeroRetentionSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**ScaleDownDelaySeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**InitializationTimeoutSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 

## Methods

### NewAutoScalingRules

`func NewAutoScalingRules() *AutoScalingRules`

NewAutoScalingRules instantiates a new AutoScalingRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAutoScalingRulesWithDefaults

`func NewAutoScalingRulesWithDefaults() *AutoScalingRules`

NewAutoScalingRulesWithDefaults instantiates a new AutoScalingRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetricThresholdPercentage

`func (o *AutoScalingRules) GetMetricThresholdPercentage() NumberRules`

GetMetricThresholdPercentage returns the MetricThresholdPercentage field if non-nil, zero value otherwise.

### GetMetricThresholdPercentageOk

`func (o *AutoScalingRules) GetMetricThresholdPercentageOk() (*NumberRules, bool)`

GetMetricThresholdPercentageOk returns a tuple with the MetricThresholdPercentage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricThresholdPercentage

`func (o *AutoScalingRules) SetMetricThresholdPercentage(v NumberRules)`

SetMetricThresholdPercentage sets MetricThresholdPercentage field to given value.

### HasMetricThresholdPercentage

`func (o *AutoScalingRules) HasMetricThresholdPercentage() bool`

HasMetricThresholdPercentage returns a boolean if a field has been set.

### SetMetricThresholdPercentageNil

`func (o *AutoScalingRules) SetMetricThresholdPercentageNil(b bool)`

 SetMetricThresholdPercentageNil sets the value for MetricThresholdPercentage to be an explicit nil

### UnsetMetricThresholdPercentage
`func (o *AutoScalingRules) UnsetMetricThresholdPercentage()`

UnsetMetricThresholdPercentage ensures that no value is present for MetricThresholdPercentage, not even an explicit nil
### GetMinReplicas

`func (o *AutoScalingRules) GetMinReplicas() IntegerRules`

GetMinReplicas returns the MinReplicas field if non-nil, zero value otherwise.

### GetMinReplicasOk

`func (o *AutoScalingRules) GetMinReplicasOk() (*IntegerRules, bool)`

GetMinReplicasOk returns a tuple with the MinReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinReplicas

`func (o *AutoScalingRules) SetMinReplicas(v IntegerRules)`

SetMinReplicas sets MinReplicas field to given value.

### HasMinReplicas

`func (o *AutoScalingRules) HasMinReplicas() bool`

HasMinReplicas returns a boolean if a field has been set.

### SetMinReplicasNil

`func (o *AutoScalingRules) SetMinReplicasNil(b bool)`

 SetMinReplicasNil sets the value for MinReplicas to be an explicit nil

### UnsetMinReplicas
`func (o *AutoScalingRules) UnsetMinReplicas()`

UnsetMinReplicas ensures that no value is present for MinReplicas, not even an explicit nil
### GetMaxReplicas

`func (o *AutoScalingRules) GetMaxReplicas() IntegerRules`

GetMaxReplicas returns the MaxReplicas field if non-nil, zero value otherwise.

### GetMaxReplicasOk

`func (o *AutoScalingRules) GetMaxReplicasOk() (*IntegerRules, bool)`

GetMaxReplicasOk returns a tuple with the MaxReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxReplicas

`func (o *AutoScalingRules) SetMaxReplicas(v IntegerRules)`

SetMaxReplicas sets MaxReplicas field to given value.

### HasMaxReplicas

`func (o *AutoScalingRules) HasMaxReplicas() bool`

HasMaxReplicas returns a boolean if a field has been set.

### SetMaxReplicasNil

`func (o *AutoScalingRules) SetMaxReplicasNil(b bool)`

 SetMaxReplicasNil sets the value for MaxReplicas to be an explicit nil

### UnsetMaxReplicas
`func (o *AutoScalingRules) UnsetMaxReplicas()`

UnsetMaxReplicas ensures that no value is present for MaxReplicas, not even an explicit nil
### GetInitialReplicas

`func (o *AutoScalingRules) GetInitialReplicas() IntegerRules`

GetInitialReplicas returns the InitialReplicas field if non-nil, zero value otherwise.

### GetInitialReplicasOk

`func (o *AutoScalingRules) GetInitialReplicasOk() (*IntegerRules, bool)`

GetInitialReplicasOk returns a tuple with the InitialReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitialReplicas

`func (o *AutoScalingRules) SetInitialReplicas(v IntegerRules)`

SetInitialReplicas sets InitialReplicas field to given value.

### HasInitialReplicas

`func (o *AutoScalingRules) HasInitialReplicas() bool`

HasInitialReplicas returns a boolean if a field has been set.

### SetInitialReplicasNil

`func (o *AutoScalingRules) SetInitialReplicasNil(b bool)`

 SetInitialReplicasNil sets the value for InitialReplicas to be an explicit nil

### UnsetInitialReplicas
`func (o *AutoScalingRules) UnsetInitialReplicas()`

UnsetInitialReplicas ensures that no value is present for InitialReplicas, not even an explicit nil
### GetActivationReplicas

`func (o *AutoScalingRules) GetActivationReplicas() IntegerRules`

GetActivationReplicas returns the ActivationReplicas field if non-nil, zero value otherwise.

### GetActivationReplicasOk

`func (o *AutoScalingRules) GetActivationReplicasOk() (*IntegerRules, bool)`

GetActivationReplicasOk returns a tuple with the ActivationReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActivationReplicas

`func (o *AutoScalingRules) SetActivationReplicas(v IntegerRules)`

SetActivationReplicas sets ActivationReplicas field to given value.

### HasActivationReplicas

`func (o *AutoScalingRules) HasActivationReplicas() bool`

HasActivationReplicas returns a boolean if a field has been set.

### SetActivationReplicasNil

`func (o *AutoScalingRules) SetActivationReplicasNil(b bool)`

 SetActivationReplicasNil sets the value for ActivationReplicas to be an explicit nil

### UnsetActivationReplicas
`func (o *AutoScalingRules) UnsetActivationReplicas()`

UnsetActivationReplicas ensures that no value is present for ActivationReplicas, not even an explicit nil
### GetMetric

`func (o *AutoScalingRules) GetMetric() AutoScalingMetricRules`

GetMetric returns the Metric field if non-nil, zero value otherwise.

### GetMetricOk

`func (o *AutoScalingRules) GetMetricOk() (*AutoScalingMetricRules, bool)`

GetMetricOk returns a tuple with the Metric field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetric

`func (o *AutoScalingRules) SetMetric(v AutoScalingMetricRules)`

SetMetric sets Metric field to given value.

### HasMetric

`func (o *AutoScalingRules) HasMetric() bool`

HasMetric returns a boolean if a field has been set.

### SetMetricNil

`func (o *AutoScalingRules) SetMetricNil(b bool)`

 SetMetricNil sets the value for Metric to be an explicit nil

### UnsetMetric
`func (o *AutoScalingRules) UnsetMetric()`

UnsetMetric ensures that no value is present for Metric, not even an explicit nil
### GetMetricThreshold

`func (o *AutoScalingRules) GetMetricThreshold() IntegerRules`

GetMetricThreshold returns the MetricThreshold field if non-nil, zero value otherwise.

### GetMetricThresholdOk

`func (o *AutoScalingRules) GetMetricThresholdOk() (*IntegerRules, bool)`

GetMetricThresholdOk returns a tuple with the MetricThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricThreshold

`func (o *AutoScalingRules) SetMetricThreshold(v IntegerRules)`

SetMetricThreshold sets MetricThreshold field to given value.

### HasMetricThreshold

`func (o *AutoScalingRules) HasMetricThreshold() bool`

HasMetricThreshold returns a boolean if a field has been set.

### SetMetricThresholdNil

`func (o *AutoScalingRules) SetMetricThresholdNil(b bool)`

 SetMetricThresholdNil sets the value for MetricThreshold to be an explicit nil

### UnsetMetricThreshold
`func (o *AutoScalingRules) UnsetMetricThreshold()`

UnsetMetricThreshold ensures that no value is present for MetricThreshold, not even an explicit nil
### GetConcurrencyHardLimit

`func (o *AutoScalingRules) GetConcurrencyHardLimit() IntegerRules`

GetConcurrencyHardLimit returns the ConcurrencyHardLimit field if non-nil, zero value otherwise.

### GetConcurrencyHardLimitOk

`func (o *AutoScalingRules) GetConcurrencyHardLimitOk() (*IntegerRules, bool)`

GetConcurrencyHardLimitOk returns a tuple with the ConcurrencyHardLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConcurrencyHardLimit

`func (o *AutoScalingRules) SetConcurrencyHardLimit(v IntegerRules)`

SetConcurrencyHardLimit sets ConcurrencyHardLimit field to given value.

### HasConcurrencyHardLimit

`func (o *AutoScalingRules) HasConcurrencyHardLimit() bool`

HasConcurrencyHardLimit returns a boolean if a field has been set.

### SetConcurrencyHardLimitNil

`func (o *AutoScalingRules) SetConcurrencyHardLimitNil(b bool)`

 SetConcurrencyHardLimitNil sets the value for ConcurrencyHardLimit to be an explicit nil

### UnsetConcurrencyHardLimit
`func (o *AutoScalingRules) UnsetConcurrencyHardLimit()`

UnsetConcurrencyHardLimit ensures that no value is present for ConcurrencyHardLimit, not even an explicit nil
### GetScaleToZeroRetentionSeconds

`func (o *AutoScalingRules) GetScaleToZeroRetentionSeconds() IntegerRules`

GetScaleToZeroRetentionSeconds returns the ScaleToZeroRetentionSeconds field if non-nil, zero value otherwise.

### GetScaleToZeroRetentionSecondsOk

`func (o *AutoScalingRules) GetScaleToZeroRetentionSecondsOk() (*IntegerRules, bool)`

GetScaleToZeroRetentionSecondsOk returns a tuple with the ScaleToZeroRetentionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScaleToZeroRetentionSeconds

`func (o *AutoScalingRules) SetScaleToZeroRetentionSeconds(v IntegerRules)`

SetScaleToZeroRetentionSeconds sets ScaleToZeroRetentionSeconds field to given value.

### HasScaleToZeroRetentionSeconds

`func (o *AutoScalingRules) HasScaleToZeroRetentionSeconds() bool`

HasScaleToZeroRetentionSeconds returns a boolean if a field has been set.

### SetScaleToZeroRetentionSecondsNil

`func (o *AutoScalingRules) SetScaleToZeroRetentionSecondsNil(b bool)`

 SetScaleToZeroRetentionSecondsNil sets the value for ScaleToZeroRetentionSeconds to be an explicit nil

### UnsetScaleToZeroRetentionSeconds
`func (o *AutoScalingRules) UnsetScaleToZeroRetentionSeconds()`

UnsetScaleToZeroRetentionSeconds ensures that no value is present for ScaleToZeroRetentionSeconds, not even an explicit nil
### GetScaleDownDelaySeconds

`func (o *AutoScalingRules) GetScaleDownDelaySeconds() IntegerRules`

GetScaleDownDelaySeconds returns the ScaleDownDelaySeconds field if non-nil, zero value otherwise.

### GetScaleDownDelaySecondsOk

`func (o *AutoScalingRules) GetScaleDownDelaySecondsOk() (*IntegerRules, bool)`

GetScaleDownDelaySecondsOk returns a tuple with the ScaleDownDelaySeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScaleDownDelaySeconds

`func (o *AutoScalingRules) SetScaleDownDelaySeconds(v IntegerRules)`

SetScaleDownDelaySeconds sets ScaleDownDelaySeconds field to given value.

### HasScaleDownDelaySeconds

`func (o *AutoScalingRules) HasScaleDownDelaySeconds() bool`

HasScaleDownDelaySeconds returns a boolean if a field has been set.

### SetScaleDownDelaySecondsNil

`func (o *AutoScalingRules) SetScaleDownDelaySecondsNil(b bool)`

 SetScaleDownDelaySecondsNil sets the value for ScaleDownDelaySeconds to be an explicit nil

### UnsetScaleDownDelaySeconds
`func (o *AutoScalingRules) UnsetScaleDownDelaySeconds()`

UnsetScaleDownDelaySeconds ensures that no value is present for ScaleDownDelaySeconds, not even an explicit nil
### GetInitializationTimeoutSeconds

`func (o *AutoScalingRules) GetInitializationTimeoutSeconds() IntegerRules`

GetInitializationTimeoutSeconds returns the InitializationTimeoutSeconds field if non-nil, zero value otherwise.

### GetInitializationTimeoutSecondsOk

`func (o *AutoScalingRules) GetInitializationTimeoutSecondsOk() (*IntegerRules, bool)`

GetInitializationTimeoutSecondsOk returns a tuple with the InitializationTimeoutSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitializationTimeoutSeconds

`func (o *AutoScalingRules) SetInitializationTimeoutSeconds(v IntegerRules)`

SetInitializationTimeoutSeconds sets InitializationTimeoutSeconds field to given value.

### HasInitializationTimeoutSeconds

`func (o *AutoScalingRules) HasInitializationTimeoutSeconds() bool`

HasInitializationTimeoutSeconds returns a boolean if a field has been set.

### SetInitializationTimeoutSecondsNil

`func (o *AutoScalingRules) SetInitializationTimeoutSecondsNil(b bool)`

 SetInitializationTimeoutSecondsNil sets the value for InitializationTimeoutSeconds to be an explicit nil

### UnsetInitializationTimeoutSeconds
`func (o *AutoScalingRules) UnsetInitializationTimeoutSeconds()`

UnsetInitializationTimeoutSeconds ensures that no value is present for InitializationTimeoutSeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


