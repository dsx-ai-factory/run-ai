# \LoginsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetLoginScreenURL**](LoginsApi.md#GetLoginScreenURL) | **Get** /api/v1/login | Get login url



## GetLoginScreenURL

> GetLoginScreenURL(ctx).UserAgent(userAgent).AuthenticationFlow(authenticationFlow).RedirectUrl(redirectUrl).Execute()

Get login url



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
    userAgent := "userAgent_example" // string | 
    authenticationFlow := "authenticationFlow_example" // string |  (optional)
    redirectUrl := "redirectUrl_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.LoginsApi.GetLoginScreenURL(context.Background()).UserAgent(userAgent).AuthenticationFlow(authenticationFlow).RedirectUrl(redirectUrl).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LoginsApi.GetLoginScreenURL``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetLoginScreenURLRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userAgent** | **string** |  | 
 **authenticationFlow** | **string** |  | 
 **redirectUrl** | **string** |  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

