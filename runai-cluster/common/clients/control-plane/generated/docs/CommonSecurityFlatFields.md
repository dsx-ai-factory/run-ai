# CommonSecurityFlatFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**UidGidSource** | Pointer to [**NullableUidGidSource**](UidGidSource.md) |  | [optional] 
**Capabilities** | Pointer to [**[]Capability**](Capability.md) | Add POSIX capabilities to running containers. Defaults to the default set of capabilities granted by the container runtime. | [optional] 
**SeccompProfileType** | Pointer to [**NullableSeccompProfileType**](SeccompProfileType.md) |  | [optional] 
**RunAsNonRoot** | Pointer to **NullableBool** | Force the container to run as a non-root user. | [optional] 
**ReadOnlyRootFilesystem** | Pointer to **NullableBool** | If true, mounts the container&#39;s root filesystem as read-only. | [optional] 
**RunAsUid** | Pointer to **NullableInt64** | The user id to run the entrypoint of the container which executes the workspace. Default to the value specified in the environment asset &#x60;runAsUid&#x60; field (optional). Use only when the source uid/gid of the environment asset is not &#x60;fromTheImage&#x60;, and &#x60;overrideUidGidInWorkspace&#x60; is enabled. | [optional] 
**RunAsGid** | Pointer to **NullableInt64** | The group id to run the entrypoint of the container which executes the workspace. Default to the value specified in the environment asset &#x60;runAsGid&#x60; field (optional). Use only when the source uid/gid of the environment asset is not &#x60;fromTheImage&#x60;, and &#x60;overrideUidGidInWorkspace&#x60; is enabled. | [optional] 
**SupplementalGroups** | Pointer to **NullableString** | Comma separated list of groups that the user running the container belongs to, in addition to the group indicated by runAsGid. Use only when the source uid/gid of the environment asset is not &#x60;fromTheImage&#x60;, and &#x60;overrideUidGidInWorkspace&#x60; is enabled. Using an empty string implies reverting the supplementary groups of the image. | [optional] 

## Methods

### NewCommonSecurityFlatFields

`func NewCommonSecurityFlatFields() *CommonSecurityFlatFields`

NewCommonSecurityFlatFields instantiates a new CommonSecurityFlatFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCommonSecurityFlatFieldsWithDefaults

`func NewCommonSecurityFlatFieldsWithDefaults() *CommonSecurityFlatFields`

NewCommonSecurityFlatFieldsWithDefaults instantiates a new CommonSecurityFlatFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUidGidSource

`func (o *CommonSecurityFlatFields) GetUidGidSource() UidGidSource`

GetUidGidSource returns the UidGidSource field if non-nil, zero value otherwise.

### GetUidGidSourceOk

`func (o *CommonSecurityFlatFields) GetUidGidSourceOk() (*UidGidSource, bool)`

GetUidGidSourceOk returns a tuple with the UidGidSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUidGidSource

`func (o *CommonSecurityFlatFields) SetUidGidSource(v UidGidSource)`

SetUidGidSource sets UidGidSource field to given value.

### HasUidGidSource

`func (o *CommonSecurityFlatFields) HasUidGidSource() bool`

HasUidGidSource returns a boolean if a field has been set.

### SetUidGidSourceNil

`func (o *CommonSecurityFlatFields) SetUidGidSourceNil(b bool)`

 SetUidGidSourceNil sets the value for UidGidSource to be an explicit nil

### UnsetUidGidSource
`func (o *CommonSecurityFlatFields) UnsetUidGidSource()`

UnsetUidGidSource ensures that no value is present for UidGidSource, not even an explicit nil
### GetCapabilities

`func (o *CommonSecurityFlatFields) GetCapabilities() []Capability`

GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.

### GetCapabilitiesOk

`func (o *CommonSecurityFlatFields) GetCapabilitiesOk() (*[]Capability, bool)`

GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapabilities

`func (o *CommonSecurityFlatFields) SetCapabilities(v []Capability)`

SetCapabilities sets Capabilities field to given value.

### HasCapabilities

`func (o *CommonSecurityFlatFields) HasCapabilities() bool`

HasCapabilities returns a boolean if a field has been set.

### SetCapabilitiesNil

`func (o *CommonSecurityFlatFields) SetCapabilitiesNil(b bool)`

 SetCapabilitiesNil sets the value for Capabilities to be an explicit nil

### UnsetCapabilities
`func (o *CommonSecurityFlatFields) UnsetCapabilities()`

UnsetCapabilities ensures that no value is present for Capabilities, not even an explicit nil
### GetSeccompProfileType

`func (o *CommonSecurityFlatFields) GetSeccompProfileType() SeccompProfileType`

GetSeccompProfileType returns the SeccompProfileType field if non-nil, zero value otherwise.

### GetSeccompProfileTypeOk

`func (o *CommonSecurityFlatFields) GetSeccompProfileTypeOk() (*SeccompProfileType, bool)`

GetSeccompProfileTypeOk returns a tuple with the SeccompProfileType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeccompProfileType

`func (o *CommonSecurityFlatFields) SetSeccompProfileType(v SeccompProfileType)`

SetSeccompProfileType sets SeccompProfileType field to given value.

### HasSeccompProfileType

`func (o *CommonSecurityFlatFields) HasSeccompProfileType() bool`

HasSeccompProfileType returns a boolean if a field has been set.

### SetSeccompProfileTypeNil

`func (o *CommonSecurityFlatFields) SetSeccompProfileTypeNil(b bool)`

 SetSeccompProfileTypeNil sets the value for SeccompProfileType to be an explicit nil

### UnsetSeccompProfileType
`func (o *CommonSecurityFlatFields) UnsetSeccompProfileType()`

UnsetSeccompProfileType ensures that no value is present for SeccompProfileType, not even an explicit nil
### GetRunAsNonRoot

`func (o *CommonSecurityFlatFields) GetRunAsNonRoot() bool`

GetRunAsNonRoot returns the RunAsNonRoot field if non-nil, zero value otherwise.

### GetRunAsNonRootOk

`func (o *CommonSecurityFlatFields) GetRunAsNonRootOk() (*bool, bool)`

GetRunAsNonRootOk returns a tuple with the RunAsNonRoot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsNonRoot

`func (o *CommonSecurityFlatFields) SetRunAsNonRoot(v bool)`

SetRunAsNonRoot sets RunAsNonRoot field to given value.

### HasRunAsNonRoot

`func (o *CommonSecurityFlatFields) HasRunAsNonRoot() bool`

HasRunAsNonRoot returns a boolean if a field has been set.

### SetRunAsNonRootNil

`func (o *CommonSecurityFlatFields) SetRunAsNonRootNil(b bool)`

 SetRunAsNonRootNil sets the value for RunAsNonRoot to be an explicit nil

### UnsetRunAsNonRoot
`func (o *CommonSecurityFlatFields) UnsetRunAsNonRoot()`

UnsetRunAsNonRoot ensures that no value is present for RunAsNonRoot, not even an explicit nil
### GetReadOnlyRootFilesystem

`func (o *CommonSecurityFlatFields) GetReadOnlyRootFilesystem() bool`

GetReadOnlyRootFilesystem returns the ReadOnlyRootFilesystem field if non-nil, zero value otherwise.

### GetReadOnlyRootFilesystemOk

`func (o *CommonSecurityFlatFields) GetReadOnlyRootFilesystemOk() (*bool, bool)`

GetReadOnlyRootFilesystemOk returns a tuple with the ReadOnlyRootFilesystem field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnlyRootFilesystem

`func (o *CommonSecurityFlatFields) SetReadOnlyRootFilesystem(v bool)`

SetReadOnlyRootFilesystem sets ReadOnlyRootFilesystem field to given value.

### HasReadOnlyRootFilesystem

`func (o *CommonSecurityFlatFields) HasReadOnlyRootFilesystem() bool`

HasReadOnlyRootFilesystem returns a boolean if a field has been set.

### SetReadOnlyRootFilesystemNil

`func (o *CommonSecurityFlatFields) SetReadOnlyRootFilesystemNil(b bool)`

 SetReadOnlyRootFilesystemNil sets the value for ReadOnlyRootFilesystem to be an explicit nil

### UnsetReadOnlyRootFilesystem
`func (o *CommonSecurityFlatFields) UnsetReadOnlyRootFilesystem()`

UnsetReadOnlyRootFilesystem ensures that no value is present for ReadOnlyRootFilesystem, not even an explicit nil
### GetRunAsUid

`func (o *CommonSecurityFlatFields) GetRunAsUid() int64`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *CommonSecurityFlatFields) GetRunAsUidOk() (*int64, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *CommonSecurityFlatFields) SetRunAsUid(v int64)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *CommonSecurityFlatFields) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *CommonSecurityFlatFields) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *CommonSecurityFlatFields) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetRunAsGid

`func (o *CommonSecurityFlatFields) GetRunAsGid() int64`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *CommonSecurityFlatFields) GetRunAsGidOk() (*int64, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *CommonSecurityFlatFields) SetRunAsGid(v int64)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *CommonSecurityFlatFields) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *CommonSecurityFlatFields) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *CommonSecurityFlatFields) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetSupplementalGroups

`func (o *CommonSecurityFlatFields) GetSupplementalGroups() string`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *CommonSecurityFlatFields) GetSupplementalGroupsOk() (*string, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *CommonSecurityFlatFields) SetSupplementalGroups(v string)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *CommonSecurityFlatFields) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *CommonSecurityFlatFields) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *CommonSecurityFlatFields) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


