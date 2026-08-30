# \WorkloadsApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CountWorkloads**](WorkloadsApi.md#CountWorkloads) | **Get** /api/v1/workloads/count | Count workloads. [Experimental]
[**DeleteWorkload**](WorkloadsApi.md#DeleteWorkload) | **Delete** /api/v1/workloads/{workloadId} | Delete the workload as is does not exist in the cluster anymore.
[**GetWorkload**](WorkloadsApi.md#GetWorkload) | **Get** /api/v1/workloads/{workloadId} | Get the workload data. [Experimental]
[**GetWorkloadAncestors**](WorkloadsApi.md#GetWorkloadAncestors) | **Get** /internal/workloads/{workloadId}/ancestors | Get the workload ancestors.
[**GetWorkloads**](WorkloadsApi.md#GetWorkloads) | **Get** /api/v1/workloads | List workloads. [Experimental]
[**InternalCreateWorkload**](WorkloadsApi.md#InternalCreateWorkload) | **Post** /internal/workloads | Create a workload by the control plane.
[**InternalSyncWorkloads**](WorkloadsApi.md#InternalSyncWorkloads) | **Post** /internal/workloads/sync | Sync workloads.
[**SyncWorkloads**](WorkloadsApi.md#SyncWorkloads) | **Post** /api/v1/workloads | Sync workloads.
[**UpsertWorkload**](WorkloadsApi.md#UpsertWorkload) | **Put** /api/v1/workloads/{workloadId} | Upsert the workload info from the cluster.



## CountWorkloads

> CountWorkloads200Response CountWorkloads(ctx).Deleted(deleted).FilterBy(filterBy).Execute()

Count workloads. [Experimental]

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
    deleted := true // bool | if true, return only deleted resources. (optional)
    filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.CountWorkloads(context.Background()).Deleted(deleted).FilterBy(filterBy).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsApi.CountWorkloads``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CountWorkloads`: CountWorkloads200Response
    fmt.Fprintf(os.Stdout, "Response from `WorkloadsApi.CountWorkloads`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountWorkloadsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deleted** | **bool** | if true, return only deleted resources. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 

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
    openapiclient "./openapi"
)

func main() {
    workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The requested workload id.

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
**workloadId** | **string** | The requested workload id. | 

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

Get the workload data. [Experimental]

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
**workloadId** | **string** | The requested workload id. | 

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
    workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The requested workload id.

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
**workloadId** | **string** | The requested workload id. | 

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


## GetWorkloads

> GetWorkloads200Response GetWorkloads(ctx).Deleted(deleted).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()

List workloads. [Experimental]

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
    deleted := true // bool | if true, return only deleted resources. (optional)
    offset := int32(56) // int32 | The offset from which to start fetching records. (optional)
    limit := int32(56) // int32 | The limit of returned records count. (optional)
    sortOrder := "sortOrder_example" // string | Sort results by a descending or ascending order. (optional) (default to "asc")
    sortBy := "sortBy_example" // string | Sort results by a parameters. (optional)
    filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)

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
 **deleted** | **bool** | if true, return only deleted resources. | 
 **offset** | **int32** | The offset from which to start fetching records. | 
 **limit** | **int32** | The limit of returned records count. | 
 **sortOrder** | **string** | Sort results by a descending or ascending order. | [default to &quot;asc&quot;]
 **sortBy** | **string** | Sort results by a parameters. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 

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
    workloadCreationFields := *openapiclient.NewWorkloadCreationFields("runai-job", "very-important-job", "Id_example", "ClusterId_example", "1", time.Now()) // WorkloadCreationFields | The workload info in the control plane.

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

> InternalSyncWorkloads(ctx).InternalWorkloadsSyncRequest(internalWorkloadsSyncRequest).Execute()

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
    internalWorkloadsSyncRequest := *openapiclient.NewInternalWorkloadsSyncRequest("ClusterId_example", []openapiclient.WorkloadSyncFields{*openapiclient.NewWorkloadSyncFields("runai-job", "very-important-job", "Id_example", "high-priority", "ClusterId_example", "proj-1", "1", "department-1", "1", "runai-proj-1", time.Now(), openapiclient.Phase("Creating"), openapiclient.Source("CLI"), "Pending")}) // InternalWorkloadsSyncRequest | The workloads to sync. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.InternalSyncWorkloads(context.Background()).InternalWorkloadsSyncRequest(internalWorkloadsSyncRequest).Execute()
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


## SyncWorkloads

> SyncWorkloads(ctx).WorkloadsSyncRequest(workloadsSyncRequest).Execute()

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
    workloadsSyncRequest := *openapiclient.NewWorkloadsSyncRequest("1", []openapiclient.WorkloadSyncFields{*openapiclient.NewWorkloadSyncFields("runai-job", "very-important-job", "Id_example", "high-priority", "ClusterId_example", "proj-1", "1", "department-1", "1", "runai-proj-1", time.Now(), openapiclient.Phase("Creating"), openapiclient.Source("CLI"), "Pending")}) // WorkloadsSyncRequest | The workloads to sync. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadsApi.SyncWorkloads(context.Background()).WorkloadsSyncRequest(workloadsSyncRequest).Execute()
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
    workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The requested workload id.
    workloadSyncFields := *openapiclient.NewWorkloadSyncFields("runai-job", "very-important-job", "Id_example", "high-priority", "ClusterId_example", "proj-1", "1", "department-1", "1", "runai-proj-1", time.Now(), openapiclient.Phase("Creating"), openapiclient.Source("CLI"), "Pending") // WorkloadSyncFields | The workload info in the cluster.

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
**workloadId** | **string** | The requested workload id. | 

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

