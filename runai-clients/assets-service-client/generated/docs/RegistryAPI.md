# \RegistryAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateRegistry**](RegistryAPI.md#CreateRegistry) | **Post** /api/v1/asset/registries | Create a registry asset.
[**DeleteRegistry**](RegistryAPI.md#DeleteRegistry) | **Delete** /api/v1/asset/registries/{AssetId} | Delete a registry asset.
[**GetRegistryById**](RegistryAPI.md#GetRegistryById) | **Get** /api/v1/asset/registries/{AssetId} | Get a registry.
[**GetRepositories**](RegistryAPI.md#GetRepositories) | **Get** /api/v1/asset/registries/{AssetId}/repositories | Get the repositories in the registry.
[**GetRepositoryTags**](RegistryAPI.md#GetRepositoryTags) | **Get** /api/v1/asset/registries/{AssetId}/repositories/tags | Get the repositories tags in the registry.
[**ListRegistries**](RegistryAPI.md#ListRegistries) | **Get** /api/v1/asset/registries | Get registries.
[**UpdateRegistry**](RegistryAPI.md#UpdateRegistry) | **Put** /api/v1/asset/registries/{AssetId} | Update a registry asset.



## CreateRegistry

> Registry CreateRegistry(ctx).RegistryCreationRequest(registryCreationRequest).Execute()

Create a registry asset.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	registryCreationRequest := *openapiclient.NewRegistryCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("system")), *openapiclient.NewRegistrySpec("https://hub.docker.com/", openapiclient.CredentialKind("password"), "admin", "12345678")) // RegistryCreationRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegistryAPI.CreateRegistry(context.Background()).RegistryCreationRequest(registryCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistryAPI.CreateRegistry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateRegistry`: Registry
	fmt.Fprintf(os.Stdout, "Response from `RegistryAPI.CreateRegistry`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateRegistryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **registryCreationRequest** | [**RegistryCreationRequest**](RegistryCreationRequest.md) |  | 

### Return type

[**Registry**](Registry.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRegistry

> DeleteRegistry(ctx, assetId).Execute()

Delete a registry asset.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RegistryAPI.DeleteRegistry(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistryAPI.DeleteRegistry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRegistryRequest struct via the builder pattern


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


## GetRegistryById

> Registry GetRegistryById(ctx, assetId).Execute()

Get a registry.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegistryAPI.GetRegistryById(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistryAPI.GetRegistryById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRegistryById`: Registry
	fmt.Fprintf(os.Stdout, "Response from `RegistryAPI.GetRegistryById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRegistryByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Registry**](Registry.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRepositories

> Repositories GetRepositories(ctx, assetId).SearchName(searchName).Execute()

Get the repositories in the registry.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	searchName := "searchName_example" // string | Filter results that contains searchName. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegistryAPI.GetRepositories(context.Background(), assetId).SearchName(searchName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistryAPI.GetRepositories``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRepositories`: Repositories
	fmt.Fprintf(os.Stdout, "Response from `RegistryAPI.GetRepositories`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRepositoriesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **searchName** | **string** | Filter results that contains searchName. | 

### Return type

[**Repositories**](Repositories.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRepositoryTags

> RepositoryTags GetRepositoryTags(ctx, assetId).Repository(repository).SearchName(searchName).Execute()

Get the repositories tags in the registry.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	repository := "repository_example" // string | The repository name that its tags are requested.
	searchName := "searchName_example" // string | Filter results that contains searchName. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegistryAPI.GetRepositoryTags(context.Background(), assetId).Repository(repository).SearchName(searchName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistryAPI.GetRepositoryTags``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRepositoryTags`: RepositoryTags
	fmt.Fprintf(os.Stdout, "Response from `RegistryAPI.GetRepositoryTags`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRepositoryTagsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **repository** | **string** | The repository name that its tags are requested. | 
 **searchName** | **string** | Filter results that contains searchName. | 

### Return type

[**RepositoryTags**](RepositoryTags.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListRegistries

> RegistryListResponse ListRegistries(ctx).Name(name).Execute()

Get registries.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	name := "name_example" // string | Filter results by name. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegistryAPI.ListRegistries(context.Background()).Name(name).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistryAPI.ListRegistries``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListRegistries`: RegistryListResponse
	fmt.Fprintf(os.Stdout, "Response from `RegistryAPI.ListRegistries`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListRegistriesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 

### Return type

[**RegistryListResponse**](RegistryListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateRegistry

> Registry UpdateRegistry(ctx, assetId).RegistryUpdateRequest(registryUpdateRequest).Execute()

Update a registry asset.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	registryUpdateRequest := *openapiclient.NewRegistryUpdateRequest(*openapiclient.NewAssetUpdateRequest("my-asset")) // RegistryUpdateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegistryAPI.UpdateRegistry(context.Background(), assetId).RegistryUpdateRequest(registryUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistryAPI.UpdateRegistry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateRegistry`: Registry
	fmt.Fprintf(os.Stdout, "Response from `RegistryAPI.UpdateRegistry`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateRegistryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **registryUpdateRequest** | [**RegistryUpdateRequest**](RegistryUpdateRequest.md) |  | 

### Return type

[**Registry**](Registry.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

