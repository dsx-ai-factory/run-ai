# BreakdownBreakdown

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Fields** | Pointer to [**map[string][]BreakdownFieldVariation**](array.md) | The detailed breakdown of fields, keyed by field name. Each field maps to the list of its variations across the requested spec selectors and item selectors. | [optional] 
**SpecSelectors** | Pointer to [**map[string]BreakdownSpecSelector**](BreakdownSpecSelector.md) | The detailed breakdown of spec selectors, keyed by spec-selector name. | [optional] 

## Methods

### NewBreakdownBreakdown

`func NewBreakdownBreakdown() *BreakdownBreakdown`

NewBreakdownBreakdown instantiates a new BreakdownBreakdown object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBreakdownBreakdownWithDefaults

`func NewBreakdownBreakdownWithDefaults() *BreakdownBreakdown`

NewBreakdownBreakdownWithDefaults instantiates a new BreakdownBreakdown object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFields

`func (o *BreakdownBreakdown) GetFields() map[string][]BreakdownFieldVariation`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *BreakdownBreakdown) GetFieldsOk() (*map[string][]BreakdownFieldVariation, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *BreakdownBreakdown) SetFields(v map[string][]BreakdownFieldVariation)`

SetFields sets Fields field to given value.

### HasFields

`func (o *BreakdownBreakdown) HasFields() bool`

HasFields returns a boolean if a field has been set.

### GetSpecSelectors

`func (o *BreakdownBreakdown) GetSpecSelectors() map[string]BreakdownSpecSelector`

GetSpecSelectors returns the SpecSelectors field if non-nil, zero value otherwise.

### GetSpecSelectorsOk

`func (o *BreakdownBreakdown) GetSpecSelectorsOk() (*map[string]BreakdownSpecSelector, bool)`

GetSpecSelectorsOk returns a tuple with the SpecSelectors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecSelectors

`func (o *BreakdownBreakdown) SetSpecSelectors(v map[string]BreakdownSpecSelector)`

SetSpecSelectors sets SpecSelectors field to given value.

### HasSpecSelectors

`func (o *BreakdownBreakdown) HasSpecSelectors() bool`

HasSpecSelectors returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


