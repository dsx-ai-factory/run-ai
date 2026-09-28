# InferenceTemplatesListResponseItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateMetaFields**](TemplateMetaFields.md) |  | 
**Spec** | Pointer to [**InferenceTemplateSpec**](InferenceTemplateSpec.md) |  | [optional] 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Runnable** | Pointer to [**RunnableInfo**](RunnableInfo.md) |  | [optional] 
**CombinedSpec** | Pointer to [**InferenceTemplateSpec**](InferenceTemplateSpec.md) |  | [optional] 
**ComplianceIssues** | Pointer to [**[]ComplianceInfoReason**](ComplianceInfoReason.md) | A list of issues preventing the use of a template or asset for creating workloads in a given project. | [optional] 

## Methods

### NewInferenceTemplatesListResponseItem

`func NewInferenceTemplatesListResponseItem(meta TemplateMetaFields, ) *InferenceTemplatesListResponseItem`

NewInferenceTemplatesListResponseItem instantiates a new InferenceTemplatesListResponseItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceTemplatesListResponseItemWithDefaults

`func NewInferenceTemplatesListResponseItemWithDefaults() *InferenceTemplatesListResponseItem`

NewInferenceTemplatesListResponseItemWithDefaults instantiates a new InferenceTemplatesListResponseItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *InferenceTemplatesListResponseItem) GetMeta() TemplateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *InferenceTemplatesListResponseItem) GetMetaOk() (*TemplateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *InferenceTemplatesListResponseItem) SetMeta(v TemplateMetaFields)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *InferenceTemplatesListResponseItem) GetSpec() InferenceTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *InferenceTemplatesListResponseItem) GetSpecOk() (*InferenceTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *InferenceTemplatesListResponseItem) SetSpec(v InferenceTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *InferenceTemplatesListResponseItem) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetAssets

`func (o *InferenceTemplatesListResponseItem) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *InferenceTemplatesListResponseItem) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *InferenceTemplatesListResponseItem) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *InferenceTemplatesListResponseItem) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetRunnable

`func (o *InferenceTemplatesListResponseItem) GetRunnable() RunnableInfo`

GetRunnable returns the Runnable field if non-nil, zero value otherwise.

### GetRunnableOk

`func (o *InferenceTemplatesListResponseItem) GetRunnableOk() (*RunnableInfo, bool)`

GetRunnableOk returns a tuple with the Runnable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnable

`func (o *InferenceTemplatesListResponseItem) SetRunnable(v RunnableInfo)`

SetRunnable sets Runnable field to given value.

### HasRunnable

`func (o *InferenceTemplatesListResponseItem) HasRunnable() bool`

HasRunnable returns a boolean if a field has been set.

### GetCombinedSpec

`func (o *InferenceTemplatesListResponseItem) GetCombinedSpec() InferenceTemplateSpec`

GetCombinedSpec returns the CombinedSpec field if non-nil, zero value otherwise.

### GetCombinedSpecOk

`func (o *InferenceTemplatesListResponseItem) GetCombinedSpecOk() (*InferenceTemplateSpec, bool)`

GetCombinedSpecOk returns a tuple with the CombinedSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedSpec

`func (o *InferenceTemplatesListResponseItem) SetCombinedSpec(v InferenceTemplateSpec)`

SetCombinedSpec sets CombinedSpec field to given value.

### HasCombinedSpec

`func (o *InferenceTemplatesListResponseItem) HasCombinedSpec() bool`

HasCombinedSpec returns a boolean if a field has been set.

### GetComplianceIssues

`func (o *InferenceTemplatesListResponseItem) GetComplianceIssues() []ComplianceInfoReason`

GetComplianceIssues returns the ComplianceIssues field if non-nil, zero value otherwise.

### GetComplianceIssuesOk

`func (o *InferenceTemplatesListResponseItem) GetComplianceIssuesOk() (*[]ComplianceInfoReason, bool)`

GetComplianceIssuesOk returns a tuple with the ComplianceIssues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComplianceIssues

`func (o *InferenceTemplatesListResponseItem) SetComplianceIssues(v []ComplianceInfoReason)`

SetComplianceIssues sets ComplianceIssues field to given value.

### HasComplianceIssues

`func (o *InferenceTemplatesListResponseItem) HasComplianceIssues() bool`

HasComplianceIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


