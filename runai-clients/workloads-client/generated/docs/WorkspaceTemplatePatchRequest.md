# WorkspaceTemplatePatchRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**NullableTemplatePatchMetaFields**](TemplatePatchMetaFields.md) |  | [optional] 
**FieldsToClear** | Pointer to **[]string** |  | [optional] 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Spec** | Pointer to [**WorkspaceTemplateSpec**](WorkspaceTemplateSpec.md) |  | [optional] 

## Methods

### NewWorkspaceTemplatePatchRequest

`func NewWorkspaceTemplatePatchRequest() *WorkspaceTemplatePatchRequest`

NewWorkspaceTemplatePatchRequest instantiates a new WorkspaceTemplatePatchRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceTemplatePatchRequestWithDefaults

`func NewWorkspaceTemplatePatchRequestWithDefaults() *WorkspaceTemplatePatchRequest`

NewWorkspaceTemplatePatchRequestWithDefaults instantiates a new WorkspaceTemplatePatchRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *WorkspaceTemplatePatchRequest) GetMeta() TemplatePatchMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *WorkspaceTemplatePatchRequest) GetMetaOk() (*TemplatePatchMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *WorkspaceTemplatePatchRequest) SetMeta(v TemplatePatchMetaFields)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *WorkspaceTemplatePatchRequest) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### SetMetaNil

`func (o *WorkspaceTemplatePatchRequest) SetMetaNil(b bool)`

 SetMetaNil sets the value for Meta to be an explicit nil

### UnsetMeta
`func (o *WorkspaceTemplatePatchRequest) UnsetMeta()`

UnsetMeta ensures that no value is present for Meta, not even an explicit nil
### GetFieldsToClear

`func (o *WorkspaceTemplatePatchRequest) GetFieldsToClear() []string`

GetFieldsToClear returns the FieldsToClear field if non-nil, zero value otherwise.

### GetFieldsToClearOk

`func (o *WorkspaceTemplatePatchRequest) GetFieldsToClearOk() (*[]string, bool)`

GetFieldsToClearOk returns a tuple with the FieldsToClear field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldsToClear

`func (o *WorkspaceTemplatePatchRequest) SetFieldsToClear(v []string)`

SetFieldsToClear sets FieldsToClear field to given value.

### HasFieldsToClear

`func (o *WorkspaceTemplatePatchRequest) HasFieldsToClear() bool`

HasFieldsToClear returns a boolean if a field has been set.

### GetAssets

`func (o *WorkspaceTemplatePatchRequest) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *WorkspaceTemplatePatchRequest) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *WorkspaceTemplatePatchRequest) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *WorkspaceTemplatePatchRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpec

`func (o *WorkspaceTemplatePatchRequest) GetSpec() WorkspaceTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *WorkspaceTemplatePatchRequest) GetSpecOk() (*WorkspaceTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *WorkspaceTemplatePatchRequest) SetSpec(v WorkspaceTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *WorkspaceTemplatePatchRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


