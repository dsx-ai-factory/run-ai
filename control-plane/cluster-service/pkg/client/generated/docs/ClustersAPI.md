# \ClustersAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateCluster**](ClustersAPI.md#CreateCluster) | **Post** /api/v1/clusters | Create a cluster.
[**DeleteCluster**](ClustersAPI.md#DeleteCluster) | **Delete** /api/v1/clusters/{clusterUuid} | Delete a cluster by id.
[**GetClusterAncestors**](ClustersAPI.md#GetClusterAncestors) | **Get** /internal/clusters/{clusterUuid}/ancestors | Get the cluster ancestors.
[**GetClusterByUuid**](ClustersAPI.md#GetClusterByUuid) | **Get** /api/v1/clusters/{clusterUuid} | Get cluster by id.
[**GetClusterInstallInfoByUuid**](ClustersAPI.md#GetClusterInstallInfoByUuid) | **Get** /api/v1/clusters/{clusterUuid}/cluster-install-info | Retrieve the installation instructions of a cluster by ID.
[**GetClusterInternal**](ClustersAPI.md#GetClusterInternal) | **Get** /internal/clusters/{clusterUuid} | Get cluster by id internal.
[**GetClusterMetrics**](ClustersAPI.md#GetClusterMetrics) | **Get** /api/v1/clusters/{clusterUuid}/metrics | Get the cluster metrics data.
[**GetClusterMinimalDataById**](ClustersAPI.md#GetClusterMinimalDataById) | **Get** /api/v1/clusters/{clusterUuid}/minimal | Get cluster minimal data by id
[**GetClusterName**](ClustersAPI.md#GetClusterName) | **Get** /internal/clusters/{clusterUuid}/name | Get the cluster name.
[**GetClusterUrl**](ClustersAPI.md#GetClusterUrl) | **Get** /internal/clusters/{clusterUuid}/url | Get the cluster url.
[**GetClusterVersions**](ClustersAPI.md#GetClusterVersions) | **Get** /api/v1/clusters/versions | Get cluster versions.
[**GetClusters**](ClustersAPI.md#GetClusters) | **Get** /api/v1/clusters | Get a list of clusters (Fern preview demo)
[**GetClustersCount**](ClustersAPI.md#GetClustersCount) | **Get** /api/v1/clusters/count | Count the amount of clusters in a tenant.
[**GetClustersMinimalData**](ClustersAPI.md#GetClustersMinimalData) | **Get** /api/v1/clusters/minimal | Get clusters minimal data
[**GetInternalClusters**](ClustersAPI.md#GetInternalClusters) | **Get** /internal/clusters | Get an internal list of clusters.
[**UpdateCluster**](ClustersAPI.md#UpdateCluster) | **Put** /api/v1/clusters/{clusterUuid} | Update a cluster by id.
[**UpdateClusterInfo**](ClustersAPI.md#UpdateClusterInfo) | **Post** /api/v1/clusters/{clusterUuid}/cluster-info | Update cluster domain, version and status for a cluster.
[**UpdateClusterLiveness**](ClustersAPI.md#UpdateClusterLiveness) | **Post** /api/v1/clusters/{clusterUuid}/liveness | Update liveness status for a cluster.



## CreateCluster

> DisplayedCluster CreateCluster(ctx).ClusterCreationRequest(clusterCreationRequest).Execute()

Create a cluster.



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
	clusterCreationRequest := *openapiclient.NewClusterCreationRequest("Name_example") // ClusterCreationRequest | The cluster to create.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.CreateCluster(context.Background()).ClusterCreationRequest(clusterCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.CreateCluster``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateCluster`: DisplayedCluster
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.CreateCluster`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateClusterRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterCreationRequest** | [**ClusterCreationRequest**](ClusterCreationRequest.md) | The cluster to create. | 

### Return type

[**DisplayedCluster**](DisplayedCluster.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteCluster

> HttpResponse DeleteCluster(ctx, clusterUuid).Force(force).Execute()

Delete a cluster by id.



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
	force := true // bool | if true will force cluster instant deletion otherwise will start cluster graceful deletion process.  (optional) (default to false)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.DeleteCluster(context.Background(), clusterUuid).Force(force).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.DeleteCluster``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteCluster`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.DeleteCluster`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteClusterRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **force** | **bool** | if true will force cluster instant deletion otherwise will start cluster graceful deletion process.  | [default to false]

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


## GetClusterAncestors

> Ancestors GetClusterAncestors(ctx, clusterUuid).Execute()

Get the cluster ancestors.

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetClusterAncestors(context.Background(), clusterUuid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetClusterAncestors``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterAncestors`: Ancestors
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetClusterAncestors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterAncestorsRequest struct via the builder pattern


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


## GetClusterByUuid

> DisplayedCluster GetClusterByUuid(ctx, clusterUuid).Verbosity(verbosity).Execute()

Get cluster by id.



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
	verbosity := "full" // string | response verbosity level.  (optional) (default to "full")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetClusterByUuid(context.Background(), clusterUuid).Verbosity(verbosity).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetClusterByUuid``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterByUuid`: DisplayedCluster
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetClusterByUuid`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterByUuidRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **verbosity** | **string** | response verbosity level.  | [default to &quot;full&quot;]

### Return type

[**DisplayedCluster**](DisplayedCluster.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClusterInstallInfoByUuid

> ClusterInstallationInfoResponse GetClusterInstallInfoByUuid(ctx, clusterUuid).Version(version).RemoteClusterUrl(remoteClusterUrl).Execute()

Retrieve the installation instructions of a cluster by ID.



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
	version := "2.16" // string | The cluster version to install
	remoteClusterUrl := "https://cluster.runai" // string | The remote URL of the runai cluster (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetClusterInstallInfoByUuid(context.Background(), clusterUuid).Version(version).RemoteClusterUrl(remoteClusterUrl).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetClusterInstallInfoByUuid``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterInstallInfoByUuid`: ClusterInstallationInfoResponse
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetClusterInstallInfoByUuid`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterInstallInfoByUuidRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **version** | **string** | The cluster version to install | 
 **remoteClusterUrl** | **string** | The remote URL of the runai cluster | 

### Return type

[**ClusterInstallationInfoResponse**](ClusterInstallationInfoResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClusterInternal

> DisplayedCluster GetClusterInternal(ctx, clusterUuid).Execute()

Get cluster by id internal.



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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetClusterInternal(context.Background(), clusterUuid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetClusterInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterInternal`: DisplayedCluster
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetClusterInternal`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DisplayedCluster**](DisplayedCluster.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClusterMetrics

> GroupedMetricsResponse GetClusterMetrics(ctx, clusterUuid).Start(start).End(end).MetricType(metricType).NumberOfSamples(numberOfSamples).GroupBy(groupBy).NodepoolName(nodepoolName).Execute()

Get the cluster metrics data.



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
	start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
	end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
	metricType := []openapiclient.MetricsType{openapiclient.MetricsType("TOTAL_GPU_NODES")} // []MetricsType | specifies what data to request
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)
	groupBy := "Category" // string | Labels to group the returned metrics data by. Grouping availability depends on the selected metric type. Supported values: \"category\" (workload category) or \"node pool\".  (optional)
	nodepoolName := "default" // string | Filter using the nodepool. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetClusterMetrics(context.Background(), clusterUuid).Start(start).End(end).MetricType(metricType).NumberOfSamples(numberOfSamples).GroupBy(groupBy).NodepoolName(nodepoolName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetClusterMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterMetrics`: GroupedMetricsResponse
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetClusterMetrics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **start** | **time.Time** | Start date of time range to fetch data in ISO 8601 timestamp format. | 
 **end** | **time.Time** | End date of time range to fetch data in ISO 8601 timestamp format. | 
 **metricType** | [**[]MetricsType**](MetricsType.md) | specifies what data to request | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]
 **groupBy** | **string** | Labels to group the returned metrics data by. Grouping availability depends on the selected metric type. Supported values: \&quot;category\&quot; (workload category) or \&quot;node pool\&quot;.  | 
 **nodepoolName** | **string** | Filter using the nodepool. | 

### Return type

[**GroupedMetricsResponse**](GroupedMetricsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, text/csv

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClusterMinimalDataById

> ClustersMinimalData GetClusterMinimalDataById(ctx, clusterUuid).Execute()

Get cluster minimal data by id



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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetClusterMinimalDataById(context.Background(), clusterUuid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetClusterMinimalDataById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterMinimalDataById`: ClustersMinimalData
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetClusterMinimalDataById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterMinimalDataByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ClustersMinimalData**](ClustersMinimalData.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClusterName

> GetClusterName200Response GetClusterName(ctx, clusterUuid).Execute()

Get the cluster name.



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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetClusterName(context.Background(), clusterUuid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetClusterName``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterName`: GetClusterName200Response
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetClusterName`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterNameRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GetClusterName200Response**](GetClusterName200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClusterUrl

> GetClusterUrl200Response GetClusterUrl(ctx, clusterUuid).Execute()

Get the cluster url.



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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetClusterUrl(context.Background(), clusterUuid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetClusterUrl``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterUrl`: GetClusterUrl200Response
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetClusterUrl`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterUrlRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GetClusterUrl200Response**](GetClusterUrl200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClusterVersions

> map[string]string GetClusterVersions(ctx).Execute()

Get cluster versions.



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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetClusterVersions(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetClusterVersions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterVersions`: map[string]string
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetClusterVersions`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterVersionsRequest struct via the builder pattern


### Return type

**map[string]string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClusters

> []DisplayedCluster GetClusters(ctx).Verbosity(verbosity).TenantId(tenantId).IncludeRequestedForDelete(includeRequestedForDelete).Execute()

Get a list of clusters (Fern preview demo)



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
	verbosity := "full" // string | response verbosity level.  (optional) (default to "full")
	tenantId := []string{"Inner_example"} // []string | The tenant Ids of the current request. (optional)
	includeRequestedForDelete := true // bool | When true, includes clusters that are marked for deletion (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetClusters(context.Background()).Verbosity(verbosity).TenantId(tenantId).IncludeRequestedForDelete(includeRequestedForDelete).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetClusters``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusters`: []DisplayedCluster
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetClusters`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetClustersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **verbosity** | **string** | response verbosity level.  | [default to &quot;full&quot;]
 **tenantId** | **[]string** | The tenant Ids of the current request. | 
 **includeRequestedForDelete** | **bool** | When true, includes clusters that are marked for deletion | 

### Return type

[**[]DisplayedCluster**](DisplayedCluster.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClustersCount

> []ClustersCount GetClustersCount(ctx).GroupBy(groupBy).Execute()

Count the amount of clusters in a tenant.



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
	groupBy := "subdomain" // string | attribute to group cluster by.  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetClustersCount(context.Background()).GroupBy(groupBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetClustersCount``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClustersCount`: []ClustersCount
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetClustersCount`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetClustersCountRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **groupBy** | **string** | attribute to group cluster by.  | 

### Return type

[**[]ClustersCount**](ClustersCount.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClustersMinimalData

> GetClustersMinimalData200Response GetClustersMinimalData(ctx).FilterBy(filterBy).Offset(offset).Limit(limit).Execute()

Get clusters minimal data



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
	filterBy := []string{"Inner_example"} // []string | Filter clusters by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetClustersMinimalData(context.Background()).FilterBy(filterBy).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetClustersMinimalData``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClustersMinimalData`: GetClustersMinimalData200Response
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetClustersMinimalData`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetClustersMinimalDataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter clusters by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**GetClustersMinimalData200Response**](GetClustersMinimalData200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInternalClusters

> []DisplayedCluster GetInternalClusters(ctx).IncludeDeleted(includeDeleted).Execute()

Get an internal list of clusters.



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
	includeDeleted := true // bool |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClustersAPI.GetInternalClusters(context.Background()).IncludeDeleted(includeDeleted).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.GetInternalClusters``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInternalClusters`: []DisplayedCluster
	fmt.Fprintf(os.Stdout, "Response from `ClustersAPI.GetInternalClusters`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetInternalClustersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **includeDeleted** | **bool** |  | 

### Return type

[**[]DisplayedCluster**](DisplayedCluster.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateCluster

> UpdateCluster(ctx, clusterUuid).ClusterUpdateRequest(clusterUpdateRequest).Execute()

Update a cluster by id.



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
	clusterUpdateRequest := *openapiclient.NewClusterUpdateRequest("Name_example") // ClusterUpdateRequest | The cluster details to update

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ClustersAPI.UpdateCluster(context.Background(), clusterUuid).ClusterUpdateRequest(clusterUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.UpdateCluster``: %v\n", err)
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

Other parameters are passed through a pointer to a apiUpdateClusterRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **clusterUpdateRequest** | [**ClusterUpdateRequest**](ClusterUpdateRequest.md) | The cluster details to update | 

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


## UpdateClusterInfo

> UpdateClusterInfo(ctx, clusterUuid).ClusterInfoSyncRequest(clusterInfoSyncRequest).Execute()

Update cluster domain, version and status for a cluster.



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
	clusterInfoSyncRequest := *openapiclient.NewClusterInfoSyncRequest("Domain_example", "Version_example") // ClusterInfoSyncRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ClustersAPI.UpdateClusterInfo(context.Background(), clusterUuid).ClusterInfoSyncRequest(clusterInfoSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.UpdateClusterInfo``: %v\n", err)
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

Other parameters are passed through a pointer to a apiUpdateClusterInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **clusterInfoSyncRequest** | [**ClusterInfoSyncRequest**](ClusterInfoSyncRequest.md) |  | 

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


## UpdateClusterLiveness

> UpdateClusterLiveness(ctx, clusterUuid).ClusterLivenessSyncRequest(clusterLivenessSyncRequest).Execute()

Update liveness status for a cluster.



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
	clusterLivenessSyncRequest := *openapiclient.NewClusterLivenessSyncRequest(int64(123)) // ClusterLivenessSyncRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ClustersAPI.UpdateClusterLiveness(context.Background(), clusterUuid).ClusterLivenessSyncRequest(clusterLivenessSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClustersAPI.UpdateClusterLiveness``: %v\n", err)
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

Other parameters are passed through a pointer to a apiUpdateClusterLivenessRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **clusterLivenessSyncRequest** | [**ClusterLivenessSyncRequest**](ClusterLivenessSyncRequest.md) |  | 

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

