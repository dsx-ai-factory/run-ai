# PvcInstance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | unique name to identify the instance. primarily used for policy locked rules. | [optional] 
**Path** | Pointer to **NullableString** | Local path within the workload to which the PVC bucket will be mapped. (mandatory) | [optional] 
**ClaimInfo** | Pointer to [**NullableClaimInfo**](ClaimInfo.md) |  | [optional] 
**ExistingPvc** | Pointer to **NullableBool** | Verify existing PVC. PVC is assumed to exist when set to &#x60;true&#x60;. If set to &#x60;false&#x60;, the PVC will be created, if it does not exist. Cannot be set to &#x60;true&#x60; when &#x60;dataSharing&#x60; is &#x60;true&#x60;. | [optional] [default to false]
**ClaimName** | Pointer to **NullableString** | Name for the PVC. Allow referencing it across workloads. If not provided, a name based on the workload name and scope will be auto-generated. | [optional] 
**ReadOnly** | Pointer to **NullableBool** | Permit only read access to PVC. | [optional] [default to false]
**Ephemeral** | Pointer to **NullableBool** | Use &#x60;true&#x60; to set PVC to ephemeral. If set to &#x60;true&#x60;, the PVC will be deleted when the workload is stopped. Not supported for inference workloads. | [optional] [default to false]
**DataSharing** | Pointer to **NullableBool** | Use &#x60;true&#x60; to share the PVC data to all projects under the selected scope. Cannot be set to &#x60;true&#x60; when &#x60;existingPvc&#x60; is &#x60;true&#x60;. | [optional] [default to false]
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the item is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] [default to false]

## Methods

### NewPvcInstance

`func NewPvcInstance() *PvcInstance`

NewPvcInstance instantiates a new PvcInstance object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPvcInstanceWithDefaults

`func NewPvcInstanceWithDefaults() *PvcInstance`

NewPvcInstanceWithDefaults instantiates a new PvcInstance object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *PvcInstance) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PvcInstance) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PvcInstance) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *PvcInstance) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *PvcInstance) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *PvcInstance) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetPath

`func (o *PvcInstance) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *PvcInstance) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *PvcInstance) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *PvcInstance) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *PvcInstance) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *PvcInstance) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetClaimInfo

`func (o *PvcInstance) GetClaimInfo() ClaimInfo`

GetClaimInfo returns the ClaimInfo field if non-nil, zero value otherwise.

### GetClaimInfoOk

`func (o *PvcInstance) GetClaimInfoOk() (*ClaimInfo, bool)`

GetClaimInfoOk returns a tuple with the ClaimInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimInfo

`func (o *PvcInstance) SetClaimInfo(v ClaimInfo)`

SetClaimInfo sets ClaimInfo field to given value.

### HasClaimInfo

`func (o *PvcInstance) HasClaimInfo() bool`

HasClaimInfo returns a boolean if a field has been set.

### SetClaimInfoNil

`func (o *PvcInstance) SetClaimInfoNil(b bool)`

 SetClaimInfoNil sets the value for ClaimInfo to be an explicit nil

### UnsetClaimInfo
`func (o *PvcInstance) UnsetClaimInfo()`

UnsetClaimInfo ensures that no value is present for ClaimInfo, not even an explicit nil
### GetExistingPvc

`func (o *PvcInstance) GetExistingPvc() bool`

GetExistingPvc returns the ExistingPvc field if non-nil, zero value otherwise.

### GetExistingPvcOk

`func (o *PvcInstance) GetExistingPvcOk() (*bool, bool)`

GetExistingPvcOk returns a tuple with the ExistingPvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingPvc

`func (o *PvcInstance) SetExistingPvc(v bool)`

SetExistingPvc sets ExistingPvc field to given value.

### HasExistingPvc

`func (o *PvcInstance) HasExistingPvc() bool`

HasExistingPvc returns a boolean if a field has been set.

### SetExistingPvcNil

`func (o *PvcInstance) SetExistingPvcNil(b bool)`

 SetExistingPvcNil sets the value for ExistingPvc to be an explicit nil

### UnsetExistingPvc
`func (o *PvcInstance) UnsetExistingPvc()`

UnsetExistingPvc ensures that no value is present for ExistingPvc, not even an explicit nil
### GetClaimName

`func (o *PvcInstance) GetClaimName() string`

GetClaimName returns the ClaimName field if non-nil, zero value otherwise.

### GetClaimNameOk

`func (o *PvcInstance) GetClaimNameOk() (*string, bool)`

GetClaimNameOk returns a tuple with the ClaimName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimName

`func (o *PvcInstance) SetClaimName(v string)`

SetClaimName sets ClaimName field to given value.

### HasClaimName

`func (o *PvcInstance) HasClaimName() bool`

HasClaimName returns a boolean if a field has been set.

### SetClaimNameNil

`func (o *PvcInstance) SetClaimNameNil(b bool)`

 SetClaimNameNil sets the value for ClaimName to be an explicit nil

### UnsetClaimName
`func (o *PvcInstance) UnsetClaimName()`

UnsetClaimName ensures that no value is present for ClaimName, not even an explicit nil
### GetReadOnly

`func (o *PvcInstance) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *PvcInstance) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *PvcInstance) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *PvcInstance) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *PvcInstance) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *PvcInstance) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil
### GetEphemeral

`func (o *PvcInstance) GetEphemeral() bool`

GetEphemeral returns the Ephemeral field if non-nil, zero value otherwise.

### GetEphemeralOk

`func (o *PvcInstance) GetEphemeralOk() (*bool, bool)`

GetEphemeralOk returns a tuple with the Ephemeral field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEphemeral

`func (o *PvcInstance) SetEphemeral(v bool)`

SetEphemeral sets Ephemeral field to given value.

### HasEphemeral

`func (o *PvcInstance) HasEphemeral() bool`

HasEphemeral returns a boolean if a field has been set.

### SetEphemeralNil

`func (o *PvcInstance) SetEphemeralNil(b bool)`

 SetEphemeralNil sets the value for Ephemeral to be an explicit nil

### UnsetEphemeral
`func (o *PvcInstance) UnsetEphemeral()`

UnsetEphemeral ensures that no value is present for Ephemeral, not even an explicit nil
### GetDataSharing

`func (o *PvcInstance) GetDataSharing() bool`

GetDataSharing returns the DataSharing field if non-nil, zero value otherwise.

### GetDataSharingOk

`func (o *PvcInstance) GetDataSharingOk() (*bool, bool)`

GetDataSharingOk returns a tuple with the DataSharing field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataSharing

`func (o *PvcInstance) SetDataSharing(v bool)`

SetDataSharing sets DataSharing field to given value.

### HasDataSharing

`func (o *PvcInstance) HasDataSharing() bool`

HasDataSharing returns a boolean if a field has been set.

### SetDataSharingNil

`func (o *PvcInstance) SetDataSharingNil(b bool)`

 SetDataSharingNil sets the value for DataSharing to be an explicit nil

### UnsetDataSharing
`func (o *PvcInstance) UnsetDataSharing()`

UnsetDataSharing ensures that no value is present for DataSharing, not even an explicit nil
### GetExclude

`func (o *PvcInstance) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *PvcInstance) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *PvcInstance) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *PvcInstance) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *PvcInstance) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *PvcInstance) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


