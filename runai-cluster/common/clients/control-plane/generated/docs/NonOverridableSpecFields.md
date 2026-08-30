# NonOverridableSpecFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Image** | Pointer to **NullableString** | Docker image name. For more information, see [Images](https://kubernetes.io/docs/concepts/containers/images). The image name is mandatory for creating a workspace. | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicy**](ImagePullPolicy.md) |  | [optional] 
**WorkingDir** | Pointer to **NullableString** | Container&#39;s working directory. If not specified, the container runtime default will be used. This may be configured in the container image. | [optional] 
**CreateHomeDir** | Pointer to **NullableBool** | When set to &#x60;true&#x60;, creates a home directory for the container. | [optional] 
**Probes** | Pointer to [**NullableProbes**](Probes.md) |  | [optional] 
**UidGidSource** | Pointer to [**NullableUidGidSource**](UidGidSource.md) |  | [optional] 
**Capabilities** | Pointer to [**[]Capability**](Capability.md) | Add POSIX capabilities to running containers. Defaults to the default set of capabilities granted by the container runtime. | [optional] 
**SeccompProfileType** | Pointer to [**NullableSeccompProfileType**](SeccompProfileType.md) |  | [optional] 
**RunAsNonRoot** | Pointer to **NullableBool** | Force the container to run as a non-root user. | [optional] 
**ReadOnlyRootFilesystem** | Pointer to **NullableBool** | If true, mounts the container&#39;s root filesystem as read-only. | [optional] 
**AllowPrivilegeEscalation** | Pointer to **NullableBool** | Allow the container running the workload and all launched processes to gain additional privileges after the workload starts. For more information consult the User Identity in Container guide at https://docs.run.ai/admin/runai-setup/config/non-root-containers/ | [optional] 
**HostIpc** | Pointer to **NullableBool** | Whether to enable host IPC. Defaults to false. | [optional] 
**HostNetwork** | Pointer to **NullableBool** | Whether to enable host networking. Default to false. | [optional] 
**Connections** | Pointer to [**[]Connection**](Connection.md) | List of connections that either expose ports from the container (each port is associated with a tool that the container runs), or URL&#39;s to be used for connecting to an external tool that is related to the action of the container (such as Weights &amp; Biases). | [optional] 
**OverrideUidGidInWorkspace** | Pointer to **bool** | Allow specifying uid/gid as part of create workspace. This is relevant only for custom uigGidSource. | [optional] [default to false]

## Methods

### NewNonOverridableSpecFields

`func NewNonOverridableSpecFields() *NonOverridableSpecFields`

NewNonOverridableSpecFields instantiates a new NonOverridableSpecFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNonOverridableSpecFieldsWithDefaults

`func NewNonOverridableSpecFieldsWithDefaults() *NonOverridableSpecFields`

NewNonOverridableSpecFieldsWithDefaults instantiates a new NonOverridableSpecFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetImage

`func (o *NonOverridableSpecFields) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *NonOverridableSpecFields) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *NonOverridableSpecFields) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *NonOverridableSpecFields) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *NonOverridableSpecFields) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *NonOverridableSpecFields) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *NonOverridableSpecFields) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *NonOverridableSpecFields) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *NonOverridableSpecFields) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *NonOverridableSpecFields) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *NonOverridableSpecFields) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *NonOverridableSpecFields) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetWorkingDir

`func (o *NonOverridableSpecFields) GetWorkingDir() string`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *NonOverridableSpecFields) GetWorkingDirOk() (*string, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *NonOverridableSpecFields) SetWorkingDir(v string)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *NonOverridableSpecFields) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *NonOverridableSpecFields) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *NonOverridableSpecFields) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil
### GetCreateHomeDir

`func (o *NonOverridableSpecFields) GetCreateHomeDir() bool`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *NonOverridableSpecFields) GetCreateHomeDirOk() (*bool, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *NonOverridableSpecFields) SetCreateHomeDir(v bool)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *NonOverridableSpecFields) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *NonOverridableSpecFields) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *NonOverridableSpecFields) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetProbes

`func (o *NonOverridableSpecFields) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *NonOverridableSpecFields) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *NonOverridableSpecFields) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *NonOverridableSpecFields) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *NonOverridableSpecFields) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *NonOverridableSpecFields) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetUidGidSource

`func (o *NonOverridableSpecFields) GetUidGidSource() UidGidSource`

GetUidGidSource returns the UidGidSource field if non-nil, zero value otherwise.

### GetUidGidSourceOk

`func (o *NonOverridableSpecFields) GetUidGidSourceOk() (*UidGidSource, bool)`

GetUidGidSourceOk returns a tuple with the UidGidSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUidGidSource

`func (o *NonOverridableSpecFields) SetUidGidSource(v UidGidSource)`

SetUidGidSource sets UidGidSource field to given value.

### HasUidGidSource

`func (o *NonOverridableSpecFields) HasUidGidSource() bool`

HasUidGidSource returns a boolean if a field has been set.

### SetUidGidSourceNil

`func (o *NonOverridableSpecFields) SetUidGidSourceNil(b bool)`

 SetUidGidSourceNil sets the value for UidGidSource to be an explicit nil

### UnsetUidGidSource
`func (o *NonOverridableSpecFields) UnsetUidGidSource()`

UnsetUidGidSource ensures that no value is present for UidGidSource, not even an explicit nil
### GetCapabilities

`func (o *NonOverridableSpecFields) GetCapabilities() []Capability`

GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.

### GetCapabilitiesOk

`func (o *NonOverridableSpecFields) GetCapabilitiesOk() (*[]Capability, bool)`

GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapabilities

`func (o *NonOverridableSpecFields) SetCapabilities(v []Capability)`

SetCapabilities sets Capabilities field to given value.

### HasCapabilities

`func (o *NonOverridableSpecFields) HasCapabilities() bool`

HasCapabilities returns a boolean if a field has been set.

### SetCapabilitiesNil

`func (o *NonOverridableSpecFields) SetCapabilitiesNil(b bool)`

 SetCapabilitiesNil sets the value for Capabilities to be an explicit nil

### UnsetCapabilities
`func (o *NonOverridableSpecFields) UnsetCapabilities()`

UnsetCapabilities ensures that no value is present for Capabilities, not even an explicit nil
### GetSeccompProfileType

`func (o *NonOverridableSpecFields) GetSeccompProfileType() SeccompProfileType`

GetSeccompProfileType returns the SeccompProfileType field if non-nil, zero value otherwise.

### GetSeccompProfileTypeOk

`func (o *NonOverridableSpecFields) GetSeccompProfileTypeOk() (*SeccompProfileType, bool)`

GetSeccompProfileTypeOk returns a tuple with the SeccompProfileType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeccompProfileType

`func (o *NonOverridableSpecFields) SetSeccompProfileType(v SeccompProfileType)`

SetSeccompProfileType sets SeccompProfileType field to given value.

### HasSeccompProfileType

`func (o *NonOverridableSpecFields) HasSeccompProfileType() bool`

HasSeccompProfileType returns a boolean if a field has been set.

### SetSeccompProfileTypeNil

`func (o *NonOverridableSpecFields) SetSeccompProfileTypeNil(b bool)`

 SetSeccompProfileTypeNil sets the value for SeccompProfileType to be an explicit nil

### UnsetSeccompProfileType
`func (o *NonOverridableSpecFields) UnsetSeccompProfileType()`

UnsetSeccompProfileType ensures that no value is present for SeccompProfileType, not even an explicit nil
### GetRunAsNonRoot

`func (o *NonOverridableSpecFields) GetRunAsNonRoot() bool`

GetRunAsNonRoot returns the RunAsNonRoot field if non-nil, zero value otherwise.

### GetRunAsNonRootOk

`func (o *NonOverridableSpecFields) GetRunAsNonRootOk() (*bool, bool)`

GetRunAsNonRootOk returns a tuple with the RunAsNonRoot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsNonRoot

`func (o *NonOverridableSpecFields) SetRunAsNonRoot(v bool)`

SetRunAsNonRoot sets RunAsNonRoot field to given value.

### HasRunAsNonRoot

`func (o *NonOverridableSpecFields) HasRunAsNonRoot() bool`

HasRunAsNonRoot returns a boolean if a field has been set.

### SetRunAsNonRootNil

`func (o *NonOverridableSpecFields) SetRunAsNonRootNil(b bool)`

 SetRunAsNonRootNil sets the value for RunAsNonRoot to be an explicit nil

### UnsetRunAsNonRoot
`func (o *NonOverridableSpecFields) UnsetRunAsNonRoot()`

UnsetRunAsNonRoot ensures that no value is present for RunAsNonRoot, not even an explicit nil
### GetReadOnlyRootFilesystem

`func (o *NonOverridableSpecFields) GetReadOnlyRootFilesystem() bool`

GetReadOnlyRootFilesystem returns the ReadOnlyRootFilesystem field if non-nil, zero value otherwise.

### GetReadOnlyRootFilesystemOk

`func (o *NonOverridableSpecFields) GetReadOnlyRootFilesystemOk() (*bool, bool)`

GetReadOnlyRootFilesystemOk returns a tuple with the ReadOnlyRootFilesystem field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnlyRootFilesystem

`func (o *NonOverridableSpecFields) SetReadOnlyRootFilesystem(v bool)`

SetReadOnlyRootFilesystem sets ReadOnlyRootFilesystem field to given value.

### HasReadOnlyRootFilesystem

`func (o *NonOverridableSpecFields) HasReadOnlyRootFilesystem() bool`

HasReadOnlyRootFilesystem returns a boolean if a field has been set.

### SetReadOnlyRootFilesystemNil

`func (o *NonOverridableSpecFields) SetReadOnlyRootFilesystemNil(b bool)`

 SetReadOnlyRootFilesystemNil sets the value for ReadOnlyRootFilesystem to be an explicit nil

### UnsetReadOnlyRootFilesystem
`func (o *NonOverridableSpecFields) UnsetReadOnlyRootFilesystem()`

UnsetReadOnlyRootFilesystem ensures that no value is present for ReadOnlyRootFilesystem, not even an explicit nil
### GetAllowPrivilegeEscalation

`func (o *NonOverridableSpecFields) GetAllowPrivilegeEscalation() bool`

GetAllowPrivilegeEscalation returns the AllowPrivilegeEscalation field if non-nil, zero value otherwise.

### GetAllowPrivilegeEscalationOk

`func (o *NonOverridableSpecFields) GetAllowPrivilegeEscalationOk() (*bool, bool)`

GetAllowPrivilegeEscalationOk returns a tuple with the AllowPrivilegeEscalation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowPrivilegeEscalation

`func (o *NonOverridableSpecFields) SetAllowPrivilegeEscalation(v bool)`

SetAllowPrivilegeEscalation sets AllowPrivilegeEscalation field to given value.

### HasAllowPrivilegeEscalation

`func (o *NonOverridableSpecFields) HasAllowPrivilegeEscalation() bool`

HasAllowPrivilegeEscalation returns a boolean if a field has been set.

### SetAllowPrivilegeEscalationNil

`func (o *NonOverridableSpecFields) SetAllowPrivilegeEscalationNil(b bool)`

 SetAllowPrivilegeEscalationNil sets the value for AllowPrivilegeEscalation to be an explicit nil

### UnsetAllowPrivilegeEscalation
`func (o *NonOverridableSpecFields) UnsetAllowPrivilegeEscalation()`

UnsetAllowPrivilegeEscalation ensures that no value is present for AllowPrivilegeEscalation, not even an explicit nil
### GetHostIpc

`func (o *NonOverridableSpecFields) GetHostIpc() bool`

GetHostIpc returns the HostIpc field if non-nil, zero value otherwise.

### GetHostIpcOk

`func (o *NonOverridableSpecFields) GetHostIpcOk() (*bool, bool)`

GetHostIpcOk returns a tuple with the HostIpc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostIpc

`func (o *NonOverridableSpecFields) SetHostIpc(v bool)`

SetHostIpc sets HostIpc field to given value.

### HasHostIpc

`func (o *NonOverridableSpecFields) HasHostIpc() bool`

HasHostIpc returns a boolean if a field has been set.

### SetHostIpcNil

`func (o *NonOverridableSpecFields) SetHostIpcNil(b bool)`

 SetHostIpcNil sets the value for HostIpc to be an explicit nil

### UnsetHostIpc
`func (o *NonOverridableSpecFields) UnsetHostIpc()`

UnsetHostIpc ensures that no value is present for HostIpc, not even an explicit nil
### GetHostNetwork

`func (o *NonOverridableSpecFields) GetHostNetwork() bool`

GetHostNetwork returns the HostNetwork field if non-nil, zero value otherwise.

### GetHostNetworkOk

`func (o *NonOverridableSpecFields) GetHostNetworkOk() (*bool, bool)`

GetHostNetworkOk returns a tuple with the HostNetwork field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostNetwork

`func (o *NonOverridableSpecFields) SetHostNetwork(v bool)`

SetHostNetwork sets HostNetwork field to given value.

### HasHostNetwork

`func (o *NonOverridableSpecFields) HasHostNetwork() bool`

HasHostNetwork returns a boolean if a field has been set.

### SetHostNetworkNil

`func (o *NonOverridableSpecFields) SetHostNetworkNil(b bool)`

 SetHostNetworkNil sets the value for HostNetwork to be an explicit nil

### UnsetHostNetwork
`func (o *NonOverridableSpecFields) UnsetHostNetwork()`

UnsetHostNetwork ensures that no value is present for HostNetwork, not even an explicit nil
### GetConnections

`func (o *NonOverridableSpecFields) GetConnections() []Connection`

GetConnections returns the Connections field if non-nil, zero value otherwise.

### GetConnectionsOk

`func (o *NonOverridableSpecFields) GetConnectionsOk() (*[]Connection, bool)`

GetConnectionsOk returns a tuple with the Connections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnections

`func (o *NonOverridableSpecFields) SetConnections(v []Connection)`

SetConnections sets Connections field to given value.

### HasConnections

`func (o *NonOverridableSpecFields) HasConnections() bool`

HasConnections returns a boolean if a field has been set.

### GetOverrideUidGidInWorkspace

`func (o *NonOverridableSpecFields) GetOverrideUidGidInWorkspace() bool`

GetOverrideUidGidInWorkspace returns the OverrideUidGidInWorkspace field if non-nil, zero value otherwise.

### GetOverrideUidGidInWorkspaceOk

`func (o *NonOverridableSpecFields) GetOverrideUidGidInWorkspaceOk() (*bool, bool)`

GetOverrideUidGidInWorkspaceOk returns a tuple with the OverrideUidGidInWorkspace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverrideUidGidInWorkspace

`func (o *NonOverridableSpecFields) SetOverrideUidGidInWorkspace(v bool)`

SetOverrideUidGidInWorkspace sets OverrideUidGidInWorkspace field to given value.

### HasOverrideUidGidInWorkspace

`func (o *NonOverridableSpecFields) HasOverrideUidGidInWorkspace() bool`

HasOverrideUidGidInWorkspace returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


