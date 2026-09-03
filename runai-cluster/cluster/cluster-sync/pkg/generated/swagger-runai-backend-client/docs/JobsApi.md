# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetClusterJobsCount**](JobsApi.md#GetClusterJobsCount) | **Get** /v1/k8s/clusters/{uuid}/jobs/count | Return the number all Jobs in the cluster
[**GetPodsByJobId**](JobsApi.md#GetPodsByJobId) | **Get** /v1/k8s/clusters/{uuid}/jobs/{jobId}/pods | Get all pods that are assoiciated for a specific job
[**ListJobs**](JobsApi.md#ListJobs) | **Get** /v1/k8s/clusters/{uuid}/jobs | List all Jobs in the cluster

# **GetClusterJobsCount**
> float64 GetClusterJobsCount(ctx, uuid, optional)
Return the number all Jobs in the cluster

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
 **optional** | ***JobsApiGetClusterJobsCountOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a JobsApiGetClusterJobsCountOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodeId** | **optional.String**| Unique identifier of the node. | 
 **filter** | **optional.String**|  | 

### Return type

**float64**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetPodsByJobId**
> []Pod GetPodsByJobId(ctx, jobId, uuid, optional)
Get all pods that are assoiciated for a specific job

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **jobId** | [**string**](.md)| Unique identifier of the job (UUID) | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
 **optional** | ***JobsApiGetPodsByJobIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a JobsApiGetPodsByJobIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **id** | **optional.String**| id | 
 **podId** | **optional.String**| Unique identifier of the pod. | 
 **podGroupId** | **optional.String**| Identifier of the pod group. | 
 **nodeID** | **optional.String**| Unique identifier of the node. | 
 **name** | **optional.String**| The name of the job. | 
 **status** | **optional.String**| The status of the job. | 

### Return type

[**[]Pod**](array.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **ListJobs**
> []DisplayedJob ListJobs(ctx, uuid, optional)
List all Jobs in the cluster

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
 **optional** | ***JobsApiListJobsOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a JobsApiListJobsOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodeId** | **optional.String**| Unique identifier of the node. | 
 **filter** | **optional.String**|  | 
 **sortBy** | **optional.String**| Order of the results. | 
 **sortDirection** | **optional.String**|  | 
 **from** | **optional.Float64**| Start the response from a given number of result. Used along with &#x27;limit&#x27; to retrieve the results paginated. | 
 **limit** | **optional.Float64**| Limit the response to a given number of results. | 
 **includeDeleted** | **optional.Bool**| True to include deleted jobs in the result. | [default to true]

### Return type

[**[]DisplayedJob**](array.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

