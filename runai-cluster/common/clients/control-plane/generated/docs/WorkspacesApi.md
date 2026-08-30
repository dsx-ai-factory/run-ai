# \WorkspacesApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CountAssets**](WorkspacesApi.md#CountAssets) | **Get** /api/v1/asset/count | Count assets.
[**CreateWorkspace**](WorkspacesApi.md#CreateWorkspace) | **Post** /api/v1/workspace | Create workspace.
[**CreateWorkspace1**](WorkspacesApi.md#CreateWorkspace1) | **Post** /api/v1/workloads/workspaces | Create a workspace [Experimental]
[**CreateWorkspaceV2**](WorkspacesApi.md#CreateWorkspaceV2) | **Post** /api/v2/workspace | Create a workspace - assets based [Experimental]
[**DeleteWorkspace**](WorkspacesApi.md#DeleteWorkspace) | **Delete** /api/v1/workloads/workspaces/{workloadId} | Delete a workspace [Experimental]
[**GetWorkspace**](WorkspacesApi.md#GetWorkspace) | **Get** /api/v1/workloads/workspaces/{workloadId} | Get workspace data [Experimental]
[**GetWorkspaceById**](WorkspacesApi.md#GetWorkspaceById) | **Get** /api/v1/workspace/{workloadId} | Get workspace content.
[**GetWorkspaceByIdV2**](WorkspacesApi.md#GetWorkspaceByIdV2) | **Get** /api/v2/workspace/{workloadId} | Get workspace data - assets based [Experimental]
[**ListWorkspaces**](WorkspacesApi.md#ListWorkspaces) | **Get** /v1/k8s/clusters/{clusterUuid}/workspaces | List workspaces in the cluster
[**ListWorkspacesUpdates**](WorkspacesApi.md#ListWorkspacesUpdates) | **Get** /api/v1/workloads/workspaces/updates | List workspaces updates
[**RequestWorkspacesUpdates**](WorkspacesApi.md#RequestWorkspacesUpdates) | **Post** /api/v1/workloads/workspaces/updates | Request workspaces updates
[**ResumeWorkspace**](WorkspacesApi.md#ResumeWorkspace) | **Post** /api/v1/workloads/workspaces/{workloadId}/resume | Resume a workspace [Experimental]
[**SuspendWorkspace**](WorkspacesApi.md#SuspendWorkspace) | **Post** /api/v1/workloads/workspaces/{workloadId}/suspend | Suspend a workspace [Experimental]



## CountAssets

> AssetCountResponse CountAssets(ctx).Kind(kind).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).Execute()

Count assets.



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
    kind := "kind_example" // string | Filter results by asset kind.
    name := "name_example" // string | Filter results by name. (optional)
    scope := "scope_example" // string | Filter results by scope. (optional)
    projectId := int32(56) // int32 | Filter results by project id. if scope filter is project, only assets from the specific project will be included in the response. otherwise, the response will include both project, department, cluster and tenant assets. (optional)
    departmentId := "1" // string | Filter using the department id. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkspacesApi.CountAssets(context.Background()).Kind(kind).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.CountAssets``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CountAssets`: AssetCountResponse
    fmt.Fprintf(os.Stdout, "Response from `WorkspacesApi.CountAssets`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountAssetsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kind** | **string** | Filter results by asset kind. | 
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filter results by scope. | 
 **projectId** | **int32** | Filter results by project id. if scope filter is project, only assets from the specific project will be included in the response. otherwise, the response will include both project, department, cluster and tenant assets. | 
 **departmentId** | **string** | Filter using the department id. | 

### Return type

[**AssetCountResponse**](AssetCountResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateWorkspace

> Workspace CreateWorkspace(ctx).WorkloadCreationRequest(workloadCreationRequest).Execute()

Create workspace.



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
    resp, r, err := apiClient.WorkspacesApi.CreateWorkspace(context.Background()).WorkloadCreationRequest(workloadCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.CreateWorkspace``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateWorkspace`: Workspace
    fmt.Fprintf(os.Stdout, "Response from `WorkspacesApi.CreateWorkspace`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateWorkspaceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadCreationRequest** | [**WorkloadCreationRequest**](WorkloadCreationRequest.md) |  | 

### Return type

[**Workspace**](Workspace.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateWorkspace1

> Workspace1 CreateWorkspace1(ctx).WorkspaceCreationRequest(workspaceCreationRequest).Execute()

Create a workspace [Experimental]



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
    workspaceCreationRequest := *openapiclient.NewWorkspaceCreationRequest("my-workload-name", "1", "71f69d83-ba66-4822-adf5-55ce55efd210") // WorkspaceCreationRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkspacesApi.CreateWorkspace1(context.Background()).WorkspaceCreationRequest(workspaceCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.CreateWorkspace1``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateWorkspace1`: Workspace1
    fmt.Fprintf(os.Stdout, "Response from `WorkspacesApi.CreateWorkspace1`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateWorkspace1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workspaceCreationRequest** | [**WorkspaceCreationRequest**](WorkspaceCreationRequest.md) |  | 

### Return type

[**Workspace1**](Workspace1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateWorkspaceV2

> WorkspaceV2 CreateWorkspaceV2(ctx).WorkspaceCreationRequestV2(workspaceCreationRequestV2).Execute()

Create a workspace - assets based [Experimental]



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
    workspaceCreationRequestV2 := *openapiclient.NewWorkspaceCreationRequestV2(*openapiclient.NewWorkloadCreationMeta("my-workload-name", "1", "71f69d83-ba66-4822-adf5-55ce55efd210"), *openapiclient.NewAssetsIds("Environment_example")) // WorkspaceCreationRequestV2 |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkspacesApi.CreateWorkspaceV2(context.Background()).WorkspaceCreationRequestV2(workspaceCreationRequestV2).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.CreateWorkspaceV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateWorkspaceV2`: WorkspaceV2
    fmt.Fprintf(os.Stdout, "Response from `WorkspacesApi.CreateWorkspaceV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateWorkspaceV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workspaceCreationRequestV2** | [**WorkspaceCreationRequestV2**](WorkspaceCreationRequestV2.md) |  | 

### Return type

[**WorkspaceV2**](WorkspaceV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteWorkspace

> DeleteWorkspace(ctx, workloadId).Execute()

Delete a workspace [Experimental]



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
    resp, r, err := apiClient.WorkspacesApi.DeleteWorkspace(context.Background(), workloadId).Execute()
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
**workloadId** | **string** | Unique identifier of the workload. | 

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


## GetWorkspace

> Workspace1 GetWorkspace(ctx, workloadId).Execute()

Get workspace data [Experimental]



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
    resp, r, err := apiClient.WorkspacesApi.GetWorkspace(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.GetWorkspace``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkspace`: Workspace1
    fmt.Fprintf(os.Stdout, "Response from `WorkspacesApi.GetWorkspace`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkspaceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Workspace1**](Workspace1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkspaceById

> Workspace GetWorkspaceById(ctx, workloadId).Execute()

Get workspace content.



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
    resp, r, err := apiClient.WorkspacesApi.GetWorkspaceById(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.GetWorkspaceById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkspaceById`: Workspace
    fmt.Fprintf(os.Stdout, "Response from `WorkspacesApi.GetWorkspaceById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkspaceByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Workspace**](Workspace.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkspaceByIdV2

> WorkspaceV2 GetWorkspaceByIdV2(ctx, workloadId).Execute()

Get workspace data - assets based [Experimental]



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
    resp, r, err := apiClient.WorkspacesApi.GetWorkspaceByIdV2(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.GetWorkspaceByIdV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkspaceByIdV2`: WorkspaceV2
    fmt.Fprintf(os.Stdout, "Response from `WorkspacesApi.GetWorkspaceByIdV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkspaceByIdV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**WorkspaceV2**](WorkspaceV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListWorkspaces

> WorkspaceListResponse ListWorkspaces(ctx, clusterUuid).NodeId(nodeId).Filter(filter).SortBy(sortBy).SortDirection(sortDirection).From(from).Limit(limit).Execute()

List workspaces in the cluster

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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    nodeId := "nodeId_example" // string | Unique identifier of the node. (optional)
    filter := "filter_example" // string |  (optional)
    sortBy := "sortBy_example" // string | Order of the results. (optional)
    sortDirection := "sortDirection_example" // string |  (optional)
    from := int32(56) // int32 | Start the response from a given number of result. Used along with 'limit' to retrieve the results paginated. (optional)
    limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkspacesApi.ListWorkspaces(context.Background(), clusterUuid).NodeId(nodeId).Filter(filter).SortBy(sortBy).SortDirection(sortDirection).From(from).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.ListWorkspaces``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListWorkspaces`: WorkspaceListResponse
    fmt.Fprintf(os.Stdout, "Response from `WorkspacesApi.ListWorkspaces`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiListWorkspacesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodeId** | **string** | Unique identifier of the node. | 
 **filter** | **string** |  | 
 **sortBy** | **string** | Order of the results. | 
 **sortDirection** | **string** |  | 
 **from** | **int32** | Start the response from a given number of result. Used along with &#39;limit&#39; to retrieve the results paginated. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**WorkspaceListResponse**](WorkspaceListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListWorkspacesUpdates

> WorkloadsSyncResponse ListWorkspacesUpdates(ctx).UpdatesAfter(updatesAfter).ClusterId(clusterId).Execute()

List workspaces updates



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
    resp, r, err := apiClient.WorkspacesApi.ListWorkspacesUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.ListWorkspacesUpdates``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListWorkspacesUpdates`: WorkloadsSyncResponse
    fmt.Fprintf(os.Stdout, "Response from `WorkspacesApi.ListWorkspacesUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListWorkspacesUpdatesRequest struct via the builder pattern


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


## RequestWorkspacesUpdates

> WorkloadsSyncResponse RequestWorkspacesUpdates(ctx).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()

Request workspaces updates



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
    resp, r, err := apiClient.WorkspacesApi.RequestWorkspacesUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.RequestWorkspacesUpdates``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `RequestWorkspacesUpdates`: WorkloadsSyncResponse
    fmt.Fprintf(os.Stdout, "Response from `WorkspacesApi.RequestWorkspacesUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRequestWorkspacesUpdatesRequest struct via the builder pattern


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


## ResumeWorkspace

> HttpResponse2 ResumeWorkspace(ctx, workloadId).Execute()

Resume a workspace [Experimental]



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
    resp, r, err := apiClient.WorkspacesApi.ResumeWorkspace(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.ResumeWorkspace``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ResumeWorkspace`: HttpResponse2
    fmt.Fprintf(os.Stdout, "Response from `WorkspacesApi.ResumeWorkspace`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiResumeWorkspaceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HttpResponse2**](HttpResponse2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SuspendWorkspace

> HttpResponse2 SuspendWorkspace(ctx, workloadId).Execute()

Suspend a workspace [Experimental]



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
    resp, r, err := apiClient.WorkspacesApi.SuspendWorkspace(context.Background(), workloadId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspacesApi.SuspendWorkspace``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SuspendWorkspace`: HttpResponse2
    fmt.Fprintf(os.Stdout, "Response from `WorkspacesApi.SuspendWorkspace`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | Unique identifier of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSuspendWorkspaceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HttpResponse2**](HttpResponse2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

