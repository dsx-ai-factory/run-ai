# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateDeployment**](DeploymentsApi.md#CreateDeployment) | **Post** /v1/k8s/clusters/{uuid}/deployments | Create a deployment
[**DeleteDeployment**](DeploymentsApi.md#DeleteDeployment) | **Delete** /v1/k8s/clusters/{uuid}/deployments/{deploymentId} | Delete a deployment by id
[**GetDeployment**](DeploymentsApi.md#GetDeployment) | **Get** /v1/k8s/clusters/{uuid}/deployments/{deploymentId} | Get a deployment by id
[**GetDeployments**](DeploymentsApi.md#GetDeployments) | **Get** /v1/k8s/clusters/{uuid}/deployments | Get all cluster&#x27;s deployments
[**GetDeploymentsCount**](DeploymentsApi.md#GetDeploymentsCount) | **Get** /v1/k8s/clusters/{uuid}/deployments/count | Get the number of deployments
[**GetDeploymentsHistory**](DeploymentsApi.md#GetDeploymentsHistory) | **Get** /v1/k8s/clusters/{uuid}/deployments/deleted | Get all the deployments that were deleted in the cluster
[**GetDeploymentsHistoryCount**](DeploymentsApi.md#GetDeploymentsHistoryCount) | **Get** /v1/k8s/clusters/{uuid}/deployments/deleted/count | Get the number of deleted deployments
[**GetPodsByDeploymentId**](DeploymentsApi.md#GetPodsByDeploymentId) | **Get** /v1/k8s/clusters/{uuid}/deployments/{deploymentId}/pods | Get all pods associated with the deployment by id
[**SyncDeployments**](DeploymentsApi.md#SyncDeployments) | **Post** /v1/k8s/clusters/{uuid}/deployments/sync | Sync deployments
[**UpdateDeployment**](DeploymentsApi.md#UpdateDeployment) | **Put** /v1/k8s/clusters/{uuid}/deployments/{deploymentId} | Update a deployment by id
[**UpdateDeployments**](DeploymentsApi.md#UpdateDeployments) | **Put** /v1/k8s/clusters/{uuid}/deployments | Sync deployments

# **CreateDeployment**
> CreateDeployment(ctx, uuid, optional)
Create a deployment

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
 **optional** | ***DeploymentsApiCreateDeploymentOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DeploymentsApiCreateDeploymentOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**optional.Interface of Deployment**](Deployment.md)|  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteDeployment**
> DeleteDeployment(ctx, uuid, deploymentId)
Delete a deployment by id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **deploymentId** | **string**| the unique Id of the deployment (UUID) | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetDeployment**
> Deployment GetDeployment(ctx, uuid, deploymentId)
Get a deployment by id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **deploymentId** | **string**| the unique Id of the deployment (UUID) | 

### Return type

[**Deployment**](Deployment.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetDeployments**
> []Deployment GetDeployments(ctx, uuid)
Get all cluster's deployments

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**[]Deployment**](array.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetDeploymentsCount**
> float64 GetDeploymentsCount(ctx, uuid)
Get the number of deployments

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

**float64**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetDeploymentsHistory**
> []Deployment GetDeploymentsHistory(ctx, uuid)
Get all the deployments that were deleted in the cluster

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**[]Deployment**](array.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetDeploymentsHistoryCount**
> float64 GetDeploymentsHistoryCount(ctx, uuid)
Get the number of deleted deployments

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

**float64**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetPodsByDeploymentId**
> []Pod GetPodsByDeploymentId(ctx, uuid, deploymentId)
Get all pods associated with the deployment by id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **deploymentId** | **string**| the unique Id of the deployment (UUID) | 

### Return type

[**[]Pod**](array.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **SyncDeployments**
> SyncDeployments(ctx, uuid, optional)
Sync deployments

Sync all deployments from k8s cluster to Run:AI backend 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
 **optional** | ***DeploymentsApiSyncDeploymentsOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DeploymentsApiSyncDeploymentsOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**optional.Interface of []Deployment**](Deployment.md)|  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateDeployment**
> UpdateDeployment(ctx, body, uuid, deploymentId)
Update a deployment by id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Deployment**](Deployment.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **deploymentId** | **string**| the unique Id of the deployment (UUID) | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateDeployments**
> IdsResponse UpdateDeployments(ctx, body, uuid)
Sync deployments

Sync all deployments in a project from k8s cluster to Run:AI backend 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**DeploymentsSync**](DeploymentsSync.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**IdsResponse**](IdsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

