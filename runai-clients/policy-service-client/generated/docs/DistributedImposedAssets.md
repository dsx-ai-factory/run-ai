# DistributedImposedAssets

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Worker** | Pointer to **[]string** | a list of datasource assets to import into the policy. these datasources will be included in any workload that will be created in the scope of the policy. | [optional] 
**Master** | Pointer to **[]string** | a list of datasource assets to import into the policy. these datasources will be included in any workload that will be created in the scope of the policy. | [optional] 

## Methods

### NewDistributedImposedAssets

`func NewDistributedImposedAssets() *DistributedImposedAssets`

NewDistributedImposedAssets instantiates a new DistributedImposedAssets object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedImposedAssetsWithDefaults

`func NewDistributedImposedAssetsWithDefaults() *DistributedImposedAssets`

NewDistributedImposedAssetsWithDefaults instantiates a new DistributedImposedAssets object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorker

`func (o *DistributedImposedAssets) GetWorker() []string`

GetWorker returns the Worker field if non-nil, zero value otherwise.

### GetWorkerOk

`func (o *DistributedImposedAssets) GetWorkerOk() (*[]string, bool)`

GetWorkerOk returns a tuple with the Worker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorker

`func (o *DistributedImposedAssets) SetWorker(v []string)`

SetWorker sets Worker field to given value.

### HasWorker

`func (o *DistributedImposedAssets) HasWorker() bool`

HasWorker returns a boolean if a field has been set.

### SetWorkerNil

`func (o *DistributedImposedAssets) SetWorkerNil(b bool)`

 SetWorkerNil sets the value for Worker to be an explicit nil

### UnsetWorker
`func (o *DistributedImposedAssets) UnsetWorker()`

UnsetWorker ensures that no value is present for Worker, not even an explicit nil
### GetMaster

`func (o *DistributedImposedAssets) GetMaster() []string`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *DistributedImposedAssets) GetMasterOk() (*[]string, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *DistributedImposedAssets) SetMaster(v []string)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *DistributedImposedAssets) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *DistributedImposedAssets) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *DistributedImposedAssets) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


