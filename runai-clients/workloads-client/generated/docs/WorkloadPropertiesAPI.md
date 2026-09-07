# \WorkloadPropertiesAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateWorkloadPriorities**](WorkloadPropertiesAPI.md#CreateWorkloadPriorities) | **Post** /api/v1/workload-priorities | Create workload priorities.
[**CreateWorkloadType**](WorkloadPropertiesAPI.md#CreateWorkloadType) | **Post** /api/v1/workload-types | Create a workload type. [Experimental]
[**DeleteWorkloadType**](WorkloadPropertiesAPI.md#DeleteWorkloadType) | **Delete** /api/v1/workload-types/{workloadTypeId} | Delete a workload type by id. [Experimental]
[**GetAvailablePrioritiesAndCategories**](WorkloadPropertiesAPI.md#GetAvailablePrioritiesAndCategories) | **Get** /internal/available-priorities-and-categories | Get available priorities and categories.
[**GetCategoryById**](WorkloadPropertiesAPI.md#GetCategoryById) | **Get** /api/v1/workload-categories/{categoryId} | Get workload category by id.
[**GetSyncWorkloadPriorities**](WorkloadPropertiesAPI.md#GetSyncWorkloadPriorities) | **Get** /api/v1/workload-priorities-sync | Get workload priorities for agent.
[**GetWorkloadClusterPriorities**](WorkloadPropertiesAPI.md#GetWorkloadClusterPriorities) | **Get** /api/v1/workload-cluster-priorities | Get workload priorities.
[**GetWorkloadPriorities**](WorkloadPropertiesAPI.md#GetWorkloadPriorities) | **Get** /api/v1/workload-priorities | Get workload priorities.
[**GetWorkloadType**](WorkloadPropertiesAPI.md#GetWorkloadType) | **Get** /api/v1/workload-types/{workloadTypeId} | List workload type by id.
[**ListCategories**](WorkloadPropertiesAPI.md#ListCategories) | **Get** /api/v1/workload-categories | List workload categories.
[**ListWorkloadTypes**](WorkloadPropertiesAPI.md#ListWorkloadTypes) | **Get** /api/v1/workload-types | List workload types.
[**SyncWorkloadTypes**](WorkloadPropertiesAPI.md#SyncWorkloadTypes) | **Post** /api/v1/workload-types/sync | Sync workload type statuses from cluster.
[**UpdateWorkloadType**](WorkloadPropertiesAPI.md#UpdateWorkloadType) | **Put** /api/v1/workload-types/{workloadTypeId} | Update a workload type by id.
[**WorkloadClusterPrioritiesStatus**](WorkloadPropertiesAPI.md#WorkloadClusterPrioritiesStatus) | **Post** /api/v1/workload-priorities/cluster-status | Sync workload cluster priorities status.
[**WorkloadPrioritiesStatus**](WorkloadPropertiesAPI.md#WorkloadPrioritiesStatus) | **Post** /api/v1/workload-priorities/runai-status | Sync workload priorities status.



## CreateWorkloadPriorities

> WorkloadPriority CreateWorkloadPriorities(ctx).PriorityForCreate(priorityForCreate).Execute()

Create workload priorities.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	priorityForCreate := *openapiclient.NewPriorityForCreate("train", int32(50)) // PriorityForCreate |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadPropertiesAPI.CreateWorkloadPriorities(context.Background()).PriorityForCreate(priorityForCreate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.CreateWorkloadPriorities``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateWorkloadPriorities`: WorkloadPriority
	fmt.Fprintf(os.Stdout, "Response from `WorkloadPropertiesAPI.CreateWorkloadPriorities`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateWorkloadPrioritiesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **priorityForCreate** | [**PriorityForCreate**](PriorityForCreate.md) |  | 

### Return type

[**WorkloadPriority**](WorkloadPriority.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateWorkloadType

> WorkloadTypeConfig CreateWorkloadType(ctx).WorkloadTypeCreateFields(workloadTypeCreateFields).Execute()

Create a workload type. [Experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	workloadTypeCreateFields := *openapiclient.NewWorkloadTypeCreateFields("CategoryId_example", "PriorityId_example", "Deployment", "apps") // WorkloadTypeCreateFields | Workload type to create.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadPropertiesAPI.CreateWorkloadType(context.Background()).WorkloadTypeCreateFields(workloadTypeCreateFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.CreateWorkloadType``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateWorkloadType`: WorkloadTypeConfig
	fmt.Fprintf(os.Stdout, "Response from `WorkloadPropertiesAPI.CreateWorkloadType`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateWorkloadTypeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadTypeCreateFields** | [**WorkloadTypeCreateFields**](WorkloadTypeCreateFields.md) | Workload type to create. | 

### Return type

[**WorkloadTypeConfig**](WorkloadTypeConfig.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteWorkloadType

> DeleteWorkloadType(ctx, workloadTypeId).Execute()

Delete a workload type by id. [Experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	workloadTypeId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier of the workload type.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadPropertiesAPI.DeleteWorkloadType(context.Background(), workloadTypeId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.DeleteWorkloadType``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadTypeId** | **string** | The unique identifier of the workload type. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteWorkloadTypeRequest struct via the builder pattern


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


## GetAvailablePrioritiesAndCategories

> GetAvailablePrioritiesAndCategories200Response GetAvailablePrioritiesAndCategories(ctx).ClusterId(clusterId).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()

Get available priorities and categories.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	sortBy := "sortBy_example" // string |  (optional)
	filterBy := []string{"Inner_example"} // []string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadPropertiesAPI.GetAvailablePrioritiesAndCategories(context.Background()).ClusterId(clusterId).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.GetAvailablePrioritiesAndCategories``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAvailablePrioritiesAndCategories`: GetAvailablePrioritiesAndCategories200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkloadPropertiesAPI.GetAvailablePrioritiesAndCategories`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetAvailablePrioritiesAndCategoriesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **sortBy** | **string** |  | 
 **filterBy** | **[]string** |  | 

### Return type

[**GetAvailablePrioritiesAndCategories200Response**](GetAvailablePrioritiesAndCategories200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetCategoryById

> WorkloadCategory GetCategoryById(ctx, categoryId).Execute()

Get workload category by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	categoryId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier of the workload category.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadPropertiesAPI.GetCategoryById(context.Background(), categoryId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.GetCategoryById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetCategoryById`: WorkloadCategory
	fmt.Fprintf(os.Stdout, "Response from `WorkloadPropertiesAPI.GetCategoryById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**categoryId** | **string** | The unique identifier of the workload category. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetCategoryByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**WorkloadCategory**](WorkloadCategory.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSyncWorkloadPriorities

> []SyncPriority GetSyncWorkloadPriorities(ctx).ClusterId(clusterId).Execute()

Get workload priorities for agent.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadPropertiesAPI.GetSyncWorkloadPriorities(context.Background()).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.GetSyncWorkloadPriorities``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSyncWorkloadPriorities`: []SyncPriority
	fmt.Fprintf(os.Stdout, "Response from `WorkloadPropertiesAPI.GetSyncWorkloadPriorities`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSyncWorkloadPrioritiesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 

### Return type

[**[]SyncPriority**](SyncPriority.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadClusterPriorities

> GetWorkloadClusterPriorities200Response GetWorkloadClusterPriorities(ctx).ClusterId(clusterId).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()

Get workload priorities.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	sortBy := "sortBy_example" // string |  (optional)
	filterBy := []string{"Inner_example"} // []string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadPropertiesAPI.GetWorkloadClusterPriorities(context.Background()).ClusterId(clusterId).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.GetWorkloadClusterPriorities``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadClusterPriorities`: GetWorkloadClusterPriorities200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkloadPropertiesAPI.GetWorkloadClusterPriorities`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadClusterPrioritiesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **sortBy** | **string** |  | 
 **filterBy** | **[]string** |  | 

### Return type

[**GetWorkloadClusterPriorities200Response**](GetWorkloadClusterPriorities200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadPriorities

> GetWorkloadPriorities200Response GetWorkloadPriorities(ctx).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()

Get workload priorities.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	sortBy := "sortBy_example" // string | Defines the field to sort the results by (optional)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >= (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadPropertiesAPI.GetWorkloadPriorities(context.Background()).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.GetWorkloadPriorities``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadPriorities`: GetWorkloadPriorities200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkloadPropertiesAPI.GetWorkloadPriorities`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadPrioritiesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **sortBy** | **string** | Defines the field to sort the results by | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D; | 

### Return type

[**GetWorkloadPriorities200Response**](GetWorkloadPriorities200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadType

> WorkloadTypeConfig GetWorkloadType(ctx, workloadTypeId).Execute()

List workload type by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	workloadTypeId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier of the workload type.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadPropertiesAPI.GetWorkloadType(context.Background(), workloadTypeId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.GetWorkloadType``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadType`: WorkloadTypeConfig
	fmt.Fprintf(os.Stdout, "Response from `WorkloadPropertiesAPI.GetWorkloadType`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadTypeId** | **string** | The unique identifier of the workload type. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadTypeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**WorkloadTypeConfig**](WorkloadTypeConfig.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListCategories

> ListCategories200Response ListCategories(ctx).Execute()

List workload categories.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadPropertiesAPI.ListCategories(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.ListCategories``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListCategories`: ListCategories200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkloadPropertiesAPI.ListCategories`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListCategoriesRequest struct via the builder pattern


### Return type

[**ListCategories200Response**](ListCategories200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListWorkloadTypes

> ListWorkloadTypes200Response ListWorkloadTypes(ctx).ExternalTypesOnly(externalTypesOnly).FilterBy(filterBy).NoStatus(noStatus).Execute()

List workload types.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	externalTypesOnly := true // bool | Return only external and user_external sourced workload types. (optional) (default to false)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)
	noStatus := true // bool | Return workload types without status. (optional) (default to false)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadPropertiesAPI.ListWorkloadTypes(context.Background()).ExternalTypesOnly(externalTypesOnly).FilterBy(filterBy).NoStatus(noStatus).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.ListWorkloadTypes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListWorkloadTypes`: ListWorkloadTypes200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkloadPropertiesAPI.ListWorkloadTypes`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListWorkloadTypesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **externalTypesOnly** | **bool** | Return only external and user_external sourced workload types. | [default to false]
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contain, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 
 **noStatus** | **bool** | Return workload types without status. | [default to false]

### Return type

[**ListWorkloadTypes200Response**](ListWorkloadTypes200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncWorkloadTypes

> SyncWorkloadTypes(ctx).ClusterId(clusterId).WorkloadTypesForSync(workloadTypesForSync).Execute()

Sync workload type statuses from cluster.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	workloadTypesForSync := *openapiclient.NewWorkloadTypesForSync() // WorkloadTypesForSync |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadPropertiesAPI.SyncWorkloadTypes(context.Background()).ClusterId(clusterId).WorkloadTypesForSync(workloadTypesForSync).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.SyncWorkloadTypes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncWorkloadTypesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **workloadTypesForSync** | [**WorkloadTypesForSync**](WorkloadTypesForSync.md) |  | 

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


## UpdateWorkloadType

> WorkloadTypeConfig UpdateWorkloadType(ctx, workloadTypeId).WorkloadTypeUpdateFields(workloadTypeUpdateFields).Execute()

Update a workload type by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	workloadTypeId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier of the workload type.
	workloadTypeUpdateFields := *openapiclient.NewWorkloadTypeUpdateFields() // WorkloadTypeUpdateFields | Workload type to update.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadPropertiesAPI.UpdateWorkloadType(context.Background(), workloadTypeId).WorkloadTypeUpdateFields(workloadTypeUpdateFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.UpdateWorkloadType``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateWorkloadType`: WorkloadTypeConfig
	fmt.Fprintf(os.Stdout, "Response from `WorkloadPropertiesAPI.UpdateWorkloadType`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadTypeId** | **string** | The unique identifier of the workload type. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateWorkloadTypeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadTypeUpdateFields** | [**WorkloadTypeUpdateFields**](WorkloadTypeUpdateFields.md) | Workload type to update. | 

### Return type

[**WorkloadTypeConfig**](WorkloadTypeConfig.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## WorkloadClusterPrioritiesStatus

> WorkloadClusterPrioritiesStatus(ctx).ClusterId(clusterId).PrioritiesForStatus(prioritiesForStatus).Execute()

Sync workload cluster priorities status.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	prioritiesForStatus := *openapiclient.NewPrioritiesForStatus() // PrioritiesForStatus |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadPropertiesAPI.WorkloadClusterPrioritiesStatus(context.Background()).ClusterId(clusterId).PrioritiesForStatus(prioritiesForStatus).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.WorkloadClusterPrioritiesStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiWorkloadClusterPrioritiesStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **prioritiesForStatus** | [**PrioritiesForStatus**](PrioritiesForStatus.md) |  | 

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


## WorkloadPrioritiesStatus

> WorkloadPrioritiesStatus(ctx).ClusterId(clusterId).PrioritiesForStatus(prioritiesForStatus).Execute()

Sync workload priorities status.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	prioritiesForStatus := *openapiclient.NewPrioritiesForStatus() // PrioritiesForStatus |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadPropertiesAPI.WorkloadPrioritiesStatus(context.Background()).ClusterId(clusterId).PrioritiesForStatus(prioritiesForStatus).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadPropertiesAPI.WorkloadPrioritiesStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiWorkloadPrioritiesStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **prioritiesForStatus** | [**PrioritiesForStatus**](PrioritiesForStatus.md) |  | 

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

