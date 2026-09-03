# \DepartmentsV1API

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetDepartmentMetricsV1**](DepartmentsV1API.md#GetDepartmentMetricsV1) | **Get** /v1/k8s/clusters/{clusterUuid}/departments/{departmentId}/metrics | Get metrics for a specific department.
[**GetDepartmentsMetricsV1**](DepartmentsV1API.md#GetDepartmentsMetricsV1) | **Get** /v1/k8s/clusters/{clusterUuid}/departments/metrics | Get metrics for all departments.
[**GetDepartmentsV1**](DepartmentsV1API.md#GetDepartmentsV1) | **Get** /v1/k8s/clusters/{clusterId}/departments | List all departments.
[**SetDepartmentAdminsV1**](DepartmentsV1API.md#SetDepartmentAdminsV1) | **Put** /v1/k8s/clusters/{clusterId}/departments/{department-id}/access-control | Set the department admins.



## GetDepartmentMetricsV1

> DepartmentMetricsV1Response GetDepartmentMetricsV1(ctx, clusterUuid, departmentId).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()

Get metrics for a specific department.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
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
	resp, r, err := apiClient.DepartmentsV1API.GetDepartmentMetricsV1(context.Background(), clusterUuid, departmentId).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsV1API.GetDepartmentMetricsV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDepartmentMetricsV1`: DepartmentMetricsV1Response
	fmt.Fprintf(os.Stdout, "Response from `DepartmentsV1API.GetDepartmentMetricsV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartmentMetricsV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **start** | **time.Time** | Start of time range to fetch data from in UTC format. | 
 **end** | **time.Time** | End of time range to fetch data from in UTC format. | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]
 **nodepoolName** | **string** | Filter by unique nodepool name. | 

### Return type

[**DepartmentMetricsV1Response**](DepartmentMetricsV1Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDepartmentsMetricsV1

> DepartmentsMetricsV1Response GetDepartmentsMetricsV1(ctx, clusterUuid).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()

Get metrics for all departments.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	start := time.Now() // time.Time | Start of time range to fetch data from in UTC format. (optional)
	end := time.Now() // time.Time | End of time range to fetch data from in UTC format. (optional)
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)
	nodepoolName := "default" // string | Filter by unique nodepool name. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DepartmentsV1API.GetDepartmentsMetricsV1(context.Background(), clusterUuid).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsV1API.GetDepartmentsMetricsV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDepartmentsMetricsV1`: DepartmentsMetricsV1Response
	fmt.Fprintf(os.Stdout, "Response from `DepartmentsV1API.GetDepartmentsMetricsV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartmentsMetricsV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **start** | **time.Time** | Start of time range to fetch data from in UTC format. | 
 **end** | **time.Time** | End of time range to fetch data from in UTC format. | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]
 **nodepoolName** | **string** | Filter by unique nodepool name. | 

### Return type

[**DepartmentsMetricsV1Response**](DepartmentsMetricsV1Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDepartmentsV1

> []DepartmentForGetV1 GetDepartmentsV1(ctx, clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()

List all departments.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	excludePermissions := true // bool | Backward compatibility of the `departmentAdmins` field. If `true`, the `departmentAdmins` field in the returned departments is not set. If `false`, the `departmentAdmins` is set in the returned departments. (optional)
	memoryUnitMb := true // bool | Memory returned in MB. When set to `false` (default) memory will be returned in MiB. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DepartmentsV1API.GetDepartmentsV1(context.Background(), clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsV1API.GetDepartmentsV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDepartmentsV1`: []DepartmentForGetV1
	fmt.Fprintf(os.Stdout, "Response from `DepartmentsV1API.GetDepartmentsV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDepartmentsV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **excludePermissions** | **bool** | Backward compatibility of the &#x60;departmentAdmins&#x60; field. If &#x60;true&#x60;, the &#x60;departmentAdmins&#x60; field in the returned departments is not set. If &#x60;false&#x60;, the &#x60;departmentAdmins&#x60; is set in the returned departments. | 
 **memoryUnitMb** | **bool** | Memory returned in MB. When set to &#x60;false&#x60; (default) memory will be returned in MiB. | 

### Return type

[**[]DepartmentForGetV1**](DepartmentForGetV1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetDepartmentAdminsV1

> DepartmentV1AccessControl SetDepartmentAdminsV1(ctx, clusterId, departmentId).DepartmentV1AccessControl(departmentV1AccessControl).Execute()

Set the department admins.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	departmentId := int32(2) // int32 | The unique id identifying the department.
	departmentV1AccessControl := *openapiclient.NewDepartmentV1AccessControl() // DepartmentV1AccessControl | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DepartmentsV1API.SetDepartmentAdminsV1(context.Background(), clusterId, departmentId).DepartmentV1AccessControl(departmentV1AccessControl).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DepartmentsV1API.SetDepartmentAdminsV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SetDepartmentAdminsV1`: DepartmentV1AccessControl
	fmt.Fprintf(os.Stdout, "Response from `DepartmentsV1API.SetDepartmentAdminsV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **int32** | The unique id identifying the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetDepartmentAdminsV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **departmentV1AccessControl** | [**DepartmentV1AccessControl**](DepartmentV1AccessControl.md) |  | 

### Return type

[**DepartmentV1AccessControl**](DepartmentV1AccessControl.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

