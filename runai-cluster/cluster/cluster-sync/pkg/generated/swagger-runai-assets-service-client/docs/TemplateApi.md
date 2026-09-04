# \TemplateApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateTemplate**](TemplateApi.md#CreateTemplate) | **Post** /api/v1/asset/workload-template | Create template.
[**DeleteTemplateById**](TemplateApi.md#DeleteTemplateById) | **Delete** /api/v1/asset/workload-template/{AssetId} | Delete template.
[**GetTemplateById**](TemplateApi.md#GetTemplateById) | **Get** /api/v1/asset/workload-template/{AssetId} | Get workspace template.
[**ListTemplates**](TemplateApi.md#ListTemplates) | **Get** /api/v1/asset/workload-template | List templates.
[**UpdateTemplate**](TemplateApi.md#UpdateTemplate) | **Put** /api/v1/asset/workload-template/{AssetId} | Update template.



## CreateTemplate

> WorkloadTemplate CreateTemplate(ctx).WorkloadTemplateCreationRequest(workloadTemplateCreationRequest).Execute()

Create template.



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
    workloadTemplateCreationRequest := *openapiclient.NewWorkloadTemplateCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("runai")), *openapiclient.NewSpecificRunCreationFields(*openapiclient.NewAssetsIds("Environment_example"))) // WorkloadTemplateCreationRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TemplateApi.CreateTemplate(context.Background()).WorkloadTemplateCreationRequest(workloadTemplateCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TemplateApi.CreateTemplate``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateTemplate`: WorkloadTemplate
    fmt.Fprintf(os.Stdout, "Response from `TemplateApi.CreateTemplate`: %v\n", resp)
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

Delete template.



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
    assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TemplateApi.DeleteTemplateById(context.Background(), assetId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TemplateApi.DeleteTemplateById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteTemplateById`: HttpResponse
    fmt.Fprintf(os.Stdout, "Response from `TemplateApi.DeleteTemplateById`: %v\n", resp)
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

Get workspace template.



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
    assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TemplateApi.GetTemplateById(context.Background(), assetId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TemplateApi.GetTemplateById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetTemplateById`: WorkloadTemplate
    fmt.Fprintf(os.Stdout, "Response from `TemplateApi.GetTemplateById`: %v\n", resp)
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

> WorkloadTemplateListResponse ListTemplates(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).DistributedFramework(distributedFramework).IsDistributed(isDistributed).IsTraining(isTraining).IsWorkspace(isWorkspace).Execute()

List templates.



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
    scope := "scope_example" // string | Filter results by scope. (optional)
    projectId := int32(56) // int32 | Filter results by project id. if scope filter is project, only assets from the specific project will be included in the response. otherwise, the response will include both project, department, cluster and tenant assets. (optional)
    departmentId := "departmentId_example" // string | Filter results by department id. if scope filter is department, only assets from the specific department will be included in the response. otherwise, the response will include department, cluster and tenant scope assets. (optional)
    clusterId := "clusterId_example" // string | Filter results by cluster id. if scope filter is cluster, only assets from the specific cluster will be included in the response. otherwise, the response will include cluster and tenant scope assets. (optional)
    distributedFramework := "distributedFramework_example" // string | Filter results to workload of type distributed and distributedFramework. (optional)
    isDistributed := true // bool | Filter results to workload of type distributed. (optional)
    isTraining := true // bool | Filter results to workload of type training. (optional)
    isWorkspace := true // bool | Filter results to workload of type workspace. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TemplateApi.ListTemplates(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).DistributedFramework(distributedFramework).IsDistributed(isDistributed).IsTraining(isTraining).IsWorkspace(isWorkspace).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TemplateApi.ListTemplates``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListTemplates`: WorkloadTemplateListResponse
    fmt.Fprintf(os.Stdout, "Response from `TemplateApi.ListTemplates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListTemplatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filter results by scope. | 
 **projectId** | **int32** | Filter results by project id. if scope filter is project, only assets from the specific project will be included in the response. otherwise, the response will include both project, department, cluster and tenant assets. | 
 **departmentId** | **string** | Filter results by department id. if scope filter is department, only assets from the specific department will be included in the response. otherwise, the response will include department, cluster and tenant scope assets. | 
 **clusterId** | **string** | Filter results by cluster id. if scope filter is cluster, only assets from the specific cluster will be included in the response. otherwise, the response will include cluster and tenant scope assets. | 
 **distributedFramework** | **string** | Filter results to workload of type distributed and distributedFramework. | 
 **isDistributed** | **bool** | Filter results to workload of type distributed. | 
 **isTraining** | **bool** | Filter results to workload of type training. | 
 **isWorkspace** | **bool** | Filter results to workload of type workspace. | 

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

Update template.



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
    assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
    workloadTemplateUpdateRequest := *openapiclient.NewWorkloadTemplateUpdateRequest(*openapiclient.NewAssetUpdateRequest("my-asset"), *openapiclient.NewSpecificRunCreationFields(*openapiclient.NewAssetsIds("Environment_example"))) // WorkloadTemplateUpdateRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TemplateApi.UpdateTemplate(context.Background(), assetId).WorkloadTemplateUpdateRequest(workloadTemplateUpdateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TemplateApi.UpdateTemplate``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateTemplate`: WorkloadTemplate
    fmt.Fprintf(os.Stdout, "Response from `TemplateApi.UpdateTemplate`: %v\n", resp)
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

