# AgentNetworkingConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExternalAccess** | Pointer to [**NullableExternalAccess**](ExternalAccess.md) |  | [optional] 
**Services** | Pointer to [**[]AgentNetworkingService**](AgentNetworkingService.md) | Kubernetes service to create or reference for this endpoint (maximum one). | [optional] 

## Methods

### NewAgentNetworkingConfiguration

`func NewAgentNetworkingConfiguration() *AgentNetworkingConfiguration`

NewAgentNetworkingConfiguration instantiates a new AgentNetworkingConfiguration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAgentNetworkingConfigurationWithDefaults

`func NewAgentNetworkingConfigurationWithDefaults() *AgentNetworkingConfiguration`

NewAgentNetworkingConfigurationWithDefaults instantiates a new AgentNetworkingConfiguration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExternalAccess

`func (o *AgentNetworkingConfiguration) GetExternalAccess() ExternalAccess`

GetExternalAccess returns the ExternalAccess field if non-nil, zero value otherwise.

### GetExternalAccessOk

`func (o *AgentNetworkingConfiguration) GetExternalAccessOk() (*ExternalAccess, bool)`

GetExternalAccessOk returns a tuple with the ExternalAccess field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalAccess

`func (o *AgentNetworkingConfiguration) SetExternalAccess(v ExternalAccess)`

SetExternalAccess sets ExternalAccess field to given value.

### HasExternalAccess

`func (o *AgentNetworkingConfiguration) HasExternalAccess() bool`

HasExternalAccess returns a boolean if a field has been set.

### SetExternalAccessNil

`func (o *AgentNetworkingConfiguration) SetExternalAccessNil(b bool)`

 SetExternalAccessNil sets the value for ExternalAccess to be an explicit nil

### UnsetExternalAccess
`func (o *AgentNetworkingConfiguration) UnsetExternalAccess()`

UnsetExternalAccess ensures that no value is present for ExternalAccess, not even an explicit nil
### GetServices

`func (o *AgentNetworkingConfiguration) GetServices() []AgentNetworkingService`

GetServices returns the Services field if non-nil, zero value otherwise.

### GetServicesOk

`func (o *AgentNetworkingConfiguration) GetServicesOk() (*[]AgentNetworkingService, bool)`

GetServicesOk returns a tuple with the Services field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServices

`func (o *AgentNetworkingConfiguration) SetServices(v []AgentNetworkingService)`

SetServices sets Services field to given value.

### HasServices

`func (o *AgentNetworkingConfiguration) HasServices() bool`

HasServices returns a boolean if a field has been set.

### SetServicesNil

`func (o *AgentNetworkingConfiguration) SetServicesNil(b bool)`

 SetServicesNil sets the value for Services to be an explicit nil

### UnsetServices
`func (o *AgentNetworkingConfiguration) UnsetServices()`

UnsetServices ensures that no value is present for Services, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


