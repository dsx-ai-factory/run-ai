# \ImagesAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetImageTags**](ImagesAPI.md#GetImageTags) | **Get** /api/v1/registries/{RegistryIntegrationId}/images/tags | Get the image&#39;s tags
[**ListImages**](ImagesAPI.md#ListImages) | **Get** /api/v1/registries/{RegistryIntegrationId}/images | List images by registry-integration id



## GetImageTags

> ImageTags GetImageTags(ctx, registryIntegrationId).ImageName(imageName).AssetId(assetId).Execute()

Get the image's tags

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
	imageName := "runai/viper-catalog-service:latest" // string | The name of the image to get the tags for
	assetId := "assetId_example" // string | The asset ID to use for accessing resources that require API key authentication (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ImagesAPI.GetImageTags(context.Background(), registryIntegrationId).ImageName(imageName).AssetId(assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ImagesAPI.GetImageTags``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetImageTags`: ImageTags
	fmt.Fprintf(os.Stdout, "Response from `ImagesAPI.GetImageTags`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**registryIntegrationId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetImageTagsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **imageName** | **string** | The name of the image to get the tags for | 
 **assetId** | **string** | The asset ID to use for accessing resources that require API key authentication | 

### Return type

[**ImageTags**](ImageTags.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListImages

> ImagesListResponse ListImages(ctx, registryIntegrationId).Search(search).AssetId(assetId).Execute()

List images by registry-integration id



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
	search := "test project" // string | Filter results by a free text search. (optional)
	assetId := "assetId_example" // string | The asset ID to use for accessing resources that require API key authentication (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ImagesAPI.ListImages(context.Background(), registryIntegrationId).Search(search).AssetId(assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ImagesAPI.ListImages``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListImages`: ImagesListResponse
	fmt.Fprintf(os.Stdout, "Response from `ImagesAPI.ListImages`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**registryIntegrationId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiListImagesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **search** | **string** | Filter results by a free text search. | 
 **assetId** | **string** | The asset ID to use for accessing resources that require API key authentication | 

### Return type

[**ImagesListResponse**](ImagesListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

