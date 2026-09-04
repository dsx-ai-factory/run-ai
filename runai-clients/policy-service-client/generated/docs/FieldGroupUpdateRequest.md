# FieldGroupUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **NullableString** | A free-text description of the element. | [optional] 
**Fields** | Pointer to [**map[string]FieldGroupField**](FieldGroupField.md) | A map of the field group fields. Map **keys** are field names. OpenAPI &#x60;additionalProperties&#x60; only constrains the **value** shape (&#x60;FieldGroupField&#x60;); it does **not** validate keys. The API enforces each key to match &#x60;MappingFieldName&#x60; (1–63 chars, &#x60;^[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?$&#x60; — alphanumeric with optional internal hyphens; no leading/trailing hyphen).  | [optional] 
**ItemSelectors** | Pointer to [**map[string]ItemSelector**](ItemSelector.md) | A map of the field group item selectors. Map **keys** are item selector names. OpenAPI &#x60;additionalProperties&#x60; only constrains the **value** shape (&#x60;ItemSelector&#x60;); it does **not** validate keys. The API enforces each key to match &#x60;ItemSelectorName&#x60; (1–63 chars, &#x60;^[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?$&#x60; — alphanumeric with optional internal hyphens; no leading/trailing hyphen).  | [optional] 

## Methods

### NewFieldGroupUpdateRequest

`func NewFieldGroupUpdateRequest() *FieldGroupUpdateRequest`

NewFieldGroupUpdateRequest instantiates a new FieldGroupUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFieldGroupUpdateRequestWithDefaults

`func NewFieldGroupUpdateRequestWithDefaults() *FieldGroupUpdateRequest`

NewFieldGroupUpdateRequestWithDefaults instantiates a new FieldGroupUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *FieldGroupUpdateRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *FieldGroupUpdateRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *FieldGroupUpdateRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *FieldGroupUpdateRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *FieldGroupUpdateRequest) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *FieldGroupUpdateRequest) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetFields

`func (o *FieldGroupUpdateRequest) GetFields() map[string]FieldGroupField`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *FieldGroupUpdateRequest) GetFieldsOk() (*map[string]FieldGroupField, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *FieldGroupUpdateRequest) SetFields(v map[string]FieldGroupField)`

SetFields sets Fields field to given value.

### HasFields

`func (o *FieldGroupUpdateRequest) HasFields() bool`

HasFields returns a boolean if a field has been set.

### GetItemSelectors

`func (o *FieldGroupUpdateRequest) GetItemSelectors() map[string]ItemSelector`

GetItemSelectors returns the ItemSelectors field if non-nil, zero value otherwise.

### GetItemSelectorsOk

`func (o *FieldGroupUpdateRequest) GetItemSelectorsOk() (*map[string]ItemSelector, bool)`

GetItemSelectorsOk returns a tuple with the ItemSelectors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItemSelectors

`func (o *FieldGroupUpdateRequest) SetItemSelectors(v map[string]ItemSelector)`

SetItemSelectors sets ItemSelectors field to given value.

### HasItemSelectors

`func (o *FieldGroupUpdateRequest) HasItemSelectors() bool`

HasItemSelectors returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


