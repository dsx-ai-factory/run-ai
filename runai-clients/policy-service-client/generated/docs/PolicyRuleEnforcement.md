# PolicyRuleEnforcement

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Set** | Pointer to **map[string]interface{}** | Enforces the value of the field. | [optional] 
**Default** | Pointer to **map[string]interface{}** | Sets the value of the field when the user does not supply one. | [optional] 
**Required** | Pointer to **NullableBool** | Requires the field to have a value, supplied either in the submission or by a policy default. | [optional] [default to false]
**Blocked** | Pointer to **NullableBool** | Prevents the field from being assigned any value. Inverse of \&quot;required\&quot;. | [optional] [default to false]
**Length** | Pointer to [**NullablePolicyEnforcementLengthLength**](PolicyEnforcementLengthLength.md) |  | [optional] 
**Options** | Pointer to **[]string** | Restricts the value of the field to one of the listed options. | [optional] 
**Match** | Pointer to **NullableString** | Requires the value of the field to match the regular expression. | [optional] 
**Range** | Pointer to [**NullablePolicyEnforcementRangeRange**](PolicyEnforcementRangeRange.md) |  | [optional] 
**Elements** | Pointer to [**NullablePolicyEnforcementElementsElements**](PolicyEnforcementElementsElements.md) |  | [optional] 
**Items** | Pointer to [**NullablePolicyEnforcementItemsItems**](PolicyEnforcementItemsItems.md) |  | [optional] 
**QuantityRange** | Pointer to [**NullablePolicyEnforcementQuantityRangeQuantityRange**](PolicyEnforcementQuantityRangeQuantityRange.md) |  | [optional] 

## Methods

### NewPolicyRuleEnforcement

`func NewPolicyRuleEnforcement() *PolicyRuleEnforcement`

NewPolicyRuleEnforcement instantiates a new PolicyRuleEnforcement object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyRuleEnforcementWithDefaults

`func NewPolicyRuleEnforcementWithDefaults() *PolicyRuleEnforcement`

NewPolicyRuleEnforcementWithDefaults instantiates a new PolicyRuleEnforcement object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSet

`func (o *PolicyRuleEnforcement) GetSet() map[string]interface{}`

GetSet returns the Set field if non-nil, zero value otherwise.

### GetSetOk

`func (o *PolicyRuleEnforcement) GetSetOk() (*map[string]interface{}, bool)`

GetSetOk returns a tuple with the Set field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSet

`func (o *PolicyRuleEnforcement) SetSet(v map[string]interface{})`

SetSet sets Set field to given value.

### HasSet

`func (o *PolicyRuleEnforcement) HasSet() bool`

HasSet returns a boolean if a field has been set.

### GetDefault

`func (o *PolicyRuleEnforcement) GetDefault() map[string]interface{}`

GetDefault returns the Default field if non-nil, zero value otherwise.

### GetDefaultOk

`func (o *PolicyRuleEnforcement) GetDefaultOk() (*map[string]interface{}, bool)`

GetDefaultOk returns a tuple with the Default field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefault

`func (o *PolicyRuleEnforcement) SetDefault(v map[string]interface{})`

SetDefault sets Default field to given value.

### HasDefault

`func (o *PolicyRuleEnforcement) HasDefault() bool`

HasDefault returns a boolean if a field has been set.

### GetRequired

`func (o *PolicyRuleEnforcement) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *PolicyRuleEnforcement) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *PolicyRuleEnforcement) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *PolicyRuleEnforcement) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *PolicyRuleEnforcement) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *PolicyRuleEnforcement) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetBlocked

`func (o *PolicyRuleEnforcement) GetBlocked() bool`

GetBlocked returns the Blocked field if non-nil, zero value otherwise.

### GetBlockedOk

`func (o *PolicyRuleEnforcement) GetBlockedOk() (*bool, bool)`

GetBlockedOk returns a tuple with the Blocked field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlocked

`func (o *PolicyRuleEnforcement) SetBlocked(v bool)`

SetBlocked sets Blocked field to given value.

### HasBlocked

`func (o *PolicyRuleEnforcement) HasBlocked() bool`

HasBlocked returns a boolean if a field has been set.

### SetBlockedNil

`func (o *PolicyRuleEnforcement) SetBlockedNil(b bool)`

 SetBlockedNil sets the value for Blocked to be an explicit nil

### UnsetBlocked
`func (o *PolicyRuleEnforcement) UnsetBlocked()`

UnsetBlocked ensures that no value is present for Blocked, not even an explicit nil
### GetLength

`func (o *PolicyRuleEnforcement) GetLength() PolicyEnforcementLengthLength`

GetLength returns the Length field if non-nil, zero value otherwise.

### GetLengthOk

`func (o *PolicyRuleEnforcement) GetLengthOk() (*PolicyEnforcementLengthLength, bool)`

GetLengthOk returns a tuple with the Length field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLength

`func (o *PolicyRuleEnforcement) SetLength(v PolicyEnforcementLengthLength)`

SetLength sets Length field to given value.

### HasLength

`func (o *PolicyRuleEnforcement) HasLength() bool`

HasLength returns a boolean if a field has been set.

### SetLengthNil

`func (o *PolicyRuleEnforcement) SetLengthNil(b bool)`

 SetLengthNil sets the value for Length to be an explicit nil

### UnsetLength
`func (o *PolicyRuleEnforcement) UnsetLength()`

UnsetLength ensures that no value is present for Length, not even an explicit nil
### GetOptions

`func (o *PolicyRuleEnforcement) GetOptions() []string`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *PolicyRuleEnforcement) GetOptionsOk() (*[]string, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *PolicyRuleEnforcement) SetOptions(v []string)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *PolicyRuleEnforcement) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *PolicyRuleEnforcement) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *PolicyRuleEnforcement) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil
### GetMatch

`func (o *PolicyRuleEnforcement) GetMatch() string`

GetMatch returns the Match field if non-nil, zero value otherwise.

### GetMatchOk

`func (o *PolicyRuleEnforcement) GetMatchOk() (*string, bool)`

GetMatchOk returns a tuple with the Match field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMatch

`func (o *PolicyRuleEnforcement) SetMatch(v string)`

SetMatch sets Match field to given value.

### HasMatch

`func (o *PolicyRuleEnforcement) HasMatch() bool`

HasMatch returns a boolean if a field has been set.

### SetMatchNil

`func (o *PolicyRuleEnforcement) SetMatchNil(b bool)`

 SetMatchNil sets the value for Match to be an explicit nil

### UnsetMatch
`func (o *PolicyRuleEnforcement) UnsetMatch()`

UnsetMatch ensures that no value is present for Match, not even an explicit nil
### GetRange

`func (o *PolicyRuleEnforcement) GetRange() PolicyEnforcementRangeRange`

GetRange returns the Range field if non-nil, zero value otherwise.

### GetRangeOk

`func (o *PolicyRuleEnforcement) GetRangeOk() (*PolicyEnforcementRangeRange, bool)`

GetRangeOk returns a tuple with the Range field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRange

`func (o *PolicyRuleEnforcement) SetRange(v PolicyEnforcementRangeRange)`

SetRange sets Range field to given value.

### HasRange

`func (o *PolicyRuleEnforcement) HasRange() bool`

HasRange returns a boolean if a field has been set.

### SetRangeNil

`func (o *PolicyRuleEnforcement) SetRangeNil(b bool)`

 SetRangeNil sets the value for Range to be an explicit nil

### UnsetRange
`func (o *PolicyRuleEnforcement) UnsetRange()`

UnsetRange ensures that no value is present for Range, not even an explicit nil
### GetElements

`func (o *PolicyRuleEnforcement) GetElements() PolicyEnforcementElementsElements`

GetElements returns the Elements field if non-nil, zero value otherwise.

### GetElementsOk

`func (o *PolicyRuleEnforcement) GetElementsOk() (*PolicyEnforcementElementsElements, bool)`

GetElementsOk returns a tuple with the Elements field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetElements

`func (o *PolicyRuleEnforcement) SetElements(v PolicyEnforcementElementsElements)`

SetElements sets Elements field to given value.

### HasElements

`func (o *PolicyRuleEnforcement) HasElements() bool`

HasElements returns a boolean if a field has been set.

### SetElementsNil

`func (o *PolicyRuleEnforcement) SetElementsNil(b bool)`

 SetElementsNil sets the value for Elements to be an explicit nil

### UnsetElements
`func (o *PolicyRuleEnforcement) UnsetElements()`

UnsetElements ensures that no value is present for Elements, not even an explicit nil
### GetItems

`func (o *PolicyRuleEnforcement) GetItems() PolicyEnforcementItemsItems`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *PolicyRuleEnforcement) GetItemsOk() (*PolicyEnforcementItemsItems, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *PolicyRuleEnforcement) SetItems(v PolicyEnforcementItemsItems)`

SetItems sets Items field to given value.

### HasItems

`func (o *PolicyRuleEnforcement) HasItems() bool`

HasItems returns a boolean if a field has been set.

### SetItemsNil

`func (o *PolicyRuleEnforcement) SetItemsNil(b bool)`

 SetItemsNil sets the value for Items to be an explicit nil

### UnsetItems
`func (o *PolicyRuleEnforcement) UnsetItems()`

UnsetItems ensures that no value is present for Items, not even an explicit nil
### GetQuantityRange

`func (o *PolicyRuleEnforcement) GetQuantityRange() PolicyEnforcementQuantityRangeQuantityRange`

GetQuantityRange returns the QuantityRange field if non-nil, zero value otherwise.

### GetQuantityRangeOk

`func (o *PolicyRuleEnforcement) GetQuantityRangeOk() (*PolicyEnforcementQuantityRangeQuantityRange, bool)`

GetQuantityRangeOk returns a tuple with the QuantityRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuantityRange

`func (o *PolicyRuleEnforcement) SetQuantityRange(v PolicyEnforcementQuantityRangeQuantityRange)`

SetQuantityRange sets QuantityRange field to given value.

### HasQuantityRange

`func (o *PolicyRuleEnforcement) HasQuantityRange() bool`

HasQuantityRange returns a boolean if a field has been set.

### SetQuantityRangeNil

`func (o *PolicyRuleEnforcement) SetQuantityRangeNil(b bool)`

 SetQuantityRangeNil sets the value for QuantityRange to be an explicit nil

### UnsetQuantityRange
`func (o *PolicyRuleEnforcement) UnsetQuantityRange()`

UnsetQuantityRange ensures that no value is present for QuantityRange, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


