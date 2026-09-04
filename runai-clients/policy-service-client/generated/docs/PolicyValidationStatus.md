# PolicyValidationStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Validation** | Pointer to [**NullablePolicyValidationStatusValidation**](PolicyValidationStatusValidation.md) |  | [optional] 

## Methods

### NewPolicyValidationStatus

`func NewPolicyValidationStatus() *PolicyValidationStatus`

NewPolicyValidationStatus instantiates a new PolicyValidationStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyValidationStatusWithDefaults

`func NewPolicyValidationStatusWithDefaults() *PolicyValidationStatus`

NewPolicyValidationStatusWithDefaults instantiates a new PolicyValidationStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetValidation

`func (o *PolicyValidationStatus) GetValidation() PolicyValidationStatusValidation`

GetValidation returns the Validation field if non-nil, zero value otherwise.

### GetValidationOk

`func (o *PolicyValidationStatus) GetValidationOk() (*PolicyValidationStatusValidation, bool)`

GetValidationOk returns a tuple with the Validation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValidation

`func (o *PolicyValidationStatus) SetValidation(v PolicyValidationStatusValidation)`

SetValidation sets Validation field to given value.

### HasValidation

`func (o *PolicyValidationStatus) HasValidation() bool`

HasValidation returns a boolean if a field has been set.

### SetValidationNil

`func (o *PolicyValidationStatus) SetValidationNil(b bool)`

 SetValidationNil sets the value for Validation to be an explicit nil

### UnsetValidation
`func (o *PolicyValidationStatus) UnsetValidation()`

UnsetValidation ensures that no value is present for Validation, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


