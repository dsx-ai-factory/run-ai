# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteTraining**](TrainingsApi.md#DeleteTraining) | **Delete** /v1/k8s/clusters/{clusterId}/trainings/{id} | Delete a training
[**SyncTrainingsStatus**](TrainingsApi.md#SyncTrainingsStatus) | **Post** /v1/k8s/clusters/{clusterId}/trainings/status | Sync trainings status
[**UpdateTrainingStatus**](TrainingsApi.md#UpdateTrainingStatus) | **Put** /v1/k8s/clusters/{clusterId}/trainings/status | Update training status

# **DeleteTraining**
> DeleteTraining(ctx, clusterId, id)
Delete a training

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

# **SyncTrainingsStatus**
> SyncTrainingsStatus(ctx, clusterId, optional)
Sync trainings status

Sync the status of all trainings from k8s cluster to Run:AI control plane 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | [**string**](.md)| Unique identifier of the cluster | 
 **optional** | ***TrainingsApiSyncTrainingsStatusOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a TrainingsApiSyncTrainingsStatusOpts struct
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

# **UpdateTrainingStatus**
> UpdateTrainingStatus(ctx, body, clusterId)
Update training status

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

