# DistributedTemplatePatchRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**NullableTemplatePatchMetaFields**](TemplatePatchMetaFields.md) |  | [optional] 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**FieldsToClear** | Pointer to **[]string** |  | [optional] 
**Spec** | Pointer to [**DistributedTemplateSpec**](DistributedTemplateSpec.md) |  | [optional] 
**MasterSpecSameAsWorker** | Pointer to **NullableBool** | used for distributed workloads to indicate that the master spec should be the same as the worker spec. in this case, masterSpec should not be specified. | [optional] 
**Master** | Pointer to [**NullableTemplateMasterSection**](TemplateMasterSection.md) |  | [optional] 

## Methods

### NewDistributedTemplatePatchRequest

`func NewDistributedTemplatePatchRequest() *DistributedTemplatePatchRequest`

NewDistributedTemplatePatchRequest instantiates a new DistributedTemplatePatchRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedTemplatePatchRequestWithDefaults

`func NewDistributedTemplatePatchRequestWithDefaults() *DistributedTemplatePatchRequest`

NewDistributedTemplatePatchRequestWithDefaults instantiates a new DistributedTemplatePatchRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *DistributedTemplatePatchRequest) GetMeta() TemplatePatchMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *DistributedTemplatePatchRequest) GetMetaOk() (*TemplatePatchMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *DistributedTemplatePatchRequest) SetMeta(v TemplatePatchMetaFields)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *DistributedTemplatePatchRequest) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### SetMetaNil

`func (o *DistributedTemplatePatchRequest) SetMetaNil(b bool)`

 SetMetaNil sets the value for Meta to be an explicit nil

### UnsetMeta
`func (o *DistributedTemplatePatchRequest) UnsetMeta()`

UnsetMeta ensures that no value is present for Meta, not even an explicit nil
### GetAssets

`func (o *DistributedTemplatePatchRequest) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *DistributedTemplatePatchRequest) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *DistributedTemplatePatchRequest) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *DistributedTemplatePatchRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetFieldsToClear

`func (o *DistributedTemplatePatchRequest) GetFieldsToClear() []string`

GetFieldsToClear returns the FieldsToClear field if non-nil, zero value otherwise.

### GetFieldsToClearOk

`func (o *DistributedTemplatePatchRequest) GetFieldsToClearOk() (*[]string, bool)`

GetFieldsToClearOk returns a tuple with the FieldsToClear field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldsToClear

`func (o *DistributedTemplatePatchRequest) SetFieldsToClear(v []string)`

SetFieldsToClear sets FieldsToClear field to given value.

### HasFieldsToClear

`func (o *DistributedTemplatePatchRequest) HasFieldsToClear() bool`

HasFieldsToClear returns a boolean if a field has been set.

### GetSpec

`func (o *DistributedTemplatePatchRequest) GetSpec() DistributedTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DistributedTemplatePatchRequest) GetSpecOk() (*DistributedTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DistributedTemplatePatchRequest) SetSpec(v DistributedTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *DistributedTemplatePatchRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetMasterSpecSameAsWorker

`func (o *DistributedTemplatePatchRequest) GetMasterSpecSameAsWorker() bool`

GetMasterSpecSameAsWorker returns the MasterSpecSameAsWorker field if non-nil, zero value otherwise.

### GetMasterSpecSameAsWorkerOk

`func (o *DistributedTemplatePatchRequest) GetMasterSpecSameAsWorkerOk() (*bool, bool)`

GetMasterSpecSameAsWorkerOk returns a tuple with the MasterSpecSameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpecSameAsWorker

`func (o *DistributedTemplatePatchRequest) SetMasterSpecSameAsWorker(v bool)`

SetMasterSpecSameAsWorker sets MasterSpecSameAsWorker field to given value.

### HasMasterSpecSameAsWorker

`func (o *DistributedTemplatePatchRequest) HasMasterSpecSameAsWorker() bool`

HasMasterSpecSameAsWorker returns a boolean if a field has been set.

### SetMasterSpecSameAsWorkerNil

`func (o *DistributedTemplatePatchRequest) SetMasterSpecSameAsWorkerNil(b bool)`

 SetMasterSpecSameAsWorkerNil sets the value for MasterSpecSameAsWorker to be an explicit nil

### UnsetMasterSpecSameAsWorker
`func (o *DistributedTemplatePatchRequest) UnsetMasterSpecSameAsWorker()`

UnsetMasterSpecSameAsWorker ensures that no value is present for MasterSpecSameAsWorker, not even an explicit nil
### GetMaster

`func (o *DistributedTemplatePatchRequest) GetMaster() TemplateMasterSection`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *DistributedTemplatePatchRequest) GetMasterOk() (*TemplateMasterSection, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *DistributedTemplatePatchRequest) SetMaster(v TemplateMasterSection)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *DistributedTemplatePatchRequest) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *DistributedTemplatePatchRequest) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *DistributedTemplatePatchRequest) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


