# WorkspaceCreationRequestV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**WorkloadCreationMeta**](WorkloadCreationMeta.md) |  | 
**Assets** | [**AssetsIds**](AssetsIds.md) |  | 
**SpecificRunParams** | Pointer to [**NullableWorkspaceSpecificRunParams**](WorkspaceSpecificRunParams.md) |  | [optional] 

## Methods

### NewWorkspaceCreationRequestV2

`func NewWorkspaceCreationRequestV2(meta WorkloadCreationMeta, assets AssetsIds, ) *WorkspaceCreationRequestV2`

NewWorkspaceCreationRequestV2 instantiates a new WorkspaceCreationRequestV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceCreationRequestV2WithDefaults

`func NewWorkspaceCreationRequestV2WithDefaults() *WorkspaceCreationRequestV2`

NewWorkspaceCreationRequestV2WithDefaults instantiates a new WorkspaceCreationRequestV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *WorkspaceCreationRequestV2) GetMeta() WorkloadCreationMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *WorkspaceCreationRequestV2) GetMetaOk() (*WorkloadCreationMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *WorkspaceCreationRequestV2) SetMeta(v WorkloadCreationMeta)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *WorkspaceCreationRequestV2) GetAssets() AssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *WorkspaceCreationRequestV2) GetAssetsOk() (*AssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *WorkspaceCreationRequestV2) SetAssets(v AssetsIds)`

SetAssets sets Assets field to given value.


### GetSpecificRunParams

`func (o *WorkspaceCreationRequestV2) GetSpecificRunParams() WorkspaceSpecificRunParams`

GetSpecificRunParams returns the SpecificRunParams field if non-nil, zero value otherwise.

### GetSpecificRunParamsOk

`func (o *WorkspaceCreationRequestV2) GetSpecificRunParamsOk() (*WorkspaceSpecificRunParams, bool)`

GetSpecificRunParamsOk returns a tuple with the SpecificRunParams field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecificRunParams

`func (o *WorkspaceCreationRequestV2) SetSpecificRunParams(v WorkspaceSpecificRunParams)`

SetSpecificRunParams sets SpecificRunParams field to given value.

### HasSpecificRunParams

`func (o *WorkspaceCreationRequestV2) HasSpecificRunParams() bool`

HasSpecificRunParams returns a boolean if a field has been set.

### SetSpecificRunParamsNil

`func (o *WorkspaceCreationRequestV2) SetSpecificRunParamsNil(b bool)`

 SetSpecificRunParamsNil sets the value for SpecificRunParams to be an explicit nil

### UnsetSpecificRunParams
`func (o *WorkspaceCreationRequestV2) UnsetSpecificRunParams()`

UnsetSpecificRunParams ensures that no value is present for SpecificRunParams, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


