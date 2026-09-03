# DistMaster

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Assets** | Pointer to [**AssetsIds**](AssetsIds.md) |  | [optional] 
**SpecificEnv** | Pointer to [**NullableSpecificRunParams**](SpecificRunParams.md) |  | [optional] 

## Methods

### NewDistMaster

`func NewDistMaster() *DistMaster`

NewDistMaster instantiates a new DistMaster object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistMasterWithDefaults

`func NewDistMasterWithDefaults() *DistMaster`

NewDistMasterWithDefaults instantiates a new DistMaster object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAssets

`func (o *DistMaster) GetAssets() AssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *DistMaster) GetAssetsOk() (*AssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *DistMaster) SetAssets(v AssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *DistMaster) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpecificEnv

`func (o *DistMaster) GetSpecificEnv() SpecificRunParams`

GetSpecificEnv returns the SpecificEnv field if non-nil, zero value otherwise.

### GetSpecificEnvOk

`func (o *DistMaster) GetSpecificEnvOk() (*SpecificRunParams, bool)`

GetSpecificEnvOk returns a tuple with the SpecificEnv field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecificEnv

`func (o *DistMaster) SetSpecificEnv(v SpecificRunParams)`

SetSpecificEnv sets SpecificEnv field to given value.

### HasSpecificEnv

`func (o *DistMaster) HasSpecificEnv() bool`

HasSpecificEnv returns a boolean if a field has been set.

### SetSpecificEnvNil

`func (o *DistMaster) SetSpecificEnvNil(b bool)`

 SetSpecificEnvNil sets the value for SpecificEnv to be an explicit nil

### UnsetSpecificEnv
`func (o *DistMaster) UnsetSpecificEnv()`

UnsetSpecificEnv ensures that no value is present for SpecificEnv, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


