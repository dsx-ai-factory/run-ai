# NimServiceRulesAllOfAutoscaling

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MaxReplicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Metric** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**MetricThreshold** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**MinReplicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**ScaleWindowSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 

## Methods

### NewNimServiceRulesAllOfAutoscaling

`func NewNimServiceRulesAllOfAutoscaling() *NimServiceRulesAllOfAutoscaling`

NewNimServiceRulesAllOfAutoscaling instantiates a new NimServiceRulesAllOfAutoscaling object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServiceRulesAllOfAutoscalingWithDefaults

`func NewNimServiceRulesAllOfAutoscalingWithDefaults() *NimServiceRulesAllOfAutoscaling`

NewNimServiceRulesAllOfAutoscalingWithDefaults instantiates a new NimServiceRulesAllOfAutoscaling object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMaxReplicas

`func (o *NimServiceRulesAllOfAutoscaling) GetMaxReplicas() IntegerRules`

GetMaxReplicas returns the MaxReplicas field if non-nil, zero value otherwise.

### GetMaxReplicasOk

`func (o *NimServiceRulesAllOfAutoscaling) GetMaxReplicasOk() (*IntegerRules, bool)`

GetMaxReplicasOk returns a tuple with the MaxReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxReplicas

`func (o *NimServiceRulesAllOfAutoscaling) SetMaxReplicas(v IntegerRules)`

SetMaxReplicas sets MaxReplicas field to given value.

### HasMaxReplicas

`func (o *NimServiceRulesAllOfAutoscaling) HasMaxReplicas() bool`

HasMaxReplicas returns a boolean if a field has been set.

### SetMaxReplicasNil

`func (o *NimServiceRulesAllOfAutoscaling) SetMaxReplicasNil(b bool)`

 SetMaxReplicasNil sets the value for MaxReplicas to be an explicit nil

### UnsetMaxReplicas
`func (o *NimServiceRulesAllOfAutoscaling) UnsetMaxReplicas()`

UnsetMaxReplicas ensures that no value is present for MaxReplicas, not even an explicit nil
### GetMetric

`func (o *NimServiceRulesAllOfAutoscaling) GetMetric() StringRules`

GetMetric returns the Metric field if non-nil, zero value otherwise.

### GetMetricOk

`func (o *NimServiceRulesAllOfAutoscaling) GetMetricOk() (*StringRules, bool)`

GetMetricOk returns a tuple with the Metric field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetric

`func (o *NimServiceRulesAllOfAutoscaling) SetMetric(v StringRules)`

SetMetric sets Metric field to given value.

### HasMetric

`func (o *NimServiceRulesAllOfAutoscaling) HasMetric() bool`

HasMetric returns a boolean if a field has been set.

### SetMetricNil

`func (o *NimServiceRulesAllOfAutoscaling) SetMetricNil(b bool)`

 SetMetricNil sets the value for Metric to be an explicit nil

### UnsetMetric
`func (o *NimServiceRulesAllOfAutoscaling) UnsetMetric()`

UnsetMetric ensures that no value is present for Metric, not even an explicit nil
### GetMetricThreshold

`func (o *NimServiceRulesAllOfAutoscaling) GetMetricThreshold() IntegerRules`

GetMetricThreshold returns the MetricThreshold field if non-nil, zero value otherwise.

### GetMetricThresholdOk

`func (o *NimServiceRulesAllOfAutoscaling) GetMetricThresholdOk() (*IntegerRules, bool)`

GetMetricThresholdOk returns a tuple with the MetricThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricThreshold

`func (o *NimServiceRulesAllOfAutoscaling) SetMetricThreshold(v IntegerRules)`

SetMetricThreshold sets MetricThreshold field to given value.

### HasMetricThreshold

`func (o *NimServiceRulesAllOfAutoscaling) HasMetricThreshold() bool`

HasMetricThreshold returns a boolean if a field has been set.

### SetMetricThresholdNil

`func (o *NimServiceRulesAllOfAutoscaling) SetMetricThresholdNil(b bool)`

 SetMetricThresholdNil sets the value for MetricThreshold to be an explicit nil

### UnsetMetricThreshold
`func (o *NimServiceRulesAllOfAutoscaling) UnsetMetricThreshold()`

UnsetMetricThreshold ensures that no value is present for MetricThreshold, not even an explicit nil
### GetMinReplicas

`func (o *NimServiceRulesAllOfAutoscaling) GetMinReplicas() IntegerRules`

GetMinReplicas returns the MinReplicas field if non-nil, zero value otherwise.

### GetMinReplicasOk

`func (o *NimServiceRulesAllOfAutoscaling) GetMinReplicasOk() (*IntegerRules, bool)`

GetMinReplicasOk returns a tuple with the MinReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinReplicas

`func (o *NimServiceRulesAllOfAutoscaling) SetMinReplicas(v IntegerRules)`

SetMinReplicas sets MinReplicas field to given value.

### HasMinReplicas

`func (o *NimServiceRulesAllOfAutoscaling) HasMinReplicas() bool`

HasMinReplicas returns a boolean if a field has been set.

### SetMinReplicasNil

`func (o *NimServiceRulesAllOfAutoscaling) SetMinReplicasNil(b bool)`

 SetMinReplicasNil sets the value for MinReplicas to be an explicit nil

### UnsetMinReplicas
`func (o *NimServiceRulesAllOfAutoscaling) UnsetMinReplicas()`

UnsetMinReplicas ensures that no value is present for MinReplicas, not even an explicit nil
### GetScaleWindowSeconds

`func (o *NimServiceRulesAllOfAutoscaling) GetScaleWindowSeconds() IntegerRules`

GetScaleWindowSeconds returns the ScaleWindowSeconds field if non-nil, zero value otherwise.

### GetScaleWindowSecondsOk

`func (o *NimServiceRulesAllOfAutoscaling) GetScaleWindowSecondsOk() (*IntegerRules, bool)`

GetScaleWindowSecondsOk returns a tuple with the ScaleWindowSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScaleWindowSeconds

`func (o *NimServiceRulesAllOfAutoscaling) SetScaleWindowSeconds(v IntegerRules)`

SetScaleWindowSeconds sets ScaleWindowSeconds field to given value.

### HasScaleWindowSeconds

`func (o *NimServiceRulesAllOfAutoscaling) HasScaleWindowSeconds() bool`

HasScaleWindowSeconds returns a boolean if a field has been set.

### SetScaleWindowSecondsNil

`func (o *NimServiceRulesAllOfAutoscaling) SetScaleWindowSecondsNil(b bool)`

 SetScaleWindowSecondsNil sets the value for ScaleWindowSeconds to be an explicit nil

### UnsetScaleWindowSeconds
`func (o *NimServiceRulesAllOfAutoscaling) UnsetScaleWindowSeconds()`

UnsetScaleWindowSeconds ensures that no value is present for ScaleWindowSeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


