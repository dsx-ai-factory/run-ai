# DistributedInferenceLeaderWorkerRulesV2Security

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Capabilities** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**HostIpc** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**HostNetwork** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**Privileged** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**ReadOnlyRootFilesystem** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**RunAsGid** | Pointer to [**NullableIntegerRulesOptional**](IntegerRulesOptional.md) |  | [optional] 
**RunAsNonRoot** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**RunAsUid** | Pointer to [**NullableIntegerRulesOptional**](IntegerRulesOptional.md) |  | [optional] 
**SeccompProfileType** | Pointer to [**NullableSeccompProfileTypeRules**](SeccompProfileTypeRules.md) |  | [optional] 
**SupplementalGroups** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**UidGidSource** | Pointer to [**NullableUidGidSourceRules**](UidGidSourceRules.md) |  | [optional] 

## Methods

### NewDistributedInferenceLeaderWorkerRulesV2Security

`func NewDistributedInferenceLeaderWorkerRulesV2Security() *DistributedInferenceLeaderWorkerRulesV2Security`

NewDistributedInferenceLeaderWorkerRulesV2Security instantiates a new DistributedInferenceLeaderWorkerRulesV2Security object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceLeaderWorkerRulesV2SecurityWithDefaults

`func NewDistributedInferenceLeaderWorkerRulesV2SecurityWithDefaults() *DistributedInferenceLeaderWorkerRulesV2Security`

NewDistributedInferenceLeaderWorkerRulesV2SecurityWithDefaults instantiates a new DistributedInferenceLeaderWorkerRulesV2Security object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCapabilities

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetCapabilities() ArrayRules`

GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.

### GetCapabilitiesOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetCapabilitiesOk() (*ArrayRules, bool)`

GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapabilities

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetCapabilities(v ArrayRules)`

SetCapabilities sets Capabilities field to given value.

### HasCapabilities

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) HasCapabilities() bool`

HasCapabilities returns a boolean if a field has been set.

### SetCapabilitiesNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetCapabilitiesNil(b bool)`

 SetCapabilitiesNil sets the value for Capabilities to be an explicit nil

### UnsetCapabilities
`func (o *DistributedInferenceLeaderWorkerRulesV2Security) UnsetCapabilities()`

UnsetCapabilities ensures that no value is present for Capabilities, not even an explicit nil
### GetHostIpc

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetHostIpc() BooleanRules`

GetHostIpc returns the HostIpc field if non-nil, zero value otherwise.

### GetHostIpcOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetHostIpcOk() (*BooleanRules, bool)`

GetHostIpcOk returns a tuple with the HostIpc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostIpc

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetHostIpc(v BooleanRules)`

SetHostIpc sets HostIpc field to given value.

### HasHostIpc

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) HasHostIpc() bool`

HasHostIpc returns a boolean if a field has been set.

### SetHostIpcNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetHostIpcNil(b bool)`

 SetHostIpcNil sets the value for HostIpc to be an explicit nil

### UnsetHostIpc
`func (o *DistributedInferenceLeaderWorkerRulesV2Security) UnsetHostIpc()`

UnsetHostIpc ensures that no value is present for HostIpc, not even an explicit nil
### GetHostNetwork

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetHostNetwork() BooleanRules`

GetHostNetwork returns the HostNetwork field if non-nil, zero value otherwise.

### GetHostNetworkOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetHostNetworkOk() (*BooleanRules, bool)`

GetHostNetworkOk returns a tuple with the HostNetwork field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostNetwork

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetHostNetwork(v BooleanRules)`

SetHostNetwork sets HostNetwork field to given value.

### HasHostNetwork

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) HasHostNetwork() bool`

HasHostNetwork returns a boolean if a field has been set.

### SetHostNetworkNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetHostNetworkNil(b bool)`

 SetHostNetworkNil sets the value for HostNetwork to be an explicit nil

### UnsetHostNetwork
`func (o *DistributedInferenceLeaderWorkerRulesV2Security) UnsetHostNetwork()`

UnsetHostNetwork ensures that no value is present for HostNetwork, not even an explicit nil
### GetPrivileged

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetPrivileged() BooleanRules`

GetPrivileged returns the Privileged field if non-nil, zero value otherwise.

### GetPrivilegedOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetPrivilegedOk() (*BooleanRules, bool)`

GetPrivilegedOk returns a tuple with the Privileged field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivileged

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetPrivileged(v BooleanRules)`

SetPrivileged sets Privileged field to given value.

### HasPrivileged

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) HasPrivileged() bool`

HasPrivileged returns a boolean if a field has been set.

### SetPrivilegedNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetPrivilegedNil(b bool)`

 SetPrivilegedNil sets the value for Privileged to be an explicit nil

### UnsetPrivileged
`func (o *DistributedInferenceLeaderWorkerRulesV2Security) UnsetPrivileged()`

UnsetPrivileged ensures that no value is present for Privileged, not even an explicit nil
### GetReadOnlyRootFilesystem

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetReadOnlyRootFilesystem() BooleanRules`

GetReadOnlyRootFilesystem returns the ReadOnlyRootFilesystem field if non-nil, zero value otherwise.

### GetReadOnlyRootFilesystemOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetReadOnlyRootFilesystemOk() (*BooleanRules, bool)`

GetReadOnlyRootFilesystemOk returns a tuple with the ReadOnlyRootFilesystem field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnlyRootFilesystem

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetReadOnlyRootFilesystem(v BooleanRules)`

SetReadOnlyRootFilesystem sets ReadOnlyRootFilesystem field to given value.

### HasReadOnlyRootFilesystem

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) HasReadOnlyRootFilesystem() bool`

HasReadOnlyRootFilesystem returns a boolean if a field has been set.

### SetReadOnlyRootFilesystemNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetReadOnlyRootFilesystemNil(b bool)`

 SetReadOnlyRootFilesystemNil sets the value for ReadOnlyRootFilesystem to be an explicit nil

### UnsetReadOnlyRootFilesystem
`func (o *DistributedInferenceLeaderWorkerRulesV2Security) UnsetReadOnlyRootFilesystem()`

UnsetReadOnlyRootFilesystem ensures that no value is present for ReadOnlyRootFilesystem, not even an explicit nil
### GetRunAsGid

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetRunAsGid() IntegerRulesOptional`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetRunAsGidOk() (*IntegerRulesOptional, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetRunAsGid(v IntegerRulesOptional)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *DistributedInferenceLeaderWorkerRulesV2Security) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetRunAsNonRoot

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetRunAsNonRoot() BooleanRules`

GetRunAsNonRoot returns the RunAsNonRoot field if non-nil, zero value otherwise.

### GetRunAsNonRootOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetRunAsNonRootOk() (*BooleanRules, bool)`

GetRunAsNonRootOk returns a tuple with the RunAsNonRoot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsNonRoot

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetRunAsNonRoot(v BooleanRules)`

SetRunAsNonRoot sets RunAsNonRoot field to given value.

### HasRunAsNonRoot

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) HasRunAsNonRoot() bool`

HasRunAsNonRoot returns a boolean if a field has been set.

### SetRunAsNonRootNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetRunAsNonRootNil(b bool)`

 SetRunAsNonRootNil sets the value for RunAsNonRoot to be an explicit nil

### UnsetRunAsNonRoot
`func (o *DistributedInferenceLeaderWorkerRulesV2Security) UnsetRunAsNonRoot()`

UnsetRunAsNonRoot ensures that no value is present for RunAsNonRoot, not even an explicit nil
### GetRunAsUid

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetRunAsUid() IntegerRulesOptional`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetRunAsUidOk() (*IntegerRulesOptional, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetRunAsUid(v IntegerRulesOptional)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *DistributedInferenceLeaderWorkerRulesV2Security) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetSeccompProfileType

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetSeccompProfileType() SeccompProfileTypeRules`

GetSeccompProfileType returns the SeccompProfileType field if non-nil, zero value otherwise.

### GetSeccompProfileTypeOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetSeccompProfileTypeOk() (*SeccompProfileTypeRules, bool)`

GetSeccompProfileTypeOk returns a tuple with the SeccompProfileType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeccompProfileType

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetSeccompProfileType(v SeccompProfileTypeRules)`

SetSeccompProfileType sets SeccompProfileType field to given value.

### HasSeccompProfileType

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) HasSeccompProfileType() bool`

HasSeccompProfileType returns a boolean if a field has been set.

### SetSeccompProfileTypeNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetSeccompProfileTypeNil(b bool)`

 SetSeccompProfileTypeNil sets the value for SeccompProfileType to be an explicit nil

### UnsetSeccompProfileType
`func (o *DistributedInferenceLeaderWorkerRulesV2Security) UnsetSeccompProfileType()`

UnsetSeccompProfileType ensures that no value is present for SeccompProfileType, not even an explicit nil
### GetSupplementalGroups

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetSupplementalGroups() StringRules`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetSupplementalGroupsOk() (*StringRules, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetSupplementalGroups(v StringRules)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *DistributedInferenceLeaderWorkerRulesV2Security) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil
### GetUidGidSource

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetUidGidSource() UidGidSourceRules`

GetUidGidSource returns the UidGidSource field if non-nil, zero value otherwise.

### GetUidGidSourceOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) GetUidGidSourceOk() (*UidGidSourceRules, bool)`

GetUidGidSourceOk returns a tuple with the UidGidSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUidGidSource

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetUidGidSource(v UidGidSourceRules)`

SetUidGidSource sets UidGidSource field to given value.

### HasUidGidSource

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) HasUidGidSource() bool`

HasUidGidSource returns a boolean if a field has been set.

### SetUidGidSourceNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Security) SetUidGidSourceNil(b bool)`

 SetUidGidSourceNil sets the value for UidGidSource to be an explicit nil

### UnsetUidGidSource
`func (o *DistributedInferenceLeaderWorkerRulesV2Security) UnsetUidGidSource()`

UnsetUidGidSource ensures that no value is present for UidGidSource, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


