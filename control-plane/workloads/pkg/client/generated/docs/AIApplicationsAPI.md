# \AIApplicationsAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CountApplications**](AIApplicationsAPI.md#CountApplications) | **Get** /api/v1/ai-applications/count | Count AI applications. [Experimental]
[**CreateApplication**](AIApplicationsAPI.md#CreateApplication) | **Post** /api/v1/ai-applications | Create AI application.
[**DeleteApplication**](AIApplicationsAPI.md#DeleteApplication) | **Delete** /api/v1/ai-applications/{aiApplicationId} | Delete AI application.
[**DownloadApplicationChart**](AIApplicationsAPI.md#DownloadApplicationChart) | **Get** /api/v1/ai-applications/charts/{chartId}/content | Download an uploaded Helm chart archive.
[**GetApplication**](AIApplicationsAPI.md#GetApplication) | **Get** /api/v1/ai-applications/{aiApplicationId} | Get AI application. [Experimental]
[**InspectApplicationChart**](AIApplicationsAPI.md#InspectApplicationChart) | **Post** /api/v1/ai-applications/charts/inspect | Inspect a Helm chart questions declaration.
[**ListApplications**](AIApplicationsAPI.md#ListApplications) | **Get** /api/v1/ai-applications | List AI applications. [Experimental]
[**RequestApplicationsUpdates**](AIApplicationsAPI.md#RequestApplicationsUpdates) | **Post** /api/v1/ai-applications/sync-report | Request applications updates
[**SyncApplications**](AIApplicationsAPI.md#SyncApplications) | **Post** /api/v1/ai-applications/sync | Sync AI applications.
[**UploadApplicationChart**](AIApplicationsAPI.md#UploadApplicationChart) | **Post** /api/v1/ai-applications/charts | Upload a Helm chart archive.



## CountApplications

> CountResponse CountApplications(ctx).FilterBy(filterBy).Search(search).Execute()

Count AI applications. [Experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIApplicationsAPI.CountApplications(context.Background()).FilterBy(filterBy).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIApplicationsAPI.CountApplications``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountApplications`: CountResponse
	fmt.Fprintf(os.Stdout, "Response from `AIApplicationsAPI.CountApplications`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountApplicationsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contain, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 
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


## CreateApplication

> AIApplication CreateApplication(ctx).AIApplicationCreationRequest(aIApplicationCreationRequest).Execute()

Create AI application.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	aIApplicationCreationRequest := *openapiclient.NewAIApplicationCreationRequest("71f69d83-ba66-4822-adf5-55ce55efd210", "1", "Name_example", *openapiclient.NewAIApplicationSpec(*openapiclient.NewAIApplicationSpecChartRef())) // AIApplicationCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIApplicationsAPI.CreateApplication(context.Background()).AIApplicationCreationRequest(aIApplicationCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIApplicationsAPI.CreateApplication``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateApplication`: AIApplication
	fmt.Fprintf(os.Stdout, "Response from `AIApplicationsAPI.CreateApplication`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateApplicationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **aIApplicationCreationRequest** | [**AIApplicationCreationRequest**](AIApplicationCreationRequest.md) |  | 

### Return type

[**AIApplication**](AIApplication.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteApplication

> HttpResponse DeleteApplication(ctx, aiApplicationId).Execute()

Delete AI application.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	aiApplicationId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier of the application.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIApplicationsAPI.DeleteApplication(context.Background(), aiApplicationId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIApplicationsAPI.DeleteApplication``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteApplication`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `AIApplicationsAPI.DeleteApplication`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**aiApplicationId** | **string** | The unique identifier of the application. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteApplicationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HttpResponse**](HttpResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DownloadApplicationChart

> *os.File DownloadApplicationChart(ctx, chartId).Execute()

Download an uploaded Helm chart archive.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	chartId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier of the uploaded Helm chart archive.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIApplicationsAPI.DownloadApplicationChart(context.Background(), chartId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIApplicationsAPI.DownloadApplicationChart``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DownloadApplicationChart`: *os.File
	fmt.Fprintf(os.Stdout, "Response from `AIApplicationsAPI.DownloadApplicationChart`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**chartId** | **string** | The unique identifier of the uploaded Helm chart archive. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDownloadApplicationChartRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[***os.File**](*os.File.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/octet-stream, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetApplication

> AIApplication GetApplication(ctx, aiApplicationId).Execute()

Get AI application. [Experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	aiApplicationId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier of the application.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIApplicationsAPI.GetApplication(context.Background(), aiApplicationId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIApplicationsAPI.GetApplication``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetApplication`: AIApplication
	fmt.Fprintf(os.Stdout, "Response from `AIApplicationsAPI.GetApplication`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**aiApplicationId** | **string** | The unique identifier of the application. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetApplicationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**AIApplication**](AIApplication.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InspectApplicationChart

> AIApplicationChartInspectionResponse InspectApplicationChart(ctx).AIApplicationChartInspectionRequest(aIApplicationChartInspectionRequest).Execute()

Inspect a Helm chart questions declaration.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	aIApplicationChartInspectionRequest := *openapiclient.NewAIApplicationChartInspectionRequest(*openapiclient.NewAIApplicationSpecChartRef()) // AIApplicationChartInspectionRequest | Chart source selection and optional saved credential reference, up to 64 KiB of JSON.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIApplicationsAPI.InspectApplicationChart(context.Background()).AIApplicationChartInspectionRequest(aIApplicationChartInspectionRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIApplicationsAPI.InspectApplicationChart``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `InspectApplicationChart`: AIApplicationChartInspectionResponse
	fmt.Fprintf(os.Stdout, "Response from `AIApplicationsAPI.InspectApplicationChart`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInspectApplicationChartRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **aIApplicationChartInspectionRequest** | [**AIApplicationChartInspectionRequest**](AIApplicationChartInspectionRequest.md) | Chart source selection and optional saved credential reference, up to 64 KiB of JSON. | 

### Return type

[**AIApplicationChartInspectionResponse**](AIApplicationChartInspectionResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListApplications

> ListApplications200Response ListApplications(ctx).Verbosity(verbosity).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Search(search).Execute()

List AI applications. [Experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	verbosity := openapiclient.AIApplicationVerbosity("brief") // AIApplicationVerbosity | response verbosity level. (optional) (default to "brief")
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	sortBy := openapiclient.AIApplicationsSortFilterFields("clusterId") // AIApplicationsSortFilterFields | Sort results by a parameter. (optional)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIApplicationsAPI.ListApplications(context.Background()).Verbosity(verbosity).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIApplicationsAPI.ListApplications``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListApplications`: ListApplications200Response
	fmt.Fprintf(os.Stdout, "Response from `AIApplicationsAPI.ListApplications`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListApplicationsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **verbosity** | [**AIApplicationVerbosity**](AIApplicationVerbosity.md) | response verbosity level. | [default to &quot;brief&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **sortBy** | [**AIApplicationsSortFilterFields**](AIApplicationsSortFilterFields.md) | Sort results by a parameter. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contain, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 
 **search** | **string** | Filter results by a free text search. | 

### Return type

[**ListApplications200Response**](ListApplications200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RequestApplicationsUpdates

> AIApplicationSyncResponse RequestApplicationsUpdates(ctx).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()

Request applications updates



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	updatesAfter := int64(12345) // int64 | list only data with update counter greater than the provided value. for example updatesAfter=5 should fetch data with updates counter 6 and above.
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIApplicationsAPI.RequestApplicationsUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIApplicationsAPI.RequestApplicationsUpdates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RequestApplicationsUpdates`: AIApplicationSyncResponse
	fmt.Fprintf(os.Stdout, "Response from `AIApplicationsAPI.RequestApplicationsUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRequestApplicationsUpdatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updatesAfter** | **int64** | list only data with update counter greater than the provided value. for example updatesAfter&#x3D;5 should fetch data with updates counter 6 and above. | 
 **clusterId** | **string** | The id of the cluster | 
 **clusterSyncReport** | [**ClusterSyncReport**](ClusterSyncReport.md) |  | 

### Return type

[**AIApplicationSyncResponse**](AIApplicationSyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncApplications

> SyncApplications(ctx).ClusterId(clusterId).AIApplicationsForSync(aIApplicationsForSync).Execute()

Sync AI applications.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	aIApplicationsForSync := *openapiclient.NewAIApplicationsForSync() // AIApplicationsForSync |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.AIApplicationsAPI.SyncApplications(context.Background()).ClusterId(clusterId).AIApplicationsForSync(aIApplicationsForSync).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIApplicationsAPI.SyncApplications``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncApplicationsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **aIApplicationsForSync** | [**AIApplicationsForSync**](AIApplicationsForSync.md) |  | 

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


## UploadApplicationChart

> AIApplicationChartUploadResponse UploadApplicationChart(ctx).Chart(chart).Execute()

Upload a Helm chart archive.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	chart := os.NewFile(1234, "some_file") // *os.File | The packaged Helm chart archive (.tgz), up to 10 MB.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AIApplicationsAPI.UploadApplicationChart(context.Background()).Chart(chart).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AIApplicationsAPI.UploadApplicationChart``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UploadApplicationChart`: AIApplicationChartUploadResponse
	fmt.Fprintf(os.Stdout, "Response from `AIApplicationsAPI.UploadApplicationChart`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUploadApplicationChartRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **chart** | ***os.File** | The packaged Helm chart archive (.tgz), up to 10 MB. | 

### Return type

[**AIApplicationChartUploadResponse**](AIApplicationChartUploadResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

