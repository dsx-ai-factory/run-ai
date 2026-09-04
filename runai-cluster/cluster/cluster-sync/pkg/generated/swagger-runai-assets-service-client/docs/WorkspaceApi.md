# \WorkspaceApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CountAssets**](WorkspaceApi.md#CountAssets) | **Get** /api/v1/asset/count | Count assets.
[**CountWorkspaces**](WorkspaceApi.md#CountWorkspaces) | **Get** /api/v1/workspace/count | Count workspaces.
[**CreateWorkspace**](WorkspaceApi.md#CreateWorkspace) | **Post** /api/v1/workspace | Create workspace.
[**GetWorkspaceById**](WorkspaceApi.md#GetWorkspaceById) | **Get** /api/v1/workspace/{id} | Get workspace content.
[**ListWorkspaces**](WorkspaceApi.md#ListWorkspaces) | **Get** /v1/k8s/clusters/{clusterUuid}/workspaces | List workspaces in the cluster



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
    departmentId := "departmentId_example" // string | Filter results by department id. if scope filter is department, only assets from the specific department will be included in the response. otherwise, the response will include department, cluster and tenant scope assets. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkspaceApi.CountAssets(context.Background()).Kind(kind).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceApi.CountAssets``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CountAssets`: AssetCountResponse
    fmt.Fprintf(os.Stdout, "Response from `WorkspaceApi.CountAssets`: %v\n", resp)
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
 **departmentId** | **string** | Filter results by department id. if scope filter is department, only assets from the specific department will be included in the response. otherwise, the response will include department, cluster and tenant scope assets. | 

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


## CountWorkspaces

> AssetCountResponse CountWorkspaces(ctx).Name(name).ProjectId(projectId).Execute()

Count workspaces.



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
    name := "name_example" // string | Filter results by name. (optional)
    projectId := int32(56) // int32 | Filter results by project id. if scope filter is project, only assets from the specific project will be included in the response. otherwise, the response will include both project, department, cluster and tenant assets. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkspaceApi.CountWorkspaces(context.Background()).Name(name).ProjectId(projectId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceApi.CountWorkspaces``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CountWorkspaces`: AssetCountResponse
    fmt.Fprintf(os.Stdout, "Response from `WorkspaceApi.CountWorkspaces`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountWorkspacesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **projectId** | **int32** | Filter results by project id. if scope filter is project, only assets from the specific project will be included in the response. otherwise, the response will include both project, department, cluster and tenant assets. | 

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
    workloadCreationRequest := *openapiclient.NewWorkloadCreationRequest("Name_example", "ClusterId_example", int32(123), "Namespace_example", *openapiclient.NewAssetsIds("Environment_example")) // WorkloadCreationRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkspaceApi.CreateWorkspace(context.Background()).WorkloadCreationRequest(workloadCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceApi.CreateWorkspace``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateWorkspace`: Workspace
    fmt.Fprintf(os.Stdout, "Response from `WorkspaceApi.CreateWorkspace`: %v\n", resp)
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


## GetWorkspaceById

> Workspace GetWorkspaceById(ctx, id).Execute()

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
    id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the workspace.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkspaceApi.GetWorkspaceById(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceApi.GetWorkspaceById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetWorkspaceById`: Workspace
    fmt.Fprintf(os.Stdout, "Response from `WorkspaceApi.GetWorkspaceById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Unique identifier of the workspace. | 

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
    clusterUuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster.
    nodeId := "nodeId_example" // string | Unique identifier of the node. (optional)
    filter := "filter_example" // string |  (optional)
    sortBy := "sortBy_example" // string | Order of the results. (optional)
    sortDirection := "sortDirection_example" // string |  (optional)
    from := int32(56) // int32 | Start the response from a given number of result. Used along with 'limit' to retrieve the results paginated. (optional)
    limit := int32(56) // int32 | Limit the response to a given number of results. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.WorkspaceApi.ListWorkspaces(context.Background(), clusterUuid).NodeId(nodeId).Filter(filter).SortBy(sortBy).SortDirection(sortDirection).From(from).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceApi.ListWorkspaces``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListWorkspaces`: WorkspaceListResponse
    fmt.Fprintf(os.Stdout, "Response from `WorkspaceApi.ListWorkspaces`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiListWorkspacesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodeId** | **string** | Unique identifier of the node. | 
 **filter** | **string** |  | 
 **sortBy** | **string** | Order of the results. | 
 **sortDirection** | **string** |  | 
 **from** | **int32** | Start the response from a given number of result. Used along with &#39;limit&#39; to retrieve the results paginated. | 
 **limit** | **int32** | Limit the response to a given number of results. | 

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

