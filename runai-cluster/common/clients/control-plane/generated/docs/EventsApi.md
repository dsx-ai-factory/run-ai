# \EventsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetWorkloadEvents**](EventsApi.md#GetWorkloadEvents) | **Get** /api/v1/workloads/{workloadId}/events | Get the workload events.
[**GetWorkloadHistory**](EventsApi.md#GetWorkloadHistory) | **Get** /api/v1/workloads/{workloadId}/history | Get the workload history.
[**SyncEvents**](EventsApi.md#SyncEvents) | **Post** /api/v1/workloads/events | Sync events.



## GetWorkloadEvents

> GetWorkloadEvents200Response GetWorkloadEvents(ctx, workloadId).Offset(offset).Limit(limit).Execute()

Get the workload events.



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
    offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
    limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.EventsApi.GetWorkloadEvents(context.Background(), workloadId).Offset(offset).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `EventsApi.GetWorkloadEvents``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkloadEvents`: GetWorkloadEvents200Response
    fmt.Fprintf(os.Stdout, "Response from `EventsApi.GetWorkloadEvents`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadEventsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**GetWorkloadEvents200Response**](GetWorkloadEvents200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadHistory

> GetWorkloadHistory200Response GetWorkloadHistory(ctx, workloadId).Offset(offset).Limit(limit).Execute()

Get the workload history.



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
    offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
    limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.EventsApi.GetWorkloadHistory(context.Background(), workloadId).Offset(offset).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `EventsApi.GetWorkloadHistory``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkloadHistory`: GetWorkloadHistory200Response
    fmt.Fprintf(os.Stdout, "Response from `EventsApi.GetWorkloadHistory`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadHistoryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**GetWorkloadHistory200Response**](GetWorkloadHistory200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncEvents

> SyncEvents(ctx).EventsSyncRequest(eventsSyncRequest).Execute()

Sync events.

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
    eventsSyncRequest := *openapiclient.NewEventsSyncRequest([]openapiclient.EventSyncFields{*openapiclient.NewEventSyncFields("Id_example", "71f69d83-ba66-4822-adf5-55ce55efd210", "Normal", "Started container z", "Started", time.Now(), "kubelet", *openapiclient.NewInvolvedObject("Uid_example", "Pod", "test-0-1", "runai-test"))}) // EventsSyncRequest | The events to sync. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.EventsApi.SyncEvents(context.Background()).EventsSyncRequest(eventsSyncRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `EventsApi.SyncEvents``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncEventsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **eventsSyncRequest** | [**EventsSyncRequest**](EventsSyncRequest.md) | The events to sync. | 

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

