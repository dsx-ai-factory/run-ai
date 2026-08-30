# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**V1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPut**](DepartmentsApi.md#V1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPut) | **Put** /v1/k8s/clusters/{clusterId}/departments/{department-id}/access-control | Set the department admins.
[**V1K8sClustersClusterIdDepartmentsDepartmentIdDelete**](DepartmentsApi.md#V1K8sClustersClusterIdDepartmentsDepartmentIdDelete) | **Delete** /v1/k8s/clusters/{clusterId}/departments/{department-id} | Delete a department.
[**V1K8sClustersClusterIdDepartmentsDepartmentIdGet**](DepartmentsApi.md#V1K8sClustersClusterIdDepartmentsDepartmentIdGet) | **Get** /v1/k8s/clusters/{clusterId}/departments/{department-id} | Get a department.
[**V1K8sClustersClusterIdDepartmentsDepartmentIdPut**](DepartmentsApi.md#V1K8sClustersClusterIdDepartmentsDepartmentIdPut) | **Put** /v1/k8s/clusters/{clusterId}/departments/{department-id} | Update a department.
[**V1K8sClustersClusterIdDepartmentsGet**](DepartmentsApi.md#V1K8sClustersClusterIdDepartmentsGet) | **Get** /v1/k8s/clusters/{clusterId}/departments | Get all departments.
[**V1K8sClustersClusterIdDepartmentsPost**](DepartmentsApi.md#V1K8sClustersClusterIdDepartmentsPost) | **Post** /v1/k8s/clusters/{clusterId}/departments | Create a new department.

# **V1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPut**
> DepartmentAccessControl V1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPut(ctx, body, clusterId, departmentId)
Set the department admins.

Set the department admins. This endpoint requires EDITOR role.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**DepartmentAccessControl**](DepartmentAccessControl.md)|  | 
  **clusterId** | **string**| The unique uuid identifying the cluster. | 
  **departmentId** | **int32**| The unique id identifying the department. | 

### Return type

[**DepartmentAccessControl**](DepartmentAccessControl.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterIdDepartmentsDepartmentIdDelete**
> Department V1K8sClustersClusterIdDepartmentsDepartmentIdDelete(ctx, clusterId, departmentId, optional)
Delete a department.

Delete a department from the cluster. This endpoint requires EDITOR role.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | **string**| The unique uuid identifying the cluster. | 
  **departmentId** | **int32**| The unique id identifying the department. | 
 **optional** | ***DepartmentsApiV1K8sClustersClusterIdDepartmentsDepartmentIdDeleteOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DepartmentsApiV1K8sClustersClusterIdDepartmentsDepartmentIdDeleteOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **excludePermissions** | **optional.Bool**| filter permissions | 

### Return type

[**Department**](Department.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterIdDepartmentsDepartmentIdGet**
> Department V1K8sClustersClusterIdDepartmentsDepartmentIdGet(ctx, clusterId, departmentId, optional)
Get a department.

Get a specific department by the ID.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | **string**| The unique uuid identifying the cluster. | 
  **departmentId** | **int32**| The unique id identifying the department. | 
 **optional** | ***DepartmentsApiV1K8sClustersClusterIdDepartmentsDepartmentIdGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DepartmentsApiV1K8sClustersClusterIdDepartmentsDepartmentIdGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **excludePermissions** | **optional.Bool**| filter permissions | 

### Return type

[**Department**](Department.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterIdDepartmentsDepartmentIdPut**
> Department V1K8sClustersClusterIdDepartmentsDepartmentIdPut(ctx, body, clusterId, departmentId, optional)
Update a department.

Update a department in the cluster. This endpoint requires EDITOR role.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Department**](Department.md)|  | 
  **clusterId** | **string**| The unique uuid identifying the cluster. | 
  **departmentId** | **int32**| The unique id identifying the department. | 
 **optional** | ***DepartmentsApiV1K8sClustersClusterIdDepartmentsDepartmentIdPutOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DepartmentsApiV1K8sClustersClusterIdDepartmentsDepartmentIdPutOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **excludePermissions** | **optional.**| filter permissions | 

### Return type

[**Department**](Department.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterIdDepartmentsGet**
> []Department V1K8sClustersClusterIdDepartmentsGet(ctx, clusterId, optional)
Get all departments.

Get list of all the department in the cluster.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterId** | **string**| The unique uuid identifying the cluster. | 
 **optional** | ***DepartmentsApiV1K8sClustersClusterIdDepartmentsGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DepartmentsApiV1K8sClustersClusterIdDepartmentsGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **excludePermissions** | **optional.Bool**| filter permissions | 

### Return type

[**[]Department**](Department.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterIdDepartmentsPost**
> Department V1K8sClustersClusterIdDepartmentsPost(ctx, body, clusterId, optional)
Create a new department.

Create a new department in the cluster. This endpoint requires EDITOR role.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**DepartmentUserDefinedProperties**](DepartmentUserDefinedProperties.md)|  | 
  **clusterId** | **string**| The unique uuid identifying the cluster. | 
 **optional** | ***DepartmentsApiV1K8sClustersClusterIdDepartmentsPostOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DepartmentsApiV1K8sClustersClusterIdDepartmentsPostOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **excludePermissions** | **optional.**| filter permissions | 

### Return type

[**Department**](Department.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

