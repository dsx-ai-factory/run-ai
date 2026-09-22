# SupersetTemplate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateMetaFields**](TemplateMetaFields.md) |  | 
**ComplianceIssues** | Pointer to [**[]ComplianceInfoReason**](ComplianceInfoReason.md) | A list of issues preventing the use of a template or asset for creating workloads in a given project. | [optional] 
**Runnable** | Pointer to [**RunnableInfo**](RunnableInfo.md) |  | [optional] 
**Spec** | Pointer to [**SupersetSpec**](SupersetSpec.md) |  | [optional] 
**Assets** | Pointer to [**NullableTemplateAssetsIdsOptional**](TemplateAssetsIdsOptional.md) |  | [optional] 
**MasterSpecSameAsWorker** | Pointer to **NullableBool** | used for distributed workloads to indicate that the master spec should be the same as the worker spec. in this case, masterSpec should not be specified. | [optional] 
**Master** | Pointer to [**NullableTemplateMasterSection**](TemplateMasterSection.md) |  | [optional] 
**CombinedSpec** | Pointer to [**SupersetSpec**](SupersetSpec.md) |  | [optional] 
**CombinedMasterSpec** | Pointer to [**NullableMasterTemplateSpec**](MasterTemplateSpec.md) |  | [optional] 

## Methods

### NewSupersetTemplate

`func NewSupersetTemplate(meta TemplateMetaFields, ) *SupersetTemplate`

NewSupersetTemplate instantiates a new SupersetTemplate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSupersetTemplateWithDefaults

`func NewSupersetTemplateWithDefaults() *SupersetTemplate`

NewSupersetTemplateWithDefaults instantiates a new SupersetTemplate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *SupersetTemplate) GetMeta() TemplateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *SupersetTemplate) GetMetaOk() (*TemplateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *SupersetTemplate) SetMeta(v TemplateMetaFields)`

SetMeta sets Meta field to given value.


### GetComplianceIssues

`func (o *SupersetTemplate) GetComplianceIssues() []ComplianceInfoReason`

GetComplianceIssues returns the ComplianceIssues field if non-nil, zero value otherwise.

### GetComplianceIssuesOk

`func (o *SupersetTemplate) GetComplianceIssuesOk() (*[]ComplianceInfoReason, bool)`

GetComplianceIssuesOk returns a tuple with the ComplianceIssues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComplianceIssues

`func (o *SupersetTemplate) SetComplianceIssues(v []ComplianceInfoReason)`

SetComplianceIssues sets ComplianceIssues field to given value.

### HasComplianceIssues

`func (o *SupersetTemplate) HasComplianceIssues() bool`

HasComplianceIssues returns a boolean if a field has been set.

### GetRunnable

`func (o *SupersetTemplate) GetRunnable() RunnableInfo`

GetRunnable returns the Runnable field if non-nil, zero value otherwise.

### GetRunnableOk

`func (o *SupersetTemplate) GetRunnableOk() (*RunnableInfo, bool)`

GetRunnableOk returns a tuple with the Runnable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnable

`func (o *SupersetTemplate) SetRunnable(v RunnableInfo)`

SetRunnable sets Runnable field to given value.

### HasRunnable

`func (o *SupersetTemplate) HasRunnable() bool`

HasRunnable returns a boolean if a field has been set.

### GetSpec

`func (o *SupersetTemplate) GetSpec() SupersetSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *SupersetTemplate) GetSpecOk() (*SupersetSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *SupersetTemplate) SetSpec(v SupersetSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *SupersetTemplate) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetAssets

`func (o *SupersetTemplate) GetAssets() TemplateAssetsIdsOptional`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *SupersetTemplate) GetAssetsOk() (*TemplateAssetsIdsOptional, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *SupersetTemplate) SetAssets(v TemplateAssetsIdsOptional)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *SupersetTemplate) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### SetAssetsNil

`func (o *SupersetTemplate) SetAssetsNil(b bool)`

 SetAssetsNil sets the value for Assets to be an explicit nil

### UnsetAssets
`func (o *SupersetTemplate) UnsetAssets()`

UnsetAssets ensures that no value is present for Assets, not even an explicit nil
### GetMasterSpecSameAsWorker

`func (o *SupersetTemplate) GetMasterSpecSameAsWorker() bool`

GetMasterSpecSameAsWorker returns the MasterSpecSameAsWorker field if non-nil, zero value otherwise.

### GetMasterSpecSameAsWorkerOk

`func (o *SupersetTemplate) GetMasterSpecSameAsWorkerOk() (*bool, bool)`

GetMasterSpecSameAsWorkerOk returns a tuple with the MasterSpecSameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpecSameAsWorker

`func (o *SupersetTemplate) SetMasterSpecSameAsWorker(v bool)`

SetMasterSpecSameAsWorker sets MasterSpecSameAsWorker field to given value.

### HasMasterSpecSameAsWorker

`func (o *SupersetTemplate) HasMasterSpecSameAsWorker() bool`

HasMasterSpecSameAsWorker returns a boolean if a field has been set.

### SetMasterSpecSameAsWorkerNil

`func (o *SupersetTemplate) SetMasterSpecSameAsWorkerNil(b bool)`

 SetMasterSpecSameAsWorkerNil sets the value for MasterSpecSameAsWorker to be an explicit nil

### UnsetMasterSpecSameAsWorker
`func (o *SupersetTemplate) UnsetMasterSpecSameAsWorker()`

UnsetMasterSpecSameAsWorker ensures that no value is present for MasterSpecSameAsWorker, not even an explicit nil
### GetMaster

`func (o *SupersetTemplate) GetMaster() TemplateMasterSection`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *SupersetTemplate) GetMasterOk() (*TemplateMasterSection, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *SupersetTemplate) SetMaster(v TemplateMasterSection)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *SupersetTemplate) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *SupersetTemplate) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *SupersetTemplate) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil
### GetCombinedSpec

`func (o *SupersetTemplate) GetCombinedSpec() SupersetSpec`

GetCombinedSpec returns the CombinedSpec field if non-nil, zero value otherwise.

### GetCombinedSpecOk

`func (o *SupersetTemplate) GetCombinedSpecOk() (*SupersetSpec, bool)`

GetCombinedSpecOk returns a tuple with the CombinedSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedSpec

`func (o *SupersetTemplate) SetCombinedSpec(v SupersetSpec)`

SetCombinedSpec sets CombinedSpec field to given value.

### HasCombinedSpec

`func (o *SupersetTemplate) HasCombinedSpec() bool`

HasCombinedSpec returns a boolean if a field has been set.

### GetCombinedMasterSpec

`func (o *SupersetTemplate) GetCombinedMasterSpec() MasterTemplateSpec`

GetCombinedMasterSpec returns the CombinedMasterSpec field if non-nil, zero value otherwise.

### GetCombinedMasterSpecOk

`func (o *SupersetTemplate) GetCombinedMasterSpecOk() (*MasterTemplateSpec, bool)`

GetCombinedMasterSpecOk returns a tuple with the CombinedMasterSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedMasterSpec

`func (o *SupersetTemplate) SetCombinedMasterSpec(v MasterTemplateSpec)`

SetCombinedMasterSpec sets CombinedMasterSpec field to given value.

### HasCombinedMasterSpec

`func (o *SupersetTemplate) HasCombinedMasterSpec() bool`

HasCombinedMasterSpec returns a boolean if a field has been set.

### SetCombinedMasterSpecNil

`func (o *SupersetTemplate) SetCombinedMasterSpecNil(b bool)`

 SetCombinedMasterSpecNil sets the value for CombinedMasterSpec to be an explicit nil

### UnsetCombinedMasterSpec
`func (o *SupersetTemplate) UnsetCombinedMasterSpec()`

UnsetCombinedMasterSpec ensures that no value is present for CombinedMasterSpec, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


