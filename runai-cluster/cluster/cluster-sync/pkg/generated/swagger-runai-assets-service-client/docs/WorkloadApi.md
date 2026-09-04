# \WorkloadApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AssetNameConflictValidation**](WorkloadApi.md#AssetNameConflictValidation) | **Get** /api/v1/asset/name-availability | Asset name conflict validation
[**CountWorkloads**](WorkloadApi.md#CountWorkloads) | **Get** /api/v1/workload/count | Count workloads.



## AssetNameConflictValidation

> AssetNameConflictValidation(ctx).Kind(kind).AssetName(assetName).ScopeType(scopeType).ScopeId(scopeId).Execute()

Asset name conflict validation



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
    kind := "kind_example" // string | Filter results by asset kind.
    assetName := "assetName_example" // string | the name of the asset
    scopeType := "scopeType_example" // string | Filter results by scope.
    scopeId := "scopeId_example" // string | The id of the cluster, department or project, depending on the scope type. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadApi.AssetNameConflictValidation(context.Background()).Kind(kind).AssetName(assetName).ScopeType(scopeType).ScopeId(scopeId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadApi.AssetNameConflictValidation``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAssetNameConflictValidationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kind** | **string** | Filter results by asset kind. | 
 **assetName** | **string** | the name of the asset | 
 **scopeType** | **string** | Filter results by scope. | 
 **scopeId** | **string** | The id of the cluster, department or project, depending on the scope type. | 

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


## CountWorkloads

> AssetCountResponse CountWorkloads(ctx).Name(name).ProjectId(projectId).Execute()

Count workloads.



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
    projectId := int32(56) // int32 | Filter results by project id. if scope filter is project, only assets from the specific project will be included in the response. otherwise, the response will include both project, department, cluster and tenant assets. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadApi.CountWorkloads(context.Background()).Name(name).ProjectId(projectId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadApi.CountWorkloads``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CountWorkloads`: AssetCountResponse
    fmt.Fprintf(os.Stdout, "Response from `WorkloadApi.CountWorkloads`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountWorkloadsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **projectId** | **int32** | Filter results by project id. if scope filter is project, only assets from the specific project will be included in the response. otherwise, the response will include both project, department, cluster and tenant assets. | 

### Return type

[**AssetCountResponse**](AssetCountResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

