# DistributedMasterData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SameAsWorker** | Pointer to **bool** | if true, indicates that the master should be identical to the worker. In this case assets and specificRunParams should not be provided. | [optional] 
**Assets** | Pointer to [**NullableAssetsRefOptional**](AssetsRefOptional.md) |  | [optional] 
**SpecificRunParams** | Pointer to [**NullableTrainingSpecificRunParams**](TrainingSpecificRunParams.md) |  | [optional] 

## Methods

### NewDistributedMasterData

`func NewDistributedMasterData() *DistributedMasterData`

NewDistributedMasterData instantiates a new DistributedMasterData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedMasterDataWithDefaults

`func NewDistributedMasterDataWithDefaults() *DistributedMasterData`

NewDistributedMasterDataWithDefaults instantiates a new DistributedMasterData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSameAsWorker

`func (o *DistributedMasterData) GetSameAsWorker() bool`

GetSameAsWorker returns the SameAsWorker field if non-nil, zero value otherwise.

### GetSameAsWorkerOk

`func (o *DistributedMasterData) GetSameAsWorkerOk() (*bool, bool)`

GetSameAsWorkerOk returns a tuple with the SameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSameAsWorker

`func (o *DistributedMasterData) SetSameAsWorker(v bool)`

SetSameAsWorker sets SameAsWorker field to given value.

### HasSameAsWorker

`func (o *DistributedMasterData) HasSameAsWorker() bool`

HasSameAsWorker returns a boolean if a field has been set.

### GetAssets

`func (o *DistributedMasterData) GetAssets() AssetsRefOptional`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *DistributedMasterData) GetAssetsOk() (*AssetsRefOptional, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *DistributedMasterData) SetAssets(v AssetsRefOptional)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *DistributedMasterData) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### SetAssetsNil

`func (o *DistributedMasterData) SetAssetsNil(b bool)`

 SetAssetsNil sets the value for Assets to be an explicit nil

### UnsetAssets
`func (o *DistributedMasterData) UnsetAssets()`

UnsetAssets ensures that no value is present for Assets, not even an explicit nil
### GetSpecificRunParams

`func (o *DistributedMasterData) GetSpecificRunParams() TrainingSpecificRunParams`

GetSpecificRunParams returns the SpecificRunParams field if non-nil, zero value otherwise.

### GetSpecificRunParamsOk

`func (o *DistributedMasterData) GetSpecificRunParamsOk() (*TrainingSpecificRunParams, bool)`

GetSpecificRunParamsOk returns a tuple with the SpecificRunParams field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecificRunParams

`func (o *DistributedMasterData) SetSpecificRunParams(v TrainingSpecificRunParams)`

SetSpecificRunParams sets SpecificRunParams field to given value.

### HasSpecificRunParams

`func (o *DistributedMasterData) HasSpecificRunParams() bool`

HasSpecificRunParams returns a boolean if a field has been set.

### SetSpecificRunParamsNil

`func (o *DistributedMasterData) SetSpecificRunParamsNil(b bool)`

 SetSpecificRunParamsNil sets the value for SpecificRunParams to be an explicit nil

### UnsetSpecificRunParams
`func (o *DistributedMasterData) UnsetSpecificRunParams()`

UnsetSpecificRunParams ensures that no value is present for SpecificRunParams, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


