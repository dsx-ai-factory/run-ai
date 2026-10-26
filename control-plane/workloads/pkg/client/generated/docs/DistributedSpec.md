# DistributedSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Spec** | Pointer to [**DistributedSpecSpec**](DistributedSpecSpec.md) |  | [optional] 
**MasterSpecSameAsWorker** | Pointer to **NullableBool** | used for distributed workloads to indicate that the master spec should be the same as the worker spec. in this case, masterSpec should not be specified. | [optional] 
**MasterSpec** | Pointer to [**NullableMasterSpec**](MasterSpec.md) |  | [optional] 
**MasterAssets** | Pointer to [**NullableAssetSubmissionIds**](AssetSubmissionIds.md) |  | [optional] 

## Methods

### NewDistributedSpec

`func NewDistributedSpec() *DistributedSpec`

NewDistributedSpec instantiates a new DistributedSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedSpecWithDefaults

`func NewDistributedSpecWithDefaults() *DistributedSpec`

NewDistributedSpecWithDefaults instantiates a new DistributedSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSpec

`func (o *DistributedSpec) GetSpec() DistributedSpecSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DistributedSpec) GetSpecOk() (*DistributedSpecSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DistributedSpec) SetSpec(v DistributedSpecSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *DistributedSpec) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetMasterSpecSameAsWorker

`func (o *DistributedSpec) GetMasterSpecSameAsWorker() bool`

GetMasterSpecSameAsWorker returns the MasterSpecSameAsWorker field if non-nil, zero value otherwise.

### GetMasterSpecSameAsWorkerOk

`func (o *DistributedSpec) GetMasterSpecSameAsWorkerOk() (*bool, bool)`

GetMasterSpecSameAsWorkerOk returns a tuple with the MasterSpecSameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpecSameAsWorker

`func (o *DistributedSpec) SetMasterSpecSameAsWorker(v bool)`

SetMasterSpecSameAsWorker sets MasterSpecSameAsWorker field to given value.

### HasMasterSpecSameAsWorker

`func (o *DistributedSpec) HasMasterSpecSameAsWorker() bool`

HasMasterSpecSameAsWorker returns a boolean if a field has been set.

### SetMasterSpecSameAsWorkerNil

`func (o *DistributedSpec) SetMasterSpecSameAsWorkerNil(b bool)`

 SetMasterSpecSameAsWorkerNil sets the value for MasterSpecSameAsWorker to be an explicit nil

### UnsetMasterSpecSameAsWorker
`func (o *DistributedSpec) UnsetMasterSpecSameAsWorker()`

UnsetMasterSpecSameAsWorker ensures that no value is present for MasterSpecSameAsWorker, not even an explicit nil
### GetMasterSpec

`func (o *DistributedSpec) GetMasterSpec() MasterSpec`

GetMasterSpec returns the MasterSpec field if non-nil, zero value otherwise.

### GetMasterSpecOk

`func (o *DistributedSpec) GetMasterSpecOk() (*MasterSpec, bool)`

GetMasterSpecOk returns a tuple with the MasterSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpec

`func (o *DistributedSpec) SetMasterSpec(v MasterSpec)`

SetMasterSpec sets MasterSpec field to given value.

### HasMasterSpec

`func (o *DistributedSpec) HasMasterSpec() bool`

HasMasterSpec returns a boolean if a field has been set.

### SetMasterSpecNil

`func (o *DistributedSpec) SetMasterSpecNil(b bool)`

 SetMasterSpecNil sets the value for MasterSpec to be an explicit nil

### UnsetMasterSpec
`func (o *DistributedSpec) UnsetMasterSpec()`

UnsetMasterSpec ensures that no value is present for MasterSpec, not even an explicit nil
### GetMasterAssets

`func (o *DistributedSpec) GetMasterAssets() AssetSubmissionIds`

GetMasterAssets returns the MasterAssets field if non-nil, zero value otherwise.

### GetMasterAssetsOk

`func (o *DistributedSpec) GetMasterAssetsOk() (*AssetSubmissionIds, bool)`

GetMasterAssetsOk returns a tuple with the MasterAssets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterAssets

`func (o *DistributedSpec) SetMasterAssets(v AssetSubmissionIds)`

SetMasterAssets sets MasterAssets field to given value.

### HasMasterAssets

`func (o *DistributedSpec) HasMasterAssets() bool`

HasMasterAssets returns a boolean if a field has been set.

### SetMasterAssetsNil

`func (o *DistributedSpec) SetMasterAssetsNil(b bool)`

 SetMasterAssetsNil sets the value for MasterAssets to be an explicit nil

### UnsetMasterAssets
`func (o *DistributedSpec) UnsetMasterAssets()`

UnsetMasterAssets ensures that no value is present for MasterAssets, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


