# PolicyRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Rules** | Pointer to [**[]PolicyRule**](PolicyRule.md) | The rules declared on this policy. Each rule targets a field in the workload spec and constrains its allowed values through its enforcements. | [optional] 

## Methods

### NewPolicyRules

`func NewPolicyRules() *PolicyRules`

NewPolicyRules instantiates a new PolicyRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyRulesWithDefaults

`func NewPolicyRulesWithDefaults() *PolicyRules`

NewPolicyRulesWithDefaults instantiates a new PolicyRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRules

`func (o *PolicyRules) GetRules() []PolicyRule`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *PolicyRules) GetRulesOk() (*[]PolicyRule, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *PolicyRules) SetRules(v []PolicyRule)`

SetRules sets Rules field to given value.

### HasRules

`func (o *PolicyRules) HasRules() bool`

HasRules returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


