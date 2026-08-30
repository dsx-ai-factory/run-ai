# \ProjectsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateProject**](ProjectsAPI.md#CreateProject) | **Post** /v1/k8s/clusters/{clusterId}/projects | Create a new project.
[**DeleteProject**](ProjectsAPI.md#DeleteProject) | **Delete** /v1/k8s/clusters/{clusterId}/projects/{id} | Delete a project.
[**GetProject**](ProjectsAPI.md#GetProject) | **Get** /v1/k8s/clusters/{clusterId}/projects/{id} | List details of a specific project.
[**GetProjects**](ProjectsAPI.md#GetProjects) | **Get** /v1/k8s/clusters/{clusterId}/projects | List all projects and their details.
[**UpdateProject**](ProjectsAPI.md#UpdateProject) | **Put** /v1/k8s/clusters/{clusterId}/projects/{id} | Update a project.
[**UpdateProjectStatus**](ProjectsAPI.md#UpdateProjectStatus) | **Put** /v1/k8s/clusters/{clusterId}/projects/status | Update Project status



## CreateProject

> Project1 CreateProject(ctx, clusterId).ProjectCreateRequest(projectCreateRequest).ExcludePermissions(excludePermissions).Execute()

Create a new project.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	projectCreateRequest := *openapiclient.NewProjectCreateRequest("team-a", int32(2)) // ProjectCreateRequest | 
	excludePermissions := true // bool | Backward compatibility of the `permissions` field. If `true`, the `permissions` field in the request body is ignored. If `false`, relevant access rules for the `permissions` field are created in the project scope.  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.CreateProject(context.Background(), clusterId).ProjectCreateRequest(projectCreateRequest).ExcludePermissions(excludePermissions).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.CreateProject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateProject`: Project1
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.CreateProject`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateProjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **projectCreateRequest** | [**ProjectCreateRequest**](ProjectCreateRequest.md) |  | 
 **excludePermissions** | **bool** | Backward compatibility of the &#x60;permissions&#x60; field. If &#x60;true&#x60;, the &#x60;permissions&#x60; field in the request body is ignored. If &#x60;false&#x60;, relevant access rules for the &#x60;permissions&#x60; field are created in the project scope.  | 

### Return type

[**Project1**](Project1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteProject

> Project1 DeleteProject(ctx, clusterId, id).Execute()

Delete a project.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	id := int32(2) // int32 | The unique id of the project.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.DeleteProject(context.Background(), clusterId, id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.DeleteProject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteProject`: Project1
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.DeleteProject`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**id** | **int32** | The unique id of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteProjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Project1**](Project1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProject

> Project1 GetProject(ctx, clusterId, id).ExcludePermissions(excludePermissions).Execute()

List details of a specific project.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	id := "id_example" // string | The unique project-id identifying the project.
	excludePermissions := true // bool | Backward compatibility of the 'permissions' field. If 'true', the 'permissions' field in the returned projects is not set. If 'false', the 'permissions' field is set in the returned projects. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.GetProject(context.Background(), clusterId, id).ExcludePermissions(excludePermissions).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.GetProject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProject`: Project1
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.GetProject`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**id** | **string** | The unique project-id identifying the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **excludePermissions** | **bool** | Backward compatibility of the &#39;permissions&#39; field. If &#39;true&#39;, the &#39;permissions&#39; field in the returned projects is not set. If &#39;false&#39;, the &#39;permissions&#39; field is set in the returned projects. | 

### Return type

[**Project1**](Project1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjects

> []Project1 GetProjects(ctx, clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()

List all projects and their details.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	excludePermissions := true // bool | Backward compatibility of the 'permissions' field. If 'true', the 'permissions' field in the returned projects is not set. If 'false', the 'permissions' field is set in the returned projects. (optional)
	memoryUnitMb := true // bool | Memory returned in MB. When set to `false` (default) memory will be returned in MiB. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.GetProjects(context.Background(), clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.GetProjects``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjects`: []Project1
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.GetProjects`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **excludePermissions** | **bool** | Backward compatibility of the &#39;permissions&#39; field. If &#39;true&#39;, the &#39;permissions&#39; field in the returned projects is not set. If &#39;false&#39;, the &#39;permissions&#39; field is set in the returned projects. | 
 **memoryUnitMb** | **bool** | Memory returned in MB. When set to &#x60;false&#x60; (default) memory will be returned in MiB. | 

### Return type

[**[]Project1**](Project1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateProject

> Project1 UpdateProject(ctx, id, clusterId).ProjectUpdateRequest(projectUpdateRequest).ExcludePermissions(excludePermissions).Execute()

Update a project.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	id := "id_example" // string | The unique project-id.
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	projectUpdateRequest := *openapiclient.NewProjectUpdateRequest() // ProjectUpdateRequest | 
	excludePermissions := true // bool | Backward compatibility of the `permissions` field. If `true`, the `permissions` field in the returned projects is not set. If `false`, the `permissions` field is set in the returned projects.  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.UpdateProject(context.Background(), id, clusterId).ProjectUpdateRequest(projectUpdateRequest).ExcludePermissions(excludePermissions).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.UpdateProject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateProject`: Project1
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.UpdateProject`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The unique project-id. | 
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateProjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **projectUpdateRequest** | [**ProjectUpdateRequest**](ProjectUpdateRequest.md) |  | 
 **excludePermissions** | **bool** | Backward compatibility of the &#x60;permissions&#x60; field. If &#x60;true&#x60;, the &#x60;permissions&#x60; field in the returned projects is not set. If &#x60;false&#x60;, the &#x60;permissions&#x60; field is set in the returned projects.  | 

### Return type

[**Project1**](Project1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


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
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
	projectStatusRequest := *openapiclient.NewProjectStatusRequest() // ProjectStatusRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProjectsAPI.UpdateProjectStatus(context.Background(), clusterId).ProjectStatusRequest(projectStatusRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.UpdateProjectStatus``: %v\n", err)
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

