# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**UpdateProjectStatus**](ProjectsApi.md#UpdateProjectStatus) | **Put** /v1/k8s/clusters/{clusterId}/projects/status | Update Project status
[**V1K8sClustersClusterIdProjectsGet**](ProjectsApi.md#V1K8sClustersClusterIdProjectsGet) | **Get** /v1/k8s/clusters/{clusterId}/projects | List all projects.
[**V1K8sClustersClusterIdProjectsIdDelete**](ProjectsApi.md#V1K8sClustersClusterIdProjectsIdDelete) | **Delete** /v1/k8s/clusters/{clusterId}/projects/{id} | Delete a project.
[**V1K8sClustersClusterIdProjectsIdGet**](ProjectsApi.md#V1K8sClustersClusterIdProjectsIdGet) | **Get** /v1/k8s/clusters/{clusterId}/projects/{id} | Get a specific project.
[**V1K8sClustersClusterIdProjectsIdPut**](ProjectsApi.md#V1K8sClustersClusterIdProjectsIdPut) | **Put** /v1/k8s/clusters/{clusterId}/projects/{id} | Update a project.
[**V1K8sClustersClusterIdProjectsPost**](ProjectsApi.md#V1K8sClustersClusterIdProjectsPost) | **Post** /v1/k8s/clusters/{clusterId}/projects | Create a new project.

# **UpdateProjectStatus**
> UpdateProjectStatus(ctx, body, clusterId)
Update Project status

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ProjectStatusRequest**](ProjectStatusRequest.md)|  | 
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterIdProjectsGet**
> []Project V1K8sClustersClusterIdProjectsGet(ctx, clusterId, optional)
List all projects.

List all the projects managed by the tenant on a specific cluster.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | **string**| The unique uuid identifying the cluster. | 
 **optional** | ***ProjectsApiV1K8sClustersClusterIdProjectsGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ProjectsApiV1K8sClustersClusterIdProjectsGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **excludePermissions** | **optional.Bool**| filter permissions | 

### Return type

[**[]Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterIdProjectsIdDelete**
> Project V1K8sClustersClusterIdProjectsIdDelete(ctx, clusterId, id)
Delete a project.

Delete a project from the cluster. This endpoint requires EDITOR or RESEARCH MANAGER  role.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | **string**| The unique uuid identifying the cluster. | 
  **id** | **int32**| The unique id identifying the project. | 

### Return type

[**Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterIdProjectsIdGet**
> Project V1K8sClustersClusterIdProjectsIdGet(ctx, clusterId, id, optional)
Get a specific project.

Get information about a specific project. This endpoint requires view permissions to the requested project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | **string**| The unique uuid identifying the cluster. | 
  **id** | **string**| The unique project-id identifying the project. | 
 **optional** | ***ProjectsApiV1K8sClustersClusterIdProjectsIdGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ProjectsApiV1K8sClustersClusterIdProjectsIdGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **excludePermissions** | **optional.Bool**| filter permissions | 

### Return type

[**Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterIdProjectsIdPut**
> Project V1K8sClustersClusterIdProjectsIdPut(ctx, body, id, clusterId, optional)
Update a project.

Updates project details. This endpoint requires EDITOR or RESEARCH MANAGER role.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Project**](Project.md)|  | 
  **id** | **string**| The unique project-id identifying the project. | 
  **clusterId** | **string**| The unique uuid identifying the cluster. | 
 **optional** | ***ProjectsApiV1K8sClustersClusterIdProjectsIdPutOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ProjectsApiV1K8sClustersClusterIdProjectsIdPutOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **excludePermissions** | **optional.**| filter permissions | 

### Return type

[**Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterIdProjectsPost**
> Project V1K8sClustersClusterIdProjectsPost(ctx, body, clusterId, optional)
Create a new project.

Create a new project in the cluster. This endpoint requires EDITOR or RESEARCH MANAGER role.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ProjectCreateRequest**](ProjectCreateRequest.md)|  | 
  **clusterId** | **string**| The unique uuid identifying the cluster. | 
 **optional** | ***ProjectsApiV1K8sClustersClusterIdProjectsPostOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ProjectsApiV1K8sClustersClusterIdProjectsPostOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **excludePermissions** | **optional.**| filter permissions | 

### Return type

[**Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

