# NonInferenceSecurityFlatFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowPrivilegeEscalation** | Pointer to **NullableBool** | Allow the container running the workload and all launched processes to gain additional privileges after the workload starts. For more information consult the User Identity in Container guide at https://docs.run.ai/admin/runai-setup/config/non-root-containers/ | [optional] 
**HostIpc** | Pointer to **NullableBool** | Whether to enable host IPC. Defaults to false. | [optional] 
**HostNetwork** | Pointer to **NullableBool** | Whether to enable host networking. Default to false. | [optional] 

## Methods

### NewNonInferenceSecurityFlatFields

`func NewNonInferenceSecurityFlatFields() *NonInferenceSecurityFlatFields`

NewNonInferenceSecurityFlatFields instantiates a new NonInferenceSecurityFlatFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNonInferenceSecurityFlatFieldsWithDefaults

`func NewNonInferenceSecurityFlatFieldsWithDefaults() *NonInferenceSecurityFlatFields`

NewNonInferenceSecurityFlatFieldsWithDefaults instantiates a new NonInferenceSecurityFlatFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllowPrivilegeEscalation

`func (o *NonInferenceSecurityFlatFields) GetAllowPrivilegeEscalation() bool`

GetAllowPrivilegeEscalation returns the AllowPrivilegeEscalation field if non-nil, zero value otherwise.

### GetAllowPrivilegeEscalationOk

`func (o *NonInferenceSecurityFlatFields) GetAllowPrivilegeEscalationOk() (*bool, bool)`

GetAllowPrivilegeEscalationOk returns a tuple with the AllowPrivilegeEscalation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowPrivilegeEscalation

`func (o *NonInferenceSecurityFlatFields) SetAllowPrivilegeEscalation(v bool)`

SetAllowPrivilegeEscalation sets AllowPrivilegeEscalation field to given value.

### HasAllowPrivilegeEscalation

`func (o *NonInferenceSecurityFlatFields) HasAllowPrivilegeEscalation() bool`

HasAllowPrivilegeEscalation returns a boolean if a field has been set.

### SetAllowPrivilegeEscalationNil

`func (o *NonInferenceSecurityFlatFields) SetAllowPrivilegeEscalationNil(b bool)`

 SetAllowPrivilegeEscalationNil sets the value for AllowPrivilegeEscalation to be an explicit nil

### UnsetAllowPrivilegeEscalation
`func (o *NonInferenceSecurityFlatFields) UnsetAllowPrivilegeEscalation()`

UnsetAllowPrivilegeEscalation ensures that no value is present for AllowPrivilegeEscalation, not even an explicit nil
### GetHostIpc

`func (o *NonInferenceSecurityFlatFields) GetHostIpc() bool`

GetHostIpc returns the HostIpc field if non-nil, zero value otherwise.

### GetHostIpcOk

`func (o *NonInferenceSecurityFlatFields) GetHostIpcOk() (*bool, bool)`

GetHostIpcOk returns a tuple with the HostIpc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostIpc

`func (o *NonInferenceSecurityFlatFields) SetHostIpc(v bool)`

SetHostIpc sets HostIpc field to given value.

### HasHostIpc

`func (o *NonInferenceSecurityFlatFields) HasHostIpc() bool`

HasHostIpc returns a boolean if a field has been set.

### SetHostIpcNil

`func (o *NonInferenceSecurityFlatFields) SetHostIpcNil(b bool)`

 SetHostIpcNil sets the value for HostIpc to be an explicit nil

### UnsetHostIpc
`func (o *NonInferenceSecurityFlatFields) UnsetHostIpc()`

UnsetHostIpc ensures that no value is present for HostIpc, not even an explicit nil
### GetHostNetwork

`func (o *NonInferenceSecurityFlatFields) GetHostNetwork() bool`

GetHostNetwork returns the HostNetwork field if non-nil, zero value otherwise.

### GetHostNetworkOk

`func (o *NonInferenceSecurityFlatFields) GetHostNetworkOk() (*bool, bool)`

GetHostNetworkOk returns a tuple with the HostNetwork field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostNetwork

`func (o *NonInferenceSecurityFlatFields) SetHostNetwork(v bool)`

SetHostNetwork sets HostNetwork field to given value.

### HasHostNetwork

`func (o *NonInferenceSecurityFlatFields) HasHostNetwork() bool`

HasHostNetwork returns a boolean if a field has been set.

### SetHostNetworkNil

`func (o *NonInferenceSecurityFlatFields) SetHostNetworkNil(b bool)`

 SetHostNetworkNil sets the value for HostNetwork to be an explicit nil

### UnsetHostNetwork
`func (o *NonInferenceSecurityFlatFields) UnsetHostNetwork()`

UnsetHostNetwork ensures that no value is present for HostNetwork, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


