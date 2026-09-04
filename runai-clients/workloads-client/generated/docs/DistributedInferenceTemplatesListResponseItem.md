# DistributedInferenceTemplatesListResponseItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateMetaFields**](TemplateMetaFields.md) |  | 
**Spec** | Pointer to [**DistributedInferenceTemplateSpec**](DistributedInferenceTemplateSpec.md) |  | [optional] 
**Runnable** | Pointer to [**RunnableInfo**](RunnableInfo.md) |  | [optional] 
**CombinedSpec** | Pointer to [**DistributedInferenceTemplateSpec**](DistributedInferenceTemplateSpec.md) |  | [optional] 
**ComplianceIssues** | Pointer to [**[]ComplianceInfoReason**](ComplianceInfoReason.md) | A list of issues preventing the use of a template or asset for creating workloads in a given project. | [optional] 

## Methods

### NewDistributedInferenceTemplatesListResponseItem

`func NewDistributedInferenceTemplatesListResponseItem(meta TemplateMetaFields, ) *DistributedInferenceTemplatesListResponseItem`

NewDistributedInferenceTemplatesListResponseItem instantiates a new DistributedInferenceTemplatesListResponseItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceTemplatesListResponseItemWithDefaults

`func NewDistributedInferenceTemplatesListResponseItemWithDefaults() *DistributedInferenceTemplatesListResponseItem`

NewDistributedInferenceTemplatesListResponseItemWithDefaults instantiates a new DistributedInferenceTemplatesListResponseItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *DistributedInferenceTemplatesListResponseItem) GetMeta() TemplateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *DistributedInferenceTemplatesListResponseItem) GetMetaOk() (*TemplateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *DistributedInferenceTemplatesListResponseItem) SetMeta(v TemplateMetaFields)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *DistributedInferenceTemplatesListResponseItem) GetSpec() DistributedInferenceTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DistributedInferenceTemplatesListResponseItem) GetSpecOk() (*DistributedInferenceTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DistributedInferenceTemplatesListResponseItem) SetSpec(v DistributedInferenceTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *DistributedInferenceTemplatesListResponseItem) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetRunnable

`func (o *DistributedInferenceTemplatesListResponseItem) GetRunnable() RunnableInfo`

GetRunnable returns the Runnable field if non-nil, zero value otherwise.

### GetRunnableOk

`func (o *DistributedInferenceTemplatesListResponseItem) GetRunnableOk() (*RunnableInfo, bool)`

GetRunnableOk returns a tuple with the Runnable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnable

`func (o *DistributedInferenceTemplatesListResponseItem) SetRunnable(v RunnableInfo)`

SetRunnable sets Runnable field to given value.

### HasRunnable

`func (o *DistributedInferenceTemplatesListResponseItem) HasRunnable() bool`

HasRunnable returns a boolean if a field has been set.

### GetCombinedSpec

`func (o *DistributedInferenceTemplatesListResponseItem) GetCombinedSpec() DistributedInferenceTemplateSpec`

GetCombinedSpec returns the CombinedSpec field if non-nil, zero value otherwise.

### GetCombinedSpecOk

`func (o *DistributedInferenceTemplatesListResponseItem) GetCombinedSpecOk() (*DistributedInferenceTemplateSpec, bool)`

GetCombinedSpecOk returns a tuple with the CombinedSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedSpec

`func (o *DistributedInferenceTemplatesListResponseItem) SetCombinedSpec(v DistributedInferenceTemplateSpec)`

SetCombinedSpec sets CombinedSpec field to given value.

### HasCombinedSpec

`func (o *DistributedInferenceTemplatesListResponseItem) HasCombinedSpec() bool`

HasCombinedSpec returns a boolean if a field has been set.

### GetComplianceIssues

`func (o *DistributedInferenceTemplatesListResponseItem) GetComplianceIssues() []ComplianceInfoReason`

GetComplianceIssues returns the ComplianceIssues field if non-nil, zero value otherwise.

### GetComplianceIssuesOk

`func (o *DistributedInferenceTemplatesListResponseItem) GetComplianceIssuesOk() (*[]ComplianceInfoReason, bool)`

GetComplianceIssuesOk returns a tuple with the ComplianceIssues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComplianceIssues

`func (o *DistributedInferenceTemplatesListResponseItem) SetComplianceIssues(v []ComplianceInfoReason)`

SetComplianceIssues sets ComplianceIssues field to given value.

### HasComplianceIssues

`func (o *DistributedInferenceTemplatesListResponseItem) HasComplianceIssues() bool`

HasComplianceIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


