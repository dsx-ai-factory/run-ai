# DistributedTemplateUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateUpdateMetaFields**](TemplateUpdateMetaFields.md) |  | 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Spec** | Pointer to [**DistributedTemplateSpec**](DistributedTemplateSpec.md) |  | [optional] 
**MasterSpecSameAsWorker** | Pointer to **NullableBool** | used for distributed workloads to indicate that the master spec should be the same as the worker spec. in this case, masterSpec should not be specified. | [optional] 
**Master** | Pointer to [**NullableTemplateMasterSection**](TemplateMasterSection.md) |  | [optional] 

## Methods

### NewDistributedTemplateUpdateRequest

`func NewDistributedTemplateUpdateRequest(meta TemplateUpdateMetaFields, ) *DistributedTemplateUpdateRequest`

NewDistributedTemplateUpdateRequest instantiates a new DistributedTemplateUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedTemplateUpdateRequestWithDefaults

`func NewDistributedTemplateUpdateRequestWithDefaults() *DistributedTemplateUpdateRequest`

NewDistributedTemplateUpdateRequestWithDefaults instantiates a new DistributedTemplateUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *DistributedTemplateUpdateRequest) GetMeta() TemplateUpdateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *DistributedTemplateUpdateRequest) GetMetaOk() (*TemplateUpdateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *DistributedTemplateUpdateRequest) SetMeta(v TemplateUpdateMetaFields)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *DistributedTemplateUpdateRequest) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *DistributedTemplateUpdateRequest) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *DistributedTemplateUpdateRequest) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *DistributedTemplateUpdateRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpec

`func (o *DistributedTemplateUpdateRequest) GetSpec() DistributedTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DistributedTemplateUpdateRequest) GetSpecOk() (*DistributedTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DistributedTemplateUpdateRequest) SetSpec(v DistributedTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *DistributedTemplateUpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetMasterSpecSameAsWorker

`func (o *DistributedTemplateUpdateRequest) GetMasterSpecSameAsWorker() bool`

GetMasterSpecSameAsWorker returns the MasterSpecSameAsWorker field if non-nil, zero value otherwise.

### GetMasterSpecSameAsWorkerOk

`func (o *DistributedTemplateUpdateRequest) GetMasterSpecSameAsWorkerOk() (*bool, bool)`

GetMasterSpecSameAsWorkerOk returns a tuple with the MasterSpecSameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpecSameAsWorker

`func (o *DistributedTemplateUpdateRequest) SetMasterSpecSameAsWorker(v bool)`

SetMasterSpecSameAsWorker sets MasterSpecSameAsWorker field to given value.

### HasMasterSpecSameAsWorker

`func (o *DistributedTemplateUpdateRequest) HasMasterSpecSameAsWorker() bool`

HasMasterSpecSameAsWorker returns a boolean if a field has been set.

### SetMasterSpecSameAsWorkerNil

`func (o *DistributedTemplateUpdateRequest) SetMasterSpecSameAsWorkerNil(b bool)`

 SetMasterSpecSameAsWorkerNil sets the value for MasterSpecSameAsWorker to be an explicit nil

### UnsetMasterSpecSameAsWorker
`func (o *DistributedTemplateUpdateRequest) UnsetMasterSpecSameAsWorker()`

UnsetMasterSpecSameAsWorker ensures that no value is present for MasterSpecSameAsWorker, not even an explicit nil
### GetMaster

`func (o *DistributedTemplateUpdateRequest) GetMaster() TemplateMasterSection`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *DistributedTemplateUpdateRequest) GetMasterOk() (*TemplateMasterSection, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *DistributedTemplateUpdateRequest) SetMaster(v TemplateMasterSection)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *DistributedTemplateUpdateRequest) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *DistributedTemplateUpdateRequest) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *DistributedTemplateUpdateRequest) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


