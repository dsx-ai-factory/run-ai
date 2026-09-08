# \TemplateAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateTemplate**](TemplateAPI.md#CreateTemplate) | **Post** /api/v1/asset/workload-template | Create a template.
[**DeleteTemplateById**](TemplateAPI.md#DeleteTemplateById) | **Delete** /api/v1/asset/workload-template/{AssetId} | Delete a template.
[**GetTemplateById**](TemplateAPI.md#GetTemplateById) | **Get** /api/v1/asset/workload-template/{AssetId} | Get a template.
[**ListTemplates**](TemplateAPI.md#ListTemplates) | **Get** /api/v1/asset/workload-template | List templates.
[**UpdateTemplate**](TemplateAPI.md#UpdateTemplate) | **Put** /api/v1/asset/workload-template/{AssetId} | Update a template.



## CreateTemplate

> WorkloadTemplate CreateTemplate(ctx).WorkloadTemplateCreationRequest(workloadTemplateCreationRequest).Execute()

Create a template.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	workloadTemplateCreationRequest := *openapiclient.NewWorkloadTemplateCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("system")), *openapiclient.NewSpecificRunCreationFields(*openapiclient.NewAssetsIds("Environment_example"))) // WorkloadTemplateCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TemplateAPI.CreateTemplate(context.Background()).WorkloadTemplateCreationRequest(workloadTemplateCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TemplateAPI.CreateTemplate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateTemplate`: WorkloadTemplate
	fmt.Fprintf(os.Stdout, "Response from `TemplateAPI.CreateTemplate`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateTemplateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadTemplateCreationRequest** | [**WorkloadTemplateCreationRequest**](WorkloadTemplateCreationRequest.md) |  | 

### Return type

[**WorkloadTemplate**](WorkloadTemplate.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteTemplateById

> HttpResponse DeleteTemplateById(ctx, assetId).Execute()

Delete a template.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TemplateAPI.DeleteTemplateById(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TemplateAPI.DeleteTemplateById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteTemplateById`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `TemplateAPI.DeleteTemplateById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteTemplateByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HttpResponse**](HttpResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTemplateById

> WorkloadTemplate GetTemplateById(ctx, assetId).Execute()

Get a template.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TemplateAPI.GetTemplateById(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TemplateAPI.GetTemplateById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTemplateById`: WorkloadTemplate
	fmt.Fprintf(os.Stdout, "Response from `TemplateAPI.GetTemplateById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTemplateByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**WorkloadTemplate**](WorkloadTemplate.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListTemplates

> WorkloadTemplateListResponse ListTemplates(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).DistributedFramework(distributedFramework).IsDistributed(isDistributed).IsTraining(isTraining).IsWorkspace(isWorkspace).IsDistributedInference(isDistributedInference).Execute()

List templates.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	name := "name_example" // string | Filter results by name. (optional)
	scope := "scope_example" // string | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. (optional)
	projectId := int32(56) // int32 | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. (optional)
	departmentId := "1" // string | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. (optional)
	distributedFramework := "distributedFramework_example" // string | Filter results to workload of type distributed and distributedFramework. (optional)
	isDistributed := true // bool | Filter results to workload of type distributed. (optional)
	isTraining := true // bool | Filter results to workload of type training. (optional)
	isWorkspace := true // bool | Filter results to workload of type workspace. (optional)
	isDistributedInference := true // bool | Filter results to workload of type distributed inference. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TemplateAPI.ListTemplates(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).DistributedFramework(distributedFramework).IsDistributed(isDistributed).IsTraining(isTraining).IsWorkspace(isWorkspace).IsDistributedInference(isDistributedInference).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TemplateAPI.ListTemplates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListTemplates`: WorkloadTemplateListResponse
	fmt.Fprintf(os.Stdout, "Response from `TemplateAPI.ListTemplates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListTemplatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. | 
 **projectId** | **int32** | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. | 
 **departmentId** | **string** | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. | 
 **clusterId** | **string** | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. | 
 **distributedFramework** | **string** | Filter results to workload of type distributed and distributedFramework. | 
 **isDistributed** | **bool** | Filter results to workload of type distributed. | 
 **isTraining** | **bool** | Filter results to workload of type training. | 
 **isWorkspace** | **bool** | Filter results to workload of type workspace. | 
 **isDistributedInference** | **bool** | Filter results to workload of type distributed inference. | 

### Return type

[**WorkloadTemplateListResponse**](WorkloadTemplateListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateTemplate

> WorkloadTemplate UpdateTemplate(ctx, assetId).WorkloadTemplateUpdateRequest(workloadTemplateUpdateRequest).Execute()

Update a template.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	workloadTemplateUpdateRequest := *openapiclient.NewWorkloadTemplateUpdateRequest(*openapiclient.NewAssetUpdateRequest("my-asset"), *openapiclient.NewSpecificRunCreationFields(*openapiclient.NewAssetsIds("Environment_example"))) // WorkloadTemplateUpdateRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TemplateAPI.UpdateTemplate(context.Background(), assetId).WorkloadTemplateUpdateRequest(workloadTemplateUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TemplateAPI.UpdateTemplate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateTemplate`: WorkloadTemplate
	fmt.Fprintf(os.Stdout, "Response from `TemplateAPI.UpdateTemplate`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateTemplateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadTemplateUpdateRequest** | [**WorkloadTemplateUpdateRequest**](WorkloadTemplateUpdateRequest.md) |  | 

### Return type

[**WorkloadTemplate**](WorkloadTemplate.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

