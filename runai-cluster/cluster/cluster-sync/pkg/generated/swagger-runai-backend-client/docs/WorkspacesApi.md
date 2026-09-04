# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteWorkspace**](WorkspacesApi.md#DeleteWorkspace) | **Delete** /v1/k8s/clusters/{clusterId}/workspaces/{id} | Delete a workspace
[**SyncWorkspacesStatus**](WorkspacesApi.md#SyncWorkspacesStatus) | **Post** /v1/k8s/clusters/{clusterId}/workspaces/status | Sync workspaces status
[**UpdateWorkspaceStatus**](WorkspacesApi.md#UpdateWorkspaceStatus) | **Put** /v1/k8s/clusters/{clusterId}/workspaces/status | Update workspace status

# **DeleteWorkspace**
> DeleteWorkspace(ctx, clusterId, id)
Delete a workspace

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

# **SyncWorkspacesStatus**
> SyncWorkspacesStatus(ctx, clusterId, optional)
Sync workspaces status

Sync the status of all workspaces from k8s cluster to Run:AI control plane 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 
 **optional** | ***WorkspacesApiSyncWorkspacesStatusOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a WorkspacesApiSyncWorkspacesStatusOpts struct
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

# **UpdateWorkspaceStatus**
> UpdateWorkspaceStatus(ctx, body, clusterId)
Update workspace status

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

