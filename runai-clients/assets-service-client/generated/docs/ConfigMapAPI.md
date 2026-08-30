# \ConfigMapAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateConfigMapAsset**](ConfigMapAPI.md#CreateConfigMapAsset) | **Post** /api/v1/asset/datasource/config-map | Create a ConfigMap datasource asset.
[**DeleteConfigMapAssetById**](ConfigMapAPI.md#DeleteConfigMapAssetById) | **Delete** /api/v1/asset/datasource/config-map/{AssetId} | Delete a ConfigMap asset.
[**GetConfigMapAssetById**](ConfigMapAPI.md#GetConfigMapAssetById) | **Get** /api/v1/asset/datasource/config-map/{AssetId} | Get a ConfigMap asset.
[**ListConfigMapAssets**](ConfigMapAPI.md#ListConfigMapAssets) | **Get** /api/v1/asset/datasource/config-map | List ConfigMap datasource assets.
[**UpdateConfigMapAssetById**](ConfigMapAPI.md#UpdateConfigMapAssetById) | **Put** /api/v1/asset/datasource/config-map/{AssetId} | Update a ConfigMap asset.



## CreateConfigMapAsset

> ConfigMapAsset CreateConfigMapAsset(ctx).ConfigMapCreationRequest(configMapCreationRequest).Execute()

Create a ConfigMap datasource asset.



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
	configMapCreationRequest := *openapiclient.NewConfigMapCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("system")), *openapiclient.NewConfigMapAssetSpec()) // ConfigMapCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ConfigMapAPI.CreateConfigMapAsset(context.Background()).ConfigMapCreationRequest(configMapCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapAPI.CreateConfigMapAsset``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateConfigMapAsset`: ConfigMapAsset
	fmt.Fprintf(os.Stdout, "Response from `ConfigMapAPI.CreateConfigMapAsset`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateConfigMapAssetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **configMapCreationRequest** | [**ConfigMapCreationRequest**](ConfigMapCreationRequest.md) |  | 

### Return type

[**ConfigMapAsset**](ConfigMapAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteConfigMapAssetById

> HttpResponse DeleteConfigMapAssetById(ctx, assetId).Execute()

Delete a ConfigMap asset.



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
	resp, r, err := apiClient.ConfigMapAPI.DeleteConfigMapAssetById(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapAPI.DeleteConfigMapAssetById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteConfigMapAssetById`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `ConfigMapAPI.DeleteConfigMapAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteConfigMapAssetByIdRequest struct via the builder pattern


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


## GetConfigMapAssetById

> ConfigMapAsset GetConfigMapAssetById(ctx, assetId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).StatusInfo(statusInfo).ComplyToReplicaType(complyToReplicaType).Execute()

Get a ConfigMap asset.



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
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	complyToProject := int32(56) // int32 | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
	complyToWorkloadType := "complyToWorkloadType_example" // string | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)
	complyToReplicaType := "complyToReplicaType_example" // string | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ConfigMapAPI.GetConfigMapAssetById(context.Background(), assetId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).StatusInfo(statusInfo).ComplyToReplicaType(complyToReplicaType).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapAPI.GetConfigMapAssetById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetConfigMapAssetById`: ConfigMapAsset
	fmt.Fprintf(os.Stdout, "Response from `ConfigMapAPI.GetConfigMapAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetConfigMapAssetByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **complyToProject** | **int32** | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **complyToWorkloadType** | **string** | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 
 **complyToReplicaType** | **string** | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. | 

### Return type

[**ConfigMapAsset**](ConfigMapAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListConfigMapAssets

> ConfigMapListResponse ListConfigMapAssets(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).StatusInfo(statusInfo).AssetIds(assetIds).ComplyToReplicaType(complyToReplicaType).Execute()

List ConfigMap datasource assets.



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
	scope := "scope_example" // string | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. (optional)
	projectId := int32(56) // int32 | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. (optional)
	departmentId := "1" // string | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. (optional)
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	complyToProject := int32(56) // int32 | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
	complyToWorkloadType := "complyToWorkloadType_example" // string | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)
	assetIds := "dbf4767e-2fa1-43b0-97a2-7c0cecda180b,550e8400-e29b-41d4-a716-44665544000a" // string | Filter results by the ids of the assets. Provided value should be a comma separated string of UUIDs. (optional)
	complyToReplicaType := "complyToReplicaType_example" // string | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ConfigMapAPI.ListConfigMapAssets(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).StatusInfo(statusInfo).AssetIds(assetIds).ComplyToReplicaType(complyToReplicaType).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapAPI.ListConfigMapAssets``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListConfigMapAssets`: ConfigMapListResponse
	fmt.Fprintf(os.Stdout, "Response from `ConfigMapAPI.ListConfigMapAssets`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListConfigMapAssetsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. | 
 **projectId** | **int32** | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. | 
 **departmentId** | **string** | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. | 
 **clusterId** | **string** | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. | 
 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **complyToProject** | **int32** | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **complyToWorkloadType** | **string** | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 
 **assetIds** | **string** | Filter results by the ids of the assets. Provided value should be a comma separated string of UUIDs. | 
 **complyToReplicaType** | **string** | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. | 

### Return type

[**ConfigMapListResponse**](ConfigMapListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateConfigMapAssetById

> ConfigMapAsset UpdateConfigMapAssetById(ctx, assetId).ConfigMapUpdateRequest(configMapUpdateRequest).Execute()

Update a ConfigMap asset.



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
	configMapUpdateRequest := *openapiclient.NewConfigMapUpdateRequest(*openapiclient.NewAssetUpdateRequest("my-asset")) // ConfigMapUpdateRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ConfigMapAPI.UpdateConfigMapAssetById(context.Background(), assetId).ConfigMapUpdateRequest(configMapUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapAPI.UpdateConfigMapAssetById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateConfigMapAssetById`: ConfigMapAsset
	fmt.Fprintf(os.Stdout, "Response from `ConfigMapAPI.UpdateConfigMapAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateConfigMapAssetByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **configMapUpdateRequest** | [**ConfigMapUpdateRequest**](ConfigMapUpdateRequest.md) |  | 

### Return type

[**ConfigMapAsset**](ConfigMapAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

