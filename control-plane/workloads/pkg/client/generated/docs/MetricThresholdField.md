# MetricThresholdField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MetricThreshold** | Pointer to **NullableInt32** | The threshold to use with the specified metric for autoscaling. Mandatory if metric is specified | [optional] 

## Methods

### NewMetricThresholdField

`func NewMetricThresholdField() *MetricThresholdField`

NewMetricThresholdField instantiates a new MetricThresholdField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetricThresholdFieldWithDefaults

`func NewMetricThresholdFieldWithDefaults() *MetricThresholdField`

NewMetricThresholdFieldWithDefaults instantiates a new MetricThresholdField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetricThreshold

`func (o *MetricThresholdField) GetMetricThreshold() int32`

GetMetricThreshold returns the MetricThreshold field if non-nil, zero value otherwise.

### GetMetricThresholdOk

`func (o *MetricThresholdField) GetMetricThresholdOk() (*int32, bool)`

GetMetricThresholdOk returns a tuple with the MetricThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricThreshold

`func (o *MetricThresholdField) SetMetricThreshold(v int32)`

SetMetricThreshold sets MetricThreshold field to given value.

### HasMetricThreshold

`func (o *MetricThresholdField) HasMetricThreshold() bool`

HasMetricThreshold returns a boolean if a field has been set.

### SetMetricThresholdNil

`func (o *MetricThresholdField) SetMetricThresholdNil(b bool)`

 SetMetricThresholdNil sets the value for MetricThreshold to be an explicit nil

### UnsetMetricThreshold
`func (o *MetricThresholdField) UnsetMetricThreshold()`

UnsetMetricThreshold ensures that no value is present for MetricThreshold, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


