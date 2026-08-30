# \ClustersApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateCluster**](ClustersApi.md#CreateCluster) | **Post** /api/v1/clusters | Create a cluster.
[**CreateCluster_0**](ClustersApi.md#CreateCluster_0) | **Post** /v1/k8s/clusters | Create a cluster
[**DeleteCluster**](ClustersApi.md#DeleteCluster) | **Delete** /api/v1/clusters/{clusterUuid} | Delete a cluster.
[**GetClusterAncestors**](ClustersApi.md#GetClusterAncestors) | **Get** /internal/clusters/{clusterUuid}/ancestors | Get the cluster ancestors.
[**GetClusterByUuid**](ClustersApi.md#GetClusterByUuid) | **Get** /api/v1/clusters/{clusterUuid} | Get cluster by id.
[**GetClusterInstallInfoByUuid**](ClustersApi.md#GetClusterInstallInfoByUuid) | **Get** /api/v1/clusters/{clusterUuid}/cluster-install-info | Retrieve the installation instructions of a cluster by ID.
[**GetClusterMetrics**](ClustersApi.md#GetClusterMetrics) | **Get** /api/v1/clusters/{clusterUuid}/metrics | Get the cluster metrics data.
[**GetClusterMetrics_0**](ClustersApi.md#GetClusterMetrics_0) | **Get** /v1/k8s/clusters/{clusterUuid}/metrics | Get cluster metrics.
[**GetClusterName**](ClustersApi.md#GetClusterName) | **Get** /internal/clusters/{clusterUuid}/name | Get the cluster name.
[**GetClusterVersions**](ClustersApi.md#GetClusterVersions) | **Get** /api/v1/clusters/versions | Get cluster versions.
[**GetClusters**](ClustersApi.md#GetClusters) | **Get** /api/v1/clusters | Get a list of clusters.
[**GetInstallFile**](ClustersApi.md#GetInstallFile) | **Get** /v1/k8s/clusters/{cluster_uuid}/installfile | Get cluster installation file by id.
[**InternalClustersUuidSubResourcesDelete**](ClustersApi.md#InternalClustersUuidSubResourcesDelete) | **Delete** /internal/clusters/{uuid}/sub-resources | Delete all cluster sub-resources
[**InternalClustersUuidSubResourcesPost**](ClustersApi.md#InternalClustersUuidSubResourcesPost) | **Post** /internal/clusters/{uuid}/sub-resources | Create all cluster sub-resources
[**UpdateCluster**](ClustersApi.md#UpdateCluster) | **Put** /api/v1/clusters/{clusterUuid} | Update a cluster by id.
[**UpdateCluster1**](ClustersApi.md#UpdateCluster1) | **Put** /v1/k8s/clusters/{uuid} | Update a Cluster by Id
[**UpdateClusterInfo**](ClustersApi.md#UpdateClusterInfo) | **Post** /api/v1/clusters/{clusterUuid}/cluster-info | Update cluster domain, version and status for a cluster.
[**UpdateClusterInfo1**](ClustersApi.md#UpdateClusterInfo1) | **Post** /v1/k8s/clusters/{uuid}/cluster-info | Send cluster-info. This endpoint is reserved for internal run:ai purposes.
[**UpdateClusterLiveness**](ClustersApi.md#UpdateClusterLiveness) | **Post** /api/v1/clusters/{clusterUuid}/liveness | Update liveness status for a cluster.
[**UpdateClusterLiveness1**](ClustersApi.md#UpdateClusterLiveness1) | **Post** /v1/k8s/clusters/{uuid}/liveness | Update liveness status for a specific cluster. This endpoint is reserved for internal run:ai purposes.
[**V1K8sClustersClusterUuidConfigCustomGet**](ClustersApi.md#V1K8sClustersClusterUuidConfigCustomGet) | **Get** /v1/k8s/clusters/{cluster_uuid}/config/custom | Get the custom cluster configuration.
[**V1K8sClustersClusterUuidConfigCustomPost**](ClustersApi.md#V1K8sClustersClusterUuidConfigCustomPost) | **Post** /v1/k8s/clusters/{cluster_uuid}/config/custom | Update a custom cluster configuration.
[**V1K8sClustersClusterUuidConfigGet**](ClustersApi.md#V1K8sClustersClusterUuidConfigGet) | **Get** /v1/k8s/clusters/{cluster_uuid}/config | Get cluster configuration.
[**V1K8sClustersGet**](ClustersApi.md#V1K8sClustersGet) | **Get** /v1/k8s/clusters | List clusters.
[**V1K8sClustersUuidDelete**](ClustersApi.md#V1K8sClustersUuidDelete) | **Delete** /v1/k8s/clusters/{uuid} | Delete a Cluster
[**V1K8sClustersUuidGet**](ClustersApi.md#V1K8sClustersUuidGet) | **Get** /v1/k8s/clusters/{uuid} | Get cluster by id



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
    openapiclient "./openapi"
)

func main() {
    clusterCreationRequest := *openapiclient.NewClusterCreationRequest("Name_example") // ClusterCreationRequest | The cluster to create.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.CreateCluster(context.Background()).ClusterCreationRequest(clusterCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.CreateCluster``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateCluster`: DisplayedCluster
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.CreateCluster`: %v\n", resp)
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


## CreateCluster_0

> Cluster CreateCluster_0(ctx).CreateCluster(createCluster).Execute()

Create a cluster



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
    createCluster := *openapiclient.NewCreateCluster("cluster-a") // CreateCluster | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.CreateCluster_0(context.Background()).CreateCluster(createCluster).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.CreateCluster_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateCluster_0`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.CreateCluster_0`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateCluster_1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createCluster** | [**CreateCluster**](CreateCluster.md) |  | 

### Return type

[**Cluster**](Cluster.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteCluster

> DeleteCluster(ctx, clusterUuid).Execute()

Delete a cluster.



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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.DeleteCluster(context.Background(), clusterUuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.DeleteCluster``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeleteClusterRequest struct via the builder pattern


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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.GetClusterAncestors(context.Background(), clusterUuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.GetClusterAncestors``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetClusterAncestors`: Ancestors
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.GetClusterAncestors`: %v\n", resp)
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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    verbosity := "full" // string | response verbosity level.  (optional) (default to "full")

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.GetClusterByUuid(context.Background(), clusterUuid).Verbosity(verbosity).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.GetClusterByUuid``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetClusterByUuid`: DisplayedCluster
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.GetClusterByUuid`: %v\n", resp)
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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    version := "2.16" // string | The cluster version to install
    remoteClusterUrl := "https://cluster.runai" // string | The remote URL of the runai cluster (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.GetClusterInstallInfoByUuid(context.Background(), clusterUuid).Version(version).RemoteClusterUrl(remoteClusterUrl).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.GetClusterInstallInfoByUuid``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetClusterInstallInfoByUuid`: ClusterInstallationInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.GetClusterInstallInfoByUuid`: %v\n", resp)
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


## GetClusterMetrics

> MetricsResponse GetClusterMetrics(ctx, clusterUuid).Start(start).End(end).MetricType(metricType).NumberOfSamples(numberOfSamples).Execute()

Get the cluster metrics data.



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
    start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
    end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
    metricType := []openapiclient.MetricsType{openapiclient.MetricsType("GPU_UTILIZATION")} // []MetricsType | specifies what data to request
    numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.GetClusterMetrics(context.Background(), clusterUuid).Start(start).End(end).MetricType(metricType).NumberOfSamples(numberOfSamples).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.GetClusterMetrics``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetClusterMetrics`: MetricsResponse
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.GetClusterMetrics`: %v\n", resp)
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


## GetClusterMetrics_0

> Cluster GetClusterMetrics_0(ctx, clusterUuid).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()

Get cluster metrics.



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
    resp, r, err := apiClient.ClustersApi.GetClusterMetrics_0(context.Background(), clusterUuid).Start(start).End(end).NumberOfSamples(numberOfSamples).NodepoolName(nodepoolName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.GetClusterMetrics_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetClusterMetrics_0`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.GetClusterMetrics_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterMetrics_2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **start** | **time.Time** | Start of time range to fetch data from in UTC format. | 
 **end** | **time.Time** | End of time range to fetch data from in UTC format. | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]
 **nodepoolName** | **string** | Filter by unique nodepool name. | 

### Return type

[**Cluster**](Cluster.md)

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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.GetClusterName(context.Background(), clusterUuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.GetClusterName``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetClusterName`: GetClusterName200Response
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.GetClusterName`: %v\n", resp)
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


## GetClusterVersions

> ClusterVersions GetClusterVersions(ctx).Execute()

Get cluster versions.



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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.GetClusterVersions(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.GetClusterVersions``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetClusterVersions`: ClusterVersions
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.GetClusterVersions`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterVersionsRequest struct via the builder pattern


### Return type

[**ClusterVersions**](ClusterVersions.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClusters

> []DisplayedCluster GetClusters(ctx).Verbosity(verbosity).TenantId(tenantId).Execute()

Get a list of clusters.



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
    verbosity := "full" // string | response verbosity level.  (optional) (default to "full")
    tenantId := []string{"Inner_example"} // []string | The tenant Ids of the current request. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.GetClusters(context.Background()).Verbosity(verbosity).TenantId(tenantId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.GetClusters``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetClusters`: []DisplayedCluster
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.GetClusters`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetClustersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **verbosity** | **string** | response verbosity level.  | [default to &quot;full&quot;]
 **tenantId** | **[]string** | The tenant Ids of the current request. | 

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


## GetInstallFile

> string GetInstallFile(ctx, clusterUuid).Cloud(cloud).Clusterip(clusterip).Format(format).Execute()

Get cluster installation file by id.



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
    clusterUuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster.
    cloud := "cloud_example" // string | Cloud type identifier. (optional)
    clusterip := "clusterip_example" // string | Comma-separated list of IP addresses that provide access to the cluster. (optional)
    format := "format_example" // string | Format of the output file. (optional) (default to "yaml")

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.GetInstallFile(context.Background(), clusterUuid).Cloud(cloud).Clusterip(clusterip).Format(format).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.GetInstallFile``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetInstallFile`: string
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.GetInstallFile`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **cloud** | **string** | Cloud type identifier. | 
 **clusterip** | **string** | Comma-separated list of IP addresses that provide access to the cluster. | 
 **format** | **string** | Format of the output file. | [default to &quot;yaml&quot;]

### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InternalClustersUuidSubResourcesDelete

> InternalClustersUuidSubResourcesDelete(ctx, uuid).Execute()

Delete all cluster sub-resources

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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.InternalClustersUuidSubResourcesDelete(context.Background(), uuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.InternalClustersUuidSubResourcesDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiInternalClustersUuidSubResourcesDeleteRequest struct via the builder pattern


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


## InternalClustersUuidSubResourcesPost

> InternalClustersUuidSubResourcesPost(ctx, uuid).ClusterMetaData(clusterMetaData).Execute()

Create all cluster sub-resources

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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster.
    clusterMetaData := *openapiclient.NewClusterMetaData() // ClusterMetaData | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.InternalClustersUuidSubResourcesPost(context.Background(), uuid).ClusterMetaData(clusterMetaData).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.InternalClustersUuidSubResourcesPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiInternalClustersUuidSubResourcesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **clusterMetaData** | [**ClusterMetaData**](ClusterMetaData.md) |  | 

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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    clusterUpdateRequest := *openapiclient.NewClusterUpdateRequest("Name_example") // ClusterUpdateRequest | The cluster details to update

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.UpdateCluster(context.Background(), clusterUuid).ClusterUpdateRequest(clusterUpdateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.UpdateCluster``: %v\n", err)
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


## UpdateCluster1

> Cluster UpdateCluster1(ctx, uuid).Cluster(cluster).Execute()

Update a Cluster by Id



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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster.
    cluster := *openapiclient.NewCluster(*openapiclient.NewClusterMetadata("71f69d83-ba66-4822-adf5-55ce55efd210", "cluster-a"), *openapiclient.NewClusterCurrent([]openapiclient.ClusterCurrentResourcesInner{*openapiclient.NewClusterCurrentResourcesInner(int32(1), *openapiclient.NewResourceData(float32(3), float32(2.5)), *openapiclient.NewResourceData(float32(3), float32(2.5)), *openapiclient.NewResourceData(float32(3), float32(2.5)), "nodepoola")}, []openapiclient.ClusterCurrentProjectResourcesInner{*openapiclient.NewClusterCurrentProjectResourcesInner(int32(1), *openapiclient.NewResourceData(float32(3), float32(2.5)), *openapiclient.NewResourceData(float32(3), float32(2.5)), *openapiclient.NewResourceData(float32(3), float32(2.5)), "project-a", "default", "nodepoola")})) // Cluster | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.UpdateCluster1(context.Background(), uuid).Cluster(cluster).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.UpdateCluster1``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateCluster1`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.UpdateCluster1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateCluster1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **cluster** | [**Cluster**](Cluster.md) |  | 

### Return type

[**Cluster**](Cluster.md)

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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    clusterInfoSyncRequest := *openapiclient.NewClusterInfoSyncRequest("Domain_example", "Version_example") // ClusterInfoSyncRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.UpdateClusterInfo(context.Background(), clusterUuid).ClusterInfoSyncRequest(clusterInfoSyncRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.UpdateClusterInfo``: %v\n", err)
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


## UpdateClusterInfo1

> UpdateClusterInfo1(ctx, uuid).ClusterInfo(clusterInfo).Execute()

Send cluster-info. This endpoint is reserved for internal run:ai purposes.

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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    clusterInfo := *openapiclient.NewClusterInfo("Domain_example", "Version_example") // ClusterInfo | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.UpdateClusterInfo1(context.Background(), uuid).ClusterInfo(clusterInfo).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.UpdateClusterInfo1``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateClusterInfo1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **clusterInfo** | [**ClusterInfo**](ClusterInfo.md) |  | 

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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    clusterLivenessSyncRequest := *openapiclient.NewClusterLivenessSyncRequest(int64(123)) // ClusterLivenessSyncRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.UpdateClusterLiveness(context.Background(), clusterUuid).ClusterLivenessSyncRequest(clusterLivenessSyncRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.UpdateClusterLiveness``: %v\n", err)
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


## UpdateClusterLiveness1

> UpdateClusterLiveness1(ctx, uuid).UpdateClusterLiveness1Request(updateClusterLiveness1Request).Execute()

Update liveness status for a specific cluster. This endpoint is reserved for internal run:ai purposes.

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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
    updateClusterLiveness1Request := *openapiclient.NewUpdateClusterLiveness1Request() // UpdateClusterLiveness1Request | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.UpdateClusterLiveness1(context.Background(), uuid).UpdateClusterLiveness1Request(updateClusterLiveness1Request).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.UpdateClusterLiveness1``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateClusterLiveness1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateClusterLiveness1Request** | [**UpdateClusterLiveness1Request**](UpdateClusterLiveness1Request.md) |  | 

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


## V1K8sClustersClusterUuidConfigCustomGet

> ClusterCustomConfig V1K8sClustersClusterUuidConfigCustomGet(ctx, clusterUuid).Execute()

Get the custom cluster configuration.



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
    clusterUuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.V1K8sClustersClusterUuidConfigCustomGet(context.Background(), clusterUuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.V1K8sClustersClusterUuidConfigCustomGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterUuidConfigCustomGet`: ClusterCustomConfig
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.V1K8sClustersClusterUuidConfigCustomGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterUuidConfigCustomGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ClusterCustomConfig**](ClusterCustomConfig.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersClusterUuidConfigCustomPost

> ClusterCustomConfig V1K8sClustersClusterUuidConfigCustomPost(ctx, clusterUuid).ClusterCustomConfig(clusterCustomConfig).Execute()

Update a custom cluster configuration.



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
    clusterUuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster.
    clusterCustomConfig := *openapiclient.NewClusterCustomConfig() // ClusterCustomConfig | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.V1K8sClustersClusterUuidConfigCustomPost(context.Background(), clusterUuid).ClusterCustomConfig(clusterCustomConfig).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.V1K8sClustersClusterUuidConfigCustomPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterUuidConfigCustomPost`: ClusterCustomConfig
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.V1K8sClustersClusterUuidConfigCustomPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterUuidConfigCustomPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **clusterCustomConfig** | [**ClusterCustomConfig**](ClusterCustomConfig.md) |  | 

### Return type

[**ClusterCustomConfig**](ClusterCustomConfig.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersClusterUuidConfigGet

> map[string]interface{} V1K8sClustersClusterUuidConfigGet(ctx, clusterUuid).ClusterDomain(clusterDomain).ClusterVersion(clusterVersion).Execute()

Get cluster configuration.



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
    clusterUuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The Universally Unique Identifier (UUID) of a cluster.
    clusterDomain := "clusterDomain_example" // string | The domain of the cluster. (optional)
    clusterVersion := "clusterVersion_example" // string | The version of the configuration. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.V1K8sClustersClusterUuidConfigGet(context.Background(), clusterUuid).ClusterDomain(clusterDomain).ClusterVersion(clusterVersion).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.V1K8sClustersClusterUuidConfigGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterUuidConfigGet`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.V1K8sClustersClusterUuidConfigGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of a cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterUuidConfigGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **clusterDomain** | **string** | The domain of the cluster. | 
 **clusterVersion** | **string** | The version of the configuration. | 

### Return type

**map[string]interface{}**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersGet

> []Cluster V1K8sClustersGet(ctx).Execute()

List clusters.



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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.V1K8sClustersGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.V1K8sClustersGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersGet`: []Cluster
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.V1K8sClustersGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersGetRequest struct via the builder pattern


### Return type

[**[]Cluster**](Cluster.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersUuidDelete

> V1K8sClustersUuidDelete200Response V1K8sClustersUuidDelete(ctx, uuid).Execute()

Delete a Cluster



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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.V1K8sClustersUuidDelete(context.Background(), uuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.V1K8sClustersUuidDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersUuidDelete`: V1K8sClustersUuidDelete200Response
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.V1K8sClustersUuidDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersUuidDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**V1K8sClustersUuidDelete200Response**](V1K8sClustersUuidDelete200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersUuidGet

> Cluster V1K8sClustersUuidGet(ctx, uuid).Execute()

Get cluster by id



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
    uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.V1K8sClustersUuidGet(context.Background(), uuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.V1K8sClustersUuidGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersUuidGet`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.V1K8sClustersUuidGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersUuidGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Cluster**](Cluster.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

