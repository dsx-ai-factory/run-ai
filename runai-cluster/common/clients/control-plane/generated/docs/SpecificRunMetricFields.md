# SpecificRunMetricFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ThresholdMetric** | Pointer to [**NullableAutoScalingMetric**](AutoScalingMetric.md) |  | [optional] 
**ThresholdValue** | Pointer to **NullableInt32** | The threshold value to use with the specified metric for autoscaling | [optional] 

## Methods

### NewSpecificRunMetricFields

`func NewSpecificRunMetricFields() *SpecificRunMetricFields`

NewSpecificRunMetricFields instantiates a new SpecificRunMetricFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSpecificRunMetricFieldsWithDefaults

`func NewSpecificRunMetricFieldsWithDefaults() *SpecificRunMetricFields`

NewSpecificRunMetricFieldsWithDefaults instantiates a new SpecificRunMetricFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetThresholdMetric

`func (o *SpecificRunMetricFields) GetThresholdMetric() AutoScalingMetric`

GetThresholdMetric returns the ThresholdMetric field if non-nil, zero value otherwise.

### GetThresholdMetricOk

`func (o *SpecificRunMetricFields) GetThresholdMetricOk() (*AutoScalingMetric, bool)`

GetThresholdMetricOk returns a tuple with the ThresholdMetric field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThresholdMetric

`func (o *SpecificRunMetricFields) SetThresholdMetric(v AutoScalingMetric)`

SetThresholdMetric sets ThresholdMetric field to given value.

### HasThresholdMetric

`func (o *SpecificRunMetricFields) HasThresholdMetric() bool`

HasThresholdMetric returns a boolean if a field has been set.

### SetThresholdMetricNil

`func (o *SpecificRunMetricFields) SetThresholdMetricNil(b bool)`

 SetThresholdMetricNil sets the value for ThresholdMetric to be an explicit nil

### UnsetThresholdMetric
`func (o *SpecificRunMetricFields) UnsetThresholdMetric()`

UnsetThresholdMetric ensures that no value is present for ThresholdMetric, not even an explicit nil
### GetThresholdValue

`func (o *SpecificRunMetricFields) GetThresholdValue() int32`

GetThresholdValue returns the ThresholdValue field if non-nil, zero value otherwise.

### GetThresholdValueOk

`func (o *SpecificRunMetricFields) GetThresholdValueOk() (*int32, bool)`

GetThresholdValueOk returns a tuple with the ThresholdValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThresholdValue

`func (o *SpecificRunMetricFields) SetThresholdValue(v int32)`

SetThresholdValue sets ThresholdValue field to given value.

### HasThresholdValue

`func (o *SpecificRunMetricFields) HasThresholdValue() bool`

HasThresholdValue returns a boolean if a field has been set.

### SetThresholdValueNil

`func (o *SpecificRunMetricFields) SetThresholdValueNil(b bool)`

 SetThresholdValueNil sets the value for ThresholdValue to be an explicit nil

### UnsetThresholdValue
`func (o *SpecificRunMetricFields) UnsetThresholdValue()`

UnsetThresholdValue ensures that no value is present for ThresholdValue, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


