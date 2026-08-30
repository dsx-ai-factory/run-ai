# \DeploymentsApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateDeployment**](DeploymentsApi.md#CreateDeployment) | **Post** /v1/k8s/clusters/{uuid}/deployments | Create a deployment
[**DeleteDeployment**](DeploymentsApi.md#DeleteDeployment) | **Delete** /v1/k8s/clusters/{uuid}/deployments/{deploymentId} | Delete a deployment by id
[**GetDeployment**](DeploymentsApi.md#GetDeployment) | **Get** /v1/k8s/clusters/{uuid}/deployments/{deploymentId} | Get a deployment by id
[**GetDeployments**](DeploymentsApi.md#GetDeployments) | **Get** /v1/k8s/clusters/{uuid}/deployments | Get all cluster&#39;s deployments
[**GetDeploymentsCount**](DeploymentsApi.md#GetDeploymentsCount) | **Get** /v1/k8s/clusters/{uuid}/deployments/count | Get the number of deployments
[**GetDeploymentsHistory**](DeploymentsApi.md#GetDeploymentsHistory) | **Get** /v1/k8s/clusters/{uuid}/deployments/deleted | Get all the deployments that were deleted in the cluster
[**GetDeploymentsHistoryCount**](DeploymentsApi.md#GetDeploymentsHistoryCount) | **Get** /v1/k8s/clusters/{uuid}/deployments/deleted/count | Get the number of deleted deployments
[**GetPodsByDeploymentId**](DeploymentsApi.md#GetPodsByDeploymentId) | **Get** /v1/k8s/clusters/{uuid}/deployments/{deploymentId}/pods | Get all pods associated with the deployment by id
[**SyncDeployments**](DeploymentsApi.md#SyncDeployments) | **Post** /v1/k8s/clusters/{uuid}/deployments/sync | Sync deployments
[**UpdateDeployment**](DeploymentsApi.md#UpdateDeployment) | **Put** /v1/k8s/clusters/{uuid}/deployments/{deploymentId} | Update a deployment by id
[**UpdateDeployments**](DeploymentsApi.md#UpdateDeployments) | **Put** /v1/k8s/clusters/{uuid}/deployments | Sync deployments



## CreateDeployment

> CreateDeployment(ctx, uuid).Deployment(deployment).Execute()

Create a deployment

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
    deployment := *openapiclient.NewDeployment("ClusterUuid_example", "Id_example", "Name_example", "Namespace_example", []openapiclient.Owner{*openapiclient.NewOwner("Kind_example", "Name_example", "Id_example")}, int32(123), *openapiclient.NewPodTemplate([]openapiclient.Container{*openapiclient.NewContainer("Image_example", "Name_example")}, map[string]float32{"key": float32(123)}), "SchedulerName_example") // Deployment |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DeploymentsApi.CreateDeployment(context.Background(), uuid).Deployment(deployment).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsApi.CreateDeployment``: %v\n", err)
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

Other parameters are passed through a pointer to a apiCreateDeploymentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deployment** | [**Deployment**](Deployment.md) |  | 

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


## DeleteDeployment

> DeleteDeployment(ctx, uuid, deploymentId).Execute()

Delete a deployment by id

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
    deploymentId := "deploymentId_example" // string | the unique Id of the deployment (UUID)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DeploymentsApi.DeleteDeployment(context.Background(), uuid, deploymentId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsApi.DeleteDeployment``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**deploymentId** | **string** | the unique Id of the deployment (UUID) | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDeploymentRequest struct via the builder pattern


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


## GetDeployment

> Deployment GetDeployment(ctx, uuid, deploymentId).Execute()

Get a deployment by id

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
    deploymentId := "deploymentId_example" // string | the unique Id of the deployment (UUID)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DeploymentsApi.GetDeployment(context.Background(), uuid, deploymentId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsApi.GetDeployment``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDeployment`: Deployment
    fmt.Fprintf(os.Stdout, "Response from `DeploymentsApi.GetDeployment`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**deploymentId** | **string** | the unique Id of the deployment (UUID) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDeploymentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Deployment**](Deployment.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDeployments

> []Deployment GetDeployments(ctx, uuid).Execute()

Get all cluster's deployments

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DeploymentsApi.GetDeployments(context.Background(), uuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsApi.GetDeployments``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDeployments`: []Deployment
    fmt.Fprintf(os.Stdout, "Response from `DeploymentsApi.GetDeployments`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDeploymentsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]Deployment**](Deployment.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDeploymentsCount

> float32 GetDeploymentsCount(ctx, uuid).Execute()

Get the number of deployments

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DeploymentsApi.GetDeploymentsCount(context.Background(), uuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsApi.GetDeploymentsCount``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDeploymentsCount`: float32
    fmt.Fprintf(os.Stdout, "Response from `DeploymentsApi.GetDeploymentsCount`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDeploymentsCountRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**float32**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDeploymentsHistory

> []Deployment GetDeploymentsHistory(ctx, uuid).Execute()

Get all the deployments that were deleted in the cluster

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DeploymentsApi.GetDeploymentsHistory(context.Background(), uuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsApi.GetDeploymentsHistory``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDeploymentsHistory`: []Deployment
    fmt.Fprintf(os.Stdout, "Response from `DeploymentsApi.GetDeploymentsHistory`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDeploymentsHistoryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]Deployment**](Deployment.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDeploymentsHistoryCount

> float32 GetDeploymentsHistoryCount(ctx, uuid).Execute()

Get the number of deleted deployments

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DeploymentsApi.GetDeploymentsHistoryCount(context.Background(), uuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsApi.GetDeploymentsHistoryCount``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDeploymentsHistoryCount`: float32
    fmt.Fprintf(os.Stdout, "Response from `DeploymentsApi.GetDeploymentsHistoryCount`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDeploymentsHistoryCountRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**float32**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPodsByDeploymentId

> []Pod GetPodsByDeploymentId(ctx, uuid, deploymentId).Execute()

Get all pods associated with the deployment by id

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
    deploymentId := "deploymentId_example" // string | the unique Id of the deployment (UUID)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DeploymentsApi.GetPodsByDeploymentId(context.Background(), uuid, deploymentId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsApi.GetPodsByDeploymentId``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPodsByDeploymentId`: []Pod
    fmt.Fprintf(os.Stdout, "Response from `DeploymentsApi.GetPodsByDeploymentId`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**deploymentId** | **string** | the unique Id of the deployment (UUID) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPodsByDeploymentIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**[]Pod**](Pod.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDeployments

> SyncDeployments(ctx, uuid).Deployment(deployment).Execute()

Sync deployments



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
    deployment := []openapiclient.Deployment{*openapiclient.NewDeployment("ClusterUuid_example", "Id_example", "Name_example", "Namespace_example", []openapiclient.Owner{*openapiclient.NewOwner("Kind_example", "Name_example", "Id_example")}, int32(123), *openapiclient.NewPodTemplate([]openapiclient.Container{*openapiclient.NewContainer("Image_example", "Name_example")}, map[string]float32{"key": float32(123)}), "SchedulerName_example")} // []Deployment |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DeploymentsApi.SyncDeployments(context.Background(), uuid).Deployment(deployment).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsApi.SyncDeployments``: %v\n", err)
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

Other parameters are passed through a pointer to a apiSyncDeploymentsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deployment** | [**[]Deployment**](Deployment.md) |  | 

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


## UpdateDeployment

> UpdateDeployment(ctx, uuid, deploymentId).Deployment(deployment).Execute()

Update a deployment by id

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
    deploymentId := "deploymentId_example" // string | the unique Id of the deployment (UUID)
    deployment := *openapiclient.NewDeployment("ClusterUuid_example", "Id_example", "Name_example", "Namespace_example", []openapiclient.Owner{*openapiclient.NewOwner("Kind_example", "Name_example", "Id_example")}, int32(123), *openapiclient.NewPodTemplate([]openapiclient.Container{*openapiclient.NewContainer("Image_example", "Name_example")}, map[string]float32{"key": float32(123)}), "SchedulerName_example") // Deployment | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DeploymentsApi.UpdateDeployment(context.Background(), uuid, deploymentId).Deployment(deployment).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsApi.UpdateDeployment``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**deploymentId** | **string** | the unique Id of the deployment (UUID) | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDeploymentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **deployment** | [**Deployment**](Deployment.md) |  | 

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


## UpdateDeployments

> IdsResponse UpdateDeployments(ctx, uuid).DeploymentsSync(deploymentsSync).Execute()

Sync deployments



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
    deploymentsSync := *openapiclient.NewDeploymentsSync("team-a", []openapiclient.Deployment{*openapiclient.NewDeployment("ClusterUuid_example", "Id_example", "Name_example", "Namespace_example", []openapiclient.Owner{*openapiclient.NewOwner("Kind_example", "Name_example", "Id_example")}, int32(123), *openapiclient.NewPodTemplate([]openapiclient.Container{*openapiclient.NewContainer("Image_example", "Name_example")}, map[string]float32{"key": float32(123)}), "SchedulerName_example")}) // DeploymentsSync | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DeploymentsApi.UpdateDeployments(context.Background(), uuid).DeploymentsSync(deploymentsSync).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsApi.UpdateDeployments``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateDeployments`: IdsResponse
    fmt.Fprintf(os.Stdout, "Response from `DeploymentsApi.UpdateDeployments`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDeploymentsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deploymentsSync** | [**DeploymentsSync**](DeploymentsSync.md) |  | 

### Return type

[**IdsResponse**](IdsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

