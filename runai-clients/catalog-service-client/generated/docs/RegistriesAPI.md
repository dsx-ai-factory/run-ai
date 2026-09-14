# \RegistriesAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateRegistryIntegration**](RegistriesAPI.md#CreateRegistryIntegration) | **Post** /api/v1/registries | Create registry integration
[**DeleteRegistryIntegration**](RegistriesAPI.md#DeleteRegistryIntegration) | **Delete** /api/v1/registries/{RegistryIntegrationId} | Delete registry integration
[**GetRegistryIntegrationById**](RegistriesAPI.md#GetRegistryIntegrationById) | **Get** /api/v1/registries/{RegistryIntegrationId} | Get registry integration by id
[**ListRegistryIntegrations**](RegistriesAPI.md#ListRegistryIntegrations) | **Get** /api/v1/registries | List registry integrations
[**UpdateRegistryIntegration**](RegistriesAPI.md#UpdateRegistryIntegration) | **Put** /api/v1/registries/{RegistryIntegrationId} | Update registry integration



## CreateRegistryIntegration

> RegistryIntegration CreateRegistryIntegration(ctx).RegistryIntegrationCreationRequest(registryIntegrationCreationRequest).Execute()

Create registry integration



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client"
)

func main() {
	registryIntegrationCreationRequest := *openapiclient.NewRegistryIntegrationCreationRequest("Name_example", "Url_example", openapiclient.ScopeType("system"), "ScopeId_example") // RegistryIntegrationCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegistriesAPI.CreateRegistryIntegration(context.Background()).RegistryIntegrationCreationRequest(registryIntegrationCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistriesAPI.CreateRegistryIntegration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateRegistryIntegration`: RegistryIntegration
	fmt.Fprintf(os.Stdout, "Response from `RegistriesAPI.CreateRegistryIntegration`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateRegistryIntegrationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **registryIntegrationCreationRequest** | [**RegistryIntegrationCreationRequest**](RegistryIntegrationCreationRequest.md) |  | 

### Return type

[**RegistryIntegration**](RegistryIntegration.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRegistryIntegration

> HttpResponse DeleteRegistryIntegration(ctx, registryIntegrationId).Execute()

Delete registry integration



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client"
)

func main() {
	registryIntegrationId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegistriesAPI.DeleteRegistryIntegration(context.Background(), registryIntegrationId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistriesAPI.DeleteRegistryIntegration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteRegistryIntegration`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `RegistriesAPI.DeleteRegistryIntegration`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**registryIntegrationId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRegistryIntegrationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HttpResponse**](HttpResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRegistryIntegrationById

> RegistryIntegration GetRegistryIntegrationById(ctx, registryIntegrationId).Execute()

Get registry integration by id



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client"
)

func main() {
	registryIntegrationId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegistriesAPI.GetRegistryIntegrationById(context.Background(), registryIntegrationId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistriesAPI.GetRegistryIntegrationById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRegistryIntegrationById`: RegistryIntegration
	fmt.Fprintf(os.Stdout, "Response from `RegistriesAPI.GetRegistryIntegrationById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**registryIntegrationId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRegistryIntegrationByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**RegistryIntegration**](RegistryIntegration.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListRegistryIntegrations

> RegistryIntegrationListResponse ListRegistryIntegrations(ctx).Execute()

List registry integrations



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegistriesAPI.ListRegistryIntegrations(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistriesAPI.ListRegistryIntegrations``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListRegistryIntegrations`: RegistryIntegrationListResponse
	fmt.Fprintf(os.Stdout, "Response from `RegistriesAPI.ListRegistryIntegrations`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListRegistryIntegrationsRequest struct via the builder pattern


### Return type

[**RegistryIntegrationListResponse**](RegistryIntegrationListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateRegistryIntegration

> RegistryIntegration UpdateRegistryIntegration(ctx, registryIntegrationId).RegistryIntegrationUpdateRequest(registryIntegrationUpdateRequest).Execute()

Update registry integration



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client"
)

func main() {
	registryIntegrationId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	registryIntegrationUpdateRequest := *openapiclient.NewRegistryIntegrationUpdateRequest() // RegistryIntegrationUpdateRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegistriesAPI.UpdateRegistryIntegration(context.Background(), registryIntegrationId).RegistryIntegrationUpdateRequest(registryIntegrationUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistriesAPI.UpdateRegistryIntegration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateRegistryIntegration`: RegistryIntegration
	fmt.Fprintf(os.Stdout, "Response from `RegistriesAPI.UpdateRegistryIntegration`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**registryIntegrationId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateRegistryIntegrationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **registryIntegrationUpdateRequest** | [**RegistryIntegrationUpdateRequest**](RegistryIntegrationUpdateRequest.md) |  | 

### Return type

[**RegistryIntegration**](RegistryIntegration.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

