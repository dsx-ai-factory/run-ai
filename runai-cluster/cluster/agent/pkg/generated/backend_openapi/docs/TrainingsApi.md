# \TrainingsApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteTraining**](TrainingsApi.md#DeleteTraining) | **Delete** /v1/k8s/clusters/{clusterId}/trainings/{id} | Delete a training
[**SyncTrainingsStatus**](TrainingsApi.md#SyncTrainingsStatus) | **Post** /v1/k8s/clusters/{clusterId}/trainings/status | Sync trainings status
[**UpdateTrainingStatus**](TrainingsApi.md#UpdateTrainingStatus) | **Put** /v1/k8s/clusters/{clusterId}/trainings/status | Update training status



## DeleteTraining

> DeleteTraining(ctx, clusterId, id).Execute()

Delete a training

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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TrainingsApi.DeleteTraining(context.Background(), clusterId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TrainingsApi.DeleteTraining``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteTrainingRequest struct via the builder pattern


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


## SyncTrainingsStatus

> SyncTrainingsStatus(ctx, clusterId).WorkloadStatus(workloadStatus).Execute()

Sync trainings status



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    workloadStatus := []openapiclient.WorkloadStatus{*openapiclient.NewWorkloadStatus("Id_example", "Status_example")} // []WorkloadStatus |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TrainingsApi.SyncTrainingsStatus(context.Background(), clusterId).WorkloadStatus(workloadStatus).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TrainingsApi.SyncTrainingsStatus``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncTrainingsStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadStatus** | [**[]WorkloadStatus**](WorkloadStatus.md) |  | 

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


## UpdateTrainingStatus

> UpdateTrainingStatus(ctx, clusterId).WorkloadStatus(workloadStatus).Execute()

Update training status

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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    workloadStatus := *openapiclient.NewWorkloadStatus("Id_example", "Status_example") // WorkloadStatus | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TrainingsApi.UpdateTrainingStatus(context.Background(), clusterId).WorkloadStatus(workloadStatus).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TrainingsApi.UpdateTrainingStatus``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateTrainingStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadStatus** | [**WorkloadStatus**](WorkloadStatus.md) |  | 

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

