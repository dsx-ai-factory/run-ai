# TemplateCreationMetaFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the template. | 
**Description** | Pointer to **NullableString** | A description of the template. | [optional] 
**ScopeType** | [**ScopeType**](ScopeType.md) |  | 
**ScopeId** | Pointer to **NullableString** | The id of the cluster, department or project, depending on the scope type. | [optional] 

## Methods

### NewTemplateCreationMetaFields

`func NewTemplateCreationMetaFields(name string, scopeType ScopeType, ) *TemplateCreationMetaFields`

NewTemplateCreationMetaFields instantiates a new TemplateCreationMetaFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTemplateCreationMetaFieldsWithDefaults

`func NewTemplateCreationMetaFieldsWithDefaults() *TemplateCreationMetaFields`

NewTemplateCreationMetaFieldsWithDefaults instantiates a new TemplateCreationMetaFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *TemplateCreationMetaFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TemplateCreationMetaFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TemplateCreationMetaFields) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *TemplateCreationMetaFields) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *TemplateCreationMetaFields) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *TemplateCreationMetaFields) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *TemplateCreationMetaFields) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *TemplateCreationMetaFields) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *TemplateCreationMetaFields) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetScopeType

`func (o *TemplateCreationMetaFields) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *TemplateCreationMetaFields) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *TemplateCreationMetaFields) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *TemplateCreationMetaFields) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *TemplateCreationMetaFields) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *TemplateCreationMetaFields) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.

### HasScopeId

`func (o *TemplateCreationMetaFields) HasScopeId() bool`

HasScopeId returns a boolean if a field has been set.

### SetScopeIdNil

`func (o *TemplateCreationMetaFields) SetScopeIdNil(b bool)`

 SetScopeIdNil sets the value for ScopeId to be an explicit nil

### UnsetScopeId
`func (o *TemplateCreationMetaFields) UnsetScopeId()`

UnsetScopeId ensures that no value is present for ScopeId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


