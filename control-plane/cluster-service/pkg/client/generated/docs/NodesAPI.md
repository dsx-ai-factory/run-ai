# \NodesAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CountNodes**](NodesAPI.md#CountNodes) | **Get** /api/v1/clusters/{clusterUuid}/nodes/count | Count nodes (Fern preview demo)
[**CountTenantNodes**](NodesAPI.md#CountTenantNodes) | **Get** /api/v1/nodes/count | Count tenant nodes
[**DeleteNode**](NodesAPI.md#DeleteNode) | **Delete** /api/v1/clusters/{clusterUuid}/nodes/{nodeName} | Delete a node from a cluster.
[**GetNode**](NodesAPI.md#GetNode) | **Get** /api/v1/nodes/{nodeId} | Get a node by ID.
[**GetNodeInternal**](NodesAPI.md#GetNodeInternal) | **Get** /internal/clusters/{clusterUuid}/nodes/{nodeName} | Get node internal data by cluster UUID and name
[**GetNodeMetrics**](NodesAPI.md#GetNodeMetrics) | **Get** /api/v1/nodes/{nodeId}/metrics | Get the node metrics data.
[**GetNodeTelemetry**](NodesAPI.md#GetNodeTelemetry) | **Get** /api/v1/nodes/telemetry | Get node telemetry data.
[**GetTenantNodes**](NodesAPI.md#GetTenantNodes) | **Get** /api/v1/nodes | Get a list of nodes.
[**SyncNodes**](NodesAPI.md#SyncNodes) | **Post** /api/v1/clusters/{clusterUuid}/nodes/sync | Sync the nodes in the cluster.
[**SyncNumaNodes**](NodesAPI.md#SyncNumaNodes) | **Post** /api/v1/clusters/{clusterUuid}/nodes/numa/sync | Sync NUMA nodes in a cluster.
[**UpsertNode**](NodesAPI.md#UpsertNode) | **Put** /api/v1/clusters/{clusterUuid}/nodes/{nodeName} | Update node details.



## CountNodes

> CountNodes200Response CountNodes(ctx, clusterUuid).FilterBy(filterBy).Search(search).Execute()

Count nodes (Fern preview demo)



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodesAPI.CountNodes(context.Background(), clusterUuid).FilterBy(filterBy).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesAPI.CountNodes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountNodes`: CountNodes200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesAPI.CountNodes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCountNodesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **search** | **string** | Filter results by a free text search. | 

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


## CountTenantNodes

> CountNodes200Response CountTenantNodes(ctx).FilterBy(filterBy).Search(search).Execute()

Count tenant nodes



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodesAPI.CountTenantNodes(context.Background()).FilterBy(filterBy).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesAPI.CountTenantNodes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountTenantNodes`: CountNodes200Response
	fmt.Fprintf(os.Stdout, "Response from `NodesAPI.CountTenantNodes`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountTenantNodesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **search** | **string** | Filter results by a free text search. | 

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


## DeleteNode

> DeleteNode(ctx, clusterUuid, nodeName).Execute()

Delete a node from a cluster.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	nodeName := "nodeName_example" // string | The node name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodesAPI.DeleteNode(context.Background(), clusterUuid, nodeName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesAPI.DeleteNode``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**nodeName** | **string** | The node name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteNodeRequest struct via the builder pattern


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


## GetNode

> NodeExtended GetNode(ctx, nodeId).Execute()

Get a node by ID.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	nodeId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The node UUID.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodesAPI.GetNode(context.Background(), nodeId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesAPI.GetNode``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNode`: NodeExtended
	fmt.Fprintf(os.Stdout, "Response from `NodesAPI.GetNode`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodeId** | **string** | The node UUID. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNodeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**NodeExtended**](NodeExtended.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNodeInternal

> NodeInternal GetNodeInternal(ctx, clusterUuid, nodeName).Execute()

Get node internal data by cluster UUID and name



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	nodeName := "nodeName_example" // string | The node name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodesAPI.GetNodeInternal(context.Background(), clusterUuid, nodeName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesAPI.GetNodeInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodeInternal`: NodeInternal
	fmt.Fprintf(os.Stdout, "Response from `NodesAPI.GetNodeInternal`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**nodeName** | **string** | The node name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNodeInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**NodeInternal**](NodeInternal.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNodeMetrics

> MetricsResponse GetNodeMetrics(ctx, nodeId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()

Get the node metrics data.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	nodeId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The node UUID.
	metricType := []openapiclient.NodeMetricType{openapiclient.NodeMetricType("GPU_UTILIZATION_PER_GPU")} // []NodeMetricType | Specify which data to request.
	start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
	end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodesAPI.GetNodeMetrics(context.Background(), nodeId).MetricType(metricType).Start(start).End(end).NumberOfSamples(numberOfSamples).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesAPI.GetNodeMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodeMetrics`: MetricsResponse
	fmt.Fprintf(os.Stdout, "Response from `NodesAPI.GetNodeMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodeId** | **string** | The node UUID. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNodeMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **metricType** | [**[]NodeMetricType**](NodeMetricType.md) | Specify which data to request. | 
 **start** | **time.Time** | Start date of time range to fetch data in ISO 8601 timestamp format. | 
 **end** | **time.Time** | End date of time range to fetch data in ISO 8601 timestamp format. | 
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


## GetNodeTelemetry

> TelemetryResponse GetNodeTelemetry(ctx).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolName(nodepoolName).GroupBy(groupBy).TenantId(tenantId).Execute()

Get node telemetry data.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	telemetryType := openapiclient.NodeTelemetryType("READY_GPU_NODES") // NodeTelemetryType | specifies what data to request
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	nodepoolName := "default" // string | Filter using the nodepool. (optional)
	groupBy := []string{"GroupBy_example"} // []string | workload fields to group the data by (optional)
	tenantId := []string{"Inner_example"} // []string | The tenant Ids of the current request. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodesAPI.GetNodeTelemetry(context.Background()).TelemetryType(telemetryType).ClusterId(clusterId).NodepoolName(nodepoolName).GroupBy(groupBy).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesAPI.GetNodeTelemetry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodeTelemetry`: TelemetryResponse
	fmt.Fprintf(os.Stdout, "Response from `NodesAPI.GetNodeTelemetry`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNodeTelemetryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **telemetryType** | [**NodeTelemetryType**](NodeTelemetryType.md) | specifies what data to request | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **nodepoolName** | **string** | Filter using the nodepool. | 
 **groupBy** | **[]string** | workload fields to group the data by | 
 **tenantId** | **[]string** | The tenant Ids of the current request. | 

### Return type

[**TelemetryResponse**](TelemetryResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, text/csv

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenantNodes

> Nodes GetTenantNodes(ctx).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Search(search).Execute()

Get a list of nodes.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	sortBy := openapiclient.NodeFilterSortFields("name") // NodeFilterSortFields | Sort results by a parameters. (optional)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodesAPI.GetTenantNodes(context.Background()).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesAPI.GetTenantNodes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTenantNodes`: Nodes
	fmt.Fprintf(os.Stdout, "Response from `NodesAPI.GetTenantNodes`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantNodesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **sortBy** | [**NodeFilterSortFields**](NodeFilterSortFields.md) | Sort results by a parameters. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **search** | **string** | Filter results by a free text search. | 

### Return type

[**Nodes**](Nodes.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncNodes

> SyncNodes(ctx, clusterUuid).NodesSyncRequest(nodesSyncRequest).Execute()

Sync the nodes in the cluster.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	nodesSyncRequest := *openapiclient.NewNodesSyncRequest([]openapiclient.NodeForSync{*openapiclient.NewNodeForSync("worker-node-1", *openapiclient.NewNodeInfo("Status_example", "node-pool-1", time.Now()))}) // NodesSyncRequest | The nodes info for sync.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodesAPI.SyncNodes(context.Background(), clusterUuid).NodesSyncRequest(nodesSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesAPI.SyncNodes``: %v\n", err)
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

Other parameters are passed through a pointer to a apiSyncNodesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodesSyncRequest** | [**NodesSyncRequest**](NodesSyncRequest.md) | The nodes info for sync. | 

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


## SyncNumaNodes

> SyncNumaNodes(ctx, clusterUuid).NumaNodesSyncRequest(numaNodesSyncRequest).Execute()

Sync NUMA nodes in a cluster.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	numaNodesSyncRequest := *openapiclient.NewNumaNodesSyncRequest([]openapiclient.NodeNumaForSync{*openapiclient.NewNodeNumaForSync("worker-node-1", []openapiclient.NumaNodeDetails{*openapiclient.NewNumaNodeDetails("node-0", float64(63.5), float64(8.5), int64(270582939648), int64(34359738368), int32(4), int32(2))})}) // NumaNodesSyncRequest | NUMA-node snapshots to sync.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodesAPI.SyncNumaNodes(context.Background(), clusterUuid).NumaNodesSyncRequest(numaNodesSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesAPI.SyncNumaNodes``: %v\n", err)
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

Other parameters are passed through a pointer to a apiSyncNumaNodesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **numaNodesSyncRequest** | [**NumaNodesSyncRequest**](NumaNodesSyncRequest.md) | NUMA-node snapshots to sync. | 

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


## UpsertNode

> UpsertNode(ctx, clusterUuid, nodeName).NodeInfo(nodeInfo).Execute()

Update node details.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	nodeName := "nodeName_example" // string | The node name.
	nodeInfo := *openapiclient.NewNodeInfo("Status_example", "node-pool-1", time.Now()) // NodeInfo | The node info in the cluster.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodesAPI.UpsertNode(context.Background(), clusterUuid, nodeName).NodeInfo(nodeInfo).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesAPI.UpsertNode``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**nodeName** | **string** | The node name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpsertNodeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **nodeInfo** | [**NodeInfo**](NodeInfo.md) | The node info in the cluster. | 

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

