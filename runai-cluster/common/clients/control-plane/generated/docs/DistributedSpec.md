# DistributedSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Spec** | Pointer to [**DistributedSpecSpec**](DistributedSpecSpec.md) |  | [optional] 
**MasterSpecSameAsWorker** | Pointer to **NullableBool** | used for distributed workloads to indicate that the master spec should be the same as the worker spec. in this case, masterSpec should not be specified. | [optional] 
**MasterSpec** | Pointer to [**NullableDistributedSpecMasterSpec**](DistributedSpecMasterSpec.md) |  | [optional] 

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

`func (o *DistributedSpec) GetMasterSpec() DistributedSpecMasterSpec`

GetMasterSpec returns the MasterSpec field if non-nil, zero value otherwise.

### GetMasterSpecOk

`func (o *DistributedSpec) GetMasterSpecOk() (*DistributedSpecMasterSpec, bool)`

GetMasterSpecOk returns a tuple with the MasterSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpec

`func (o *DistributedSpec) SetMasterSpec(v DistributedSpecMasterSpec)`

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

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


