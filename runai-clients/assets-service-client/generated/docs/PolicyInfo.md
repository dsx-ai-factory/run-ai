# PolicyInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AttributeDefaults** | Pointer to [**NullableAttributeDefaults**](AttributeDefaults.md) |  | [optional] 
**Rules** | Pointer to [**NullableRules**](Rules.md) |  | [optional] 

## Methods

### NewPolicyInfo

`func NewPolicyInfo() *PolicyInfo`

NewPolicyInfo instantiates a new PolicyInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyInfoWithDefaults

`func NewPolicyInfoWithDefaults() *PolicyInfo`

NewPolicyInfoWithDefaults instantiates a new PolicyInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributeDefaults

`func (o *PolicyInfo) GetAttributeDefaults() AttributeDefaults`

GetAttributeDefaults returns the AttributeDefaults field if non-nil, zero value otherwise.

### GetAttributeDefaultsOk

`func (o *PolicyInfo) GetAttributeDefaultsOk() (*AttributeDefaults, bool)`

GetAttributeDefaultsOk returns a tuple with the AttributeDefaults field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributeDefaults

`func (o *PolicyInfo) SetAttributeDefaults(v AttributeDefaults)`

SetAttributeDefaults sets AttributeDefaults field to given value.

### HasAttributeDefaults

`func (o *PolicyInfo) HasAttributeDefaults() bool`

HasAttributeDefaults returns a boolean if a field has been set.

### SetAttributeDefaultsNil

`func (o *PolicyInfo) SetAttributeDefaultsNil(b bool)`

 SetAttributeDefaultsNil sets the value for AttributeDefaults to be an explicit nil

### UnsetAttributeDefaults
`func (o *PolicyInfo) UnsetAttributeDefaults()`

UnsetAttributeDefaults ensures that no value is present for AttributeDefaults, not even an explicit nil
### GetRules

`func (o *PolicyInfo) GetRules() Rules`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *PolicyInfo) GetRulesOk() (*Rules, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *PolicyInfo) SetRules(v Rules)`

SetRules sets Rules field to given value.

### HasRules

`func (o *PolicyInfo) HasRules() bool`

HasRules returns a boolean if a field has been set.

### SetRulesNil

`func (o *PolicyInfo) SetRulesNil(b bool)`

 SetRulesNil sets the value for Rules to be an explicit nil

### UnsetRules
`func (o *PolicyInfo) UnsetRules()`

UnsetRules ensures that no value is present for Rules, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


