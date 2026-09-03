# \NodePoolsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateNodePool**](NodePoolsApi.md#CreateNodePool) | **Post** /v1/k8s/clusters/{clusterId}/node-pools | Create a Node Pool.
[**DeleteNodePool**](NodePoolsApi.md#DeleteNodePool) | **Delete** /v1/k8s/clusters/{clusterId}/node-pools/{id} | Delete a Node Pool by id.§
[**GetNodePools**](NodePoolsApi.md#GetNodePools) | **Get** /v1/k8s/clusters/{clusterId}/node-pools | Get the cluster&#39;s Node Pools.
[**GetNodepoolMetrics**](NodePoolsApi.md#GetNodepoolMetrics) | **Get** /api/v1/clusters/{clusterUuid}/nodepools/{nodepoolName}/metrics | Get the node pool metrics data. [Experimental]
[**SyncNodePools**](NodePoolsApi.md#SyncNodePools) | **Post** /v1/k8s/clusters/{clusterId}/node-pools/sync | Sync Node Pools.
[**UpdateNodePool**](NodePoolsApi.md#UpdateNodePool) | **Put** /v1/k8s/clusters/{clusterId}/node-pools/{id} | Update a Node Pool.
[**UpdateNodePoolLabels**](NodePoolsApi.md#UpdateNodePoolLabels) | **Put** /v1/k8s/clusters/{clusterId}/node-pools/{id}/labels | Update labels of a Node Pool.
[**UpdateNodePoolStatus**](NodePoolsApi.md#UpdateNodePoolStatus) | **Put** /v1/k8s/clusters/{clusterId}/node-pools/status | Update the status of a Node Pool.



## CreateNodePool

> CreateNodePool(ctx, clusterId).NodePoolCreateRequest(nodePoolCreateRequest).Execute()

Create a Node Pool.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    nodePoolCreateRequest := *openapiclient.NewNodePoolCreateRequest() // NodePoolCreateRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodePoolsApi.CreateNodePool(context.Background(), clusterId).NodePoolCreateRequest(nodePoolCreateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsApi.CreateNodePool``: %v\n", err)
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

Other parameters are passed through a pointer to a apiCreateNodePoolRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodePoolCreateRequest** | [**NodePoolCreateRequest**](NodePoolCreateRequest.md) |  | 

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


## DeleteNodePool

> DeleteNodePool(ctx, clusterId, id).Execute()

Delete a Node Pool by id.§



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    id := int32(56) // int32 | The unique id identifying the Node Pool.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodePoolsApi.DeleteNodePool(context.Background(), clusterId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsApi.DeleteNodePool``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 
**id** | **int32** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteNodePoolRequest struct via the builder pattern


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


## GetNodePools

> []NodePool1 GetNodePools(ctx, clusterId).Execute()

Get the cluster's Node Pools.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodePoolsApi.GetNodePools(context.Background(), clusterId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsApi.GetNodePools``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetNodePools`: []NodePool1
    fmt.Fprintf(os.Stdout, "Response from `NodePoolsApi.GetNodePools`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNodePoolsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]NodePool1**](NodePool1.md)

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

Get the node pool metrics data. [Experimental]



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
    nodepoolName := "default" // string | The unique nodepool name.
    start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
    end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
    metricType := []openapiclient.MetricsType{openapiclient.MetricsType("GPU_UTILIZATION")} // []MetricsType | specifies what data to request
    numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodePoolsApi.GetNodepoolMetrics(context.Background(), clusterUuid, nodepoolName).Start(start).End(end).MetricType(metricType).NumberOfSamples(numberOfSamples).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsApi.GetNodepoolMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetNodepoolMetrics`: MetricsResponse
    fmt.Fprintf(os.Stdout, "Response from `NodePoolsApi.GetNodepoolMetrics`: %v\n", resp)
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


## SyncNodePools

> SyncNodePools(ctx, clusterId).NodePoolStatusRequest(nodePoolStatusRequest).Execute()

Sync Node Pools.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    nodePoolStatusRequest := []openapiclient.NodePoolStatusRequest{*openapiclient.NewNodePoolStatusRequest()} // []NodePoolStatusRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodePoolsApi.SyncNodePools(context.Background(), clusterId).NodePoolStatusRequest(nodePoolStatusRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsApi.SyncNodePools``: %v\n", err)
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

Other parameters are passed through a pointer to a apiSyncNodePoolsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodePoolStatusRequest** | [**[]NodePoolStatusRequest**](NodePoolStatusRequest.md) |  | 

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


## UpdateNodePool

> UpdateNodePool(ctx, clusterId, id).UpdateNodePoolRequest(updateNodePoolRequest).Execute()

Update a Node Pool.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    id := int32(56) // int32 | The unique id identifying the Node Pool.
    updateNodePoolRequest := *openapiclient.NewUpdateNodePoolRequest() // UpdateNodePoolRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodePoolsApi.UpdateNodePool(context.Background(), clusterId, id).UpdateNodePoolRequest(updateNodePoolRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsApi.UpdateNodePool``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 
**id** | **int32** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNodePoolRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **updateNodePoolRequest** | [**UpdateNodePoolRequest**](UpdateNodePoolRequest.md) |  | 

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


## UpdateNodePoolLabels

> UpdateNodePoolLabels(ctx, clusterId, id).NodePoolLabelsRequest(nodePoolLabelsRequest).Execute()

Update labels of a Node Pool.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    id := int32(56) // int32 | The unique id identifying the Node Pool.
    nodePoolLabelsRequest := *openapiclient.NewNodePoolLabelsRequest() // NodePoolLabelsRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodePoolsApi.UpdateNodePoolLabels(context.Background(), clusterId, id).NodePoolLabelsRequest(nodePoolLabelsRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsApi.UpdateNodePoolLabels``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | Unique identifier of the cluster | 
**id** | **int32** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNodePoolLabelsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **nodePoolLabelsRequest** | [**NodePoolLabelsRequest**](NodePoolLabelsRequest.md) |  | 

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


## UpdateNodePoolStatus

> UpdateNodePoolStatus(ctx, clusterId).Name(name).NodePoolStatusRequest(nodePoolStatusRequest).Execute()

Update the status of a Node Pool.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    name := "name_example" // string | The unique name identifying the Node Pool.
    nodePoolStatusRequest := *openapiclient.NewNodePoolStatusRequest() // NodePoolStatusRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.NodePoolsApi.UpdateNodePoolStatus(context.Background(), clusterId).Name(name).NodePoolStatusRequest(nodePoolStatusRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `NodePoolsApi.UpdateNodePoolStatus``: %v\n", err)
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

Other parameters are passed through a pointer to a apiUpdateNodePoolStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **name** | **string** | The unique name identifying the Node Pool. | 
 **nodePoolStatusRequest** | [**NodePoolStatusRequest**](NodePoolStatusRequest.md) |  | 

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

