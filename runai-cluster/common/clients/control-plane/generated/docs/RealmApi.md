# \RealmApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateRealm**](RealmApi.md#CreateRealm) | **Post** /api/v1/realms | Create runai realm
[**DeleteRealm**](RealmApi.md#DeleteRealm) | **Delete** /internal/realms/{realm} | delete realm



## CreateRealm

> RealmCreationResponse CreateRealm(ctx).RealmCreationRequest(realmCreationRequest).Execute()

Create runai realm



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
    realmCreationRequest := *openapiclient.NewRealmCreationRequest("RealmName_example", "TenantId_example") // RealmCreationRequest | The tenant on behalf realm will be created.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.RealmApi.CreateRealm(context.Background()).RealmCreationRequest(realmCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `RealmApi.CreateRealm``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateRealm`: RealmCreationResponse
    fmt.Fprintf(os.Stdout, "Response from `RealmApi.CreateRealm`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateRealmRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **realmCreationRequest** | [**RealmCreationRequest**](RealmCreationRequest.md) | The tenant on behalf realm will be created. | 

### Return type

[**RealmCreationResponse**](RealmCreationResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRealm

> DeleteRealm200Response DeleteRealm(ctx, realm).IsHardDelete(isHardDelete).Execute()

delete realm

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
    realm := "realm_example" // string | The name of the realm
    isHardDelete := true // bool | Boolean indicator for soft/hard delete (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.RealmApi.DeleteRealm(context.Background(), realm).IsHardDelete(isHardDelete).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `RealmApi.DeleteRealm``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteRealm`: DeleteRealm200Response
    fmt.Fprintf(os.Stdout, "Response from `RealmApi.DeleteRealm`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**realm** | **string** | The name of the realm | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRealmRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **isHardDelete** | **bool** | Boolean indicator for soft/hard delete | 

### Return type

[**DeleteRealm200Response**](DeleteRealm200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

