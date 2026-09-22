# DistributedTemplatesListResponseItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateMetaFields**](TemplateMetaFields.md) |  | 
**Spec** | Pointer to [**DistributedTemplateSpec**](DistributedTemplateSpec.md) |  | [optional] 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**MasterSpecSameAsWorker** | Pointer to **NullableBool** | used for distributed workloads to indicate that the master spec should be the same as the worker spec. in this case, masterSpec should not be specified. | [optional] 
**Master** | Pointer to [**NullableTemplateMasterSection**](TemplateMasterSection.md) |  | [optional] 
**Runnable** | Pointer to [**RunnableInfo**](RunnableInfo.md) |  | [optional] 
**CombinedSpec** | Pointer to [**DistributedSpecSpec**](DistributedSpecSpec.md) |  | [optional] 
**CombinedMasterSpec** | Pointer to [**NullableMasterTemplateSpec**](MasterTemplateSpec.md) |  | [optional] 
**ComplianceIssues** | Pointer to [**[]ComplianceInfoReason**](ComplianceInfoReason.md) | A list of issues preventing the use of a template or asset for creating workloads in a given project. | [optional] 

## Methods

### NewDistributedTemplatesListResponseItem

`func NewDistributedTemplatesListResponseItem(meta TemplateMetaFields, ) *DistributedTemplatesListResponseItem`

NewDistributedTemplatesListResponseItem instantiates a new DistributedTemplatesListResponseItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedTemplatesListResponseItemWithDefaults

`func NewDistributedTemplatesListResponseItemWithDefaults() *DistributedTemplatesListResponseItem`

NewDistributedTemplatesListResponseItemWithDefaults instantiates a new DistributedTemplatesListResponseItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *DistributedTemplatesListResponseItem) GetMeta() TemplateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *DistributedTemplatesListResponseItem) GetMetaOk() (*TemplateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *DistributedTemplatesListResponseItem) SetMeta(v TemplateMetaFields)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *DistributedTemplatesListResponseItem) GetSpec() DistributedTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DistributedTemplatesListResponseItem) GetSpecOk() (*DistributedTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DistributedTemplatesListResponseItem) SetSpec(v DistributedTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *DistributedTemplatesListResponseItem) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetAssets

`func (o *DistributedTemplatesListResponseItem) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *DistributedTemplatesListResponseItem) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *DistributedTemplatesListResponseItem) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *DistributedTemplatesListResponseItem) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetMasterSpecSameAsWorker

`func (o *DistributedTemplatesListResponseItem) GetMasterSpecSameAsWorker() bool`

GetMasterSpecSameAsWorker returns the MasterSpecSameAsWorker field if non-nil, zero value otherwise.

### GetMasterSpecSameAsWorkerOk

`func (o *DistributedTemplatesListResponseItem) GetMasterSpecSameAsWorkerOk() (*bool, bool)`

GetMasterSpecSameAsWorkerOk returns a tuple with the MasterSpecSameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpecSameAsWorker

`func (o *DistributedTemplatesListResponseItem) SetMasterSpecSameAsWorker(v bool)`

SetMasterSpecSameAsWorker sets MasterSpecSameAsWorker field to given value.

### HasMasterSpecSameAsWorker

`func (o *DistributedTemplatesListResponseItem) HasMasterSpecSameAsWorker() bool`

HasMasterSpecSameAsWorker returns a boolean if a field has been set.

### SetMasterSpecSameAsWorkerNil

`func (o *DistributedTemplatesListResponseItem) SetMasterSpecSameAsWorkerNil(b bool)`

 SetMasterSpecSameAsWorkerNil sets the value for MasterSpecSameAsWorker to be an explicit nil

### UnsetMasterSpecSameAsWorker
`func (o *DistributedTemplatesListResponseItem) UnsetMasterSpecSameAsWorker()`

UnsetMasterSpecSameAsWorker ensures that no value is present for MasterSpecSameAsWorker, not even an explicit nil
### GetMaster

`func (o *DistributedTemplatesListResponseItem) GetMaster() TemplateMasterSection`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *DistributedTemplatesListResponseItem) GetMasterOk() (*TemplateMasterSection, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *DistributedTemplatesListResponseItem) SetMaster(v TemplateMasterSection)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *DistributedTemplatesListResponseItem) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *DistributedTemplatesListResponseItem) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *DistributedTemplatesListResponseItem) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil
### GetRunnable

`func (o *DistributedTemplatesListResponseItem) GetRunnable() RunnableInfo`

GetRunnable returns the Runnable field if non-nil, zero value otherwise.

### GetRunnableOk

`func (o *DistributedTemplatesListResponseItem) GetRunnableOk() (*RunnableInfo, bool)`

GetRunnableOk returns a tuple with the Runnable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnable

`func (o *DistributedTemplatesListResponseItem) SetRunnable(v RunnableInfo)`

SetRunnable sets Runnable field to given value.

### HasRunnable

`func (o *DistributedTemplatesListResponseItem) HasRunnable() bool`

HasRunnable returns a boolean if a field has been set.

### GetCombinedSpec

`func (o *DistributedTemplatesListResponseItem) GetCombinedSpec() DistributedSpecSpec`

GetCombinedSpec returns the CombinedSpec field if non-nil, zero value otherwise.

### GetCombinedSpecOk

`func (o *DistributedTemplatesListResponseItem) GetCombinedSpecOk() (*DistributedSpecSpec, bool)`

GetCombinedSpecOk returns a tuple with the CombinedSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedSpec

`func (o *DistributedTemplatesListResponseItem) SetCombinedSpec(v DistributedSpecSpec)`

SetCombinedSpec sets CombinedSpec field to given value.

### HasCombinedSpec

`func (o *DistributedTemplatesListResponseItem) HasCombinedSpec() bool`

HasCombinedSpec returns a boolean if a field has been set.

### GetCombinedMasterSpec

`func (o *DistributedTemplatesListResponseItem) GetCombinedMasterSpec() MasterTemplateSpec`

GetCombinedMasterSpec returns the CombinedMasterSpec field if non-nil, zero value otherwise.

### GetCombinedMasterSpecOk

`func (o *DistributedTemplatesListResponseItem) GetCombinedMasterSpecOk() (*MasterTemplateSpec, bool)`

GetCombinedMasterSpecOk returns a tuple with the CombinedMasterSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedMasterSpec

`func (o *DistributedTemplatesListResponseItem) SetCombinedMasterSpec(v MasterTemplateSpec)`

SetCombinedMasterSpec sets CombinedMasterSpec field to given value.

### HasCombinedMasterSpec

`func (o *DistributedTemplatesListResponseItem) HasCombinedMasterSpec() bool`

HasCombinedMasterSpec returns a boolean if a field has been set.

### SetCombinedMasterSpecNil

`func (o *DistributedTemplatesListResponseItem) SetCombinedMasterSpecNil(b bool)`

 SetCombinedMasterSpecNil sets the value for CombinedMasterSpec to be an explicit nil

### UnsetCombinedMasterSpec
`func (o *DistributedTemplatesListResponseItem) UnsetCombinedMasterSpec()`

UnsetCombinedMasterSpec ensures that no value is present for CombinedMasterSpec, not even an explicit nil
### GetComplianceIssues

`func (o *DistributedTemplatesListResponseItem) GetComplianceIssues() []ComplianceInfoReason`

GetComplianceIssues returns the ComplianceIssues field if non-nil, zero value otherwise.

### GetComplianceIssuesOk

`func (o *DistributedTemplatesListResponseItem) GetComplianceIssuesOk() (*[]ComplianceInfoReason, bool)`

GetComplianceIssuesOk returns a tuple with the ComplianceIssues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComplianceIssues

`func (o *DistributedTemplatesListResponseItem) SetComplianceIssues(v []ComplianceInfoReason)`

SetComplianceIssues sets ComplianceIssues field to given value.

### HasComplianceIssues

`func (o *DistributedTemplatesListResponseItem) HasComplianceIssues() bool`

HasComplianceIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


