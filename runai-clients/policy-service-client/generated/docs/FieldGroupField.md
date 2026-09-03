# FieldGroupField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **NullableString** | A free-text description of the element. | [optional] 
**Formatter** | Pointer to **string** | The name of the formatter used to read and write the field. | [optional] 
**SpecRef** | Pointer to **string** | A direct dotted path into the workload spec, used in place of a formatter for simple fields. For example, securityContext.privileged. | [optional] 
**Type** | Pointer to [**MappingFieldType**](MappingFieldType.md) | The data type of the field. | [optional] 
**ItemSelectors** | Pointer to **[]string** | The item selectors that apply to this field. | [optional] 

## Methods

### NewFieldGroupField

`func NewFieldGroupField() *FieldGroupField`

NewFieldGroupField instantiates a new FieldGroupField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFieldGroupFieldWithDefaults

`func NewFieldGroupFieldWithDefaults() *FieldGroupField`

NewFieldGroupFieldWithDefaults instantiates a new FieldGroupField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *FieldGroupField) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *FieldGroupField) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *FieldGroupField) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *FieldGroupField) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *FieldGroupField) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *FieldGroupField) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetFormatter

`func (o *FieldGroupField) GetFormatter() string`

GetFormatter returns the Formatter field if non-nil, zero value otherwise.

### GetFormatterOk

`func (o *FieldGroupField) GetFormatterOk() (*string, bool)`

GetFormatterOk returns a tuple with the Formatter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFormatter

`func (o *FieldGroupField) SetFormatter(v string)`

SetFormatter sets Formatter field to given value.

### HasFormatter

`func (o *FieldGroupField) HasFormatter() bool`

HasFormatter returns a boolean if a field has been set.

### GetSpecRef

`func (o *FieldGroupField) GetSpecRef() string`

GetSpecRef returns the SpecRef field if non-nil, zero value otherwise.

### GetSpecRefOk

`func (o *FieldGroupField) GetSpecRefOk() (*string, bool)`

GetSpecRefOk returns a tuple with the SpecRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecRef

`func (o *FieldGroupField) SetSpecRef(v string)`

SetSpecRef sets SpecRef field to given value.

### HasSpecRef

`func (o *FieldGroupField) HasSpecRef() bool`

HasSpecRef returns a boolean if a field has been set.

### GetType

`func (o *FieldGroupField) GetType() MappingFieldType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *FieldGroupField) GetTypeOk() (*MappingFieldType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *FieldGroupField) SetType(v MappingFieldType)`

SetType sets Type field to given value.

### HasType

`func (o *FieldGroupField) HasType() bool`

HasType returns a boolean if a field has been set.

### GetItemSelectors

`func (o *FieldGroupField) GetItemSelectors() []string`

GetItemSelectors returns the ItemSelectors field if non-nil, zero value otherwise.

### GetItemSelectorsOk

`func (o *FieldGroupField) GetItemSelectorsOk() (*[]string, bool)`

GetItemSelectorsOk returns a tuple with the ItemSelectors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItemSelectors

`func (o *FieldGroupField) SetItemSelectors(v []string)`

SetItemSelectors sets ItemSelectors field to given value.

### HasItemSelectors

`func (o *FieldGroupField) HasItemSelectors() bool`

HasItemSelectors returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


