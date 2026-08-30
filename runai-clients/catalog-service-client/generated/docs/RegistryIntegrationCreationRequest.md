# RegistryIntegrationCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the registry integration. | 
**Url** | **string** | The URL of the registry integration. | 
**ScopeType** | [**ScopeType**](ScopeType.md) |  | 
**ScopeId** | **string** | The id of the cluster, department or project, depending on the scope type. | 
**Description** | Pointer to **string** | The description of the registry integration. | [optional] 
**Version** | Pointer to **string** | The version of the registry integration. | [optional] 

## Methods

### NewRegistryIntegrationCreationRequest

`func NewRegistryIntegrationCreationRequest(name string, url string, scopeType ScopeType, scopeId string, ) *RegistryIntegrationCreationRequest`

NewRegistryIntegrationCreationRequest instantiates a new RegistryIntegrationCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegistryIntegrationCreationRequestWithDefaults

`func NewRegistryIntegrationCreationRequestWithDefaults() *RegistryIntegrationCreationRequest`

NewRegistryIntegrationCreationRequestWithDefaults instantiates a new RegistryIntegrationCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *RegistryIntegrationCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *RegistryIntegrationCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *RegistryIntegrationCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetUrl

`func (o *RegistryIntegrationCreationRequest) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *RegistryIntegrationCreationRequest) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *RegistryIntegrationCreationRequest) SetUrl(v string)`

SetUrl sets Url field to given value.


### GetScopeType

`func (o *RegistryIntegrationCreationRequest) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *RegistryIntegrationCreationRequest) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *RegistryIntegrationCreationRequest) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *RegistryIntegrationCreationRequest) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *RegistryIntegrationCreationRequest) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *RegistryIntegrationCreationRequest) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetDescription

`func (o *RegistryIntegrationCreationRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *RegistryIntegrationCreationRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *RegistryIntegrationCreationRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *RegistryIntegrationCreationRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetVersion

`func (o *RegistryIntegrationCreationRequest) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *RegistryIntegrationCreationRequest) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *RegistryIntegrationCreationRequest) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *RegistryIntegrationCreationRequest) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


