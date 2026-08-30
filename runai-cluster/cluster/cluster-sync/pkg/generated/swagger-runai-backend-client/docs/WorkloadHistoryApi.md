# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetWorkloadHistory**](WorkloadHistoryApi.md#GetWorkloadHistory) | **Get** /v1/k8s/clusters/{uuid}/jobs/{job_id}/workload-history | list workload history

# **GetWorkloadHistory**
> WorkloadHistoryResponse GetWorkloadHistory(ctx, uuid, jobId)
list workload history

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 
  **jobId** | [**string**](.md)| Unique identifier of the job | 

### Return type

[**WorkloadHistoryResponse**](workloadHistoryResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

