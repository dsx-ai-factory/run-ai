# PVCAssetSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | Pointer to **NullableString** | Local path within the workload to which the PVC bucket will be mapped. (mandatory) | [optional] 
**ClaimInfo** | Pointer to [**NullableClaimInfo**](ClaimInfo.md) |  | [optional] 
**ExistingPvc** | Pointer to **NullableBool** | Verify existing PVC. PVC is assumed to exist when set to &#x60;true&#x60;. If set to &#x60;false&#x60;, the PVC will be created, if it does not exist. Cannot be set to &#x60;true&#x60; when &#x60;dataSharing&#x60; is &#x60;true&#x60;. | [optional] [default to false]
**ClaimName** | Pointer to **NullableString** | Name for the PVC. Allow referencing it across workloads. If not provided, a name based on the workload name and scope will be auto-generated. | [optional] 
**ReadOnly** | Pointer to **NullableBool** | Permit only read access to PVC. | [optional] [default to false]
**Ephemeral** | Pointer to **NullableBool** | Use &#x60;true&#x60; to set PVC to ephemeral. If set to &#x60;true&#x60;, the PVC will be deleted when the workload is stopped. Not supported for inference workloads. | [optional] [default to false]
**DataSharing** | Pointer to **NullableBool** | Use &#x60;true&#x60; to share the PVC data to all projects under the selected scope. Cannot be set to &#x60;true&#x60; when &#x60;existingPvc&#x60; is &#x60;true&#x60;. | [optional] [default to false]

## Methods

### NewPVCAssetSpec

`func NewPVCAssetSpec() *PVCAssetSpec`

NewPVCAssetSpec instantiates a new PVCAssetSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCAssetSpecWithDefaults

`func NewPVCAssetSpecWithDefaults() *PVCAssetSpec`

NewPVCAssetSpecWithDefaults instantiates a new PVCAssetSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *PVCAssetSpec) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *PVCAssetSpec) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *PVCAssetSpec) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *PVCAssetSpec) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *PVCAssetSpec) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *PVCAssetSpec) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetClaimInfo

`func (o *PVCAssetSpec) GetClaimInfo() ClaimInfo`

GetClaimInfo returns the ClaimInfo field if non-nil, zero value otherwise.

### GetClaimInfoOk

`func (o *PVCAssetSpec) GetClaimInfoOk() (*ClaimInfo, bool)`

GetClaimInfoOk returns a tuple with the ClaimInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimInfo

`func (o *PVCAssetSpec) SetClaimInfo(v ClaimInfo)`

SetClaimInfo sets ClaimInfo field to given value.

### HasClaimInfo

`func (o *PVCAssetSpec) HasClaimInfo() bool`

HasClaimInfo returns a boolean if a field has been set.

### SetClaimInfoNil

`func (o *PVCAssetSpec) SetClaimInfoNil(b bool)`

 SetClaimInfoNil sets the value for ClaimInfo to be an explicit nil

### UnsetClaimInfo
`func (o *PVCAssetSpec) UnsetClaimInfo()`

UnsetClaimInfo ensures that no value is present for ClaimInfo, not even an explicit nil
### GetExistingPvc

`func (o *PVCAssetSpec) GetExistingPvc() bool`

GetExistingPvc returns the ExistingPvc field if non-nil, zero value otherwise.

### GetExistingPvcOk

`func (o *PVCAssetSpec) GetExistingPvcOk() (*bool, bool)`

GetExistingPvcOk returns a tuple with the ExistingPvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingPvc

`func (o *PVCAssetSpec) SetExistingPvc(v bool)`

SetExistingPvc sets ExistingPvc field to given value.

### HasExistingPvc

`func (o *PVCAssetSpec) HasExistingPvc() bool`

HasExistingPvc returns a boolean if a field has been set.

### SetExistingPvcNil

`func (o *PVCAssetSpec) SetExistingPvcNil(b bool)`

 SetExistingPvcNil sets the value for ExistingPvc to be an explicit nil

### UnsetExistingPvc
`func (o *PVCAssetSpec) UnsetExistingPvc()`

UnsetExistingPvc ensures that no value is present for ExistingPvc, not even an explicit nil
### GetClaimName

`func (o *PVCAssetSpec) GetClaimName() string`

GetClaimName returns the ClaimName field if non-nil, zero value otherwise.

### GetClaimNameOk

`func (o *PVCAssetSpec) GetClaimNameOk() (*string, bool)`

GetClaimNameOk returns a tuple with the ClaimName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimName

`func (o *PVCAssetSpec) SetClaimName(v string)`

SetClaimName sets ClaimName field to given value.

### HasClaimName

`func (o *PVCAssetSpec) HasClaimName() bool`

HasClaimName returns a boolean if a field has been set.

### SetClaimNameNil

`func (o *PVCAssetSpec) SetClaimNameNil(b bool)`

 SetClaimNameNil sets the value for ClaimName to be an explicit nil

### UnsetClaimName
`func (o *PVCAssetSpec) UnsetClaimName()`

UnsetClaimName ensures that no value is present for ClaimName, not even an explicit nil
### GetReadOnly

`func (o *PVCAssetSpec) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *PVCAssetSpec) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *PVCAssetSpec) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *PVCAssetSpec) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *PVCAssetSpec) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *PVCAssetSpec) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetEphemeral

`func (o *PVCAssetSpec) GetEphemeral() bool`

GetEphemeral returns the Ephemeral field if non-nil, zero value otherwise.

### GetEphemeralOk

`func (o *PVCAssetSpec) GetEphemeralOk() (*bool, bool)`

GetEphemeralOk returns a tuple with the Ephemeral field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEphemeral

`func (o *PVCAssetSpec) SetEphemeral(v bool)`

SetEphemeral sets Ephemeral field to given value.

### HasEphemeral

`func (o *PVCAssetSpec) HasEphemeral() bool`

HasEphemeral returns a boolean if a field has been set.

### SetEphemeralNil

`func (o *PVCAssetSpec) SetEphemeralNil(b bool)`

 SetEphemeralNil sets the value for Ephemeral to be an explicit nil

### UnsetEphemeral
`func (o *PVCAssetSpec) UnsetEphemeral()`

UnsetEphemeral ensures that no value is present for Ephemeral, not even an explicit nil
### GetDataSharing

`func (o *PVCAssetSpec) GetDataSharing() bool`

GetDataSharing returns the DataSharing field if non-nil, zero value otherwise.

### GetDataSharingOk

`func (o *PVCAssetSpec) GetDataSharingOk() (*bool, bool)`

GetDataSharingOk returns a tuple with the DataSharing field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataSharing

`func (o *PVCAssetSpec) SetDataSharing(v bool)`

SetDataSharing sets DataSharing field to given value.

### HasDataSharing

`func (o *PVCAssetSpec) HasDataSharing() bool`

HasDataSharing returns a boolean if a field has been set.

### SetDataSharingNil

`func (o *PVCAssetSpec) SetDataSharingNil(b bool)`

 SetDataSharingNil sets the value for DataSharing to be an explicit nil

### UnsetDataSharing
`func (o *PVCAssetSpec) UnsetDataSharing()`

UnsetDataSharing ensures that no value is present for DataSharing, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


