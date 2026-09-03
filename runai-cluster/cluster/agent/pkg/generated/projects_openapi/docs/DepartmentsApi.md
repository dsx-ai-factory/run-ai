# \DepartmentsApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**V1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPut**](DepartmentsApi.md#V1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPut) | **Put** /v1/k8s/clusters/{clusterId}/departments/{department-id}/access-control | Set the department admins.
[**V1K8sClustersClusterIdDepartmentsDepartmentIdDelete**](DepartmentsApi.md#V1K8sClustersClusterIdDepartmentsDepartmentIdDelete) | **Delete** /v1/k8s/clusters/{clusterId}/departments/{department-id} | Delete a department.
[**V1K8sClustersClusterIdDepartmentsDepartmentIdGet**](DepartmentsApi.md#V1K8sClustersClusterIdDepartmentsDepartmentIdGet) | **Get** /v1/k8s/clusters/{clusterId}/departments/{department-id} | Get a department.
[**V1K8sClustersClusterIdDepartmentsDepartmentIdPut**](DepartmentsApi.md#V1K8sClustersClusterIdDepartmentsDepartmentIdPut) | **Put** /v1/k8s/clusters/{clusterId}/departments/{department-id} | Update a department.
[**V1K8sClustersClusterIdDepartmentsGet**](DepartmentsApi.md#V1K8sClustersClusterIdDepartmentsGet) | **Get** /v1/k8s/clusters/{clusterId}/departments | Get all departments.
[**V1K8sClustersClusterIdDepartmentsPost**](DepartmentsApi.md#V1K8sClustersClusterIdDepartmentsPost) | **Post** /v1/k8s/clusters/{clusterId}/departments | Create a new department.



## V1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPut

> DepartmentAccessControl V1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPut(ctx, clusterId, departmentId).DepartmentAccessControl(departmentAccessControl).Execute()

Set the department admins.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
    departmentId := int32(2) // int32 | The unique id identifying the department.
    departmentAccessControl := *openapiclient.NewDepartmentAccessControl() // DepartmentAccessControl | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.V1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPut(context.Background(), clusterId, departmentId).DepartmentAccessControl(departmentAccessControl).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.V1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPut`: DepartmentAccessControl
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.V1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 
**departmentId** | **int32** | The unique id identifying the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterIdDepartmentsDepartmentIdAccessControlPutRequest struct via the builder pattern


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


## V1K8sClustersClusterIdDepartmentsDepartmentIdDelete

> Department V1K8sClustersClusterIdDepartmentsDepartmentIdDelete(ctx, clusterId, departmentId).ExcludePermissions(excludePermissions).Execute()

Delete a department.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
    departmentId := int32(2) // int32 | The unique id identifying the department.
    excludePermissions := true // bool | filter permissions (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.V1K8sClustersClusterIdDepartmentsDepartmentIdDelete(context.Background(), clusterId, departmentId).ExcludePermissions(excludePermissions).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.V1K8sClustersClusterIdDepartmentsDepartmentIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterIdDepartmentsDepartmentIdDelete`: Department
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.V1K8sClustersClusterIdDepartmentsDepartmentIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 
**departmentId** | **int32** | The unique id identifying the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterIdDepartmentsDepartmentIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **excludePermissions** | **bool** | filter permissions | 

### Return type

[**Department**](Department.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersClusterIdDepartmentsDepartmentIdGet

> Department V1K8sClustersClusterIdDepartmentsDepartmentIdGet(ctx, clusterId, departmentId).ExcludePermissions(excludePermissions).Execute()

Get a department.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
    departmentId := int32(2) // int32 | The unique id identifying the department.
    excludePermissions := true // bool | filter permissions (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.V1K8sClustersClusterIdDepartmentsDepartmentIdGet(context.Background(), clusterId, departmentId).ExcludePermissions(excludePermissions).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.V1K8sClustersClusterIdDepartmentsDepartmentIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterIdDepartmentsDepartmentIdGet`: Department
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.V1K8sClustersClusterIdDepartmentsDepartmentIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 
**departmentId** | **int32** | The unique id identifying the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterIdDepartmentsDepartmentIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **excludePermissions** | **bool** | filter permissions | 

### Return type

[**Department**](Department.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersClusterIdDepartmentsDepartmentIdPut

> Department V1K8sClustersClusterIdDepartmentsDepartmentIdPut(ctx, clusterId, departmentId).Department(department).ExcludePermissions(excludePermissions).Execute()

Update a department.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
    departmentId := int32(2) // int32 | The unique id identifying the department.
    department := *openapiclient.NewDepartment() // Department | 
    excludePermissions := true // bool | filter permissions (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.V1K8sClustersClusterIdDepartmentsDepartmentIdPut(context.Background(), clusterId, departmentId).Department(department).ExcludePermissions(excludePermissions).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.V1K8sClustersClusterIdDepartmentsDepartmentIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterIdDepartmentsDepartmentIdPut`: Department
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.V1K8sClustersClusterIdDepartmentsDepartmentIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 
**departmentId** | **int32** | The unique id identifying the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterIdDepartmentsDepartmentIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **department** | [**Department**](Department.md) |  | 
 **excludePermissions** | **bool** | filter permissions | 

### Return type

[**Department**](Department.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersClusterIdDepartmentsGet

> []Department V1K8sClustersClusterIdDepartmentsGet(ctx, clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()

Get all departments.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
    excludePermissions := true // bool | filter permissions (optional)
    memoryUnitMb := true // bool | should memory be returned in MB. default is \"false\", in which case it will be returned in MiB. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.V1K8sClustersClusterIdDepartmentsGet(context.Background(), clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.V1K8sClustersClusterIdDepartmentsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterIdDepartmentsGet`: []Department
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.V1K8sClustersClusterIdDepartmentsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterIdDepartmentsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **excludePermissions** | **bool** | filter permissions | 
 **memoryUnitMb** | **bool** | should memory be returned in MB. default is \&quot;false\&quot;, in which case it will be returned in MiB. | 

### Return type

[**[]Department**](Department.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersClusterIdDepartmentsPost

> Department V1K8sClustersClusterIdDepartmentsPost(ctx, clusterId).DepartmentUserDefinedProperties(departmentUserDefinedProperties).ExcludePermissions(excludePermissions).Execute()

Create a new department.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
    departmentUserDefinedProperties := *openapiclient.NewDepartmentUserDefinedProperties() // DepartmentUserDefinedProperties | 
    excludePermissions := true // bool | filter permissions (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.V1K8sClustersClusterIdDepartmentsPost(context.Background(), clusterId).DepartmentUserDefinedProperties(departmentUserDefinedProperties).ExcludePermissions(excludePermissions).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.V1K8sClustersClusterIdDepartmentsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterIdDepartmentsPost`: Department
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.V1K8sClustersClusterIdDepartmentsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterIdDepartmentsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **departmentUserDefinedProperties** | [**DepartmentUserDefinedProperties**](DepartmentUserDefinedProperties.md) |  | 
 **excludePermissions** | **bool** | filter permissions | 

### Return type

[**Department**](Department.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

