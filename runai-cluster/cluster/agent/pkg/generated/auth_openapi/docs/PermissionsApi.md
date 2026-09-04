# \PermissionsApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetPermissions**](PermissionsApi.md#GetPermissions) | **Get** /api/v1/authorization/permissions | Get user permissions summary.
[**GetPermittedScopes**](PermissionsApi.md#GetPermittedScopes) | **Post** /api/v1/authorization/permitted-scopes | Calculate user permitted scopes for an action on a resource.



## GetPermissions

> []Permission GetPermissions(ctx).Execute()

Get user permissions summary.

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
    resp, r, err := apiClient.PermissionsApi.GetPermissions(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PermissionsApi.GetPermissions``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPermissions`: []Permission
    fmt.Fprintf(os.Stdout, "Response from `PermissionsApi.GetPermissions`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetPermissionsRequest struct via the builder pattern


### Return type

[**[]Permission**](Permission.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPermittedScopes

> PermittedScopesActions GetPermittedScopes(ctx).GetPermittedScopesRequest(getPermittedScopesRequest).Execute()

Calculate user permitted scopes for an action on a resource.

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
    getPermittedScopesRequest := *openapiclient.NewGetPermittedScopesRequest(openapiclient.ResourceType("department")) // GetPermittedScopesRequest | The request parameters.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PermissionsApi.GetPermittedScopes(context.Background()).GetPermittedScopesRequest(getPermittedScopesRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PermissionsApi.GetPermittedScopes``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPermittedScopes`: PermittedScopesActions
    fmt.Fprintf(os.Stdout, "Response from `PermissionsApi.GetPermittedScopes`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetPermittedScopesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **getPermittedScopesRequest** | [**GetPermittedScopesRequest**](GetPermittedScopesRequest.md) | The request parameters. | 

### Return type

[**PermittedScopesActions**](PermittedScopesActions.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

