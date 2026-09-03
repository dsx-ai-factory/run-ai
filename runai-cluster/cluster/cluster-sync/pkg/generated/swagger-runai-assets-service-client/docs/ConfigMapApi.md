# \ConfigMapApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateConfigMapAsset**](ConfigMapApi.md#CreateConfigMapAsset) | **Post** /api/v1/asset/datasource/config-map | Create ConfigMap asset.
[**DeleteConfigMapAssetById**](ConfigMapApi.md#DeleteConfigMapAssetById) | **Delete** /api/v1/asset/datasource/config-map/{AssetId} | Delete ConfigMap asset.
[**GetConfigMapAssetById**](ConfigMapApi.md#GetConfigMapAssetById) | **Get** /api/v1/asset/datasource/config-map/{AssetId} | Get ConfigMap asset.
[**ListConfigMapAssets**](ConfigMapApi.md#ListConfigMapAssets) | **Get** /api/v1/asset/datasource/config-map | List ConfigMap datasource assets.
[**UpdateConfigMapAssetById**](ConfigMapApi.md#UpdateConfigMapAssetById) | **Put** /api/v1/asset/datasource/config-map/{AssetId} | Update ConfigMap asset.



## CreateConfigMapAsset

> ConfigMapAsset CreateConfigMapAsset(ctx).ConfigMapCreationRequest(configMapCreationRequest).Execute()

Create ConfigMap asset.



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
    configMapCreationRequest := *openapiclient.NewConfigMapCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("runai")), *openapiclient.NewConfigMapAssetSpec("ConfigMap_example", "MountPath_example")) // ConfigMapCreationRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConfigMapApi.CreateConfigMapAsset(context.Background()).ConfigMapCreationRequest(configMapCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapApi.CreateConfigMapAsset``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateConfigMapAsset`: ConfigMapAsset
    fmt.Fprintf(os.Stdout, "Response from `ConfigMapApi.CreateConfigMapAsset`: %v\n", resp)
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

Delete ConfigMap asset.



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
    assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConfigMapApi.DeleteConfigMapAssetById(context.Background(), assetId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapApi.DeleteConfigMapAssetById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteConfigMapAssetById`: HttpResponse
    fmt.Fprintf(os.Stdout, "Response from `ConfigMapApi.DeleteConfigMapAssetById`: %v\n", resp)
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

> ConfigMapAsset GetConfigMapAssetById(ctx, assetId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).Execute()

Get ConfigMap asset.



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
    assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
    usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
    complyToProject := int32(56) // int32 | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
    complyToWorkloadType := "complyToWorkloadType_example" // string | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConfigMapApi.GetConfigMapAssetById(context.Background(), assetId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapApi.GetConfigMapAssetById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetConfigMapAssetById`: ConfigMapAsset
    fmt.Fprintf(os.Stdout, "Response from `ConfigMapApi.GetConfigMapAssetById`: %v\n", resp)
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

> ConfigMapListResponse ListConfigMapAssets(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).Execute()

List ConfigMap datasource assets.



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
    name := "name_example" // string | Filter results by name. (optional)
    scope := "scope_example" // string | Filter results by scope. (optional)
    projectId := int32(56) // int32 | Filter results by project id. if scope filter is project, only assets from the specific project will be included in the response. otherwise, the response will include both project, department, cluster and tenant assets. (optional)
    departmentId := "departmentId_example" // string | Filter results by department id. if scope filter is department, only assets from the specific department will be included in the response. otherwise, the response will include department, cluster and tenant scope assets. (optional)
    clusterId := "clusterId_example" // string | Filter results by cluster id. if scope filter is cluster, only assets from the specific cluster will be included in the response. otherwise, the response will include cluster and tenant scope assets. (optional)
    usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
    complyToProject := int32(56) // int32 | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
    complyToWorkloadType := "complyToWorkloadType_example" // string | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConfigMapApi.ListConfigMapAssets(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapApi.ListConfigMapAssets``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListConfigMapAssets`: ConfigMapListResponse
    fmt.Fprintf(os.Stdout, "Response from `ConfigMapApi.ListConfigMapAssets`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListConfigMapAssetsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filter results by scope. | 
 **projectId** | **int32** | Filter results by project id. if scope filter is project, only assets from the specific project will be included in the response. otherwise, the response will include both project, department, cluster and tenant assets. | 
 **departmentId** | **string** | Filter results by department id. if scope filter is department, only assets from the specific department will be included in the response. otherwise, the response will include department, cluster and tenant scope assets. | 
 **clusterId** | **string** | Filter results by cluster id. if scope filter is cluster, only assets from the specific cluster will be included in the response. otherwise, the response will include cluster and tenant scope assets. | 
 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **complyToProject** | **int32** | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **complyToWorkloadType** | **string** | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. | 

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

Update ConfigMap asset.



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
    assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
    configMapUpdateRequest := *openapiclient.NewConfigMapUpdateRequest(*openapiclient.NewAssetUpdateRequest("my-asset"), *openapiclient.NewConfigMapAssetSpec("ConfigMap_example", "MountPath_example")) // ConfigMapUpdateRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConfigMapApi.UpdateConfigMapAssetById(context.Background(), assetId).ConfigMapUpdateRequest(configMapUpdateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapApi.UpdateConfigMapAssetById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateConfigMapAssetById`: ConfigMapAsset
    fmt.Fprintf(os.Stdout, "Response from `ConfigMapApi.UpdateConfigMapAssetById`: %v\n", resp)
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

