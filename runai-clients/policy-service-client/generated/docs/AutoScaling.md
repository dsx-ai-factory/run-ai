# AutoScaling

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MetricThresholdPercentage** | Pointer to **NullableFloat32** | The percentage of metric threshold value to use for autoscaling. Defaults to 70. Applicable only with the &#39;throughput&#39; and &#39;concurrency&#39; metrics | [optional] 
**MinReplicas** | Pointer to **NullableInt32** | The minimum number of replicas for autoscaling. Defaults to 1. Use 0 to allow scale-to-zero | [optional] 
**MaxReplicas** | Pointer to **NullableInt32** | The maximum number of replicas for autoscaling. Defaults to minReplicas, or to 1 if minReplicas is set to 0 | [optional] 
**InitialReplicas** | Pointer to **NullableInt32** | The number of replicas to run when initializing the workload for the first time. Defaults to minReplicas, or to 1 if minReplicas is set to 0 | [optional] 
**ActivationReplicas** | Pointer to **NullableInt32** | The number of replicas to run when scaling-up from zero. Defaults to minReplicas, or to 1 if minReplicas is set to 0 | [optional] 
**ConcurrencyHardLimit** | Pointer to **NullableInt32** | The maximum number of requests allowed to flow to a single replica at any time. 0 means no limit | [optional] 
**ScaleToZeroRetentionSeconds** | Pointer to **NullableInt32** | The minimum amount of time (in seconds) that the last replica will remain active after a scale-to-zero decision. Defaults to 0. Available only if minReplicas is set to 0 | [optional] 
**ScaleDownDelaySeconds** | Pointer to **NullableInt32** | The minimum amount of time (in seconds) that a replica will remain active after a scale-down decision | [optional] 
**InitializationTimeoutSeconds** | Pointer to **NullableInt32** | Use &#x60;servingConfiguration.initializationTimeoutSeconds&#x60; instead.  If this field is set, it will be ignored and the value under &#x60;servingConfiguration&#x60; will be used. The maximum amount of time (in seconds) to wait for the container to become ready. | [optional] 
**Metric** | Pointer to **NullableString** | The metric to use for autoscaling. Mandatory if minReplicas &lt; maxReplicas, except for the special case where minReplicas is set to 0 and maxReplicas is set to 1, as in this case autoscaling decisions are made according to network activity rather than metrics. Use one of the built-in metrics of &#39;throughput&#39;, &#39;concurrency&#39; or &#39;latency&#39;, or any other available custom metric. Only the &#39;throughput&#39; and &#39;concurrency&#39; metrics support scale-to-zero | [optional] 
**MetricThreshold** | Pointer to **NullableInt32** | The threshold to use with the specified metric for autoscaling. Mandatory if metric is specified | [optional] 

## Methods

### NewAutoScaling

`func NewAutoScaling() *AutoScaling`

NewAutoScaling instantiates a new AutoScaling object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAutoScalingWithDefaults

`func NewAutoScalingWithDefaults() *AutoScaling`

NewAutoScalingWithDefaults instantiates a new AutoScaling object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetricThresholdPercentage

`func (o *AutoScaling) GetMetricThresholdPercentage() float32`

GetMetricThresholdPercentage returns the MetricThresholdPercentage field if non-nil, zero value otherwise.

### GetMetricThresholdPercentageOk

`func (o *AutoScaling) GetMetricThresholdPercentageOk() (*float32, bool)`

GetMetricThresholdPercentageOk returns a tuple with the MetricThresholdPercentage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricThresholdPercentage

`func (o *AutoScaling) SetMetricThresholdPercentage(v float32)`

SetMetricThresholdPercentage sets MetricThresholdPercentage field to given value.

### HasMetricThresholdPercentage

`func (o *AutoScaling) HasMetricThresholdPercentage() bool`

HasMetricThresholdPercentage returns a boolean if a field has been set.

### SetMetricThresholdPercentageNil

`func (o *AutoScaling) SetMetricThresholdPercentageNil(b bool)`

 SetMetricThresholdPercentageNil sets the value for MetricThresholdPercentage to be an explicit nil

### UnsetMetricThresholdPercentage
`func (o *AutoScaling) UnsetMetricThresholdPercentage()`

UnsetMetricThresholdPercentage ensures that no value is present for MetricThresholdPercentage, not even an explicit nil
### GetMinReplicas

`func (o *AutoScaling) GetMinReplicas() int32`

GetMinReplicas returns the MinReplicas field if non-nil, zero value otherwise.

### GetMinReplicasOk

`func (o *AutoScaling) GetMinReplicasOk() (*int32, bool)`

GetMinReplicasOk returns a tuple with the MinReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinReplicas

`func (o *AutoScaling) SetMinReplicas(v int32)`

SetMinReplicas sets MinReplicas field to given value.

### HasMinReplicas

`func (o *AutoScaling) HasMinReplicas() bool`

HasMinReplicas returns a boolean if a field has been set.

### SetMinReplicasNil

`func (o *AutoScaling) SetMinReplicasNil(b bool)`

 SetMinReplicasNil sets the value for MinReplicas to be an explicit nil

### UnsetMinReplicas
`func (o *AutoScaling) UnsetMinReplicas()`

UnsetMinReplicas ensures that no value is present for MinReplicas, not even an explicit nil
### GetMaxReplicas

`func (o *AutoScaling) GetMaxReplicas() int32`

GetMaxReplicas returns the MaxReplicas field if non-nil, zero value otherwise.

### GetMaxReplicasOk

`func (o *AutoScaling) GetMaxReplicasOk() (*int32, bool)`

GetMaxReplicasOk returns a tuple with the MaxReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxReplicas

`func (o *AutoScaling) SetMaxReplicas(v int32)`

SetMaxReplicas sets MaxReplicas field to given value.

### HasMaxReplicas

`func (o *AutoScaling) HasMaxReplicas() bool`

HasMaxReplicas returns a boolean if a field has been set.

### SetMaxReplicasNil

`func (o *AutoScaling) SetMaxReplicasNil(b bool)`

 SetMaxReplicasNil sets the value for MaxReplicas to be an explicit nil

### UnsetMaxReplicas
`func (o *AutoScaling) UnsetMaxReplicas()`

UnsetMaxReplicas ensures that no value is present for MaxReplicas, not even an explicit nil
### GetInitialReplicas

`func (o *AutoScaling) GetInitialReplicas() int32`

GetInitialReplicas returns the InitialReplicas field if non-nil, zero value otherwise.

### GetInitialReplicasOk

`func (o *AutoScaling) GetInitialReplicasOk() (*int32, bool)`

GetInitialReplicasOk returns a tuple with the InitialReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitialReplicas

`func (o *AutoScaling) SetInitialReplicas(v int32)`

SetInitialReplicas sets InitialReplicas field to given value.

### HasInitialReplicas

`func (o *AutoScaling) HasInitialReplicas() bool`

HasInitialReplicas returns a boolean if a field has been set.

### SetInitialReplicasNil

`func (o *AutoScaling) SetInitialReplicasNil(b bool)`

 SetInitialReplicasNil sets the value for InitialReplicas to be an explicit nil

### UnsetInitialReplicas
`func (o *AutoScaling) UnsetInitialReplicas()`

UnsetInitialReplicas ensures that no value is present for InitialReplicas, not even an explicit nil
### GetActivationReplicas

`func (o *AutoScaling) GetActivationReplicas() int32`

GetActivationReplicas returns the ActivationReplicas field if non-nil, zero value otherwise.

### GetActivationReplicasOk

`func (o *AutoScaling) GetActivationReplicasOk() (*int32, bool)`

GetActivationReplicasOk returns a tuple with the ActivationReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActivationReplicas

`func (o *AutoScaling) SetActivationReplicas(v int32)`

SetActivationReplicas sets ActivationReplicas field to given value.

### HasActivationReplicas

`func (o *AutoScaling) HasActivationReplicas() bool`

HasActivationReplicas returns a boolean if a field has been set.

### SetActivationReplicasNil

`func (o *AutoScaling) SetActivationReplicasNil(b bool)`

 SetActivationReplicasNil sets the value for ActivationReplicas to be an explicit nil

### UnsetActivationReplicas
`func (o *AutoScaling) UnsetActivationReplicas()`

UnsetActivationReplicas ensures that no value is present for ActivationReplicas, not even an explicit nil
### GetConcurrencyHardLimit

`func (o *AutoScaling) GetConcurrencyHardLimit() int32`

GetConcurrencyHardLimit returns the ConcurrencyHardLimit field if non-nil, zero value otherwise.

### GetConcurrencyHardLimitOk

`func (o *AutoScaling) GetConcurrencyHardLimitOk() (*int32, bool)`

GetConcurrencyHardLimitOk returns a tuple with the ConcurrencyHardLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConcurrencyHardLimit

`func (o *AutoScaling) SetConcurrencyHardLimit(v int32)`

SetConcurrencyHardLimit sets ConcurrencyHardLimit field to given value.

### HasConcurrencyHardLimit

`func (o *AutoScaling) HasConcurrencyHardLimit() bool`

HasConcurrencyHardLimit returns a boolean if a field has been set.

### SetConcurrencyHardLimitNil

`func (o *AutoScaling) SetConcurrencyHardLimitNil(b bool)`

 SetConcurrencyHardLimitNil sets the value for ConcurrencyHardLimit to be an explicit nil

### UnsetConcurrencyHardLimit
`func (o *AutoScaling) UnsetConcurrencyHardLimit()`

UnsetConcurrencyHardLimit ensures that no value is present for ConcurrencyHardLimit, not even an explicit nil
### GetScaleToZeroRetentionSeconds

`func (o *AutoScaling) GetScaleToZeroRetentionSeconds() int32`

GetScaleToZeroRetentionSeconds returns the ScaleToZeroRetentionSeconds field if non-nil, zero value otherwise.

### GetScaleToZeroRetentionSecondsOk

`func (o *AutoScaling) GetScaleToZeroRetentionSecondsOk() (*int32, bool)`

GetScaleToZeroRetentionSecondsOk returns a tuple with the ScaleToZeroRetentionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScaleToZeroRetentionSeconds

`func (o *AutoScaling) SetScaleToZeroRetentionSeconds(v int32)`

SetScaleToZeroRetentionSeconds sets ScaleToZeroRetentionSeconds field to given value.

### HasScaleToZeroRetentionSeconds

`func (o *AutoScaling) HasScaleToZeroRetentionSeconds() bool`

HasScaleToZeroRetentionSeconds returns a boolean if a field has been set.

### SetScaleToZeroRetentionSecondsNil

`func (o *AutoScaling) SetScaleToZeroRetentionSecondsNil(b bool)`

 SetScaleToZeroRetentionSecondsNil sets the value for ScaleToZeroRetentionSeconds to be an explicit nil

### UnsetScaleToZeroRetentionSeconds
`func (o *AutoScaling) UnsetScaleToZeroRetentionSeconds()`

UnsetScaleToZeroRetentionSeconds ensures that no value is present for ScaleToZeroRetentionSeconds, not even an explicit nil
### GetScaleDownDelaySeconds

`func (o *AutoScaling) GetScaleDownDelaySeconds() int32`

GetScaleDownDelaySeconds returns the ScaleDownDelaySeconds field if non-nil, zero value otherwise.

### GetScaleDownDelaySecondsOk

`func (o *AutoScaling) GetScaleDownDelaySecondsOk() (*int32, bool)`

GetScaleDownDelaySecondsOk returns a tuple with the ScaleDownDelaySeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScaleDownDelaySeconds

`func (o *AutoScaling) SetScaleDownDelaySeconds(v int32)`

SetScaleDownDelaySeconds sets ScaleDownDelaySeconds field to given value.

### HasScaleDownDelaySeconds

`func (o *AutoScaling) HasScaleDownDelaySeconds() bool`

HasScaleDownDelaySeconds returns a boolean if a field has been set.

### SetScaleDownDelaySecondsNil

`func (o *AutoScaling) SetScaleDownDelaySecondsNil(b bool)`

 SetScaleDownDelaySecondsNil sets the value for ScaleDownDelaySeconds to be an explicit nil

### UnsetScaleDownDelaySeconds
`func (o *AutoScaling) UnsetScaleDownDelaySeconds()`

UnsetScaleDownDelaySeconds ensures that no value is present for ScaleDownDelaySeconds, not even an explicit nil
### GetInitializationTimeoutSeconds

`func (o *AutoScaling) GetInitializationTimeoutSeconds() int32`

GetInitializationTimeoutSeconds returns the InitializationTimeoutSeconds field if non-nil, zero value otherwise.

### GetInitializationTimeoutSecondsOk

`func (o *AutoScaling) GetInitializationTimeoutSecondsOk() (*int32, bool)`

GetInitializationTimeoutSecondsOk returns a tuple with the InitializationTimeoutSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitializationTimeoutSeconds

`func (o *AutoScaling) SetInitializationTimeoutSeconds(v int32)`

SetInitializationTimeoutSeconds sets InitializationTimeoutSeconds field to given value.

### HasInitializationTimeoutSeconds

`func (o *AutoScaling) HasInitializationTimeoutSeconds() bool`

HasInitializationTimeoutSeconds returns a boolean if a field has been set.

### SetInitializationTimeoutSecondsNil

`func (o *AutoScaling) SetInitializationTimeoutSecondsNil(b bool)`

 SetInitializationTimeoutSecondsNil sets the value for InitializationTimeoutSeconds to be an explicit nil

### UnsetInitializationTimeoutSeconds
`func (o *AutoScaling) UnsetInitializationTimeoutSeconds()`

UnsetInitializationTimeoutSeconds ensures that no value is present for InitializationTimeoutSeconds, not even an explicit nil
### GetMetric

`func (o *AutoScaling) GetMetric() string`

GetMetric returns the Metric field if non-nil, zero value otherwise.

### GetMetricOk

`func (o *AutoScaling) GetMetricOk() (*string, bool)`

GetMetricOk returns a tuple with the Metric field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetric

`func (o *AutoScaling) SetMetric(v string)`

SetMetric sets Metric field to given value.

### HasMetric

`func (o *AutoScaling) HasMetric() bool`

HasMetric returns a boolean if a field has been set.

### SetMetricNil

`func (o *AutoScaling) SetMetricNil(b bool)`

 SetMetricNil sets the value for Metric to be an explicit nil

### UnsetMetric
`func (o *AutoScaling) UnsetMetric()`

UnsetMetric ensures that no value is present for Metric, not even an explicit nil
### GetMetricThreshold

`func (o *AutoScaling) GetMetricThreshold() int32`

GetMetricThreshold returns the MetricThreshold field if non-nil, zero value otherwise.

### GetMetricThresholdOk

`func (o *AutoScaling) GetMetricThresholdOk() (*int32, bool)`

GetMetricThresholdOk returns a tuple with the MetricThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricThreshold

`func (o *AutoScaling) SetMetricThreshold(v int32)`

SetMetricThreshold sets MetricThreshold field to given value.

### HasMetricThreshold

`func (o *AutoScaling) HasMetricThreshold() bool`

HasMetricThreshold returns a boolean if a field has been set.

### SetMetricThresholdNil

`func (o *AutoScaling) SetMetricThresholdNil(b bool)`

 SetMetricThresholdNil sets the value for MetricThreshold to be an explicit nil

### UnsetMetricThreshold
`func (o *AutoScaling) UnsetMetricThreshold()`

UnsetMetricThreshold ensures that no value is present for MetricThreshold, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


