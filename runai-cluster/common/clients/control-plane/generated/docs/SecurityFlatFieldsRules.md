# SecurityFlatFieldsRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**UidGidSource** | Pointer to [**NullableUidGidSourceRules**](UidGidSourceRules.md) |  | [optional] 
**Capabilities** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**SeccompProfileType** | Pointer to [**NullableSeccompProfileTypeRules**](SeccompProfileTypeRules.md) |  | [optional] 
**RunAsNonRoot** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**ReadOnlyRootFilesystem** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**RunAsUid** | Pointer to [**NullableIntegerRulesOptional**](IntegerRulesOptional.md) |  | [optional] 
**RunAsGid** | Pointer to [**NullableIntegerRulesOptional**](IntegerRulesOptional.md) |  | [optional] 
**SupplementalGroups** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**AllowPrivilegeEscalation** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**HostIpc** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**HostNetwork** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 

## Methods

### NewSecurityFlatFieldsRules

`func NewSecurityFlatFieldsRules() *SecurityFlatFieldsRules`

NewSecurityFlatFieldsRules instantiates a new SecurityFlatFieldsRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecurityFlatFieldsRulesWithDefaults

`func NewSecurityFlatFieldsRulesWithDefaults() *SecurityFlatFieldsRules`

NewSecurityFlatFieldsRulesWithDefaults instantiates a new SecurityFlatFieldsRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUidGidSource

`func (o *SecurityFlatFieldsRules) GetUidGidSource() UidGidSourceRules`

GetUidGidSource returns the UidGidSource field if non-nil, zero value otherwise.

### GetUidGidSourceOk

`func (o *SecurityFlatFieldsRules) GetUidGidSourceOk() (*UidGidSourceRules, bool)`

GetUidGidSourceOk returns a tuple with the UidGidSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUidGidSource

`func (o *SecurityFlatFieldsRules) SetUidGidSource(v UidGidSourceRules)`

SetUidGidSource sets UidGidSource field to given value.

### HasUidGidSource

`func (o *SecurityFlatFieldsRules) HasUidGidSource() bool`

HasUidGidSource returns a boolean if a field has been set.

### SetUidGidSourceNil

`func (o *SecurityFlatFieldsRules) SetUidGidSourceNil(b bool)`

 SetUidGidSourceNil sets the value for UidGidSource to be an explicit nil

### UnsetUidGidSource
`func (o *SecurityFlatFieldsRules) UnsetUidGidSource()`

UnsetUidGidSource ensures that no value is present for UidGidSource, not even an explicit nil
### GetCapabilities

`func (o *SecurityFlatFieldsRules) GetCapabilities() ArrayRules`

GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.

### GetCapabilitiesOk

`func (o *SecurityFlatFieldsRules) GetCapabilitiesOk() (*ArrayRules, bool)`

GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapabilities

`func (o *SecurityFlatFieldsRules) SetCapabilities(v ArrayRules)`

SetCapabilities sets Capabilities field to given value.

### HasCapabilities

`func (o *SecurityFlatFieldsRules) HasCapabilities() bool`

HasCapabilities returns a boolean if a field has been set.

### SetCapabilitiesNil

`func (o *SecurityFlatFieldsRules) SetCapabilitiesNil(b bool)`

 SetCapabilitiesNil sets the value for Capabilities to be an explicit nil

### UnsetCapabilities
`func (o *SecurityFlatFieldsRules) UnsetCapabilities()`

UnsetCapabilities ensures that no value is present for Capabilities, not even an explicit nil
### GetSeccompProfileType

`func (o *SecurityFlatFieldsRules) GetSeccompProfileType() SeccompProfileTypeRules`

GetSeccompProfileType returns the SeccompProfileType field if non-nil, zero value otherwise.

### GetSeccompProfileTypeOk

`func (o *SecurityFlatFieldsRules) GetSeccompProfileTypeOk() (*SeccompProfileTypeRules, bool)`

GetSeccompProfileTypeOk returns a tuple with the SeccompProfileType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeccompProfileType

`func (o *SecurityFlatFieldsRules) SetSeccompProfileType(v SeccompProfileTypeRules)`

SetSeccompProfileType sets SeccompProfileType field to given value.

### HasSeccompProfileType

`func (o *SecurityFlatFieldsRules) HasSeccompProfileType() bool`

HasSeccompProfileType returns a boolean if a field has been set.

### SetSeccompProfileTypeNil

`func (o *SecurityFlatFieldsRules) SetSeccompProfileTypeNil(b bool)`

 SetSeccompProfileTypeNil sets the value for SeccompProfileType to be an explicit nil

### UnsetSeccompProfileType
`func (o *SecurityFlatFieldsRules) UnsetSeccompProfileType()`

UnsetSeccompProfileType ensures that no value is present for SeccompProfileType, not even an explicit nil
### GetRunAsNonRoot

`func (o *SecurityFlatFieldsRules) GetRunAsNonRoot() BooleanRules`

GetRunAsNonRoot returns the RunAsNonRoot field if non-nil, zero value otherwise.

### GetRunAsNonRootOk

`func (o *SecurityFlatFieldsRules) GetRunAsNonRootOk() (*BooleanRules, bool)`

GetRunAsNonRootOk returns a tuple with the RunAsNonRoot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsNonRoot

`func (o *SecurityFlatFieldsRules) SetRunAsNonRoot(v BooleanRules)`

SetRunAsNonRoot sets RunAsNonRoot field to given value.

### HasRunAsNonRoot

`func (o *SecurityFlatFieldsRules) HasRunAsNonRoot() bool`

HasRunAsNonRoot returns a boolean if a field has been set.

### SetRunAsNonRootNil

`func (o *SecurityFlatFieldsRules) SetRunAsNonRootNil(b bool)`

 SetRunAsNonRootNil sets the value for RunAsNonRoot to be an explicit nil

### UnsetRunAsNonRoot
`func (o *SecurityFlatFieldsRules) UnsetRunAsNonRoot()`

UnsetRunAsNonRoot ensures that no value is present for RunAsNonRoot, not even an explicit nil
### GetReadOnlyRootFilesystem

`func (o *SecurityFlatFieldsRules) GetReadOnlyRootFilesystem() BooleanRules`

GetReadOnlyRootFilesystem returns the ReadOnlyRootFilesystem field if non-nil, zero value otherwise.

### GetReadOnlyRootFilesystemOk

`func (o *SecurityFlatFieldsRules) GetReadOnlyRootFilesystemOk() (*BooleanRules, bool)`

GetReadOnlyRootFilesystemOk returns a tuple with the ReadOnlyRootFilesystem field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnlyRootFilesystem

`func (o *SecurityFlatFieldsRules) SetReadOnlyRootFilesystem(v BooleanRules)`

SetReadOnlyRootFilesystem sets ReadOnlyRootFilesystem field to given value.

### HasReadOnlyRootFilesystem

`func (o *SecurityFlatFieldsRules) HasReadOnlyRootFilesystem() bool`

HasReadOnlyRootFilesystem returns a boolean if a field has been set.

### SetReadOnlyRootFilesystemNil

`func (o *SecurityFlatFieldsRules) SetReadOnlyRootFilesystemNil(b bool)`

 SetReadOnlyRootFilesystemNil sets the value for ReadOnlyRootFilesystem to be an explicit nil

### UnsetReadOnlyRootFilesystem
`func (o *SecurityFlatFieldsRules) UnsetReadOnlyRootFilesystem()`

UnsetReadOnlyRootFilesystem ensures that no value is present for ReadOnlyRootFilesystem, not even an explicit nil
### GetRunAsUid

`func (o *SecurityFlatFieldsRules) GetRunAsUid() IntegerRulesOptional`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *SecurityFlatFieldsRules) GetRunAsUidOk() (*IntegerRulesOptional, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *SecurityFlatFieldsRules) SetRunAsUid(v IntegerRulesOptional)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *SecurityFlatFieldsRules) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *SecurityFlatFieldsRules) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *SecurityFlatFieldsRules) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetRunAsGid

`func (o *SecurityFlatFieldsRules) GetRunAsGid() IntegerRulesOptional`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *SecurityFlatFieldsRules) GetRunAsGidOk() (*IntegerRulesOptional, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *SecurityFlatFieldsRules) SetRunAsGid(v IntegerRulesOptional)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *SecurityFlatFieldsRules) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *SecurityFlatFieldsRules) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *SecurityFlatFieldsRules) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetSupplementalGroups

`func (o *SecurityFlatFieldsRules) GetSupplementalGroups() StringRules`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *SecurityFlatFieldsRules) GetSupplementalGroupsOk() (*StringRules, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *SecurityFlatFieldsRules) SetSupplementalGroups(v StringRules)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *SecurityFlatFieldsRules) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *SecurityFlatFieldsRules) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *SecurityFlatFieldsRules) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil
### GetAllowPrivilegeEscalation

`func (o *SecurityFlatFieldsRules) GetAllowPrivilegeEscalation() BooleanRules`

GetAllowPrivilegeEscalation returns the AllowPrivilegeEscalation field if non-nil, zero value otherwise.

### GetAllowPrivilegeEscalationOk

`func (o *SecurityFlatFieldsRules) GetAllowPrivilegeEscalationOk() (*BooleanRules, bool)`

GetAllowPrivilegeEscalationOk returns a tuple with the AllowPrivilegeEscalation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowPrivilegeEscalation

`func (o *SecurityFlatFieldsRules) SetAllowPrivilegeEscalation(v BooleanRules)`

SetAllowPrivilegeEscalation sets AllowPrivilegeEscalation field to given value.

### HasAllowPrivilegeEscalation

`func (o *SecurityFlatFieldsRules) HasAllowPrivilegeEscalation() bool`

HasAllowPrivilegeEscalation returns a boolean if a field has been set.

### SetAllowPrivilegeEscalationNil

`func (o *SecurityFlatFieldsRules) SetAllowPrivilegeEscalationNil(b bool)`

 SetAllowPrivilegeEscalationNil sets the value for AllowPrivilegeEscalation to be an explicit nil

### UnsetAllowPrivilegeEscalation
`func (o *SecurityFlatFieldsRules) UnsetAllowPrivilegeEscalation()`

UnsetAllowPrivilegeEscalation ensures that no value is present for AllowPrivilegeEscalation, not even an explicit nil
### GetHostIpc

`func (o *SecurityFlatFieldsRules) GetHostIpc() BooleanRules`

GetHostIpc returns the HostIpc field if non-nil, zero value otherwise.

### GetHostIpcOk

`func (o *SecurityFlatFieldsRules) GetHostIpcOk() (*BooleanRules, bool)`

GetHostIpcOk returns a tuple with the HostIpc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostIpc

`func (o *SecurityFlatFieldsRules) SetHostIpc(v BooleanRules)`

SetHostIpc sets HostIpc field to given value.

### HasHostIpc

`func (o *SecurityFlatFieldsRules) HasHostIpc() bool`

HasHostIpc returns a boolean if a field has been set.

### SetHostIpcNil

`func (o *SecurityFlatFieldsRules) SetHostIpcNil(b bool)`

 SetHostIpcNil sets the value for HostIpc to be an explicit nil

### UnsetHostIpc
`func (o *SecurityFlatFieldsRules) UnsetHostIpc()`

UnsetHostIpc ensures that no value is present for HostIpc, not even an explicit nil
### GetHostNetwork

`func (o *SecurityFlatFieldsRules) GetHostNetwork() BooleanRules`

GetHostNetwork returns the HostNetwork field if non-nil, zero value otherwise.

### GetHostNetworkOk

`func (o *SecurityFlatFieldsRules) GetHostNetworkOk() (*BooleanRules, bool)`

GetHostNetworkOk returns a tuple with the HostNetwork field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostNetwork

`func (o *SecurityFlatFieldsRules) SetHostNetwork(v BooleanRules)`

SetHostNetwork sets HostNetwork field to given value.

### HasHostNetwork

`func (o *SecurityFlatFieldsRules) HasHostNetwork() bool`

HasHostNetwork returns a boolean if a field has been set.

### SetHostNetworkNil

`func (o *SecurityFlatFieldsRules) SetHostNetworkNil(b bool)`

 SetHostNetworkNil sets the value for HostNetwork to be an explicit nil

### UnsetHostNetwork
`func (o *SecurityFlatFieldsRules) UnsetHostNetwork()`

UnsetHostNetwork ensures that no value is present for HostNetwork, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


