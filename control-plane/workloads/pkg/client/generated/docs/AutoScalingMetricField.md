# AutoScalingMetricField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metric** | Pointer to **NullableString** | The metric to use for autoscaling. Mandatory if minReplicas &lt; maxReplicas, except for the special case where minReplicas is set to 0 and maxReplicas is set to 1, as in this case autoscaling decisions are made according to network activity rather than metrics. Use one of the built-in metrics of &#39;throughput&#39;, &#39;concurrency&#39; or &#39;latency&#39;, or any other available custom metric. Only the &#39;throughput&#39; and &#39;concurrency&#39; metrics support scale-to-zero | [optional] 

## Methods

### NewAutoScalingMetricField

`func NewAutoScalingMetricField() *AutoScalingMetricField`

NewAutoScalingMetricField instantiates a new AutoScalingMetricField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAutoScalingMetricFieldWithDefaults

`func NewAutoScalingMetricFieldWithDefaults() *AutoScalingMetricField`

NewAutoScalingMetricFieldWithDefaults instantiates a new AutoScalingMetricField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetric

`func (o *AutoScalingMetricField) GetMetric() string`

GetMetric returns the Metric field if non-nil, zero value otherwise.

### GetMetricOk

`func (o *AutoScalingMetricField) GetMetricOk() (*string, bool)`

GetMetricOk returns a tuple with the Metric field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetric

`func (o *AutoScalingMetricField) SetMetric(v string)`

SetMetric sets Metric field to given value.

### HasMetric

`func (o *AutoScalingMetricField) HasMetric() bool`

HasMetric returns a boolean if a field has been set.

### SetMetricNil

`func (o *AutoScalingMetricField) SetMetricNil(b bool)`

 SetMetricNil sets the value for Metric to be an explicit nil

### UnsetMetric
`func (o *AutoScalingMetricField) UnsetMetric()`

UnsetMetric ensures that no value is present for Metric, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


