# \PodsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CountPods**](PodsApi.md#CountPods) | **Get** /api/v1/workloads/pods/count | Get pods count.
[**CreatePods**](PodsApi.md#CreatePods) | **Post** /v1/k8s/clusters/{uuid}/pods | create pods
[**DeletePod**](PodsApi.md#DeletePod) | **Delete** /v1/k8s/clusters/{uuid}/pods/{podId} | Delete a pod by id
[**DeletePod_0**](PodsApi.md#DeletePod_0) | **Delete** /api/v1/workloads/pods/{podId} | Delete a pod as it does not exist in the cluster anymore.
[**GetPods**](PodsApi.md#GetPods) | **Get** /v1/k8s/clusters/{uuid}/pods | get all pods from a specific cluster. Deprecated - please use api/v1/workloads/pods instead
[**GetWorkloadPodMetrics**](PodsApi.md#GetWorkloadPodMetrics) | **Get** /api/v1/workloads/{workloadId}/pods/{podId}/metrics | Get pod metrics data. [Experimental]
[**GetWorkloadPods**](PodsApi.md#GetWorkloadPods) | **Get** /api/v1/workloads/{workloadId}/pods | Get workload pods by id.
[**InternalSyncPods**](PodsApi.md#InternalSyncPods) | **Post** /internal/workloads/pods/sync | Sync pods.
[**ListPods**](PodsApi.md#ListPods) | **Get** /api/v1/workloads/pods | List pods.
[**SyncPodsData**](PodsApi.md#SyncPodsData) | **Post** /v1/k8s/clusters/{uuid}/pods/sync | Sync All cluster pods data to DB
[**SyncWorkloadsPods**](PodsApi.md#SyncWorkloadsPods) | **Post** /api/v1/workloads/pods | Sync workloads pods.
[**UpdatePod**](PodsApi.md#UpdatePod) | **Put** /v1/k8s/clusters/{uuid}/pods/{podId} | Update a pod by id
[**UpdatePods**](PodsApi.md#UpdatePods) | **Put** /v1/k8s/clusters/{uuid}/pods | Sync all pods in a namespace to DB
[**UpsertPod**](PodsApi.md#UpsertPod) | **Put** /api/v1/workloads/pods/{podId} | Upsert a single pod info from the cluster.



## CountPods

> CountAccessRules200Response CountPods(ctx).Deleted(deleted).FilterBy(filterBy).Execute()

Get pods count.



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
    deleted := true // bool | Return only deleted resources when `true`. (optional)
    filterBy := []string{"Inner_example"} // []string | Filter results using a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contains, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.CountPods(context.Background()).Deleted(deleted).FilterBy(filterBy).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.CountPods``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CountPods`: CountAccessRules200Response
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.CountPods`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountPodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deleted** | **bool** | Return only deleted resources when &#x60;true&#x60;. | 
 **filterBy** | **[]string** | Filter results using a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contains, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 

### Return type

[**CountAccessRules200Response**](CountAccessRules200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreatePods

> []Pod CreatePods(ctx, uuid).Pod(pod).Execute()

create pods

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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    pod := []openapiclient.Pod{*openapiclient.NewPod("PodId_example", "JobId_example", "ClusterUuid_example", "PodName_example", "ImageName_example", "Phase_example", int64(123), int64(123), int64(123))} // []Pod | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.CreatePods(context.Background(), uuid).Pod(pod).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.CreatePods``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreatePods`: []Pod
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.CreatePods`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreatePodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **pod** | [**[]Pod**](Pod.md) |  | 

### Return type

[**[]Pod**](Pod.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeletePod

> string DeletePod(ctx, uuid, podId).Execute()

Delete a pod by id

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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    podId := "podId_example" // string | Unique identifier of the pod.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.DeletePod(context.Background(), uuid, podId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.DeletePod``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeletePod`: string
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.DeletePod`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**podId** | **string** | Unique identifier of the pod. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeletePodRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeletePod_0

> DeletePod_0(ctx, podId).Execute()

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
    resp, r, err := apiClient.PodsApi.DeletePod_0(context.Background(), podId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.DeletePod_0``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeletePod_1Request struct via the builder pattern


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


## GetPods

> []Pod GetPods(ctx, uuid).Execute()

get all pods from a specific cluster. Deprecated - please use api/v1/workloads/pods instead

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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.GetPods(context.Background(), uuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.GetPods``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPods`: []Pod
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.GetPods`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]Pod**](Pod.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadPodMetrics

> MetricsResponse GetWorkloadPodMetrics(ctx, workloadId, podId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()

Get pod metrics data. [Experimental]



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
    workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the workload.
    podId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The requested pod id.
    metricType := []openapiclient.PodMetricType{openapiclient.PodMetricType("GPU_UTILIZATION_PER_GPU")} // []PodMetricType | Specify which metric data to request. Advanced GPU metrics are only supported if the 'Advanced GPU Metrics' feature flag in the settings is enabled.
    start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
    end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
    numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.GetWorkloadPodMetrics(context.Background(), workloadId, podId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.GetWorkloadPodMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkloadPodMetrics`: MetricsResponse
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.GetWorkloadPodMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 
**podId** | **string** | The requested pod id. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadPodMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metricType** | [**[]PodMetricType**](PodMetricType.md) | Specify which metric data to request. Advanced GPU metrics are only supported if the &#39;Advanced GPU Metrics&#39; feature flag in the settings is enabled. | 
 **start** | **time.Time** | Start date of time range to fetch data in ISO 8601 timestamp format. | 
 **end** | **time.Time** | End date of time range to fetch data in ISO 8601 timestamp format. | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]

### Return type

[**MetricsResponse**](MetricsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, text/csv

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadPods

> ListPods200Response GetWorkloadPods(ctx, workloadId).Deleted(deleted).Offset(offset).Limit(limit).Execute()

Get workload pods by id.



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
    deleted := true // bool | Return only deleted resources when `true`. (optional)
    offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
    limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.GetWorkloadPods(context.Background(), workloadId).Deleted(deleted).Offset(offset).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.GetWorkloadPods``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkloadPods`: ListPods200Response
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.GetWorkloadPods`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadPodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deleted** | **bool** | Return only deleted resources when &#x60;true&#x60;. | 
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**ListPods200Response**](ListPods200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InternalSyncPods

> InternalSyncPods(ctx).InternalPodsSyncRequest(internalPodsSyncRequest).Execute()

Sync pods.

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
    internalPodsSyncRequest := *openapiclient.NewInternalPodsSyncRequest("71f69d83-ba66-4822-adf5-55ce55efd210", []openapiclient.PodSyncFields{*openapiclient.NewPodSyncFields("pod-of-a-very-important-job", "high-priority", "Id_example", "WorkloadId_example", "71f69d83-ba66-4822-adf5-55ce55efd210", time.Now(), []openapiclient.Container1{*openapiclient.NewContainer1("alpine", "alpine:latest")}, "Pending")}) // InternalPodsSyncRequest | The pods to sync. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.InternalSyncPods(context.Background()).InternalPodsSyncRequest(internalPodsSyncRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.InternalSyncPods``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInternalSyncPodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **internalPodsSyncRequest** | [**InternalPodsSyncRequest**](InternalPodsSyncRequest.md) | The pods to sync. | 

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


## ListPods

> ListPods200Response ListPods(ctx).Deleted(deleted).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Verbosity(verbosity).Completed(completed).Execute()

List pods.



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
    deleted := true // bool | Return only deleted resources when `true`. (optional)
    offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
    limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
    sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
    sortBy := "sortBy_example" // string | Sort results using a parameter. (optional)
    filterBy := []string{"Inner_example"} // []string | Filter results using a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contains, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)
    verbosity := openapiclient.PodVerbosity("brief") // PodVerbosity | response verbosity level. if full, the response includes workloadName and projectName fields.  (optional) (default to "brief")
    completed := "completed_example" // string | Return only completed resources when 'true', return only non-completed resources when 'false'. By default, or when empty, returns all resources. (optional) (default to "all")

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.ListPods(context.Background()).Deleted(deleted).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Verbosity(verbosity).Completed(completed).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.ListPods``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListPods`: ListPods200Response
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.ListPods`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListPodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deleted** | **bool** | Return only deleted resources when &#x60;true&#x60;. | 
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **sortBy** | **string** | Sort results using a parameter. | 
 **filterBy** | **[]string** | Filter results using a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contains, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 
 **verbosity** | [**PodVerbosity**](PodVerbosity.md) | response verbosity level. if full, the response includes workloadName and projectName fields.  | [default to &quot;brief&quot;]
 **completed** | **string** | Return only completed resources when &#39;true&#39;, return only non-completed resources when &#39;false&#39;. By default, or when empty, returns all resources. | [default to &quot;all&quot;]

### Return type

[**ListPods200Response**](ListPods200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncPodsData

> SyncPodsData(ctx, uuid).Pod(pod).Execute()

Sync All cluster pods data to DB

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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    pod := []openapiclient.Pod{*openapiclient.NewPod("PodId_example", "JobId_example", "ClusterUuid_example", "PodName_example", "ImageName_example", "Phase_example", int64(123), int64(123), int64(123))} // []Pod | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.SyncPodsData(context.Background(), uuid).Pod(pod).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.SyncPodsData``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncPodsDataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **pod** | [**[]Pod**](Pod.md) |  | 

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
    podsSyncRequest := *openapiclient.NewPodsSyncRequest("1", []openapiclient.PodSyncFields{*openapiclient.NewPodSyncFields("pod-of-a-very-important-job", "high-priority", "Id_example", "WorkloadId_example", "71f69d83-ba66-4822-adf5-55ce55efd210", time.Now(), []openapiclient.Container1{*openapiclient.NewContainer1("alpine", "alpine:latest")}, "Pending")}) // PodsSyncRequest | The pods to sync. (optional)

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


## UpdatePod

> UpdatePod(ctx, uuid, podId).Pod(pod).Execute()

Update a pod by id

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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    podId := "podId_example" // string | Unique identifier of the pod.
    pod := *openapiclient.NewPod("PodId_example", "JobId_example", "ClusterUuid_example", "PodName_example", "ImageName_example", "Phase_example", int64(123), int64(123), int64(123)) // Pod | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.UpdatePod(context.Background(), uuid, podId).Pod(pod).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.UpdatePod``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**podId** | **string** | Unique identifier of the pod. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdatePodRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **pod** | [**Pod**](Pod.md) |  | 

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


## UpdatePods

> IdsResponse UpdatePods(ctx, uuid).PodsSync(podsSync).Execute()

Sync all pods in a namespace to DB

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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    podsSync := *openapiclient.NewPodsSync("runai-team-a", []openapiclient.Pod{*openapiclient.NewPod("PodId_example", "JobId_example", "ClusterUuid_example", "PodName_example", "ImageName_example", "Phase_example", int64(123), int64(123), int64(123))}) // PodsSync | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.UpdatePods(context.Background(), uuid).PodsSync(podsSync).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.UpdatePods``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdatePods`: IdsResponse
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.UpdatePods`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdatePodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **podsSync** | [**PodsSync**](PodsSync.md) |  | 

### Return type

[**IdsResponse**](IdsResponse.md)

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
    podSyncFields := *openapiclient.NewPodSyncFields("pod-of-a-very-important-job", "high-priority", "Id_example", "WorkloadId_example", "71f69d83-ba66-4822-adf5-55ce55efd210", time.Now(), []openapiclient.Container1{*openapiclient.NewContainer1("alpine", "alpine:latest")}, "Pending") // PodSyncFields | The pod info in the cluster.

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

