# WorkspaceV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**WorkloadMeta**](WorkloadMeta.md) |  | 
**Assets** | [**AssetsRef**](AssetsRef.md) |  | 
**SpecificRunParams** | Pointer to [**NullableWorkspaceSpecificRunParams**](WorkspaceSpecificRunParams.md) |  | [optional] 

## Methods

### NewWorkspaceV2

`func NewWorkspaceV2(meta WorkloadMeta, assets AssetsRef, ) *WorkspaceV2`

NewWorkspaceV2 instantiates a new WorkspaceV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceV2WithDefaults

`func NewWorkspaceV2WithDefaults() *WorkspaceV2`

NewWorkspaceV2WithDefaults instantiates a new WorkspaceV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *WorkspaceV2) GetMeta() WorkloadMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *WorkspaceV2) GetMetaOk() (*WorkloadMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *WorkspaceV2) SetMeta(v WorkloadMeta)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *WorkspaceV2) GetAssets() AssetsRef`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *WorkspaceV2) GetAssetsOk() (*AssetsRef, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *WorkspaceV2) SetAssets(v AssetsRef)`

SetAssets sets Assets field to given value.


### GetSpecificRunParams

`func (o *WorkspaceV2) GetSpecificRunParams() WorkspaceSpecificRunParams`

GetSpecificRunParams returns the SpecificRunParams field if non-nil, zero value otherwise.

### GetSpecificRunParamsOk

`func (o *WorkspaceV2) GetSpecificRunParamsOk() (*WorkspaceSpecificRunParams, bool)`

GetSpecificRunParamsOk returns a tuple with the SpecificRunParams field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecificRunParams

`func (o *WorkspaceV2) SetSpecificRunParams(v WorkspaceSpecificRunParams)`

SetSpecificRunParams sets SpecificRunParams field to given value.

### HasSpecificRunParams

`func (o *WorkspaceV2) HasSpecificRunParams() bool`

HasSpecificRunParams returns a boolean if a field has been set.

### SetSpecificRunParamsNil

`func (o *WorkspaceV2) SetSpecificRunParamsNil(b bool)`

 SetSpecificRunParamsNil sets the value for SpecificRunParams to be an explicit nil

### UnsetSpecificRunParams
`func (o *WorkspaceV2) UnsetSpecificRunParams()`

UnsetSpecificRunParams ensures that no value is present for SpecificRunParams, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


