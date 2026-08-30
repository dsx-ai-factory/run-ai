# InferencePolicyRulesV2AllOfSecurity

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Capabilities** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**HostIpc** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**HostNetwork** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**ReadOnlyRootFilesystem** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**RunAsGid** | Pointer to [**NullableIntegerRulesOptional**](IntegerRulesOptional.md) |  | [optional] 
**RunAsNonRoot** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**RunAsUid** | Pointer to [**NullableIntegerRulesOptional**](IntegerRulesOptional.md) |  | [optional] 
**SeccompProfileType** | Pointer to [**NullableSeccompProfileTypeRules**](SeccompProfileTypeRules.md) |  | [optional] 
**SupplementalGroups** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**UidGidSource** | Pointer to [**NullableUidGidSourceRules**](UidGidSourceRules.md) |  | [optional] 

## Methods

### NewInferencePolicyRulesV2AllOfSecurity

`func NewInferencePolicyRulesV2AllOfSecurity() *InferencePolicyRulesV2AllOfSecurity`

NewInferencePolicyRulesV2AllOfSecurity instantiates a new InferencePolicyRulesV2AllOfSecurity object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferencePolicyRulesV2AllOfSecurityWithDefaults

`func NewInferencePolicyRulesV2AllOfSecurityWithDefaults() *InferencePolicyRulesV2AllOfSecurity`

NewInferencePolicyRulesV2AllOfSecurityWithDefaults instantiates a new InferencePolicyRulesV2AllOfSecurity object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCapabilities

`func (o *InferencePolicyRulesV2AllOfSecurity) GetCapabilities() ArrayRules`

GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.

### GetCapabilitiesOk

`func (o *InferencePolicyRulesV2AllOfSecurity) GetCapabilitiesOk() (*ArrayRules, bool)`

GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapabilities

`func (o *InferencePolicyRulesV2AllOfSecurity) SetCapabilities(v ArrayRules)`

SetCapabilities sets Capabilities field to given value.

### HasCapabilities

`func (o *InferencePolicyRulesV2AllOfSecurity) HasCapabilities() bool`

HasCapabilities returns a boolean if a field has been set.

### SetCapabilitiesNil

`func (o *InferencePolicyRulesV2AllOfSecurity) SetCapabilitiesNil(b bool)`

 SetCapabilitiesNil sets the value for Capabilities to be an explicit nil

### UnsetCapabilities
`func (o *InferencePolicyRulesV2AllOfSecurity) UnsetCapabilities()`

UnsetCapabilities ensures that no value is present for Capabilities, not even an explicit nil
### GetHostIpc

`func (o *InferencePolicyRulesV2AllOfSecurity) GetHostIpc() BooleanRules`

GetHostIpc returns the HostIpc field if non-nil, zero value otherwise.

### GetHostIpcOk

`func (o *InferencePolicyRulesV2AllOfSecurity) GetHostIpcOk() (*BooleanRules, bool)`

GetHostIpcOk returns a tuple with the HostIpc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostIpc

`func (o *InferencePolicyRulesV2AllOfSecurity) SetHostIpc(v BooleanRules)`

SetHostIpc sets HostIpc field to given value.

### HasHostIpc

`func (o *InferencePolicyRulesV2AllOfSecurity) HasHostIpc() bool`

HasHostIpc returns a boolean if a field has been set.

### SetHostIpcNil

`func (o *InferencePolicyRulesV2AllOfSecurity) SetHostIpcNil(b bool)`

 SetHostIpcNil sets the value for HostIpc to be an explicit nil

### UnsetHostIpc
`func (o *InferencePolicyRulesV2AllOfSecurity) UnsetHostIpc()`

UnsetHostIpc ensures that no value is present for HostIpc, not even an explicit nil
### GetHostNetwork

`func (o *InferencePolicyRulesV2AllOfSecurity) GetHostNetwork() BooleanRules`

GetHostNetwork returns the HostNetwork field if non-nil, zero value otherwise.

### GetHostNetworkOk

`func (o *InferencePolicyRulesV2AllOfSecurity) GetHostNetworkOk() (*BooleanRules, bool)`

GetHostNetworkOk returns a tuple with the HostNetwork field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostNetwork

`func (o *InferencePolicyRulesV2AllOfSecurity) SetHostNetwork(v BooleanRules)`

SetHostNetwork sets HostNetwork field to given value.

### HasHostNetwork

`func (o *InferencePolicyRulesV2AllOfSecurity) HasHostNetwork() bool`

HasHostNetwork returns a boolean if a field has been set.

### SetHostNetworkNil

`func (o *InferencePolicyRulesV2AllOfSecurity) SetHostNetworkNil(b bool)`

 SetHostNetworkNil sets the value for HostNetwork to be an explicit nil

### UnsetHostNetwork
`func (o *InferencePolicyRulesV2AllOfSecurity) UnsetHostNetwork()`

UnsetHostNetwork ensures that no value is present for HostNetwork, not even an explicit nil
### GetReadOnlyRootFilesystem

`func (o *InferencePolicyRulesV2AllOfSecurity) GetReadOnlyRootFilesystem() BooleanRules`

GetReadOnlyRootFilesystem returns the ReadOnlyRootFilesystem field if non-nil, zero value otherwise.

### GetReadOnlyRootFilesystemOk

`func (o *InferencePolicyRulesV2AllOfSecurity) GetReadOnlyRootFilesystemOk() (*BooleanRules, bool)`

GetReadOnlyRootFilesystemOk returns a tuple with the ReadOnlyRootFilesystem field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnlyRootFilesystem

`func (o *InferencePolicyRulesV2AllOfSecurity) SetReadOnlyRootFilesystem(v BooleanRules)`

SetReadOnlyRootFilesystem sets ReadOnlyRootFilesystem field to given value.

### HasReadOnlyRootFilesystem

`func (o *InferencePolicyRulesV2AllOfSecurity) HasReadOnlyRootFilesystem() bool`

HasReadOnlyRootFilesystem returns a boolean if a field has been set.

### SetReadOnlyRootFilesystemNil

`func (o *InferencePolicyRulesV2AllOfSecurity) SetReadOnlyRootFilesystemNil(b bool)`

 SetReadOnlyRootFilesystemNil sets the value for ReadOnlyRootFilesystem to be an explicit nil

### UnsetReadOnlyRootFilesystem
`func (o *InferencePolicyRulesV2AllOfSecurity) UnsetReadOnlyRootFilesystem()`

UnsetReadOnlyRootFilesystem ensures that no value is present for ReadOnlyRootFilesystem, not even an explicit nil
### GetRunAsGid

`func (o *InferencePolicyRulesV2AllOfSecurity) GetRunAsGid() IntegerRulesOptional`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *InferencePolicyRulesV2AllOfSecurity) GetRunAsGidOk() (*IntegerRulesOptional, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *InferencePolicyRulesV2AllOfSecurity) SetRunAsGid(v IntegerRulesOptional)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *InferencePolicyRulesV2AllOfSecurity) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *InferencePolicyRulesV2AllOfSecurity) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *InferencePolicyRulesV2AllOfSecurity) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetRunAsNonRoot

`func (o *InferencePolicyRulesV2AllOfSecurity) GetRunAsNonRoot() BooleanRules`

GetRunAsNonRoot returns the RunAsNonRoot field if non-nil, zero value otherwise.

### GetRunAsNonRootOk

`func (o *InferencePolicyRulesV2AllOfSecurity) GetRunAsNonRootOk() (*BooleanRules, bool)`

GetRunAsNonRootOk returns a tuple with the RunAsNonRoot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsNonRoot

`func (o *InferencePolicyRulesV2AllOfSecurity) SetRunAsNonRoot(v BooleanRules)`

SetRunAsNonRoot sets RunAsNonRoot field to given value.

### HasRunAsNonRoot

`func (o *InferencePolicyRulesV2AllOfSecurity) HasRunAsNonRoot() bool`

HasRunAsNonRoot returns a boolean if a field has been set.

### SetRunAsNonRootNil

`func (o *InferencePolicyRulesV2AllOfSecurity) SetRunAsNonRootNil(b bool)`

 SetRunAsNonRootNil sets the value for RunAsNonRoot to be an explicit nil

### UnsetRunAsNonRoot
`func (o *InferencePolicyRulesV2AllOfSecurity) UnsetRunAsNonRoot()`

UnsetRunAsNonRoot ensures that no value is present for RunAsNonRoot, not even an explicit nil
### GetRunAsUid

`func (o *InferencePolicyRulesV2AllOfSecurity) GetRunAsUid() IntegerRulesOptional`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *InferencePolicyRulesV2AllOfSecurity) GetRunAsUidOk() (*IntegerRulesOptional, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *InferencePolicyRulesV2AllOfSecurity) SetRunAsUid(v IntegerRulesOptional)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *InferencePolicyRulesV2AllOfSecurity) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *InferencePolicyRulesV2AllOfSecurity) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *InferencePolicyRulesV2AllOfSecurity) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetSeccompProfileType

`func (o *InferencePolicyRulesV2AllOfSecurity) GetSeccompProfileType() SeccompProfileTypeRules`

GetSeccompProfileType returns the SeccompProfileType field if non-nil, zero value otherwise.

### GetSeccompProfileTypeOk

`func (o *InferencePolicyRulesV2AllOfSecurity) GetSeccompProfileTypeOk() (*SeccompProfileTypeRules, bool)`

GetSeccompProfileTypeOk returns a tuple with the SeccompProfileType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeccompProfileType

`func (o *InferencePolicyRulesV2AllOfSecurity) SetSeccompProfileType(v SeccompProfileTypeRules)`

SetSeccompProfileType sets SeccompProfileType field to given value.

### HasSeccompProfileType

`func (o *InferencePolicyRulesV2AllOfSecurity) HasSeccompProfileType() bool`

HasSeccompProfileType returns a boolean if a field has been set.

### SetSeccompProfileTypeNil

`func (o *InferencePolicyRulesV2AllOfSecurity) SetSeccompProfileTypeNil(b bool)`

 SetSeccompProfileTypeNil sets the value for SeccompProfileType to be an explicit nil

### UnsetSeccompProfileType
`func (o *InferencePolicyRulesV2AllOfSecurity) UnsetSeccompProfileType()`

UnsetSeccompProfileType ensures that no value is present for SeccompProfileType, not even an explicit nil
### GetSupplementalGroups

`func (o *InferencePolicyRulesV2AllOfSecurity) GetSupplementalGroups() StringRules`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *InferencePolicyRulesV2AllOfSecurity) GetSupplementalGroupsOk() (*StringRules, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *InferencePolicyRulesV2AllOfSecurity) SetSupplementalGroups(v StringRules)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *InferencePolicyRulesV2AllOfSecurity) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *InferencePolicyRulesV2AllOfSecurity) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *InferencePolicyRulesV2AllOfSecurity) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil
### GetUidGidSource

`func (o *InferencePolicyRulesV2AllOfSecurity) GetUidGidSource() UidGidSourceRules`

GetUidGidSource returns the UidGidSource field if non-nil, zero value otherwise.

### GetUidGidSourceOk

`func (o *InferencePolicyRulesV2AllOfSecurity) GetUidGidSourceOk() (*UidGidSourceRules, bool)`

GetUidGidSourceOk returns a tuple with the UidGidSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUidGidSource

`func (o *InferencePolicyRulesV2AllOfSecurity) SetUidGidSource(v UidGidSourceRules)`

SetUidGidSource sets UidGidSource field to given value.

### HasUidGidSource

`func (o *InferencePolicyRulesV2AllOfSecurity) HasUidGidSource() bool`

HasUidGidSource returns a boolean if a field has been set.

### SetUidGidSourceNil

`func (o *InferencePolicyRulesV2AllOfSecurity) SetUidGidSourceNil(b bool)`

 SetUidGidSourceNil sets the value for UidGidSource to be an explicit nil

### UnsetUidGidSource
`func (o *InferencePolicyRulesV2AllOfSecurity) UnsetUidGidSource()`

UnsetUidGidSource ensures that no value is present for UidGidSource, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


