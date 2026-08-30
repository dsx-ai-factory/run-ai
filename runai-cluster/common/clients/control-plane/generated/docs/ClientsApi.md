# \ClientsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateInternalClient**](ClientsApi.md#CreateInternalClient) | **Post** /internal/clients | create internal client
[**DeleteClient**](ClientsApi.md#DeleteClient) | **Delete** /internal/clients/{id} | delete client
[**GetAllClients**](ClientsApi.md#GetAllClients) | **Get** /internal/clients | get all clients
[**GetClientSecret**](ClientsApi.md#GetClientSecret) | **Get** /internal/clients/{id}/secret | get client secret
[**UpdateRedirectUris**](ClientsApi.md#UpdateRedirectUris) | **Put** /internal/clients/{id}/redirectUris | update client redirect uris by id



## CreateInternalClient

> CreateInternalClient201Response CreateInternalClient(ctx).ClientCreationRequest(clientCreationRequest).Execute()

create internal client

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
    clientCreationRequest := *openapiclient.NewClientCreationRequest(openapiclient.ClientType("Agent"), "ClientId_example") // ClientCreationRequest | Client object to create

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClientsApi.CreateInternalClient(context.Background()).ClientCreationRequest(clientCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClientsApi.CreateInternalClient``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateInternalClient`: CreateInternalClient201Response
    fmt.Fprintf(os.Stdout, "Response from `ClientsApi.CreateInternalClient`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateInternalClientRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clientCreationRequest** | [**ClientCreationRequest**](ClientCreationRequest.md) | Client object to create | 

### Return type

[**CreateInternalClient201Response**](CreateInternalClient201Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteClient

> DeleteClient(ctx, id).Execute()

delete client

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
    id := "id_example" // string | ID of the client

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClientsApi.DeleteClient(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClientsApi.DeleteClient``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | ID of the client | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteClientRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## GetAllClients

> []Client GetAllClients(ctx).Filter(filter).Execute()

get all clients

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
    filter := "runai.cluster_id:123" // string | Filter results by client attribute. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClientsApi.GetAllClients(context.Background()).Filter(filter).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClientsApi.GetAllClients``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAllClients`: []Client
    fmt.Fprintf(os.Stdout, "Response from `ClientsApi.GetAllClients`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetAllClientsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filter** | **string** | Filter results by client attribute. | 

### Return type

[**[]Client**](Client.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClientSecret

> GetClientSecret200Response GetClientSecret(ctx, id).Execute()

get client secret

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
    id := "id_example" // string | The client \"id\" to retrieve

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClientsApi.GetClientSecret(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClientsApi.GetClientSecret``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetClientSecret`: GetClientSecret200Response
    fmt.Fprintf(os.Stdout, "Response from `ClientsApi.GetClientSecret`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The client \&quot;id\&quot; to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClientSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GetClientSecret200Response**](GetClientSecret200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateRedirectUris

> UpdateRedirectUris(ctx, id).ClientRedirectUrisUpdateRequest(clientRedirectUrisUpdateRequest).Execute()

update client redirect uris by id

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
    id := "id_example" // string | The client \"id\" to retrieve
    clientRedirectUrisUpdateRequest := *openapiclient.NewClientRedirectUrisUpdateRequest([]string{"RedirectUris_example"}) // ClientRedirectUrisUpdateRequest | Updated Client redirect uris

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClientsApi.UpdateRedirectUris(context.Background(), id).ClientRedirectUrisUpdateRequest(clientRedirectUrisUpdateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClientsApi.UpdateRedirectUris``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The client \&quot;id\&quot; to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateRedirectUrisRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **clientRedirectUrisUpdateRequest** | [**ClientRedirectUrisUpdateRequest**](ClientRedirectUrisUpdateRequest.md) | Updated Client redirect uris | 

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

