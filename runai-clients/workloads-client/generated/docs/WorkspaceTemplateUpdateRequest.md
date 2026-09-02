# WorkspaceTemplateUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateUpdateMetaFields**](TemplateUpdateMetaFields.md) |  | 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Spec** | Pointer to [**WorkspaceTemplateSpec**](WorkspaceTemplateSpec.md) |  | [optional] 

## Methods

### NewWorkspaceTemplateUpdateRequest

`func NewWorkspaceTemplateUpdateRequest(meta TemplateUpdateMetaFields, ) *WorkspaceTemplateUpdateRequest`

NewWorkspaceTemplateUpdateRequest instantiates a new WorkspaceTemplateUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceTemplateUpdateRequestWithDefaults

`func NewWorkspaceTemplateUpdateRequestWithDefaults() *WorkspaceTemplateUpdateRequest`

NewWorkspaceTemplateUpdateRequestWithDefaults instantiates a new WorkspaceTemplateUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *WorkspaceTemplateUpdateRequest) GetMeta() TemplateUpdateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *WorkspaceTemplateUpdateRequest) GetMetaOk() (*TemplateUpdateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *WorkspaceTemplateUpdateRequest) SetMeta(v TemplateUpdateMetaFields)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *WorkspaceTemplateUpdateRequest) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *WorkspaceTemplateUpdateRequest) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *WorkspaceTemplateUpdateRequest) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *WorkspaceTemplateUpdateRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpec

`func (o *WorkspaceTemplateUpdateRequest) GetSpec() WorkspaceTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *WorkspaceTemplateUpdateRequest) GetSpecOk() (*WorkspaceTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *WorkspaceTemplateUpdateRequest) SetSpec(v WorkspaceTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *WorkspaceTemplateUpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


