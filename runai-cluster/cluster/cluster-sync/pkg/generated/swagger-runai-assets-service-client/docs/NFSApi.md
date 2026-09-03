# \NFSApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateNfsAsset**](NFSApi.md#CreateNfsAsset) | **Post** /api/v1/asset/datasource/nfs | Create NFS asset.
[**DeleteNfsAssetById**](NFSApi.md#DeleteNfsAssetById) | **Delete** /api/v1/asset/datasource/nfs/{AssetId} | Delete NFS asset.
[**GetNfsAssetById**](NFSApi.md#GetNfsAssetById) | **Get** /api/v1/asset/datasource/nfs/{AssetId} | Get NFS asset.
[**ListNfsAssets**](NFSApi.md#ListNfsAssets) | **Get** /api/v1/asset/datasource/nfs | List NFS assets.
[**UpdateNfsAssetById**](NFSApi.md#UpdateNfsAssetById) | **Put** /api/v1/asset/datasource/nfs/{AssetId} | Update NFS asset.



## CreateNfsAsset

> S3Asset CreateNfsAsset(ctx).NFSCreationRequest(nFSCreationRequest).Execute()

Create NFS asset.



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
    nFSCreationRequest := *openapiclient.NewNFSCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("runai")), *openapiclient.NewNFSAssetSpec("Path_example", "Server_example", "MountPath_example")) // NFSCreationRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NFSApi.CreateNfsAsset(context.Background()).NFSCreationRequest(nFSCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NFSApi.CreateNfsAsset``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateNfsAsset`: S3Asset
    fmt.Fprintf(os.Stdout, "Response from `NFSApi.CreateNfsAsset`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateNfsAssetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **nFSCreationRequest** | [**NFSCreationRequest**](NFSCreationRequest.md) |  | 

### Return type

[**S3Asset**](S3Asset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteNfsAssetById

> HttpResponse DeleteNfsAssetById(ctx, assetId).Execute()

Delete NFS asset.



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
    resp, r, err := apiClient.NFSApi.DeleteNfsAssetById(context.Background(), assetId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NFSApi.DeleteNfsAssetById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteNfsAssetById`: HttpResponse
    fmt.Fprintf(os.Stdout, "Response from `NFSApi.DeleteNfsAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteNfsAssetByIdRequest struct via the builder pattern


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


## GetNfsAssetById

> NFSAsset GetNfsAssetById(ctx, assetId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).Execute()

Get NFS asset.



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
    resp, r, err := apiClient.NFSApi.GetNfsAssetById(context.Background(), assetId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NFSApi.GetNfsAssetById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetNfsAssetById`: NFSAsset
    fmt.Fprintf(os.Stdout, "Response from `NFSApi.GetNfsAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNfsAssetByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **complyToProject** | **int32** | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **complyToWorkloadType** | **string** | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. | 

### Return type

[**NFSAsset**](NFSAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListNfsAssets

> NFSListResponse ListNfsAssets(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).Execute()

List NFS assets.



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
    resp, r, err := apiClient.NFSApi.ListNfsAssets(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NFSApi.ListNfsAssets``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListNfsAssets`: NFSListResponse
    fmt.Fprintf(os.Stdout, "Response from `NFSApi.ListNfsAssets`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListNfsAssetsRequest struct via the builder pattern


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

[**NFSListResponse**](NFSListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateNfsAssetById

> NFSAsset UpdateNfsAssetById(ctx, assetId).NFSUpdateRequest(nFSUpdateRequest).Execute()

Update NFS asset.



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
    nFSUpdateRequest := *openapiclient.NewNFSUpdateRequest(*openapiclient.NewAssetUpdateRequest("my-asset"), *openapiclient.NewNFSAssetSpec("Path_example", "Server_example", "MountPath_example")) // NFSUpdateRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NFSApi.UpdateNfsAssetById(context.Background(), assetId).NFSUpdateRequest(nFSUpdateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NFSApi.UpdateNfsAssetById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateNfsAssetById`: NFSAsset
    fmt.Fprintf(os.Stdout, "Response from `NFSApi.UpdateNfsAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNfsAssetByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nFSUpdateRequest** | [**NFSUpdateRequest**](NFSUpdateRequest.md) |  | 

### Return type

[**NFSAsset**](NFSAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

