# OAuthAuthorizationServer

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Issuer** | **string** | The authorization server&#39;s issuer identifier URL | 
**AuthorizationEndpoint** | Pointer to **string** | URL of the OAuth2 authorization endpoint | [optional] 
**TokenEndpoint** | Pointer to **string** | URL of the OAuth2 token endpoint | [optional] 
**UserinfoEndpoint** | Pointer to **string** | URL of the UserInfo endpoint | [optional] 
**JwksUri** | **string** | URL of the JSON Web Key Set document | 
**RegistrationEndpoint** | Pointer to **string** | URL of the Dynamic Client Registration endpoint | [optional] 
**ScopesSupported** | Pointer to **[]string** | List of OAuth2 scope values supported | [optional] 
**ResponseTypesSupported** | Pointer to **[]string** | List of OAuth2 response_type values supported | [optional] 
**ResponseModesSupported** | Pointer to **[]string** | List of OAuth2 response_mode values supported | [optional] 
**GrantTypesSupported** | Pointer to **[]string** | List of OAuth2 grant type values supported | [optional] 
**TokenEndpointAuthMethodsSupported** | Pointer to **[]string** | List of client authentication methods supported by token endpoint | [optional] 
**ClaimsSupported** | Pointer to **[]string** | List of claim names supported | [optional] 
**CodeChallengeMethodsSupported** | Pointer to **[]string** | List of PKCE code challenge methods supported | [optional] 
**EndSessionEndpoint** | Pointer to **string** | URL of the logout endpoint | [optional] 

## Methods

### NewOAuthAuthorizationServer

`func NewOAuthAuthorizationServer(issuer string, jwksUri string, ) *OAuthAuthorizationServer`

NewOAuthAuthorizationServer instantiates a new OAuthAuthorizationServer object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOAuthAuthorizationServerWithDefaults

`func NewOAuthAuthorizationServerWithDefaults() *OAuthAuthorizationServer`

NewOAuthAuthorizationServerWithDefaults instantiates a new OAuthAuthorizationServer object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIssuer

`func (o *OAuthAuthorizationServer) GetIssuer() string`

GetIssuer returns the Issuer field if non-nil, zero value otherwise.

### GetIssuerOk

`func (o *OAuthAuthorizationServer) GetIssuerOk() (*string, bool)`

GetIssuerOk returns a tuple with the Issuer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssuer

`func (o *OAuthAuthorizationServer) SetIssuer(v string)`

SetIssuer sets Issuer field to given value.


### GetAuthorizationEndpoint

`func (o *OAuthAuthorizationServer) GetAuthorizationEndpoint() string`

GetAuthorizationEndpoint returns the AuthorizationEndpoint field if non-nil, zero value otherwise.

### GetAuthorizationEndpointOk

`func (o *OAuthAuthorizationServer) GetAuthorizationEndpointOk() (*string, bool)`

GetAuthorizationEndpointOk returns a tuple with the AuthorizationEndpoint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizationEndpoint

`func (o *OAuthAuthorizationServer) SetAuthorizationEndpoint(v string)`

SetAuthorizationEndpoint sets AuthorizationEndpoint field to given value.

### HasAuthorizationEndpoint

`func (o *OAuthAuthorizationServer) HasAuthorizationEndpoint() bool`

HasAuthorizationEndpoint returns a boolean if a field has been set.

### GetTokenEndpoint

`func (o *OAuthAuthorizationServer) GetTokenEndpoint() string`

GetTokenEndpoint returns the TokenEndpoint field if non-nil, zero value otherwise.

### GetTokenEndpointOk

`func (o *OAuthAuthorizationServer) GetTokenEndpointOk() (*string, bool)`

GetTokenEndpointOk returns a tuple with the TokenEndpoint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTokenEndpoint

`func (o *OAuthAuthorizationServer) SetTokenEndpoint(v string)`

SetTokenEndpoint sets TokenEndpoint field to given value.

### HasTokenEndpoint

`func (o *OAuthAuthorizationServer) HasTokenEndpoint() bool`

HasTokenEndpoint returns a boolean if a field has been set.

### GetUserinfoEndpoint

`func (o *OAuthAuthorizationServer) GetUserinfoEndpoint() string`

GetUserinfoEndpoint returns the UserinfoEndpoint field if non-nil, zero value otherwise.

### GetUserinfoEndpointOk

`func (o *OAuthAuthorizationServer) GetUserinfoEndpointOk() (*string, bool)`

GetUserinfoEndpointOk returns a tuple with the UserinfoEndpoint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserinfoEndpoint

`func (o *OAuthAuthorizationServer) SetUserinfoEndpoint(v string)`

SetUserinfoEndpoint sets UserinfoEndpoint field to given value.

### HasUserinfoEndpoint

`func (o *OAuthAuthorizationServer) HasUserinfoEndpoint() bool`

HasUserinfoEndpoint returns a boolean if a field has been set.

### GetJwksUri

`func (o *OAuthAuthorizationServer) GetJwksUri() string`

GetJwksUri returns the JwksUri field if non-nil, zero value otherwise.

### GetJwksUriOk

`func (o *OAuthAuthorizationServer) GetJwksUriOk() (*string, bool)`

GetJwksUriOk returns a tuple with the JwksUri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJwksUri

`func (o *OAuthAuthorizationServer) SetJwksUri(v string)`

SetJwksUri sets JwksUri field to given value.


### GetRegistrationEndpoint

`func (o *OAuthAuthorizationServer) GetRegistrationEndpoint() string`

GetRegistrationEndpoint returns the RegistrationEndpoint field if non-nil, zero value otherwise.

### GetRegistrationEndpointOk

`func (o *OAuthAuthorizationServer) GetRegistrationEndpointOk() (*string, bool)`

GetRegistrationEndpointOk returns a tuple with the RegistrationEndpoint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegistrationEndpoint

`func (o *OAuthAuthorizationServer) SetRegistrationEndpoint(v string)`

SetRegistrationEndpoint sets RegistrationEndpoint field to given value.

### HasRegistrationEndpoint

`func (o *OAuthAuthorizationServer) HasRegistrationEndpoint() bool`

HasRegistrationEndpoint returns a boolean if a field has been set.

### GetScopesSupported

`func (o *OAuthAuthorizationServer) GetScopesSupported() []string`

GetScopesSupported returns the ScopesSupported field if non-nil, zero value otherwise.

### GetScopesSupportedOk

`func (o *OAuthAuthorizationServer) GetScopesSupportedOk() (*[]string, bool)`

GetScopesSupportedOk returns a tuple with the ScopesSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopesSupported

`func (o *OAuthAuthorizationServer) SetScopesSupported(v []string)`

SetScopesSupported sets ScopesSupported field to given value.

### HasScopesSupported

`func (o *OAuthAuthorizationServer) HasScopesSupported() bool`

HasScopesSupported returns a boolean if a field has been set.

### GetResponseTypesSupported

`func (o *OAuthAuthorizationServer) GetResponseTypesSupported() []string`

GetResponseTypesSupported returns the ResponseTypesSupported field if non-nil, zero value otherwise.

### GetResponseTypesSupportedOk

`func (o *OAuthAuthorizationServer) GetResponseTypesSupportedOk() (*[]string, bool)`

GetResponseTypesSupportedOk returns a tuple with the ResponseTypesSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponseTypesSupported

`func (o *OAuthAuthorizationServer) SetResponseTypesSupported(v []string)`

SetResponseTypesSupported sets ResponseTypesSupported field to given value.

### HasResponseTypesSupported

`func (o *OAuthAuthorizationServer) HasResponseTypesSupported() bool`

HasResponseTypesSupported returns a boolean if a field has been set.

### GetResponseModesSupported

`func (o *OAuthAuthorizationServer) GetResponseModesSupported() []string`

GetResponseModesSupported returns the ResponseModesSupported field if non-nil, zero value otherwise.

### GetResponseModesSupportedOk

`func (o *OAuthAuthorizationServer) GetResponseModesSupportedOk() (*[]string, bool)`

GetResponseModesSupportedOk returns a tuple with the ResponseModesSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponseModesSupported

`func (o *OAuthAuthorizationServer) SetResponseModesSupported(v []string)`

SetResponseModesSupported sets ResponseModesSupported field to given value.

### HasResponseModesSupported

`func (o *OAuthAuthorizationServer) HasResponseModesSupported() bool`

HasResponseModesSupported returns a boolean if a field has been set.

### GetGrantTypesSupported

`func (o *OAuthAuthorizationServer) GetGrantTypesSupported() []string`

GetGrantTypesSupported returns the GrantTypesSupported field if non-nil, zero value otherwise.

### GetGrantTypesSupportedOk

`func (o *OAuthAuthorizationServer) GetGrantTypesSupportedOk() (*[]string, bool)`

GetGrantTypesSupportedOk returns a tuple with the GrantTypesSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGrantTypesSupported

`func (o *OAuthAuthorizationServer) SetGrantTypesSupported(v []string)`

SetGrantTypesSupported sets GrantTypesSupported field to given value.

### HasGrantTypesSupported

`func (o *OAuthAuthorizationServer) HasGrantTypesSupported() bool`

HasGrantTypesSupported returns a boolean if a field has been set.

### GetTokenEndpointAuthMethodsSupported

`func (o *OAuthAuthorizationServer) GetTokenEndpointAuthMethodsSupported() []string`

GetTokenEndpointAuthMethodsSupported returns the TokenEndpointAuthMethodsSupported field if non-nil, zero value otherwise.

### GetTokenEndpointAuthMethodsSupportedOk

`func (o *OAuthAuthorizationServer) GetTokenEndpointAuthMethodsSupportedOk() (*[]string, bool)`

GetTokenEndpointAuthMethodsSupportedOk returns a tuple with the TokenEndpointAuthMethodsSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTokenEndpointAuthMethodsSupported

`func (o *OAuthAuthorizationServer) SetTokenEndpointAuthMethodsSupported(v []string)`

SetTokenEndpointAuthMethodsSupported sets TokenEndpointAuthMethodsSupported field to given value.

### HasTokenEndpointAuthMethodsSupported

`func (o *OAuthAuthorizationServer) HasTokenEndpointAuthMethodsSupported() bool`

HasTokenEndpointAuthMethodsSupported returns a boolean if a field has been set.

### GetClaimsSupported

`func (o *OAuthAuthorizationServer) GetClaimsSupported() []string`

GetClaimsSupported returns the ClaimsSupported field if non-nil, zero value otherwise.

### GetClaimsSupportedOk

`func (o *OAuthAuthorizationServer) GetClaimsSupportedOk() (*[]string, bool)`

GetClaimsSupportedOk returns a tuple with the ClaimsSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimsSupported

`func (o *OAuthAuthorizationServer) SetClaimsSupported(v []string)`

SetClaimsSupported sets ClaimsSupported field to given value.

### HasClaimsSupported

`func (o *OAuthAuthorizationServer) HasClaimsSupported() bool`

HasClaimsSupported returns a boolean if a field has been set.

### GetCodeChallengeMethodsSupported

`func (o *OAuthAuthorizationServer) GetCodeChallengeMethodsSupported() []string`

GetCodeChallengeMethodsSupported returns the CodeChallengeMethodsSupported field if non-nil, zero value otherwise.

### GetCodeChallengeMethodsSupportedOk

`func (o *OAuthAuthorizationServer) GetCodeChallengeMethodsSupportedOk() (*[]string, bool)`

GetCodeChallengeMethodsSupportedOk returns a tuple with the CodeChallengeMethodsSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCodeChallengeMethodsSupported

`func (o *OAuthAuthorizationServer) SetCodeChallengeMethodsSupported(v []string)`

SetCodeChallengeMethodsSupported sets CodeChallengeMethodsSupported field to given value.

### HasCodeChallengeMethodsSupported

`func (o *OAuthAuthorizationServer) HasCodeChallengeMethodsSupported() bool`

HasCodeChallengeMethodsSupported returns a boolean if a field has been set.

### GetEndSessionEndpoint

`func (o *OAuthAuthorizationServer) GetEndSessionEndpoint() string`

GetEndSessionEndpoint returns the EndSessionEndpoint field if non-nil, zero value otherwise.

### GetEndSessionEndpointOk

`func (o *OAuthAuthorizationServer) GetEndSessionEndpointOk() (*string, bool)`

GetEndSessionEndpointOk returns a tuple with the EndSessionEndpoint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEndSessionEndpoint

`func (o *OAuthAuthorizationServer) SetEndSessionEndpoint(v string)`

SetEndSessionEndpoint sets EndSessionEndpoint field to given value.

### HasEndSessionEndpoint

`func (o *OAuthAuthorizationServer) HasEndSessionEndpoint() bool`

HasEndSessionEndpoint returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


