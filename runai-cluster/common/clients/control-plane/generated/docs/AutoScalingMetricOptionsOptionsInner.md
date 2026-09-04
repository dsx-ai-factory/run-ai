# AutoScalingMetricOptionsOptionsInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Value** | [**NullableAutoScalingMetric**](AutoScalingMetric.md) |  | 
**Displayed** | Pointer to **NullableString** | Textual description of the value | [optional] 

## Methods

### NewAutoScalingMetricOptionsOptionsInner

`func NewAutoScalingMetricOptionsOptionsInner(value NullableAutoScalingMetric, ) *AutoScalingMetricOptionsOptionsInner`

NewAutoScalingMetricOptionsOptionsInner instantiates a new AutoScalingMetricOptionsOptionsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAutoScalingMetricOptionsOptionsInnerWithDefaults

`func NewAutoScalingMetricOptionsOptionsInnerWithDefaults() *AutoScalingMetricOptionsOptionsInner`

NewAutoScalingMetricOptionsOptionsInnerWithDefaults instantiates a new AutoScalingMetricOptionsOptionsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetValue

`func (o *AutoScalingMetricOptionsOptionsInner) GetValue() AutoScalingMetric`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *AutoScalingMetricOptionsOptionsInner) GetValueOk() (*AutoScalingMetric, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *AutoScalingMetricOptionsOptionsInner) SetValue(v AutoScalingMetric)`

SetValue sets Value field to given value.


### SetValueNil

`func (o *AutoScalingMetricOptionsOptionsInner) SetValueNil(b bool)`

 SetValueNil sets the value for Value to be an explicit nil

### UnsetValue
`func (o *AutoScalingMetricOptionsOptionsInner) UnsetValue()`

UnsetValue ensures that no value is present for Value, not even an explicit nil
### GetDisplayed

`func (o *AutoScalingMetricOptionsOptionsInner) GetDisplayed() string`

GetDisplayed returns the Displayed field if non-nil, zero value otherwise.

### GetDisplayedOk

`func (o *AutoScalingMetricOptionsOptionsInner) GetDisplayedOk() (*string, bool)`

GetDisplayedOk returns a tuple with the Displayed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayed

`func (o *AutoScalingMetricOptionsOptionsInner) SetDisplayed(v string)`

SetDisplayed sets Displayed field to given value.

### HasDisplayed

`func (o *AutoScalingMetricOptionsOptionsInner) HasDisplayed() bool`

HasDisplayed returns a boolean if a field has been set.

### SetDisplayedNil

`func (o *AutoScalingMetricOptionsOptionsInner) SetDisplayedNil(b bool)`

 SetDisplayedNil sets the value for Displayed to be an explicit nil

### UnsetDisplayed
`func (o *AutoScalingMetricOptionsOptionsInner) UnsetDisplayed()`

UnsetDisplayed ensures that no value is present for Displayed, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


