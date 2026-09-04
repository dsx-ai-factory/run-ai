# WorkspaceTemplate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateMetaFields**](TemplateMetaFields.md) |  | 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Spec** | Pointer to [**WorkspaceTemplateSpec**](WorkspaceTemplateSpec.md) |  | [optional] 
**Runnable** | Pointer to [**RunnableInfo**](RunnableInfo.md) |  | [optional] 
**CombinedSpec** | Pointer to [**WorkspaceTemplateSpec**](WorkspaceTemplateSpec.md) |  | [optional] 

## Methods

### NewWorkspaceTemplate

`func NewWorkspaceTemplate(meta TemplateMetaFields, ) *WorkspaceTemplate`

NewWorkspaceTemplate instantiates a new WorkspaceTemplate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceTemplateWithDefaults

`func NewWorkspaceTemplateWithDefaults() *WorkspaceTemplate`

NewWorkspaceTemplateWithDefaults instantiates a new WorkspaceTemplate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *WorkspaceTemplate) GetMeta() TemplateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *WorkspaceTemplate) GetMetaOk() (*TemplateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *WorkspaceTemplate) SetMeta(v TemplateMetaFields)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *WorkspaceTemplate) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *WorkspaceTemplate) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *WorkspaceTemplate) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *WorkspaceTemplate) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpec

`func (o *WorkspaceTemplate) GetSpec() WorkspaceTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *WorkspaceTemplate) GetSpecOk() (*WorkspaceTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *WorkspaceTemplate) SetSpec(v WorkspaceTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *WorkspaceTemplate) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetRunnable

`func (o *WorkspaceTemplate) GetRunnable() RunnableInfo`

GetRunnable returns the Runnable field if non-nil, zero value otherwise.

### GetRunnableOk

`func (o *WorkspaceTemplate) GetRunnableOk() (*RunnableInfo, bool)`

GetRunnableOk returns a tuple with the Runnable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnable

`func (o *WorkspaceTemplate) SetRunnable(v RunnableInfo)`

SetRunnable sets Runnable field to given value.

### HasRunnable

`func (o *WorkspaceTemplate) HasRunnable() bool`

HasRunnable returns a boolean if a field has been set.

### GetCombinedSpec

`func (o *WorkspaceTemplate) GetCombinedSpec() WorkspaceTemplateSpec`

GetCombinedSpec returns the CombinedSpec field if non-nil, zero value otherwise.

### GetCombinedSpecOk

`func (o *WorkspaceTemplate) GetCombinedSpecOk() (*WorkspaceTemplateSpec, bool)`

GetCombinedSpecOk returns a tuple with the CombinedSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedSpec

`func (o *WorkspaceTemplate) SetCombinedSpec(v WorkspaceTemplateSpec)`

SetCombinedSpec sets CombinedSpec field to given value.

### HasCombinedSpec

`func (o *WorkspaceTemplate) HasCombinedSpec() bool`

HasCombinedSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


