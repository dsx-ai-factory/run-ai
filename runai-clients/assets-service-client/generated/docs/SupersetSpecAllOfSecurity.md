# SupersetSpecAllOfSecurity

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowPrivilegeEscalation** | Pointer to **NullableBool** | Allow the container running the workload and all launched processes to gain additional privileges after the workload starts. For more information consult the User Identity in Container guide at https://run-ai-docs.nvidia.com/self-hosted/latest/infrastructure-setup/advanced-setup/container-access/user-identity-in-containers | [optional] 
**Capabilities** | Pointer to [**[]Capability**](Capability.md) | Add POSIX capabilities to running containers. Defaults to the default set of capabilities granted by the container runtime. | [optional] 
**HostIpc** | Pointer to **NullableBool** | Whether to enable host IPC. Defaults to false. | [optional] 
**HostNetwork** | Pointer to **NullableBool** | Whether to enable host networking. Default to false. | [optional] 
**Privileged** | Pointer to **NullableBool** | Allow the container running the workload to operate in privileged mode, granting it elevated access to the host system. When enabled, the container receives all Linux capabilities, unrestricted access to host devices, and is exempt from security constraints. This effectively removes many of the isolation boundaries between the container and the host. | [optional] 
**ReadOnlyRootFilesystem** | Pointer to **NullableBool** | If true, mounts the container&#39;s root filesystem as read-only. | [optional] 
**RunAsGid** | Pointer to **NullableInt64** | The group id to run the entrypoint of the container which executes the workspace. Default to the value specified in the environment asset &#x60;runAsGid&#x60; field (optional). Use only when the source uid/gid of the environment asset is not &#x60;fromTheImage&#x60;, and &#x60;overrideUidGidInWorkspace&#x60; is enabled. | [optional] 
**RunAsNonRoot** | Pointer to **NullableBool** | Force the container to run as a non-root user. | [optional] 
**RunAsUid** | Pointer to **NullableInt64** | The user id to run the entrypoint of the container which executes the workspace. Default to the value specified in the environment asset &#x60;runAsUid&#x60; field (optional). Use only when the source uid/gid of the environment asset is not &#x60;fromTheImage&#x60;, and &#x60;overrideUidGidInWorkspace&#x60; is enabled. | [optional] 
**SeccompProfileType** | Pointer to [**NullableSeccompProfileType**](SeccompProfileType.md) |  | [optional] 
**SupplementalGroups** | Pointer to **NullableString** | Comma separated list of groups that the user running the container belongs to, in addition to the group indicated by runAsGid. Use only when the source uid/gid of the environment asset is not &#x60;fromTheImage&#x60;, and &#x60;overrideUidGidInWorkspace&#x60; is enabled. Using an empty string implies reverting the supplementary groups of the image. | [optional] 
**UidGidSource** | Pointer to [**NullableUidGidSource**](UidGidSource.md) |  | [optional] 

## Methods

### NewSupersetSpecAllOfSecurity

`func NewSupersetSpecAllOfSecurity() *SupersetSpecAllOfSecurity`

NewSupersetSpecAllOfSecurity instantiates a new SupersetSpecAllOfSecurity object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSupersetSpecAllOfSecurityWithDefaults

`func NewSupersetSpecAllOfSecurityWithDefaults() *SupersetSpecAllOfSecurity`

NewSupersetSpecAllOfSecurityWithDefaults instantiates a new SupersetSpecAllOfSecurity object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllowPrivilegeEscalation

`func (o *SupersetSpecAllOfSecurity) GetAllowPrivilegeEscalation() bool`

GetAllowPrivilegeEscalation returns the AllowPrivilegeEscalation field if non-nil, zero value otherwise.

### GetAllowPrivilegeEscalationOk

`func (o *SupersetSpecAllOfSecurity) GetAllowPrivilegeEscalationOk() (*bool, bool)`

GetAllowPrivilegeEscalationOk returns a tuple with the AllowPrivilegeEscalation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowPrivilegeEscalation

`func (o *SupersetSpecAllOfSecurity) SetAllowPrivilegeEscalation(v bool)`

SetAllowPrivilegeEscalation sets AllowPrivilegeEscalation field to given value.

### HasAllowPrivilegeEscalation

`func (o *SupersetSpecAllOfSecurity) HasAllowPrivilegeEscalation() bool`

HasAllowPrivilegeEscalation returns a boolean if a field has been set.

### SetAllowPrivilegeEscalationNil

`func (o *SupersetSpecAllOfSecurity) SetAllowPrivilegeEscalationNil(b bool)`

 SetAllowPrivilegeEscalationNil sets the value for AllowPrivilegeEscalation to be an explicit nil

### UnsetAllowPrivilegeEscalation
`func (o *SupersetSpecAllOfSecurity) UnsetAllowPrivilegeEscalation()`

UnsetAllowPrivilegeEscalation ensures that no value is present for AllowPrivilegeEscalation, not even an explicit nil
### GetCapabilities

`func (o *SupersetSpecAllOfSecurity) GetCapabilities() []Capability`

GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.

### GetCapabilitiesOk

`func (o *SupersetSpecAllOfSecurity) GetCapabilitiesOk() (*[]Capability, bool)`

GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapabilities

`func (o *SupersetSpecAllOfSecurity) SetCapabilities(v []Capability)`

SetCapabilities sets Capabilities field to given value.

### HasCapabilities

`func (o *SupersetSpecAllOfSecurity) HasCapabilities() bool`

HasCapabilities returns a boolean if a field has been set.

### SetCapabilitiesNil

`func (o *SupersetSpecAllOfSecurity) SetCapabilitiesNil(b bool)`

 SetCapabilitiesNil sets the value for Capabilities to be an explicit nil

### UnsetCapabilities
`func (o *SupersetSpecAllOfSecurity) UnsetCapabilities()`

UnsetCapabilities ensures that no value is present for Capabilities, not even an explicit nil
### GetHostIpc

`func (o *SupersetSpecAllOfSecurity) GetHostIpc() bool`

GetHostIpc returns the HostIpc field if non-nil, zero value otherwise.

### GetHostIpcOk

`func (o *SupersetSpecAllOfSecurity) GetHostIpcOk() (*bool, bool)`

GetHostIpcOk returns a tuple with the HostIpc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostIpc

`func (o *SupersetSpecAllOfSecurity) SetHostIpc(v bool)`

SetHostIpc sets HostIpc field to given value.

### HasHostIpc

`func (o *SupersetSpecAllOfSecurity) HasHostIpc() bool`

HasHostIpc returns a boolean if a field has been set.

### SetHostIpcNil

`func (o *SupersetSpecAllOfSecurity) SetHostIpcNil(b bool)`

 SetHostIpcNil sets the value for HostIpc to be an explicit nil

### UnsetHostIpc
`func (o *SupersetSpecAllOfSecurity) UnsetHostIpc()`

UnsetHostIpc ensures that no value is present for HostIpc, not even an explicit nil
### GetHostNetwork

`func (o *SupersetSpecAllOfSecurity) GetHostNetwork() bool`

GetHostNetwork returns the HostNetwork field if non-nil, zero value otherwise.

### GetHostNetworkOk

`func (o *SupersetSpecAllOfSecurity) GetHostNetworkOk() (*bool, bool)`

GetHostNetworkOk returns a tuple with the HostNetwork field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostNetwork

`func (o *SupersetSpecAllOfSecurity) SetHostNetwork(v bool)`

SetHostNetwork sets HostNetwork field to given value.

### HasHostNetwork

`func (o *SupersetSpecAllOfSecurity) HasHostNetwork() bool`

HasHostNetwork returns a boolean if a field has been set.

### SetHostNetworkNil

`func (o *SupersetSpecAllOfSecurity) SetHostNetworkNil(b bool)`

 SetHostNetworkNil sets the value for HostNetwork to be an explicit nil

### UnsetHostNetwork
`func (o *SupersetSpecAllOfSecurity) UnsetHostNetwork()`

UnsetHostNetwork ensures that no value is present for HostNetwork, not even an explicit nil
### GetPrivileged

`func (o *SupersetSpecAllOfSecurity) GetPrivileged() bool`

GetPrivileged returns the Privileged field if non-nil, zero value otherwise.

### GetPrivilegedOk

`func (o *SupersetSpecAllOfSecurity) GetPrivilegedOk() (*bool, bool)`

GetPrivilegedOk returns a tuple with the Privileged field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivileged

`func (o *SupersetSpecAllOfSecurity) SetPrivileged(v bool)`

SetPrivileged sets Privileged field to given value.

### HasPrivileged

`func (o *SupersetSpecAllOfSecurity) HasPrivileged() bool`

HasPrivileged returns a boolean if a field has been set.

### SetPrivilegedNil

`func (o *SupersetSpecAllOfSecurity) SetPrivilegedNil(b bool)`

 SetPrivilegedNil sets the value for Privileged to be an explicit nil

### UnsetPrivileged
`func (o *SupersetSpecAllOfSecurity) UnsetPrivileged()`

UnsetPrivileged ensures that no value is present for Privileged, not even an explicit nil
### GetReadOnlyRootFilesystem

`func (o *SupersetSpecAllOfSecurity) GetReadOnlyRootFilesystem() bool`

GetReadOnlyRootFilesystem returns the ReadOnlyRootFilesystem field if non-nil, zero value otherwise.

### GetReadOnlyRootFilesystemOk

`func (o *SupersetSpecAllOfSecurity) GetReadOnlyRootFilesystemOk() (*bool, bool)`

GetReadOnlyRootFilesystemOk returns a tuple with the ReadOnlyRootFilesystem field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnlyRootFilesystem

`func (o *SupersetSpecAllOfSecurity) SetReadOnlyRootFilesystem(v bool)`

SetReadOnlyRootFilesystem sets ReadOnlyRootFilesystem field to given value.

### HasReadOnlyRootFilesystem

`func (o *SupersetSpecAllOfSecurity) HasReadOnlyRootFilesystem() bool`

HasReadOnlyRootFilesystem returns a boolean if a field has been set.

### SetReadOnlyRootFilesystemNil

`func (o *SupersetSpecAllOfSecurity) SetReadOnlyRootFilesystemNil(b bool)`

 SetReadOnlyRootFilesystemNil sets the value for ReadOnlyRootFilesystem to be an explicit nil

### UnsetReadOnlyRootFilesystem
`func (o *SupersetSpecAllOfSecurity) UnsetReadOnlyRootFilesystem()`

UnsetReadOnlyRootFilesystem ensures that no value is present for ReadOnlyRootFilesystem, not even an explicit nil
### GetRunAsGid

`func (o *SupersetSpecAllOfSecurity) GetRunAsGid() int64`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *SupersetSpecAllOfSecurity) GetRunAsGidOk() (*int64, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *SupersetSpecAllOfSecurity) SetRunAsGid(v int64)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *SupersetSpecAllOfSecurity) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *SupersetSpecAllOfSecurity) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *SupersetSpecAllOfSecurity) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetRunAsNonRoot

`func (o *SupersetSpecAllOfSecurity) GetRunAsNonRoot() bool`

GetRunAsNonRoot returns the RunAsNonRoot field if non-nil, zero value otherwise.

### GetRunAsNonRootOk

`func (o *SupersetSpecAllOfSecurity) GetRunAsNonRootOk() (*bool, bool)`

GetRunAsNonRootOk returns a tuple with the RunAsNonRoot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsNonRoot

`func (o *SupersetSpecAllOfSecurity) SetRunAsNonRoot(v bool)`

SetRunAsNonRoot sets RunAsNonRoot field to given value.

### HasRunAsNonRoot

`func (o *SupersetSpecAllOfSecurity) HasRunAsNonRoot() bool`

HasRunAsNonRoot returns a boolean if a field has been set.

### SetRunAsNonRootNil

`func (o *SupersetSpecAllOfSecurity) SetRunAsNonRootNil(b bool)`

 SetRunAsNonRootNil sets the value for RunAsNonRoot to be an explicit nil

### UnsetRunAsNonRoot
`func (o *SupersetSpecAllOfSecurity) UnsetRunAsNonRoot()`

UnsetRunAsNonRoot ensures that no value is present for RunAsNonRoot, not even an explicit nil
### GetRunAsUid

`func (o *SupersetSpecAllOfSecurity) GetRunAsUid() int64`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *SupersetSpecAllOfSecurity) GetRunAsUidOk() (*int64, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *SupersetSpecAllOfSecurity) SetRunAsUid(v int64)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *SupersetSpecAllOfSecurity) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *SupersetSpecAllOfSecurity) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *SupersetSpecAllOfSecurity) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetSeccompProfileType

`func (o *SupersetSpecAllOfSecurity) GetSeccompProfileType() SeccompProfileType`

GetSeccompProfileType returns the SeccompProfileType field if non-nil, zero value otherwise.

### GetSeccompProfileTypeOk

`func (o *SupersetSpecAllOfSecurity) GetSeccompProfileTypeOk() (*SeccompProfileType, bool)`

GetSeccompProfileTypeOk returns a tuple with the SeccompProfileType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeccompProfileType

`func (o *SupersetSpecAllOfSecurity) SetSeccompProfileType(v SeccompProfileType)`

SetSeccompProfileType sets SeccompProfileType field to given value.

### HasSeccompProfileType

`func (o *SupersetSpecAllOfSecurity) HasSeccompProfileType() bool`

HasSeccompProfileType returns a boolean if a field has been set.

### SetSeccompProfileTypeNil

`func (o *SupersetSpecAllOfSecurity) SetSeccompProfileTypeNil(b bool)`

 SetSeccompProfileTypeNil sets the value for SeccompProfileType to be an explicit nil

### UnsetSeccompProfileType
`func (o *SupersetSpecAllOfSecurity) UnsetSeccompProfileType()`

UnsetSeccompProfileType ensures that no value is present for SeccompProfileType, not even an explicit nil
### GetSupplementalGroups

`func (o *SupersetSpecAllOfSecurity) GetSupplementalGroups() string`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *SupersetSpecAllOfSecurity) GetSupplementalGroupsOk() (*string, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *SupersetSpecAllOfSecurity) SetSupplementalGroups(v string)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *SupersetSpecAllOfSecurity) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *SupersetSpecAllOfSecurity) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *SupersetSpecAllOfSecurity) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil
### GetUidGidSource

`func (o *SupersetSpecAllOfSecurity) GetUidGidSource() UidGidSource`

GetUidGidSource returns the UidGidSource field if non-nil, zero value otherwise.

### GetUidGidSourceOk

`func (o *SupersetSpecAllOfSecurity) GetUidGidSourceOk() (*UidGidSource, bool)`

GetUidGidSourceOk returns a tuple with the UidGidSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUidGidSource

`func (o *SupersetSpecAllOfSecurity) SetUidGidSource(v UidGidSource)`

SetUidGidSource sets UidGidSource field to given value.

### HasUidGidSource

`func (o *SupersetSpecAllOfSecurity) HasUidGidSource() bool`

HasUidGidSource returns a boolean if a field has been set.

### SetUidGidSourceNil

`func (o *SupersetSpecAllOfSecurity) SetUidGidSourceNil(b bool)`

 SetUidGidSourceNil sets the value for UidGidSource to be an explicit nil

### UnsetUidGidSource
`func (o *SupersetSpecAllOfSecurity) UnsetUidGidSource()`

UnsetUidGidSource ensures that no value is present for UidGidSource, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


