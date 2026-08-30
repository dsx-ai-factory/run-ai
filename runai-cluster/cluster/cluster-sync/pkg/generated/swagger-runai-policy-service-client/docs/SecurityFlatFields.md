# SecurityFlatFields

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
**AllowPrivilegeEscalation** | Pointer to **NullableBool** | Allow the container running the workload and all launched processes to gain additional privileges after the workload starts. For more information consult the User Identity in Container guide at https://docs.run.ai/admin/runai-setup/config/non-root-containers/ | [optional] 
**HostIpc** | Pointer to **NullableBool** | Whether to enable host IPC. Defaults to false. | [optional] 
**HostNetwork** | Pointer to **NullableBool** | Whether to enable host networking. Default to false. | [optional] 

## Methods

### NewSecurityFlatFields

`func NewSecurityFlatFields() *SecurityFlatFields`

NewSecurityFlatFields instantiates a new SecurityFlatFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecurityFlatFieldsWithDefaults

`func NewSecurityFlatFieldsWithDefaults() *SecurityFlatFields`

NewSecurityFlatFieldsWithDefaults instantiates a new SecurityFlatFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUidGidSource

`func (o *SecurityFlatFields) GetUidGidSource() UidGidSource`

GetUidGidSource returns the UidGidSource field if non-nil, zero value otherwise.

### GetUidGidSourceOk

`func (o *SecurityFlatFields) GetUidGidSourceOk() (*UidGidSource, bool)`

GetUidGidSourceOk returns a tuple with the UidGidSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUidGidSource

`func (o *SecurityFlatFields) SetUidGidSource(v UidGidSource)`

SetUidGidSource sets UidGidSource field to given value.

### HasUidGidSource

`func (o *SecurityFlatFields) HasUidGidSource() bool`

HasUidGidSource returns a boolean if a field has been set.

### SetUidGidSourceNil

`func (o *SecurityFlatFields) SetUidGidSourceNil(b bool)`

 SetUidGidSourceNil sets the value for UidGidSource to be an explicit nil

### UnsetUidGidSource
`func (o *SecurityFlatFields) UnsetUidGidSource()`

UnsetUidGidSource ensures that no value is present for UidGidSource, not even an explicit nil
### GetCapabilities

`func (o *SecurityFlatFields) GetCapabilities() []Capability`

GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.

### GetCapabilitiesOk

`func (o *SecurityFlatFields) GetCapabilitiesOk() (*[]Capability, bool)`

GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapabilities

`func (o *SecurityFlatFields) SetCapabilities(v []Capability)`

SetCapabilities sets Capabilities field to given value.

### HasCapabilities

`func (o *SecurityFlatFields) HasCapabilities() bool`

HasCapabilities returns a boolean if a field has been set.

### SetCapabilitiesNil

`func (o *SecurityFlatFields) SetCapabilitiesNil(b bool)`

 SetCapabilitiesNil sets the value for Capabilities to be an explicit nil

### UnsetCapabilities
`func (o *SecurityFlatFields) UnsetCapabilities()`

UnsetCapabilities ensures that no value is present for Capabilities, not even an explicit nil
### GetSeccompProfileType

`func (o *SecurityFlatFields) GetSeccompProfileType() SeccompProfileType`

GetSeccompProfileType returns the SeccompProfileType field if non-nil, zero value otherwise.

### GetSeccompProfileTypeOk

`func (o *SecurityFlatFields) GetSeccompProfileTypeOk() (*SeccompProfileType, bool)`

GetSeccompProfileTypeOk returns a tuple with the SeccompProfileType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeccompProfileType

`func (o *SecurityFlatFields) SetSeccompProfileType(v SeccompProfileType)`

SetSeccompProfileType sets SeccompProfileType field to given value.

### HasSeccompProfileType

`func (o *SecurityFlatFields) HasSeccompProfileType() bool`

HasSeccompProfileType returns a boolean if a field has been set.

### SetSeccompProfileTypeNil

`func (o *SecurityFlatFields) SetSeccompProfileTypeNil(b bool)`

 SetSeccompProfileTypeNil sets the value for SeccompProfileType to be an explicit nil

### UnsetSeccompProfileType
`func (o *SecurityFlatFields) UnsetSeccompProfileType()`

UnsetSeccompProfileType ensures that no value is present for SeccompProfileType, not even an explicit nil
### GetRunAsNonRoot

`func (o *SecurityFlatFields) GetRunAsNonRoot() bool`

GetRunAsNonRoot returns the RunAsNonRoot field if non-nil, zero value otherwise.

### GetRunAsNonRootOk

`func (o *SecurityFlatFields) GetRunAsNonRootOk() (*bool, bool)`

GetRunAsNonRootOk returns a tuple with the RunAsNonRoot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsNonRoot

`func (o *SecurityFlatFields) SetRunAsNonRoot(v bool)`

SetRunAsNonRoot sets RunAsNonRoot field to given value.

### HasRunAsNonRoot

`func (o *SecurityFlatFields) HasRunAsNonRoot() bool`

HasRunAsNonRoot returns a boolean if a field has been set.

### SetRunAsNonRootNil

`func (o *SecurityFlatFields) SetRunAsNonRootNil(b bool)`

 SetRunAsNonRootNil sets the value for RunAsNonRoot to be an explicit nil

### UnsetRunAsNonRoot
`func (o *SecurityFlatFields) UnsetRunAsNonRoot()`

UnsetRunAsNonRoot ensures that no value is present for RunAsNonRoot, not even an explicit nil
### GetReadOnlyRootFilesystem

`func (o *SecurityFlatFields) GetReadOnlyRootFilesystem() bool`

GetReadOnlyRootFilesystem returns the ReadOnlyRootFilesystem field if non-nil, zero value otherwise.

### GetReadOnlyRootFilesystemOk

`func (o *SecurityFlatFields) GetReadOnlyRootFilesystemOk() (*bool, bool)`

GetReadOnlyRootFilesystemOk returns a tuple with the ReadOnlyRootFilesystem field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnlyRootFilesystem

`func (o *SecurityFlatFields) SetReadOnlyRootFilesystem(v bool)`

SetReadOnlyRootFilesystem sets ReadOnlyRootFilesystem field to given value.

### HasReadOnlyRootFilesystem

`func (o *SecurityFlatFields) HasReadOnlyRootFilesystem() bool`

HasReadOnlyRootFilesystem returns a boolean if a field has been set.

### SetReadOnlyRootFilesystemNil

`func (o *SecurityFlatFields) SetReadOnlyRootFilesystemNil(b bool)`

 SetReadOnlyRootFilesystemNil sets the value for ReadOnlyRootFilesystem to be an explicit nil

### UnsetReadOnlyRootFilesystem
`func (o *SecurityFlatFields) UnsetReadOnlyRootFilesystem()`

UnsetReadOnlyRootFilesystem ensures that no value is present for ReadOnlyRootFilesystem, not even an explicit nil
### GetRunAsUid

`func (o *SecurityFlatFields) GetRunAsUid() int64`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *SecurityFlatFields) GetRunAsUidOk() (*int64, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *SecurityFlatFields) SetRunAsUid(v int64)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *SecurityFlatFields) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *SecurityFlatFields) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *SecurityFlatFields) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetRunAsGid

`func (o *SecurityFlatFields) GetRunAsGid() int64`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *SecurityFlatFields) GetRunAsGidOk() (*int64, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *SecurityFlatFields) SetRunAsGid(v int64)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *SecurityFlatFields) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *SecurityFlatFields) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *SecurityFlatFields) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetSupplementalGroups

`func (o *SecurityFlatFields) GetSupplementalGroups() string`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *SecurityFlatFields) GetSupplementalGroupsOk() (*string, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *SecurityFlatFields) SetSupplementalGroups(v string)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *SecurityFlatFields) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *SecurityFlatFields) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *SecurityFlatFields) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil
### GetAllowPrivilegeEscalation

`func (o *SecurityFlatFields) GetAllowPrivilegeEscalation() bool`

GetAllowPrivilegeEscalation returns the AllowPrivilegeEscalation field if non-nil, zero value otherwise.

### GetAllowPrivilegeEscalationOk

`func (o *SecurityFlatFields) GetAllowPrivilegeEscalationOk() (*bool, bool)`

GetAllowPrivilegeEscalationOk returns a tuple with the AllowPrivilegeEscalation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowPrivilegeEscalation

`func (o *SecurityFlatFields) SetAllowPrivilegeEscalation(v bool)`

SetAllowPrivilegeEscalation sets AllowPrivilegeEscalation field to given value.

### HasAllowPrivilegeEscalation

`func (o *SecurityFlatFields) HasAllowPrivilegeEscalation() bool`

HasAllowPrivilegeEscalation returns a boolean if a field has been set.

### SetAllowPrivilegeEscalationNil

`func (o *SecurityFlatFields) SetAllowPrivilegeEscalationNil(b bool)`

 SetAllowPrivilegeEscalationNil sets the value for AllowPrivilegeEscalation to be an explicit nil

### UnsetAllowPrivilegeEscalation
`func (o *SecurityFlatFields) UnsetAllowPrivilegeEscalation()`

UnsetAllowPrivilegeEscalation ensures that no value is present for AllowPrivilegeEscalation, not even an explicit nil
### GetHostIpc

`func (o *SecurityFlatFields) GetHostIpc() bool`

GetHostIpc returns the HostIpc field if non-nil, zero value otherwise.

### GetHostIpcOk

`func (o *SecurityFlatFields) GetHostIpcOk() (*bool, bool)`

GetHostIpcOk returns a tuple with the HostIpc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostIpc

`func (o *SecurityFlatFields) SetHostIpc(v bool)`

SetHostIpc sets HostIpc field to given value.

### HasHostIpc

`func (o *SecurityFlatFields) HasHostIpc() bool`

HasHostIpc returns a boolean if a field has been set.

### SetHostIpcNil

`func (o *SecurityFlatFields) SetHostIpcNil(b bool)`

 SetHostIpcNil sets the value for HostIpc to be an explicit nil

### UnsetHostIpc
`func (o *SecurityFlatFields) UnsetHostIpc()`

UnsetHostIpc ensures that no value is present for HostIpc, not even an explicit nil
### GetHostNetwork

`func (o *SecurityFlatFields) GetHostNetwork() bool`

GetHostNetwork returns the HostNetwork field if non-nil, zero value otherwise.

### GetHostNetworkOk

`func (o *SecurityFlatFields) GetHostNetworkOk() (*bool, bool)`

GetHostNetworkOk returns a tuple with the HostNetwork field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostNetwork

`func (o *SecurityFlatFields) SetHostNetwork(v bool)`

SetHostNetwork sets HostNetwork field to given value.

### HasHostNetwork

`func (o *SecurityFlatFields) HasHostNetwork() bool`

HasHostNetwork returns a boolean if a field has been set.

### SetHostNetworkNil

`func (o *SecurityFlatFields) SetHostNetworkNil(b bool)`

 SetHostNetworkNil sets the value for HostNetwork to be an explicit nil

### UnsetHostNetwork
`func (o *SecurityFlatFields) UnsetHostNetwork()`

UnsetHostNetwork ensures that no value is present for HostNetwork, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


