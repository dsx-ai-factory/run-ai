# \WorkspacesApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteWorkspace**](WorkspacesApi.md#DeleteWorkspace) | **Delete** /v1/k8s/clusters/{clusterId}/workspaces/{id} | Delete a workspace
[**SyncWorkspacesStatus**](WorkspacesApi.md#SyncWorkspacesStatus) | **Post** /v1/k8s/clusters/{clusterId}/workspaces/status | Sync workspaces status
[**UpdateWorkspaceStatus**](WorkspacesApi.md#UpdateWorkspaceStatus) | **Put** /v1/k8s/clusters/{clusterId}/workspaces/status | Update workspace status



## DeleteWorkspace

> DeleteWorkspace(ctx, clusterId, id).Execute()

Delete a workspace

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
    resp, r, err := apiClient.WorkspacesApi.DeleteWorkspace(context.Background(), clusterId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.DeleteWorkspace``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeleteWorkspaceRequest struct via the builder pattern


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


## SyncWorkspacesStatus

> SyncWorkspacesStatus(ctx, clusterId).WorkloadStatus(workloadStatus).Execute()

Sync workspaces status



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
    resp, r, err := apiClient.WorkspacesApi.SyncWorkspacesStatus(context.Background(), clusterId).WorkloadStatus(workloadStatus).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.SyncWorkspacesStatus``: %v\n", err)
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

Other parameters are passed through a pointer to a apiSyncWorkspacesStatusRequest struct via the builder pattern


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


## UpdateWorkspaceStatus

> UpdateWorkspaceStatus(ctx, clusterId).WorkloadStatus(workloadStatus).Execute()

Update workspace status

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
    resp, r, err := apiClient.WorkspacesApi.UpdateWorkspaceStatus(context.Background(), clusterId).WorkloadStatus(workloadStatus).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.UpdateWorkspaceStatus``: %v\n", err)
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

Other parameters are passed through a pointer to a apiUpdateWorkspaceStatusRequest struct via the builder pattern


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

