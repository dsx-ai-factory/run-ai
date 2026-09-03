# SupersetWorkloadAllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Spec** | Pointer to [**SupersetSpec**](SupersetSpec.md) |  | [optional] 
**MasterSpecSameAsWorker** | Pointer to **NullableBool** | used for distributed workloads to indicate that the master spec should be the same as the worker spec. in this case, masterSpec should not be specified. | [optional] 
**MasterSpec** | Pointer to [**NullableSupersetSpec**](SupersetSpec.md) |  | [optional] 

## Methods

### NewSupersetWorkloadAllOf

`func NewSupersetWorkloadAllOf() *SupersetWorkloadAllOf`

NewSupersetWorkloadAllOf instantiates a new SupersetWorkloadAllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSupersetWorkloadAllOfWithDefaults

`func NewSupersetWorkloadAllOfWithDefaults() *SupersetWorkloadAllOf`

NewSupersetWorkloadAllOfWithDefaults instantiates a new SupersetWorkloadAllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSpec

`func (o *SupersetWorkloadAllOf) GetSpec() SupersetSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *SupersetWorkloadAllOf) GetSpecOk() (*SupersetSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *SupersetWorkloadAllOf) SetSpec(v SupersetSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *SupersetWorkloadAllOf) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetMasterSpecSameAsWorker

`func (o *SupersetWorkloadAllOf) GetMasterSpecSameAsWorker() bool`

GetMasterSpecSameAsWorker returns the MasterSpecSameAsWorker field if non-nil, zero value otherwise.

### GetMasterSpecSameAsWorkerOk

`func (o *SupersetWorkloadAllOf) GetMasterSpecSameAsWorkerOk() (*bool, bool)`

GetMasterSpecSameAsWorkerOk returns a tuple with the MasterSpecSameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpecSameAsWorker

`func (o *SupersetWorkloadAllOf) SetMasterSpecSameAsWorker(v bool)`

SetMasterSpecSameAsWorker sets MasterSpecSameAsWorker field to given value.

### HasMasterSpecSameAsWorker

`func (o *SupersetWorkloadAllOf) HasMasterSpecSameAsWorker() bool`

HasMasterSpecSameAsWorker returns a boolean if a field has been set.

### SetMasterSpecSameAsWorkerNil

`func (o *SupersetWorkloadAllOf) SetMasterSpecSameAsWorkerNil(b bool)`

 SetMasterSpecSameAsWorkerNil sets the value for MasterSpecSameAsWorker to be an explicit nil

### UnsetMasterSpecSameAsWorker
`func (o *SupersetWorkloadAllOf) UnsetMasterSpecSameAsWorker()`

UnsetMasterSpecSameAsWorker ensures that no value is present for MasterSpecSameAsWorker, not even an explicit nil
### GetMasterSpec

`func (o *SupersetWorkloadAllOf) GetMasterSpec() SupersetSpec`

GetMasterSpec returns the MasterSpec field if non-nil, zero value otherwise.

### GetMasterSpecOk

`func (o *SupersetWorkloadAllOf) GetMasterSpecOk() (*SupersetSpec, bool)`

GetMasterSpecOk returns a tuple with the MasterSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpec

`func (o *SupersetWorkloadAllOf) SetMasterSpec(v SupersetSpec)`

SetMasterSpec sets MasterSpec field to given value.

### HasMasterSpec

`func (o *SupersetWorkloadAllOf) HasMasterSpec() bool`

HasMasterSpec returns a boolean if a field has been set.

### SetMasterSpecNil

`func (o *SupersetWorkloadAllOf) SetMasterSpecNil(b bool)`

 SetMasterSpecNil sets the value for MasterSpec to be an explicit nil

### UnsetMasterSpec
`func (o *SupersetWorkloadAllOf) UnsetMasterSpec()`

UnsetMasterSpec ensures that no value is present for MasterSpec, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


