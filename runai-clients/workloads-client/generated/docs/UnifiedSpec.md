# UnifiedSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Worker** | Pointer to [**NullableSupersetSpec**](SupersetSpec.md) |  | [optional] 
**MasterSpecSameAsWorker** | Pointer to **NullableBool** | used for distributed workloads to indicate that the master spec should be the same as the worker spec. in this case, masterSpec should not be specified. | [optional] 
**Master** | Pointer to [**NullableMasterSpec**](MasterSpec.md) |  | [optional] 

## Methods

### NewUnifiedSpec

`func NewUnifiedSpec() *UnifiedSpec`

NewUnifiedSpec instantiates a new UnifiedSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUnifiedSpecWithDefaults

`func NewUnifiedSpecWithDefaults() *UnifiedSpec`

NewUnifiedSpecWithDefaults instantiates a new UnifiedSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorker

`func (o *UnifiedSpec) GetWorker() SupersetSpec`

GetWorker returns the Worker field if non-nil, zero value otherwise.

### GetWorkerOk

`func (o *UnifiedSpec) GetWorkerOk() (*SupersetSpec, bool)`

GetWorkerOk returns a tuple with the Worker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorker

`func (o *UnifiedSpec) SetWorker(v SupersetSpec)`

SetWorker sets Worker field to given value.

### HasWorker

`func (o *UnifiedSpec) HasWorker() bool`

HasWorker returns a boolean if a field has been set.

### SetWorkerNil

`func (o *UnifiedSpec) SetWorkerNil(b bool)`

 SetWorkerNil sets the value for Worker to be an explicit nil

### UnsetWorker
`func (o *UnifiedSpec) UnsetWorker()`

UnsetWorker ensures that no value is present for Worker, not even an explicit nil
### GetMasterSpecSameAsWorker

`func (o *UnifiedSpec) GetMasterSpecSameAsWorker() bool`

GetMasterSpecSameAsWorker returns the MasterSpecSameAsWorker field if non-nil, zero value otherwise.

### GetMasterSpecSameAsWorkerOk

`func (o *UnifiedSpec) GetMasterSpecSameAsWorkerOk() (*bool, bool)`

GetMasterSpecSameAsWorkerOk returns a tuple with the MasterSpecSameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpecSameAsWorker

`func (o *UnifiedSpec) SetMasterSpecSameAsWorker(v bool)`

SetMasterSpecSameAsWorker sets MasterSpecSameAsWorker field to given value.

### HasMasterSpecSameAsWorker

`func (o *UnifiedSpec) HasMasterSpecSameAsWorker() bool`

HasMasterSpecSameAsWorker returns a boolean if a field has been set.

### SetMasterSpecSameAsWorkerNil

`func (o *UnifiedSpec) SetMasterSpecSameAsWorkerNil(b bool)`

 SetMasterSpecSameAsWorkerNil sets the value for MasterSpecSameAsWorker to be an explicit nil

### UnsetMasterSpecSameAsWorker
`func (o *UnifiedSpec) UnsetMasterSpecSameAsWorker()`

UnsetMasterSpecSameAsWorker ensures that no value is present for MasterSpecSameAsWorker, not even an explicit nil
### GetMaster

`func (o *UnifiedSpec) GetMaster() MasterSpec`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *UnifiedSpec) GetMasterOk() (*MasterSpec, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *UnifiedSpec) SetMaster(v MasterSpec)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *UnifiedSpec) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *UnifiedSpec) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *UnifiedSpec) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


