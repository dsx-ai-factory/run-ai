# \PodsApi

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreatePods**](PodsApi.md#CreatePods) | **Post** /v1/k8s/clusters/{uuid}/pods | create pods
[**DeletePod**](PodsApi.md#DeletePod) | **Delete** /v1/k8s/clusters/{uuid}/pods/{podId} | Delete a pod by id
[**GetPods**](PodsApi.md#GetPods) | **Get** /v1/k8s/clusters/{uuid}/pods | get all pods from a specific cluster
[**SyncPodsData**](PodsApi.md#SyncPodsData) | **Post** /v1/k8s/clusters/{uuid}/pods/sync | Sync All cluster pods data to DB
[**UpdatePod**](PodsApi.md#UpdatePod) | **Put** /v1/k8s/clusters/{uuid}/pods/{podId} | Update a pod by id
[**UpdatePods**](PodsApi.md#UpdatePods) | **Put** /v1/k8s/clusters/{uuid}/pods | Sync all pods in a namespace to DB



## CreatePods

> []Pod CreatePods(ctx, uuid).Pod(pod).Execute()

create pods

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
    pod := []openapiclient.Pod{*openapiclient.NewPod("PodId_example", "JobId_example", "ClusterUuid_example", "PodName_example", "ImageName_example", "Phase_example", int64(123), int64(123), int64(123))} // []Pod | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.CreatePods(context.Background(), uuid).Pod(pod).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.CreatePods``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreatePods`: []Pod
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.CreatePods`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreatePodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **pod** | [**[]Pod**](Pod.md) |  | 

### Return type

[**[]Pod**](Pod.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeletePod

> string DeletePod(ctx, uuid, podId).Execute()

Delete a pod by id

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
    podId := "podId_example" // string | Unique identifier of the pod.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.DeletePod(context.Background(), uuid, podId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.DeletePod``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeletePod`: string
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.DeletePod`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**podId** | **string** | Unique identifier of the pod. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeletePodRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



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


## GetPods

> []Pod GetPods(ctx, uuid).Execute()

get all pods from a specific cluster

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
    resp, r, err := apiClient.PodsApi.GetPods(context.Background(), uuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.GetPods``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPods`: []Pod
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.GetPods`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPodsRequest struct via the builder pattern


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


## SyncPodsData

> SyncPodsData(ctx, uuid).Pod(pod).Execute()

Sync All cluster pods data to DB

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
    pod := []openapiclient.Pod{*openapiclient.NewPod("PodId_example", "JobId_example", "ClusterUuid_example", "PodName_example", "ImageName_example", "Phase_example", int64(123), int64(123), int64(123))} // []Pod | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.SyncPodsData(context.Background(), uuid).Pod(pod).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.SyncPodsData``: %v\n", err)
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

Other parameters are passed through a pointer to a apiSyncPodsDataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **pod** | [**[]Pod**](Pod.md) |  | 

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


## UpdatePod

> UpdatePod(ctx, uuid, podId).Pod(pod).Execute()

Update a pod by id

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
    podId := "podId_example" // string | Unique identifier of the pod.
    pod := *openapiclient.NewPod("PodId_example", "JobId_example", "ClusterUuid_example", "PodName_example", "ImageName_example", "Phase_example", int64(123), int64(123), int64(123)) // Pod | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.UpdatePod(context.Background(), uuid, podId).Pod(pod).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.UpdatePod``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**podId** | **string** | Unique identifier of the pod. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdatePodRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **pod** | [**Pod**](Pod.md) |  | 

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


## UpdatePods

> IdsResponse UpdatePods(ctx, uuid).PodsSync(podsSync).Execute()

Sync all pods in a namespace to DB

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
    podsSync := *openapiclient.NewPodsSync("runai-team-a", []openapiclient.Pod{*openapiclient.NewPod("PodId_example", "JobId_example", "ClusterUuid_example", "PodName_example", "ImageName_example", "Phase_example", int64(123), int64(123), int64(123))}) // PodsSync | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodsApi.UpdatePods(context.Background(), uuid).PodsSync(podsSync).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodsApi.UpdatePods``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdatePods`: IdsResponse
    fmt.Fprintf(os.Stdout, "Response from `PodsApi.UpdatePods`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdatePodsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **podsSync** | [**PodsSync**](PodsSync.md) |  | 

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

