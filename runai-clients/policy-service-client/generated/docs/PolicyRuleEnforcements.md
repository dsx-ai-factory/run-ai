# PolicyRuleEnforcements

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enforcements** | [**[]PolicyRuleEnforcement**](PolicyRuleEnforcement.md) | The set of enforcements that define the rule&#39;s behavior. Each enforcement pairs an enforcement type with its value, and multiple enforcements may combine on a single rule to constrain the targeted field. | 

## Methods

### NewPolicyRuleEnforcements

`func NewPolicyRuleEnforcements(enforcements []PolicyRuleEnforcement, ) *PolicyRuleEnforcements`

NewPolicyRuleEnforcements instantiates a new PolicyRuleEnforcements object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyRuleEnforcementsWithDefaults

`func NewPolicyRuleEnforcementsWithDefaults() *PolicyRuleEnforcements`

NewPolicyRuleEnforcementsWithDefaults instantiates a new PolicyRuleEnforcements object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnforcements

`func (o *PolicyRuleEnforcements) GetEnforcements() []PolicyRuleEnforcement`

GetEnforcements returns the Enforcements field if non-nil, zero value otherwise.

### GetEnforcementsOk

`func (o *PolicyRuleEnforcements) GetEnforcementsOk() (*[]PolicyRuleEnforcement, bool)`

GetEnforcementsOk returns a tuple with the Enforcements field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnforcements

`func (o *PolicyRuleEnforcements) SetEnforcements(v []PolicyRuleEnforcement)`

SetEnforcements sets Enforcements field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


