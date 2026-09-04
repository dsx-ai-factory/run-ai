# \DepartmentsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateDepartment**](DepartmentsApi.md#CreateDepartment) | **Post** /api/v1/org-unit/departments | Create department
[**CreateDepartment_0**](DepartmentsApi.md#CreateDepartment_0) | **Post** /v1/k8s/clusters/{clusterId}/departments | Create a new department.
[**DeleteDepartment**](DepartmentsApi.md#DeleteDepartment) | **Delete** /api/v1/org-unit/departments/{departmentId} | Delete department
[**DeleteDepartment_0**](DepartmentsApi.md#DeleteDepartment_0) | **Delete** /v1/k8s/clusters/{clusterId}/departments/{department-id} | Delete a department.
[**GetDepartment**](DepartmentsApi.md#GetDepartment) | **Get** /api/v1/org-unit/departments/{departmentId} | Get department
[**GetDepartmentAncestors**](DepartmentsApi.md#GetDepartmentAncestors) | **Get** /internal/org-unit/departments/{departmentId}/ancestors | Get department ancestors
[**GetDepartmentMetrics**](DepartmentsApi.md#GetDepartmentMetrics) | **Get** /v1/k8s/clusters/{clusterUuid}/departments/{departmentId}/metrics | Get metrics for a specific department.
[**GetDepartmentMetrics_0**](DepartmentsApi.md#GetDepartmentMetrics_0) | **Get** /v2/k8s/clusters/{clusterUuid}/departments/{departmentId}/metrics | Get the department requested metrics data. [Experimental]
[**GetDepartment_0**](DepartmentsApi.md#GetDepartment_0) | **Get** /v1/k8s/clusters/{clusterId}/departments/{department-id} | Get a specific department.
[**GetDepartments**](DepartmentsApi.md#GetDepartments) | **Get** /api/v1/org-unit/departments | Get departments
[**GetDepartmentsMetrics**](DepartmentsApi.md#GetDepartmentsMetrics) | **Get** /v1/k8s/clusters/{clusterUuid}/departments/metrics | Get metrics for all departments.
[**GetDepartmentsTelemetry**](DepartmentsApi.md#GetDepartmentsTelemetry) | **Get** /api/v1/org-unit/departments/telemetry | Get departments telemetry
[**GetDepartments_0**](DepartmentsApi.md#GetDepartments_0) | **Get** /v1/k8s/clusters/{clusterId}/departments | List all departments.
[**PatchDepartmentResources**](DepartmentsApi.md#PatchDepartmentResources) | **Patch** /api/v1/org-unit/departments/{departmentId}/resources | Patch department resources
[**UpdateDepartment**](DepartmentsApi.md#UpdateDepartment) | **Put** /api/v1/org-unit/departments/{departmentId} | Update department
[**UpdateDepartmentAdmins**](DepartmentsApi.md#UpdateDepartmentAdmins) | **Put** /v1/k8s/clusters/{clusterId}/departments/{department-id}/access-control | Set the department admins.
[**UpdateDepartmentResources**](DepartmentsApi.md#UpdateDepartmentResources) | **Put** /api/v1/org-unit/departments/{departmentId}/resources | Update department resources
[**UpdateDepartment_0**](DepartmentsApi.md#UpdateDepartment_0) | **Put** /v1/k8s/clusters/{clusterId}/departments/{department-id} | Update a department.



## CreateDepartment

> Department CreateDepartment(ctx).DepartmentCreationRequest(departmentCreationRequest).Execute()

Create department



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
    departmentCreationRequest := *openapiclient.NewDepartmentCreationRequest([]openapiclient.Resources{*openapiclient.NewResources(*openapiclient.NewResourcesGpu(float32(1000)))}, "organization1", "71f69d83-ba66-4822-adf5-55ce55efd210") // DepartmentCreationRequest | Department to create.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.CreateDepartment(context.Background()).DepartmentCreationRequest(departmentCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.CreateDepartment``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateDepartment`: Department
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.CreateDepartment`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateDepartmentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **departmentCreationRequest** | [**DepartmentCreationRequest**](DepartmentCreationRequest.md) | Department to create. | 

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


## CreateDepartment_0

> Department2 CreateDepartment_0(ctx, clusterId).DepartmentCreateRequest(departmentCreateRequest).Execute()

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
    departmentCreateRequest := *openapiclient.NewDepartmentCreateRequest("default") // DepartmentCreateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.CreateDepartment_0(context.Background(), clusterId).DepartmentCreateRequest(departmentCreateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.CreateDepartment_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateDepartment_0`: Department2
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.CreateDepartment_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateDepartment_1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **departmentCreateRequest** | [**DepartmentCreateRequest**](DepartmentCreateRequest.md) |  | 

### Return type

[**Department2**](Department2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteDepartment

> DeleteDepartment(ctx, departmentId).Execute()

Delete department



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
    departmentId := "1" // string | The id of the department.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.DeleteDepartment(context.Background(), departmentId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.DeleteDepartment``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDepartmentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteDepartment_0

> Department2 DeleteDepartment_0(ctx, clusterId, departmentId).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.DeleteDepartment_0(context.Background(), clusterId, departmentId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.DeleteDepartment_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteDepartment_0`: Department2
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.DeleteDepartment_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 
**departmentId** | **int32** | The unique id identifying the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDepartment_2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Department2**](Department2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDepartment

> GetDepartment200Response GetDepartment(ctx, departmentId).Execute()

Get department



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
    departmentId := "1" // string | The id of the department.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.GetDepartment(context.Background(), departmentId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.GetDepartment``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDepartment`: GetDepartment200Response
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.GetDepartment`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartmentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GetDepartment200Response**](GetDepartment200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDepartmentAncestors

> Ancestors GetDepartmentAncestors(ctx, departmentId).Execute()

Get department ancestors

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
    departmentId := "1" // string | The id of the department.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.GetDepartmentAncestors(context.Background(), departmentId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.GetDepartmentAncestors``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDepartmentAncestors`: Ancestors
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.GetDepartmentAncestors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartmentAncestorsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Ancestors**](Ancestors.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDepartmentMetrics

> Department GetDepartmentMetrics(ctx, clusterUuid, departmentId).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()

Get metrics for a specific department.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    departmentId := "1" // string | The id of the department.
    start := time.Now() // time.Time | Start of time range to fetch data from in UTC format. (optional)
    end := time.Now() // time.Time | End of time range to fetch data from in UTC format. (optional)
    numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)
    nodepoolName := "default" // string | Filter by unique nodepool name. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.GetDepartmentMetrics(context.Background(), clusterUuid, departmentId).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.GetDepartmentMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDepartmentMetrics`: Department
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.GetDepartmentMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartmentMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **start** | **time.Time** | Start of time range to fetch data from in UTC format. | 
 **end** | **time.Time** | End of time range to fetch data from in UTC format. | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]
 **nodepoolName** | **string** | Filter by unique nodepool name. | 

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


## GetDepartmentMetrics_0

> MetricsResponse GetDepartmentMetrics_0(ctx, clusterUuid, departmentId).Start(start).End(end).MetricType(metricType).NumberOfSamples(numberOfSamples).Execute()

Get the department requested metrics data. [Experimental]

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    departmentId := "1" // string | The id of the department.
    start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
    end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
    metricType := []openapiclient.MetricsType{openapiclient.MetricsType("GPU_UTILIZATION")} // []MetricsType | specifies what data to request
    numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.GetDepartmentMetrics_0(context.Background(), clusterUuid, departmentId).Start(start).End(end).MetricType(metricType).NumberOfSamples(numberOfSamples).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.GetDepartmentMetrics_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDepartmentMetrics_0`: MetricsResponse
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.GetDepartmentMetrics_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartmentMetrics_3Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **start** | **time.Time** | Start date of time range to fetch data in ISO 8601 timestamp format. | 
 **end** | **time.Time** | End date of time range to fetch data in ISO 8601 timestamp format. | 
 **metricType** | [**[]MetricsType**](MetricsType.md) | specifies what data to request | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]

### Return type

[**MetricsResponse**](MetricsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDepartment_0

> Department2 GetDepartment_0(ctx, clusterId, departmentId).ExcludePermissions(excludePermissions).Execute()

Get a specific department.



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
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    departmentId := int32(2) // int32 | The unique id of the department.
    excludePermissions := true // bool | backwards compatibility of the 'departmentAdmins' field. if 'true', will not set the 'departmentAdmins' field in the returned department. if 'false', will set the 'departmentAdmins' field in the returned department. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.GetDepartment_0(context.Background(), clusterId, departmentId).ExcludePermissions(excludePermissions).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.GetDepartment_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDepartment_0`: Department2
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.GetDepartment_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **int32** | The unique id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartment_4Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **excludePermissions** | **bool** | backwards compatibility of the &#39;departmentAdmins&#39; field. if &#39;true&#39;, will not set the &#39;departmentAdmins&#39; field in the returned department. if &#39;false&#39;, will set the &#39;departmentAdmins&#39; field in the returned department. | 

### Return type

[**Department2**](Department2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDepartments

> GetDepartments200Response GetDepartments(ctx).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Execute()

Get departments



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
    filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
    sortBy := "sortBy_example" // string | Sort results by a parameters. (optional)
    sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
    offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
    limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.GetDepartments(context.Background()).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.GetDepartments``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDepartments`: GetDepartments200Response
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.GetDepartments`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartmentsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **sortBy** | **string** | Sort results by a parameters. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**GetDepartments200Response**](GetDepartments200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDepartmentsMetrics

> Departments GetDepartmentsMetrics(ctx, clusterUuid).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()

Get metrics for all departments.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    start := time.Now() // time.Time | Start of time range to fetch data from in UTC format. (optional)
    end := time.Now() // time.Time | End of time range to fetch data from in UTC format. (optional)
    numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)
    nodepoolName := "default" // string | Filter by unique nodepool name. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.GetDepartmentsMetrics(context.Background(), clusterUuid).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.GetDepartmentsMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDepartmentsMetrics`: Departments
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.GetDepartmentsMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartmentsMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **start** | **time.Time** | Start of time range to fetch data from in UTC format. | 
 **end** | **time.Time** | End of time range to fetch data from in UTC format. | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]
 **nodepoolName** | **string** | Filter by unique nodepool name. | 

### Return type

[**Departments**](Departments.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDepartmentsTelemetry

> TelemetryResponse GetDepartmentsTelemetry(ctx).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolName(nodepoolName).DepartmentId(departmentId).GroupBy(groupBy).Execute()

Get departments telemetry

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
    telemetryType := "telemetryType_example" // string | specifies what data to request
    clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
    nodepoolName := "default" // string | Filter using the nodepool. (optional)
    departmentId := "1" // string | Filter using the department id. (optional)
    groupBy := []string{"GroupBy_example"} // []string | department fields to group the data by (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.GetDepartmentsTelemetry(context.Background()).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolName(nodepoolName).DepartmentId(departmentId).GroupBy(groupBy).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.GetDepartmentsTelemetry``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDepartmentsTelemetry`: TelemetryResponse
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.GetDepartmentsTelemetry`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartmentsTelemetryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **telemetryType** | **string** | specifies what data to request | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **nodepoolName** | **string** | Filter using the nodepool. | 
 **departmentId** | **string** | Filter using the department id. | 
 **groupBy** | **[]string** | department fields to group the data by | 

### Return type

[**TelemetryResponse**](TelemetryResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDepartments_0

> []Department2 GetDepartments_0(ctx, clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()

List all departments.



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
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    excludePermissions := true // bool | Backward compatibility of the `departmentAdmins` field. If `true`, the `departmentAdmins` field in the returned departments is not set. If `false`, the `departmentAdmins` is set in the returned departments. (optional)
    memoryUnitMb := true // bool | Memory returned in MB. When set to `false` (default) memory will be returned in MiB. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.GetDepartments_0(context.Background(), clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.GetDepartments_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDepartments_0`: []Department2
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.GetDepartments_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartments_5Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **excludePermissions** | **bool** | Backward compatibility of the &#x60;departmentAdmins&#x60; field. If &#x60;true&#x60;, the &#x60;departmentAdmins&#x60; field in the returned departments is not set. If &#x60;false&#x60;, the &#x60;departmentAdmins&#x60; is set in the returned departments. | 
 **memoryUnitMb** | **bool** | Memory returned in MB. When set to &#x60;false&#x60; (default) memory will be returned in MiB. | 

### Return type

[**[]Department2**](Department2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PatchDepartmentResources

> []Resources PatchDepartmentResources(ctx, departmentId).Resources(resources).Execute()

Patch department resources



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
    departmentId := "1" // string | The id of the department.
    resources := []openapiclient.Resources{*openapiclient.NewResources(*openapiclient.NewResourcesGpu(float32(1000)))} // []Resources | Department resources to update.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.PatchDepartmentResources(context.Background(), departmentId).Resources(resources).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.PatchDepartmentResources``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `PatchDepartmentResources`: []Resources
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.PatchDepartmentResources`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchDepartmentResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **resources** | [**[]Resources**](Resources.md) | Department resources to update. | 

### Return type

[**[]Resources**](Resources.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDepartment

> Department UpdateDepartment(ctx, departmentId).DepartmentUpdateRequest(departmentUpdateRequest).Execute()

Update department



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
    departmentId := "1" // string | The id of the department.
    departmentUpdateRequest := *openapiclient.NewDepartmentUpdateRequest([]openapiclient.Resources{*openapiclient.NewResources(*openapiclient.NewResourcesGpu(float32(1000)))}) // DepartmentUpdateRequest | Department to update.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.UpdateDepartment(context.Background(), departmentId).DepartmentUpdateRequest(departmentUpdateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.UpdateDepartment``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateDepartment`: Department
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.UpdateDepartment`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDepartmentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **departmentUpdateRequest** | [**DepartmentUpdateRequest**](DepartmentUpdateRequest.md) | Department to update. | 

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
    openapiclient "./openapi"
)

func main() {
    clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The unique uuid identifying the cluster.
    departmentId := int32(2) // int32 | The unique id identifying the department.
    departmentAccessControl := *openapiclient.NewDepartmentAccessControl() // DepartmentAccessControl | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.UpdateDepartmentAdmins(context.Background(), clusterId, departmentId).DepartmentAccessControl(departmentAccessControl).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.UpdateDepartmentAdmins``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateDepartmentAdmins`: DepartmentAccessControl
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.UpdateDepartmentAdmins`: %v\n", resp)
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


## UpdateDepartmentResources

> []Resources UpdateDepartmentResources(ctx, departmentId).Resources(resources).Execute()

Update department resources



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
    departmentId := "1" // string | The id of the department.
    resources := []openapiclient.Resources{*openapiclient.NewResources(*openapiclient.NewResourcesGpu(float32(1000)))} // []Resources | Department resources to update.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.UpdateDepartmentResources(context.Background(), departmentId).Resources(resources).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.UpdateDepartmentResources``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateDepartmentResources`: []Resources
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.UpdateDepartmentResources`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDepartmentResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **resources** | [**[]Resources**](Resources.md) | Department resources to update. | 

### Return type

[**[]Resources**](Resources.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDepartment_0

> Department2 UpdateDepartment_0(ctx, clusterId, departmentId).DepartmentUpdateRequest1(departmentUpdateRequest1).Execute()

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
    departmentUpdateRequest1 := *openapiclient.NewDepartmentUpdateRequest1("default") // DepartmentUpdateRequest1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DepartmentsApi.UpdateDepartment_0(context.Background(), clusterId, departmentId).DepartmentUpdateRequest1(departmentUpdateRequest1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsApi.UpdateDepartment_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateDepartment_0`: Department2
    fmt.Fprintf(os.Stdout, "Response from `DepartmentsApi.UpdateDepartment_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The unique uuid identifying the cluster. | 
**departmentId** | **int32** | The unique id identifying the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDepartment_6Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **departmentUpdateRequest1** | [**DepartmentUpdateRequest1**](DepartmentUpdateRequest1.md) |  | 

### Return type

[**Department2**](Department2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

