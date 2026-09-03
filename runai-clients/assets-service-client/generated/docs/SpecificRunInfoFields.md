# SpecificRunInfoFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Assets** | [**AssetsRef**](AssetsRef.md) |  | 
**SpecificEnv** | Pointer to [**NullableSpecificRunParams**](SpecificRunParams.md) |  | [optional] 
**Distributed** | Pointer to [**NullableInfoDistributed**](InfoDistributed.md) |  | [optional] 

## Methods

### NewSpecificRunInfoFields

`func NewSpecificRunInfoFields(assets AssetsRef, ) *SpecificRunInfoFields`

NewSpecificRunInfoFields instantiates a new SpecificRunInfoFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSpecificRunInfoFieldsWithDefaults

`func NewSpecificRunInfoFieldsWithDefaults() *SpecificRunInfoFields`

NewSpecificRunInfoFieldsWithDefaults instantiates a new SpecificRunInfoFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAssets

`func (o *SpecificRunInfoFields) GetAssets() AssetsRef`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *SpecificRunInfoFields) GetAssetsOk() (*AssetsRef, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *SpecificRunInfoFields) SetAssets(v AssetsRef)`

SetAssets sets Assets field to given value.


### GetSpecificEnv

`func (o *SpecificRunInfoFields) GetSpecificEnv() SpecificRunParams`

GetSpecificEnv returns the SpecificEnv field if non-nil, zero value otherwise.

### GetSpecificEnvOk

`func (o *SpecificRunInfoFields) GetSpecificEnvOk() (*SpecificRunParams, bool)`

GetSpecificEnvOk returns a tuple with the SpecificEnv field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecificEnv

`func (o *SpecificRunInfoFields) SetSpecificEnv(v SpecificRunParams)`

SetSpecificEnv sets SpecificEnv field to given value.

### HasSpecificEnv

`func (o *SpecificRunInfoFields) HasSpecificEnv() bool`

HasSpecificEnv returns a boolean if a field has been set.

### SetSpecificEnvNil

`func (o *SpecificRunInfoFields) SetSpecificEnvNil(b bool)`

 SetSpecificEnvNil sets the value for SpecificEnv to be an explicit nil

### UnsetSpecificEnv
`func (o *SpecificRunInfoFields) UnsetSpecificEnv()`

UnsetSpecificEnv ensures that no value is present for SpecificEnv, not even an explicit nil
### GetDistributed

`func (o *SpecificRunInfoFields) GetDistributed() InfoDistributed`

GetDistributed returns the Distributed field if non-nil, zero value otherwise.

### GetDistributedOk

`func (o *SpecificRunInfoFields) GetDistributedOk() (*InfoDistributed, bool)`

GetDistributedOk returns a tuple with the Distributed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributed

`func (o *SpecificRunInfoFields) SetDistributed(v InfoDistributed)`

SetDistributed sets Distributed field to given value.

### HasDistributed

`func (o *SpecificRunInfoFields) HasDistributed() bool`

HasDistributed returns a boolean if a field has been set.

### SetDistributedNil

`func (o *SpecificRunInfoFields) SetDistributedNil(b bool)`

 SetDistributedNil sets the value for Distributed to be an explicit nil

### UnsetDistributed
`func (o *SpecificRunInfoFields) UnsetDistributed()`

UnsetDistributed ensures that no value is present for Distributed, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


