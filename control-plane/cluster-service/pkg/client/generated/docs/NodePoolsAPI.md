# \NodePoolsAPI

All URIs are relative to *https://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ClusterNotification**](NodePoolsAPI.md#ClusterNotification) | **Post** /internal/node-pools/notification | Cluster notification
[**CountNodepools**](NodePoolsAPI.md#CountNodepools) | **Get** /api/v1/node-pools/count | Count nodepools
[**CreateNodepool**](NodePoolsAPI.md#CreateNodepool) | **Post** /api/v1/node-pools | Create nodepool
[**DeleteNodepool**](NodePoolsAPI.md#DeleteNodepool) | **Delete** /api/v1/node-pools/{nodepoolId} | Delete nodepool
[**GetNodepool**](NodePoolsAPI.md#GetNodepool) | **Get** /api/v1/node-pools/{nodepoolId} | Get nodepool
[**GetNodepoolAncestors**](NodePoolsAPI.md#GetNodepoolAncestors) | **Get** /internal/node-pools/{nodepoolId}/ancestors | Get the nodepool ancestors.
[**GetNodepoolMetrics**](NodePoolsAPI.md#GetNodepoolMetrics) | **Get** /api/v1/clusters/{clusterUuid}/nodepools/{nodepoolName}/metrics | Get the node pool metrics data.
[**GetNodepoolMinimal**](NodePoolsAPI.md#GetNodepoolMinimal) | **Get** /api/v1/node-pools/{nodepoolId}/minimal | Get nodepool minimal data by ID
[**GetNodepools**](NodePoolsAPI.md#GetNodepools) | **Get** /api/v1/node-pools | Get nodepools
[**GetNodepoolsInternal**](NodePoolsAPI.md#GetNodepoolsInternal) | **Get** /internal/node-pools | Get nodepools internal
[**GetNodepoolsMinimal**](NodePoolsAPI.md#GetNodepoolsMinimal) | **Get** /api/v1/node-pools/minimal | Get nodepools minimal
[**GetNodepoolsMinimalInternal**](NodePoolsAPI.md#GetNodepoolsMinimalInternal) | **Get** /internal/node-pools/minimal | Get nodepools minimal internal
[**LegacyCreateNodepool**](NodePoolsAPI.md#LegacyCreateNodepool) | **Post** /api/v1/clusters/{clusterId}/node-pools | Create a Node Pool.
[**LegacyDeleteNodepool**](NodePoolsAPI.md#LegacyDeleteNodepool) | **Delete** /api/v1/clusters/{clusterId}/node-pools/{nodepoolId} | Delete a Node Pool by id.
[**LegacyGetNodepools**](NodePoolsAPI.md#LegacyGetNodepools) | **Get** /api/v1/clusters/{clusterId}/node-pools | Get the cluster&#39;s Node Pools.
[**LegacySyncNodepools**](NodePoolsAPI.md#LegacySyncNodepools) | **Post** /api/v1/clusters/{clusterId}/node-pools/sync | Sync Node Pools.
[**LegacyUpdateNodepool**](NodePoolsAPI.md#LegacyUpdateNodepool) | **Put** /api/v1/clusters/{clusterId}/node-pools/{nodepoolId} | Update a Node Pool.
[**LegacyUpdateNodepoolLabels**](NodePoolsAPI.md#LegacyUpdateNodepoolLabels) | **Put** /api/v1/clusters/{clusterId}/node-pools/{nodepoolId}/labels | Update labels of a Node Pool.
[**LegacyUpdateNodepoolStatus**](NodePoolsAPI.md#LegacyUpdateNodepoolStatus) | **Put** /api/v1/clusters/{clusterId}/node-pools/status | Update the status of a Node Pool.
[**PatchNodepool**](NodePoolsAPI.md#PatchNodepool) | **Patch** /api/v1/node-pools/{nodepoolId} | Patch nodepool fields
[**SyncNodepoolsStatus**](NodePoolsAPI.md#SyncNodepoolsStatus) | **Post** /api/v1/clusters/{clusterUuid}/nodepools/status | Sync nodepools status
[**UpdateNodepool**](NodePoolsAPI.md#UpdateNodepool) | **Put** /api/v1/node-pools/{nodepoolId} | Update nodepool



## ClusterNotification

> ClusterNotification(ctx).ClusterNotification(clusterNotification).Execute()

Cluster notification



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterNotification := *openapiclient.NewClusterNotification(openapiclient.NotificationOperation("Create"), "Id_example") // ClusterNotification | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodePoolsAPI.ClusterNotification(context.Background()).ClusterNotification(clusterNotification).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.ClusterNotification``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiClusterNotificationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterNotification** | [**ClusterNotification**](ClusterNotification.md) |  | 

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


## CountNodepools

> CountNodes200Response CountNodepools(ctx).FilterBy(filterBy).Execute()

Count nodepools



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.CountNodepools(context.Background()).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.CountNodepools``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountNodepools`: CountNodes200Response
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.CountNodepools`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountNodepoolsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 

### Return type

[**CountNodes200Response**](CountNodes200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateNodepool

> Nodepool CreateNodepool(ctx).NodepoolCreateFields(nodepoolCreateFields).Execute()

Create nodepool



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	nodepoolCreateFields := *openapiclient.NewNodepoolCreateFields("v100", "LabelKey_example", "LabelValue_example", "d73a738f-fab3-430a-8fa3-5241493d7128") // NodepoolCreateFields | Nodepool to create.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.CreateNodepool(context.Background()).NodepoolCreateFields(nodepoolCreateFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.CreateNodepool``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateNodepool`: Nodepool
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.CreateNodepool`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateNodepoolRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **nodepoolCreateFields** | [**NodepoolCreateFields**](NodepoolCreateFields.md) | Nodepool to create. | 

### Return type

[**Nodepool**](Nodepool.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteNodepool

> HttpResponse DeleteNodepool(ctx, nodepoolId).Execute()

Delete nodepool



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	nodepoolId := "17" // string | The unique id identifying the Node Pool.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.DeleteNodepool(context.Background(), nodepoolId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.DeleteNodepool``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteNodepool`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.DeleteNodepool`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodepoolId** | **string** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteNodepoolRequest struct via the builder pattern


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


## GetNodepool

> Nodepool GetNodepool(ctx, nodepoolId).Execute()

Get nodepool



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	nodepoolId := "17" // string | The unique id identifying the Node Pool.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.GetNodepool(context.Background(), nodepoolId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.GetNodepool``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodepool`: Nodepool
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.GetNodepool`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodepoolId** | **string** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNodepoolRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Nodepool**](Nodepool.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNodepoolAncestors

> Ancestors GetNodepoolAncestors(ctx, nodepoolId).Execute()

Get the nodepool ancestors.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	nodepoolId := "17" // string | The unique id identifying the Node Pool.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.GetNodepoolAncestors(context.Background(), nodepoolId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.GetNodepoolAncestors``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodepoolAncestors`: Ancestors
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.GetNodepoolAncestors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodepoolId** | **string** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNodepoolAncestorsRequest struct via the builder pattern


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


## GetNodepoolMetrics

> MetricsResponse GetNodepoolMetrics(ctx, clusterUuid, nodepoolName).Start(start).End(end).MetricType(metricType).NumberOfSamples(numberOfSamples).Execute()

Get the node pool metrics data.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	nodepoolName := "default" // string | The unique nodepool name.
	start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
	end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
	metricType := []openapiclient.MetricsType{openapiclient.MetricsType("TOTAL_GPU_NODES")} // []MetricsType | specifies what data to request
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.GetNodepoolMetrics(context.Background(), clusterUuid, nodepoolName).Start(start).End(end).MetricType(metricType).NumberOfSamples(numberOfSamples).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.GetNodepoolMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodepoolMetrics`: MetricsResponse
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.GetNodepoolMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**nodepoolName** | **string** | The unique nodepool name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNodepoolMetricsRequest struct via the builder pattern


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
- **Accept**: application/json, text/csv

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNodepoolMinimal

> NodepoolMinimal GetNodepoolMinimal(ctx, nodepoolId).Execute()

Get nodepool minimal data by ID



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	nodepoolId := "17" // string | The unique id identifying the Node Pool.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.GetNodepoolMinimal(context.Background(), nodepoolId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.GetNodepoolMinimal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodepoolMinimal`: NodepoolMinimal
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.GetNodepoolMinimal`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodepoolId** | **string** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNodepoolMinimalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**NodepoolMinimal**](NodepoolMinimal.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNodepools

> GetNodepools200Response GetNodepools(ctx).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Execute()

Get nodepools



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	sortBy := openapiclient.NodepoolSortFilterFields("name") // NodepoolSortFilterFields | Sort results by a parameters. (optional)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.GetNodepools(context.Background()).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.GetNodepools``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodepools`: GetNodepools200Response
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.GetNodepools`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNodepoolsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **sortBy** | [**NodepoolSortFilterFields**](NodepoolSortFilterFields.md) | Sort results by a parameters. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**GetNodepools200Response**](GetNodepools200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNodepoolsInternal

> GetNodepoolsInternal200Response GetNodepoolsInternal(ctx).ClusterId(clusterId).Execute()

Get nodepools internal



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.GetNodepoolsInternal(context.Background()).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.GetNodepoolsInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodepoolsInternal`: GetNodepoolsInternal200Response
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.GetNodepoolsInternal`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNodepoolsInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 

### Return type

[**GetNodepoolsInternal200Response**](GetNodepoolsInternal200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNodepoolsMinimal

> GetNodepoolsMinimal200Response GetNodepoolsMinimal(ctx).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Execute()

Get nodepools minimal



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	sortBy := openapiclient.NodepoolMinimalSortFields("clusterId") // NodepoolMinimalSortFields | Sort results by a parameter. (optional)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.GetNodepoolsMinimal(context.Background()).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.GetNodepoolsMinimal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodepoolsMinimal`: GetNodepoolsMinimal200Response
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.GetNodepoolsMinimal`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNodepoolsMinimalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **sortBy** | [**NodepoolMinimalSortFields**](NodepoolMinimalSortFields.md) | Sort results by a parameter. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**GetNodepoolsMinimal200Response**](GetNodepoolsMinimal200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNodepoolsMinimalInternal

> GetNodepoolsMinimalInternal200Response GetNodepoolsMinimalInternal(ctx).TenantId(tenantId).Execute()

Get nodepools minimal internal



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	tenantId := []string{"Inner_example"} // []string | The tenant Ids of the current request. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.GetNodepoolsMinimalInternal(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.GetNodepoolsMinimalInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodepoolsMinimalInternal`: GetNodepoolsMinimalInternal200Response
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.GetNodepoolsMinimalInternal`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNodepoolsMinimalInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **[]string** | The tenant Ids of the current request. | 

### Return type

[**GetNodepoolsMinimalInternal200Response**](GetNodepoolsMinimalInternal200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyCreateNodepool

> LegacyCreateNodepool(ctx, clusterId).LegacyNodePoolCreateRequest(legacyNodePoolCreateRequest).Execute()

Create a Node Pool.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
	legacyNodePoolCreateRequest := *openapiclient.NewLegacyNodePoolCreateRequest() // LegacyNodePoolCreateRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodePoolsAPI.LegacyCreateNodepool(context.Background(), clusterId).LegacyNodePoolCreateRequest(legacyNodePoolCreateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.LegacyCreateNodepool``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyCreateNodepoolRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **legacyNodePoolCreateRequest** | [**LegacyNodePoolCreateRequest**](LegacyNodePoolCreateRequest.md) |  | 

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


## LegacyDeleteNodepool

> LegacyDeleteNodepool(ctx, clusterId, nodepoolId).Execute()

Delete a Node Pool by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
	nodepoolId := "17" // string | The unique id identifying the Node Pool.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodePoolsAPI.LegacyDeleteNodepool(context.Background(), clusterId, nodepoolId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.LegacyDeleteNodepool``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 
**nodepoolId** | **string** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyDeleteNodepoolRequest struct via the builder pattern


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


## LegacyGetNodepools

> []LegacyNodePool LegacyGetNodepools(ctx, clusterId).Execute()

Get the cluster's Node Pools.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.LegacyGetNodepools(context.Background(), clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.LegacyGetNodepools``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `LegacyGetNodepools`: []LegacyNodePool
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.LegacyGetNodepools`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyGetNodepoolsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]LegacyNodePool**](LegacyNodePool.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacySyncNodepools

> LegacySyncNodepools(ctx, clusterId).LegacyNodePoolStatusRequest(legacyNodePoolStatusRequest).Execute()

Sync Node Pools.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
	legacyNodePoolStatusRequest := []openapiclient.LegacyNodePoolStatusRequest{*openapiclient.NewLegacyNodePoolStatusRequest()} // []LegacyNodePoolStatusRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodePoolsAPI.LegacySyncNodepools(context.Background(), clusterId).LegacyNodePoolStatusRequest(legacyNodePoolStatusRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.LegacySyncNodepools``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacySyncNodepoolsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **legacyNodePoolStatusRequest** | [**[]LegacyNodePoolStatusRequest**](LegacyNodePoolStatusRequest.md) |  | 

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


## LegacyUpdateNodepool

> LegacyUpdateNodepool(ctx, clusterId, nodepoolId).LegacyUpdateNodePoolRequest(legacyUpdateNodePoolRequest).Execute()

Update a Node Pool.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
	nodepoolId := "17" // string | The unique id identifying the Node Pool.
	legacyUpdateNodePoolRequest := *openapiclient.NewLegacyUpdateNodePoolRequest() // LegacyUpdateNodePoolRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodePoolsAPI.LegacyUpdateNodepool(context.Background(), clusterId, nodepoolId).LegacyUpdateNodePoolRequest(legacyUpdateNodePoolRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.LegacyUpdateNodepool``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 
**nodepoolId** | **string** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyUpdateNodepoolRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **legacyUpdateNodePoolRequest** | [**LegacyUpdateNodePoolRequest**](LegacyUpdateNodePoolRequest.md) |  | 

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


## LegacyUpdateNodepoolLabels

> LegacyUpdateNodepoolLabels(ctx, clusterId, nodepoolId).LegacyNodePoolLabelsRequest(legacyNodePoolLabelsRequest).Execute()

Update labels of a Node Pool.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
	nodepoolId := "17" // string | The unique id identifying the Node Pool.
	legacyNodePoolLabelsRequest := *openapiclient.NewLegacyNodePoolLabelsRequest() // LegacyNodePoolLabelsRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodePoolsAPI.LegacyUpdateNodepoolLabels(context.Background(), clusterId, nodepoolId).LegacyNodePoolLabelsRequest(legacyNodePoolLabelsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.LegacyUpdateNodepoolLabels``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 
**nodepoolId** | **string** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyUpdateNodepoolLabelsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **legacyNodePoolLabelsRequest** | [**LegacyNodePoolLabelsRequest**](LegacyNodePoolLabelsRequest.md) |  | 

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


## LegacyUpdateNodepoolStatus

> LegacyUpdateNodepoolStatus(ctx, clusterId).Name(name).LegacyNodePoolStatusRequest(legacyNodePoolStatusRequest).Execute()

Update the status of a Node Pool.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
	name := "name_example" // string | The unique name identifying the Node Pool.
	legacyNodePoolStatusRequest := *openapiclient.NewLegacyNodePoolStatusRequest() // LegacyNodePoolStatusRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodePoolsAPI.LegacyUpdateNodepoolStatus(context.Background(), clusterId).Name(name).LegacyNodePoolStatusRequest(legacyNodePoolStatusRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.LegacyUpdateNodepoolStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyUpdateNodepoolStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **name** | **string** | The unique name identifying the Node Pool. | 
 **legacyNodePoolStatusRequest** | [**LegacyNodePoolStatusRequest**](LegacyNodePoolStatusRequest.md) |  | 

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


## PatchNodepool

> Nodepool PatchNodepool(ctx, nodepoolId).NodepoolUpdateFields(nodepoolUpdateFields).Execute()

Patch nodepool fields



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	nodepoolId := "17" // string | The unique id identifying the Node Pool.
	nodepoolUpdateFields := *openapiclient.NewNodepoolUpdateFields() // NodepoolUpdateFields | Fields to update.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.PatchNodepool(context.Background(), nodepoolId).NodepoolUpdateFields(nodepoolUpdateFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.PatchNodepool``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PatchNodepool`: Nodepool
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.PatchNodepool`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodepoolId** | **string** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchNodepoolRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodepoolUpdateFields** | [**NodepoolUpdateFields**](NodepoolUpdateFields.md) | Fields to update. | 

### Return type

[**Nodepool**](Nodepool.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncNodepoolsStatus

> SyncNodepoolsStatus(ctx, clusterUuid).SyncNodepoolsStatusRequest(syncNodepoolsStatusRequest).Execute()

Sync nodepools status



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	syncNodepoolsStatusRequest := *openapiclient.NewSyncNodepoolsStatusRequest([]openapiclient.NodepoolSyncFields{*openapiclient.NewNodepoolSyncFields("v100")}) // SyncNodepoolsStatusRequest | Nodepools sync data.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodePoolsAPI.SyncNodepoolsStatus(context.Background(), clusterUuid).SyncNodepoolsStatusRequest(syncNodepoolsStatusRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.SyncNodepoolsStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncNodepoolsStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **syncNodepoolsStatusRequest** | [**SyncNodepoolsStatusRequest**](SyncNodepoolsStatusRequest.md) | Nodepools sync data. | 

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


## UpdateNodepool

> Nodepool UpdateNodepool(ctx, nodepoolId).NodepoolUpdateFields(nodepoolUpdateFields).Execute()

Update nodepool



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	nodepoolId := "17" // string | The unique id identifying the Node Pool.
	nodepoolUpdateFields := *openapiclient.NewNodepoolUpdateFields() // NodepoolUpdateFields | Nodepool to update.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodePoolsAPI.UpdateNodepool(context.Background(), nodepoolId).NodepoolUpdateFields(nodepoolUpdateFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsAPI.UpdateNodepool``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateNodepool`: Nodepool
	fmt.Fprintf(os.Stdout, "Response from `NodePoolsAPI.UpdateNodepool`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodepoolId** | **string** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNodepoolRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodepoolUpdateFields** | [**NodepoolUpdateFields**](NodepoolUpdateFields.md) | Nodepool to update. | 

### Return type

[**Nodepool**](Nodepool.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

