# SecurityNonOverridableRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowPrivilegeEscalation** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**UidGidSource** | Pointer to [**NullableUidGidSourceRules**](UidGidSourceRules.md) |  | [optional] 
**Capabilities** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**SeccompProfileType** | Pointer to [**NullableSeccompProfileTypeRules**](SeccompProfileTypeRules.md) |  | [optional] 
**RunAsNonRoot** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**HostIpc** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**HostNetwork** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 

## Methods

### NewSecurityNonOverridableRules

`func NewSecurityNonOverridableRules() *SecurityNonOverridableRules`

NewSecurityNonOverridableRules instantiates a new SecurityNonOverridableRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecurityNonOverridableRulesWithDefaults

`func NewSecurityNonOverridableRulesWithDefaults() *SecurityNonOverridableRules`

NewSecurityNonOverridableRulesWithDefaults instantiates a new SecurityNonOverridableRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllowPrivilegeEscalation

`func (o *SecurityNonOverridableRules) GetAllowPrivilegeEscalation() BooleanRules`

GetAllowPrivilegeEscalation returns the AllowPrivilegeEscalation field if non-nil, zero value otherwise.

### GetAllowPrivilegeEscalationOk

`func (o *SecurityNonOverridableRules) GetAllowPrivilegeEscalationOk() (*BooleanRules, bool)`

GetAllowPrivilegeEscalationOk returns a tuple with the AllowPrivilegeEscalation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowPrivilegeEscalation

`func (o *SecurityNonOverridableRules) SetAllowPrivilegeEscalation(v BooleanRules)`

SetAllowPrivilegeEscalation sets AllowPrivilegeEscalation field to given value.

### HasAllowPrivilegeEscalation

`func (o *SecurityNonOverridableRules) HasAllowPrivilegeEscalation() bool`

HasAllowPrivilegeEscalation returns a boolean if a field has been set.

### SetAllowPrivilegeEscalationNil

`func (o *SecurityNonOverridableRules) SetAllowPrivilegeEscalationNil(b bool)`

 SetAllowPrivilegeEscalationNil sets the value for AllowPrivilegeEscalation to be an explicit nil

### UnsetAllowPrivilegeEscalation
`func (o *SecurityNonOverridableRules) UnsetAllowPrivilegeEscalation()`

UnsetAllowPrivilegeEscalation ensures that no value is present for AllowPrivilegeEscalation, not even an explicit nil
### GetUidGidSource

`func (o *SecurityNonOverridableRules) GetUidGidSource() UidGidSourceRules`

GetUidGidSource returns the UidGidSource field if non-nil, zero value otherwise.

### GetUidGidSourceOk

`func (o *SecurityNonOverridableRules) GetUidGidSourceOk() (*UidGidSourceRules, bool)`

GetUidGidSourceOk returns a tuple with the UidGidSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUidGidSource

`func (o *SecurityNonOverridableRules) SetUidGidSource(v UidGidSourceRules)`

SetUidGidSource sets UidGidSource field to given value.

### HasUidGidSource

`func (o *SecurityNonOverridableRules) HasUidGidSource() bool`

HasUidGidSource returns a boolean if a field has been set.

### SetUidGidSourceNil

`func (o *SecurityNonOverridableRules) SetUidGidSourceNil(b bool)`

 SetUidGidSourceNil sets the value for UidGidSource to be an explicit nil

### UnsetUidGidSource
`func (o *SecurityNonOverridableRules) UnsetUidGidSource()`

UnsetUidGidSource ensures that no value is present for UidGidSource, not even an explicit nil
### GetCapabilities

`func (o *SecurityNonOverridableRules) GetCapabilities() ArrayRules`

GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.

### GetCapabilitiesOk

`func (o *SecurityNonOverridableRules) GetCapabilitiesOk() (*ArrayRules, bool)`

GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapabilities

`func (o *SecurityNonOverridableRules) SetCapabilities(v ArrayRules)`

SetCapabilities sets Capabilities field to given value.

### HasCapabilities

`func (o *SecurityNonOverridableRules) HasCapabilities() bool`

HasCapabilities returns a boolean if a field has been set.

### SetCapabilitiesNil

`func (o *SecurityNonOverridableRules) SetCapabilitiesNil(b bool)`

 SetCapabilitiesNil sets the value for Capabilities to be an explicit nil

### UnsetCapabilities
`func (o *SecurityNonOverridableRules) UnsetCapabilities()`

UnsetCapabilities ensures that no value is present for Capabilities, not even an explicit nil
### GetSeccompProfileType

`func (o *SecurityNonOverridableRules) GetSeccompProfileType() SeccompProfileTypeRules`

GetSeccompProfileType returns the SeccompProfileType field if non-nil, zero value otherwise.

### GetSeccompProfileTypeOk

`func (o *SecurityNonOverridableRules) GetSeccompProfileTypeOk() (*SeccompProfileTypeRules, bool)`

GetSeccompProfileTypeOk returns a tuple with the SeccompProfileType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeccompProfileType

`func (o *SecurityNonOverridableRules) SetSeccompProfileType(v SeccompProfileTypeRules)`

SetSeccompProfileType sets SeccompProfileType field to given value.

### HasSeccompProfileType

`func (o *SecurityNonOverridableRules) HasSeccompProfileType() bool`

HasSeccompProfileType returns a boolean if a field has been set.

### SetSeccompProfileTypeNil

`func (o *SecurityNonOverridableRules) SetSeccompProfileTypeNil(b bool)`

 SetSeccompProfileTypeNil sets the value for SeccompProfileType to be an explicit nil

### UnsetSeccompProfileType
`func (o *SecurityNonOverridableRules) UnsetSeccompProfileType()`

UnsetSeccompProfileType ensures that no value is present for SeccompProfileType, not even an explicit nil
### GetRunAsNonRoot

`func (o *SecurityNonOverridableRules) GetRunAsNonRoot() BooleanRules`

GetRunAsNonRoot returns the RunAsNonRoot field if non-nil, zero value otherwise.

### GetRunAsNonRootOk

`func (o *SecurityNonOverridableRules) GetRunAsNonRootOk() (*BooleanRules, bool)`

GetRunAsNonRootOk returns a tuple with the RunAsNonRoot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsNonRoot

`func (o *SecurityNonOverridableRules) SetRunAsNonRoot(v BooleanRules)`

SetRunAsNonRoot sets RunAsNonRoot field to given value.

### HasRunAsNonRoot

`func (o *SecurityNonOverridableRules) HasRunAsNonRoot() bool`

HasRunAsNonRoot returns a boolean if a field has been set.

### SetRunAsNonRootNil

`func (o *SecurityNonOverridableRules) SetRunAsNonRootNil(b bool)`

 SetRunAsNonRootNil sets the value for RunAsNonRoot to be an explicit nil

### UnsetRunAsNonRoot
`func (o *SecurityNonOverridableRules) UnsetRunAsNonRoot()`

UnsetRunAsNonRoot ensures that no value is present for RunAsNonRoot, not even an explicit nil
### GetHostIpc

`func (o *SecurityNonOverridableRules) GetHostIpc() BooleanRules`

GetHostIpc returns the HostIpc field if non-nil, zero value otherwise.

### GetHostIpcOk

`func (o *SecurityNonOverridableRules) GetHostIpcOk() (*BooleanRules, bool)`

GetHostIpcOk returns a tuple with the HostIpc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostIpc

`func (o *SecurityNonOverridableRules) SetHostIpc(v BooleanRules)`

SetHostIpc sets HostIpc field to given value.

### HasHostIpc

`func (o *SecurityNonOverridableRules) HasHostIpc() bool`

HasHostIpc returns a boolean if a field has been set.

### SetHostIpcNil

`func (o *SecurityNonOverridableRules) SetHostIpcNil(b bool)`

 SetHostIpcNil sets the value for HostIpc to be an explicit nil

### UnsetHostIpc
`func (o *SecurityNonOverridableRules) UnsetHostIpc()`

UnsetHostIpc ensures that no value is present for HostIpc, not even an explicit nil
### GetHostNetwork

`func (o *SecurityNonOverridableRules) GetHostNetwork() BooleanRules`

GetHostNetwork returns the HostNetwork field if non-nil, zero value otherwise.

### GetHostNetworkOk

`func (o *SecurityNonOverridableRules) GetHostNetworkOk() (*BooleanRules, bool)`

GetHostNetworkOk returns a tuple with the HostNetwork field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostNetwork

`func (o *SecurityNonOverridableRules) SetHostNetwork(v BooleanRules)`

SetHostNetwork sets HostNetwork field to given value.

### HasHostNetwork

`func (o *SecurityNonOverridableRules) HasHostNetwork() bool`

HasHostNetwork returns a boolean if a field has been set.

### SetHostNetworkNil

`func (o *SecurityNonOverridableRules) SetHostNetworkNil(b bool)`

 SetHostNetworkNil sets the value for HostNetwork to be an explicit nil

### UnsetHostNetwork
`func (o *SecurityNonOverridableRules) UnsetHostNetwork()`

UnsetHostNetwork ensures that no value is present for HostNetwork, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


