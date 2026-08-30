# \TokensAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AppToken**](TokensAPI.md#AppToken) | **Post** /v1/k8s/auth/oauth/apptoken | get application token
[**GrantToken**](TokensAPI.md#GrantToken) | **Post** /api/v1/token | Create an access token v1
[**GrantTokenV2**](TokensAPI.md#GrantTokenV2) | **Post** /api/v2/token | Create an access token v2



## AppToken

> AppTokenResponse AppToken(ctx).AppTokenRequest(appTokenRequest).Execute()

get application token



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
)

func main() {
	appTokenRequest := *openapiclient.NewAppTokenRequest("Id_example", "Name_example", "Secret_example") // AppTokenRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TokensAPI.AppToken(context.Background()).AppTokenRequest(appTokenRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TokensAPI.AppToken``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AppToken`: AppTokenResponse
	fmt.Fprintf(os.Stdout, "Response from `TokensAPI.AppToken`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAppTokenRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **appTokenRequest** | [**AppTokenRequest**](AppTokenRequest.md) |  | 

### Return type

[**AppTokenResponse**](AppTokenResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GrantToken

> TokenResponse GrantToken(ctx).TokenRequest(tokenRequest).Execute()

Create an access token v1



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
)

func main() {
	tokenRequest := *openapiclient.NewTokenRequest() // TokenRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TokensAPI.GrantToken(context.Background()).TokenRequest(tokenRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TokensAPI.GrantToken``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GrantToken`: TokenResponse
	fmt.Fprintf(os.Stdout, "Response from `TokensAPI.GrantToken`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGrantTokenRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tokenRequest** | [**TokenRequest**](TokenRequest.md) |  | 

### Return type

[**TokenResponse**](TokenResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GrantTokenV2

> OAuth2TokenResponse GrantTokenV2(ctx).GrantType(grantType).ClientId(clientId).ClientSecret(clientSecret).Username(username).Password(password).ExternalToken(externalToken).RefreshToken(refreshToken).Code(code).RedirectUri(redirectUri).CodeVerifier(codeVerifier).Execute()

Create an access token v2



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
)

func main() {
	grantType := "grantType_example" // string | The OAuth2 grant type. Determines the authentication flow and which parameters are required:  `client_credentials` - Authenticate as a service account or user access key using clientId and clientSecret.  `password` - Authenticate as a local user using username and password.  `external_token_exchange` - Authenticate as a service account or user using a token issued by the configured identity provider (IdP) for a Run:ai token. Requires externalToken. 
	clientId := "clientId_example" // string | The client identifier for a service account or user access key. Required when using the client_credentials grant type. Must be used together with clientSecret. (optional)
	clientSecret := "clientSecret_example" // string | The client secret corresponding to the clientId. Required when using the client_credentials grant type. Must be used together with clientId. (optional)
	username := "username_example" // string | The username of a local user account. Required when using the password grant type. Must be used together with password. (optional)
	password := "password_example" // string | The password for the specified username. Required when using the password grant type. Must be used together with username. (optional)
	externalToken := "externalToken_example" // string | A valid token issued by the identity provider (IdP) that is configured in NVIDIA Run:ai. Required when using the external_token_exchange grant type. (optional)
	refreshToken := "refreshToken_example" // string | The refresh token used to obtain a new access token. Required only when using the refresh_token grant type. (optional)
	code := "code_example" // string | Authorization code issued by the authorization server. Used in the authorization_code grant to exchange the code for an access token. (optional)
	redirectUri := "redirectUri_example" // string | The redirect URI used during the authorization request. Required for the authorization_code grant to validate the redirect destination. (optional)
	codeVerifier := "codeVerifier_example" // string | PKCE code verifier. Required when the authorization code was requested with a code_challenge. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TokensAPI.GrantTokenV2(context.Background()).GrantType(grantType).ClientId(clientId).ClientSecret(clientSecret).Username(username).Password(password).ExternalToken(externalToken).RefreshToken(refreshToken).Code(code).RedirectUri(redirectUri).CodeVerifier(codeVerifier).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TokensAPI.GrantTokenV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GrantTokenV2`: OAuth2TokenResponse
	fmt.Fprintf(os.Stdout, "Response from `TokensAPI.GrantTokenV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGrantTokenV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **grantType** | **string** | The OAuth2 grant type. Determines the authentication flow and which parameters are required:  &#x60;client_credentials&#x60; - Authenticate as a service account or user access key using clientId and clientSecret.  &#x60;password&#x60; - Authenticate as a local user using username and password.  &#x60;external_token_exchange&#x60; - Authenticate as a service account or user using a token issued by the configured identity provider (IdP) for a Run:ai token. Requires externalToken.  | 
 **clientId** | **string** | The client identifier for a service account or user access key. Required when using the client_credentials grant type. Must be used together with clientSecret. | 
 **clientSecret** | **string** | The client secret corresponding to the clientId. Required when using the client_credentials grant type. Must be used together with clientId. | 
 **username** | **string** | The username of a local user account. Required when using the password grant type. Must be used together with password. | 
 **password** | **string** | The password for the specified username. Required when using the password grant type. Must be used together with username. | 
 **externalToken** | **string** | A valid token issued by the identity provider (IdP) that is configured in NVIDIA Run:ai. Required when using the external_token_exchange grant type. | 
 **refreshToken** | **string** | The refresh token used to obtain a new access token. Required only when using the refresh_token grant type. | 
 **code** | **string** | Authorization code issued by the authorization server. Used in the authorization_code grant to exchange the code for an access token. | 
 **redirectUri** | **string** | The redirect URI used during the authorization request. Required for the authorization_code grant to validate the redirect destination. | 
 **codeVerifier** | **string** | PKCE code verifier. Required when the authorization code was requested with a code_challenge. | 

### Return type

[**OAuth2TokenResponse**](OAuth2TokenResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/x-www-form-urlencoded
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

