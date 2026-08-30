# \ProjectsV1API

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetProjectMetricsV1**](ProjectsV1API.md#GetProjectMetricsV1) | **Get** /v1/k8s/clusters/{clusterUuid}/projects/{projectId}/metrics | Get metrics data for a specific project.
[**GetProjectsMetricsV1**](ProjectsV1API.md#GetProjectsMetricsV1) | **Get** /v1/k8s/clusters/{clusterUuid}/projects/metrics | Get metrics for all projects.
[**GetProjectsQuotasV1**](ProjectsV1API.md#GetProjectsQuotasV1) | **Get** /v1/k8s/clusters/{clusterId}/projects/quotas | Get project quotas.
[**GetProjectsV1**](ProjectsV1API.md#GetProjectsV1) | **Get** /v1/k8s/clusters/{clusterId}/projects | List all projects and their details.
[**UpdateProjectStatusV1**](ProjectsV1API.md#UpdateProjectStatusV1) | **Put** /v1/k8s/clusters/{clusterId}/projects/status | Update Project status



## GetProjectMetricsV1

> ProjectMetricsV1Response GetProjectMetricsV1(ctx, clusterUuid, projectId).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()

Get metrics data for a specific project.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/run-ai/runai/runai-clients/org-unit-service-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id
	start := time.Now() // time.Time | Start of time range to fetch data from in UTC format. (optional)
	end := time.Now() // time.Time | End of time range to fetch data from in UTC format. (optional)
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)
	nodepoolName := "default" // string | Filter by unique nodepool name. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsV1API.GetProjectMetricsV1(context.Background(), clusterUuid, projectId).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsV1API.GetProjectMetricsV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjectMetricsV1`: ProjectMetricsV1Response
	fmt.Fprintf(os.Stdout, "Response from `ProjectsV1API.GetProjectMetricsV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectMetricsV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **start** | **time.Time** | Start of time range to fetch data from in UTC format. | 
 **end** | **time.Time** | End of time range to fetch data from in UTC format. | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]
 **nodepoolName** | **string** | Filter by unique nodepool name. | 

### Return type

[**ProjectMetricsV1Response**](ProjectMetricsV1Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectsMetricsV1

> ProjectsMetricsV1Response GetProjectsMetricsV1(ctx, clusterUuid).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()

Get metrics for all projects.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/run-ai/runai/runai-clients/org-unit-service-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	start := time.Now() // time.Time | Start of time range to fetch data from in UTC format. (optional)
	end := time.Now() // time.Time | End of time range to fetch data from in UTC format. (optional)
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)
	nodepoolName := "default" // string | Filter by unique nodepool name. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsV1API.GetProjectsMetricsV1(context.Background(), clusterUuid).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsV1API.GetProjectsMetricsV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjectsMetricsV1`: ProjectsMetricsV1Response
	fmt.Fprintf(os.Stdout, "Response from `ProjectsV1API.GetProjectsMetricsV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectsMetricsV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **start** | **time.Time** | Start of time range to fetch data from in UTC format. | 
 **end** | **time.Time** | End of time range to fetch data from in UTC format. | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]
 **nodepoolName** | **string** | Filter by unique nodepool name. | 

### Return type

[**ProjectsMetricsV1Response**](ProjectsMetricsV1Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectsQuotasV1

> []ProjectQuotasV1Inner GetProjectsQuotasV1(ctx, clusterId).Execute()

Get project quotas.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/org-unit-service-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsV1API.GetProjectsQuotasV1(context.Background(), clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsV1API.GetProjectsQuotasV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjectsQuotasV1`: []ProjectQuotasV1Inner
	fmt.Fprintf(os.Stdout, "Response from `ProjectsV1API.GetProjectsQuotasV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectsQuotasV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]ProjectQuotasV1Inner**](ProjectQuotasV1Inner.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectsV1

> []ProjectV1 GetProjectsV1(ctx, clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()

List all projects and their details.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/org-unit-service-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	excludePermissions := true // bool | Backward compatibility of the `departmentAdmins` field. If `true`, the `departmentAdmins` field in the returned departments is not set. If `false`, the `departmentAdmins` is set in the returned departments. (optional)
	memoryUnitMb := true // bool | Memory returned in MB. When set to `false` (default) memory will be returned in MiB. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsV1API.GetProjectsV1(context.Background(), clusterId).ExcludePermissions(excludePermissions).MemoryUnitMb(memoryUnitMb).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsV1API.GetProjectsV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjectsV1`: []ProjectV1
	fmt.Fprintf(os.Stdout, "Response from `ProjectsV1API.GetProjectsV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectsV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **excludePermissions** | **bool** | Backward compatibility of the &#x60;departmentAdmins&#x60; field. If &#x60;true&#x60;, the &#x60;departmentAdmins&#x60; field in the returned departments is not set. If &#x60;false&#x60;, the &#x60;departmentAdmins&#x60; is set in the returned departments. | 
 **memoryUnitMb** | **bool** | Memory returned in MB. When set to &#x60;false&#x60; (default) memory will be returned in MiB. | 

### Return type

[**[]ProjectV1**](ProjectV1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateProjectStatusV1

> UpdateProjectStatusV1(ctx, clusterId).ProjectV1StatusRequest(projectV1StatusRequest).Execute()

Update Project status

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/org-unit-service-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	projectV1StatusRequest := *openapiclient.NewProjectV1StatusRequest() // ProjectV1StatusRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProjectsV1API.UpdateProjectStatusV1(context.Background(), clusterId).ProjectV1StatusRequest(projectV1StatusRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsV1API.UpdateProjectStatusV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateProjectStatusV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **projectV1StatusRequest** | [**ProjectV1StatusRequest**](ProjectV1StatusRequest.md) |  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

