# WorkspaceTemplatesListResponseItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateMetaFields**](TemplateMetaFields.md) |  | 
**Spec** | Pointer to [**WorkspaceTemplateSpec**](WorkspaceTemplateSpec.md) |  | [optional] 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Runnable** | Pointer to [**RunnableInfo**](RunnableInfo.md) |  | [optional] 
**CombinedSpec** | Pointer to [**WorkspaceTemplateSpec**](WorkspaceTemplateSpec.md) |  | [optional] 
**ComplianceIssues** | Pointer to [**[]ComplianceInfoReason**](ComplianceInfoReason.md) | A list of issues preventing the use of a template or asset for creating workloads in a given project. | [optional] 

## Methods

### NewWorkspaceTemplatesListResponseItem

`func NewWorkspaceTemplatesListResponseItem(meta TemplateMetaFields, ) *WorkspaceTemplatesListResponseItem`

NewWorkspaceTemplatesListResponseItem instantiates a new WorkspaceTemplatesListResponseItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceTemplatesListResponseItemWithDefaults

`func NewWorkspaceTemplatesListResponseItemWithDefaults() *WorkspaceTemplatesListResponseItem`

NewWorkspaceTemplatesListResponseItemWithDefaults instantiates a new WorkspaceTemplatesListResponseItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *WorkspaceTemplatesListResponseItem) GetMeta() TemplateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *WorkspaceTemplatesListResponseItem) GetMetaOk() (*TemplateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *WorkspaceTemplatesListResponseItem) SetMeta(v TemplateMetaFields)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *WorkspaceTemplatesListResponseItem) GetSpec() WorkspaceTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *WorkspaceTemplatesListResponseItem) GetSpecOk() (*WorkspaceTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *WorkspaceTemplatesListResponseItem) SetSpec(v WorkspaceTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *WorkspaceTemplatesListResponseItem) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetAssets

`func (o *WorkspaceTemplatesListResponseItem) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *WorkspaceTemplatesListResponseItem) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *WorkspaceTemplatesListResponseItem) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *WorkspaceTemplatesListResponseItem) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetRunnable

`func (o *WorkspaceTemplatesListResponseItem) GetRunnable() RunnableInfo`

GetRunnable returns the Runnable field if non-nil, zero value otherwise.

### GetRunnableOk

`func (o *WorkspaceTemplatesListResponseItem) GetRunnableOk() (*RunnableInfo, bool)`

GetRunnableOk returns a tuple with the Runnable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnable

`func (o *WorkspaceTemplatesListResponseItem) SetRunnable(v RunnableInfo)`

SetRunnable sets Runnable field to given value.

### HasRunnable

`func (o *WorkspaceTemplatesListResponseItem) HasRunnable() bool`

HasRunnable returns a boolean if a field has been set.

### GetCombinedSpec

`func (o *WorkspaceTemplatesListResponseItem) GetCombinedSpec() WorkspaceTemplateSpec`

GetCombinedSpec returns the CombinedSpec field if non-nil, zero value otherwise.

### GetCombinedSpecOk

`func (o *WorkspaceTemplatesListResponseItem) GetCombinedSpecOk() (*WorkspaceTemplateSpec, bool)`

GetCombinedSpecOk returns a tuple with the CombinedSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedSpec

`func (o *WorkspaceTemplatesListResponseItem) SetCombinedSpec(v WorkspaceTemplateSpec)`

SetCombinedSpec sets CombinedSpec field to given value.

### HasCombinedSpec

`func (o *WorkspaceTemplatesListResponseItem) HasCombinedSpec() bool`

HasCombinedSpec returns a boolean if a field has been set.

### GetComplianceIssues

`func (o *WorkspaceTemplatesListResponseItem) GetComplianceIssues() []ComplianceInfoReason`

GetComplianceIssues returns the ComplianceIssues field if non-nil, zero value otherwise.

### GetComplianceIssuesOk

`func (o *WorkspaceTemplatesListResponseItem) GetComplianceIssuesOk() (*[]ComplianceInfoReason, bool)`

GetComplianceIssuesOk returns a tuple with the ComplianceIssues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComplianceIssues

`func (o *WorkspaceTemplatesListResponseItem) SetComplianceIssues(v []ComplianceInfoReason)`

SetComplianceIssues sets ComplianceIssues field to given value.

### HasComplianceIssues

`func (o *WorkspaceTemplatesListResponseItem) HasComplianceIssues() bool`

HasComplianceIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


