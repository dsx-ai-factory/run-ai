# MeasurementResponseValuesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Value** | **string** |  | 
**Timestamp** | **NullableTime** |  | 

## Methods

### NewMeasurementResponseValuesInner

`func NewMeasurementResponseValuesInner(value string, timestamp NullableTime, ) *MeasurementResponseValuesInner`

NewMeasurementResponseValuesInner instantiates a new MeasurementResponseValuesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMeasurementResponseValuesInnerWithDefaults

`func NewMeasurementResponseValuesInnerWithDefaults() *MeasurementResponseValuesInner`

NewMeasurementResponseValuesInnerWithDefaults instantiates a new MeasurementResponseValuesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetValue

`func (o *MeasurementResponseValuesInner) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *MeasurementResponseValuesInner) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *MeasurementResponseValuesInner) SetValue(v string)`

SetValue sets Value field to given value.


### GetTimestamp

`func (o *MeasurementResponseValuesInner) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *MeasurementResponseValuesInner) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *MeasurementResponseValuesInner) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.


### SetTimestampNil

`func (o *MeasurementResponseValuesInner) SetTimestampNil(b bool)`

 SetTimestampNil sets the value for Timestamp to be an explicit nil

### UnsetTimestamp
`func (o *MeasurementResponseValuesInner) UnsetTimestamp()`

UnsetTimestamp ensures that no value is present for Timestamp, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


