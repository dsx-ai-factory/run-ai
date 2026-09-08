# \HelmReposAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**BrowseHelmRepoCharts**](HelmReposAPI.md#BrowseHelmRepoCharts) | **Get** /api/v1/helm-repos/{HelmRepoId}/charts | Browse helm repo charts. [Experimental]
[**CreateHelmRepo**](HelmReposAPI.md#CreateHelmRepo) | **Post** /api/v1/helm-repos | Create helm repo. [Experimental]
[**DeleteHelmRepo**](HelmReposAPI.md#DeleteHelmRepo) | **Delete** /api/v1/helm-repos/{HelmRepoId} | Delete helm repo. [Experimental]
[**GetHelmRepoById**](HelmReposAPI.md#GetHelmRepoById) | **Get** /api/v1/helm-repos/{HelmRepoId} | Get helm repo by id. [Experimental]
[**ListHelmRepos**](HelmReposAPI.md#ListHelmRepos) | **Get** /api/v1/helm-repos | List helm repos. [Experimental]
[**UpdateHelmRepo**](HelmReposAPI.md#UpdateHelmRepo) | **Put** /api/v1/helm-repos/{HelmRepoId} | Update helm repo. [Experimental]



## BrowseHelmRepoCharts

> HelmChartListResponse BrowseHelmRepoCharts(ctx, helmRepoId).AssetId(assetId).Execute()

Browse helm repo charts. [Experimental]



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
	helmRepoId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	assetId := "assetId_example" // string | The asset ID to use for accessing resources that require API key authentication (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.HelmReposAPI.BrowseHelmRepoCharts(context.Background(), helmRepoId).AssetId(assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `HelmReposAPI.BrowseHelmRepoCharts``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BrowseHelmRepoCharts`: HelmChartListResponse
	fmt.Fprintf(os.Stdout, "Response from `HelmReposAPI.BrowseHelmRepoCharts`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**helmRepoId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiBrowseHelmRepoChartsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **assetId** | **string** | The asset ID to use for accessing resources that require API key authentication | 

### Return type

[**HelmChartListResponse**](HelmChartListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateHelmRepo

> HelmRepo CreateHelmRepo(ctx).HelmRepoCreationRequest(helmRepoCreationRequest).Execute()

Create helm repo. [Experimental]



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
	helmRepoCreationRequest := *openapiclient.NewHelmRepoCreationRequest("Name_example", "Url_example", openapiclient.ScopeType("system"), "ScopeId_example") // HelmRepoCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.HelmReposAPI.CreateHelmRepo(context.Background()).HelmRepoCreationRequest(helmRepoCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `HelmReposAPI.CreateHelmRepo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateHelmRepo`: HelmRepo
	fmt.Fprintf(os.Stdout, "Response from `HelmReposAPI.CreateHelmRepo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateHelmRepoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **helmRepoCreationRequest** | [**HelmRepoCreationRequest**](HelmRepoCreationRequest.md) |  | 

### Return type

[**HelmRepo**](HelmRepo.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteHelmRepo

> HttpResponse DeleteHelmRepo(ctx, helmRepoId).Execute()

Delete helm repo. [Experimental]



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
	helmRepoId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.HelmReposAPI.DeleteHelmRepo(context.Background(), helmRepoId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `HelmReposAPI.DeleteHelmRepo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteHelmRepo`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `HelmReposAPI.DeleteHelmRepo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**helmRepoId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteHelmRepoRequest struct via the builder pattern


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


## GetHelmRepoById

> HelmRepo GetHelmRepoById(ctx, helmRepoId).Execute()

Get helm repo by id. [Experimental]



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
	helmRepoId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.HelmReposAPI.GetHelmRepoById(context.Background(), helmRepoId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `HelmReposAPI.GetHelmRepoById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetHelmRepoById`: HelmRepo
	fmt.Fprintf(os.Stdout, "Response from `HelmReposAPI.GetHelmRepoById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**helmRepoId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetHelmRepoByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HelmRepo**](HelmRepo.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListHelmRepos

> HelmRepoListResponse ListHelmRepos(ctx).Execute()

List helm repos. [Experimental]



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
	resp, r, err := apiClient.HelmReposAPI.ListHelmRepos(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `HelmReposAPI.ListHelmRepos``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListHelmRepos`: HelmRepoListResponse
	fmt.Fprintf(os.Stdout, "Response from `HelmReposAPI.ListHelmRepos`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListHelmReposRequest struct via the builder pattern


### Return type

[**HelmRepoListResponse**](HelmRepoListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateHelmRepo

> HelmRepo UpdateHelmRepo(ctx, helmRepoId).HelmRepoUpdateRequest(helmRepoUpdateRequest).Execute()

Update helm repo. [Experimental]



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
	helmRepoId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	helmRepoUpdateRequest := *openapiclient.NewHelmRepoUpdateRequest() // HelmRepoUpdateRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.HelmReposAPI.UpdateHelmRepo(context.Background(), helmRepoId).HelmRepoUpdateRequest(helmRepoUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `HelmReposAPI.UpdateHelmRepo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateHelmRepo`: HelmRepo
	fmt.Fprintf(os.Stdout, "Response from `HelmReposAPI.UpdateHelmRepo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**helmRepoId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateHelmRepoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **helmRepoUpdateRequest** | [**HelmRepoUpdateRequest**](HelmRepoUpdateRequest.md) |  | 

### Return type

[**HelmRepo**](HelmRepo.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

