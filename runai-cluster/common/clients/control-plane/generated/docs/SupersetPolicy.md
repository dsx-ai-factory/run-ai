# SupersetPolicy

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Defaults** | Pointer to [**NullableSupersetPolicyDefaults**](SupersetPolicyDefaults.md) |  | [optional] 
**Rules** | Pointer to [**NullableSupersetPolicyRules**](SupersetPolicyRules.md) |  | [optional] 

## Methods

### NewSupersetPolicy

`func NewSupersetPolicy() *SupersetPolicy`

NewSupersetPolicy instantiates a new SupersetPolicy object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSupersetPolicyWithDefaults

`func NewSupersetPolicyWithDefaults() *SupersetPolicy`

NewSupersetPolicyWithDefaults instantiates a new SupersetPolicy object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefaults

`func (o *SupersetPolicy) GetDefaults() SupersetPolicyDefaults`

GetDefaults returns the Defaults field if non-nil, zero value otherwise.

### GetDefaultsOk

`func (o *SupersetPolicy) GetDefaultsOk() (*SupersetPolicyDefaults, bool)`

GetDefaultsOk returns a tuple with the Defaults field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaults

`func (o *SupersetPolicy) SetDefaults(v SupersetPolicyDefaults)`

SetDefaults sets Defaults field to given value.

### HasDefaults

`func (o *SupersetPolicy) HasDefaults() bool`

HasDefaults returns a boolean if a field has been set.

### SetDefaultsNil

`func (o *SupersetPolicy) SetDefaultsNil(b bool)`

 SetDefaultsNil sets the value for Defaults to be an explicit nil

### UnsetDefaults
`func (o *SupersetPolicy) UnsetDefaults()`

UnsetDefaults ensures that no value is present for Defaults, not even an explicit nil
### GetRules

`func (o *SupersetPolicy) GetRules() SupersetPolicyRules`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *SupersetPolicy) GetRulesOk() (*SupersetPolicyRules, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *SupersetPolicy) SetRules(v SupersetPolicyRules)`

SetRules sets Rules field to given value.

### HasRules

`func (o *SupersetPolicy) HasRules() bool`

HasRules returns a boolean if a field has been set.

### SetRulesNil

`func (o *SupersetPolicy) SetRulesNil(b bool)`

 SetRulesNil sets the value for Rules to be an explicit nil

### UnsetRules
`func (o *SupersetPolicy) UnsetRules()`

UnsetRules ensures that no value is present for Rules, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


