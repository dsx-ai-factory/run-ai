# \ClustersApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateCluster**](ClustersApi.md#CreateCluster) | **Post** /v1/k8s/clusters | Create a cluster
[**UpdateCluster**](ClustersApi.md#UpdateCluster) | **Put** /v1/k8s/clusters/{uuid} | Update a Cluster by Id
[**UpdateClusterInfo**](ClustersApi.md#UpdateClusterInfo) | **Post** /v1/k8s/clusters/{uuid}/cluster-info | Send cluster-info. This endpoint is reserved for internal run:ai purposes.
[**UpdateClusterLiveness**](ClustersApi.md#UpdateClusterLiveness) | **Post** /v1/k8s/clusters/{uuid}/liveness | Update liveness status for a specific cluster. This endpoint is reserved for internal run:ai purposes.
[**V1K8sClustersClusterUuidConfigControlPlaneGet**](ClustersApi.md#V1K8sClustersClusterUuidConfigControlPlaneGet) | **Get** /v1/k8s/clusters/{cluster_uuid}/config/control-plane | Returns cluster&#39;s control plane configuration
[**V1K8sClustersClusterUuidConfigCustomGet**](ClustersApi.md#V1K8sClustersClusterUuidConfigCustomGet) | **Get** /v1/k8s/clusters/{cluster_uuid}/config/custom | Returns custom cluster configuration
[**V1K8sClustersClusterUuidConfigCustomPost**](ClustersApi.md#V1K8sClustersClusterUuidConfigCustomPost) | **Post** /v1/k8s/clusters/{cluster_uuid}/config/custom | Update custom cluster configuration
[**V1K8sClustersClusterUuidConfigGet**](ClustersApi.md#V1K8sClustersClusterUuidConfigGet) | **Get** /v1/k8s/clusters/{cluster_uuid}/config | Returns cluster configuration
[**V1K8sClustersClusterUuidInstallfileGet**](ClustersApi.md#V1K8sClustersClusterUuidInstallfileGet) | **Get** /v1/k8s/clusters/{cluster_uuid}/installfile | Returns cluster installation file by Id
[**V1K8sClustersGet**](ClustersApi.md#V1K8sClustersGet) | **Get** /v1/k8s/clusters | List clusters.
[**V1K8sClustersUuidDelete**](ClustersApi.md#V1K8sClustersUuidDelete) | **Delete** /v1/k8s/clusters/{uuid} | Delete a Cluster
[**V1K8sClustersUuidGet**](ClustersApi.md#V1K8sClustersUuidGet) | **Get** /v1/k8s/clusters/{uuid} | Get cluster by id



## CreateCluster

> Cluster CreateCluster(ctx).CreateCluster(createCluster).Execute()

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
    resp, r, err := apiClient.ClustersApi.CreateCluster(context.Background()).CreateCluster(createCluster).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.CreateCluster``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateCluster`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.CreateCluster`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateClusterRequest struct via the builder pattern


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


## UpdateCluster

> Cluster UpdateCluster(ctx, uuid).Cluster(cluster).Execute()

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
    cluster := *openapiclient.NewCluster() // Cluster | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.UpdateCluster(context.Background(), uuid).Cluster(cluster).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.UpdateCluster``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateCluster`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.UpdateCluster`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateClusterRequest struct via the builder pattern


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

> UpdateClusterInfo(ctx, uuid).ClusterInfo(clusterInfo).Execute()

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
    resp, r, err := apiClient.ClustersApi.UpdateClusterInfo(context.Background(), uuid).ClusterInfo(clusterInfo).Execute()
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
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateClusterInfoRequest struct via the builder pattern


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

> UpdateClusterLiveness(ctx, uuid).UpdateClusterLivenessRequest(updateClusterLivenessRequest).Execute()

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
    updateClusterLivenessRequest := *openapiclient.NewUpdateClusterLivenessRequest() // UpdateClusterLivenessRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClustersApi.UpdateClusterLiveness(context.Background(), uuid).UpdateClusterLivenessRequest(updateClusterLivenessRequest).Execute()
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
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateClusterLivenessRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateClusterLivenessRequest** | [**UpdateClusterLivenessRequest**](UpdateClusterLivenessRequest.md) |  | 

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


## V1K8sClustersClusterUuidConfigControlPlaneGet

> ClusterControlPlaneConfig V1K8sClustersClusterUuidConfigControlPlaneGet(ctx, clusterUuid).Execute()

Returns cluster's control plane configuration



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
    resp, r, err := apiClient.ClustersApi.V1K8sClustersClusterUuidConfigControlPlaneGet(context.Background(), clusterUuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.V1K8sClustersClusterUuidConfigControlPlaneGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterUuidConfigControlPlaneGet`: ClusterControlPlaneConfig
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.V1K8sClustersClusterUuidConfigControlPlaneGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterUuidConfigControlPlaneGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ClusterControlPlaneConfig**](ClusterControlPlaneConfig.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1K8sClustersClusterUuidConfigCustomGet

> ClusterCustomConfig V1K8sClustersClusterUuidConfigCustomGet(ctx, clusterUuid).Execute()

Returns custom cluster configuration



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

Update custom cluster configuration



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

Returns cluster configuration



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
**clusterUuid** | **string** | Unique identifier of the cluster. | 

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


## V1K8sClustersClusterUuidInstallfileGet

> string V1K8sClustersClusterUuidInstallfileGet(ctx, clusterUuid).Cloud(cloud).Clusterip(clusterip).Format(format).Execute()

Returns cluster installation file by Id



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
    resp, r, err := apiClient.ClustersApi.V1K8sClustersClusterUuidInstallfileGet(context.Background(), clusterUuid).Cloud(cloud).Clusterip(clusterip).Format(format).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClustersApi.V1K8sClustersClusterUuidInstallfileGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `V1K8sClustersClusterUuidInstallfileGet`: string
    fmt.Fprintf(os.Stdout, "Response from `ClustersApi.V1K8sClustersClusterUuidInstallfileGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1K8sClustersClusterUuidInstallfileGetRequest struct via the builder pattern


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

