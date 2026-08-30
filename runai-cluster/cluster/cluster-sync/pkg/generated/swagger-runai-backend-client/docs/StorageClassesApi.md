# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateStorageClass**](StorageClassesApi.md#CreateStorageClass) | **Post** /v1/k8s/clusters/{uuid}/storage-classes | create storageClass
[**DeleteStorageClass**](StorageClassesApi.md#DeleteStorageClass) | **Delete** /v1/k8s/clusters/{uuid}/storage-classes/{storage_class_name} | delete storageClass
[**GetStorageClasses**](StorageClassesApi.md#GetStorageClasses) | **Get** /v1/k8s/clusters/{uuid}/storage-classes | get all storageClasses from a specific cluster
[**SyncStorageClass**](StorageClassesApi.md#SyncStorageClass) | **Post** /v1/k8s/clusters/{uuid}/storage-classes/sync | sync storageClass
[**UpdateStorageClass**](StorageClassesApi.md#UpdateStorageClass) | **Put** /v1/k8s/clusters/{uuid}/storage-classes/{storage_class_name} | update storageClass

# **CreateStorageClass**
> StorageClass CreateStorageClass(ctx, body, uuid)
create storageClass

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**StorageClass**](StorageClass.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**StorageClass**](StorageClass.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteStorageClass**
> DeleteStorageClass(ctx, uuid, storageClassName)
delete storageClass

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **storageClassName** | **string**| Unique identifier of the storageClass | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetStorageClasses**
> []StorageClass GetStorageClasses(ctx, uuid)
get all storageClasses from a specific cluster

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**[]StorageClass**](array.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **SyncStorageClass**
> StorageClass SyncStorageClass(ctx, body, uuid)
sync storageClass

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]StorageClass**](StorageClass.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**StorageClass**](StorageClass.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateStorageClass**
> []StorageClass UpdateStorageClass(ctx, body, uuid, storageClassName)
update storageClass

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**StorageClass**](StorageClass.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **storageClassName** | **string**| Unique identifier of the storageClass | 

### Return type

[**[]StorageClass**](array.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

