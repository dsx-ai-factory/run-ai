# DefaultFromRule

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Field** | Pointer to **NullableString** | The default of this field is taken from another field, with optional factor. The specified field must belong to the same workload section (compute, security, etc). For itemized field, the value of the field will be taken from the same instance. | [optional] 
**Factor** | Pointer to **NullableFloat64** | Multiple the value of the fromField by this factor to get the value for this field For example, if factor is 2 and defaultFrom field has value of 8, the calculated default will be determined to be 16. | [optional] [default to 1]

## Methods

### NewDefaultFromRule

`func NewDefaultFromRule() *DefaultFromRule`

NewDefaultFromRule instantiates a new DefaultFromRule object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDefaultFromRuleWithDefaults

`func NewDefaultFromRuleWithDefaults() *DefaultFromRule`

NewDefaultFromRuleWithDefaults instantiates a new DefaultFromRule object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetField

`func (o *DefaultFromRule) GetField() string`

GetField returns the Field field if non-nil, zero value otherwise.

### GetFieldOk

`func (o *DefaultFromRule) GetFieldOk() (*string, bool)`

GetFieldOk returns a tuple with the Field field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetField

`func (o *DefaultFromRule) SetField(v string)`

SetField sets Field field to given value.

### HasField

`func (o *DefaultFromRule) HasField() bool`

HasField returns a boolean if a field has been set.

### SetFieldNil

`func (o *DefaultFromRule) SetFieldNil(b bool)`

 SetFieldNil sets the value for Field to be an explicit nil

### UnsetField
`func (o *DefaultFromRule) UnsetField()`

UnsetField ensures that no value is present for Field, not even an explicit nil
### GetFactor

`func (o *DefaultFromRule) GetFactor() float64`

GetFactor returns the Factor field if non-nil, zero value otherwise.

### GetFactorOk

`func (o *DefaultFromRule) GetFactorOk() (*float64, bool)`

GetFactorOk returns a tuple with the Factor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFactor

`func (o *DefaultFromRule) SetFactor(v float64)`

SetFactor sets Factor field to given value.

### HasFactor

`func (o *DefaultFromRule) HasFactor() bool`

HasFactor returns a boolean if a field has been set.

### SetFactorNil

`func (o *DefaultFromRule) SetFactorNil(b bool)`

 SetFactorNil sets the value for Factor to be an explicit nil

### UnsetFactor
`func (o *DefaultFromRule) UnsetFactor()`

UnsetFactor ensures that no value is present for Factor, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


