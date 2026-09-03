# AgentNetworkingService

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Create** | Pointer to [**NullableAgentServiceCreate**](AgentServiceCreate.md) |  | [optional] 
**Ref** | Pointer to [**NullableServiceRef**](ServiceRef.md) |  | [optional] 
**Authorization** | Pointer to [**NullableWorkloadv2Authorization**](Workloadv2Authorization.md) |  | [optional] 

## Methods

### NewAgentNetworkingService

`func NewAgentNetworkingService() *AgentNetworkingService`

NewAgentNetworkingService instantiates a new AgentNetworkingService object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAgentNetworkingServiceWithDefaults

`func NewAgentNetworkingServiceWithDefaults() *AgentNetworkingService`

NewAgentNetworkingServiceWithDefaults instantiates a new AgentNetworkingService object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreate

`func (o *AgentNetworkingService) GetCreate() AgentServiceCreate`

GetCreate returns the Create field if non-nil, zero value otherwise.

### GetCreateOk

`func (o *AgentNetworkingService) GetCreateOk() (*AgentServiceCreate, bool)`

GetCreateOk returns a tuple with the Create field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreate

`func (o *AgentNetworkingService) SetCreate(v AgentServiceCreate)`

SetCreate sets Create field to given value.

### HasCreate

`func (o *AgentNetworkingService) HasCreate() bool`

HasCreate returns a boolean if a field has been set.

### SetCreateNil

`func (o *AgentNetworkingService) SetCreateNil(b bool)`

 SetCreateNil sets the value for Create to be an explicit nil

### UnsetCreate
`func (o *AgentNetworkingService) UnsetCreate()`

UnsetCreate ensures that no value is present for Create, not even an explicit nil
### GetRef

`func (o *AgentNetworkingService) GetRef() ServiceRef`

GetRef returns the Ref field if non-nil, zero value otherwise.

### GetRefOk

`func (o *AgentNetworkingService) GetRefOk() (*ServiceRef, bool)`

GetRefOk returns a tuple with the Ref field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRef

`func (o *AgentNetworkingService) SetRef(v ServiceRef)`

SetRef sets Ref field to given value.

### HasRef

`func (o *AgentNetworkingService) HasRef() bool`

HasRef returns a boolean if a field has been set.

### SetRefNil

`func (o *AgentNetworkingService) SetRefNil(b bool)`

 SetRefNil sets the value for Ref to be an explicit nil

### UnsetRef
`func (o *AgentNetworkingService) UnsetRef()`

UnsetRef ensures that no value is present for Ref, not even an explicit nil
### GetAuthorization

`func (o *AgentNetworkingService) GetAuthorization() Workloadv2Authorization`

GetAuthorization returns the Authorization field if non-nil, zero value otherwise.

### GetAuthorizationOk

`func (o *AgentNetworkingService) GetAuthorizationOk() (*Workloadv2Authorization, bool)`

GetAuthorizationOk returns a tuple with the Authorization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorization

`func (o *AgentNetworkingService) SetAuthorization(v Workloadv2Authorization)`

SetAuthorization sets Authorization field to given value.

### HasAuthorization

`func (o *AgentNetworkingService) HasAuthorization() bool`

HasAuthorization returns a boolean if a field has been set.

### SetAuthorizationNil

`func (o *AgentNetworkingService) SetAuthorizationNil(b bool)`

 SetAuthorizationNil sets the value for Authorization to be an explicit nil

### UnsetAuthorization
`func (o *AgentNetworkingService) UnsetAuthorization()`

UnsetAuthorization ensures that no value is present for Authorization, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


