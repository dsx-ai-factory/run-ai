# \TokensApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AppToken**](TokensApi.md#AppToken) | **Post** /v1/k8s/auth/oauth/apptoken | get application token
[**ExchangeCodeForToken**](TokensApi.md#ExchangeCodeForToken) | **Get** /v1/k8s/auth/token/exchange | exchange code for token
[**GrantToken**](TokensApi.md#GrantToken) | **Post** /api/v1/token | Create an application token.
[**RefreshToken**](TokensApi.md#RefreshToken) | **Post** /v1/k8s/auth/oauth/tokens/refresh | refresh token



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
    openapiclient "./openapi"
)

func main() {
    appTokenRequest := *openapiclient.NewAppTokenRequest("Id_example", "Name_example", "Secret_example") // AppTokenRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TokensApi.AppToken(context.Background()).AppTokenRequest(appTokenRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TokensApi.AppToken``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AppToken`: AppTokenResponse
    fmt.Fprintf(os.Stdout, "Response from `TokensApi.AppToken`: %v\n", resp)
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


## ExchangeCodeForToken

> TokenResponse ExchangeCodeForToken(ctx).RedirectUri(redirectUri).Code(code).Execute()

exchange code for token



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    redirectUri := "redirectUri_example" // string | The redirect uri to redirect to after the authorization server completes the authorization flow
    code := "code_example" // string | The exchange code retrieved from the idp server

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TokensApi.ExchangeCodeForToken(context.Background()).RedirectUri(redirectUri).Code(code).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TokensApi.ExchangeCodeForToken``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ExchangeCodeForToken`: TokenResponse
    fmt.Fprintf(os.Stdout, "Response from `TokensApi.ExchangeCodeForToken`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiExchangeCodeForTokenRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **redirectUri** | **string** | The redirect uri to redirect to after the authorization server completes the authorization flow | 
 **code** | **string** | The exchange code retrieved from the idp server | 

### Return type

[**TokenResponse**](TokenResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GrantToken

> TokenResponse GrantToken(ctx).UserAgent(userAgent).TokenRequest(tokenRequest).Execute()

Create an application token.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    userAgent := "userAgent_example" // string |  (optional)
    tokenRequest := *openapiclient.NewTokenRequest() // TokenRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TokensApi.GrantToken(context.Background()).UserAgent(userAgent).TokenRequest(tokenRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TokensApi.GrantToken``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GrantToken`: TokenResponse
    fmt.Fprintf(os.Stdout, "Response from `TokensApi.GrantToken`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGrantTokenRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userAgent** | **string** |  | 
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


## RefreshToken

> TokenResponse RefreshToken(ctx).RefreshToken(refreshToken).Execute()

refresh token



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    refreshToken := "refreshToken_example" // string | The refresh token retrieved from the idp server

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TokensApi.RefreshToken(context.Background()).RefreshToken(refreshToken).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TokensApi.RefreshToken``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `RefreshToken`: TokenResponse
    fmt.Fprintf(os.Stdout, "Response from `TokensApi.RefreshToken`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRefreshTokenRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **refreshToken** | **string** | The refresh token retrieved from the idp server | 

### Return type

[**TokenResponse**](TokenResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

