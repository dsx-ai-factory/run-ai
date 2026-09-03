# FormatterCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The unique name of the formatter | 
**Description** | Pointer to **NullableString** | A free-text description of the element. | [optional] 
**Fields** | Pointer to [**map[string]FormatterField**](FormatterField.md) | A map of the formatter fields. Map **keys** are field names. OpenAPI &#x60;additionalProperties&#x60; only constrains the **value** shape (&#x60;FormatterField&#x60;); it does **not** validate keys. The API enforces each key to match &#x60;MappingFieldName&#x60; (1–63 chars, &#x60;^[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?$&#x60; — alphanumeric with optional internal hyphens; no leading/trailing hyphen).  | [optional] 

## Methods

### NewFormatterCreationRequest

`func NewFormatterCreationRequest(name string, ) *FormatterCreationRequest`

NewFormatterCreationRequest instantiates a new FormatterCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFormatterCreationRequestWithDefaults

`func NewFormatterCreationRequestWithDefaults() *FormatterCreationRequest`

NewFormatterCreationRequestWithDefaults instantiates a new FormatterCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *FormatterCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *FormatterCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *FormatterCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *FormatterCreationRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *FormatterCreationRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *FormatterCreationRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *FormatterCreationRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *FormatterCreationRequest) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *FormatterCreationRequest) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetFields

`func (o *FormatterCreationRequest) GetFields() map[string]FormatterField`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *FormatterCreationRequest) GetFieldsOk() (*map[string]FormatterField, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *FormatterCreationRequest) SetFields(v map[string]FormatterField)`

SetFields sets Fields field to given value.

### HasFields

`func (o *FormatterCreationRequest) HasFields() bool`

HasFields returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


