# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteInference**](InferencesApi.md#DeleteInference) | **Delete** /v1/k8s/clusters/{clusterId}/inferences/{id} | Delete an inference

# **DeleteInference**
> DeleteInference(ctx, clusterId, id)
Delete an inference

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

