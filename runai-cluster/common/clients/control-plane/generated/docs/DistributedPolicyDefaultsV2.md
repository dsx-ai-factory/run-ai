# DistributedPolicyDefaultsV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Worker** | Pointer to [**NullableDistributedPolicyDefaultsV2Worker**](DistributedPolicyDefaultsV2Worker.md) |  | [optional] 
**Master** | Pointer to [**NullableReplicaDefaultsV2**](ReplicaDefaultsV2.md) |  | [optional] 

## Methods

### NewDistributedPolicyDefaultsV2

`func NewDistributedPolicyDefaultsV2() *DistributedPolicyDefaultsV2`

NewDistributedPolicyDefaultsV2 instantiates a new DistributedPolicyDefaultsV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedPolicyDefaultsV2WithDefaults

`func NewDistributedPolicyDefaultsV2WithDefaults() *DistributedPolicyDefaultsV2`

NewDistributedPolicyDefaultsV2WithDefaults instantiates a new DistributedPolicyDefaultsV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorker

`func (o *DistributedPolicyDefaultsV2) GetWorker() DistributedPolicyDefaultsV2Worker`

GetWorker returns the Worker field if non-nil, zero value otherwise.

### GetWorkerOk

`func (o *DistributedPolicyDefaultsV2) GetWorkerOk() (*DistributedPolicyDefaultsV2Worker, bool)`

GetWorkerOk returns a tuple with the Worker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorker

`func (o *DistributedPolicyDefaultsV2) SetWorker(v DistributedPolicyDefaultsV2Worker)`

SetWorker sets Worker field to given value.

### HasWorker

`func (o *DistributedPolicyDefaultsV2) HasWorker() bool`

HasWorker returns a boolean if a field has been set.

### SetWorkerNil

`func (o *DistributedPolicyDefaultsV2) SetWorkerNil(b bool)`

 SetWorkerNil sets the value for Worker to be an explicit nil

### UnsetWorker
`func (o *DistributedPolicyDefaultsV2) UnsetWorker()`

UnsetWorker ensures that no value is present for Worker, not even an explicit nil
### GetMaster

`func (o *DistributedPolicyDefaultsV2) GetMaster() ReplicaDefaultsV2`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *DistributedPolicyDefaultsV2) GetMasterOk() (*ReplicaDefaultsV2, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *DistributedPolicyDefaultsV2) SetMaster(v ReplicaDefaultsV2)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *DistributedPolicyDefaultsV2) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *DistributedPolicyDefaultsV2) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *DistributedPolicyDefaultsV2) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


