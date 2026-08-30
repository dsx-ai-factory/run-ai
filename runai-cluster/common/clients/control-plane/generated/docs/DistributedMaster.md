# DistributedMaster

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SameAsWorker** | Pointer to **bool** | if true, indicates that the master should be identical to the worker. In this case assets and specificRunParams should not be provided. | [optional] 
**Assets** | Pointer to [**NullableOptionalAssetIds**](OptionalAssetIds.md) |  | [optional] 
**SpecificRunParams** | Pointer to [**NullableTrainingSpecificRunParams**](TrainingSpecificRunParams.md) |  | [optional] 

## Methods

### NewDistributedMaster

`func NewDistributedMaster() *DistributedMaster`

NewDistributedMaster instantiates a new DistributedMaster object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedMasterWithDefaults

`func NewDistributedMasterWithDefaults() *DistributedMaster`

NewDistributedMasterWithDefaults instantiates a new DistributedMaster object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSameAsWorker

`func (o *DistributedMaster) GetSameAsWorker() bool`

GetSameAsWorker returns the SameAsWorker field if non-nil, zero value otherwise.

### GetSameAsWorkerOk

`func (o *DistributedMaster) GetSameAsWorkerOk() (*bool, bool)`

GetSameAsWorkerOk returns a tuple with the SameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSameAsWorker

`func (o *DistributedMaster) SetSameAsWorker(v bool)`

SetSameAsWorker sets SameAsWorker field to given value.

### HasSameAsWorker

`func (o *DistributedMaster) HasSameAsWorker() bool`

HasSameAsWorker returns a boolean if a field has been set.

### GetAssets

`func (o *DistributedMaster) GetAssets() OptionalAssetIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *DistributedMaster) GetAssetsOk() (*OptionalAssetIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *DistributedMaster) SetAssets(v OptionalAssetIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *DistributedMaster) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### SetAssetsNil

`func (o *DistributedMaster) SetAssetsNil(b bool)`

 SetAssetsNil sets the value for Assets to be an explicit nil

### UnsetAssets
`func (o *DistributedMaster) UnsetAssets()`

UnsetAssets ensures that no value is present for Assets, not even an explicit nil
### GetSpecificRunParams

`func (o *DistributedMaster) GetSpecificRunParams() TrainingSpecificRunParams`

GetSpecificRunParams returns the SpecificRunParams field if non-nil, zero value otherwise.

### GetSpecificRunParamsOk

`func (o *DistributedMaster) GetSpecificRunParamsOk() (*TrainingSpecificRunParams, bool)`

GetSpecificRunParamsOk returns a tuple with the SpecificRunParams field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecificRunParams

`func (o *DistributedMaster) SetSpecificRunParams(v TrainingSpecificRunParams)`

SetSpecificRunParams sets SpecificRunParams field to given value.

### HasSpecificRunParams

`func (o *DistributedMaster) HasSpecificRunParams() bool`

HasSpecificRunParams returns a boolean if a field has been set.

### SetSpecificRunParamsNil

`func (o *DistributedMaster) SetSpecificRunParamsNil(b bool)`

 SetSpecificRunParamsNil sets the value for SpecificRunParams to be an explicit nil

### UnsetSpecificRunParams
`func (o *DistributedMaster) UnsetSpecificRunParams()`

UnsetSpecificRunParams ensures that no value is present for SpecificRunParams, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


