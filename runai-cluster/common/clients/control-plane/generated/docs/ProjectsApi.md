# \ProjectsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateProject**](ProjectsApi.md#CreateProject) | **Post** /api/v1/org-unit/projects | Create project
[**CreateProject_0**](ProjectsApi.md#CreateProject_0) | **Post** /v1/k8s/clusters/{clusterId}/projects | Create a new project.
[**DeleteProject**](ProjectsApi.md#DeleteProject) | **Delete** /api/v1/org-unit/projects/{projectId} | Delete project
[**DeleteProject_0**](ProjectsApi.md#DeleteProject_0) | **Delete** /v1/k8s/clusters/{clusterId}/projects/{id} | Delete a project.
[**GetProject**](ProjectsApi.md#GetProject) | **Get** /api/v1/org-unit/projects/{projectId} | Get project
[**GetProjectAncestors**](ProjectsApi.md#GetProjectAncestors) | **Get** /internal/org-unit/projects/{projectId}/ancestors | Get project ancestors
[**GetProjectMetrics**](ProjectsApi.md#GetProjectMetrics) | **Get** /v1/k8s/clusters/{clusterUuid}/projects/{projectId}/metrics | Get metrics data for a specific project.
[**GetProject_0**](ProjectsApi.md#GetProject_0) | **Get** /v1/k8s/clusters/{clusterId}/projects/{id} | List details of a specific project.
[**GetProjects**](ProjectsApi.md#GetProjects) | **Get** /api/v1/org-unit/projects | Get projects
[**GetProjectsMetrics**](ProjectsApi.md#GetProjectsMetrics) | **Get** /v1/k8s/clusters/{clusterUuid}/projects/metrics | Get metrics data for all projects.
[**GetProjectsTelemetry**](ProjectsApi.md#GetProjectsTelemetry) | **Get** /api/v1/org-unit/projects/telemetry | Get projects telemetry
[**GetProjects_0**](ProjectsApi.md#GetProjects_0) | **Get** /v1/k8s/clusters/{clusterId}/projects | List all projects and their details.
[**PatchProjectResources**](ProjectsApi.md#PatchProjectResources) | **Patch** /api/v1/org-unit/projects/{projectId}/resources | Patch project resources
[**UpdateProject**](ProjectsApi.md#UpdateProject) | **Put** /api/v1/org-unit/projects/{projectId} | Update project
[**UpdateProjectResources**](ProjectsApi.md#UpdateProjectResources) | **Put** /api/v1/org-unit/projects/{projectId}/resources | Update project resources
[**UpdateProjectStatus**](ProjectsApi.md#UpdateProjectStatus) | **Put** /api/v1/org-unit/projects/status | Update project status from cluster
[**UpdateProjectStatus_0**](ProjectsApi.md#UpdateProjectStatus_0) | **Put** /v1/k8s/clusters/{clusterId}/projects/status | Update Project status
[**UpdateProject_0**](ProjectsApi.md#UpdateProject_0) | **Put** /v1/k8s/clusters/{clusterId}/projects/{id} | Update a project.



## CreateProject

> Project CreateProject(ctx).ProjectCreationRequest(projectCreationRequest).Execute()

Create project



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
    projectCreationRequest := *openapiclient.NewProjectCreationRequest([]openapiclient.Resources{*openapiclient.NewResources(*openapiclient.NewResourcesGpu(float32(1000)))}, "organization1", "71f69d83-ba66-4822-adf5-55ce55efd210") // ProjectCreationRequest | Project to create.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.CreateProject(context.Background()).ProjectCreationRequest(projectCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.CreateProject``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateProject`: Project
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.CreateProject`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateProjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectCreationRequest** | [**ProjectCreationRequest**](ProjectCreationRequest.md) | Project to create. | 

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


## CreateProject_0

> Project2 CreateProject_0(ctx, clusterId).ProjectCreateRequest(projectCreateRequest).ExcludePermissions(excludePermissions).Execute()

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
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    projectCreateRequest := *openapiclient.NewProjectCreateRequest("team-a", int32(2)) // ProjectCreateRequest | 
    excludePermissions := true // bool | Backward compatibility of the `permissions` field. If `true`, the `permissions` field in the request body is ignored. If `false`, relevant access rules for the `permissions` field are created in the project scope.  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.CreateProject_0(context.Background(), clusterId).ProjectCreateRequest(projectCreateRequest).ExcludePermissions(excludePermissions).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.CreateProject_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateProject_0`: Project2
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.CreateProject_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateProject_1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **projectCreateRequest** | [**ProjectCreateRequest**](ProjectCreateRequest.md) |  | 
 **excludePermissions** | **bool** | Backward compatibility of the &#x60;permissions&#x60; field. If &#x60;true&#x60;, the &#x60;permissions&#x60; field in the request body is ignored. If &#x60;false&#x60;, relevant access rules for the &#x60;permissions&#x60; field are created in the project scope.  | 

### Return type

[**Project2**](Project2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteProject

> DeleteProject(ctx, projectId).Execute()

Delete project



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
    projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.DeleteProject(context.Background(), projectId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.DeleteProject``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteProjectRequest struct via the builder pattern


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


## DeleteProject_0

> Project2 DeleteProject_0(ctx, clusterId, id).Execute()

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
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    id := int32(2) // int32 | The unique id of the project.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.DeleteProject_0(context.Background(), clusterId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.DeleteProject_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteProject_0`: Project2
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.DeleteProject_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**id** | **int32** | The unique id of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteProject_2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Project2**](Project2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProject

> GetProjects200ResponseProjectsInner GetProject(ctx, projectId).Execute()

Get project



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
    projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.GetProject(context.Background(), projectId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.GetProject``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetProject`: GetProjects200ResponseProjectsInner
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.GetProject`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GetProjects200ResponseProjectsInner**](GetProjects200ResponseProjectsInner.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectAncestors

> Ancestors GetProjectAncestors(ctx, projectId).Execute()

Get project ancestors

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
    projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.GetProjectAncestors(context.Background(), projectId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.GetProjectAncestors``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetProjectAncestors`: Ancestors
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.GetProjectAncestors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectAncestorsRequest struct via the builder pattern


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


## GetProjectMetrics

> Project GetProjectMetrics(ctx, clusterUuid, projectId).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()

Get metrics data for a specific project.



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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    projectId := "projectId_example" // string | The unique project-id of the project.
    start := time.Now() // time.Time | Start of time range to fetch data from in UTC format. (optional)
    end := time.Now() // time.Time | End of time range to fetch data from in UTC format. (optional)
    numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)
    nodepoolName := "default" // string | Filter by unique nodepool name. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.GetProjectMetrics(context.Background(), clusterUuid, projectId).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.GetProjectMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetProjectMetrics`: Project
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.GetProjectMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**projectId** | **string** | The unique project-id of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **start** | **time.Time** | Start of time range to fetch data from in UTC format. | 
 **end** | **time.Time** | End of time range to fetch data from in UTC format. | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]
 **nodepoolName** | **string** | Filter by unique nodepool name. | 

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


## GetProject_0

> Project2 GetProject_0(ctx, clusterId, id).ExcludePermissions(excludePermissions).Execute()

List details of a specific project.



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
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    id := "id_example" // string | The unique project-id identifying the project.
    excludePermissions := true // bool | Backward compatibility of the 'permissions' field. If 'true', the 'permissions' field in the returned projects is not set. If 'false', the 'permissions' field is set in the returned projects. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.GetProject_0(context.Background(), clusterId, id).ExcludePermissions(excludePermissions).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.GetProject_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetProject_0`: Project2
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.GetProject_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**id** | **string** | The unique project-id identifying the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProject_3Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **excludePermissions** | **bool** | Backward compatibility of the &#39;permissions&#39; field. If &#39;true&#39;, the &#39;permissions&#39; field in the returned projects is not set. If &#39;false&#39;, the &#39;permissions&#39; field is set in the returned projects. | 

### Return type

[**Project2**](Project2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjects

> GetProjects200Response GetProjects(ctx).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Execute()

Get projects



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
    filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
    sortBy := "sortBy_example" // string | Sort results by a parameters. (optional)
    sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
    offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
    limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.GetProjects(context.Background()).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.GetProjects``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetProjects`: GetProjects200Response
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.GetProjects`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **sortBy** | **string** | Sort results by a parameters. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**GetProjects200Response**](GetProjects200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectsMetrics

> Projects GetProjectsMetrics(ctx, clusterUuid).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()

Get metrics data for all projects.



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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    start := time.Now() // time.Time | Start of time range to fetch data from in UTC format. (optional)
    end := time.Now() // time.Time | End of time range to fetch data from in UTC format. (optional)
    numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)
    nodepoolName := "default" // string | Filter by unique nodepool name. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.GetProjectsMetrics(context.Background(), clusterUuid).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.GetProjectsMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetProjectsMetrics`: Projects
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.GetProjectsMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectsMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **start** | **time.Time** | Start of time range to fetch data from in UTC format. | 
 **end** | **time.Time** | End of time range to fetch data from in UTC format. | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]
 **nodepoolName** | **string** | Filter by unique nodepool name. | 

### Return type

[**Projects**](Projects.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectsTelemetry

> TelemetryResponse GetProjectsTelemetry(ctx).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolName(nodepoolName).DepartmentId(departmentId).ProjectId(projectId).GroupBy(groupBy).Execute()

Get projects telemetry



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
    telemetryType := "telemetryType_example" // string | specifies what data to request
    clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
    nodepoolName := "default" // string | Filter using the nodepool. (optional)
    departmentId := "1" // string | Filter using the department id. (optional)
    projectId := "1" // string | project id to filter by (optional)
    groupBy := []string{"GroupBy_example"} // []string | project fields to group the data by (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.GetProjectsTelemetry(context.Background()).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolName(nodepoolName).DepartmentId(departmentId).ProjectId(projectId).GroupBy(groupBy).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.GetProjectsTelemetry``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetProjectsTelemetry`: TelemetryResponse
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.GetProjectsTelemetry`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectsTelemetryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **telemetryType** | **string** | specifies what data to request | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **nodepoolName** | **string** | Filter using the nodepool. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **groupBy** | **[]string** | project fields to group the data by | 

### Return type

[**TelemetryResponse**](TelemetryResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjects_0

> []Project2 GetProjects_0(ctx, clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()

List all projects and their details.



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
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    excludePermissions := true // bool | Backward compatibility of the 'permissions' field. If 'true', the 'permissions' field in the returned projects is not set. If 'false', the 'permissions' field is set in the returned projects. (optional)
    memoryUnitMb := true // bool | Memory returned in MB. When set to `false` (default) memory will be returned in MiB. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.GetProjects_0(context.Background(), clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.GetProjects_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetProjects_0`: []Project2
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.GetProjects_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjects_4Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **excludePermissions** | **bool** | Backward compatibility of the &#39;permissions&#39; field. If &#39;true&#39;, the &#39;permissions&#39; field in the returned projects is not set. If &#39;false&#39;, the &#39;permissions&#39; field is set in the returned projects. | 
 **memoryUnitMb** | **bool** | Memory returned in MB. When set to &#x60;false&#x60; (default) memory will be returned in MiB. | 

### Return type

[**[]Project2**](Project2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PatchProjectResources

> []Resources PatchProjectResources(ctx, projectId).Resources(resources).Execute()

Patch project resources



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
    projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id
    resources := []openapiclient.Resources{*openapiclient.NewResources(*openapiclient.NewResourcesGpu(float32(1000)))} // []Resources | Project resources to patch.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.PatchProjectResources(context.Background(), projectId).Resources(resources).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.PatchProjectResources``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `PatchProjectResources`: []Resources
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.PatchProjectResources`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchProjectResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **resources** | [**[]Resources**](Resources.md) | Project resources to patch. | 

### Return type

[**[]Resources**](Resources.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateProject

> Project UpdateProject(ctx, projectId).ProjectUpdateRequest(projectUpdateRequest).Execute()

Update project



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
    projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id
    projectUpdateRequest := *openapiclient.NewProjectUpdateRequest([]openapiclient.Resources{*openapiclient.NewResources(*openapiclient.NewResourcesGpu(float32(1000)))}) // ProjectUpdateRequest | Project to update.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.UpdateProject(context.Background(), projectId).ProjectUpdateRequest(projectUpdateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.UpdateProject``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateProject`: Project
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.UpdateProject`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateProjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **projectUpdateRequest** | [**ProjectUpdateRequest**](ProjectUpdateRequest.md) | Project to update. | 

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


## UpdateProjectResources

> []Resources UpdateProjectResources(ctx, projectId).Resources(resources).Execute()

Update project resources



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
    projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id
    resources := []openapiclient.Resources{*openapiclient.NewResources(*openapiclient.NewResourcesGpu(float32(1000)))} // []Resources | Project resources to update.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.UpdateProjectResources(context.Background(), projectId).Resources(resources).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.UpdateProjectResources``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateProjectResources`: []Resources
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.UpdateProjectResources`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateProjectResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **resources** | [**[]Resources**](Resources.md) | Project resources to update. | 

### Return type

[**[]Resources**](Resources.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateProjectStatus

> UpdateProjectStatus(ctx).ProjectStatusUpdateRequest(projectStatusUpdateRequest).Execute()

Update project status from cluster

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
    projectStatusUpdateRequest := *openapiclient.NewProjectStatusUpdateRequest("project1", "71f69d83-ba66-4822-adf5-55ce55efd210", "{"namespace":"runai-team-a"}") // ProjectStatusUpdateRequest | Project to update by name

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.UpdateProjectStatus(context.Background()).ProjectStatusUpdateRequest(projectStatusUpdateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.UpdateProjectStatus``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateProjectStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectStatusUpdateRequest** | [**ProjectStatusUpdateRequest**](ProjectStatusUpdateRequest.md) | Project to update by name | 

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


## UpdateProjectStatus_0

> UpdateProjectStatus_0(ctx, clusterId).ProjectStatusRequest(projectStatusRequest).Execute()

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
    resp, r, err := apiClient.ProjectsApi.UpdateProjectStatus_0(context.Background(), clusterId).ProjectStatusRequest(projectStatusRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.UpdateProjectStatus_0``: %v\n", err)
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

Other parameters are passed through a pointer to a apiUpdateProjectStatus_5Request struct via the builder pattern


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


## UpdateProject_0

> Project2 UpdateProject_0(ctx, id, clusterId).ProjectUpdateRequest1(projectUpdateRequest1).ExcludePermissions(excludePermissions).Execute()

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
    id := "id_example" // string | The unique project-id.
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    projectUpdateRequest1 := *openapiclient.NewProjectUpdateRequest1() // ProjectUpdateRequest1 | 
    excludePermissions := true // bool | Backward compatibility of the `permissions` field. If `true`, the `permissions` field in the returned projects is not set. If `false`, the `permissions` field is set in the returned projects.  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ProjectsApi.UpdateProject_0(context.Background(), id, clusterId).ProjectUpdateRequest1(projectUpdateRequest1).ExcludePermissions(excludePermissions).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ProjectsApi.UpdateProject_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateProject_0`: Project2
    fmt.Fprintf(os.Stdout, "Response from `ProjectsApi.UpdateProject_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The unique project-id. | 
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateProject_6Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **projectUpdateRequest1** | [**ProjectUpdateRequest1**](ProjectUpdateRequest1.md) |  | 
 **excludePermissions** | **bool** | Backward compatibility of the &#x60;permissions&#x60; field. If &#x60;true&#x60;, the &#x60;permissions&#x60; field in the returned projects is not set. If &#x60;false&#x60;, the &#x60;permissions&#x60; field is set in the returned projects.  | 

### Return type

[**Project2**](Project2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

