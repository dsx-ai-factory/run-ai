# WorkloadV2UpdateConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InferenceEndpoint** | Pointer to [**NullableAgentInferenceEndpoint**](AgentInferenceEndpoint.md) |  | [optional] 
**Mnnvl** | Pointer to [**WorkloadV2UpdateConfigurationMnnvl**](WorkloadV2UpdateConfigurationMnnvl.md) |  | [optional] 
**Networking** | Pointer to [**NullableAgentNetworkingConfiguration**](AgentNetworkingConfiguration.md) |  | [optional] 
**NetworkingEndpoints** | Pointer to [**[]NetworkingEndpoint**](NetworkingEndpoint.md) | Endpoint-array networking configuration sent to agents that support the new networking shape. Supported for cluster version 2.26.0 and above.  | [optional] 
**NetworkingSubmissionProfile** | Pointer to [**NetworkingSubmissionProfile**](NetworkingSubmissionProfile.md) |  | [optional] 
**UserCredentials** | Pointer to [**NullableWorkloadV2UpdateConfigurationUserCredentials**](WorkloadV2UpdateConfigurationUserCredentials.md) |  | [optional] 
**Pvcs** | Pointer to [**[]WorkloadV2UpdateConfigurationPvc**](WorkloadV2UpdateConfigurationPvc.md) | PVCs to create and reconcile alongside the workload. | [optional] 
**PvcsConfig** | Pointer to [**NullableWorkloadV2UpdateConfigurationPvcsConfig**](WorkloadV2UpdateConfigurationPvcsConfig.md) |  | [optional] 

## Methods

### NewWorkloadV2UpdateConfiguration

`func NewWorkloadV2UpdateConfiguration() *WorkloadV2UpdateConfiguration`

NewWorkloadV2UpdateConfiguration instantiates a new WorkloadV2UpdateConfiguration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2UpdateConfigurationWithDefaults

`func NewWorkloadV2UpdateConfigurationWithDefaults() *WorkloadV2UpdateConfiguration`

NewWorkloadV2UpdateConfigurationWithDefaults instantiates a new WorkloadV2UpdateConfiguration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInferenceEndpoint

`func (o *WorkloadV2UpdateConfiguration) GetInferenceEndpoint() AgentInferenceEndpoint`

GetInferenceEndpoint returns the InferenceEndpoint field if non-nil, zero value otherwise.

### GetInferenceEndpointOk

`func (o *WorkloadV2UpdateConfiguration) GetInferenceEndpointOk() (*AgentInferenceEndpoint, bool)`

GetInferenceEndpointOk returns a tuple with the InferenceEndpoint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInferenceEndpoint

`func (o *WorkloadV2UpdateConfiguration) SetInferenceEndpoint(v AgentInferenceEndpoint)`

SetInferenceEndpoint sets InferenceEndpoint field to given value.

### HasInferenceEndpoint

`func (o *WorkloadV2UpdateConfiguration) HasInferenceEndpoint() bool`

HasInferenceEndpoint returns a boolean if a field has been set.

### SetInferenceEndpointNil

`func (o *WorkloadV2UpdateConfiguration) SetInferenceEndpointNil(b bool)`

 SetInferenceEndpointNil sets the value for InferenceEndpoint to be an explicit nil

### UnsetInferenceEndpoint
`func (o *WorkloadV2UpdateConfiguration) UnsetInferenceEndpoint()`

UnsetInferenceEndpoint ensures that no value is present for InferenceEndpoint, not even an explicit nil
### GetMnnvl

`func (o *WorkloadV2UpdateConfiguration) GetMnnvl() WorkloadV2UpdateConfigurationMnnvl`

GetMnnvl returns the Mnnvl field if non-nil, zero value otherwise.

### GetMnnvlOk

`func (o *WorkloadV2UpdateConfiguration) GetMnnvlOk() (*WorkloadV2UpdateConfigurationMnnvl, bool)`

GetMnnvlOk returns a tuple with the Mnnvl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMnnvl

`func (o *WorkloadV2UpdateConfiguration) SetMnnvl(v WorkloadV2UpdateConfigurationMnnvl)`

SetMnnvl sets Mnnvl field to given value.

### HasMnnvl

`func (o *WorkloadV2UpdateConfiguration) HasMnnvl() bool`

HasMnnvl returns a boolean if a field has been set.

### GetNetworking

`func (o *WorkloadV2UpdateConfiguration) GetNetworking() AgentNetworkingConfiguration`

GetNetworking returns the Networking field if non-nil, zero value otherwise.

### GetNetworkingOk

`func (o *WorkloadV2UpdateConfiguration) GetNetworkingOk() (*AgentNetworkingConfiguration, bool)`

GetNetworkingOk returns a tuple with the Networking field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworking

`func (o *WorkloadV2UpdateConfiguration) SetNetworking(v AgentNetworkingConfiguration)`

SetNetworking sets Networking field to given value.

### HasNetworking

`func (o *WorkloadV2UpdateConfiguration) HasNetworking() bool`

HasNetworking returns a boolean if a field has been set.

### SetNetworkingNil

`func (o *WorkloadV2UpdateConfiguration) SetNetworkingNil(b bool)`

 SetNetworkingNil sets the value for Networking to be an explicit nil

### UnsetNetworking
`func (o *WorkloadV2UpdateConfiguration) UnsetNetworking()`

UnsetNetworking ensures that no value is present for Networking, not even an explicit nil
### GetNetworkingEndpoints

`func (o *WorkloadV2UpdateConfiguration) GetNetworkingEndpoints() []NetworkingEndpoint`

GetNetworkingEndpoints returns the NetworkingEndpoints field if non-nil, zero value otherwise.

### GetNetworkingEndpointsOk

`func (o *WorkloadV2UpdateConfiguration) GetNetworkingEndpointsOk() (*[]NetworkingEndpoint, bool)`

GetNetworkingEndpointsOk returns a tuple with the NetworkingEndpoints field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkingEndpoints

`func (o *WorkloadV2UpdateConfiguration) SetNetworkingEndpoints(v []NetworkingEndpoint)`

SetNetworkingEndpoints sets NetworkingEndpoints field to given value.

### HasNetworkingEndpoints

`func (o *WorkloadV2UpdateConfiguration) HasNetworkingEndpoints() bool`

HasNetworkingEndpoints returns a boolean if a field has been set.

### SetNetworkingEndpointsNil

`func (o *WorkloadV2UpdateConfiguration) SetNetworkingEndpointsNil(b bool)`

 SetNetworkingEndpointsNil sets the value for NetworkingEndpoints to be an explicit nil

### UnsetNetworkingEndpoints
`func (o *WorkloadV2UpdateConfiguration) UnsetNetworkingEndpoints()`

UnsetNetworkingEndpoints ensures that no value is present for NetworkingEndpoints, not even an explicit nil
### GetNetworkingSubmissionProfile

`func (o *WorkloadV2UpdateConfiguration) GetNetworkingSubmissionProfile() NetworkingSubmissionProfile`

GetNetworkingSubmissionProfile returns the NetworkingSubmissionProfile field if non-nil, zero value otherwise.

### GetNetworkingSubmissionProfileOk

`func (o *WorkloadV2UpdateConfiguration) GetNetworkingSubmissionProfileOk() (*NetworkingSubmissionProfile, bool)`

GetNetworkingSubmissionProfileOk returns a tuple with the NetworkingSubmissionProfile field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkingSubmissionProfile

`func (o *WorkloadV2UpdateConfiguration) SetNetworkingSubmissionProfile(v NetworkingSubmissionProfile)`

SetNetworkingSubmissionProfile sets NetworkingSubmissionProfile field to given value.

### HasNetworkingSubmissionProfile

`func (o *WorkloadV2UpdateConfiguration) HasNetworkingSubmissionProfile() bool`

HasNetworkingSubmissionProfile returns a boolean if a field has been set.

### GetUserCredentials

`func (o *WorkloadV2UpdateConfiguration) GetUserCredentials() WorkloadV2UpdateConfigurationUserCredentials`

GetUserCredentials returns the UserCredentials field if non-nil, zero value otherwise.

### GetUserCredentialsOk

`func (o *WorkloadV2UpdateConfiguration) GetUserCredentialsOk() (*WorkloadV2UpdateConfigurationUserCredentials, bool)`

GetUserCredentialsOk returns a tuple with the UserCredentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserCredentials

`func (o *WorkloadV2UpdateConfiguration) SetUserCredentials(v WorkloadV2UpdateConfigurationUserCredentials)`

SetUserCredentials sets UserCredentials field to given value.

### HasUserCredentials

`func (o *WorkloadV2UpdateConfiguration) HasUserCredentials() bool`

HasUserCredentials returns a boolean if a field has been set.

### SetUserCredentialsNil

`func (o *WorkloadV2UpdateConfiguration) SetUserCredentialsNil(b bool)`

 SetUserCredentialsNil sets the value for UserCredentials to be an explicit nil

### UnsetUserCredentials
`func (o *WorkloadV2UpdateConfiguration) UnsetUserCredentials()`

UnsetUserCredentials ensures that no value is present for UserCredentials, not even an explicit nil
### GetPvcs

`func (o *WorkloadV2UpdateConfiguration) GetPvcs() []WorkloadV2UpdateConfigurationPvc`

GetPvcs returns the Pvcs field if non-nil, zero value otherwise.

### GetPvcsOk

`func (o *WorkloadV2UpdateConfiguration) GetPvcsOk() (*[]WorkloadV2UpdateConfigurationPvc, bool)`

GetPvcsOk returns a tuple with the Pvcs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvcs

`func (o *WorkloadV2UpdateConfiguration) SetPvcs(v []WorkloadV2UpdateConfigurationPvc)`

SetPvcs sets Pvcs field to given value.

### HasPvcs

`func (o *WorkloadV2UpdateConfiguration) HasPvcs() bool`

HasPvcs returns a boolean if a field has been set.

### SetPvcsNil

`func (o *WorkloadV2UpdateConfiguration) SetPvcsNil(b bool)`

 SetPvcsNil sets the value for Pvcs to be an explicit nil

### UnsetPvcs
`func (o *WorkloadV2UpdateConfiguration) UnsetPvcs()`

UnsetPvcs ensures that no value is present for Pvcs, not even an explicit nil
### GetPvcsConfig

`func (o *WorkloadV2UpdateConfiguration) GetPvcsConfig() WorkloadV2UpdateConfigurationPvcsConfig`

GetPvcsConfig returns the PvcsConfig field if non-nil, zero value otherwise.

### GetPvcsConfigOk

`func (o *WorkloadV2UpdateConfiguration) GetPvcsConfigOk() (*WorkloadV2UpdateConfigurationPvcsConfig, bool)`

GetPvcsConfigOk returns a tuple with the PvcsConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvcsConfig

`func (o *WorkloadV2UpdateConfiguration) SetPvcsConfig(v WorkloadV2UpdateConfigurationPvcsConfig)`

SetPvcsConfig sets PvcsConfig field to given value.

### HasPvcsConfig

`func (o *WorkloadV2UpdateConfiguration) HasPvcsConfig() bool`

HasPvcsConfig returns a boolean if a field has been set.

### SetPvcsConfigNil

`func (o *WorkloadV2UpdateConfiguration) SetPvcsConfigNil(b bool)`

 SetPvcsConfigNil sets the value for PvcsConfig to be an explicit nil

### UnsetPvcsConfig
`func (o *WorkloadV2UpdateConfiguration) UnsetPvcsConfig()`

UnsetPvcsConfig ensures that no value is present for PvcsConfig, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


