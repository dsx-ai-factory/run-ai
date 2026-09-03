# TokenRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GrantType** | Pointer to [**TokenRequestGrantTypeEnum**](TokenRequestGrantTypeEnum.md) |  | [optional] 
**AppID** | Pointer to **string** |  | [optional] 
**AppSecret** | Pointer to **string** |  | [optional] 
**Code** | Pointer to **string** |  | [optional] 
**RedirectUri** | Pointer to **string** |  | [optional] 
**RefreshToken** | Pointer to **string** |  | [optional] 
**Username** | Pointer to **string** |  | [optional] 
**Password** | Pointer to **string** |  | [optional] 
**ClientID** | Pointer to **string** |  | [optional] 
**ClientSecret** | Pointer to **string** |  | [optional] 
**ExternalToken** | Pointer to **string** |  | [optional] 
**CodeVerifier** | Pointer to **string** |  | [optional] 

## Methods

### NewTokenRequest

`func NewTokenRequest() *TokenRequest`

NewTokenRequest instantiates a new TokenRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTokenRequestWithDefaults

`func NewTokenRequestWithDefaults() *TokenRequest`

NewTokenRequestWithDefaults instantiates a new TokenRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGrantType

`func (o *TokenRequest) GetGrantType() TokenRequestGrantTypeEnum`

GetGrantType returns the GrantType field if non-nil, zero value otherwise.

### GetGrantTypeOk

`func (o *TokenRequest) GetGrantTypeOk() (*TokenRequestGrantTypeEnum, bool)`

GetGrantTypeOk returns a tuple with the GrantType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGrantType

`func (o *TokenRequest) SetGrantType(v TokenRequestGrantTypeEnum)`

SetGrantType sets GrantType field to given value.

### HasGrantType

`func (o *TokenRequest) HasGrantType() bool`

HasGrantType returns a boolean if a field has been set.

### GetAppID

`func (o *TokenRequest) GetAppID() string`

GetAppID returns the AppID field if non-nil, zero value otherwise.

### GetAppIDOk

`func (o *TokenRequest) GetAppIDOk() (*string, bool)`

GetAppIDOk returns a tuple with the AppID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAppID

`func (o *TokenRequest) SetAppID(v string)`

SetAppID sets AppID field to given value.

### HasAppID

`func (o *TokenRequest) HasAppID() bool`

HasAppID returns a boolean if a field has been set.

### GetAppSecret

`func (o *TokenRequest) GetAppSecret() string`

GetAppSecret returns the AppSecret field if non-nil, zero value otherwise.

### GetAppSecretOk

`func (o *TokenRequest) GetAppSecretOk() (*string, bool)`

GetAppSecretOk returns a tuple with the AppSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAppSecret

`func (o *TokenRequest) SetAppSecret(v string)`

SetAppSecret sets AppSecret field to given value.

### HasAppSecret

`func (o *TokenRequest) HasAppSecret() bool`

HasAppSecret returns a boolean if a field has been set.

### GetCode

`func (o *TokenRequest) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *TokenRequest) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *TokenRequest) SetCode(v string)`

SetCode sets Code field to given value.

### HasCode

`func (o *TokenRequest) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetRedirectUri

`func (o *TokenRequest) GetRedirectUri() string`

GetRedirectUri returns the RedirectUri field if non-nil, zero value otherwise.

### GetRedirectUriOk

`func (o *TokenRequest) GetRedirectUriOk() (*string, bool)`

GetRedirectUriOk returns a tuple with the RedirectUri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRedirectUri

`func (o *TokenRequest) SetRedirectUri(v string)`

SetRedirectUri sets RedirectUri field to given value.

### HasRedirectUri

`func (o *TokenRequest) HasRedirectUri() bool`

HasRedirectUri returns a boolean if a field has been set.

### GetRefreshToken

`func (o *TokenRequest) GetRefreshToken() string`

GetRefreshToken returns the RefreshToken field if non-nil, zero value otherwise.

### GetRefreshTokenOk

`func (o *TokenRequest) GetRefreshTokenOk() (*string, bool)`

GetRefreshTokenOk returns a tuple with the RefreshToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRefreshToken

`func (o *TokenRequest) SetRefreshToken(v string)`

SetRefreshToken sets RefreshToken field to given value.

### HasRefreshToken

`func (o *TokenRequest) HasRefreshToken() bool`

HasRefreshToken returns a boolean if a field has been set.

### GetUsername

`func (o *TokenRequest) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *TokenRequest) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *TokenRequest) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *TokenRequest) HasUsername() bool`

HasUsername returns a boolean if a field has been set.

### GetPassword

`func (o *TokenRequest) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *TokenRequest) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *TokenRequest) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *TokenRequest) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetClientID

`func (o *TokenRequest) GetClientID() string`

GetClientID returns the ClientID field if non-nil, zero value otherwise.

### GetClientIDOk

`func (o *TokenRequest) GetClientIDOk() (*string, bool)`

GetClientIDOk returns a tuple with the ClientID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientID

`func (o *TokenRequest) SetClientID(v string)`

SetClientID sets ClientID field to given value.

### HasClientID

`func (o *TokenRequest) HasClientID() bool`

HasClientID returns a boolean if a field has been set.

### GetClientSecret

`func (o *TokenRequest) GetClientSecret() string`

GetClientSecret returns the ClientSecret field if non-nil, zero value otherwise.

### GetClientSecretOk

`func (o *TokenRequest) GetClientSecretOk() (*string, bool)`

GetClientSecretOk returns a tuple with the ClientSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientSecret

`func (o *TokenRequest) SetClientSecret(v string)`

SetClientSecret sets ClientSecret field to given value.

### HasClientSecret

`func (o *TokenRequest) HasClientSecret() bool`

HasClientSecret returns a boolean if a field has been set.

### GetExternalToken

`func (o *TokenRequest) GetExternalToken() string`

GetExternalToken returns the ExternalToken field if non-nil, zero value otherwise.

### GetExternalTokenOk

`func (o *TokenRequest) GetExternalTokenOk() (*string, bool)`

GetExternalTokenOk returns a tuple with the ExternalToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalToken

`func (o *TokenRequest) SetExternalToken(v string)`

SetExternalToken sets ExternalToken field to given value.

### HasExternalToken

`func (o *TokenRequest) HasExternalToken() bool`

HasExternalToken returns a boolean if a field has been set.

### GetCodeVerifier

`func (o *TokenRequest) GetCodeVerifier() string`

GetCodeVerifier returns the CodeVerifier field if non-nil, zero value otherwise.

### GetCodeVerifierOk

`func (o *TokenRequest) GetCodeVerifierOk() (*string, bool)`

GetCodeVerifierOk returns a tuple with the CodeVerifier field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCodeVerifier

`func (o *TokenRequest) SetCodeVerifier(v string)`

SetCodeVerifier sets CodeVerifier field to given value.

### HasCodeVerifier

`func (o *TokenRequest) HasCodeVerifier() bool`

HasCodeVerifier returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


