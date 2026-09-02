# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreatePodGroups**](PodGroupsApi.md#CreatePodGroups) | **Post** /v1/k8s/clusters/{uuid}/podGroups | Create podGroups
[**DeletePodGroup**](PodGroupsApi.md#DeletePodGroup) | **Delete** /v1/k8s/clusters/{uuid}/podGroups/{podGroupId} | Delete a podGroup by id
[**SyncPodGroupData**](PodGroupsApi.md#SyncPodGroupData) | **Post** /v1/k8s/clusters/{uuid}/podGroups/sync | Sync All cluster podGroup data to DB
[**UpdatePodGroup**](PodGroupsApi.md#UpdatePodGroup) | **Put** /v1/k8s/clusters/{uuid}/podGroups/{podGroupId} | Update a podGroup by id
[**UpdatePodGroups**](PodGroupsApi.md#UpdatePodGroups) | **Put** /v1/k8s/clusters/{uuid}/podGroups | Sync all podGroups in a project to DB

# **CreatePodGroups**
> CreatePodGroups(ctx, body, uuid)
Create podGroups

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]PodGroup**](PodGroup.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeletePodGroup**
> DeletePodGroup(ctx, uuid, podGroupId)
Delete a podGroup by id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **podGroupId** | [**string**](.md)| Unique identifier of the podGroup (UUID) | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **SyncPodGroupData**
> SyncPodGroupData(ctx, body, uuid)
Sync All cluster podGroup data to DB

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]PodGroup**](PodGroup.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdatePodGroup**
> UpdatePodGroup(ctx, body, uuid, podGroupId)
Update a podGroup by id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**PodGroup**](PodGroup.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **podGroupId** | [**string**](.md)| Unique identifier of the podGroup (UUID) | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdatePodGroups**
> IdsResponse UpdatePodGroups(ctx, body, uuid)
Sync all podGroups in a project to DB

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**PodGroupsSync**](PodGroupsSync.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**IdsResponse**](IdsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

