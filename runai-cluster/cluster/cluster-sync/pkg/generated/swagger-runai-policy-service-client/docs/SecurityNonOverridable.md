# SecurityNonOverridable

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowPrivilegeEscalation** | Pointer to **NullableBool** | Allow the container running the workload and all launched processes to gain additional privileges after the workload starts. For more information consult the User Identity in Container guide at https://docs.run.ai/admin/runai-setup/config/non-root-containers/ | [optional] 
**UidGidSource** | Pointer to [**NullableUidGidSource**](UidGidSource.md) |  | [optional] 
**Capabilities** | Pointer to [**[]Capability**](Capability.md) | The POSIX capabilities to add when running containers. Defaults to the default set of capabilities granted by the container runtime. | [optional] 
**SeccompProfileType** | Pointer to [**NullableSeccompProfileType**](SeccompProfileType.md) |  | [optional] 
**RunAsNonRoot** | Pointer to **NullableBool** | Indicates that the container must run as a non-root user. | [optional] 
**HostIpc** | Pointer to **NullableBool** | Whether to enable host IPC. Defaults to false. | [optional] 
**HostNetwork** | Pointer to **NullableBool** | Whether to enable host networking. Default to false. | [optional] 

## Methods

### NewSecurityNonOverridable

`func NewSecurityNonOverridable() *SecurityNonOverridable`

NewSecurityNonOverridable instantiates a new SecurityNonOverridable object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecurityNonOverridableWithDefaults

`func NewSecurityNonOverridableWithDefaults() *SecurityNonOverridable`

NewSecurityNonOverridableWithDefaults instantiates a new SecurityNonOverridable object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllowPrivilegeEscalation

`func (o *SecurityNonOverridable) GetAllowPrivilegeEscalation() bool`

GetAllowPrivilegeEscalation returns the AllowPrivilegeEscalation field if non-nil, zero value otherwise.

### GetAllowPrivilegeEscalationOk

`func (o *SecurityNonOverridable) GetAllowPrivilegeEscalationOk() (*bool, bool)`

GetAllowPrivilegeEscalationOk returns a tuple with the AllowPrivilegeEscalation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowPrivilegeEscalation

`func (o *SecurityNonOverridable) SetAllowPrivilegeEscalation(v bool)`

SetAllowPrivilegeEscalation sets AllowPrivilegeEscalation field to given value.

### HasAllowPrivilegeEscalation

`func (o *SecurityNonOverridable) HasAllowPrivilegeEscalation() bool`

HasAllowPrivilegeEscalation returns a boolean if a field has been set.

### SetAllowPrivilegeEscalationNil

`func (o *SecurityNonOverridable) SetAllowPrivilegeEscalationNil(b bool)`

 SetAllowPrivilegeEscalationNil sets the value for AllowPrivilegeEscalation to be an explicit nil

### UnsetAllowPrivilegeEscalation
`func (o *SecurityNonOverridable) UnsetAllowPrivilegeEscalation()`

UnsetAllowPrivilegeEscalation ensures that no value is present for AllowPrivilegeEscalation, not even an explicit nil
### GetUidGidSource

`func (o *SecurityNonOverridable) GetUidGidSource() UidGidSource`

GetUidGidSource returns the UidGidSource field if non-nil, zero value otherwise.

### GetUidGidSourceOk

`func (o *SecurityNonOverridable) GetUidGidSourceOk() (*UidGidSource, bool)`

GetUidGidSourceOk returns a tuple with the UidGidSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUidGidSource

`func (o *SecurityNonOverridable) SetUidGidSource(v UidGidSource)`

SetUidGidSource sets UidGidSource field to given value.

### HasUidGidSource

`func (o *SecurityNonOverridable) HasUidGidSource() bool`

HasUidGidSource returns a boolean if a field has been set.

### SetUidGidSourceNil

`func (o *SecurityNonOverridable) SetUidGidSourceNil(b bool)`

 SetUidGidSourceNil sets the value for UidGidSource to be an explicit nil

### UnsetUidGidSource
`func (o *SecurityNonOverridable) UnsetUidGidSource()`

UnsetUidGidSource ensures that no value is present for UidGidSource, not even an explicit nil
### GetCapabilities

`func (o *SecurityNonOverridable) GetCapabilities() []Capability`

GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.

### GetCapabilitiesOk

`func (o *SecurityNonOverridable) GetCapabilitiesOk() (*[]Capability, bool)`

GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapabilities

`func (o *SecurityNonOverridable) SetCapabilities(v []Capability)`

SetCapabilities sets Capabilities field to given value.

### HasCapabilities

`func (o *SecurityNonOverridable) HasCapabilities() bool`

HasCapabilities returns a boolean if a field has been set.

### SetCapabilitiesNil

`func (o *SecurityNonOverridable) SetCapabilitiesNil(b bool)`

 SetCapabilitiesNil sets the value for Capabilities to be an explicit nil

### UnsetCapabilities
`func (o *SecurityNonOverridable) UnsetCapabilities()`

UnsetCapabilities ensures that no value is present for Capabilities, not even an explicit nil
### GetSeccompProfileType

`func (o *SecurityNonOverridable) GetSeccompProfileType() SeccompProfileType`

GetSeccompProfileType returns the SeccompProfileType field if non-nil, zero value otherwise.

### GetSeccompProfileTypeOk

`func (o *SecurityNonOverridable) GetSeccompProfileTypeOk() (*SeccompProfileType, bool)`

GetSeccompProfileTypeOk returns a tuple with the SeccompProfileType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeccompProfileType

`func (o *SecurityNonOverridable) SetSeccompProfileType(v SeccompProfileType)`

SetSeccompProfileType sets SeccompProfileType field to given value.

### HasSeccompProfileType

`func (o *SecurityNonOverridable) HasSeccompProfileType() bool`

HasSeccompProfileType returns a boolean if a field has been set.

### SetSeccompProfileTypeNil

`func (o *SecurityNonOverridable) SetSeccompProfileTypeNil(b bool)`

 SetSeccompProfileTypeNil sets the value for SeccompProfileType to be an explicit nil

### UnsetSeccompProfileType
`func (o *SecurityNonOverridable) UnsetSeccompProfileType()`

UnsetSeccompProfileType ensures that no value is present for SeccompProfileType, not even an explicit nil
### GetRunAsNonRoot

`func (o *SecurityNonOverridable) GetRunAsNonRoot() bool`

GetRunAsNonRoot returns the RunAsNonRoot field if non-nil, zero value otherwise.

### GetRunAsNonRootOk

`func (o *SecurityNonOverridable) GetRunAsNonRootOk() (*bool, bool)`

GetRunAsNonRootOk returns a tuple with the RunAsNonRoot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsNonRoot

`func (o *SecurityNonOverridable) SetRunAsNonRoot(v bool)`

SetRunAsNonRoot sets RunAsNonRoot field to given value.

### HasRunAsNonRoot

`func (o *SecurityNonOverridable) HasRunAsNonRoot() bool`

HasRunAsNonRoot returns a boolean if a field has been set.

### SetRunAsNonRootNil

`func (o *SecurityNonOverridable) SetRunAsNonRootNil(b bool)`

 SetRunAsNonRootNil sets the value for RunAsNonRoot to be an explicit nil

### UnsetRunAsNonRoot
`func (o *SecurityNonOverridable) UnsetRunAsNonRoot()`

UnsetRunAsNonRoot ensures that no value is present for RunAsNonRoot, not even an explicit nil
### GetHostIpc

`func (o *SecurityNonOverridable) GetHostIpc() bool`

GetHostIpc returns the HostIpc field if non-nil, zero value otherwise.

### GetHostIpcOk

`func (o *SecurityNonOverridable) GetHostIpcOk() (*bool, bool)`

GetHostIpcOk returns a tuple with the HostIpc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostIpc

`func (o *SecurityNonOverridable) SetHostIpc(v bool)`

SetHostIpc sets HostIpc field to given value.

### HasHostIpc

`func (o *SecurityNonOverridable) HasHostIpc() bool`

HasHostIpc returns a boolean if a field has been set.

### SetHostIpcNil

`func (o *SecurityNonOverridable) SetHostIpcNil(b bool)`

 SetHostIpcNil sets the value for HostIpc to be an explicit nil

### UnsetHostIpc
`func (o *SecurityNonOverridable) UnsetHostIpc()`

UnsetHostIpc ensures that no value is present for HostIpc, not even an explicit nil
### GetHostNetwork

`func (o *SecurityNonOverridable) GetHostNetwork() bool`

GetHostNetwork returns the HostNetwork field if non-nil, zero value otherwise.

### GetHostNetworkOk

`func (o *SecurityNonOverridable) GetHostNetworkOk() (*bool, bool)`

GetHostNetworkOk returns a tuple with the HostNetwork field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostNetwork

`func (o *SecurityNonOverridable) SetHostNetwork(v bool)`

SetHostNetwork sets HostNetwork field to given value.

### HasHostNetwork

`func (o *SecurityNonOverridable) HasHostNetwork() bool`

HasHostNetwork returns a boolean if a field has been set.

### SetHostNetworkNil

`func (o *SecurityNonOverridable) SetHostNetworkNil(b bool)`

 SetHostNetworkNil sets the value for HostNetwork to be an explicit nil

### UnsetHostNetwork
`func (o *SecurityNonOverridable) UnsetHostNetwork()`

UnsetHostNetwork ensures that no value is present for HostNetwork, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


