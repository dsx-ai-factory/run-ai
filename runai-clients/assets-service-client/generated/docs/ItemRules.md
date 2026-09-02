# ItemRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**CanAdd** | Pointer to **NullableBool** | Whether it is permitted to add items. Default to true. | [optional] 
**Locked** | Pointer to **[]string** | Set of keys for items that are \&quot;locked\&quot;, i.e. cannot be removed or deleted. | [optional] 

## Methods

### NewItemRules

`func NewItemRules() *ItemRules`

NewItemRules instantiates a new ItemRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewItemRulesWithDefaults

`func NewItemRulesWithDefaults() *ItemRules`

NewItemRulesWithDefaults instantiates a new ItemRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *ItemRules) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *ItemRules) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *ItemRules) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *ItemRules) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *ItemRules) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *ItemRules) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetCanAdd

`func (o *ItemRules) GetCanAdd() bool`

GetCanAdd returns the CanAdd field if non-nil, zero value otherwise.

### GetCanAddOk

`func (o *ItemRules) GetCanAddOk() (*bool, bool)`

GetCanAddOk returns a tuple with the CanAdd field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanAdd

`func (o *ItemRules) SetCanAdd(v bool)`

SetCanAdd sets CanAdd field to given value.

### HasCanAdd

`func (o *ItemRules) HasCanAdd() bool`

HasCanAdd returns a boolean if a field has been set.

### SetCanAddNil

`func (o *ItemRules) SetCanAddNil(b bool)`

 SetCanAddNil sets the value for CanAdd to be an explicit nil

### UnsetCanAdd
`func (o *ItemRules) UnsetCanAdd()`

UnsetCanAdd ensures that no value is present for CanAdd, not even an explicit nil
### GetLocked

`func (o *ItemRules) GetLocked() []string`

GetLocked returns the Locked field if non-nil, zero value otherwise.

### GetLockedOk

`func (o *ItemRules) GetLockedOk() (*[]string, bool)`

GetLockedOk returns a tuple with the Locked field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocked

`func (o *ItemRules) SetLocked(v []string)`

SetLocked sets Locked field to given value.

### HasLocked

`func (o *ItemRules) HasLocked() bool`

HasLocked returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


