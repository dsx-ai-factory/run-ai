# BreakdownFieldRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FieldName** | Pointer to **string** | The name of the field to retrieve the breakdown for. | [optional] 
**SpecSelector** | Pointer to **NullableString** | The name of the spec selector that exposes the field. Optional; when omitted, the breakdown includes all spec selectors that expose the field. | [optional] 
**ItemSelector** | Pointer to **NullableString** | The name of the item selector that applies to the field. Optional; omit for non-itemized fields. | [optional] 

## Methods

### NewBreakdownFieldRef

`func NewBreakdownFieldRef() *BreakdownFieldRef`

NewBreakdownFieldRef instantiates a new BreakdownFieldRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBreakdownFieldRefWithDefaults

`func NewBreakdownFieldRefWithDefaults() *BreakdownFieldRef`

NewBreakdownFieldRefWithDefaults instantiates a new BreakdownFieldRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFieldName

`func (o *BreakdownFieldRef) GetFieldName() string`

GetFieldName returns the FieldName field if non-nil, zero value otherwise.

### GetFieldNameOk

`func (o *BreakdownFieldRef) GetFieldNameOk() (*string, bool)`

GetFieldNameOk returns a tuple with the FieldName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldName

`func (o *BreakdownFieldRef) SetFieldName(v string)`

SetFieldName sets FieldName field to given value.

### HasFieldName

`func (o *BreakdownFieldRef) HasFieldName() bool`

HasFieldName returns a boolean if a field has been set.

### GetSpecSelector

`func (o *BreakdownFieldRef) GetSpecSelector() string`

GetSpecSelector returns the SpecSelector field if non-nil, zero value otherwise.

### GetSpecSelectorOk

`func (o *BreakdownFieldRef) GetSpecSelectorOk() (*string, bool)`

GetSpecSelectorOk returns a tuple with the SpecSelector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecSelector

`func (o *BreakdownFieldRef) SetSpecSelector(v string)`

SetSpecSelector sets SpecSelector field to given value.

### HasSpecSelector

`func (o *BreakdownFieldRef) HasSpecSelector() bool`

HasSpecSelector returns a boolean if a field has been set.

### SetSpecSelectorNil

`func (o *BreakdownFieldRef) SetSpecSelectorNil(b bool)`

 SetSpecSelectorNil sets the value for SpecSelector to be an explicit nil

### UnsetSpecSelector
`func (o *BreakdownFieldRef) UnsetSpecSelector()`

UnsetSpecSelector ensures that no value is present for SpecSelector, not even an explicit nil
### GetItemSelector

`func (o *BreakdownFieldRef) GetItemSelector() string`

GetItemSelector returns the ItemSelector field if non-nil, zero value otherwise.

### GetItemSelectorOk

`func (o *BreakdownFieldRef) GetItemSelectorOk() (*string, bool)`

GetItemSelectorOk returns a tuple with the ItemSelector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItemSelector

`func (o *BreakdownFieldRef) SetItemSelector(v string)`

SetItemSelector sets ItemSelector field to given value.

### HasItemSelector

`func (o *BreakdownFieldRef) HasItemSelector() bool`

HasItemSelector returns a boolean if a field has been set.

### SetItemSelectorNil

`func (o *BreakdownFieldRef) SetItemSelectorNil(b bool)`

 SetItemSelectorNil sets the value for ItemSelector to be an explicit nil

### UnsetItemSelector
`func (o *BreakdownFieldRef) UnsetItemSelector()`

UnsetItemSelector ensures that no value is present for ItemSelector, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


