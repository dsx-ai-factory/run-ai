# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetClusterGroups**](GroupsApi.md#GetClusterGroups) | **Get** /v1/k8s/clusters/{uuid}/groups | Get all cluster groups.

# **GetClusterGroups**
> Groups GetClusterGroups(ctx, uuid)
Get all cluster groups.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster. | 

### Return type

[**Groups**](Groups.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

