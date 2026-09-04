# Formatter

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The unique name of the formatter | 
**Description** | Pointer to **NullableString** | A free-text description of the element. | [optional] 
**Fields** | Pointer to [**map[string]FormatterField**](FormatterField.md) | A map of the formatter fields. Map **keys** are field names. OpenAPI &#x60;additionalProperties&#x60; only constrains the **value** shape (&#x60;FormatterField&#x60;); it does **not** validate keys. The API enforces each key to match &#x60;MappingFieldName&#x60; (1–63 chars, &#x60;^[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?$&#x60; — alphanumeric with optional internal hyphens; no leading/trailing hyphen).  | [optional] 
**Id** | **string** | The unique identifier for the formatter. | 
**Scope** | Pointer to [**MappingScope**](MappingScope.md) |  | [optional] 
**CreatedBy** | Pointer to **string** | The identifier of the user who created the formatter. | [optional] 
**CreatedAt** | Pointer to **time.Time** | The time at which the formatter was created. | [optional] 
**UpdatedBy** | Pointer to **string** | The identifier of the user who last updated the formatter. | [optional] 
**UpdatedAt** | Pointer to **time.Time** | The time at which the formatter was last updated. | [optional] 

## Methods

### NewFormatter

`func NewFormatter(name string, id string, ) *Formatter`

NewFormatter instantiates a new Formatter object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFormatterWithDefaults

`func NewFormatterWithDefaults() *Formatter`

NewFormatterWithDefaults instantiates a new Formatter object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Formatter) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Formatter) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Formatter) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *Formatter) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Formatter) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Formatter) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Formatter) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *Formatter) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *Formatter) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetFields

`func (o *Formatter) GetFields() map[string]FormatterField`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *Formatter) GetFieldsOk() (*map[string]FormatterField, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *Formatter) SetFields(v map[string]FormatterField)`

SetFields sets Fields field to given value.

### HasFields

`func (o *Formatter) HasFields() bool`

HasFields returns a boolean if a field has been set.

### GetId

`func (o *Formatter) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Formatter) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Formatter) SetId(v string)`

SetId sets Id field to given value.


### GetScope

`func (o *Formatter) GetScope() MappingScope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *Formatter) GetScopeOk() (*MappingScope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *Formatter) SetScope(v MappingScope)`

SetScope sets Scope field to given value.

### HasScope

`func (o *Formatter) HasScope() bool`

HasScope returns a boolean if a field has been set.

### GetCreatedBy

`func (o *Formatter) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *Formatter) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *Formatter) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *Formatter) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Formatter) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Formatter) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Formatter) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Formatter) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *Formatter) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *Formatter) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *Formatter) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *Formatter) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *Formatter) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Formatter) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Formatter) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *Formatter) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


