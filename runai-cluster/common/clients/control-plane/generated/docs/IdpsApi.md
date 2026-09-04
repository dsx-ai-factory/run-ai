# \IdpsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateIdp**](IdpsApi.md#CreateIdp) | **Post** /api/v1/idps | configure external idp
[**DeleteIdp**](IdpsApi.md#DeleteIdp) | **Delete** /api/v1/idps/{idp} | delete external idp
[**GetIdp**](IdpsApi.md#GetIdp) | **Get** /api/v1/idps/{idp} | get external idp
[**GetIdpMappers**](IdpsApi.md#GetIdpMappers) | **Get** /api/v1/idps/{idp}/mappers | get idp mappers
[**GetIdpResources**](IdpsApi.md#GetIdpResources) | **Get** /api/v1/idps/{idp}/resources | get idp resources
[**GetIdps**](IdpsApi.md#GetIdps) | **Get** /api/v1/idps | get external idps
[**UpdateIdp**](IdpsApi.md#UpdateIdp) | **Put** /api/v1/idps/{idp} | Update external idp
[**UpdateIdpMappers**](IdpsApi.md#UpdateIdpMappers) | **Put** /api/v1/idps/{idp}/mappers | update idp mappers



## CreateIdp

> IdpPostResponse CreateIdp(ctx).IdpCreationRequest(idpCreationRequest).Execute()

configure external idp

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
    idpCreationRequest := *openapiclient.NewIdpCreationRequest("Type_example") // IdpCreationRequest | idp configuration

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.IdpsApi.CreateIdp(context.Background()).IdpCreationRequest(idpCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `IdpsApi.CreateIdp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateIdp`: IdpPostResponse
    fmt.Fprintf(os.Stdout, "Response from `IdpsApi.CreateIdp`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateIdpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **idpCreationRequest** | [**IdpCreationRequest**](IdpCreationRequest.md) | idp configuration | 

### Return type

[**IdpPostResponse**](IdpPostResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteIdp

> DeleteIdp(ctx, idp).Execute()

delete external idp

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
    idp := "idp_example" // string | The alias of the idp

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.IdpsApi.DeleteIdp(context.Background(), idp).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `IdpsApi.DeleteIdp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**idp** | **string** | The alias of the idp | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteIdpRequest struct via the builder pattern


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


## GetIdp

> Idp1 GetIdp(ctx, idp).Execute()

get external idp

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
    idp := "idp_example" // string | The alias of the idp

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.IdpsApi.GetIdp(context.Background(), idp).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `IdpsApi.GetIdp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetIdp`: Idp1
    fmt.Fprintf(os.Stdout, "Response from `IdpsApi.GetIdp`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**idp** | **string** | The alias of the idp | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetIdpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Idp1**](Idp1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetIdpMappers

> Mappers GetIdpMappers(ctx, idp).Execute()

get idp mappers

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
    idp := "idp_example" // string | The alias of the idp

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.IdpsApi.GetIdpMappers(context.Background(), idp).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `IdpsApi.GetIdpMappers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetIdpMappers`: Mappers
    fmt.Fprintf(os.Stdout, "Response from `IdpsApi.GetIdpMappers`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**idp** | **string** | The alias of the idp | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetIdpMappersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Mappers**](Mappers.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetIdpResources

> IdpResources GetIdpResources(ctx, idp).Execute()

get idp resources

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
    idp := "idp_example" // string | The alias of the idp

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.IdpsApi.GetIdpResources(context.Background(), idp).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `IdpsApi.GetIdpResources``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetIdpResources`: IdpResources
    fmt.Fprintf(os.Stdout, "Response from `IdpsApi.GetIdpResources`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**idp** | **string** | The alias of the idp | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetIdpResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**IdpResources**](IdpResources.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetIdps

> []Idp1 GetIdps(ctx).Execute()

get external idps

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
    resp, r, err := apiClient.IdpsApi.GetIdps(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `IdpsApi.GetIdps``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetIdps`: []Idp1
    fmt.Fprintf(os.Stdout, "Response from `IdpsApi.GetIdps`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetIdpsRequest struct via the builder pattern


### Return type

[**[]Idp1**](Idp1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateIdp

> UpdateIdp(ctx, idp).IdpCreationRequest(idpCreationRequest).Execute()

Update external idp

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
    idp := "idp_example" // string | The alias of the idp
    idpCreationRequest := *openapiclient.NewIdpCreationRequest("Type_example") // IdpCreationRequest | Updated IdP object

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.IdpsApi.UpdateIdp(context.Background(), idp).IdpCreationRequest(idpCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `IdpsApi.UpdateIdp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**idp** | **string** | The alias of the idp | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateIdpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **idpCreationRequest** | [**IdpCreationRequest**](IdpCreationRequest.md) | Updated IdP object | 

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


## UpdateIdpMappers

> UpdateIdpMappers(ctx, idp).Mappers(mappers).Execute()

update idp mappers

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
    idp := "idp_example" // string | The alias of the idp
    mappers := *openapiclient.NewMappers() // Mappers | map of our mappers and their source in the external idp

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.IdpsApi.UpdateIdpMappers(context.Background(), idp).Mappers(mappers).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `IdpsApi.UpdateIdpMappers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**idp** | **string** | The alias of the idp | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateIdpMappersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **mappers** | [**Mappers**](Mappers.md) | map of our mappers and their source in the external idp | 

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

