# OidcCreationData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClientId** | **string** |  | 
**ClientSecret** | **string** |  | 
**DiscoverDocumentUrl** | **string** |  | 
**Scopes** | Pointer to **[]string** |  | [optional] 
**MandatoryClaim** | Pointer to [**NullableMandatoryClaim**](MandatoryClaim.md) |  | [optional] 

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


### GetScopes

`func (o *OidcCreationData) GetScopes() []string`

GetScopes returns the Scopes field if non-nil, zero value otherwise.

### GetScopesOk

`func (o *OidcCreationData) GetScopesOk() (*[]string, bool)`

GetScopesOk returns a tuple with the Scopes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopes

`func (o *OidcCreationData) SetScopes(v []string)`

SetScopes sets Scopes field to given value.

### HasScopes

`func (o *OidcCreationData) HasScopes() bool`

HasScopes returns a boolean if a field has been set.

### GetMandatoryClaim

`func (o *OidcCreationData) GetMandatoryClaim() MandatoryClaim`

GetMandatoryClaim returns the MandatoryClaim field if non-nil, zero value otherwise.

### GetMandatoryClaimOk

`func (o *OidcCreationData) GetMandatoryClaimOk() (*MandatoryClaim, bool)`

GetMandatoryClaimOk returns a tuple with the MandatoryClaim field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMandatoryClaim

`func (o *OidcCreationData) SetMandatoryClaim(v MandatoryClaim)`

SetMandatoryClaim sets MandatoryClaim field to given value.

### HasMandatoryClaim

`func (o *OidcCreationData) HasMandatoryClaim() bool`

HasMandatoryClaim returns a boolean if a field has been set.

### SetMandatoryClaimNil

`func (o *OidcCreationData) SetMandatoryClaimNil(b bool)`

 SetMandatoryClaimNil sets the value for MandatoryClaim to be an explicit nil

### UnsetMandatoryClaim
`func (o *OidcCreationData) UnsetMandatoryClaim()`

UnsetMandatoryClaim ensures that no value is present for MandatoryClaim, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


