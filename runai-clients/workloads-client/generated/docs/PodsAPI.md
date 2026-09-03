# \PodsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CountPods**](PodsAPI.md#CountPods) | **Get** /api/v1/workloads/pods/count | Get pods count.
[**DeletePod**](PodsAPI.md#DeletePod) | **Delete** /api/v1/workloads/pods/{podId} | Delete a pod as it does not exist in the cluster anymore.
[**GetRevisionPods**](PodsAPI.md#GetRevisionPods) | **Get** /api/v1/workloads/inferences/revisions/{revisionId}/pods | Get revision pods by id.
[**GetWorkloadPod**](PodsAPI.md#GetWorkloadPod) | **Get** /api/v1/workloads/{workloadId}/pods/{podId} | Get a single pod scoped to a workload.
[**GetWorkloadPodMetrics**](PodsAPI.md#GetWorkloadPodMetrics) | **Get** /api/v1/workloads/{workloadId}/pods/{podId}/metrics | Get pod metrics data.
[**GetWorkloadPods**](PodsAPI.md#GetWorkloadPods) | **Get** /api/v1/workloads/{workloadId}/pods | Get workload pods by id.
[**InternalSyncPods**](PodsAPI.md#InternalSyncPods) | **Post** /internal/workloads/pods/sync | Sync pods.
[**InternalUpdateNodePods**](PodsAPI.md#InternalUpdateNodePods) | **Put** /internal/workloads/clusters/{clusterId}/nodes/{nodeName}/pods | Update pods on a specific node in a cluster
[**ListPods**](PodsAPI.md#ListPods) | **Get** /api/v1/workloads/pods | List pods.
[**SyncPods**](PodsAPI.md#SyncPods) | **Post** /api/v1/workloads/pods/sync | Sync workloads pods.
[**SyncWorkloadsPods**](PodsAPI.md#SyncWorkloadsPods) | **Post** /api/v1/workloads/pods | Sync workloads pods.
[**UpsertPod**](PodsAPI.md#UpsertPod) | **Put** /api/v1/workloads/pods/{podId} | Upsert a single pod info from the cluster.



## CountPods

> CountWorkloads200Response CountPods(ctx).Deleted(deleted).FilterBy(filterBy).Search(search).Completed(completed).Execute()

Get pods count.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	deleted := true // bool | Return only deleted resources when `true`. (optional)
	filterBy := []string{"Inner_example"} // []string | Filter results using a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contains, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)
	search := "test project" // string | Filter results by a free text search. (optional)
	completed := "completed_example" // string | Return only completed resources when 'true', return only non-completed resources when 'false'. By default, or when empty, returns all resources. (optional) (default to "all")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PodsAPI.CountPods(context.Background()).Deleted(deleted).FilterBy(filterBy).Search(search).Completed(completed).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodsAPI.CountPods``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountPods`: CountWorkloads200Response
	fmt.Fprintf(os.Stdout, "Response from `PodsAPI.CountPods`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountPodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deleted** | **bool** | Return only deleted resources when &#x60;true&#x60;. | 
 **filterBy** | **[]string** | Filter results using a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contains, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 
 **search** | **string** | Filter results by a free text search. | 
 **completed** | **string** | Return only completed resources when &#39;true&#39;, return only non-completed resources when &#39;false&#39;. By default, or when empty, returns all resources. | [default to &quot;all&quot;]

### Return type

[**CountWorkloads200Response**](CountWorkloads200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


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
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	podId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The requested pod id.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PodsAPI.DeletePod(context.Background(), podId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodsAPI.DeletePod``: %v\n", err)
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


## GetRevisionPods

> GetRevisionPods200Response GetRevisionPods(ctx, revisionId).Deleted(deleted).Execute()

Get revision pods by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	revisionId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the revision.
	deleted := true // bool | Return only deleted resources when `true`. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PodsAPI.GetRevisionPods(context.Background(), revisionId).Deleted(deleted).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodsAPI.GetRevisionPods``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRevisionPods`: GetRevisionPods200Response
	fmt.Fprintf(os.Stdout, "Response from `PodsAPI.GetRevisionPods`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**revisionId** | **string** | The  Universally Unique Identifier (UUID) of the revision. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRevisionPodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deleted** | **bool** | Return only deleted resources when &#x60;true&#x60;. | 

### Return type

[**GetRevisionPods200Response**](GetRevisionPods200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadPod

> Pod GetWorkloadPod(ctx, workloadId, podId).Execute()

Get a single pod scoped to a workload.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	podId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The requested pod id.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PodsAPI.GetWorkloadPod(context.Background(), workloadId, podId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodsAPI.GetWorkloadPod``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadPod`: Pod
	fmt.Fprintf(os.Stdout, "Response from `PodsAPI.GetWorkloadPod`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 
**podId** | **string** | The requested pod id. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadPodRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Pod**](Pod.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadPodMetrics

> MetricsCompositeResponse GetWorkloadPodMetrics(ctx, workloadId, podId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()

Get pod metrics data.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	podId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The requested pod id.
	metricType := []openapiclient.PodMetricType{openapiclient.PodMetricType("GPU_UTILIZATION_PER_GPU")} // []PodMetricType | Specify which metric data to request. Advanced GPU metrics are only supported if the 'Advanced GPU Metrics' feature flag in the settings is enabled.
	start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
	end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PodsAPI.GetWorkloadPodMetrics(context.Background(), workloadId, podId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodsAPI.GetWorkloadPodMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadPodMetrics`: MetricsCompositeResponse
	fmt.Fprintf(os.Stdout, "Response from `PodsAPI.GetWorkloadPodMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 
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

[**MetricsCompositeResponse**](MetricsCompositeResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, text/csv

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadPods

> ListPods200Response GetWorkloadPods(ctx, workloadId).Deleted(deleted).Offset(offset).Limit(limit).Search(search).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()

Get workload pods by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	deleted := true // bool | Return only deleted resources when `true`. (optional)
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	search := "test project" // string | Filter results by a free text search. (optional)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	sortBy := "sortBy_example" // string | Sort results using a parameter. (optional)
	filterBy := []string{"Inner_example"} // []string | Filter results using a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contains, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PodsAPI.GetWorkloadPods(context.Background(), workloadId).Deleted(deleted).Offset(offset).Limit(limit).Search(search).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodsAPI.GetWorkloadPods``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadPods`: ListPods200Response
	fmt.Fprintf(os.Stdout, "Response from `PodsAPI.GetWorkloadPods`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadPodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deleted** | **bool** | Return only deleted resources when &#x60;true&#x60;. | 
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **search** | **string** | Filter results by a free text search. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **sortBy** | **string** | Sort results using a parameter. | 
 **filterBy** | **[]string** | Filter results using a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contains, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 

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
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	internalPodsSyncRequest := *openapiclient.NewInternalPodsSyncRequest("71f69d83-ba66-4822-adf5-55ce55efd210", []openapiclient.PodSyncFields{*openapiclient.NewPodSyncFields("pod-of-a-very-important-job", "high-priority", "Id_example", "WorkloadId_example", "71f69d83-ba66-4822-adf5-55ce55efd210", time.Now(), []openapiclient.Container{*openapiclient.NewContainer("alpine", "alpine:latest")}, "Pending")}) // InternalPodsSyncRequest | The pods to sync. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PodsAPI.InternalSyncPods(context.Background()).InternalPodsSyncRequest(internalPodsSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodsAPI.InternalSyncPods``: %v\n", err)
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


## InternalUpdateNodePods

> InternalUpdateNodePods(ctx, clusterId, nodeName).UpdateNodePodsRequest(updateNodePodsRequest).Execute()

Update pods on a specific node in a cluster



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	nodeName := "nodeName_example" // string | The node name where the pods are running.
	updateNodePodsRequest := *openapiclient.NewUpdateNodePodsRequest() // UpdateNodePodsRequest | The data to update for all pods on this node

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PodsAPI.InternalUpdateNodePods(context.Background(), clusterId, nodeName).UpdateNodePodsRequest(updateNodePodsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodsAPI.InternalUpdateNodePods``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**nodeName** | **string** | The node name where the pods are running. | 

### Other Parameters

Other parameters are passed through a pointer to a apiInternalUpdateNodePodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **updateNodePodsRequest** | [**UpdateNodePodsRequest**](UpdateNodePodsRequest.md) | The data to update for all pods on this node | 

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

> ListPods200Response ListPods(ctx).Deleted(deleted).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Verbosity(verbosity).Completed(completed).Search(search).Execute()

List pods.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
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
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PodsAPI.ListPods(context.Background()).Deleted(deleted).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Verbosity(verbosity).Completed(completed).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodsAPI.ListPods``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPods`: ListPods200Response
	fmt.Fprintf(os.Stdout, "Response from `PodsAPI.ListPods`: %v\n", resp)
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
 **search** | **string** | Filter results by a free text search. | 

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


## SyncPods

> SyncPods(ctx).WorkloadPodsSyncRequest(workloadPodsSyncRequest).Execute()

Sync workloads pods.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	workloadPodsSyncRequest := []openapiclient.WorkloadPodsSyncRequest{*openapiclient.NewWorkloadPodsSyncRequest("1", "1", []openapiclient.PodSyncFields{*openapiclient.NewPodSyncFields("pod-of-a-very-important-job", "high-priority", "Id_example", "WorkloadId_example", "71f69d83-ba66-4822-adf5-55ce55efd210", time.Now(), []openapiclient.Container{*openapiclient.NewContainer("alpine", "alpine:latest")}, "Pending")})} // []WorkloadPodsSyncRequest | The pods to sync. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PodsAPI.SyncPods(context.Background()).WorkloadPodsSyncRequest(workloadPodsSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodsAPI.SyncPods``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncPodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadPodsSyncRequest** | [**[]WorkloadPodsSyncRequest**](WorkloadPodsSyncRequest.md) | The pods to sync. | 

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
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	podsSyncRequest := *openapiclient.NewPodsSyncRequest("1", []openapiclient.PodSyncFields{*openapiclient.NewPodSyncFields("pod-of-a-very-important-job", "high-priority", "Id_example", "WorkloadId_example", "71f69d83-ba66-4822-adf5-55ce55efd210", time.Now(), []openapiclient.Container{*openapiclient.NewContainer("alpine", "alpine:latest")}, "Pending")}) // PodsSyncRequest | The pods to sync. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PodsAPI.SyncWorkloadsPods(context.Background()).PodsSyncRequest(podsSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodsAPI.SyncWorkloadsPods``: %v\n", err)
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
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	podId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The requested pod id.
	podSyncFields := *openapiclient.NewPodSyncFields("pod-of-a-very-important-job", "high-priority", "Id_example", "WorkloadId_example", "71f69d83-ba66-4822-adf5-55ce55efd210", time.Now(), []openapiclient.Container{*openapiclient.NewContainer("alpine", "alpine:latest")}, "Pending") // PodSyncFields | The pod info in the cluster.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PodsAPI.UpsertPod(context.Background(), podId).PodSyncFields(podSyncFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodsAPI.UpsertPod``: %v\n", err)
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

