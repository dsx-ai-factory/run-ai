# OcpCreationData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClientId** | Pointer to **string** |  | [optional] 
**IdpBaseUrl** | **string** |  | 
**ClientSecret** | **string** |  | 
**Scopes** | Pointer to **[]string** |  | [optional] 

## Methods

### NewOcpCreationData

`func NewOcpCreationData(idpBaseUrl string, clientSecret string, ) *OcpCreationData`

NewOcpCreationData instantiates a new OcpCreationData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOcpCreationDataWithDefaults

`func NewOcpCreationDataWithDefaults() *OcpCreationData`

NewOcpCreationDataWithDefaults instantiates a new OcpCreationData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClientId

`func (o *OcpCreationData) GetClientId() string`

GetClientId returns the ClientId field if non-nil, zero value otherwise.

### GetClientIdOk

`func (o *OcpCreationData) GetClientIdOk() (*string, bool)`

GetClientIdOk returns a tuple with the ClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientId

`func (o *OcpCreationData) SetClientId(v string)`

SetClientId sets ClientId field to given value.

### HasClientId

`func (o *OcpCreationData) HasClientId() bool`

HasClientId returns a boolean if a field has been set.

### GetIdpBaseUrl

`func (o *OcpCreationData) GetIdpBaseUrl() string`

GetIdpBaseUrl returns the IdpBaseUrl field if non-nil, zero value otherwise.

### GetIdpBaseUrlOk

`func (o *OcpCreationData) GetIdpBaseUrlOk() (*string, bool)`

GetIdpBaseUrlOk returns a tuple with the IdpBaseUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdpBaseUrl

`func (o *OcpCreationData) SetIdpBaseUrl(v string)`

SetIdpBaseUrl sets IdpBaseUrl field to given value.


### GetClientSecret

`func (o *OcpCreationData) GetClientSecret() string`

GetClientSecret returns the ClientSecret field if non-nil, zero value otherwise.

### GetClientSecretOk

`func (o *OcpCreationData) GetClientSecretOk() (*string, bool)`

GetClientSecretOk returns a tuple with the ClientSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientSecret

`func (o *OcpCreationData) SetClientSecret(v string)`

SetClientSecret sets ClientSecret field to given value.


### GetScopes

`func (o *OcpCreationData) GetScopes() []string`

GetScopes returns the Scopes field if non-nil, zero value otherwise.

### GetScopesOk

`func (o *OcpCreationData) GetScopesOk() (*[]string, bool)`

GetScopesOk returns a tuple with the Scopes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopes

`func (o *OcpCreationData) SetScopes(v []string)`

SetScopes sets Scopes field to given value.

### HasScopes

`func (o *OcpCreationData) HasScopes() bool`

HasScopes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


