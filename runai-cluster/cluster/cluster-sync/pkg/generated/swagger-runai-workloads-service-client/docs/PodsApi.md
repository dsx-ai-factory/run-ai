# \PodsApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeletePod**](PodsApi.md#DeletePod) | **Delete** /api/v1/workloads/pods/{podId} | Delete a pod as it does not exist in the cluster anymore.
[**GetWorkloadPods**](PodsApi.md#GetWorkloadPods) | **Get** /api/v1/workloads/{workloadId}/pods | Get the workload pods. [Experimental]
[**SyncWorkloadsPods**](PodsApi.md#SyncWorkloadsPods) | **Post** /api/v1/workloads/pods | Sync workloads pods.
[**UpsertPod**](PodsApi.md#UpsertPod) | **Put** /api/v1/workloads/pods/{podId} | Upsert a single pod info from the cluster.



## DeletePod

> DeletePod(ctx, podId).Execute()

Delete a pod as it does not exist in the cluster anymore.

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
    podId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The requested pod id.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.DeletePod(context.Background(), podId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.DeletePod``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**podId** | **string** | The requested pod id. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeletePodRequest struct via the builder pattern


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


## GetWorkloadPods

> GetWorkloadPods200Response GetWorkloadPods(ctx, workloadId).Deleted(deleted).Offset(offset).Limit(limit).Execute()

Get the workload pods. [Experimental]

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
    workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The requested workload id.
    deleted := true // bool | if true, return only deleted resources. (optional)
    offset := int32(56) // int32 | The offset from which to start fetching records. (optional)
    limit := int32(56) // int32 | The limit of returned records count. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.GetWorkloadPods(context.Background(), workloadId).Deleted(deleted).Offset(offset).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.GetWorkloadPods``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkloadPods`: GetWorkloadPods200Response
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.GetWorkloadPods`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The requested workload id. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadPodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deleted** | **bool** | if true, return only deleted resources. | 
 **offset** | **int32** | The offset from which to start fetching records. | 
 **limit** | **int32** | The limit of returned records count. | 

### Return type

[**GetWorkloadPods200Response**](GetWorkloadPods200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncWorkloadsPods

> SyncWorkloadsPods(ctx).PodsSyncRequest(podsSyncRequest).Execute()

Sync workloads pods.

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    openapiclient "./openapi"
)

func main() {
    podsSyncRequest := *openapiclient.NewPodsSyncRequest("1", []openapiclient.PodSyncFields{*openapiclient.NewPodSyncFields("pod-of-a-very-important-job", "high-priority", "Id_example", "WorkloadId_example", "ClusterId_example", "1", time.Now(), []openapiclient.Container{*openapiclient.NewContainer("busybox", "busybox:latest")}, "Pending")}) // PodsSyncRequest | The pods to sync. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.SyncWorkloadsPods(context.Background()).PodsSyncRequest(podsSyncRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.SyncWorkloadsPods``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncWorkloadsPodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **podsSyncRequest** | [**PodsSyncRequest**](PodsSyncRequest.md) | The pods to sync. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpsertPod

> UpsertPod(ctx, podId).PodSyncFields(podSyncFields).Execute()

Upsert a single pod info from the cluster.

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    openapiclient "./openapi"
)

func main() {
    podId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The requested pod id.
    podSyncFields := *openapiclient.NewPodSyncFields("pod-of-a-very-important-job", "high-priority", "Id_example", "WorkloadId_example", "ClusterId_example", "1", time.Now(), []openapiclient.Container{*openapiclient.NewContainer("busybox", "busybox:latest")}, "Pending") // PodSyncFields | The pod info in the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.UpsertPod(context.Background(), podId).PodSyncFields(podSyncFields).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.UpsertPod``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**podId** | **string** | The requested pod id. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpsertPodRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **podSyncFields** | [**PodSyncFields**](PodSyncFields.md) | The pod info in the cluster. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

