# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteDistributed**](DistributedApi.md#DeleteDistributed) | **Delete** /v1/k8s/clusters/{clusterId}/distributed/{id} | Delete a distributed
[**SyncDistributedStatus**](DistributedApi.md#SyncDistributedStatus) | **Post** /v1/k8s/clusters/{clusterId}/distributed/status | Sync distributed status
[**UpdateDistributedStatus**](DistributedApi.md#UpdateDistributedStatus) | **Put** /v1/k8s/clusters/{clusterId}/distributed/status | Update distributed status

# **DeleteDistributed**
> DeleteDistributed(ctx, clusterId, id)
Delete a distributed

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 
  **id** | [**string**](.md)|  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **SyncDistributedStatus**
> SyncDistributedStatus(ctx, clusterId, optional)
Sync distributed status

Sync the status of all distributed from k8s cluster to Run:AI control plane 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 
 **optional** | ***DistributedApiSyncDistributedStatusOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DistributedApiSyncDistributedStatusOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**optional.Interface of []WorkloadStatus**](WorkloadStatus.md)|  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateDistributedStatus**
> UpdateDistributedStatus(ctx, body, clusterId)
Update distributed status

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**WorkloadStatus**](WorkloadStatus.md)|  | 
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

