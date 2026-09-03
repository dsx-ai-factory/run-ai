# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateNodePool**](NodePoolsApi.md#CreateNodePool) | **Post** /v1/k8s/clusters/{clusterId}/node-pools | Create a Node Pool
[**DeleteNodePool**](NodePoolsApi.md#DeleteNodePool) | **Delete** /v1/k8s/clusters/{clusterId}/node-pools/{id} | Delete a Node Pool by id
[**GetNodePools**](NodePoolsApi.md#GetNodePools) | **Get** /v1/k8s/clusters/{clusterId}/node-pools | Get all cluster&#x27;s Node Pools.
[**SyncNodePools**](NodePoolsApi.md#SyncNodePools) | **Post** /v1/k8s/clusters/{clusterId}/node-pools/sync | Sync Node Pools
[**UpdateNodePool**](NodePoolsApi.md#UpdateNodePool) | **Put** /v1/k8s/clusters/{clusterId}/node-pools/{id} | Update Node Pool
[**UpdateNodePoolLabels**](NodePoolsApi.md#UpdateNodePoolLabels) | **Put** /v1/k8s/clusters/{clusterId}/node-pools/{id}/labels | Update Node Pool labels
[**UpdateNodePoolStatus**](NodePoolsApi.md#UpdateNodePoolStatus) | **Put** /v1/k8s/clusters/{clusterId}/node-pools/status | Update Node Pool status

# **CreateNodePool**
> CreateNodePool(ctx, clusterId, optional)
Create a Node Pool

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 
 **optional** | ***NodePoolsApiCreateNodePoolOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NodePoolsApiCreateNodePoolOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**optional.Interface of NodePoolCreateRequest**](NodePoolCreateRequest.md)|  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteNodePool**
> DeleteNodePool(ctx, clusterId, id)
Delete a Node Pool by id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 
  **id** | **int32**| The unique id identifying the Node Pool. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetNodePools**
> []NodePool GetNodePools(ctx, clusterId)
Get all cluster's Node Pools.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**[]NodePool**](NodePool.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **SyncNodePools**
> SyncNodePools(ctx, clusterId, optional)
Sync Node Pools

Sync all Node Pools statuses from k8s cluster to Run:AI backend 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 
 **optional** | ***NodePoolsApiSyncNodePoolsOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NodePoolsApiSyncNodePoolsOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**optional.Interface of []NodePoolStatusRequest**](NodePoolStatusRequest.md)|  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateNodePool**
> UpdateNodePool(ctx, body, clusterId, id)
Update Node Pool

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UpdateNodePoolRequest**](UpdateNodePoolRequest.md)|  | 
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 
  **id** | **int32**| The unique id identifying the Node Pool. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateNodePoolLabels**
> UpdateNodePoolLabels(ctx, body, clusterId, id)
Update Node Pool labels

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**NodePoolLabelsRequest**](NodePoolLabelsRequest.md)|  | 
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 
  **id** | **int32**| The unique id identifying the Node Pool. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateNodePoolStatus**
> UpdateNodePoolStatus(ctx, body, clusterId, name)
Update Node Pool status

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**NodePoolStatusRequest**](NodePoolStatusRequest.md)|  | 
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 
  **name** | **string**| The unique name identifying the Node Pool. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

