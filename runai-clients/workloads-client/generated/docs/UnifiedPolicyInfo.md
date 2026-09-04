# UnifiedPolicyInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Worker** | Pointer to [**NullableUnifiedPolicyInfoPerReplicaType**](UnifiedPolicyInfoPerReplicaType.md) |  | [optional] 
**Master** | Pointer to [**NullableUnifiedPolicyInfoPerReplicaType**](UnifiedPolicyInfoPerReplicaType.md) |  | [optional] 

## Methods

### NewUnifiedPolicyInfo

`func NewUnifiedPolicyInfo() *UnifiedPolicyInfo`

NewUnifiedPolicyInfo instantiates a new UnifiedPolicyInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUnifiedPolicyInfoWithDefaults

`func NewUnifiedPolicyInfoWithDefaults() *UnifiedPolicyInfo`

NewUnifiedPolicyInfoWithDefaults instantiates a new UnifiedPolicyInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorker

`func (o *UnifiedPolicyInfo) GetWorker() UnifiedPolicyInfoPerReplicaType`

GetWorker returns the Worker field if non-nil, zero value otherwise.

### GetWorkerOk

`func (o *UnifiedPolicyInfo) GetWorkerOk() (*UnifiedPolicyInfoPerReplicaType, bool)`

GetWorkerOk returns a tuple with the Worker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorker

`func (o *UnifiedPolicyInfo) SetWorker(v UnifiedPolicyInfoPerReplicaType)`

SetWorker sets Worker field to given value.

### HasWorker

`func (o *UnifiedPolicyInfo) HasWorker() bool`

HasWorker returns a boolean if a field has been set.

### SetWorkerNil

`func (o *UnifiedPolicyInfo) SetWorkerNil(b bool)`

 SetWorkerNil sets the value for Worker to be an explicit nil

### UnsetWorker
`func (o *UnifiedPolicyInfo) UnsetWorker()`

UnsetWorker ensures that no value is present for Worker, not even an explicit nil
### GetMaster

`func (o *UnifiedPolicyInfo) GetMaster() UnifiedPolicyInfoPerReplicaType`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *UnifiedPolicyInfo) GetMasterOk() (*UnifiedPolicyInfoPerReplicaType, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *UnifiedPolicyInfo) SetMaster(v UnifiedPolicyInfoPerReplicaType)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *UnifiedPolicyInfo) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *UnifiedPolicyInfo) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *UnifiedPolicyInfo) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


