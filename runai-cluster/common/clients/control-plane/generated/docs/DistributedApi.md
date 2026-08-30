# \DistributedApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateDistributed**](DistributedApi.md#CreateDistributed) | **Post** /api/v1/workloads/distributed | Create a distributed training. [Experimental]
[**DeleteDistributed**](DistributedApi.md#DeleteDistributed) | **Delete** /api/v1/workloads/distributed/{workloadId} | Delete a distributed training by id. [Experimental]
[**GetDistributed**](DistributedApi.md#GetDistributed) | **Get** /api/v1/workloads/distributed/{workloadId} | Get distributed training&#39;s data. [Experimental]
[**ListDistributedUpdates**](DistributedApi.md#ListDistributedUpdates) | **Get** /api/v1/workloads/distributed/updates | Get distributed training updates.
[**RequestDistributedUpdates**](DistributedApi.md#RequestDistributedUpdates) | **Post** /api/v1/workloads/distributed/updates | Request distributed training updates



## CreateDistributed

> DistributedWorkload CreateDistributed(ctx).DistributedCreationRequest(distributedCreationRequest).Execute()

Create a distributed training. [Experimental]



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
    distributedCreationRequest := *openapiclient.NewDistributedCreationRequest("my-workload-name", "1", "71f69d83-ba66-4822-adf5-55ce55efd210") // DistributedCreationRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DistributedApi.CreateDistributed(context.Background()).DistributedCreationRequest(distributedCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DistributedApi.CreateDistributed``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateDistributed`: DistributedWorkload
    fmt.Fprintf(os.Stdout, "Response from `DistributedApi.CreateDistributed`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateDistributedRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **distributedCreationRequest** | [**DistributedCreationRequest**](DistributedCreationRequest.md) |  | 

### Return type

[**DistributedWorkload**](DistributedWorkload.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteDistributed

> DeleteDistributed(ctx, workloadId).Execute()

Delete a distributed training by id. [Experimental]



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
    workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the workload.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DistributedApi.DeleteDistributed(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DistributedApi.DeleteDistributed``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDistributedRequest struct via the builder pattern


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


## GetDistributed

> DistributedWorkload GetDistributed(ctx, workloadId).Execute()

Get distributed training's data. [Experimental]



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
    workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the workload.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DistributedApi.GetDistributed(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DistributedApi.GetDistributed``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDistributed`: DistributedWorkload
    fmt.Fprintf(os.Stdout, "Response from `DistributedApi.GetDistributed`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDistributedRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DistributedWorkload**](DistributedWorkload.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListDistributedUpdates

> WorkloadsSyncResponse ListDistributedUpdates(ctx).UpdatesAfter(updatesAfter).ClusterId(clusterId).Execute()

Get distributed training updates.



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
    updatesAfter := int64(12345) // int64 | list only data with update counter greater than the provided value. for example updatesAfter=5 should fetch data with updates counter 6 and above.
    clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DistributedApi.ListDistributedUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DistributedApi.ListDistributedUpdates``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListDistributedUpdates`: WorkloadsSyncResponse
    fmt.Fprintf(os.Stdout, "Response from `DistributedApi.ListDistributedUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListDistributedUpdatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updatesAfter** | **int64** | list only data with update counter greater than the provided value. for example updatesAfter&#x3D;5 should fetch data with updates counter 6 and above. | 
 **clusterId** | **string** | The id of the cluster | 

### Return type

[**WorkloadsSyncResponse**](WorkloadsSyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RequestDistributedUpdates

> WorkloadsSyncResponse RequestDistributedUpdates(ctx).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()

Request distributed training updates



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
    updatesAfter := int64(12345) // int64 | list only data with update counter greater than the provided value. for example updatesAfter=5 should fetch data with updates counter 6 and above.
    clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
    clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DistributedApi.RequestDistributedUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DistributedApi.RequestDistributedUpdates``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `RequestDistributedUpdates`: WorkloadsSyncResponse
    fmt.Fprintf(os.Stdout, "Response from `DistributedApi.RequestDistributedUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRequestDistributedUpdatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updatesAfter** | **int64** | list only data with update counter greater than the provided value. for example updatesAfter&#x3D;5 should fetch data with updates counter 6 and above. | 
 **clusterId** | **string** | The id of the cluster | 
 **clusterSyncReport** | [**ClusterSyncReport**](ClusterSyncReport.md) |  | 

### Return type

[**WorkloadsSyncResponse**](WorkloadsSyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

