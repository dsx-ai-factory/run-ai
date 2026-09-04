# \StorageClassesApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateStorageClass**](StorageClassesApi.md#CreateStorageClass) | **Post** /v1/k8s/clusters/{uuid}/storage-classes | Create a storageClass.
[**DeleteStorageClass**](StorageClassesApi.md#DeleteStorageClass) | **Delete** /v1/k8s/clusters/{uuid}/storage-classes/{storage_class_name} | Delete a storageClass.
[**GetStorageClasses**](StorageClassesApi.md#GetStorageClasses) | **Get** /v1/k8s/clusters/{uuid}/storage-classes | Get all storageClasses from a cluster.
[**SyncStorageClass**](StorageClassesApi.md#SyncStorageClass) | **Post** /v1/k8s/clusters/{uuid}/storage-classes/sync | Sync storageClass to a cluster.
[**UpdateStorageClass**](StorageClassesApi.md#UpdateStorageClass) | **Put** /v1/k8s/clusters/{uuid}/storage-classes/{storage_class_name} | Update a storageClass.



## CreateStorageClass

> StorageClass CreateStorageClass(ctx, uuid).StorageClass(storageClass).Execute()

Create a storageClass.



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
    storageClass := *openapiclient.NewStorageClass("StorageClassName_example") // StorageClass | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.StorageClassesApi.CreateStorageClass(context.Background(), uuid).StorageClass(storageClass).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesApi.CreateStorageClass``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateStorageClass`: StorageClass
    fmt.Fprintf(os.Stdout, "Response from `StorageClassesApi.CreateStorageClass`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateStorageClassRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **storageClass** | [**StorageClass**](StorageClass.md) |  | 

### Return type

[**StorageClass**](StorageClass.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteStorageClass

> DeleteStorageClass(ctx, uuid, storageClassName).Execute()

Delete a storageClass.



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
    storageClassName := "storageClassName_example" // string | Unique identifier of the storageClass

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.StorageClassesApi.DeleteStorageClass(context.Background(), uuid, storageClassName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesApi.DeleteStorageClass``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**storageClassName** | **string** | Unique identifier of the storageClass | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteStorageClassRequest struct via the builder pattern


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


## GetStorageClasses

> []StorageClass GetStorageClasses(ctx, uuid).Execute()

Get all storageClasses from a cluster.



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
    resp, r, err := apiClient.StorageClassesApi.GetStorageClasses(context.Background(), uuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesApi.GetStorageClasses``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetStorageClasses`: []StorageClass
    fmt.Fprintf(os.Stdout, "Response from `StorageClassesApi.GetStorageClasses`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetStorageClassesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]StorageClass**](StorageClass.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncStorageClass

> StorageClass SyncStorageClass(ctx, uuid).StorageClass(storageClass).Execute()

Sync storageClass to a cluster.



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
    storageClass := []openapiclient.StorageClass{*openapiclient.NewStorageClass("StorageClassName_example")} // []StorageClass | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.StorageClassesApi.SyncStorageClass(context.Background(), uuid).StorageClass(storageClass).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesApi.SyncStorageClass``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncStorageClass`: StorageClass
    fmt.Fprintf(os.Stdout, "Response from `StorageClassesApi.SyncStorageClass`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncStorageClassRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **storageClass** | [**[]StorageClass**](StorageClass.md) |  | 

### Return type

[**StorageClass**](StorageClass.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateStorageClass

> []StorageClass UpdateStorageClass(ctx, uuid, storageClassName).StorageClass(storageClass).Execute()

Update a storageClass.



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
    storageClassName := "storageClassName_example" // string | Unique identifier of the storageClass
    storageClass := *openapiclient.NewStorageClass("StorageClassName_example") // StorageClass | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.StorageClassesApi.UpdateStorageClass(context.Background(), uuid, storageClassName).StorageClass(storageClass).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesApi.UpdateStorageClass``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateStorageClass`: []StorageClass
    fmt.Fprintf(os.Stdout, "Response from `StorageClassesApi.UpdateStorageClass`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**storageClassName** | **string** | Unique identifier of the storageClass | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateStorageClassRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **storageClass** | [**StorageClass**](StorageClass.md) |  | 

### Return type

[**[]StorageClass**](StorageClass.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

