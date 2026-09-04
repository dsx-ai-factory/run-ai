# DistributedTemplate

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

## Methods

### NewDistributedTemplate

`func NewDistributedTemplate(meta TemplateMetaFields, ) *DistributedTemplate`

NewDistributedTemplate instantiates a new DistributedTemplate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedTemplateWithDefaults

`func NewDistributedTemplateWithDefaults() *DistributedTemplate`

NewDistributedTemplateWithDefaults instantiates a new DistributedTemplate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *DistributedTemplate) GetMeta() TemplateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *DistributedTemplate) GetMetaOk() (*TemplateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *DistributedTemplate) SetMeta(v TemplateMetaFields)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *DistributedTemplate) GetSpec() DistributedTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DistributedTemplate) GetSpecOk() (*DistributedTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DistributedTemplate) SetSpec(v DistributedTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *DistributedTemplate) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetAssets

`func (o *DistributedTemplate) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *DistributedTemplate) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *DistributedTemplate) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *DistributedTemplate) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetMasterSpecSameAsWorker

`func (o *DistributedTemplate) GetMasterSpecSameAsWorker() bool`

GetMasterSpecSameAsWorker returns the MasterSpecSameAsWorker field if non-nil, zero value otherwise.

### GetMasterSpecSameAsWorkerOk

`func (o *DistributedTemplate) GetMasterSpecSameAsWorkerOk() (*bool, bool)`

GetMasterSpecSameAsWorkerOk returns a tuple with the MasterSpecSameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpecSameAsWorker

`func (o *DistributedTemplate) SetMasterSpecSameAsWorker(v bool)`

SetMasterSpecSameAsWorker sets MasterSpecSameAsWorker field to given value.

### HasMasterSpecSameAsWorker

`func (o *DistributedTemplate) HasMasterSpecSameAsWorker() bool`

HasMasterSpecSameAsWorker returns a boolean if a field has been set.

### SetMasterSpecSameAsWorkerNil

`func (o *DistributedTemplate) SetMasterSpecSameAsWorkerNil(b bool)`

 SetMasterSpecSameAsWorkerNil sets the value for MasterSpecSameAsWorker to be an explicit nil

### UnsetMasterSpecSameAsWorker
`func (o *DistributedTemplate) UnsetMasterSpecSameAsWorker()`

UnsetMasterSpecSameAsWorker ensures that no value is present for MasterSpecSameAsWorker, not even an explicit nil
### GetMaster

`func (o *DistributedTemplate) GetMaster() TemplateMasterSection`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *DistributedTemplate) GetMasterOk() (*TemplateMasterSection, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *DistributedTemplate) SetMaster(v TemplateMasterSection)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *DistributedTemplate) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *DistributedTemplate) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *DistributedTemplate) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil
### GetRunnable

`func (o *DistributedTemplate) GetRunnable() RunnableInfo`

GetRunnable returns the Runnable field if non-nil, zero value otherwise.

### GetRunnableOk

`func (o *DistributedTemplate) GetRunnableOk() (*RunnableInfo, bool)`

GetRunnableOk returns a tuple with the Runnable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnable

`func (o *DistributedTemplate) SetRunnable(v RunnableInfo)`

SetRunnable sets Runnable field to given value.

### HasRunnable

`func (o *DistributedTemplate) HasRunnable() bool`

HasRunnable returns a boolean if a field has been set.

### GetCombinedSpec

`func (o *DistributedTemplate) GetCombinedSpec() DistributedSpecSpec`

GetCombinedSpec returns the CombinedSpec field if non-nil, zero value otherwise.

### GetCombinedSpecOk

`func (o *DistributedTemplate) GetCombinedSpecOk() (*DistributedSpecSpec, bool)`

GetCombinedSpecOk returns a tuple with the CombinedSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedSpec

`func (o *DistributedTemplate) SetCombinedSpec(v DistributedSpecSpec)`

SetCombinedSpec sets CombinedSpec field to given value.

### HasCombinedSpec

`func (o *DistributedTemplate) HasCombinedSpec() bool`

HasCombinedSpec returns a boolean if a field has been set.

### GetCombinedMasterSpec

`func (o *DistributedTemplate) GetCombinedMasterSpec() MasterTemplateSpec`

GetCombinedMasterSpec returns the CombinedMasterSpec field if non-nil, zero value otherwise.

### GetCombinedMasterSpecOk

`func (o *DistributedTemplate) GetCombinedMasterSpecOk() (*MasterTemplateSpec, bool)`

GetCombinedMasterSpecOk returns a tuple with the CombinedMasterSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedMasterSpec

`func (o *DistributedTemplate) SetCombinedMasterSpec(v MasterTemplateSpec)`

SetCombinedMasterSpec sets CombinedMasterSpec field to given value.

### HasCombinedMasterSpec

`func (o *DistributedTemplate) HasCombinedMasterSpec() bool`

HasCombinedMasterSpec returns a boolean if a field has been set.

### SetCombinedMasterSpecNil

`func (o *DistributedTemplate) SetCombinedMasterSpecNil(b bool)`

 SetCombinedMasterSpecNil sets the value for CombinedMasterSpec to be an explicit nil

### UnsetCombinedMasterSpec
`func (o *DistributedTemplate) UnsetCombinedMasterSpec()`

UnsetCombinedMasterSpec ensures that no value is present for CombinedMasterSpec, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


