# PolicyRulesRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Rules** | [**[]PolicyRuleRef**](PolicyRuleRef.md) | References to the rules to remove from the policy. Each entry identifies a rule by its identity fields only (field, spec selector, and optional item selector and selector parameters). | 

## Methods

### NewPolicyRulesRef

`func NewPolicyRulesRef(rules []PolicyRuleRef, ) *PolicyRulesRef`

NewPolicyRulesRef instantiates a new PolicyRulesRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyRulesRefWithDefaults

`func NewPolicyRulesRefWithDefaults() *PolicyRulesRef`

NewPolicyRulesRefWithDefaults instantiates a new PolicyRulesRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRules

`func (o *PolicyRulesRef) GetRules() []PolicyRuleRef`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *PolicyRulesRef) GetRulesOk() (*[]PolicyRuleRef, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *PolicyRulesRef) SetRules(v []PolicyRuleRef)`

SetRules sets Rules field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


