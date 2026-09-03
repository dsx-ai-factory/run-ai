# \WorkloadsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CountWorkloads**](WorkloadsApi.md#CountWorkloads) | **Get** /api/v1/workloads/count | Count workloads.
[**DeleteWorkload**](WorkloadsApi.md#DeleteWorkload) | **Delete** /api/v1/workloads/{workloadId} | Delete the workload as is does not exist in the cluster anymore.
[**GetWorkload**](WorkloadsApi.md#GetWorkload) | **Get** /api/v1/workloads/{workloadId} | Get a workload.
[**GetWorkloadAncestors**](WorkloadsApi.md#GetWorkloadAncestors) | **Get** /internal/workloads/{workloadId}/ancestors | Get the workload ancestors.
[**GetWorkloadMetrics**](WorkloadsApi.md#GetWorkloadMetrics) | **Get** /api/v1/workloads/{workloadId}/metrics | Get workload metrics data. [Experimental]
[**GetWorkloads**](WorkloadsApi.md#GetWorkloads) | **Get** /api/v1/workloads | List workloads.
[**GetWorkloadsTelemetry**](WorkloadsApi.md#GetWorkloadsTelemetry) | **Get** /api/v1/workloads/telemetry | Get the workloads telemetry. [Experimental]
[**InternalCreateWorkload**](WorkloadsApi.md#InternalCreateWorkload) | **Post** /internal/workloads | Create a workload by the control plane.
[**InternalSyncWorkloads**](WorkloadsApi.md#InternalSyncWorkloads) | **Post** /internal/workloads/sync | Sync workloads.
[**PatchWorkloadPhase**](WorkloadsApi.md#PatchWorkloadPhase) | **Patch** /api/v1/workloads/{workloadId} | patch the workload phase from the client.
[**SyncWorkloads**](WorkloadsApi.md#SyncWorkloads) | **Post** /api/v1/workloads | Sync workloads.
[**UpsertWorkload**](WorkloadsApi.md#UpsertWorkload) | **Put** /api/v1/workloads/{workloadId} | Upsert the workload info from the cluster.



## CountWorkloads

> CountAccessRules200Response CountWorkloads(ctx).Deleted(deleted).FilterBy(filterBy).Execute()

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
    deleted := true // bool | Return only deleted resources when `true`. (optional)
    filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.CountWorkloads(context.Background()).Deleted(deleted).FilterBy(filterBy).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.CountWorkloads``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CountWorkloads`: CountAccessRules200Response
    fmt.Fprintf(os.Stdout, "Response from `WorkloadsApi.CountWorkloads`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountWorkloadsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deleted** | **bool** | Return only deleted resources when &#x60;true&#x60;. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contain, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 

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
    openapiclient "./openapi"
)

func main() {
    workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the workload.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.DeleteWorkload(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.DeleteWorkload``: %v\n", err)
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


## GetWorkload

> Workload GetWorkload(ctx, workloadId).Execute()

Get a workload.



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
    resp, r, err := apiClient.WorkloadsApi.GetWorkload(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.GetWorkload``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkload`: Workload
    fmt.Fprintf(os.Stdout, "Response from `WorkloadsApi.GetWorkload`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Workload**](Workload.md)

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
    openapiclient "./openapi"
)

func main() {
    workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the workload.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.GetWorkloadAncestors(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.GetWorkloadAncestors``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkloadAncestors`: Ancestors
    fmt.Fprintf(os.Stdout, "Response from `WorkloadsApi.GetWorkloadAncestors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

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


## GetWorkloadMetrics

> MetricsResponse GetWorkloadMetrics(ctx, workloadId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()

Get workload metrics data. [Experimental]



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
    metricType := []openapiclient.WorkloadMetricType{openapiclient.WorkloadMetricType("GPU_UTILIZATION")} // []WorkloadMetricType | Specify which data to request.
    start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
    end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
    numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.GetWorkloadMetrics(context.Background(), workloadId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.GetWorkloadMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkloadMetrics`: MetricsResponse
    fmt.Fprintf(os.Stdout, "Response from `WorkloadsApi.GetWorkloadMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **metricType** | [**[]WorkloadMetricType**](WorkloadMetricType.md) | Specify which data to request. | 
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


## GetWorkloads

> GetWorkloads200Response GetWorkloads(ctx).Deleted(deleted).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()

List workloads.



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
    sortBy := "sortBy_example" // string | Sort results by a parameter. (optional)
    filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.GetWorkloads(context.Background()).Deleted(deleted).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.GetWorkloads``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkloads`: GetWorkloads200Response
    fmt.Fprintf(os.Stdout, "Response from `WorkloadsApi.GetWorkloads`: %v\n", resp)
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


## GetWorkloadsTelemetry

> TelemetryResponse GetWorkloadsTelemetry(ctx).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolName(nodepoolName).DepartmentId(departmentId).GroupBy(groupBy).Execute()

Get the workloads telemetry. [Experimental]



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
    telemetryType := openapiclient.WorkloadTelemetryType("WORKLOADS_COUNT") // WorkloadTelemetryType | Specifies the telemetry type.
    clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
    nodepoolName := "default" // string | Filter using the nodepool. (optional)
    departmentId := "1" // string | Filter using the department id. (optional)
    groupBy := []string{"GroupBy_example"} // []string | Group workloads by field. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.GetWorkloadsTelemetry(context.Background()).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolName(nodepoolName).DepartmentId(departmentId).GroupBy(groupBy).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.GetWorkloadsTelemetry``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkloadsTelemetry`: TelemetryResponse
    fmt.Fprintf(os.Stdout, "Response from `WorkloadsApi.GetWorkloadsTelemetry`: %v\n", resp)
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
 **groupBy** | **[]string** | Group workloads by field. | 

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
    openapiclient "./openapi"
)

func main() {
    workloadCreationFields := *openapiclient.NewWorkloadCreationFields("runai-job", "very-important-job", "Id_example", "71f69d83-ba66-4822-adf5-55ce55efd210", "1", time.Now()) // WorkloadCreationFields | The workload info in the control plane.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.InternalCreateWorkload(context.Background()).WorkloadCreationFields(workloadCreationFields).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.InternalCreateWorkload``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InternalCreateWorkload`: Workload
    fmt.Fprintf(os.Stdout, "Response from `WorkloadsApi.InternalCreateWorkload`: %v\n", resp)
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
    openapiclient "./openapi"
)

func main() {
    filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)
    internalWorkloadsSyncRequest := *openapiclient.NewInternalWorkloadsSyncRequest("71f69d83-ba66-4822-adf5-55ce55efd210", []openapiclient.WorkloadSyncFields{*openapiclient.NewWorkloadSyncFields("runai-job", "very-important-job", "Id_example", "high-priority", "71f69d83-ba66-4822-adf5-55ce55efd210", "proj-1", "1", "department-1", "1", "runai-proj-1", time.Now(), openapiclient.Phase("Creating"), openapiclient.Source("CLI"), "Pending")}) // InternalWorkloadsSyncRequest | The workloads to sync. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.InternalSyncWorkloads(context.Background()).FilterBy(filterBy).InternalWorkloadsSyncRequest(internalWorkloadsSyncRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.InternalSyncWorkloads``: %v\n", err)
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
    openapiclient "./openapi"
)

func main() {
    workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the workload.
    workloadPatchFields := *openapiclient.NewWorkloadPatchFields(openapiclient.Phase("Creating")) // WorkloadPatchFields | The workload info in the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.PatchWorkloadPhase(context.Background(), workloadId).WorkloadPatchFields(workloadPatchFields).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.PatchWorkloadPhase``: %v\n", err)
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
    openapiclient "./openapi"
)

func main() {
    filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)
    workloadsSyncRequest := *openapiclient.NewWorkloadsSyncRequest("1", []openapiclient.WorkloadSyncFields{*openapiclient.NewWorkloadSyncFields("runai-job", "very-important-job", "Id_example", "high-priority", "71f69d83-ba66-4822-adf5-55ce55efd210", "proj-1", "1", "department-1", "1", "runai-proj-1", time.Now(), openapiclient.Phase("Creating"), openapiclient.Source("CLI"), "Pending")}) // WorkloadsSyncRequest | The workloads to sync. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.SyncWorkloads(context.Background()).FilterBy(filterBy).WorkloadsSyncRequest(workloadsSyncRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.SyncWorkloads``: %v\n", err)
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
    openapiclient "./openapi"
)

func main() {
    workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the workload.
    workloadSyncFields := *openapiclient.NewWorkloadSyncFields("runai-job", "very-important-job", "Id_example", "high-priority", "71f69d83-ba66-4822-adf5-55ce55efd210", "proj-1", "1", "department-1", "1", "runai-proj-1", time.Now(), openapiclient.Phase("Creating"), openapiclient.Source("CLI"), "Pending") // WorkloadSyncFields | The workload info in the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.UpsertWorkload(context.Background(), workloadId).WorkloadSyncFields(workloadSyncFields).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.UpsertWorkload``: %v\n", err)
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

