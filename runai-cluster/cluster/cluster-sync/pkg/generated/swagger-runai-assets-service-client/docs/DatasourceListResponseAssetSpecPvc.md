# DatasourceListResponseAssetSpecPvc

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExistingPvc** | Pointer to **NullableBool** | Whether to assume that the PVC exists. If set to true, PVC is assumed to exist. If set to false, the PVC will be create if it does not exist. | [optional] [default to false]
**ClaimName** | **NullableString** | A given name for the PVC. Allowed referencing it across workspaces. ClaimName is mandatory for creating a workspace. | 
**Path** | **NullableString** | Local path within the workspace to which the PVC bucket will be mapped. Path is mandatory for creating a workspace. | 
**ReadOnly** | Pointer to **NullableBool** | Whether the path to the PVC permits only read access. | [optional] [default to false]
**Ephemeral** | Pointer to **NullableBool** | Whether the PVC is ephemeral. If set to true, the PVC will be deleted when the workspace is stopped. | [optional] [default to false]
**ClaimInfo** | Pointer to [**NullableClaimInfo**](ClaimInfo.md) |  | [optional] 

## Methods

### NewDatasourceListResponseAssetSpecPvc

`func NewDatasourceListResponseAssetSpecPvc(claimName NullableString, path NullableString, ) *DatasourceListResponseAssetSpecPvc`

NewDatasourceListResponseAssetSpecPvc instantiates a new DatasourceListResponseAssetSpecPvc object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatasourceListResponseAssetSpecPvcWithDefaults

`func NewDatasourceListResponseAssetSpecPvcWithDefaults() *DatasourceListResponseAssetSpecPvc`

NewDatasourceListResponseAssetSpecPvcWithDefaults instantiates a new DatasourceListResponseAssetSpecPvc object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExistingPvc

`func (o *DatasourceListResponseAssetSpecPvc) GetExistingPvc() bool`

GetExistingPvc returns the ExistingPvc field if non-nil, zero value otherwise.

### GetExistingPvcOk

`func (o *DatasourceListResponseAssetSpecPvc) GetExistingPvcOk() (*bool, bool)`

GetExistingPvcOk returns a tuple with the ExistingPvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingPvc

`func (o *DatasourceListResponseAssetSpecPvc) SetExistingPvc(v bool)`

SetExistingPvc sets ExistingPvc field to given value.

### HasExistingPvc

`func (o *DatasourceListResponseAssetSpecPvc) HasExistingPvc() bool`

HasExistingPvc returns a boolean if a field has been set.

### SetExistingPvcNil

`func (o *DatasourceListResponseAssetSpecPvc) SetExistingPvcNil(b bool)`

 SetExistingPvcNil sets the value for ExistingPvc to be an explicit nil

### UnsetExistingPvc
`func (o *DatasourceListResponseAssetSpecPvc) UnsetExistingPvc()`

UnsetExistingPvc ensures that no value is present for ExistingPvc, not even an explicit nil
### GetClaimName

`func (o *DatasourceListResponseAssetSpecPvc) GetClaimName() string`

GetClaimName returns the ClaimName field if non-nil, zero value otherwise.

### GetClaimNameOk

`func (o *DatasourceListResponseAssetSpecPvc) GetClaimNameOk() (*string, bool)`

GetClaimNameOk returns a tuple with the ClaimName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimName

`func (o *DatasourceListResponseAssetSpecPvc) SetClaimName(v string)`

SetClaimName sets ClaimName field to given value.


### SetClaimNameNil

`func (o *DatasourceListResponseAssetSpecPvc) SetClaimNameNil(b bool)`

 SetClaimNameNil sets the value for ClaimName to be an explicit nil

### UnsetClaimName
`func (o *DatasourceListResponseAssetSpecPvc) UnsetClaimName()`

UnsetClaimName ensures that no value is present for ClaimName, not even an explicit nil
### GetPath

`func (o *DatasourceListResponseAssetSpecPvc) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *DatasourceListResponseAssetSpecPvc) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *DatasourceListResponseAssetSpecPvc) SetPath(v string)`

SetPath sets Path field to given value.


### SetPathNil

`func (o *DatasourceListResponseAssetSpecPvc) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *DatasourceListResponseAssetSpecPvc) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetReadOnly

`func (o *DatasourceListResponseAssetSpecPvc) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *DatasourceListResponseAssetSpecPvc) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *DatasourceListResponseAssetSpecPvc) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *DatasourceListResponseAssetSpecPvc) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *DatasourceListResponseAssetSpecPvc) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *DatasourceListResponseAssetSpecPvc) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetEphemeral

`func (o *DatasourceListResponseAssetSpecPvc) GetEphemeral() bool`

GetEphemeral returns the Ephemeral field if non-nil, zero value otherwise.

### GetEphemeralOk

`func (o *DatasourceListResponseAssetSpecPvc) GetEphemeralOk() (*bool, bool)`

GetEphemeralOk returns a tuple with the Ephemeral field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEphemeral

`func (o *DatasourceListResponseAssetSpecPvc) SetEphemeral(v bool)`

SetEphemeral sets Ephemeral field to given value.

### HasEphemeral

`func (o *DatasourceListResponseAssetSpecPvc) HasEphemeral() bool`

HasEphemeral returns a boolean if a field has been set.

### SetEphemeralNil

`func (o *DatasourceListResponseAssetSpecPvc) SetEphemeralNil(b bool)`

 SetEphemeralNil sets the value for Ephemeral to be an explicit nil

### UnsetEphemeral
`func (o *DatasourceListResponseAssetSpecPvc) UnsetEphemeral()`

UnsetEphemeral ensures that no value is present for Ephemeral, not even an explicit nil
### GetClaimInfo

`func (o *DatasourceListResponseAssetSpecPvc) GetClaimInfo() ClaimInfo`

GetClaimInfo returns the ClaimInfo field if non-nil, zero value otherwise.

### GetClaimInfoOk

`func (o *DatasourceListResponseAssetSpecPvc) GetClaimInfoOk() (*ClaimInfo, bool)`

GetClaimInfoOk returns a tuple with the ClaimInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimInfo

`func (o *DatasourceListResponseAssetSpecPvc) SetClaimInfo(v ClaimInfo)`

SetClaimInfo sets ClaimInfo field to given value.

### HasClaimInfo

`func (o *DatasourceListResponseAssetSpecPvc) HasClaimInfo() bool`

HasClaimInfo returns a boolean if a field has been set.

### SetClaimInfoNil

`func (o *DatasourceListResponseAssetSpecPvc) SetClaimInfoNil(b bool)`

 SetClaimInfoNil sets the value for ClaimInfo to be an explicit nil

### UnsetClaimInfo
`func (o *DatasourceListResponseAssetSpecPvc) UnsetClaimInfo()`

UnsetClaimInfo ensures that no value is present for ClaimInfo, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


