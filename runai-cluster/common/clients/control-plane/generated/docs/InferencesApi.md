# \InferencesApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateInference**](InferencesApi.md#CreateInference) | **Post** /api/v1/inference | Create inference workload.
[**CreateInference1**](InferencesApi.md#CreateInference1) | **Post** /api/v1/workloads/inferences | Create an inference. [Experimental]
[**CreateInferenceV2**](InferencesApi.md#CreateInferenceV2) | **Post** /api/v2/inference | Create an inference - assets based [Experimental]
[**DeleteInference**](InferencesApi.md#DeleteInference) | **Delete** /api/v1/workloads/inferences/{workloadId} | Delete an inference. [Experimental]
[**DeleteInferenceById**](InferencesApi.md#DeleteInferenceById) | **Delete** /api/v1/inference/{workloadId} | Delete an inference workload.
[**GetInference**](InferencesApi.md#GetInference) | **Get** /api/v1/workloads/inferences/{workloadId} | Get inference data. [Experimental]
[**GetInferenceById**](InferencesApi.md#GetInferenceById) | **Get** /api/v1/inference/{workloadId} | Get inference content.
[**GetInferenceByIdV2**](InferencesApi.md#GetInferenceByIdV2) | **Get** /api/v2/inference/{workloadId} | Get inference data - assets based [Experimental].
[**GetInferenceWorkloadMetrics**](InferencesApi.md#GetInferenceWorkloadMetrics) | **Get** /api/v1/workloads/inferences/{workloadId}/metrics | Get inference metrics data. [Experimental]
[**GetInferenceWorkloadPodMetrics**](InferencesApi.md#GetInferenceWorkloadPodMetrics) | **Get** /api/v1/workloads/inferences/{workloadId}/pods/{podId}/metrics | Get inference pod&#39;s metrics data. [Experimental]
[**ListInferencesUpdates**](InferencesApi.md#ListInferencesUpdates) | **Get** /api/v1/workloads/inferences/updates | List inference updates.
[**RequestInferencesUpdates**](InferencesApi.md#RequestInferencesUpdates) | **Post** /api/v1/workloads/inferences/updates | Request inference updates



## CreateInference

> Inference CreateInference(ctx).WorkloadCreationRequest(workloadCreationRequest).Execute()

Create inference workload.



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
    workloadCreationRequest := *openapiclient.NewWorkloadCreationRequest("Name_example", "71f69d83-ba66-4822-adf5-55ce55efd210", int32(123), "Namespace_example", *openapiclient.NewAssetsIds("Environment_example")) // WorkloadCreationRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InferencesApi.CreateInference(context.Background()).WorkloadCreationRequest(workloadCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InferencesApi.CreateInference``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateInference`: Inference
    fmt.Fprintf(os.Stdout, "Response from `InferencesApi.CreateInference`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateInferenceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadCreationRequest** | [**WorkloadCreationRequest**](WorkloadCreationRequest.md) |  | 

### Return type

[**Inference**](Inference.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateInference1

> Inference1 CreateInference1(ctx).InferenceCreationRequest(inferenceCreationRequest).Execute()

Create an inference. [Experimental]



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
    inferenceCreationRequest := *openapiclient.NewInferenceCreationRequest("my-workload-name", "1", "71f69d83-ba66-4822-adf5-55ce55efd210") // InferenceCreationRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InferencesApi.CreateInference1(context.Background()).InferenceCreationRequest(inferenceCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InferencesApi.CreateInference1``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateInference1`: Inference1
    fmt.Fprintf(os.Stdout, "Response from `InferencesApi.CreateInference1`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateInference1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **inferenceCreationRequest** | [**InferenceCreationRequest**](InferenceCreationRequest.md) |  | 

### Return type

[**Inference1**](Inference1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateInferenceV2

> InferenceV2 CreateInferenceV2(ctx).InferenceCreationRequestV2(inferenceCreationRequestV2).Execute()

Create an inference - assets based [Experimental]



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
    inferenceCreationRequestV2 := *openapiclient.NewInferenceCreationRequestV2(*openapiclient.NewWorkloadCreationMeta("my-workload-name", "1", "71f69d83-ba66-4822-adf5-55ce55efd210"), *openapiclient.NewAssetsIds("Environment_example")) // InferenceCreationRequestV2 |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InferencesApi.CreateInferenceV2(context.Background()).InferenceCreationRequestV2(inferenceCreationRequestV2).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InferencesApi.CreateInferenceV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateInferenceV2`: InferenceV2
    fmt.Fprintf(os.Stdout, "Response from `InferencesApi.CreateInferenceV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateInferenceV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **inferenceCreationRequestV2** | [**InferenceCreationRequestV2**](InferenceCreationRequestV2.md) |  | 

### Return type

[**InferenceV2**](InferenceV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteInference

> DeleteInference(ctx, workloadId).Execute()

Delete an inference. [Experimental]



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
    resp, r, err := apiClient.InferencesApi.DeleteInference(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InferencesApi.DeleteInference``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeleteInferenceRequest struct via the builder pattern


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


## DeleteInferenceById

> HttpResponse1 DeleteInferenceById(ctx, workloadId).Execute()

Delete an inference workload.



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
    resp, r, err := apiClient.InferencesApi.DeleteInferenceById(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InferencesApi.DeleteInferenceById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteInferenceById`: HttpResponse1
    fmt.Fprintf(os.Stdout, "Response from `InferencesApi.DeleteInferenceById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteInferenceByIdRequest struct via the builder pattern


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


## GetInference

> Inference1 GetInference(ctx, workloadId).Execute()

Get inference data. [Experimental]



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
    resp, r, err := apiClient.InferencesApi.GetInference(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InferencesApi.GetInference``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetInference`: Inference1
    fmt.Fprintf(os.Stdout, "Response from `InferencesApi.GetInference`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInferenceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Inference1**](Inference1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInferenceById

> Inference GetInferenceById(ctx, workloadId).Execute()

Get inference content.



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
    resp, r, err := apiClient.InferencesApi.GetInferenceById(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InferencesApi.GetInferenceById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetInferenceById`: Inference
    fmt.Fprintf(os.Stdout, "Response from `InferencesApi.GetInferenceById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInferenceByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Inference**](Inference.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInferenceByIdV2

> InferenceV2 GetInferenceByIdV2(ctx, workloadId).Execute()

Get inference data - assets based [Experimental].



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
    resp, r, err := apiClient.InferencesApi.GetInferenceByIdV2(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InferencesApi.GetInferenceByIdV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetInferenceByIdV2`: InferenceV2
    fmt.Fprintf(os.Stdout, "Response from `InferencesApi.GetInferenceByIdV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInferenceByIdV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**InferenceV2**](InferenceV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInferenceWorkloadMetrics

> MetricsResponse GetInferenceWorkloadMetrics(ctx, workloadId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()

Get inference metrics data. [Experimental]



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
    metricType := []openapiclient.InferenceWorkloadMetricType{openapiclient.InferenceWorkloadMetricType("THROUGHPUT")} // []InferenceWorkloadMetricType | Specify which data to request.
    start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
    end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
    numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InferencesApi.GetInferenceWorkloadMetrics(context.Background(), workloadId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InferencesApi.GetInferenceWorkloadMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetInferenceWorkloadMetrics`: MetricsResponse
    fmt.Fprintf(os.Stdout, "Response from `InferencesApi.GetInferenceWorkloadMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInferenceWorkloadMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **metricType** | [**[]InferenceWorkloadMetricType**](InferenceWorkloadMetricType.md) | Specify which data to request. | 
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


## GetInferenceWorkloadPodMetrics

> MetricsResponse GetInferenceWorkloadPodMetrics(ctx, workloadId, podId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()

Get inference pod's metrics data. [Experimental]



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
    metricType := []openapiclient.InferencePodMetricType{openapiclient.InferencePodMetricType("THROUGHPUT")} // []InferencePodMetricType | Specifies metrics data to request. Inference metrics are only available for inference workloads.
    start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
    end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
    numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InferencesApi.GetInferenceWorkloadPodMetrics(context.Background(), workloadId, podId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InferencesApi.GetInferenceWorkloadPodMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetInferenceWorkloadPodMetrics`: MetricsResponse
    fmt.Fprintf(os.Stdout, "Response from `InferencesApi.GetInferenceWorkloadPodMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 
**podId** | **string** | The requested pod id. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInferenceWorkloadPodMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metricType** | [**[]InferencePodMetricType**](InferencePodMetricType.md) | Specifies metrics data to request. Inference metrics are only available for inference workloads. | 
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


## ListInferencesUpdates

> WorkloadsSyncResponse ListInferencesUpdates(ctx).UpdatesAfter(updatesAfter).ClusterId(clusterId).Execute()

List inference updates.



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
    resp, r, err := apiClient.InferencesApi.ListInferencesUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InferencesApi.ListInferencesUpdates``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListInferencesUpdates`: WorkloadsSyncResponse
    fmt.Fprintf(os.Stdout, "Response from `InferencesApi.ListInferencesUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListInferencesUpdatesRequest struct via the builder pattern


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


## RequestInferencesUpdates

> WorkloadsSyncResponse RequestInferencesUpdates(ctx).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()

Request inference updates



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
    resp, r, err := apiClient.InferencesApi.RequestInferencesUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InferencesApi.RequestInferencesUpdates``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `RequestInferencesUpdates`: WorkloadsSyncResponse
    fmt.Fprintf(os.Stdout, "Response from `InferencesApi.RequestInferencesUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRequestInferencesUpdatesRequest struct via the builder pattern


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

