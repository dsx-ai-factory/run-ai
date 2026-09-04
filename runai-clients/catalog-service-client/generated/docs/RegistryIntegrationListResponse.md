# RegistryIntegrationListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Registries** | [**[]RegistryIntegration**](RegistryIntegration.md) | The list of registry-integrations. | 

## Methods

### NewRegistryIntegrationListResponse

`func NewRegistryIntegrationListResponse(registries []RegistryIntegration, ) *RegistryIntegrationListResponse`

NewRegistryIntegrationListResponse instantiates a new RegistryIntegrationListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegistryIntegrationListResponseWithDefaults

`func NewRegistryIntegrationListResponseWithDefaults() *RegistryIntegrationListResponse`

NewRegistryIntegrationListResponseWithDefaults instantiates a new RegistryIntegrationListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRegistries

`func (o *RegistryIntegrationListResponse) GetRegistries() []RegistryIntegration`

GetRegistries returns the Registries field if non-nil, zero value otherwise.

### GetRegistriesOk

`func (o *RegistryIntegrationListResponse) GetRegistriesOk() (*[]RegistryIntegration, bool)`

GetRegistriesOk returns a tuple with the Registries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegistries

`func (o *RegistryIntegrationListResponse) SetRegistries(v []RegistryIntegration)`

SetRegistries sets Registries field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


