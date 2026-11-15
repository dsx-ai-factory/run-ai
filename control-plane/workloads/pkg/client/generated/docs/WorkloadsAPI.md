# \WorkloadsAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CountWorkloadElements**](WorkloadsAPI.md#CountWorkloadElements) | **Get** /api/v1/workloads/{workloadId}/elements/count | Count workload structure elements.
[**CountWorkloads**](WorkloadsAPI.md#CountWorkloads) | **Get** /api/v1/workloads/count | Count workloads.
[**DeleteWorkload**](WorkloadsAPI.md#DeleteWorkload) | **Delete** /api/v1/workloads/{workloadId} | Delete the workload as is does not exist in the cluster anymore.
[**GetInternalWorkloadsTelemetry**](WorkloadsAPI.md#GetInternalWorkloadsTelemetry) | **Get** /internal/workloads/telemetry | Get the internal workloads telemetry.
[**GetWorkload**](WorkloadsAPI.md#GetWorkload) | **Get** /api/v1/workloads/{workloadId} | Get a workload.
[**GetWorkloadAncestors**](WorkloadsAPI.md#GetWorkloadAncestors) | **Get** /internal/workloads/{workloadId}/ancestors | Get the workload ancestors.
[**GetWorkloadElement**](WorkloadsAPI.md#GetWorkloadElement) | **Get** /api/v1/workloads/{workloadId}/elements/{elementId} | Get a single workload structure element by element id.
[**GetWorkloadElementMetrics**](WorkloadsAPI.md#GetWorkloadElementMetrics) | **Get** /api/v1/workloads/{workloadId}/elements/{elementId}/metrics | Get element metrics data.
[**GetWorkloadElements**](WorkloadsAPI.md#GetWorkloadElements) | **Get** /api/v1/workloads/{workloadId}/elements | Get workload structure elements by workload id.
[**GetWorkloadEndpoints**](WorkloadsAPI.md#GetWorkloadEndpoints) | **Get** /api/v1/workloads/{workloadId}/endpoints | Get workload endpoints.
[**GetWorkloadMetrics**](WorkloadsAPI.md#GetWorkloadMetrics) | **Get** /api/v1/workloads/{workloadId}/metrics | Get workload metrics data.
[**GetWorkloadResourceClaims**](WorkloadsAPI.md#GetWorkloadResourceClaims) | **Get** /api/v1/workloads/{workloadId}/resource-claims | Get workload resource claims by workload id.
[**GetWorkloads**](WorkloadsAPI.md#GetWorkloads) | **Get** /api/v1/workloads | List workloads.
[**GetWorkloadsSlim**](WorkloadsAPI.md#GetWorkloadsSlim) | **Get** /api/v1/workloads/slim | List workloads, slim.
[**GetWorkloadsTelemetry**](WorkloadsAPI.md#GetWorkloadsTelemetry) | **Get** /api/v1/workloads/telemetry | Get the workloads telemetry.
[**InternalCreateWorkload**](WorkloadsAPI.md#InternalCreateWorkload) | **Post** /internal/workloads | Create a workload by the control plane.
[**InternalEnrichOrgUnitRanking**](WorkloadsAPI.md#InternalEnrichOrgUnitRanking) | **Post** /internal/workloads/enrich-ranking | Enrich org unit ranking with workload counts (Internal API).
[**InternalGetResourceClaimTemplateWorkloads**](WorkloadsAPI.md#InternalGetResourceClaimTemplateWorkloads) | **Post** /internal/workloads/resource-claim-templates/workloads | List the workloads referencing a resource claim template (Internal API).
[**InternalPatchWorkloadPhase**](WorkloadsAPI.md#InternalPatchWorkloadPhase) | **Patch** /internal/workloads/{workloadId} | patch the workload phase from the client.
[**InternalSyncWorkloads**](WorkloadsAPI.md#InternalSyncWorkloads) | **Post** /internal/workloads/sync | Sync workloads.
[**PatchWorkloadPhase**](WorkloadsAPI.md#PatchWorkloadPhase) | **Patch** /api/v1/workloads/{workloadId} | patch the workload phase from the client.
[**PostSyncPodgroupsBackwards**](WorkloadsAPI.md#PostSyncPodgroupsBackwards) | **Post** /v1/k8s/clusters/{clusterId}/podGroups/sync | Sync pod groups.
[**PostSyncPodsBackwards**](WorkloadsAPI.md#PostSyncPodsBackwards) | **Post** /v1/k8s/clusters/{clusterId}/pods/sync | Sync pods.
[**SyncJobsBackwards**](WorkloadsAPI.md#SyncJobsBackwards) | **Post** /v1/k8s/clusters/{clusterId}/jobs | Sync jobs.
[**SyncPodgroupsBackwards**](WorkloadsAPI.md#SyncPodgroupsBackwards) | **Put** /v1/k8s/clusters/{clusterId}/podGroups | Sync pod groups.
[**SyncPodsBackwards**](WorkloadsAPI.md#SyncPodsBackwards) | **Put** /v1/k8s/clusters/{clusterId}/pods | Sync pods.
[**SyncWorkloads**](WorkloadsAPI.md#SyncWorkloads) | **Post** /api/v1/workloads | Sync workloads.
[**UpsertWorkload**](WorkloadsAPI.md#UpsertWorkload) | **Put** /api/v1/workloads/{workloadId} | Upsert the workload info from the cluster.



## CountWorkloadElements

> CountWorkloadElements200Response CountWorkloadElements(ctx, workloadId).Execute()

Count workload structure elements.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.CountWorkloadElements(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.CountWorkloadElements``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountWorkloadElements`: CountWorkloadElements200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.CountWorkloadElements`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCountWorkloadElementsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**CountWorkloadElements200Response**](CountWorkloadElements200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CountWorkloads

> CountWorkloads200Response CountWorkloads(ctx).Deleted(deleted).FilterBy(filterBy).Search(search).Precise(precise).Execute()

Count workloads.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	deleted := true // bool | Return only deleted resources when `true`. (optional)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)
	search := "test project" // string | Filter results by a free text search. (optional)
	precise := true // bool | Use this if you want to count the precise amount of workloads, uncached. Defaults to false. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.CountWorkloads(context.Background()).Deleted(deleted).FilterBy(filterBy).Search(search).Precise(precise).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.CountWorkloads``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountWorkloads`: CountWorkloads200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.CountWorkloads`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountWorkloadsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deleted** | **bool** | Return only deleted resources when &#x60;true&#x60;. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contain, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 
 **search** | **string** | Filter results by a free text search. | 
 **precise** | **bool** | Use this if you want to count the precise amount of workloads, uncached. Defaults to false. | 

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


## DeleteWorkload

> DeleteWorkload(ctx, workloadId).Execute()

Delete the workload as is does not exist in the cluster anymore.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadsAPI.DeleteWorkload(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.DeleteWorkload``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteWorkloadRequest struct via the builder pattern


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


## GetInternalWorkloadsTelemetry

> TelemetryResponse GetInternalWorkloadsTelemetry(ctx).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolName(nodepoolName).DepartmentId(departmentId).ProjectId(projectId).GroupBy(groupBy).FilterByPhases(filterByPhases).FilterBy(filterBy).Execute()

Get the internal workloads telemetry.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	telemetryType := openapiclient.WorkloadTelemetryType("WORKLOADS_COUNT") // WorkloadTelemetryType | Specifies the telemetry type.
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	nodepoolName := "default" // string | Filter using the nodepool. (optional)
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	groupBy := []string{"GroupBy_example"} // []string | Group workloads by field. (optional)
	filterByPhases := []openapiclient.Phase{openapiclient.Phase("Creating")} // []Phase | Filter workloads by specific phases. If not specified, all phases are included. (optional)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal. Dates are in ISO 8601 timestamp format. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.GetInternalWorkloadsTelemetry(context.Background()).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolName(nodepoolName).DepartmentId(departmentId).ProjectId(projectId).GroupBy(groupBy).FilterByPhases(filterByPhases).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.GetInternalWorkloadsTelemetry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInternalWorkloadsTelemetry`: TelemetryResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.GetInternalWorkloadsTelemetry`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetInternalWorkloadsTelemetryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **telemetryType** | [**WorkloadTelemetryType**](WorkloadTelemetryType.md) | Specifies the telemetry type. | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **nodepoolName** | **string** | Filter using the nodepool. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **groupBy** | **[]string** | Group workloads by field. | 
 **filterByPhases** | [**[]Phase**](Phase.md) | Filter workloads by specific phases. If not specified, all phases are included. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal. Dates are in ISO 8601 timestamp format. | 

### Return type

[**TelemetryResponse**](TelemetryResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkload

> WorkloadDetailed GetWorkload(ctx, workloadId).Execute()

Get a workload.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.GetWorkload(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.GetWorkload``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkload`: WorkloadDetailed
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.GetWorkload`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**WorkloadDetailed**](WorkloadDetailed.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadAncestors

> Ancestors GetWorkloadAncestors(ctx, workloadId).Execute()

Get the workload ancestors.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.GetWorkloadAncestors(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.GetWorkloadAncestors``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadAncestors`: Ancestors
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.GetWorkloadAncestors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadAncestorsRequest struct via the builder pattern


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


## GetWorkloadElement

> WorkloadElementResponse GetWorkloadElement(ctx, workloadId, elementId).Execute()

Get a single workload structure element by element id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	elementId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier of the structure element.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.GetWorkloadElement(context.Background(), workloadId, elementId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.GetWorkloadElement``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadElement`: WorkloadElementResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.GetWorkloadElement`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 
**elementId** | **string** | The unique identifier of the structure element. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadElementRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**WorkloadElementResponse**](WorkloadElementResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadElementMetrics

> MetricsCompositeResponse GetWorkloadElementMetrics(ctx, workloadId, elementId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()

Get element metrics data.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	elementId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier of the structure element.
	metricType := []openapiclient.ElementMetricType{openapiclient.ElementMetricType("GPU_UTILIZATION")} // []ElementMetricType | Specify which metric data to request for the element.
	start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
	end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.GetWorkloadElementMetrics(context.Background(), workloadId, elementId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.GetWorkloadElementMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadElementMetrics`: MetricsCompositeResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.GetWorkloadElementMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 
**elementId** | **string** | The unique identifier of the structure element. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadElementMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metricType** | [**[]ElementMetricType**](ElementMetricType.md) | Specify which metric data to request for the element. | 
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


## GetWorkloadElements

> WorkloadElementsResponse GetWorkloadElements(ctx, workloadId).Offset(offset).Limit(limit).Execute()

Get workload structure elements by workload id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.GetWorkloadElements(context.Background(), workloadId).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.GetWorkloadElements``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadElements`: WorkloadElementsResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.GetWorkloadElements`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadElementsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**WorkloadElementsResponse**](WorkloadElementsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadEndpoints

> GetWorkloadEndpoints200Response GetWorkloadEndpoints(ctx, workloadId).Execute()

Get workload endpoints.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.GetWorkloadEndpoints(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.GetWorkloadEndpoints``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadEndpoints`: GetWorkloadEndpoints200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.GetWorkloadEndpoints`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadEndpointsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GetWorkloadEndpoints200Response**](GetWorkloadEndpoints200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadMetrics

> MetricsCompositeResponse GetWorkloadMetrics(ctx, workloadId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()

Get workload metrics data.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	metricType := []openapiclient.WorkloadMetricType{openapiclient.WorkloadMetricType("GPU_UTILIZATION")} // []WorkloadMetricType | Specify which data to request.
	start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
	end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.GetWorkloadMetrics(context.Background(), workloadId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.GetWorkloadMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadMetrics`: MetricsCompositeResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.GetWorkloadMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **metricType** | [**[]WorkloadMetricType**](WorkloadMetricType.md) | Specify which data to request. | 
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


## GetWorkloadResourceClaims

> WorkloadResourceClaimsResponse GetWorkloadResourceClaims(ctx, workloadId).Offset(offset).Limit(limit).Execute()

Get workload resource claims by workload id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.GetWorkloadResourceClaims(context.Background(), workloadId).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.GetWorkloadResourceClaims``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadResourceClaims`: WorkloadResourceClaimsResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.GetWorkloadResourceClaims`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadResourceClaimsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**WorkloadResourceClaimsResponse**](WorkloadResourceClaimsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloads

> GetWorkloads200Response GetWorkloads(ctx).Deleted(deleted).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Search(search).Execute()

List workloads.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	deleted := true // bool | Return only deleted resources when `true`. (optional)
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	sortBy := "sortBy_example" // string | Sort results by a parameter. (optional)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.GetWorkloads(context.Background()).Deleted(deleted).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.GetWorkloads``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloads`: GetWorkloads200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.GetWorkloads`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deleted** | **bool** | Return only deleted resources when &#x60;true&#x60;. | 
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **sortBy** | **string** | Sort results by a parameter. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contain, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 
 **search** | **string** | Filter results by a free text search. | 

### Return type

[**GetWorkloads200Response**](GetWorkloads200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadsSlim

> GetWorkloadsSlimResponse GetWorkloadsSlim(ctx).Offset(offset).Limit(limit).FilterBy(filterBy).Execute()

List workloads, slim.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.GetWorkloadsSlim(context.Background()).Offset(offset).Limit(limit).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.GetWorkloadsSlim``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadsSlim`: GetWorkloadsSlimResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.GetWorkloadsSlim`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadsSlimRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contain, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. | 

### Return type

[**GetWorkloadsSlimResponse**](GetWorkloadsSlimResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadsTelemetry

> TelemetryResponse GetWorkloadsTelemetry(ctx).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolName(nodepoolName).DepartmentId(departmentId).ProjectId(projectId).GroupBy(groupBy).FilterByPhases(filterByPhases).FilterBy(filterBy).Execute()

Get the workloads telemetry.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	telemetryType := openapiclient.WorkloadTelemetryType("WORKLOADS_COUNT") // WorkloadTelemetryType | Specifies the telemetry type.
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	nodepoolName := "default" // string | Filter using the nodepool. (optional)
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	groupBy := []string{"GroupBy_example"} // []string | Group workloads by field. (optional)
	filterByPhases := []openapiclient.Phase{openapiclient.Phase("Creating")} // []Phase | Filter workloads by specific phases. If not specified, all phases are included. (optional)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal. Dates are in ISO 8601 timestamp format. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.GetWorkloadsTelemetry(context.Background()).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolName(nodepoolName).DepartmentId(departmentId).ProjectId(projectId).GroupBy(groupBy).FilterByPhases(filterByPhases).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.GetWorkloadsTelemetry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadsTelemetry`: TelemetryResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.GetWorkloadsTelemetry`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadsTelemetryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **telemetryType** | [**WorkloadTelemetryType**](WorkloadTelemetryType.md) | Specifies the telemetry type. | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **nodepoolName** | **string** | Filter using the nodepool. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **groupBy** | **[]string** | Group workloads by field. | 
 **filterByPhases** | [**[]Phase**](Phase.md) | Filter workloads by specific phases. If not specified, all phases are included. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal. Dates are in ISO 8601 timestamp format. | 

### Return type

[**TelemetryResponse**](TelemetryResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, text/csv

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InternalCreateWorkload

> Workload InternalCreateWorkload(ctx).WorkloadCreationFields(workloadCreationFields).Execute()

Create a workload by the control plane.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadCreationFields := *openapiclient.NewWorkloadCreationFields("runai-job", "very-important-job", "Id_example", "71f69d83-ba66-4822-adf5-55ce55efd210", "1", time.Now()) // WorkloadCreationFields | The workload info in the control plane.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.InternalCreateWorkload(context.Background()).WorkloadCreationFields(workloadCreationFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.InternalCreateWorkload``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `InternalCreateWorkload`: Workload
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.InternalCreateWorkload`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInternalCreateWorkloadRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadCreationFields** | [**WorkloadCreationFields**](WorkloadCreationFields.md) | The workload info in the control plane. | 

### Return type

[**Workload**](Workload.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InternalEnrichOrgUnitRanking

> InternalEnrichOrgUnitRankingResponse InternalEnrichOrgUnitRanking(ctx).InternalEnrichOrgUnitRankingRequest(internalEnrichOrgUnitRankingRequest).Execute()

Enrich org unit ranking with workload counts (Internal API).



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	internalEnrichOrgUnitRankingRequest := *openapiclient.NewInternalEnrichOrgUnitRankingRequest("abc-cluster-123", "project", []string{"Ids_example"}) // InternalEnrichOrgUnitRankingRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.InternalEnrichOrgUnitRanking(context.Background()).InternalEnrichOrgUnitRankingRequest(internalEnrichOrgUnitRankingRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.InternalEnrichOrgUnitRanking``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `InternalEnrichOrgUnitRanking`: InternalEnrichOrgUnitRankingResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.InternalEnrichOrgUnitRanking`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInternalEnrichOrgUnitRankingRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **internalEnrichOrgUnitRankingRequest** | [**InternalEnrichOrgUnitRankingRequest**](InternalEnrichOrgUnitRankingRequest.md) |  | 

### Return type

[**InternalEnrichOrgUnitRankingResponse**](InternalEnrichOrgUnitRankingResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InternalGetResourceClaimTemplateWorkloads

> ResourceClaimTemplateWorkloadsResponse InternalGetResourceClaimTemplateWorkloads(ctx).InternalResourceClaimTemplateWorkloadsRequest(internalResourceClaimTemplateWorkloadsRequest).Offset(offset).Limit(limit).Execute()

List the workloads referencing a resource claim template (Internal API).



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	internalResourceClaimTemplateWorkloadsRequest := *openapiclient.NewInternalResourceClaimTemplateWorkloadsRequest("single-gpu", openapiclient.ScopeType("system"), "ScopeId_example") // InternalResourceClaimTemplateWorkloadsRequest | 
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.InternalGetResourceClaimTemplateWorkloads(context.Background()).InternalResourceClaimTemplateWorkloadsRequest(internalResourceClaimTemplateWorkloadsRequest).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.InternalGetResourceClaimTemplateWorkloads``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `InternalGetResourceClaimTemplateWorkloads`: ResourceClaimTemplateWorkloadsResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.InternalGetResourceClaimTemplateWorkloads`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInternalGetResourceClaimTemplateWorkloadsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **internalResourceClaimTemplateWorkloadsRequest** | [**InternalResourceClaimTemplateWorkloadsRequest**](InternalResourceClaimTemplateWorkloadsRequest.md) |  | 
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**ResourceClaimTemplateWorkloadsResponse**](ResourceClaimTemplateWorkloadsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InternalPatchWorkloadPhase

> InternalPatchWorkloadPhase(ctx, workloadId).WorkloadPatchFields(workloadPatchFields).Execute()

patch the workload phase from the client.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	workloadPatchFields := *openapiclient.NewWorkloadPatchFields(openapiclient.Phase("Creating")) // WorkloadPatchFields | The workload info in the cluster.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadsAPI.InternalPatchWorkloadPhase(context.Background(), workloadId).WorkloadPatchFields(workloadPatchFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.InternalPatchWorkloadPhase``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiInternalPatchWorkloadPhaseRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadPatchFields** | [**WorkloadPatchFields**](WorkloadPatchFields.md) | The workload info in the cluster. | 

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


## InternalSyncWorkloads

> InternalSyncWorkloads(ctx).FilterBy(filterBy).InternalWorkloadsSyncRequest(internalWorkloadsSyncRequest).Execute()

Sync workloads.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)
	internalWorkloadsSyncRequest := *openapiclient.NewInternalWorkloadsSyncRequest("71f69d83-ba66-4822-adf5-55ce55efd210", []openapiclient.WorkloadSyncFields{*openapiclient.NewWorkloadSyncFields("runai-job", "very-important-job", "Id_example", "high-priority", "71f69d83-ba66-4822-adf5-55ce55efd210", "proj-1", "1", "department-1", "1", "runai-proj-1", time.Now(), openapiclient.Phase("Creating"), openapiclient.Source("CLI"), "Pending")}) // InternalWorkloadsSyncRequest | The workloads to sync. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadsAPI.InternalSyncWorkloads(context.Background()).FilterBy(filterBy).InternalWorkloadsSyncRequest(internalWorkloadsSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.InternalSyncWorkloads``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInternalSyncWorkloadsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contain, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 
 **internalWorkloadsSyncRequest** | [**InternalWorkloadsSyncRequest**](InternalWorkloadsSyncRequest.md) | The workloads to sync. | 

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


## PatchWorkloadPhase

> PatchWorkloadPhase(ctx, workloadId).WorkloadPatchFields(workloadPatchFields).Execute()

patch the workload phase from the client.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	workloadPatchFields := *openapiclient.NewWorkloadPatchFields(openapiclient.Phase("Creating")) // WorkloadPatchFields | The workload info in the cluster.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadsAPI.PatchWorkloadPhase(context.Background(), workloadId).WorkloadPatchFields(workloadPatchFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.PatchWorkloadPhase``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchWorkloadPhaseRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadPatchFields** | [**WorkloadPatchFields**](WorkloadPatchFields.md) | The workload info in the cluster. | 

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


## PostSyncPodgroupsBackwards

> string PostSyncPodgroupsBackwards(ctx, clusterId).SyncJobsBackwardsRequest(syncJobsBackwardsRequest).Execute()

Sync pod groups.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	syncJobsBackwardsRequest := *openapiclient.NewSyncJobsBackwardsRequest() // SyncJobsBackwardsRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.PostSyncPodgroupsBackwards(context.Background(), clusterId).SyncJobsBackwardsRequest(syncJobsBackwardsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.PostSyncPodgroupsBackwards``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PostSyncPodgroupsBackwards`: string
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.PostSyncPodgroupsBackwards`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiPostSyncPodgroupsBackwardsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **syncJobsBackwardsRequest** | [**SyncJobsBackwardsRequest**](SyncJobsBackwardsRequest.md) |  | 

### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PostSyncPodsBackwards

> string PostSyncPodsBackwards(ctx, clusterId).SyncJobsBackwardsRequest(syncJobsBackwardsRequest).Execute()

Sync pods.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	syncJobsBackwardsRequest := *openapiclient.NewSyncJobsBackwardsRequest() // SyncJobsBackwardsRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.PostSyncPodsBackwards(context.Background(), clusterId).SyncJobsBackwardsRequest(syncJobsBackwardsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.PostSyncPodsBackwards``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PostSyncPodsBackwards`: string
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.PostSyncPodsBackwards`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiPostSyncPodsBackwardsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **syncJobsBackwardsRequest** | [**SyncJobsBackwardsRequest**](SyncJobsBackwardsRequest.md) |  | 

### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncJobsBackwards

> SyncJobsBackwards(ctx, clusterId).SyncJobsBackwardsRequest(syncJobsBackwardsRequest).Execute()

Sync jobs.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	syncJobsBackwardsRequest := *openapiclient.NewSyncJobsBackwardsRequest() // SyncJobsBackwardsRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadsAPI.SyncJobsBackwards(context.Background(), clusterId).SyncJobsBackwardsRequest(syncJobsBackwardsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.SyncJobsBackwards``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncJobsBackwardsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **syncJobsBackwardsRequest** | [**SyncJobsBackwardsRequest**](SyncJobsBackwardsRequest.md) |  | 

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


## SyncPodgroupsBackwards

> SyncPodsBackwards200Response SyncPodgroupsBackwards(ctx, clusterId).ProjectName(projectName).SyncJobsBackwardsRequest(syncJobsBackwardsRequest).Execute()

Sync pod groups.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	projectName := "projectName_example" // string |  (optional)
	syncJobsBackwardsRequest := *openapiclient.NewSyncJobsBackwardsRequest() // SyncJobsBackwardsRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.SyncPodgroupsBackwards(context.Background(), clusterId).ProjectName(projectName).SyncJobsBackwardsRequest(syncJobsBackwardsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.SyncPodgroupsBackwards``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncPodgroupsBackwards`: SyncPodsBackwards200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.SyncPodgroupsBackwards`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncPodgroupsBackwardsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **projectName** | **string** |  | 
 **syncJobsBackwardsRequest** | [**SyncJobsBackwardsRequest**](SyncJobsBackwardsRequest.md) |  | 

### Return type

[**SyncPodsBackwards200Response**](SyncPodsBackwards200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncPodsBackwards

> SyncPodsBackwards200Response SyncPodsBackwards(ctx, clusterId).SyncJobsBackwardsRequest(syncJobsBackwardsRequest).Execute()

Sync pods.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	syncJobsBackwardsRequest := *openapiclient.NewSyncJobsBackwardsRequest() // SyncJobsBackwardsRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsAPI.SyncPodsBackwards(context.Background(), clusterId).SyncJobsBackwardsRequest(syncJobsBackwardsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.SyncPodsBackwards``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncPodsBackwards`: SyncPodsBackwards200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsAPI.SyncPodsBackwards`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncPodsBackwardsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **syncJobsBackwardsRequest** | [**SyncJobsBackwardsRequest**](SyncJobsBackwardsRequest.md) |  | 

### Return type

[**SyncPodsBackwards200Response**](SyncPodsBackwards200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncWorkloads

> SyncWorkloads(ctx).FilterBy(filterBy).WorkloadsSyncRequest(workloadsSyncRequest).Execute()

Sync workloads.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)
	workloadsSyncRequest := *openapiclient.NewWorkloadsSyncRequest("1", []openapiclient.WorkloadSyncFields{*openapiclient.NewWorkloadSyncFields("runai-job", "very-important-job", "Id_example", "high-priority", "71f69d83-ba66-4822-adf5-55ce55efd210", "proj-1", "1", "department-1", "1", "runai-proj-1", time.Now(), openapiclient.Phase("Creating"), openapiclient.Source("CLI"), "Pending")}) // WorkloadsSyncRequest | The workloads to sync. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadsAPI.SyncWorkloads(context.Background()).FilterBy(filterBy).WorkloadsSyncRequest(workloadsSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.SyncWorkloads``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncWorkloadsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contain, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 
 **workloadsSyncRequest** | [**WorkloadsSyncRequest**](WorkloadsSyncRequest.md) | The workloads to sync. | 

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


## UpsertWorkload

> UpsertWorkload(ctx, workloadId).WorkloadSyncFields(workloadSyncFields).Execute()

Upsert the workload info from the cluster.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	workloadSyncFields := *openapiclient.NewWorkloadSyncFields("runai-job", "very-important-job", "Id_example", "high-priority", "71f69d83-ba66-4822-adf5-55ce55efd210", "proj-1", "1", "department-1", "1", "runai-proj-1", time.Now(), openapiclient.Phase("Creating"), openapiclient.Source("CLI"), "Pending") // WorkloadSyncFields | The workload info in the cluster.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadsAPI.UpsertWorkload(context.Background(), workloadId).WorkloadSyncFields(workloadSyncFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsAPI.UpsertWorkload``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpsertWorkloadRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadSyncFields** | [**WorkloadSyncFields**](WorkloadSyncFields.md) | The workload info in the cluster. | 

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

