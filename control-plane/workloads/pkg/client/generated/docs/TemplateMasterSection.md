# TemplateMasterSection

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Spec** | Pointer to [**NullableMasterTemplateSpec**](MasterTemplateSpec.md) |  | [optional] 
**Assets** | Pointer to [**NullableTemplateAssetsIdsOptional**](TemplateAssetsIdsOptional.md) |  | [optional] 

## Methods

### NewTemplateMasterSection

`func NewTemplateMasterSection() *TemplateMasterSection`

NewTemplateMasterSection instantiates a new TemplateMasterSection object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTemplateMasterSectionWithDefaults

`func NewTemplateMasterSectionWithDefaults() *TemplateMasterSection`

NewTemplateMasterSectionWithDefaults instantiates a new TemplateMasterSection object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSpec

`func (o *TemplateMasterSection) GetSpec() MasterTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *TemplateMasterSection) GetSpecOk() (*MasterTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *TemplateMasterSection) SetSpec(v MasterTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *TemplateMasterSection) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### SetSpecNil

`func (o *TemplateMasterSection) SetSpecNil(b bool)`

 SetSpecNil sets the value for Spec to be an explicit nil

### UnsetSpec
`func (o *TemplateMasterSection) UnsetSpec()`

UnsetSpec ensures that no value is present for Spec, not even an explicit nil
### GetAssets

`func (o *TemplateMasterSection) GetAssets() TemplateAssetsIdsOptional`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *TemplateMasterSection) GetAssetsOk() (*TemplateAssetsIdsOptional, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *TemplateMasterSection) SetAssets(v TemplateAssetsIdsOptional)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *TemplateMasterSection) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### SetAssetsNil

`func (o *TemplateMasterSection) SetAssetsNil(b bool)`

 SetAssetsNil sets the value for Assets to be an explicit nil

### UnsetAssets
`func (o *TemplateMasterSection) UnsetAssets()`

UnsetAssets ensures that no value is present for Assets, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


