# \DepartmentsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateDepartment**](DepartmentsAPI.md#CreateDepartment) | **Post** /v1/k8s/clusters/{clusterId}/departments | Create a new department.
[**DeleteDepartment**](DepartmentsAPI.md#DeleteDepartment) | **Delete** /v1/k8s/clusters/{clusterId}/departments/{department-id} | Delete a department.
[**GetDepartment**](DepartmentsAPI.md#GetDepartment) | **Get** /v1/k8s/clusters/{clusterId}/departments/{department-id} | Get a specific department.
[**GetDepartments**](DepartmentsAPI.md#GetDepartments) | **Get** /v1/k8s/clusters/{clusterId}/departments | List all departments.
[**UpdateDepartment**](DepartmentsAPI.md#UpdateDepartment) | **Put** /v1/k8s/clusters/{clusterId}/departments/{department-id} | Update a department.
[**UpdateDepartmentAdmins**](DepartmentsAPI.md#UpdateDepartmentAdmins) | **Put** /v1/k8s/clusters/{clusterId}/departments/{department-id}/access-control | Set the department admins.



## CreateDepartment

> Department1 CreateDepartment(ctx, clusterId).DepartmentCreateRequest(departmentCreateRequest).Execute()

Create a new department.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
	departmentCreateRequest := *openapiclient.NewDepartmentCreateRequest("default") // DepartmentCreateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DepartmentsAPI.CreateDepartment(context.Background(), clusterId).DepartmentCreateRequest(departmentCreateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsAPI.CreateDepartment``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateDepartment`: Department1
	fmt.Fprintf(os.Stdout, "Response from `DepartmentsAPI.CreateDepartment`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateDepartmentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **departmentCreateRequest** | [**DepartmentCreateRequest**](DepartmentCreateRequest.md) |  | 

### Return type

[**Department1**](Department1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteDepartment

> Department1 DeleteDepartment(ctx, clusterId, departmentId).Execute()

Delete a department.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
	departmentId := int32(2) // int32 | The unique id identifying the department.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DepartmentsAPI.DeleteDepartment(context.Background(), clusterId, departmentId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsAPI.DeleteDepartment``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteDepartment`: Department1
	fmt.Fprintf(os.Stdout, "Response from `DepartmentsAPI.DeleteDepartment`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 
**departmentId** | **int32** | The unique id identifying the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDepartmentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Department1**](Department1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDepartment

> Department1 GetDepartment(ctx, clusterId, departmentId).ExcludePermissions(excludePermissions).Execute()

Get a specific department.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	departmentId := int32(2) // int32 | The unique id of the department.
	excludePermissions := true // bool | backwards compatibility of the 'departmentAdmins' field. if 'true', will not set the 'departmentAdmins' field in the returned department. if 'false', will set the 'departmentAdmins' field in the returned department. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DepartmentsAPI.GetDepartment(context.Background(), clusterId, departmentId).ExcludePermissions(excludePermissions).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsAPI.GetDepartment``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDepartment`: Department1
	fmt.Fprintf(os.Stdout, "Response from `DepartmentsAPI.GetDepartment`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **int32** | The unique id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartmentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **excludePermissions** | **bool** | backwards compatibility of the &#39;departmentAdmins&#39; field. if &#39;true&#39;, will not set the &#39;departmentAdmins&#39; field in the returned department. if &#39;false&#39;, will set the &#39;departmentAdmins&#39; field in the returned department. | 

### Return type

[**Department1**](Department1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDepartments

> []Department1 GetDepartments(ctx, clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()

List all departments.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	excludePermissions := true // bool | Backward compatibility of the `departmentAdmins` field. If `true`, the `departmentAdmins` field in the returned departments is not set. If `false`, the `departmentAdmins` is set in the returned departments. (optional)
	memoryUnitMb := true // bool | Memory returned in MB. When set to `false` (default) memory will be returned in MiB. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DepartmentsAPI.GetDepartments(context.Background(), clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsAPI.GetDepartments``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDepartments`: []Department1
	fmt.Fprintf(os.Stdout, "Response from `DepartmentsAPI.GetDepartments`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartmentsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **excludePermissions** | **bool** | Backward compatibility of the &#x60;departmentAdmins&#x60; field. If &#x60;true&#x60;, the &#x60;departmentAdmins&#x60; field in the returned departments is not set. If &#x60;false&#x60;, the &#x60;departmentAdmins&#x60; is set in the returned departments. | 
 **memoryUnitMb** | **bool** | Memory returned in MB. When set to &#x60;false&#x60; (default) memory will be returned in MiB. | 

### Return type

[**[]Department1**](Department1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDepartment

> Department1 UpdateDepartment(ctx, clusterId, departmentId).DepartmentUpdateRequest(departmentUpdateRequest).Execute()

Update a department.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
	departmentId := int32(2) // int32 | The unique id identifying the department.
	departmentUpdateRequest := *openapiclient.NewDepartmentUpdateRequest("default") // DepartmentUpdateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DepartmentsAPI.UpdateDepartment(context.Background(), clusterId, departmentId).DepartmentUpdateRequest(departmentUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsAPI.UpdateDepartment``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateDepartment`: Department1
	fmt.Fprintf(os.Stdout, "Response from `DepartmentsAPI.UpdateDepartment`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 
**departmentId** | **int32** | The unique id identifying the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDepartmentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **departmentUpdateRequest** | [**DepartmentUpdateRequest**](DepartmentUpdateRequest.md) |  | 

### Return type

[**Department1**](Department1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDepartmentAdmins

> DepartmentAccessControl UpdateDepartmentAdmins(ctx, clusterId, departmentId).DepartmentAccessControl(departmentAccessControl).Execute()

Set the department admins.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
	departmentId := int32(2) // int32 | The unique id identifying the department.
	departmentAccessControl := *openapiclient.NewDepartmentAccessControl() // DepartmentAccessControl | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DepartmentsAPI.UpdateDepartmentAdmins(context.Background(), clusterId, departmentId).DepartmentAccessControl(departmentAccessControl).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsAPI.UpdateDepartmentAdmins``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateDepartmentAdmins`: DepartmentAccessControl
	fmt.Fprintf(os.Stdout, "Response from `DepartmentsAPI.UpdateDepartmentAdmins`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 
**departmentId** | **int32** | The unique id identifying the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDepartmentAdminsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **departmentAccessControl** | [**DepartmentAccessControl**](DepartmentAccessControl.md) |  | 

### Return type

[**DepartmentAccessControl**](DepartmentAccessControl.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

