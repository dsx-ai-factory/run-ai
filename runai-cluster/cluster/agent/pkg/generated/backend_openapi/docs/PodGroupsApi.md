# \PodGroupsApi

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreatePodGroups**](PodGroupsApi.md#CreatePodGroups) | **Post** /v1/k8s/clusters/{uuid}/podGroups | Create podGroups
[**DeletePodGroup**](PodGroupsApi.md#DeletePodGroup) | **Delete** /v1/k8s/clusters/{uuid}/podGroups/{podGroupId} | Delete a podGroup by id
[**SyncPodGroupData**](PodGroupsApi.md#SyncPodGroupData) | **Post** /v1/k8s/clusters/{uuid}/podGroups/sync | Sync All cluster podGroup data to DB
[**UpdatePodGroup**](PodGroupsApi.md#UpdatePodGroup) | **Put** /v1/k8s/clusters/{uuid}/podGroups/{podGroupId} | Update a podGroup by id
[**UpdatePodGroups**](PodGroupsApi.md#UpdatePodGroups) | **Put** /v1/k8s/clusters/{uuid}/podGroups | Sync all podGroups in a project to DB



## CreatePodGroups

> CreatePodGroups(ctx, uuid).PodGroup(podGroup).Execute()

Create podGroups

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
    podGroup := []openapiclient.PodGroup{*openapiclient.NewPodGroup("JobId_example", "PodGroupId_example", "WorkloadName_example", "interactive", "runaijob", "ClusterUuid_example", "tensorflow", "JobPhase_example", "WorkloadUser_example", "Project_example", int64(123), int64(123), int32(123), int32(123), float32(123), float32(123), float32(123), float32(123), "parentWorkloadName")} // []PodGroup | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodGroupsApi.CreatePodGroups(context.Background(), uuid).PodGroup(podGroup).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodGroupsApi.CreatePodGroups``: %v\n", err)
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

Other parameters are passed through a pointer to a apiCreatePodGroupsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **podGroup** | [**[]PodGroup**](PodGroup.md) |  | 

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


## DeletePodGroup

> DeletePodGroup(ctx, uuid, podGroupId).Execute()

Delete a podGroup by id

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
    podGroupId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the podGroup (UUID)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodGroupsApi.DeletePodGroup(context.Background(), uuid, podGroupId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodGroupsApi.DeletePodGroup``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**podGroupId** | **string** | Unique identifier of the podGroup (UUID) | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeletePodGroupRequest struct via the builder pattern


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


## SyncPodGroupData

> SyncPodGroupData(ctx, uuid).PodGroup(podGroup).Execute()

Sync All cluster podGroup data to DB

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
    podGroup := []openapiclient.PodGroup{*openapiclient.NewPodGroup("JobId_example", "PodGroupId_example", "WorkloadName_example", "interactive", "runaijob", "ClusterUuid_example", "tensorflow", "JobPhase_example", "WorkloadUser_example", "Project_example", int64(123), int64(123), int32(123), int32(123), float32(123), float32(123), float32(123), float32(123), "parentWorkloadName")} // []PodGroup | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodGroupsApi.SyncPodGroupData(context.Background(), uuid).PodGroup(podGroup).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodGroupsApi.SyncPodGroupData``: %v\n", err)
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

Other parameters are passed through a pointer to a apiSyncPodGroupDataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **podGroup** | [**[]PodGroup**](PodGroup.md) |  | 

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


## UpdatePodGroup

> UpdatePodGroup(ctx, uuid, podGroupId).PodGroup(podGroup).Execute()

Update a podGroup by id

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
    podGroupId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the podGroup (UUID)
    podGroup := *openapiclient.NewPodGroup("JobId_example", "PodGroupId_example", "WorkloadName_example", "interactive", "runaijob", "ClusterUuid_example", "tensorflow", "JobPhase_example", "WorkloadUser_example", "Project_example", int64(123), int64(123), int32(123), int32(123), float32(123), float32(123), float32(123), float32(123), "parentWorkloadName") // PodGroup | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodGroupsApi.UpdatePodGroup(context.Background(), uuid, podGroupId).PodGroup(podGroup).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodGroupsApi.UpdatePodGroup``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**podGroupId** | **string** | Unique identifier of the podGroup (UUID) | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdatePodGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **podGroup** | [**PodGroup**](PodGroup.md) |  | 

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


## UpdatePodGroups

> IdsResponse UpdatePodGroups(ctx, uuid).PodGroupsSync(podGroupsSync).Execute()

Sync all podGroups in a project to DB

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
    podGroupsSync := *openapiclient.NewPodGroupsSync("team-a", []openapiclient.PodGroup{*openapiclient.NewPodGroup("JobId_example", "PodGroupId_example", "WorkloadName_example", "interactive", "runaijob", "ClusterUuid_example", "tensorflow", "JobPhase_example", "WorkloadUser_example", "Project_example", int64(123), int64(123), int32(123), int32(123), float32(123), float32(123), float32(123), float32(123), "parentWorkloadName")}) // PodGroupsSync | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PodGroupsApi.UpdatePodGroups(context.Background(), uuid).PodGroupsSync(podGroupsSync).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PodGroupsApi.UpdatePodGroups``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdatePodGroups`: IdsResponse
    fmt.Fprintf(os.Stdout, "Response from `PodGroupsApi.UpdatePodGroups`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdatePodGroupsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **podGroupsSync** | [**PodGroupsSync**](PodGroupsSync.md) |  | 

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

