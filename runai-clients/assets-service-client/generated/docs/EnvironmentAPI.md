# \EnvironmentAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateEnvironmentAsset**](EnvironmentAPI.md#CreateEnvironmentAsset) | **Post** /api/v1/asset/environment | Create an environment asset.
[**DeleteEnvironmentAssetById**](EnvironmentAPI.md#DeleteEnvironmentAssetById) | **Delete** /api/v1/asset/environment/{AssetId} | Delete an environment asset.
[**GetEnvironmentAssetById**](EnvironmentAPI.md#GetEnvironmentAssetById) | **Get** /api/v1/asset/environment/{AssetId} | Get an environment asset.
[**ListEnvironmentAssets**](EnvironmentAPI.md#ListEnvironmentAssets) | **Get** /api/v1/asset/environment | List environment assets.
[**UpdateEnvironmentAssetById**](EnvironmentAPI.md#UpdateEnvironmentAssetById) | **Put** /api/v1/asset/environment/{AssetId} | Update an environment asset.



## CreateEnvironmentAsset

> EnvironmentAsset CreateEnvironmentAsset(ctx).EnvironmentCreationRequest(environmentCreationRequest).Execute()

Create an environment asset.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	environmentCreationRequest := *openapiclient.NewEnvironmentCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("system")), *openapiclient.NewEnvironmentAssetSpec()) // EnvironmentCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EnvironmentAPI.CreateEnvironmentAsset(context.Background()).EnvironmentCreationRequest(environmentCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EnvironmentAPI.CreateEnvironmentAsset``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateEnvironmentAsset`: EnvironmentAsset
	fmt.Fprintf(os.Stdout, "Response from `EnvironmentAPI.CreateEnvironmentAsset`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateEnvironmentAssetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **environmentCreationRequest** | [**EnvironmentCreationRequest**](EnvironmentCreationRequest.md) |  | 

### Return type

[**EnvironmentAsset**](EnvironmentAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteEnvironmentAssetById

> HttpResponse DeleteEnvironmentAssetById(ctx, assetId).Execute()

Delete an environment asset.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EnvironmentAPI.DeleteEnvironmentAssetById(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EnvironmentAPI.DeleteEnvironmentAssetById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteEnvironmentAssetById`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `EnvironmentAPI.DeleteEnvironmentAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteEnvironmentAssetByIdRequest struct via the builder pattern


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


## GetEnvironmentAssetById

> EnvironmentAsset GetEnvironmentAssetById(ctx, assetId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).ComplyToReplicaType(complyToReplicaType).StatusInfo(statusInfo).Execute()

Get an environment asset.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	complyToProject := int32(56) // int32 | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
	complyToWorkloadType := "complyToWorkloadType_example" // string | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
	complyToReplicaType := "complyToReplicaType_example" // string | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EnvironmentAPI.GetEnvironmentAssetById(context.Background(), assetId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).ComplyToReplicaType(complyToReplicaType).StatusInfo(statusInfo).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EnvironmentAPI.GetEnvironmentAssetById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetEnvironmentAssetById`: EnvironmentAsset
	fmt.Fprintf(os.Stdout, "Response from `EnvironmentAPI.GetEnvironmentAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetEnvironmentAssetByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **complyToProject** | **int32** | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **complyToWorkloadType** | **string** | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **complyToReplicaType** | **string** | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 

### Return type

[**EnvironmentAsset**](EnvironmentAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListEnvironmentAssets

> EnvironmentListResponse ListEnvironmentAssets(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).DistributedFramework(distributedFramework).IsDistributed(isDistributed).IsTraining(isTraining).IsWorkspace(isWorkspace).IsInference(isInference).ComplyToReplicaType(complyToReplicaType).StatusInfo(statusInfo).IncludeDescendants(includeDescendants).IsDistributedInference(isDistributedInference).Execute()

List environment assets.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
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
	distributedFramework := "distributedFramework_example" // string | Filter results to workload of type distributed and distributedFramework. (optional)
	isDistributed := true // bool | Filter results to workload of type distributed. (optional)
	isTraining := true // bool | Filter results to workload of type training. (optional)
	isWorkspace := true // bool | Filter results to workload of type workspace. (optional)
	isInference := true // bool | Filter results to workload of type inference. (optional)
	complyToReplicaType := "complyToReplicaType_example" // string | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)
	includeDescendants := true // bool | Indicates whether to include assets from all descendant scopes under the specified scope. When set to true, and a specific scope identifier (for example, clusterId) is provided, the query also returns assets from all nested scopes within it (for example, departments and projects).  Mutually exclusive with scope. (optional)
	isDistributedInference := true // bool | Filter results to workload of type distributed inference. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EnvironmentAPI.ListEnvironmentAssets(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).DistributedFramework(distributedFramework).IsDistributed(isDistributed).IsTraining(isTraining).IsWorkspace(isWorkspace).IsInference(isInference).ComplyToReplicaType(complyToReplicaType).StatusInfo(statusInfo).IncludeDescendants(includeDescendants).IsDistributedInference(isDistributedInference).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EnvironmentAPI.ListEnvironmentAssets``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListEnvironmentAssets`: EnvironmentListResponse
	fmt.Fprintf(os.Stdout, "Response from `EnvironmentAPI.ListEnvironmentAssets`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListEnvironmentAssetsRequest struct via the builder pattern


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
 **distributedFramework** | **string** | Filter results to workload of type distributed and distributedFramework. | 
 **isDistributed** | **bool** | Filter results to workload of type distributed. | 
 **isTraining** | **bool** | Filter results to workload of type training. | 
 **isWorkspace** | **bool** | Filter results to workload of type workspace. | 
 **isInference** | **bool** | Filter results to workload of type inference. | 
 **complyToReplicaType** | **string** | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 
 **includeDescendants** | **bool** | Indicates whether to include assets from all descendant scopes under the specified scope. When set to true, and a specific scope identifier (for example, clusterId) is provided, the query also returns assets from all nested scopes within it (for example, departments and projects).  Mutually exclusive with scope. | 
 **isDistributedInference** | **bool** | Filter results to workload of type distributed inference. | 

### Return type

[**EnvironmentListResponse**](EnvironmentListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateEnvironmentAssetById

> EnvironmentAsset UpdateEnvironmentAssetById(ctx, assetId).EnvironmentUpdateRequest(environmentUpdateRequest).Execute()

Update an environment asset.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	environmentUpdateRequest := *openapiclient.NewEnvironmentUpdateRequest(*openapiclient.NewAssetUpdateRequest("my-asset"), *openapiclient.NewEnvironmentAssetSpec()) // EnvironmentUpdateRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EnvironmentAPI.UpdateEnvironmentAssetById(context.Background(), assetId).EnvironmentUpdateRequest(environmentUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EnvironmentAPI.UpdateEnvironmentAssetById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateEnvironmentAssetById`: EnvironmentAsset
	fmt.Fprintf(os.Stdout, "Response from `EnvironmentAPI.UpdateEnvironmentAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateEnvironmentAssetByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **environmentUpdateRequest** | [**EnvironmentUpdateRequest**](EnvironmentUpdateRequest.md) |  | 

### Return type

[**EnvironmentAsset**](EnvironmentAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

