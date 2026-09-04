# DistributedPolicyRulesV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Master** | Pointer to [**NullableReplicaRulesV2**](ReplicaRulesV2.md) |  | [optional] 
**Worker** | Pointer to [**NullableDistributedPolicyRulesV2Worker**](DistributedPolicyRulesV2Worker.md) |  | [optional] 

## Methods

### NewDistributedPolicyRulesV2

`func NewDistributedPolicyRulesV2() *DistributedPolicyRulesV2`

NewDistributedPolicyRulesV2 instantiates a new DistributedPolicyRulesV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedPolicyRulesV2WithDefaults

`func NewDistributedPolicyRulesV2WithDefaults() *DistributedPolicyRulesV2`

NewDistributedPolicyRulesV2WithDefaults instantiates a new DistributedPolicyRulesV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMaster

`func (o *DistributedPolicyRulesV2) GetMaster() ReplicaRulesV2`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *DistributedPolicyRulesV2) GetMasterOk() (*ReplicaRulesV2, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *DistributedPolicyRulesV2) SetMaster(v ReplicaRulesV2)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *DistributedPolicyRulesV2) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *DistributedPolicyRulesV2) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *DistributedPolicyRulesV2) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil
### GetWorker

`func (o *DistributedPolicyRulesV2) GetWorker() DistributedPolicyRulesV2Worker`

GetWorker returns the Worker field if non-nil, zero value otherwise.

### GetWorkerOk

`func (o *DistributedPolicyRulesV2) GetWorkerOk() (*DistributedPolicyRulesV2Worker, bool)`

GetWorkerOk returns a tuple with the Worker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorker

`func (o *DistributedPolicyRulesV2) SetWorker(v DistributedPolicyRulesV2Worker)`

SetWorker sets Worker field to given value.

### HasWorker

`func (o *DistributedPolicyRulesV2) HasWorker() bool`

HasWorker returns a boolean if a field has been set.

### SetWorkerNil

`func (o *DistributedPolicyRulesV2) SetWorkerNil(b bool)`

 SetWorkerNil sets the value for Worker to be an explicit nil

### UnsetWorker
`func (o *DistributedPolicyRulesV2) UnsetWorker()`

UnsetWorker ensures that no value is present for Worker, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


