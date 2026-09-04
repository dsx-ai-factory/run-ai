# OidcData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClientId** | **string** |  | 
**ClientSecret** | **string** |  | 
**Scopes** | Pointer to **[]string** |  | [optional] 
**MandatoryClaim** | Pointer to [**NullableMandatoryClaim**](MandatoryClaim.md) |  | [optional] 

## Methods

### NewOidcData

`func NewOidcData(clientId string, clientSecret string, ) *OidcData`

NewOidcData instantiates a new OidcData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOidcDataWithDefaults

`func NewOidcDataWithDefaults() *OidcData`

NewOidcDataWithDefaults instantiates a new OidcData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClientId

`func (o *OidcData) GetClientId() string`

GetClientId returns the ClientId field if non-nil, zero value otherwise.

### GetClientIdOk

`func (o *OidcData) GetClientIdOk() (*string, bool)`

GetClientIdOk returns a tuple with the ClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientId

`func (o *OidcData) SetClientId(v string)`

SetClientId sets ClientId field to given value.


### GetClientSecret

`func (o *OidcData) GetClientSecret() string`

GetClientSecret returns the ClientSecret field if non-nil, zero value otherwise.

### GetClientSecretOk

`func (o *OidcData) GetClientSecretOk() (*string, bool)`

GetClientSecretOk returns a tuple with the ClientSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientSecret

`func (o *OidcData) SetClientSecret(v string)`

SetClientSecret sets ClientSecret field to given value.


### GetScopes

`func (o *OidcData) GetScopes() []string`

GetScopes returns the Scopes field if non-nil, zero value otherwise.

### GetScopesOk

`func (o *OidcData) GetScopesOk() (*[]string, bool)`

GetScopesOk returns a tuple with the Scopes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopes

`func (o *OidcData) SetScopes(v []string)`

SetScopes sets Scopes field to given value.

### HasScopes

`func (o *OidcData) HasScopes() bool`

HasScopes returns a boolean if a field has been set.

### GetMandatoryClaim

`func (o *OidcData) GetMandatoryClaim() MandatoryClaim`

GetMandatoryClaim returns the MandatoryClaim field if non-nil, zero value otherwise.

### GetMandatoryClaimOk

`func (o *OidcData) GetMandatoryClaimOk() (*MandatoryClaim, bool)`

GetMandatoryClaimOk returns a tuple with the MandatoryClaim field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMandatoryClaim

`func (o *OidcData) SetMandatoryClaim(v MandatoryClaim)`

SetMandatoryClaim sets MandatoryClaim field to given value.

### HasMandatoryClaim

`func (o *OidcData) HasMandatoryClaim() bool`

HasMandatoryClaim returns a boolean if a field has been set.

### SetMandatoryClaimNil

`func (o *OidcData) SetMandatoryClaimNil(b bool)`

 SetMandatoryClaimNil sets the value for MandatoryClaim to be an explicit nil

### UnsetMandatoryClaim
`func (o *OidcData) UnsetMandatoryClaim()`

UnsetMandatoryClaim ensures that no value is present for MandatoryClaim, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


