# BreakdownFieldVariation

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SpecSelectors** | Pointer to **[]string** | The spec selectors under which this variation of the field appears. | [optional] 
**Formatter** | Pointer to [**BreakdownFormatter**](BreakdownFormatter.md) |  | [optional] 
**SpecRef** | Pointer to **string** | A direct dotted path into the workload spec, used in place of a formatter for simple fields. | [optional] 
**Type** | Pointer to [**MappingFieldType**](MappingFieldType.md) | The data type of the field. | [optional] 
**ItemSelectors** | Pointer to [**map[string]BreakdownItemSelector**](BreakdownItemSelector.md) | The item selectors that apply to the field, keyed by item-selector name, with each value containing the item selector&#39;s JQ query and argument metadata. | [optional] 
**PossibleEnforcements** | Pointer to [**[]MappingPossibleEnforcement**](MappingPossibleEnforcement.md) | The enforcement types that may be applied to this variation of the field. Derived from the field&#39;s data type, with &#x60;items&#x60; added when the field is itemized (has an itemSelector). | [optional] 

## Methods

### NewBreakdownFieldVariation

`func NewBreakdownFieldVariation() *BreakdownFieldVariation`

NewBreakdownFieldVariation instantiates a new BreakdownFieldVariation object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBreakdownFieldVariationWithDefaults

`func NewBreakdownFieldVariationWithDefaults() *BreakdownFieldVariation`

NewBreakdownFieldVariationWithDefaults instantiates a new BreakdownFieldVariation object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSpecSelectors

`func (o *BreakdownFieldVariation) GetSpecSelectors() []string`

GetSpecSelectors returns the SpecSelectors field if non-nil, zero value otherwise.

### GetSpecSelectorsOk

`func (o *BreakdownFieldVariation) GetSpecSelectorsOk() (*[]string, bool)`

GetSpecSelectorsOk returns a tuple with the SpecSelectors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecSelectors

`func (o *BreakdownFieldVariation) SetSpecSelectors(v []string)`

SetSpecSelectors sets SpecSelectors field to given value.

### HasSpecSelectors

`func (o *BreakdownFieldVariation) HasSpecSelectors() bool`

HasSpecSelectors returns a boolean if a field has been set.

### GetFormatter

`func (o *BreakdownFieldVariation) GetFormatter() BreakdownFormatter`

GetFormatter returns the Formatter field if non-nil, zero value otherwise.

### GetFormatterOk

`func (o *BreakdownFieldVariation) GetFormatterOk() (*BreakdownFormatter, bool)`

GetFormatterOk returns a tuple with the Formatter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFormatter

`func (o *BreakdownFieldVariation) SetFormatter(v BreakdownFormatter)`

SetFormatter sets Formatter field to given value.

### HasFormatter

`func (o *BreakdownFieldVariation) HasFormatter() bool`

HasFormatter returns a boolean if a field has been set.

### GetSpecRef

`func (o *BreakdownFieldVariation) GetSpecRef() string`

GetSpecRef returns the SpecRef field if non-nil, zero value otherwise.

### GetSpecRefOk

`func (o *BreakdownFieldVariation) GetSpecRefOk() (*string, bool)`

GetSpecRefOk returns a tuple with the SpecRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecRef

`func (o *BreakdownFieldVariation) SetSpecRef(v string)`

SetSpecRef sets SpecRef field to given value.

### HasSpecRef

`func (o *BreakdownFieldVariation) HasSpecRef() bool`

HasSpecRef returns a boolean if a field has been set.

### GetType

`func (o *BreakdownFieldVariation) GetType() MappingFieldType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *BreakdownFieldVariation) GetTypeOk() (*MappingFieldType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *BreakdownFieldVariation) SetType(v MappingFieldType)`

SetType sets Type field to given value.

### HasType

`func (o *BreakdownFieldVariation) HasType() bool`

HasType returns a boolean if a field has been set.

### GetItemSelectors

`func (o *BreakdownFieldVariation) GetItemSelectors() map[string]BreakdownItemSelector`

GetItemSelectors returns the ItemSelectors field if non-nil, zero value otherwise.

### GetItemSelectorsOk

`func (o *BreakdownFieldVariation) GetItemSelectorsOk() (*map[string]BreakdownItemSelector, bool)`

GetItemSelectorsOk returns a tuple with the ItemSelectors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItemSelectors

`func (o *BreakdownFieldVariation) SetItemSelectors(v map[string]BreakdownItemSelector)`

SetItemSelectors sets ItemSelectors field to given value.

### HasItemSelectors

`func (o *BreakdownFieldVariation) HasItemSelectors() bool`

HasItemSelectors returns a boolean if a field has been set.

### GetPossibleEnforcements

`func (o *BreakdownFieldVariation) GetPossibleEnforcements() []MappingPossibleEnforcement`

GetPossibleEnforcements returns the PossibleEnforcements field if non-nil, zero value otherwise.

### GetPossibleEnforcementsOk

`func (o *BreakdownFieldVariation) GetPossibleEnforcementsOk() (*[]MappingPossibleEnforcement, bool)`

GetPossibleEnforcementsOk returns a tuple with the PossibleEnforcements field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPossibleEnforcements

`func (o *BreakdownFieldVariation) SetPossibleEnforcements(v []MappingPossibleEnforcement)`

SetPossibleEnforcements sets PossibleEnforcements field to given value.

### HasPossibleEnforcements

`func (o *BreakdownFieldVariation) HasPossibleEnforcements() bool`

HasPossibleEnforcements returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


