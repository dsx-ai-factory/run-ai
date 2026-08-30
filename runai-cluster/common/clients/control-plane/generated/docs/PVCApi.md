# \PVCApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreatePvcAsset**](PVCApi.md#CreatePvcAsset) | **Post** /api/v1/asset/datasource/pvc | Create a PVC asset.
[**DeletePvcAssetById**](PVCApi.md#DeletePvcAssetById) | **Delete** /api/v1/asset/datasource/pvc/{AssetId} | Delete a PVC asset.
[**GetPvcAssetById**](PVCApi.md#GetPvcAssetById) | **Get** /api/v1/asset/datasource/pvc/{AssetId} | Get a PVC asset.
[**ListPvcAssets**](PVCApi.md#ListPvcAssets) | **Get** /api/v1/asset/datasource/pvc | List PVC assets.
[**UpdatePvcAssetById**](PVCApi.md#UpdatePvcAssetById) | **Put** /api/v1/asset/datasource/pvc/{AssetId} | Update a PVC asset.



## CreatePvcAsset

> PVCAsset CreatePvcAsset(ctx).PVCCreationRequest(pVCCreationRequest).Execute()

Create a PVC asset.



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
    pVCCreationRequest := *openapiclient.NewPVCCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("system")), *openapiclient.NewPVCAssetSpec()) // PVCCreationRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCApi.CreatePvcAsset(context.Background()).PVCCreationRequest(pVCCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCApi.CreatePvcAsset``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreatePvcAsset`: PVCAsset
    fmt.Fprintf(os.Stdout, "Response from `PVCApi.CreatePvcAsset`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreatePvcAssetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **pVCCreationRequest** | [**PVCCreationRequest**](PVCCreationRequest.md) |  | 

### Return type

[**PVCAsset**](PVCAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeletePvcAssetById

> HttpResponse1 DeletePvcAssetById(ctx, assetId).Execute()

Delete a PVC asset.



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
    resp, r, err := apiClient.PVCApi.DeletePvcAssetById(context.Background(), assetId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCApi.DeletePvcAssetById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeletePvcAssetById`: HttpResponse1
    fmt.Fprintf(os.Stdout, "Response from `PVCApi.DeletePvcAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeletePvcAssetByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HttpResponse1**](HttpResponse1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPvcAssetById

> PVCAsset GetPvcAssetById(ctx, assetId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).StatusInfo(statusInfo).ComplyToReplicaType(complyToReplicaType).Execute()

Get a PVC asset.



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
    statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)
    complyToReplicaType := "complyToReplicaType_example" // string | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCApi.GetPvcAssetById(context.Background(), assetId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).StatusInfo(statusInfo).ComplyToReplicaType(complyToReplicaType).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCApi.GetPvcAssetById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPvcAssetById`: PVCAsset
    fmt.Fprintf(os.Stdout, "Response from `PVCApi.GetPvcAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPvcAssetByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **complyToProject** | **int32** | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **complyToWorkloadType** | **string** | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 
 **complyToReplicaType** | **string** | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. | 

### Return type

[**PVCAsset**](PVCAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPvcAssets

> PVCListResponse ListPvcAssets(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).StatusInfo(statusInfo).AssetIds(assetIds).ComplyToReplicaType(complyToReplicaType).Execute()

List PVC assets.



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
    departmentId := "1" // string | Filter using the department id. (optional)
    clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
    usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
    complyToProject := int32(56) // int32 | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
    complyToWorkloadType := "complyToWorkloadType_example" // string | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
    statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)
    assetIds := "dbf4767e-2fa1-43b0-97a2-7c0cecda180b,550e8400-e29b-41d4-a716-44665544000a" // string | Filter results by the ids of the assets. Provided value should be a comma separated string of UUIDs. (optional)
    complyToReplicaType := "complyToReplicaType_example" // string | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCApi.ListPvcAssets(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).StatusInfo(statusInfo).AssetIds(assetIds).ComplyToReplicaType(complyToReplicaType).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCApi.ListPvcAssets``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListPvcAssets`: PVCListResponse
    fmt.Fprintf(os.Stdout, "Response from `PVCApi.ListPvcAssets`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListPvcAssetsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filter results by scope. | 
 **projectId** | **int32** | Filter results by project id. if scope filter is project, only assets from the specific project will be included in the response. otherwise, the response will include both project, department, cluster and tenant assets. | 
 **departmentId** | **string** | Filter using the department id. | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **complyToProject** | **int32** | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **complyToWorkloadType** | **string** | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 
 **assetIds** | **string** | Filter results by the ids of the assets. Provided value should be a comma separated string of UUIDs. | 
 **complyToReplicaType** | **string** | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. | 

### Return type

[**PVCListResponse**](PVCListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdatePvcAssetById

> PVCAsset UpdatePvcAssetById(ctx, assetId).PVCUpdateRequest(pVCUpdateRequest).Execute()

Update a PVC asset.



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
    pVCUpdateRequest := *openapiclient.NewPVCUpdateRequest(*openapiclient.NewAssetUpdateRequest("my-asset"), *openapiclient.NewPVCAssetSpec()) // PVCUpdateRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCApi.UpdatePvcAssetById(context.Background(), assetId).PVCUpdateRequest(pVCUpdateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCApi.UpdatePvcAssetById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdatePvcAssetById`: PVCAsset
    fmt.Fprintf(os.Stdout, "Response from `PVCApi.UpdatePvcAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdatePvcAssetByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **pVCUpdateRequest** | [**PVCUpdateRequest**](PVCUpdateRequest.md) |  | 

### Return type

[**PVCAsset**](PVCAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

