# OidcCreationData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClientId** | **string** |  | 
**ClientSecret** | **string** |  | 
**DiscoverDocumentUrl** | **string** |  | 
**DefaultScope** | Pointer to **[]string** |  | [optional] 

## Methods

### NewOidcCreationData

`func NewOidcCreationData(clientId string, clientSecret string, discoverDocumentUrl string, ) *OidcCreationData`

NewOidcCreationData instantiates a new OidcCreationData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOidcCreationDataWithDefaults

`func NewOidcCreationDataWithDefaults() *OidcCreationData`

NewOidcCreationDataWithDefaults instantiates a new OidcCreationData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClientId

`func (o *OidcCreationData) GetClientId() string`

GetClientId returns the ClientId field if non-nil, zero value otherwise.

### GetClientIdOk

`func (o *OidcCreationData) GetClientIdOk() (*string, bool)`

GetClientIdOk returns a tuple with the ClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientId

`func (o *OidcCreationData) SetClientId(v string)`

SetClientId sets ClientId field to given value.


### GetClientSecret

`func (o *OidcCreationData) GetClientSecret() string`

GetClientSecret returns the ClientSecret field if non-nil, zero value otherwise.

### GetClientSecretOk

`func (o *OidcCreationData) GetClientSecretOk() (*string, bool)`

GetClientSecretOk returns a tuple with the ClientSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientSecret

`func (o *OidcCreationData) SetClientSecret(v string)`

SetClientSecret sets ClientSecret field to given value.


### GetDiscoverDocumentUrl

`func (o *OidcCreationData) GetDiscoverDocumentUrl() string`

GetDiscoverDocumentUrl returns the DiscoverDocumentUrl field if non-nil, zero value otherwise.

### GetDiscoverDocumentUrlOk

`func (o *OidcCreationData) GetDiscoverDocumentUrlOk() (*string, bool)`

GetDiscoverDocumentUrlOk returns a tuple with the DiscoverDocumentUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDiscoverDocumentUrl

`func (o *OidcCreationData) SetDiscoverDocumentUrl(v string)`

SetDiscoverDocumentUrl sets DiscoverDocumentUrl field to given value.


### GetDefaultScope

`func (o *OidcCreationData) GetDefaultScope() []string`

GetDefaultScope returns the DefaultScope field if non-nil, zero value otherwise.

### GetDefaultScopeOk

`func (o *OidcCreationData) GetDefaultScopeOk() (*[]string, bool)`

GetDefaultScopeOk returns a tuple with the DefaultScope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultScope

`func (o *OidcCreationData) SetDefaultScope(v []string)`

SetDefaultScope sets DefaultScope field to given value.

### HasDefaultScope

`func (o *OidcCreationData) HasDefaultScope() bool`

HasDefaultScope returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


