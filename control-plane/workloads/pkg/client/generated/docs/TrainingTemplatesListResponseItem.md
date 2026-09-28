# TrainingTemplatesListResponseItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateMetaFields**](TemplateMetaFields.md) |  | 
**Spec** | Pointer to [**TrainingTemplateSpec**](TrainingTemplateSpec.md) |  | [optional] 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Runnable** | Pointer to [**RunnableInfo**](RunnableInfo.md) |  | [optional] 
**CombinedSpec** | Pointer to [**TrainingTemplateSpec**](TrainingTemplateSpec.md) |  | [optional] 
**ComplianceIssues** | Pointer to [**[]ComplianceInfoReason**](ComplianceInfoReason.md) | A list of issues preventing the use of a template or asset for creating workloads in a given project. | [optional] 

## Methods

### NewTrainingTemplatesListResponseItem

`func NewTrainingTemplatesListResponseItem(meta TemplateMetaFields, ) *TrainingTemplatesListResponseItem`

NewTrainingTemplatesListResponseItem instantiates a new TrainingTemplatesListResponseItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingTemplatesListResponseItemWithDefaults

`func NewTrainingTemplatesListResponseItemWithDefaults() *TrainingTemplatesListResponseItem`

NewTrainingTemplatesListResponseItemWithDefaults instantiates a new TrainingTemplatesListResponseItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *TrainingTemplatesListResponseItem) GetMeta() TemplateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *TrainingTemplatesListResponseItem) GetMetaOk() (*TemplateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *TrainingTemplatesListResponseItem) SetMeta(v TemplateMetaFields)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *TrainingTemplatesListResponseItem) GetSpec() TrainingTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *TrainingTemplatesListResponseItem) GetSpecOk() (*TrainingTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *TrainingTemplatesListResponseItem) SetSpec(v TrainingTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *TrainingTemplatesListResponseItem) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetAssets

`func (o *TrainingTemplatesListResponseItem) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *TrainingTemplatesListResponseItem) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *TrainingTemplatesListResponseItem) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *TrainingTemplatesListResponseItem) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetRunnable

`func (o *TrainingTemplatesListResponseItem) GetRunnable() RunnableInfo`

GetRunnable returns the Runnable field if non-nil, zero value otherwise.

### GetRunnableOk

`func (o *TrainingTemplatesListResponseItem) GetRunnableOk() (*RunnableInfo, bool)`

GetRunnableOk returns a tuple with the Runnable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnable

`func (o *TrainingTemplatesListResponseItem) SetRunnable(v RunnableInfo)`

SetRunnable sets Runnable field to given value.

### HasRunnable

`func (o *TrainingTemplatesListResponseItem) HasRunnable() bool`

HasRunnable returns a boolean if a field has been set.

### GetCombinedSpec

`func (o *TrainingTemplatesListResponseItem) GetCombinedSpec() TrainingTemplateSpec`

GetCombinedSpec returns the CombinedSpec field if non-nil, zero value otherwise.

### GetCombinedSpecOk

`func (o *TrainingTemplatesListResponseItem) GetCombinedSpecOk() (*TrainingTemplateSpec, bool)`

GetCombinedSpecOk returns a tuple with the CombinedSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedSpec

`func (o *TrainingTemplatesListResponseItem) SetCombinedSpec(v TrainingTemplateSpec)`

SetCombinedSpec sets CombinedSpec field to given value.

### HasCombinedSpec

`func (o *TrainingTemplatesListResponseItem) HasCombinedSpec() bool`

HasCombinedSpec returns a boolean if a field has been set.

### GetComplianceIssues

`func (o *TrainingTemplatesListResponseItem) GetComplianceIssues() []ComplianceInfoReason`

GetComplianceIssues returns the ComplianceIssues field if non-nil, zero value otherwise.

### GetComplianceIssuesOk

`func (o *TrainingTemplatesListResponseItem) GetComplianceIssuesOk() (*[]ComplianceInfoReason, bool)`

GetComplianceIssuesOk returns a tuple with the ComplianceIssues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComplianceIssues

`func (o *TrainingTemplatesListResponseItem) SetComplianceIssues(v []ComplianceInfoReason)`

SetComplianceIssues sets ComplianceIssues field to given value.

### HasComplianceIssues

`func (o *TrainingTemplatesListResponseItem) HasComplianceIssues() bool`

HasComplianceIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


