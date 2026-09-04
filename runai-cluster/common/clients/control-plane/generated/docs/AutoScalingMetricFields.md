# AutoScalingMetricFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metric** | Pointer to [**NullableAutoScalingMetric**](AutoScalingMetric.md) |  | [optional] 
**MetricThreshold** | Pointer to **NullableInt32** | The threshold to use with the specified metric for autoscaling. Mandatory if metric is specified | [optional] 

## Methods

### NewAutoScalingMetricFields

`func NewAutoScalingMetricFields() *AutoScalingMetricFields`

NewAutoScalingMetricFields instantiates a new AutoScalingMetricFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAutoScalingMetricFieldsWithDefaults

`func NewAutoScalingMetricFieldsWithDefaults() *AutoScalingMetricFields`

NewAutoScalingMetricFieldsWithDefaults instantiates a new AutoScalingMetricFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetric

`func (o *AutoScalingMetricFields) GetMetric() AutoScalingMetric`

GetMetric returns the Metric field if non-nil, zero value otherwise.

### GetMetricOk

`func (o *AutoScalingMetricFields) GetMetricOk() (*AutoScalingMetric, bool)`

GetMetricOk returns a tuple with the Metric field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetric

`func (o *AutoScalingMetricFields) SetMetric(v AutoScalingMetric)`

SetMetric sets Metric field to given value.

### HasMetric

`func (o *AutoScalingMetricFields) HasMetric() bool`

HasMetric returns a boolean if a field has been set.

### SetMetricNil

`func (o *AutoScalingMetricFields) SetMetricNil(b bool)`

 SetMetricNil sets the value for Metric to be an explicit nil

### UnsetMetric
`func (o *AutoScalingMetricFields) UnsetMetric()`

UnsetMetric ensures that no value is present for Metric, not even an explicit nil
### GetMetricThreshold

`func (o *AutoScalingMetricFields) GetMetricThreshold() int32`

GetMetricThreshold returns the MetricThreshold field if non-nil, zero value otherwise.

### GetMetricThresholdOk

`func (o *AutoScalingMetricFields) GetMetricThresholdOk() (*int32, bool)`

GetMetricThresholdOk returns a tuple with the MetricThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricThreshold

`func (o *AutoScalingMetricFields) SetMetricThreshold(v int32)`

SetMetricThreshold sets MetricThreshold field to given value.

### HasMetricThreshold

`func (o *AutoScalingMetricFields) HasMetricThreshold() bool`

HasMetricThreshold returns a boolean if a field has been set.

### SetMetricThresholdNil

`func (o *AutoScalingMetricFields) SetMetricThresholdNil(b bool)`

 SetMetricThresholdNil sets the value for MetricThreshold to be an explicit nil

### UnsetMetricThreshold
`func (o *AutoScalingMetricFields) UnsetMetricThreshold()`

UnsetMetricThreshold ensures that no value is present for MetricThreshold, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


