# \WorkloadHistoryApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetWorkloadHistory**](WorkloadHistoryApi.md#GetWorkloadHistory) | **Get** /v1/k8s/clusters/{uuid}/jobs/{job_id}/workload-history | list workload history



## GetWorkloadHistory

> WorkloadHistoryResponse GetWorkloadHistory(ctx, uuid, jobId).Execute()

list workload history

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
    jobId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the job

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkloadHistoryApi.GetWorkloadHistory(context.Background(), uuid, jobId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkloadHistoryApi.GetWorkloadHistory``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkloadHistory`: WorkloadHistoryResponse
    fmt.Fprintf(os.Stdout, "Response from `WorkloadHistoryApi.GetWorkloadHistory`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**jobId** | **string** | Unique identifier of the job | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadHistoryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**WorkloadHistoryResponse**](WorkloadHistoryResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

