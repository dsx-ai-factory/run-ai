# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreatePods**](PodsApi.md#CreatePods) | **Post** /v1/k8s/clusters/{uuid}/pods | create pods
[**DeletePod**](PodsApi.md#DeletePod) | **Delete** /v1/k8s/clusters/{uuid}/pods/{podId} | Delete a pod by id
[**GetPods**](PodsApi.md#GetPods) | **Get** /v1/k8s/clusters/{uuid}/pods | get all pods from a specific cluster
[**SyncPodsData**](PodsApi.md#SyncPodsData) | **Post** /v1/k8s/clusters/{uuid}/pods/sync | Sync All cluster pods data to DB
[**UpdatePod**](PodsApi.md#UpdatePod) | **Put** /v1/k8s/clusters/{uuid}/pods/{podId} | Update a pod by id
[**UpdatePods**](PodsApi.md#UpdatePods) | **Put** /v1/k8s/clusters/{uuid}/pods | Sync all pods in a namespace to DB

# **CreatePods**
> []Pod CreatePods(ctx, body, uuid)
create pods

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]Pod**](Pod.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**[]Pod**](array.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeletePod**
> string DeletePod(ctx, uuid, podId)
Delete a pod by id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **podId** | **string**| Unique identifier of the pod. | 

### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetPods**
> []Pod GetPods(ctx, uuid)
get all pods from a specific cluster

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**[]Pod**](array.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **SyncPodsData**
> SyncPodsData(ctx, body, uuid)
Sync All cluster pods data to DB

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]Pod**](Pod.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdatePod**
> UpdatePod(ctx, body, uuid, podId)
Update a pod by id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Pod**](Pod.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **podId** | **string**| Unique identifier of the pod. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdatePods**
> IdsResponse UpdatePods(ctx, body, uuid)
Sync all pods in a namespace to DB

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**PodsSync**](PodsSync.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**IdsResponse**](IdsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

