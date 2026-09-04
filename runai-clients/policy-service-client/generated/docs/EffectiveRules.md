# EffectiveRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**EffectiveRules** | Pointer to [**[]EffectiveRulesEffectiveRulesInner**](EffectiveRulesEffectiveRulesInner.md) | The merged rule set in effect for this policy, including rules inherited from broader scopes (cluster, then department, then project). Each entry adds the sourceOfRule indicating where the rule was defined. Returned in responses only; not accepted in requests. | [optional] 

## Methods

### NewEffectiveRules

`func NewEffectiveRules() *EffectiveRules`

NewEffectiveRules instantiates a new EffectiveRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEffectiveRulesWithDefaults

`func NewEffectiveRulesWithDefaults() *EffectiveRules`

NewEffectiveRulesWithDefaults instantiates a new EffectiveRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEffectiveRules

`func (o *EffectiveRules) GetEffectiveRules() []EffectiveRulesEffectiveRulesInner`

GetEffectiveRules returns the EffectiveRules field if non-nil, zero value otherwise.

### GetEffectiveRulesOk

`func (o *EffectiveRules) GetEffectiveRulesOk() (*[]EffectiveRulesEffectiveRulesInner, bool)`

GetEffectiveRulesOk returns a tuple with the EffectiveRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffectiveRules

`func (o *EffectiveRules) SetEffectiveRules(v []EffectiveRulesEffectiveRulesInner)`

SetEffectiveRules sets EffectiveRules field to given value.

### HasEffectiveRules

`func (o *EffectiveRules) HasEffectiveRules() bool`

HasEffectiveRules returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


