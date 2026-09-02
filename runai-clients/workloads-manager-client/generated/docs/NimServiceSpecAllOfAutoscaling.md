# NimServiceSpecAllOfAutoscaling

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MaxReplicas** | Pointer to **NullableInt32** | The maximum number of replicas for autoscaling. Defaults to minReplicas. Must be no less than minReplicas. | [optional] 
**Metric** | Pointer to **NullableString** | The metric to use for autoscaling (mandatory). | [optional] 
**MetricThreshold** | Pointer to **NullableInt32** | The threshold to use with the specified metric for autoscaling (mandatory). | [optional] 
**MinReplicas** | Pointer to **NullableInt32** | The minimum number of replicas for autoscaling. Defaults to 1. | [optional] 
**ScaleWindowSeconds** | Pointer to **NullableInt32** | The time window for autoscaling decisions, in seconds. Defaults to 300 seconds. | [optional] 

## Methods

### NewNimServiceSpecAllOfAutoscaling

`func NewNimServiceSpecAllOfAutoscaling() *NimServiceSpecAllOfAutoscaling`

NewNimServiceSpecAllOfAutoscaling instantiates a new NimServiceSpecAllOfAutoscaling object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServiceSpecAllOfAutoscalingWithDefaults

`func NewNimServiceSpecAllOfAutoscalingWithDefaults() *NimServiceSpecAllOfAutoscaling`

NewNimServiceSpecAllOfAutoscalingWithDefaults instantiates a new NimServiceSpecAllOfAutoscaling object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMaxReplicas

`func (o *NimServiceSpecAllOfAutoscaling) GetMaxReplicas() int32`

GetMaxReplicas returns the MaxReplicas field if non-nil, zero value otherwise.

### GetMaxReplicasOk

`func (o *NimServiceSpecAllOfAutoscaling) GetMaxReplicasOk() (*int32, bool)`

GetMaxReplicasOk returns a tuple with the MaxReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxReplicas

`func (o *NimServiceSpecAllOfAutoscaling) SetMaxReplicas(v int32)`

SetMaxReplicas sets MaxReplicas field to given value.

### HasMaxReplicas

`func (o *NimServiceSpecAllOfAutoscaling) HasMaxReplicas() bool`

HasMaxReplicas returns a boolean if a field has been set.

### SetMaxReplicasNil

`func (o *NimServiceSpecAllOfAutoscaling) SetMaxReplicasNil(b bool)`

 SetMaxReplicasNil sets the value for MaxReplicas to be an explicit nil

### UnsetMaxReplicas
`func (o *NimServiceSpecAllOfAutoscaling) UnsetMaxReplicas()`

UnsetMaxReplicas ensures that no value is present for MaxReplicas, not even an explicit nil
### GetMetric

`func (o *NimServiceSpecAllOfAutoscaling) GetMetric() string`

GetMetric returns the Metric field if non-nil, zero value otherwise.

### GetMetricOk

`func (o *NimServiceSpecAllOfAutoscaling) GetMetricOk() (*string, bool)`

GetMetricOk returns a tuple with the Metric field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetric

`func (o *NimServiceSpecAllOfAutoscaling) SetMetric(v string)`

SetMetric sets Metric field to given value.

### HasMetric

`func (o *NimServiceSpecAllOfAutoscaling) HasMetric() bool`

HasMetric returns a boolean if a field has been set.

### SetMetricNil

`func (o *NimServiceSpecAllOfAutoscaling) SetMetricNil(b bool)`

 SetMetricNil sets the value for Metric to be an explicit nil

### UnsetMetric
`func (o *NimServiceSpecAllOfAutoscaling) UnsetMetric()`

UnsetMetric ensures that no value is present for Metric, not even an explicit nil
### GetMetricThreshold

`func (o *NimServiceSpecAllOfAutoscaling) GetMetricThreshold() int32`

GetMetricThreshold returns the MetricThreshold field if non-nil, zero value otherwise.

### GetMetricThresholdOk

`func (o *NimServiceSpecAllOfAutoscaling) GetMetricThresholdOk() (*int32, bool)`

GetMetricThresholdOk returns a tuple with the MetricThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricThreshold

`func (o *NimServiceSpecAllOfAutoscaling) SetMetricThreshold(v int32)`

SetMetricThreshold sets MetricThreshold field to given value.

### HasMetricThreshold

`func (o *NimServiceSpecAllOfAutoscaling) HasMetricThreshold() bool`

HasMetricThreshold returns a boolean if a field has been set.

### SetMetricThresholdNil

`func (o *NimServiceSpecAllOfAutoscaling) SetMetricThresholdNil(b bool)`

 SetMetricThresholdNil sets the value for MetricThreshold to be an explicit nil

### UnsetMetricThreshold
`func (o *NimServiceSpecAllOfAutoscaling) UnsetMetricThreshold()`

UnsetMetricThreshold ensures that no value is present for MetricThreshold, not even an explicit nil
### GetMinReplicas

`func (o *NimServiceSpecAllOfAutoscaling) GetMinReplicas() int32`

GetMinReplicas returns the MinReplicas field if non-nil, zero value otherwise.

### GetMinReplicasOk

`func (o *NimServiceSpecAllOfAutoscaling) GetMinReplicasOk() (*int32, bool)`

GetMinReplicasOk returns a tuple with the MinReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinReplicas

`func (o *NimServiceSpecAllOfAutoscaling) SetMinReplicas(v int32)`

SetMinReplicas sets MinReplicas field to given value.

### HasMinReplicas

`func (o *NimServiceSpecAllOfAutoscaling) HasMinReplicas() bool`

HasMinReplicas returns a boolean if a field has been set.

### SetMinReplicasNil

`func (o *NimServiceSpecAllOfAutoscaling) SetMinReplicasNil(b bool)`

 SetMinReplicasNil sets the value for MinReplicas to be an explicit nil

### UnsetMinReplicas
`func (o *NimServiceSpecAllOfAutoscaling) UnsetMinReplicas()`

UnsetMinReplicas ensures that no value is present for MinReplicas, not even an explicit nil
### GetScaleWindowSeconds

`func (o *NimServiceSpecAllOfAutoscaling) GetScaleWindowSeconds() int32`

GetScaleWindowSeconds returns the ScaleWindowSeconds field if non-nil, zero value otherwise.

### GetScaleWindowSecondsOk

`func (o *NimServiceSpecAllOfAutoscaling) GetScaleWindowSecondsOk() (*int32, bool)`

GetScaleWindowSecondsOk returns a tuple with the ScaleWindowSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScaleWindowSeconds

`func (o *NimServiceSpecAllOfAutoscaling) SetScaleWindowSeconds(v int32)`

SetScaleWindowSeconds sets ScaleWindowSeconds field to given value.

### HasScaleWindowSeconds

`func (o *NimServiceSpecAllOfAutoscaling) HasScaleWindowSeconds() bool`

HasScaleWindowSeconds returns a boolean if a field has been set.

### SetScaleWindowSecondsNil

`func (o *NimServiceSpecAllOfAutoscaling) SetScaleWindowSecondsNil(b bool)`

 SetScaleWindowSecondsNil sets the value for ScaleWindowSeconds to be an explicit nil

### UnsetScaleWindowSeconds
`func (o *NimServiceSpecAllOfAutoscaling) UnsetScaleWindowSeconds()`

UnsetScaleWindowSeconds ensures that no value is present for ScaleWindowSeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


