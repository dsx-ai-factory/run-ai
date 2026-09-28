# \ProjectsAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**BatchPatchProjectNodePoolResources**](ProjectsAPI.md#BatchPatchProjectNodePoolResources) | **Post** /api/v1/org-unit/projects/node-pool-resources/batch-patch | Patch multiple projects node pool resources in batch.
[**CountProjectNodePoolResources**](ProjectsAPI.md#CountProjectNodePoolResources) | **Get** /api/v1/org-unit/projects/node-pool-resources/count | Count projects node pool resources
[**CountProjects**](ProjectsAPI.md#CountProjects) | **Get** /api/v1/org-unit/projects/count | Count projects
[**CreateProject**](ProjectsAPI.md#CreateProject) | **Post** /api/v1/org-unit/projects | Create project
[**DeleteProject**](ProjectsAPI.md#DeleteProject) | **Delete** /api/v1/org-unit/projects/{projectId} | Delete project
[**GetInternalProjectsTelemetry**](ProjectsAPI.md#GetInternalProjectsTelemetry) | **Get** /internal/org-unit/projects/telemetry | Get internal projects telemetry
[**GetProject**](ProjectsAPI.md#GetProject) | **Get** /api/v1/org-unit/projects/{projectId} | Get project
[**GetProjectAncestors**](ProjectsAPI.md#GetProjectAncestors) | **Get** /internal/org-unit/projects/{projectId}/ancestors | Get project ancestors
[**GetProjectMetrics**](ProjectsAPI.md#GetProjectMetrics) | **Get** /api/v1/org-unit/projects/{projectId}/metrics | Get project metrics data.
[**GetProjectNodePoolResources**](ProjectsAPI.md#GetProjectNodePoolResources) | **Get** /api/v1/org-unit/projects/node-pool-resources | Get projects node pool resources
[**GetProjects**](ProjectsAPI.md#GetProjects) | **Get** /api/v1/org-unit/projects | Get projects
[**GetProjectsMetadata**](ProjectsAPI.md#GetProjectsMetadata) | **Get** /api/v1/org-unit/projects/metadata | Get projects metadata
[**GetProjectsRanking**](ProjectsAPI.md#GetProjectsRanking) | **Get** /api/v1/org-unit/projects/ranking | Get top 15 projects by highest or lowest average resource utilization, allocation, or allocation ratio
[**GetProjectsTelemetry**](ProjectsAPI.md#GetProjectsTelemetry) | **Get** /api/v1/org-unit/projects/telemetry | Get projects telemetry
[**GetSyncProjects**](ProjectsAPI.md#GetSyncProjects) | **Get** /api/v1/org-unit/projects/sync | Get projects for sync
[**PatchProjectResources**](ProjectsAPI.md#PatchProjectResources) | **Patch** /api/v1/org-unit/projects/{projectId}/resources | Patch project resources
[**ProjectsResetResources**](ProjectsAPI.md#ProjectsResetResources) | **Post** /api/v1/org-unit/projects/resources/reset | Reset resources of a specific nodepool for all projects in the cluster.
[**SumProjectNodePoolResources**](ProjectsAPI.md#SumProjectNodePoolResources) | **Get** /api/v1/org-unit/projects/node-pool-resources/sum | sum projects node pool resources
[**SyncProjects**](ProjectsAPI.md#SyncProjects) | **Post** /api/v1/org-unit/projects/sync | Sync cluster projects
[**UpdateProject**](ProjectsAPI.md#UpdateProject) | **Put** /api/v1/org-unit/projects/{projectId} | Update project
[**UpdateProjectResources**](ProjectsAPI.md#UpdateProjectResources) | **Put** /api/v1/org-unit/projects/{projectId}/resources | Update project resources
[**UpdateProjectStatus**](ProjectsAPI.md#UpdateProjectStatus) | **Put** /api/v1/org-unit/projects/status | Update project status from cluster



## BatchPatchProjectNodePoolResources

> BatchPatchProjectResourcesResponse BatchPatchProjectNodePoolResources(ctx).BatchPatchNodePoolResourcesRequest(batchPatchNodePoolResourcesRequest).Execute()

Patch multiple projects node pool resources in batch.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	batchPatchNodePoolResourcesRequest := *openapiclient.NewBatchPatchNodePoolResourcesRequest([]openapiclient.NodePoolResourcesPatchRequest{*openapiclient.NewNodePoolResourcesPatchRequest("OrgUnitId_example", "NodePoolId_example")}) // BatchPatchNodePoolResourcesRequest | Batch of project resources patch requests.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.BatchPatchProjectNodePoolResources(context.Background()).BatchPatchNodePoolResourcesRequest(batchPatchNodePoolResourcesRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.BatchPatchProjectNodePoolResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BatchPatchProjectNodePoolResources`: BatchPatchProjectResourcesResponse
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.BatchPatchProjectNodePoolResources`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiBatchPatchProjectNodePoolResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **batchPatchNodePoolResourcesRequest** | [**BatchPatchNodePoolResourcesRequest**](BatchPatchNodePoolResourcesRequest.md) | Batch of project resources patch requests. | 

### Return type

[**BatchPatchProjectResourcesResponse**](BatchPatchProjectResourcesResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CountProjectNodePoolResources

> CountResponse CountProjectNodePoolResources(ctx).FilterBy(filterBy).Search(search).Execute()

Count projects node pool resources



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.CountProjectNodePoolResources(context.Background()).FilterBy(filterBy).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.CountProjectNodePoolResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountProjectNodePoolResources`: CountResponse
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.CountProjectNodePoolResources`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountProjectNodePoolResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **search** | **string** | Filter results by a free text search. | 

### Return type

[**CountResponse**](CountResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CountProjects

> CountResponse CountProjects(ctx).FilterBy(filterBy).Search(search).Execute()

Count projects



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.CountProjects(context.Background()).FilterBy(filterBy).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.CountProjects``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountProjects`: CountResponse
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.CountProjects`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountProjectsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **search** | **string** | Filter results by a free text search. | 

### Return type

[**CountResponse**](CountResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateProject

> Project CreateProject(ctx).ProjectCreationRequest(projectCreationRequest).Execute()

Create project



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	projectCreationRequest := *openapiclient.NewProjectCreationRequest("organization1", "71f69d83-ba66-4822-adf5-55ce55efd210", []openapiclient.ResourcesNullable{*openapiclient.NewResourcesNullable()}) // ProjectCreationRequest | Project to create.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.CreateProject(context.Background()).ProjectCreationRequest(projectCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.CreateProject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateProject`: Project
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.CreateProject`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateProjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectCreationRequest** | [**ProjectCreationRequest**](ProjectCreationRequest.md) | Project to create. | 

### Return type

[**Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteProject

> DeleteProject(ctx, projectId).DeleteType(deleteType).Force(force).Execute()

Delete project



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id
	deleteType := openapiclient.DeletionTypes("IfEmpty") // DeletionTypes | Define project deletion type. (optional)
	force := true // bool | Sets the project to be forced deleted (could leave orphan objects in the system) (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProjectsAPI.DeleteProject(context.Background(), projectId).DeleteType(deleteType).Force(force).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.DeleteProject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteProjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deleteType** | [**DeletionTypes**](DeletionTypes.md) | Define project deletion type. | 
 **force** | **bool** | Sets the project to be forced deleted (could leave orphan objects in the system) | 

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


## GetInternalProjectsTelemetry

> TelemetryResponse GetInternalProjectsTelemetry(ctx).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolId(nodepoolId).DepartmentId(departmentId).ProjectId(projectId).FilterBy(filterBy).GroupBy(groupBy).Execute()

Get internal projects telemetry



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	telemetryType := openapiclient.OrgUnitTelemetryTypeInternal("GPU_QUOTA") // OrgUnitTelemetryTypeInternal | specifies what data to request
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	nodepoolId := "1" // string | Filter using the nodepool. (optional)
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal. Dates are in ISO 8601 timestamp format. (optional)
	groupBy := []string{"GroupBy_example"} // []string | project fields to group the data by. Rank is supported only on the internal projects telemetry API. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.GetInternalProjectsTelemetry(context.Background()).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolId(nodepoolId).DepartmentId(departmentId).ProjectId(projectId).FilterBy(filterBy).GroupBy(groupBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.GetInternalProjectsTelemetry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInternalProjectsTelemetry`: TelemetryResponse
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.GetInternalProjectsTelemetry`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetInternalProjectsTelemetryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **telemetryType** | [**OrgUnitTelemetryTypeInternal**](OrgUnitTelemetryTypeInternal.md) | specifies what data to request | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **nodepoolId** | **string** | Filter using the nodepool. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal. Dates are in ISO 8601 timestamp format. | 
 **groupBy** | **[]string** | project fields to group the data by. Rank is supported only on the internal projects telemetry API. | 

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


## GetProject

> Project GetProject(ctx, projectId).Execute()

Get project



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.GetProject(context.Background(), projectId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.GetProject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProject`: Project
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.GetProject`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectAncestors

> Ancestors GetProjectAncestors(ctx, projectId).Execute()

Get project ancestors

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.GetProjectAncestors(context.Background(), projectId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.GetProjectAncestors``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjectAncestors`: Ancestors
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.GetProjectAncestors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectAncestorsRequest struct via the builder pattern


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


## GetProjectMetrics

> MetricsResponse GetProjectMetrics(ctx, projectId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).GroupBy(groupBy).Execute()

Get project metrics data.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id
	metricType := []openapiclient.OrgUnitMetricType{openapiclient.OrgUnitMetricType("GPU_QUOTA")} // []OrgUnitMetricType | Specify which data to request.
	start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
	end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)
	nodepoolName := "default" // string | Filter using the nodepool. (optional)
	groupBy := "groupBy_example" // string | Field to group the metrics data by. Currently only supported for WORKLOADS_COUNT metric type. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.GetProjectMetrics(context.Background(), projectId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).GroupBy(groupBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.GetProjectMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjectMetrics`: MetricsResponse
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.GetProjectMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **metricType** | [**[]OrgUnitMetricType**](OrgUnitMetricType.md) | Specify which data to request. | 
 **start** | **time.Time** | Start date of time range to fetch data in ISO 8601 timestamp format. | 
 **end** | **time.Time** | End date of time range to fetch data in ISO 8601 timestamp format. | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]
 **nodepoolName** | **string** | Filter using the nodepool. | 
 **groupBy** | **string** | Field to group the metrics data by. Currently only supported for WORKLOADS_COUNT metric type. | 

### Return type

[**MetricsResponse**](MetricsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, text/csv

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectNodePoolResources

> GetProjectNodePoolResources200Response GetProjectNodePoolResources(ctx).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Search(search).Execute()

Get projects node pool resources



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	sortBy := openapiclient.ProjectNodePoolResourcesFilterSortFields("nodePoolName") // ProjectNodePoolResourcesFilterSortFields | Sort results by a parameters. (optional)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.GetProjectNodePoolResources(context.Background()).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.GetProjectNodePoolResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjectNodePoolResources`: GetProjectNodePoolResources200Response
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.GetProjectNodePoolResources`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectNodePoolResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **sortBy** | [**ProjectNodePoolResourcesFilterSortFields**](ProjectNodePoolResourcesFilterSortFields.md) | Sort results by a parameters. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **search** | **string** | Filter results by a free text search. | 

### Return type

[**GetProjectNodePoolResources200Response**](GetProjectNodePoolResources200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjects

> GetProjects200Response GetProjects(ctx).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Search(search).Execute()

Get projects



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	sortBy := openapiclient.ProjectFilterSortFields("name") // ProjectFilterSortFields | Sort results by a parameters. (optional)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.GetProjects(context.Background()).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.GetProjects``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjects`: GetProjects200Response
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.GetProjects`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **sortBy** | [**ProjectFilterSortFields**](ProjectFilterSortFields.md) | Sort results by a parameters. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **search** | **string** | Filter results by a free text search. | 

### Return type

[**GetProjects200Response**](GetProjects200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectsMetadata

> GetProjectsMetadata200Response GetProjectsMetadata(ctx).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Search(search).Execute()

Get projects metadata



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	sortBy := "sortBy_example" // string | Sort results by a parameters. (optional)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.GetProjectsMetadata(context.Background()).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.GetProjectsMetadata``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjectsMetadata`: GetProjectsMetadata200Response
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.GetProjectsMetadata`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectsMetadataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **sortBy** | **string** | Sort results by a parameters. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **search** | **string** | Filter results by a free text search. | 

### Return type

[**GetProjectsMetadata200Response**](GetProjectsMetadata200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectsRanking

> GetProjectsRanking200Response GetProjectsRanking(ctx).MetricType(metricType).Order(order).ClusterId(clusterId).NodepoolName(nodepoolName).DepartmentId(departmentId).Execute()

Get top 15 projects by highest or lowest average resource utilization, allocation, or allocation ratio



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	metricType := openapiclient.RankingMetricType("avgGpuAllocation24h") // RankingMetricType | The type of metric to rank by (top/bottom gpu allocation/utilization)
	order := openapiclient.RankingOrderType("top") // RankingOrderType | The order of the ranking (top/bottom)
	clusterId := "clusterId_example" // string | 
	nodepoolName := "nodepoolName_example" // string | The name of the node pool to filter by (optional)
	departmentId := "departmentId_example" // string | Filter projects by department ID. If the department does not exist or belongs to a different cluster, an empty list is returned. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.GetProjectsRanking(context.Background()).MetricType(metricType).Order(order).ClusterId(clusterId).NodepoolName(nodepoolName).DepartmentId(departmentId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.GetProjectsRanking``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjectsRanking`: GetProjectsRanking200Response
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.GetProjectsRanking`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectsRankingRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **metricType** | [**RankingMetricType**](RankingMetricType.md) | The type of metric to rank by (top/bottom gpu allocation/utilization) | 
 **order** | [**RankingOrderType**](RankingOrderType.md) | The order of the ranking (top/bottom) | 
 **clusterId** | **string** |  | 
 **nodepoolName** | **string** | The name of the node pool to filter by | 
 **departmentId** | **string** | Filter projects by department ID. If the department does not exist or belongs to a different cluster, an empty list is returned. | 

### Return type

[**GetProjectsRanking200Response**](GetProjectsRanking200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetProjectsTelemetry

> TelemetryResponse GetProjectsTelemetry(ctx).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolId(nodepoolId).DepartmentId(departmentId).ProjectId(projectId).GroupBy(groupBy).Execute()

Get projects telemetry



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	telemetryType := openapiclient.OrgUnitTelemetryType("GPU_QUOTA") // OrgUnitTelemetryType | specifies what data to request
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	nodepoolId := "1" // string | Filter using the nodepool. (optional)
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	groupBy := []string{"GroupBy_example"} // []string | project fields to group the data by (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.GetProjectsTelemetry(context.Background()).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolId(nodepoolId).DepartmentId(departmentId).ProjectId(projectId).GroupBy(groupBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.GetProjectsTelemetry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetProjectsTelemetry`: TelemetryResponse
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.GetProjectsTelemetry`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetProjectsTelemetryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **telemetryType** | [**OrgUnitTelemetryType**](OrgUnitTelemetryType.md) | specifies what data to request | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **nodepoolId** | **string** | Filter using the nodepool. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **groupBy** | **[]string** | project fields to group the data by | 

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


## GetSyncProjects

> GetSyncProjects200Response GetSyncProjects(ctx).ClusterId(clusterId).UpdatedAfter(updatedAfter).Offset(offset).Limit(limit).ExcludePermissions(excludePermissions).Execute()

Get projects for sync

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	updatedAfter := time.Now() // time.Time |  (optional)
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	excludePermissions := true // bool |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.GetSyncProjects(context.Background()).ClusterId(clusterId).UpdatedAfter(updatedAfter).Offset(offset).Limit(limit).ExcludePermissions(excludePermissions).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.GetSyncProjects``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSyncProjects`: GetSyncProjects200Response
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.GetSyncProjects`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSyncProjectsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **updatedAfter** | **time.Time** |  | 
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **excludePermissions** | **bool** |  | 

### Return type

[**GetSyncProjects200Response**](GetSyncProjects200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PatchProjectResources

> []Resources PatchProjectResources(ctx, projectId).ResourcesNullable(resourcesNullable).Execute()

Patch project resources



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id
	resourcesNullable := []openapiclient.ResourcesNullable{*openapiclient.NewResourcesNullable()} // []ResourcesNullable | Project resources to patch.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.PatchProjectResources(context.Background(), projectId).ResourcesNullable(resourcesNullable).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.PatchProjectResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PatchProjectResources`: []Resources
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.PatchProjectResources`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchProjectResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **resourcesNullable** | [**[]ResourcesNullable**](ResourcesNullable.md) | Project resources to patch. | 

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


## ProjectsResetResources

> ProjectsResetResources(ctx).ProjectsResourcesResetRequest(projectsResourcesResetRequest).Execute()

Reset resources of a specific nodepool for all projects in the cluster.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	projectsResourcesResetRequest := *openapiclient.NewProjectsResourcesResetRequest("71f69d83-ba66-4822-adf5-55ce55efd210", []string{"22"}, openapiclient.ResourcesResetAction("QUOTAS_TO_DEFAULT")) // ProjectsResourcesResetRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProjectsAPI.ProjectsResetResources(context.Background()).ProjectsResourcesResetRequest(projectsResourcesResetRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.ProjectsResetResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiProjectsResetResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectsResourcesResetRequest** | [**ProjectsResourcesResetRequest**](ProjectsResourcesResetRequest.md) |  | 

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


## SumProjectNodePoolResources

> SumResponse SumProjectNodePoolResources(ctx).NodepoolId(nodepoolId).FilterBy(filterBy).Execute()

sum projects node pool resources



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	nodepoolId := "17" // string | The unique id identifying the Node Pool.
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.SumProjectNodePoolResources(context.Background()).NodepoolId(nodepoolId).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.SumProjectNodePoolResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SumProjectNodePoolResources`: SumResponse
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.SumProjectNodePoolResources`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSumProjectNodePoolResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **nodepoolId** | **string** | The unique id identifying the Node Pool. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 

### Return type

[**SumResponse**](SumResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncProjects

> SyncProjects(ctx).SyncProjectsRequest(syncProjectsRequest).Execute()

Sync cluster projects

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	syncProjectsRequest := *openapiclient.NewSyncProjectsRequest("ClusterId_example", []openapiclient.ProjectInSyncRequest{*openapiclient.NewProjectInSyncRequest("33")}) // SyncProjectsRequest | Projects sync data.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProjectsAPI.SyncProjects(context.Background()).SyncProjectsRequest(syncProjectsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.SyncProjects``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncProjectsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncProjectsRequest** | [**SyncProjectsRequest**](SyncProjectsRequest.md) | Projects sync data. | 

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


## UpdateProject

> Project UpdateProject(ctx, projectId).ProjectUpdateRequest(projectUpdateRequest).Execute()

Update project



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id
	projectUpdateRequest := *openapiclient.NewProjectUpdateRequest([]openapiclient.ResourcesNullable{*openapiclient.NewResourcesNullable()}) // ProjectUpdateRequest | Project to update.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.UpdateProject(context.Background(), projectId).ProjectUpdateRequest(projectUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.UpdateProject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateProject`: Project
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.UpdateProject`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateProjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **projectUpdateRequest** | [**ProjectUpdateRequest**](ProjectUpdateRequest.md) | Project to update. | 

### Return type

[**Project**](Project.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateProjectResources

> []Resources UpdateProjectResources(ctx, projectId).ResourcesNullable(resourcesNullable).Execute()

Update project resources



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	projectId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The project id
	resourcesNullable := []openapiclient.ResourcesNullable{*openapiclient.NewResourcesNullable()} // []ResourcesNullable | Project resources to update.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ProjectsAPI.UpdateProjectResources(context.Background(), projectId).ResourcesNullable(resourcesNullable).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.UpdateProjectResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateProjectResources`: []Resources
	fmt.Fprintf(os.Stdout, "Response from `ProjectsAPI.UpdateProjectResources`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectId** | **string** | The project id | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateProjectResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **resourcesNullable** | [**[]ResourcesNullable**](ResourcesNullable.md) | Project resources to update. | 

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


## UpdateProjectStatus

> UpdateProjectStatus(ctx).ProjectStatusUpdateRequest(projectStatusUpdateRequest).Execute()

Update project status from cluster

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

func main() {
	projectStatusUpdateRequest := *openapiclient.NewProjectStatusUpdateRequest("33", "71f69d83-ba66-4822-adf5-55ce55efd210", *openapiclient.NewProjectStatus()) // ProjectStatusUpdateRequest | Project to update by name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ProjectsAPI.UpdateProjectStatus(context.Background()).ProjectStatusUpdateRequest(projectStatusUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ProjectsAPI.UpdateProjectStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateProjectStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectStatusUpdateRequest** | [**ProjectStatusUpdateRequest**](ProjectStatusUpdateRequest.md) | Project to update by name | 

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

