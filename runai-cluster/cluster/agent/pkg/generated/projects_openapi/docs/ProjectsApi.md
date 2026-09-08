# \ProjectsApi

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**UpdateProjectStatus**](ProjectsApi.md#UpdateProjectStatus) | **Put** /v1/k8s/clusters/{clusterId}/projects/status | Update Project status
[**V1K8sClustersClusterIdProjectsGet**](ProjectsApi.md#V1K8sClustersClusterIdProjectsGet) | **Get** /v1/k8s/clusters/{clusterId}/projects | List all projects.
[**V1K8sClustersClusterIdProjectsIdDelete**](ProjectsApi.md#V1K8sClustersClusterIdProjectsIdDelete) | **Delete** /v1/k8s/clusters/{clusterId}/projects/{id} | Delete a project.
[**V1K8sClustersClusterIdProjectsIdGet**](ProjectsApi.md#V1K8sClustersClusterIdProjectsIdGet) | **Get** /v1/k8s/clusters/{clusterId}/projects/{id} | Get a specific project.
[**V1K8sClustersClusterIdProjectsIdPut**](ProjectsApi.md#V1K8sClustersClusterIdProjectsIdPut) | **Put** /v1/k8s/clusters/{clusterId}/projects/{id} | Update a project.
[**V1K8sClustersClusterIdProjectsPost**](ProjectsApi.md#V1K8sClustersClusterIdProjectsPost) | **Post** /v1/k8s/clusters/{clusterId}/projects | Create a new project.



## UpdateProjectStatus

> UpdateProjectStatus(ctx, clusterId).ProjectStatusRequest(projectStatusRequest).Execute()

Update Project status

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
    projectStatusRequest := *openapiclient.NewProjectStatusRequest() // ProjectStatusRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.UpdateProjectStatus(context.Background(), clusterId).ProjectStatusRequest(projectStatusRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.UpdateProjectStatus``: %v\n", err)
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

Other parameters are passed through a pointer to a apiUpdateProjectStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **projectStatusRequest** | [**ProjectStatusRequest**](ProjectStatusRequest.md) |  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersClusterIdProjectsGet

> []Project V1K8sClustersClusterIdProjectsGet(ctx, clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()

List all projects.



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
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
    excludePermissions := true // bool | filter permissions (optional)
    memoryUnitMb := true // bool | should memory be returned in MB. default is \"false\", in which case it will be returned in MiB. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.V1K8sClustersClusterIdProjectsGet(context.Background(), clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.V1K8sClustersClusterIdProjectsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterIdProjectsGet`: []Project
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.V1K8sClustersClusterIdProjectsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterIdProjectsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **excludePermissions** | **bool** | filter permissions | 
 **memoryUnitMb** | **bool** | should memory be returned in MB. default is \&quot;false\&quot;, in which case it will be returned in MiB. | 

### Return type

[**[]Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersClusterIdProjectsIdDelete

> Project V1K8sClustersClusterIdProjectsIdDelete(ctx, clusterId, id).Execute()

Delete a project.



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
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
    id := int32(2) // int32 | The unique id identifying the project.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.V1K8sClustersClusterIdProjectsIdDelete(context.Background(), clusterId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.V1K8sClustersClusterIdProjectsIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterIdProjectsIdDelete`: Project
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.V1K8sClustersClusterIdProjectsIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 
**id** | **int32** | The unique id identifying the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterIdProjectsIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersClusterIdProjectsIdGet

> Project V1K8sClustersClusterIdProjectsIdGet(ctx, clusterId, id).ExcludePermissions(excludePermissions).Execute()

Get a specific project.



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
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
    id := "id_example" // string | The unique project-id identifying the project.
    excludePermissions := true // bool | filter permissions (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.V1K8sClustersClusterIdProjectsIdGet(context.Background(), clusterId, id).ExcludePermissions(excludePermissions).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.V1K8sClustersClusterIdProjectsIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterIdProjectsIdGet`: Project
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.V1K8sClustersClusterIdProjectsIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 
**id** | **string** | The unique project-id identifying the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterIdProjectsIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **excludePermissions** | **bool** | filter permissions | 

### Return type

[**Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersClusterIdProjectsIdPut

> Project V1K8sClustersClusterIdProjectsIdPut(ctx, id, clusterId).Project(project).ExcludePermissions(excludePermissions).Execute()

Update a project.



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
    id := "id_example" // string | The unique project-id identifying the project.
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
    project := *openapiclient.NewProject() // Project | 
    excludePermissions := true // bool | filter permissions (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.V1K8sClustersClusterIdProjectsIdPut(context.Background(), id, clusterId).Project(project).ExcludePermissions(excludePermissions).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.V1K8sClustersClusterIdProjectsIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterIdProjectsIdPut`: Project
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.V1K8sClustersClusterIdProjectsIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The unique project-id identifying the project. | 
**clusterId** | **string** | The unique uuid identifying the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterIdProjectsIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **project** | [**Project**](Project.md) |  | 
 **excludePermissions** | **bool** | filter permissions | 

### Return type

[**Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersClusterIdProjectsPost

> Project V1K8sClustersClusterIdProjectsPost(ctx, clusterId).ProjectCreateRequest(projectCreateRequest).ExcludePermissions(excludePermissions).Execute()

Create a new project.



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
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
    projectCreateRequest := *openapiclient.NewProjectCreateRequest() // ProjectCreateRequest | 
    excludePermissions := true // bool | filter permissions (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.V1K8sClustersClusterIdProjectsPost(context.Background(), clusterId).ProjectCreateRequest(projectCreateRequest).ExcludePermissions(excludePermissions).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.V1K8sClustersClusterIdProjectsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterIdProjectsPost`: Project
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.V1K8sClustersClusterIdProjectsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterIdProjectsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **projectCreateRequest** | [**ProjectCreateRequest**](ProjectCreateRequest.md) |  | 
 **excludePermissions** | **bool** | filter permissions | 

### Return type

[**Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

