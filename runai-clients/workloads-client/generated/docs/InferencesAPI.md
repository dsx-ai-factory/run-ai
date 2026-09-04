# \InferencesAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateInference**](InferencesAPI.md#CreateInference) | **Post** /api/v1/workloads/inferences | Create an inference.
[**DeleteInference**](InferencesAPI.md#DeleteInference) | **Delete** /api/v1/workloads/inferences/{workloadId} | Delete an inference.
[**GetInference**](InferencesAPI.md#GetInference) | **Get** /api/v1/workloads/inferences/{workloadId} | Get inference data.
[**GetInferenceAncestors**](InferencesAPI.md#GetInferenceAncestors) | **Get** /internal/workloads/inferences/{workloadId}/ancestors | Get the inference or revision ancestors.
[**GetInferenceWorkloadMetrics**](InferencesAPI.md#GetInferenceWorkloadMetrics) | **Get** /api/v1/workloads/inferences/{workloadId}/metrics | Get inference metrics data.
[**GetInferenceWorkloadPodMetrics**](InferencesAPI.md#GetInferenceWorkloadPodMetrics) | **Get** /api/v1/workloads/inferences/{workloadId}/pods/{podId}/metrics | Get inference pod&#39;s metrics data.
[**GetNimCatalog**](InferencesAPI.md#GetNimCatalog) | **Get** /api/v1/workloads/nim/catalog | Search NGC Catalog Resources
[**ListInferencesUpdates**](InferencesAPI.md#ListInferencesUpdates) | **Get** /api/v1/workloads/inferences/updates | List inference updates.
[**RequestInferencesUpdates**](InferencesAPI.md#RequestInferencesUpdates) | **Post** /api/v1/workloads/inferences/updates | Request inference updates
[**UpdateInferenceSpec**](InferencesAPI.md#UpdateInferenceSpec) | **Patch** /api/v1/workloads/inferences/{workloadId} | Update inference spec.



## CreateInference

> Inference CreateInference(ctx).InferenceCreationRequest(inferenceCreationRequest).Execute()

Create an inference.



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
	inferenceCreationRequest := *openapiclient.NewInferenceCreationRequest("my-workload-name", "1", "71f69d83-ba66-4822-adf5-55ce55efd210") // InferenceCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.CreateInference(context.Background()).InferenceCreationRequest(inferenceCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.CreateInference``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateInference`: Inference
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.CreateInference`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateInferenceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **inferenceCreationRequest** | [**InferenceCreationRequest**](InferenceCreationRequest.md) |  | 

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


## DeleteInference

> HttpResponse DeleteInference(ctx, workloadId).Execute()

Delete an inference.



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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.DeleteInference(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.DeleteInference``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteInference`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.DeleteInference`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteInferenceRequest struct via the builder pattern


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


## GetInference

> Inference GetInference(ctx, workloadId).Execute()

Get inference data.



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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.GetInference(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.GetInference``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInference`: Inference
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.GetInference`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInferenceRequest struct via the builder pattern


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


## GetInferenceAncestors

> Ancestors GetInferenceAncestors(ctx, workloadId).Execute()

Get the inference or revision ancestors.

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.GetInferenceAncestors(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.GetInferenceAncestors``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInferenceAncestors`: Ancestors
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.GetInferenceAncestors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInferenceAncestorsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Ancestors**](Ancestors.md)

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

Get inference metrics data.



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
	metricType := []openapiclient.InferenceWorkloadMetricType{openapiclient.InferenceWorkloadMetricType("THROUGHPUT")} // []InferenceWorkloadMetricType | Specify which data to request.
	start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
	end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.GetInferenceWorkloadMetrics(context.Background(), workloadId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.GetInferenceWorkloadMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInferenceWorkloadMetrics`: MetricsResponse
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.GetInferenceWorkloadMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

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

Get inference pod's metrics data.



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
	metricType := []openapiclient.InferencePodMetricType{openapiclient.InferencePodMetricType("THROUGHPUT")} // []InferencePodMetricType | Specifies metrics data to request. Inference metrics are only available for inference workloads.
	start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
	end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.GetInferenceWorkloadPodMetrics(context.Background(), workloadId, podId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.GetInferenceWorkloadPodMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInferenceWorkloadPodMetrics`: MetricsResponse
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.GetInferenceWorkloadPodMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 
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


## GetNimCatalog

> Catalog GetNimCatalog(ctx).Filter(filter).Query(query).OrderBy(orderBy).Page(page).PageSize(pageSize).Execute()

Search NGC Catalog Resources



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
	filter := []string{"Inner_example"} // []string | Repeated filter in the form `filter=field:value`. Exact match only (e.g., `filter=orgName:nvidia`).  (optional)
	query := "*:*" // string | Full-text Lucene query (e.g., `name:llama*`, `*:*`, or `orgName:\"nvidia\"`).  (optional)
	orderBy := "dateCreated,DESC" // string | Sorting field and direction, comma separated (e.g., `dateCreated,DESC`).  (optional)
	page := int32(0) // int32 | Page number for pagination (0-based indexing).  (optional) (default to 0)
	pageSize := int32(50) // int32 | Number of results per page.  (optional) (default to 200)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.GetNimCatalog(context.Background()).Filter(filter).Query(query).OrderBy(orderBy).Page(page).PageSize(pageSize).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.GetNimCatalog``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNimCatalog`: Catalog
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.GetNimCatalog`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNimCatalogRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filter** | **[]string** | Repeated filter in the form &#x60;filter&#x3D;field:value&#x60;. Exact match only (e.g., &#x60;filter&#x3D;orgName:nvidia&#x60;).  | 
 **query** | **string** | Full-text Lucene query (e.g., &#x60;name:llama*&#x60;, &#x60;*:*&#x60;, or &#x60;orgName:\&quot;nvidia\&quot;&#x60;).  | 
 **orderBy** | **string** | Sorting field and direction, comma separated (e.g., &#x60;dateCreated,DESC&#x60;).  | 
 **page** | **int32** | Page number for pagination (0-based indexing).  | [default to 0]
 **pageSize** | **int32** | Number of results per page.  | [default to 200]

### Return type

[**Catalog**](Catalog.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

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
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	updatesAfter := int64(12345) // int64 | list only data with update counter greater than the provided value. for example updatesAfter=5 should fetch data with updates counter 6 and above.
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.ListInferencesUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.ListInferencesUpdates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListInferencesUpdates`: WorkloadsSyncResponse
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.ListInferencesUpdates`: %v\n", resp)
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
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	updatesAfter := int64(12345) // int64 | list only data with update counter greater than the provided value. for example updatesAfter=5 should fetch data with updates counter 6 and above.
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.RequestInferencesUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.RequestInferencesUpdates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RequestInferencesUpdates`: WorkloadsSyncResponse
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.RequestInferencesUpdates`: %v\n", resp)
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


## UpdateInferenceSpec

> Inference UpdateInferenceSpec(ctx, workloadId).InferenceUpdateRequest(inferenceUpdateRequest).Execute()

Update inference spec.



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
	inferenceUpdateRequest := *openapiclient.NewInferenceUpdateRequest() // InferenceUpdateRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.UpdateInferenceSpec(context.Background(), workloadId).InferenceUpdateRequest(inferenceUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.UpdateInferenceSpec``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateInferenceSpec`: Inference
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.UpdateInferenceSpec`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateInferenceSpecRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **inferenceUpdateRequest** | [**InferenceUpdateRequest**](InferenceUpdateRequest.md) |  | 

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

