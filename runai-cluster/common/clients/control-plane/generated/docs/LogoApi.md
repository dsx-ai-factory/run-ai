# \LogoApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetTenantLogo**](LogoApi.md#GetTenantLogo) | **Get** /api/v1/logo | Get tenant logo. (base64)
[**UploadTenantLogo1**](LogoApi.md#UploadTenantLogo1) | **Post** /api/v1/logo | Upload logo for tenant. (base64)



## GetTenantLogo

> Logo200Response GetTenantLogo(ctx).Execute()

Get tenant logo. (base64)



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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.LogoApi.GetTenantLogo(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LogoApi.GetTenantLogo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetTenantLogo`: Logo200Response
    fmt.Fprintf(os.Stdout, "Response from `LogoApi.GetTenantLogo`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantLogoRequest struct via the builder pattern


### Return type

[**Logo200Response**](Logo200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UploadTenantLogo1

> UploadTenantLogo1(ctx).Logo200Response(logo200Response).Execute()

Upload logo for tenant. (base64)



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
    logo200Response := *openapiclient.NewLogo200Response("Logo_example") // Logo200Response |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.LogoApi.UploadTenantLogo1(context.Background()).Logo200Response(logo200Response).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LogoApi.UploadTenantLogo1``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUploadTenantLogo1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **logo200Response** | [**Logo200Response**](Logo200Response.md) |  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

