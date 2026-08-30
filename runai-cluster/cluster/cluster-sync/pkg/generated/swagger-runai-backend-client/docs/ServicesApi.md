# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateService**](ServicesApi.md#CreateService) | **Post** /v1/k8s/clusters/{uuid}/services | Create a service
[**DeleteService**](ServicesApi.md#DeleteService) | **Delete** /v1/k8s/clusters/{uuid}/services/{serviceId} | Delete a service by id
[**GetService**](ServicesApi.md#GetService) | **Get** /v1/k8s/clusters/{uuid}/services/{serviceId} | Get a service by id
[**GetServices**](ServicesApi.md#GetServices) | **Get** /v1/k8s/clusters/{uuid}/services | Get all cluster&#x27;s services.
[**GetServicesCount**](ServicesApi.md#GetServicesCount) | **Get** /v1/k8s/clusters/{uuid}/services/count | Get the number of services
[**SyncServices**](ServicesApi.md#SyncServices) | **Post** /v1/k8s/clusters/{uuid}/services/sync | Sync services
[**UpdateService**](ServicesApi.md#UpdateService) | **Put** /v1/k8s/clusters/{uuid}/services/{serviceId} | Update a service by id
[**UpdateServices**](ServicesApi.md#UpdateServices) | **Put** /v1/k8s/clusters/{uuid}/services | Sync services

# **CreateService**
> CreateService(ctx, uuid, optional)
Create a service

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
 **optional** | ***ServicesApiCreateServiceOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ServicesApiCreateServiceOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**optional.Interface of Service**](Service.md)|  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteService**
> DeleteService(ctx, uuid, serviceId)
Delete a service by id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **serviceId** | **string**| the unique Id of the service (UUID) | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetService**
> Service GetService(ctx, uuid, serviceId)
Get a service by id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **serviceId** | **string**| the unique Id of the service (UUID) | 

### Return type

[**Service**](Service.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetServices**
> []Service GetServices(ctx, uuid)
Get all cluster's services.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**[]Service**](array.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetServicesCount**
> float64 GetServicesCount(ctx, uuid)
Get the number of services

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

# **SyncServices**
> SyncServices(ctx, uuid, optional)
Sync services

Sync all services from k8s cluster to Run:AI backend 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
 **optional** | ***ServicesApiSyncServicesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ServicesApiSyncServicesOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**optional.Interface of []Service**](Service.md)|  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateService**
> UpdateService(ctx, body, uuid, serviceId)
Update a service by id

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Service**](Service.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **serviceId** | **string**| the unique Id of the service (UUID) | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateServices**
> IdsResponse UpdateServices(ctx, body, uuid)
Sync services

Sync all services in a namespace from k8s cluster to Run:AI backend 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ServicesSync**](ServicesSync.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

[**IdsResponse**](IdsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

