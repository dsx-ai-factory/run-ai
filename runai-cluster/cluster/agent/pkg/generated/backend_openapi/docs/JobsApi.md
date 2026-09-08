# \JobsApi

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetClusterJobsCount**](JobsApi.md#GetClusterJobsCount) | **Get** /v1/k8s/clusters/{uuid}/jobs/count | Return the number all Jobs in the cluster
[**GetPodsByJobId**](JobsApi.md#GetPodsByJobId) | **Get** /v1/k8s/clusters/{uuid}/jobs/{jobId}/pods | Get all pods that are assoiciated for a specific job
[**ListJobs**](JobsApi.md#ListJobs) | **Get** /v1/k8s/clusters/{uuid}/jobs | List all Jobs in the cluster



## GetClusterJobsCount

> float32 GetClusterJobsCount(ctx, uuid).NodeId(nodeId).Filter(filter).Execute()

Return the number all Jobs in the cluster

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
    nodeId := "nodeId_example" // string | Unique identifier of the node. (optional)
    filter := "filter_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.JobsApi.GetClusterJobsCount(context.Background(), uuid).NodeId(nodeId).Filter(filter).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `JobsApi.GetClusterJobsCount``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetClusterJobsCount`: float32
    fmt.Fprintf(os.Stdout, "Response from `JobsApi.GetClusterJobsCount`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterJobsCountRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodeId** | **string** | Unique identifier of the node. | 
 **filter** | **string** |  | 

### Return type

**float32**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPodsByJobId

> []Pod GetPodsByJobId(ctx, jobId, uuid).Id(id).PodId(podId).PodGroupId(podGroupId).NodeID(nodeID).Name(name).Status(status).Execute()

Get all pods that are assoiciated for a specific job

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
    jobId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the job (UUID)
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    id := "id_example" // string | id (optional)
    podId := "podId_example" // string | Unique identifier of the pod. (optional)
    podGroupId := "podGroupId_example" // string | Identifier of the pod group. (optional)
    nodeID := "nodeID_example" // string | Unique identifier of the node. (optional)
    name := "name_example" // string | The name of the job. (optional)
    status := "status_example" // string | The status of the job. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.JobsApi.GetPodsByJobId(context.Background(), jobId, uuid).Id(id).PodId(podId).PodGroupId(podGroupId).NodeID(nodeID).Name(name).Status(status).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `JobsApi.GetPodsByJobId``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPodsByJobId`: []Pod
    fmt.Fprintf(os.Stdout, "Response from `JobsApi.GetPodsByJobId`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**jobId** | **string** | Unique identifier of the job (UUID) | 
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPodsByJobIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **id** | **string** | id | 
 **podId** | **string** | Unique identifier of the pod. | 
 **podGroupId** | **string** | Identifier of the pod group. | 
 **nodeID** | **string** | Unique identifier of the node. | 
 **name** | **string** | The name of the job. | 
 **status** | **string** | The status of the job. | 

### Return type

[**[]Pod**](Pod.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListJobs

> []DisplayedJob ListJobs(ctx, uuid).NodeId(nodeId).Filter(filter).SortBy(sortBy).SortDirection(sortDirection).From(from).Limit(limit).IncludeDeleted(includeDeleted).Execute()

List all Jobs in the cluster

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
    nodeId := "nodeId_example" // string | Unique identifier of the node. (optional)
    filter := "filter_example" // string |  (optional)
    sortBy := "sortBy_example" // string | Order of the results. (optional)
    sortDirection := "sortDirection_example" // string |  (optional)
    from := float32(8.14) // float32 | Start the response from a given number of result. Used along with 'limit' to retrieve the results paginated. (optional)
    limit := float32(8.14) // float32 | Limit the response to a given number of results. (optional)
    includeDeleted := true // bool | True to include deleted jobs in the result. (optional) (default to true)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.JobsApi.ListJobs(context.Background(), uuid).NodeId(nodeId).Filter(filter).SortBy(sortBy).SortDirection(sortDirection).From(from).Limit(limit).IncludeDeleted(includeDeleted).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `JobsApi.ListJobs``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListJobs`: []DisplayedJob
    fmt.Fprintf(os.Stdout, "Response from `JobsApi.ListJobs`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiListJobsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodeId** | **string** | Unique identifier of the node. | 
 **filter** | **string** |  | 
 **sortBy** | **string** | Order of the results. | 
 **sortDirection** | **string** |  | 
 **from** | **float32** | Start the response from a given number of result. Used along with &#39;limit&#39; to retrieve the results paginated. | 
 **limit** | **float32** | Limit the response to a given number of results. | 
 **includeDeleted** | **bool** | True to include deleted jobs in the result. | [default to true]

### Return type

[**[]DisplayedJob**](DisplayedJob.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

