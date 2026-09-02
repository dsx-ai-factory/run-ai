# PolicyEnforcementRangeRange

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Min** | Pointer to **NullableFloat32** | Enforces the minimum numeric value of the field. | [optional] 
**Max** | Pointer to **NullableFloat32** | Enforces the maximum numeric value of the field. | [optional] 
**Step** | Pointer to **NullableFloat32** | Restricts the field value to specific numeric increments starting from &#x60;min&#x60; (or from 0 when &#x60;min&#x60; is absent). Must be greater than 0.  | [optional] 

## Methods

### NewPolicyEnforcementRangeRange

`func NewPolicyEnforcementRangeRange() *PolicyEnforcementRangeRange`

NewPolicyEnforcementRangeRange instantiates a new PolicyEnforcementRangeRange object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyEnforcementRangeRangeWithDefaults

`func NewPolicyEnforcementRangeRangeWithDefaults() *PolicyEnforcementRangeRange`

NewPolicyEnforcementRangeRangeWithDefaults instantiates a new PolicyEnforcementRangeRange object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMin

`func (o *PolicyEnforcementRangeRange) GetMin() float32`

GetMin returns the Min field if non-nil, zero value otherwise.

### GetMinOk

`func (o *PolicyEnforcementRangeRange) GetMinOk() (*float32, bool)`

GetMinOk returns a tuple with the Min field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMin

`func (o *PolicyEnforcementRangeRange) SetMin(v float32)`

SetMin sets Min field to given value.

### HasMin

`func (o *PolicyEnforcementRangeRange) HasMin() bool`

HasMin returns a boolean if a field has been set.

### SetMinNil

`func (o *PolicyEnforcementRangeRange) SetMinNil(b bool)`

 SetMinNil sets the value for Min to be an explicit nil

### UnsetMin
`func (o *PolicyEnforcementRangeRange) UnsetMin()`

UnsetMin ensures that no value is present for Min, not even an explicit nil
### GetMax

`func (o *PolicyEnforcementRangeRange) GetMax() float32`

GetMax returns the Max field if non-nil, zero value otherwise.

### GetMaxOk

`func (o *PolicyEnforcementRangeRange) GetMaxOk() (*float32, bool)`

GetMaxOk returns a tuple with the Max field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMax

`func (o *PolicyEnforcementRangeRange) SetMax(v float32)`

SetMax sets Max field to given value.

### HasMax

`func (o *PolicyEnforcementRangeRange) HasMax() bool`

HasMax returns a boolean if a field has been set.

### SetMaxNil

`func (o *PolicyEnforcementRangeRange) SetMaxNil(b bool)`

 SetMaxNil sets the value for Max to be an explicit nil

### UnsetMax
`func (o *PolicyEnforcementRangeRange) UnsetMax()`

UnsetMax ensures that no value is present for Max, not even an explicit nil
### GetStep

`func (o *PolicyEnforcementRangeRange) GetStep() float32`

GetStep returns the Step field if non-nil, zero value otherwise.

### GetStepOk

`func (o *PolicyEnforcementRangeRange) GetStepOk() (*float32, bool)`

GetStepOk returns a tuple with the Step field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStep

`func (o *PolicyEnforcementRangeRange) SetStep(v float32)`

SetStep sets Step field to given value.

### HasStep

`func (o *PolicyEnforcementRangeRange) HasStep() bool`

HasStep returns a boolean if a field has been set.

### SetStepNil

`func (o *PolicyEnforcementRangeRange) SetStepNil(b bool)`

 SetStepNil sets the value for Step to be an explicit nil

### UnsetStep
`func (o *PolicyEnforcementRangeRange) UnsetStep()`

UnsetStep ensures that no value is present for Step, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


