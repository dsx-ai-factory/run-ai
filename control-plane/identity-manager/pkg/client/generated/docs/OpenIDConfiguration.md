# OpenIDConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Issuer** | **string** | The authorization server&#39;s issuer identifier URL | 
**AuthorizationEndpoint** | **string** | URL of the OAuth2 authorization endpoint | 
**TokenEndpoint** | **string** | URL of the OAuth2 token endpoint | 
**UserinfoEndpoint** | Pointer to **string** | URL of the UserInfo endpoint | [optional] 
**JwksUri** | **string** | URL of the JSON Web Key Set document | 
**RegistrationEndpoint** | Pointer to **string** | URL of the Dynamic Client Registration endpoint | [optional] 
**ScopesSupported** | Pointer to **[]string** | List of OAuth2 scope values supported | [optional] 
**ResponseTypesSupported** | **[]string** | List of OAuth2 response_type values supported | 
**ResponseModesSupported** | Pointer to **[]string** | List of OAuth2 response_mode values supported | [optional] 
**GrantTypesSupported** | Pointer to **[]string** | List of OAuth2 grant type values supported | [optional] 
**SubjectTypesSupported** | **[]string** | List of Subject Identifier types supported | 
**IdTokenSigningAlgValuesSupported** | **[]string** | List of JWS signing algorithms supported for ID Token | 
**TokenEndpointAuthMethodsSupported** | Pointer to **[]string** | List of client authentication methods supported by token endpoint | [optional] 
**ClaimsSupported** | Pointer to **[]string** | List of claim names supported | [optional] 
**CodeChallengeMethodsSupported** | Pointer to **[]string** | List of PKCE code challenge methods supported | [optional] 
**EndSessionEndpoint** | Pointer to **string** | URL of the logout endpoint | [optional] 

## Methods

### NewOpenIDConfiguration

`func NewOpenIDConfiguration(issuer string, authorizationEndpoint string, tokenEndpoint string, jwksUri string, responseTypesSupported []string, subjectTypesSupported []string, idTokenSigningAlgValuesSupported []string, ) *OpenIDConfiguration`

NewOpenIDConfiguration instantiates a new OpenIDConfiguration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOpenIDConfigurationWithDefaults

`func NewOpenIDConfigurationWithDefaults() *OpenIDConfiguration`

NewOpenIDConfigurationWithDefaults instantiates a new OpenIDConfiguration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIssuer

`func (o *OpenIDConfiguration) GetIssuer() string`

GetIssuer returns the Issuer field if non-nil, zero value otherwise.

### GetIssuerOk

`func (o *OpenIDConfiguration) GetIssuerOk() (*string, bool)`

GetIssuerOk returns a tuple with the Issuer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssuer

`func (o *OpenIDConfiguration) SetIssuer(v string)`

SetIssuer sets Issuer field to given value.


### GetAuthorizationEndpoint

`func (o *OpenIDConfiguration) GetAuthorizationEndpoint() string`

GetAuthorizationEndpoint returns the AuthorizationEndpoint field if non-nil, zero value otherwise.

### GetAuthorizationEndpointOk

`func (o *OpenIDConfiguration) GetAuthorizationEndpointOk() (*string, bool)`

GetAuthorizationEndpointOk returns a tuple with the AuthorizationEndpoint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizationEndpoint

`func (o *OpenIDConfiguration) SetAuthorizationEndpoint(v string)`

SetAuthorizationEndpoint sets AuthorizationEndpoint field to given value.


### GetTokenEndpoint

`func (o *OpenIDConfiguration) GetTokenEndpoint() string`

GetTokenEndpoint returns the TokenEndpoint field if non-nil, zero value otherwise.

### GetTokenEndpointOk

`func (o *OpenIDConfiguration) GetTokenEndpointOk() (*string, bool)`

GetTokenEndpointOk returns a tuple with the TokenEndpoint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTokenEndpoint

`func (o *OpenIDConfiguration) SetTokenEndpoint(v string)`

SetTokenEndpoint sets TokenEndpoint field to given value.


### GetUserinfoEndpoint

`func (o *OpenIDConfiguration) GetUserinfoEndpoint() string`

GetUserinfoEndpoint returns the UserinfoEndpoint field if non-nil, zero value otherwise.

### GetUserinfoEndpointOk

`func (o *OpenIDConfiguration) GetUserinfoEndpointOk() (*string, bool)`

GetUserinfoEndpointOk returns a tuple with the UserinfoEndpoint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserinfoEndpoint

`func (o *OpenIDConfiguration) SetUserinfoEndpoint(v string)`

SetUserinfoEndpoint sets UserinfoEndpoint field to given value.

### HasUserinfoEndpoint

`func (o *OpenIDConfiguration) HasUserinfoEndpoint() bool`

HasUserinfoEndpoint returns a boolean if a field has been set.

### GetJwksUri

`func (o *OpenIDConfiguration) GetJwksUri() string`

GetJwksUri returns the JwksUri field if non-nil, zero value otherwise.

### GetJwksUriOk

`func (o *OpenIDConfiguration) GetJwksUriOk() (*string, bool)`

GetJwksUriOk returns a tuple with the JwksUri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJwksUri

`func (o *OpenIDConfiguration) SetJwksUri(v string)`

SetJwksUri sets JwksUri field to given value.


### GetRegistrationEndpoint

`func (o *OpenIDConfiguration) GetRegistrationEndpoint() string`

GetRegistrationEndpoint returns the RegistrationEndpoint field if non-nil, zero value otherwise.

### GetRegistrationEndpointOk

`func (o *OpenIDConfiguration) GetRegistrationEndpointOk() (*string, bool)`

GetRegistrationEndpointOk returns a tuple with the RegistrationEndpoint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegistrationEndpoint

`func (o *OpenIDConfiguration) SetRegistrationEndpoint(v string)`

SetRegistrationEndpoint sets RegistrationEndpoint field to given value.

### HasRegistrationEndpoint

`func (o *OpenIDConfiguration) HasRegistrationEndpoint() bool`

HasRegistrationEndpoint returns a boolean if a field has been set.

### GetScopesSupported

`func (o *OpenIDConfiguration) GetScopesSupported() []string`

GetScopesSupported returns the ScopesSupported field if non-nil, zero value otherwise.

### GetScopesSupportedOk

`func (o *OpenIDConfiguration) GetScopesSupportedOk() (*[]string, bool)`

GetScopesSupportedOk returns a tuple with the ScopesSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopesSupported

`func (o *OpenIDConfiguration) SetScopesSupported(v []string)`

SetScopesSupported sets ScopesSupported field to given value.

### HasScopesSupported

`func (o *OpenIDConfiguration) HasScopesSupported() bool`

HasScopesSupported returns a boolean if a field has been set.

### GetResponseTypesSupported

`func (o *OpenIDConfiguration) GetResponseTypesSupported() []string`

GetResponseTypesSupported returns the ResponseTypesSupported field if non-nil, zero value otherwise.

### GetResponseTypesSupportedOk

`func (o *OpenIDConfiguration) GetResponseTypesSupportedOk() (*[]string, bool)`

GetResponseTypesSupportedOk returns a tuple with the ResponseTypesSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponseTypesSupported

`func (o *OpenIDConfiguration) SetResponseTypesSupported(v []string)`

SetResponseTypesSupported sets ResponseTypesSupported field to given value.


### GetResponseModesSupported

`func (o *OpenIDConfiguration) GetResponseModesSupported() []string`

GetResponseModesSupported returns the ResponseModesSupported field if non-nil, zero value otherwise.

### GetResponseModesSupportedOk

`func (o *OpenIDConfiguration) GetResponseModesSupportedOk() (*[]string, bool)`

GetResponseModesSupportedOk returns a tuple with the ResponseModesSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponseModesSupported

`func (o *OpenIDConfiguration) SetResponseModesSupported(v []string)`

SetResponseModesSupported sets ResponseModesSupported field to given value.

### HasResponseModesSupported

`func (o *OpenIDConfiguration) HasResponseModesSupported() bool`

HasResponseModesSupported returns a boolean if a field has been set.

### GetGrantTypesSupported

`func (o *OpenIDConfiguration) GetGrantTypesSupported() []string`

GetGrantTypesSupported returns the GrantTypesSupported field if non-nil, zero value otherwise.

### GetGrantTypesSupportedOk

`func (o *OpenIDConfiguration) GetGrantTypesSupportedOk() (*[]string, bool)`

GetGrantTypesSupportedOk returns a tuple with the GrantTypesSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGrantTypesSupported

`func (o *OpenIDConfiguration) SetGrantTypesSupported(v []string)`

SetGrantTypesSupported sets GrantTypesSupported field to given value.

### HasGrantTypesSupported

`func (o *OpenIDConfiguration) HasGrantTypesSupported() bool`

HasGrantTypesSupported returns a boolean if a field has been set.

### GetSubjectTypesSupported

`func (o *OpenIDConfiguration) GetSubjectTypesSupported() []string`

GetSubjectTypesSupported returns the SubjectTypesSupported field if non-nil, zero value otherwise.

### GetSubjectTypesSupportedOk

`func (o *OpenIDConfiguration) GetSubjectTypesSupportedOk() (*[]string, bool)`

GetSubjectTypesSupportedOk returns a tuple with the SubjectTypesSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectTypesSupported

`func (o *OpenIDConfiguration) SetSubjectTypesSupported(v []string)`

SetSubjectTypesSupported sets SubjectTypesSupported field to given value.


### GetIdTokenSigningAlgValuesSupported

`func (o *OpenIDConfiguration) GetIdTokenSigningAlgValuesSupported() []string`

GetIdTokenSigningAlgValuesSupported returns the IdTokenSigningAlgValuesSupported field if non-nil, zero value otherwise.

### GetIdTokenSigningAlgValuesSupportedOk

`func (o *OpenIDConfiguration) GetIdTokenSigningAlgValuesSupportedOk() (*[]string, bool)`

GetIdTokenSigningAlgValuesSupportedOk returns a tuple with the IdTokenSigningAlgValuesSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdTokenSigningAlgValuesSupported

`func (o *OpenIDConfiguration) SetIdTokenSigningAlgValuesSupported(v []string)`

SetIdTokenSigningAlgValuesSupported sets IdTokenSigningAlgValuesSupported field to given value.


### GetTokenEndpointAuthMethodsSupported

`func (o *OpenIDConfiguration) GetTokenEndpointAuthMethodsSupported() []string`

GetTokenEndpointAuthMethodsSupported returns the TokenEndpointAuthMethodsSupported field if non-nil, zero value otherwise.

### GetTokenEndpointAuthMethodsSupportedOk

`func (o *OpenIDConfiguration) GetTokenEndpointAuthMethodsSupportedOk() (*[]string, bool)`

GetTokenEndpointAuthMethodsSupportedOk returns a tuple with the TokenEndpointAuthMethodsSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTokenEndpointAuthMethodsSupported

`func (o *OpenIDConfiguration) SetTokenEndpointAuthMethodsSupported(v []string)`

SetTokenEndpointAuthMethodsSupported sets TokenEndpointAuthMethodsSupported field to given value.

### HasTokenEndpointAuthMethodsSupported

`func (o *OpenIDConfiguration) HasTokenEndpointAuthMethodsSupported() bool`

HasTokenEndpointAuthMethodsSupported returns a boolean if a field has been set.

### GetClaimsSupported

`func (o *OpenIDConfiguration) GetClaimsSupported() []string`

GetClaimsSupported returns the ClaimsSupported field if non-nil, zero value otherwise.

### GetClaimsSupportedOk

`func (o *OpenIDConfiguration) GetClaimsSupportedOk() (*[]string, bool)`

GetClaimsSupportedOk returns a tuple with the ClaimsSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimsSupported

`func (o *OpenIDConfiguration) SetClaimsSupported(v []string)`

SetClaimsSupported sets ClaimsSupported field to given value.

### HasClaimsSupported

`func (o *OpenIDConfiguration) HasClaimsSupported() bool`

HasClaimsSupported returns a boolean if a field has been set.

### GetCodeChallengeMethodsSupported

`func (o *OpenIDConfiguration) GetCodeChallengeMethodsSupported() []string`

GetCodeChallengeMethodsSupported returns the CodeChallengeMethodsSupported field if non-nil, zero value otherwise.

### GetCodeChallengeMethodsSupportedOk

`func (o *OpenIDConfiguration) GetCodeChallengeMethodsSupportedOk() (*[]string, bool)`

GetCodeChallengeMethodsSupportedOk returns a tuple with the CodeChallengeMethodsSupported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCodeChallengeMethodsSupported

`func (o *OpenIDConfiguration) SetCodeChallengeMethodsSupported(v []string)`

SetCodeChallengeMethodsSupported sets CodeChallengeMethodsSupported field to given value.

### HasCodeChallengeMethodsSupported

`func (o *OpenIDConfiguration) HasCodeChallengeMethodsSupported() bool`

HasCodeChallengeMethodsSupported returns a boolean if a field has been set.

### GetEndSessionEndpoint

`func (o *OpenIDConfiguration) GetEndSessionEndpoint() string`

GetEndSessionEndpoint returns the EndSessionEndpoint field if non-nil, zero value otherwise.

### GetEndSessionEndpointOk

`func (o *OpenIDConfiguration) GetEndSessionEndpointOk() (*string, bool)`

GetEndSessionEndpointOk returns a tuple with the EndSessionEndpoint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEndSessionEndpoint

`func (o *OpenIDConfiguration) SetEndSessionEndpoint(v string)`

SetEndSessionEndpoint sets EndSessionEndpoint field to given value.

### HasEndSessionEndpoint

`func (o *OpenIDConfiguration) HasEndSessionEndpoint() bool`

HasEndSessionEndpoint returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


