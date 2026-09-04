# NetworkingEndpoint

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExternalAccess** | Pointer to [**NullableExternalAccess**](ExternalAccess.md) |  | [optional] 
**Services** | Pointer to [**[]NetworkingService**](NetworkingService.md) | Kubernetes service to create or reference for this endpoint (maximum one). | [optional] 

## Methods

### NewNetworkingEndpoint

`func NewNetworkingEndpoint() *NetworkingEndpoint`

NewNetworkingEndpoint instantiates a new NetworkingEndpoint object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNetworkingEndpointWithDefaults

`func NewNetworkingEndpointWithDefaults() *NetworkingEndpoint`

NewNetworkingEndpointWithDefaults instantiates a new NetworkingEndpoint object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExternalAccess

`func (o *NetworkingEndpoint) GetExternalAccess() ExternalAccess`

GetExternalAccess returns the ExternalAccess field if non-nil, zero value otherwise.

### GetExternalAccessOk

`func (o *NetworkingEndpoint) GetExternalAccessOk() (*ExternalAccess, bool)`

GetExternalAccessOk returns a tuple with the ExternalAccess field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalAccess

`func (o *NetworkingEndpoint) SetExternalAccess(v ExternalAccess)`

SetExternalAccess sets ExternalAccess field to given value.

### HasExternalAccess

`func (o *NetworkingEndpoint) HasExternalAccess() bool`

HasExternalAccess returns a boolean if a field has been set.

### SetExternalAccessNil

`func (o *NetworkingEndpoint) SetExternalAccessNil(b bool)`

 SetExternalAccessNil sets the value for ExternalAccess to be an explicit nil

### UnsetExternalAccess
`func (o *NetworkingEndpoint) UnsetExternalAccess()`

UnsetExternalAccess ensures that no value is present for ExternalAccess, not even an explicit nil
### GetServices

`func (o *NetworkingEndpoint) GetServices() []NetworkingService`

GetServices returns the Services field if non-nil, zero value otherwise.

### GetServicesOk

`func (o *NetworkingEndpoint) GetServicesOk() (*[]NetworkingService, bool)`

GetServicesOk returns a tuple with the Services field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServices

`func (o *NetworkingEndpoint) SetServices(v []NetworkingService)`

SetServices sets Services field to given value.

### HasServices

`func (o *NetworkingEndpoint) HasServices() bool`

HasServices returns a boolean if a field has been set.

### SetServicesNil

`func (o *NetworkingEndpoint) SetServicesNil(b bool)`

 SetServicesNil sets the value for Services to be an explicit nil

### UnsetServices
`func (o *NetworkingEndpoint) UnsetServices()`

UnsetServices ensures that no value is present for Services, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


