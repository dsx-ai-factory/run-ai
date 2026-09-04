# FieldGroup

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The unique name of the field group | 
**Description** | Pointer to **NullableString** | A free-text description of the element. | [optional] 
**Fields** | Pointer to [**map[string]FieldGroupFieldsWithScopeValue**](FieldGroupFieldsWithScopeValue.md) | A map of the field group fields. Map **keys** are field names. OpenAPI &#x60;additionalProperties&#x60; only constrains the **value** shape (&#x60;FieldGroupField&#x60; plus scope); it does **not** validate keys. The API enforces each key to match &#x60;MappingFieldName&#x60; (1–63 chars, &#x60;^[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?$&#x60; — alphanumeric with optional internal hyphens; no leading/trailing hyphen).  | [optional] 
**ItemSelectors** | Pointer to [**map[string]ItemSelectorsWithScopeValue**](ItemSelectorsWithScopeValue.md) | A map of the field group item selectors. Map **keys** are item selector names. OpenAPI &#x60;additionalProperties&#x60; only constrains the **value** shape (&#x60;FieldGroupItemSelector&#x60;); it does **not** validate keys. The API enforces each key to match &#x60;ItemSelectorName&#x60; (1–63 chars, &#x60;^[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?$&#x60; — alphanumeric with optional internal hyphens; no leading/trailing hyphen).  | [optional] 
**Id** | Pointer to **string** | The unique identifier for the field group. | [optional] 
**Scope** | Pointer to [**MappingScope**](MappingScope.md) |  | [optional] 
**CreatedBy** | Pointer to **string** | The identifier of the user who created the field group. | [optional] 
**CreatedAt** | Pointer to **time.Time** | The time at which the field group was created. | [optional] 
**UpdatedBy** | Pointer to **string** | The identifier of the user who last updated the field group. | [optional] 
**UpdatedAt** | Pointer to **time.Time** | The time at which the field group was last updated. | [optional] 

## Methods

### NewFieldGroup

`func NewFieldGroup(name string, ) *FieldGroup`

NewFieldGroup instantiates a new FieldGroup object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFieldGroupWithDefaults

`func NewFieldGroupWithDefaults() *FieldGroup`

NewFieldGroupWithDefaults instantiates a new FieldGroup object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *FieldGroup) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *FieldGroup) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *FieldGroup) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *FieldGroup) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *FieldGroup) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *FieldGroup) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *FieldGroup) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *FieldGroup) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *FieldGroup) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetFields

`func (o *FieldGroup) GetFields() map[string]FieldGroupFieldsWithScopeValue`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *FieldGroup) GetFieldsOk() (*map[string]FieldGroupFieldsWithScopeValue, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *FieldGroup) SetFields(v map[string]FieldGroupFieldsWithScopeValue)`

SetFields sets Fields field to given value.

### HasFields

`func (o *FieldGroup) HasFields() bool`

HasFields returns a boolean if a field has been set.

### GetItemSelectors

`func (o *FieldGroup) GetItemSelectors() map[string]ItemSelectorsWithScopeValue`

GetItemSelectors returns the ItemSelectors field if non-nil, zero value otherwise.

### GetItemSelectorsOk

`func (o *FieldGroup) GetItemSelectorsOk() (*map[string]ItemSelectorsWithScopeValue, bool)`

GetItemSelectorsOk returns a tuple with the ItemSelectors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItemSelectors

`func (o *FieldGroup) SetItemSelectors(v map[string]ItemSelectorsWithScopeValue)`

SetItemSelectors sets ItemSelectors field to given value.

### HasItemSelectors

`func (o *FieldGroup) HasItemSelectors() bool`

HasItemSelectors returns a boolean if a field has been set.

### GetId

`func (o *FieldGroup) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *FieldGroup) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *FieldGroup) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *FieldGroup) HasId() bool`

HasId returns a boolean if a field has been set.

### GetScope

`func (o *FieldGroup) GetScope() MappingScope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *FieldGroup) GetScopeOk() (*MappingScope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *FieldGroup) SetScope(v MappingScope)`

SetScope sets Scope field to given value.

### HasScope

`func (o *FieldGroup) HasScope() bool`

HasScope returns a boolean if a field has been set.

### GetCreatedBy

`func (o *FieldGroup) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *FieldGroup) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *FieldGroup) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *FieldGroup) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetCreatedAt

`func (o *FieldGroup) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *FieldGroup) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *FieldGroup) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *FieldGroup) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *FieldGroup) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *FieldGroup) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *FieldGroup) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *FieldGroup) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *FieldGroup) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *FieldGroup) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *FieldGroup) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *FieldGroup) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


