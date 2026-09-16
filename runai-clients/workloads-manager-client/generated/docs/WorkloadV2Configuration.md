# WorkloadV2Configuration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Mnnvl** | Pointer to [**NullableMnnvlMode**](MnnvlMode.md) |  | [optional] 
**Networking** | Pointer to [**[]NetworkingEndpoint**](NetworkingEndpoint.md) | Networking endpoint configurations for the workload (external access, services with auth). Direct submissions require cluster version 3.0.0 or above for networking ownership validation.  | [optional] 
**UserCredentials** | Pointer to [**NullableWorkloadV2UpdateConfigurationUserCredentials**](WorkloadV2UpdateConfigurationUserCredentials.md) |  | [optional] 
**Pvcs** | Pointer to [**[]WorkloadV2UpdateConfigurationPvc**](WorkloadV2UpdateConfigurationPvc.md) | PVCs to create and reconcile alongside the workload. | [optional] 
**PvcsConfig** | Pointer to [**NullableWorkloadV2UpdateConfigurationPvcsConfig**](WorkloadV2UpdateConfigurationPvcsConfig.md) |  | [optional] 

## Methods

### NewWorkloadV2Configuration

`func NewWorkloadV2Configuration() *WorkloadV2Configuration`

NewWorkloadV2Configuration instantiates a new WorkloadV2Configuration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2ConfigurationWithDefaults

`func NewWorkloadV2ConfigurationWithDefaults() *WorkloadV2Configuration`

NewWorkloadV2ConfigurationWithDefaults instantiates a new WorkloadV2Configuration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMnnvl

`func (o *WorkloadV2Configuration) GetMnnvl() MnnvlMode`

GetMnnvl returns the Mnnvl field if non-nil, zero value otherwise.

### GetMnnvlOk

`func (o *WorkloadV2Configuration) GetMnnvlOk() (*MnnvlMode, bool)`

GetMnnvlOk returns a tuple with the Mnnvl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMnnvl

`func (o *WorkloadV2Configuration) SetMnnvl(v MnnvlMode)`

SetMnnvl sets Mnnvl field to given value.

### HasMnnvl

`func (o *WorkloadV2Configuration) HasMnnvl() bool`

HasMnnvl returns a boolean if a field has been set.

### SetMnnvlNil

`func (o *WorkloadV2Configuration) SetMnnvlNil(b bool)`

 SetMnnvlNil sets the value for Mnnvl to be an explicit nil

### UnsetMnnvl
`func (o *WorkloadV2Configuration) UnsetMnnvl()`

UnsetMnnvl ensures that no value is present for Mnnvl, not even an explicit nil
### GetNetworking

`func (o *WorkloadV2Configuration) GetNetworking() []NetworkingEndpoint`

GetNetworking returns the Networking field if non-nil, zero value otherwise.

### GetNetworkingOk

`func (o *WorkloadV2Configuration) GetNetworkingOk() (*[]NetworkingEndpoint, bool)`

GetNetworkingOk returns a tuple with the Networking field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworking

`func (o *WorkloadV2Configuration) SetNetworking(v []NetworkingEndpoint)`

SetNetworking sets Networking field to given value.

### HasNetworking

`func (o *WorkloadV2Configuration) HasNetworking() bool`

HasNetworking returns a boolean if a field has been set.

### SetNetworkingNil

`func (o *WorkloadV2Configuration) SetNetworkingNil(b bool)`

 SetNetworkingNil sets the value for Networking to be an explicit nil

### UnsetNetworking
`func (o *WorkloadV2Configuration) UnsetNetworking()`

UnsetNetworking ensures that no value is present for Networking, not even an explicit nil
### GetUserCredentials

`func (o *WorkloadV2Configuration) GetUserCredentials() WorkloadV2UpdateConfigurationUserCredentials`

GetUserCredentials returns the UserCredentials field if non-nil, zero value otherwise.

### GetUserCredentialsOk

`func (o *WorkloadV2Configuration) GetUserCredentialsOk() (*WorkloadV2UpdateConfigurationUserCredentials, bool)`

GetUserCredentialsOk returns a tuple with the UserCredentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserCredentials

`func (o *WorkloadV2Configuration) SetUserCredentials(v WorkloadV2UpdateConfigurationUserCredentials)`

SetUserCredentials sets UserCredentials field to given value.

### HasUserCredentials

`func (o *WorkloadV2Configuration) HasUserCredentials() bool`

HasUserCredentials returns a boolean if a field has been set.

### SetUserCredentialsNil

`func (o *WorkloadV2Configuration) SetUserCredentialsNil(b bool)`

 SetUserCredentialsNil sets the value for UserCredentials to be an explicit nil

### UnsetUserCredentials
`func (o *WorkloadV2Configuration) UnsetUserCredentials()`

UnsetUserCredentials ensures that no value is present for UserCredentials, not even an explicit nil
### GetPvcs

`func (o *WorkloadV2Configuration) GetPvcs() []WorkloadV2UpdateConfigurationPvc`

GetPvcs returns the Pvcs field if non-nil, zero value otherwise.

### GetPvcsOk

`func (o *WorkloadV2Configuration) GetPvcsOk() (*[]WorkloadV2UpdateConfigurationPvc, bool)`

GetPvcsOk returns a tuple with the Pvcs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvcs

`func (o *WorkloadV2Configuration) SetPvcs(v []WorkloadV2UpdateConfigurationPvc)`

SetPvcs sets Pvcs field to given value.

### HasPvcs

`func (o *WorkloadV2Configuration) HasPvcs() bool`

HasPvcs returns a boolean if a field has been set.

### SetPvcsNil

`func (o *WorkloadV2Configuration) SetPvcsNil(b bool)`

 SetPvcsNil sets the value for Pvcs to be an explicit nil

### UnsetPvcs
`func (o *WorkloadV2Configuration) UnsetPvcs()`

UnsetPvcs ensures that no value is present for Pvcs, not even an explicit nil
### GetPvcsConfig

`func (o *WorkloadV2Configuration) GetPvcsConfig() WorkloadV2UpdateConfigurationPvcsConfig`

GetPvcsConfig returns the PvcsConfig field if non-nil, zero value otherwise.

### GetPvcsConfigOk

`func (o *WorkloadV2Configuration) GetPvcsConfigOk() (*WorkloadV2UpdateConfigurationPvcsConfig, bool)`

GetPvcsConfigOk returns a tuple with the PvcsConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvcsConfig

`func (o *WorkloadV2Configuration) SetPvcsConfig(v WorkloadV2UpdateConfigurationPvcsConfig)`

SetPvcsConfig sets PvcsConfig field to given value.

### HasPvcsConfig

`func (o *WorkloadV2Configuration) HasPvcsConfig() bool`

HasPvcsConfig returns a boolean if a field has been set.

### SetPvcsConfigNil

`func (o *WorkloadV2Configuration) SetPvcsConfigNil(b bool)`

 SetPvcsConfigNil sets the value for PvcsConfig to be an explicit nil

### UnsetPvcsConfig
`func (o *WorkloadV2Configuration) UnsetPvcsConfig()`

UnsetPvcsConfig ensures that no value is present for PvcsConfig, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


