# UpdateGvkPolicyReq

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **NullableString** | A description of the policy. | [optional] 
**Rules** | Pointer to [**[]PolicyRule**](PolicyRule.md) | The rules declared on this policy. Each rule targets a field in the workload spec and constrains its allowed values through its enforcements. | [optional] 

## Methods

### NewUpdateGvkPolicyReq

`func NewUpdateGvkPolicyReq() *UpdateGvkPolicyReq`

NewUpdateGvkPolicyReq instantiates a new UpdateGvkPolicyReq object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateGvkPolicyReqWithDefaults

`func NewUpdateGvkPolicyReqWithDefaults() *UpdateGvkPolicyReq`

NewUpdateGvkPolicyReqWithDefaults instantiates a new UpdateGvkPolicyReq object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *UpdateGvkPolicyReq) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *UpdateGvkPolicyReq) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *UpdateGvkPolicyReq) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *UpdateGvkPolicyReq) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *UpdateGvkPolicyReq) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *UpdateGvkPolicyReq) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetRules

`func (o *UpdateGvkPolicyReq) GetRules() []PolicyRule`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *UpdateGvkPolicyReq) GetRulesOk() (*[]PolicyRule, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *UpdateGvkPolicyReq) SetRules(v []PolicyRule)`

SetRules sets Rules field to given value.

### HasRules

`func (o *UpdateGvkPolicyReq) HasRules() bool`

HasRules returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


