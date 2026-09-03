# \SynchronizationApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AccessKeySyncInfo**](SynchronizationApi.md#AccessKeySyncInfo) | **Get** /api/v1/asset/sync/credentials/access-key | Get access keys sync info.
[**AccessKeySyncInfoV2**](SynchronizationApi.md#AccessKeySyncInfoV2) | **Get** /api/v2/asset/sync/credentials/access-key | Get access keys sync info.
[**DockerRegistrySyncInfo**](SynchronizationApi.md#DockerRegistrySyncInfo) | **Get** /api/v1/asset/sync/credentials/docker-registry | Get docker registry sync info.
[**DockerRegistrySyncInfoV2**](SynchronizationApi.md#DockerRegistrySyncInfoV2) | **Get** /api/v2/asset/sync/credentials/docker-registry | Get docker registry sync info.
[**ListAssetsById**](SynchronizationApi.md#ListAssetsById) | **Get** /api/v1/asset | List assets by their ids
[**PasswordSyncInfo**](SynchronizationApi.md#PasswordSyncInfo) | **Get** /api/v1/asset/sync/credentials/password | Get passwords sync info.
[**PasswordSyncInfoV2**](SynchronizationApi.md#PasswordSyncInfoV2) | **Get** /api/v2/asset/sync/credentials/password | Get passwords sync info.
[**PvcSyncInfo**](SynchronizationApi.md#PvcSyncInfo) | **Get** /api/v1/asset/sync/storage/pvc | Get pvcs sync info.
[**PvcSyncInfoV2**](SynchronizationApi.md#PvcSyncInfoV2) | **Get** /api/v2/asset/sync/storage/pvc | Get pvcs sync info.
[**SyncAccessKeys**](SynchronizationApi.md#SyncAccessKeys) | **Post** /api/v1/asset/sync/credentials/access-key | Sync access key assets.
[**SyncAccessKeysV2**](SynchronizationApi.md#SyncAccessKeysV2) | **Post** /api/v2/asset/sync/credentials/access-key | Sync access key assets.
[**SyncDockerRegistry**](SynchronizationApi.md#SyncDockerRegistry) | **Post** /api/v1/asset/sync/credentials/docker-registry | Sync docker registry credentials assets.
[**SyncDockerRegistryV2**](SynchronizationApi.md#SyncDockerRegistryV2) | **Post** /api/v2/asset/sync/credentials/docker-registry | Sync docker registry credentials assets.
[**SyncPasswords**](SynchronizationApi.md#SyncPasswords) | **Post** /api/v1/asset/sync/credentials/password | Sync password assets.
[**SyncPasswordsV2**](SynchronizationApi.md#SyncPasswordsV2) | **Post** /api/v2/asset/sync/credentials/password | Sync password assets.
[**SyncPvcs**](SynchronizationApi.md#SyncPvcs) | **Post** /api/v1/asset/sync/storage/pvc | Sync pvc assets.
[**SyncPvcsStatus**](SynchronizationApi.md#SyncPvcsStatus) | **Put** /api/v2/asset/sync/storage/pvc/{clusterUuid}/status | Sync pvc assets status.
[**SyncPvcsV2**](SynchronizationApi.md#SyncPvcsV2) | **Post** /api/v2/asset/sync/storage/pvc | Sync pvc assets.



## AccessKeySyncInfo

> SyncInfoResponse AccessKeySyncInfo(ctx).ClusterId(clusterId).Execute()

Get access keys sync info.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The id of the cluster. If empty - tenant scoped assets. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.AccessKeySyncInfo(context.Background()).ClusterId(clusterId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.AccessKeySyncInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AccessKeySyncInfo`: SyncInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.AccessKeySyncInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAccessKeySyncInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster. If empty - tenant scoped assets. | 

### Return type

[**SyncInfoResponse**](SyncInfoResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccessKeySyncInfoV2

> SyncInfoResponseV2 AccessKeySyncInfoV2(ctx).ClusterId(clusterId).Scope(scope).Execute()

Get access keys sync info.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The id of the cluster. If empty - tenant scoped assets. (optional)
    scope := "scope_example" // string | the scope level of the asset. If empty - tenant scoped assets. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.AccessKeySyncInfoV2(context.Background()).ClusterId(clusterId).Scope(scope).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.AccessKeySyncInfoV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AccessKeySyncInfoV2`: SyncInfoResponseV2
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.AccessKeySyncInfoV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAccessKeySyncInfoV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster. If empty - tenant scoped assets. | 
 **scope** | **string** | the scope level of the asset. If empty - tenant scoped assets. | 

### Return type

[**SyncInfoResponseV2**](SyncInfoResponseV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DockerRegistrySyncInfo

> SyncInfoResponse DockerRegistrySyncInfo(ctx).ClusterId(clusterId).Execute()

Get docker registry sync info.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The id of the cluster. If empty - tenant scoped assets. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.DockerRegistrySyncInfo(context.Background()).ClusterId(clusterId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.DockerRegistrySyncInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DockerRegistrySyncInfo`: SyncInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.DockerRegistrySyncInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDockerRegistrySyncInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster. If empty - tenant scoped assets. | 

### Return type

[**SyncInfoResponse**](SyncInfoResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DockerRegistrySyncInfoV2

> SyncInfoResponseV2 DockerRegistrySyncInfoV2(ctx).ClusterId(clusterId).Scope(scope).Execute()

Get docker registry sync info.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The id of the cluster. If empty - tenant scoped assets. (optional)
    scope := "scope_example" // string | the scope level of the asset. If empty - tenant scoped assets. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.DockerRegistrySyncInfoV2(context.Background()).ClusterId(clusterId).Scope(scope).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.DockerRegistrySyncInfoV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DockerRegistrySyncInfoV2`: SyncInfoResponseV2
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.DockerRegistrySyncInfoV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDockerRegistrySyncInfoV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster. If empty - tenant scoped assets. | 
 **scope** | **string** | the scope level of the asset. If empty - tenant scoped assets. | 

### Return type

[**SyncInfoResponseV2**](SyncInfoResponseV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListAssetsById

> AssetListByIdsResponse ListAssetsById(ctx).Id(id).Execute()

List assets by their ids



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
    id := "id_example" // string | comma seperated list of asset ids

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.ListAssetsById(context.Background()).Id(id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.ListAssetsById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListAssetsById`: AssetListByIdsResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.ListAssetsById`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListAssetsByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | comma seperated list of asset ids | 

### Return type

[**AssetListByIdsResponse**](AssetListByIdsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PasswordSyncInfo

> SyncInfoResponse PasswordSyncInfo(ctx).ClusterId(clusterId).Execute()

Get passwords sync info.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The id of the cluster. If empty - tenant scoped assets. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.PasswordSyncInfo(context.Background()).ClusterId(clusterId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.PasswordSyncInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `PasswordSyncInfo`: SyncInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.PasswordSyncInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPasswordSyncInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster. If empty - tenant scoped assets. | 

### Return type

[**SyncInfoResponse**](SyncInfoResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PasswordSyncInfoV2

> SyncInfoResponseV2 PasswordSyncInfoV2(ctx).ClusterId(clusterId).Scope(scope).Execute()

Get passwords sync info.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The id of the cluster. If empty - tenant scoped assets. (optional)
    scope := "scope_example" // string | the scope level of the asset. If empty - tenant scoped assets. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.PasswordSyncInfoV2(context.Background()).ClusterId(clusterId).Scope(scope).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.PasswordSyncInfoV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `PasswordSyncInfoV2`: SyncInfoResponseV2
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.PasswordSyncInfoV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPasswordSyncInfoV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster. If empty - tenant scoped assets. | 
 **scope** | **string** | the scope level of the asset. If empty - tenant scoped assets. | 

### Return type

[**SyncInfoResponseV2**](SyncInfoResponseV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PvcSyncInfo

> SyncInfoResponse PvcSyncInfo(ctx).ClusterId(clusterId).Execute()

Get pvcs sync info.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The id of the cluster. If empty - tenant scoped assets. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.PvcSyncInfo(context.Background()).ClusterId(clusterId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.PvcSyncInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `PvcSyncInfo`: SyncInfoResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.PvcSyncInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPvcSyncInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster. If empty - tenant scoped assets. | 

### Return type

[**SyncInfoResponse**](SyncInfoResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PvcSyncInfoV2

> SyncInfoResponseV2 PvcSyncInfoV2(ctx).ClusterId(clusterId).Scope(scope).Execute()

Get pvcs sync info.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The id of the cluster. If empty - tenant scoped assets. (optional)
    scope := "scope_example" // string | the scope level of the asset. If empty - tenant scoped assets. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.PvcSyncInfoV2(context.Background()).ClusterId(clusterId).Scope(scope).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.PvcSyncInfoV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `PvcSyncInfoV2`: SyncInfoResponseV2
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.PvcSyncInfoV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPvcSyncInfoV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster. If empty - tenant scoped assets. | 
 **scope** | **string** | the scope level of the asset. If empty - tenant scoped assets. | 

### Return type

[**SyncInfoResponseV2**](SyncInfoResponseV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncAccessKeys

> AccessKeysSyncResponse SyncAccessKeys(ctx).SyncRequest(syncRequest).Execute()

Sync access key assets.



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
    syncRequest := *openapiclient.NewSyncRequest() // SyncRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.SyncAccessKeys(context.Background()).SyncRequest(syncRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.SyncAccessKeys``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncAccessKeys`: AccessKeysSyncResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.SyncAccessKeys`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncAccessKeysRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncRequest** | [**SyncRequest**](SyncRequest.md) |  | 

### Return type

[**AccessKeysSyncResponse**](AccessKeysSyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncAccessKeysV2

> AccessKeysSyncResponse SyncAccessKeysV2(ctx).SyncRequestV2(syncRequestV2).Execute()

Sync access key assets.



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
    syncRequestV2 := *openapiclient.NewSyncRequestV2() // SyncRequestV2 |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.SyncAccessKeysV2(context.Background()).SyncRequestV2(syncRequestV2).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.SyncAccessKeysV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncAccessKeysV2`: AccessKeysSyncResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.SyncAccessKeysV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncAccessKeysV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncRequestV2** | [**SyncRequestV2**](SyncRequestV2.md) |  | 

### Return type

[**AccessKeysSyncResponse**](AccessKeysSyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDockerRegistry

> DockerRegistrySyncResponse SyncDockerRegistry(ctx).SyncRequest(syncRequest).Execute()

Sync docker registry credentials assets.



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
    syncRequest := *openapiclient.NewSyncRequest() // SyncRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.SyncDockerRegistry(context.Background()).SyncRequest(syncRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.SyncDockerRegistry``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncDockerRegistry`: DockerRegistrySyncResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.SyncDockerRegistry`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncDockerRegistryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncRequest** | [**SyncRequest**](SyncRequest.md) |  | 

### Return type

[**DockerRegistrySyncResponse**](DockerRegistrySyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDockerRegistryV2

> DockerRegistrySyncResponse SyncDockerRegistryV2(ctx).SyncRequestV2(syncRequestV2).Execute()

Sync docker registry credentials assets.



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
    syncRequestV2 := *openapiclient.NewSyncRequestV2() // SyncRequestV2 |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.SyncDockerRegistryV2(context.Background()).SyncRequestV2(syncRequestV2).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.SyncDockerRegistryV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncDockerRegistryV2`: DockerRegistrySyncResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.SyncDockerRegistryV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncDockerRegistryV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncRequestV2** | [**SyncRequestV2**](SyncRequestV2.md) |  | 

### Return type

[**DockerRegistrySyncResponse**](DockerRegistrySyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncPasswords

> PasswordsSyncResponse SyncPasswords(ctx).SyncRequest(syncRequest).Execute()

Sync password assets.



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
    syncRequest := *openapiclient.NewSyncRequest() // SyncRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.SyncPasswords(context.Background()).SyncRequest(syncRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.SyncPasswords``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncPasswords`: PasswordsSyncResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.SyncPasswords`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncPasswordsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncRequest** | [**SyncRequest**](SyncRequest.md) |  | 

### Return type

[**PasswordsSyncResponse**](PasswordsSyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncPasswordsV2

> PasswordsSyncResponse SyncPasswordsV2(ctx).SyncRequestV2(syncRequestV2).Execute()

Sync password assets.



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
    syncRequestV2 := *openapiclient.NewSyncRequestV2() // SyncRequestV2 |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.SyncPasswordsV2(context.Background()).SyncRequestV2(syncRequestV2).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.SyncPasswordsV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncPasswordsV2`: PasswordsSyncResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.SyncPasswordsV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncPasswordsV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncRequestV2** | [**SyncRequestV2**](SyncRequestV2.md) |  | 

### Return type

[**PasswordsSyncResponse**](PasswordsSyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncPvcs

> PVCsSyncResponse SyncPvcs(ctx).SyncRequest(syncRequest).Execute()

Sync pvc assets.



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
    syncRequest := *openapiclient.NewSyncRequest() // SyncRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.SyncPvcs(context.Background()).SyncRequest(syncRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.SyncPvcs``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncPvcs`: PVCsSyncResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.SyncPvcs`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncPvcsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncRequest** | [**SyncRequest**](SyncRequest.md) |  | 

### Return type

[**PVCsSyncResponse**](PVCsSyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncPvcsStatus

> SyncPvcsStatus(ctx, clusterUuid).PVCStatusRequest(pVCStatusRequest).Execute()

Sync pvc assets status.



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
    pVCStatusRequest := *openapiclient.NewPVCStatusRequest() // PVCStatusRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.SyncPvcsStatus(context.Background(), clusterUuid).PVCStatusRequest(pVCStatusRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.SyncPvcsStatus``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | Unique identifier of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncPvcsStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **pVCStatusRequest** | [**PVCStatusRequest**](PVCStatusRequest.md) |  | 

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


## SyncPvcsV2

> PVCsSyncResponse SyncPvcsV2(ctx).SyncRequestV2(syncRequestV2).Execute()

Sync pvc assets.



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
    syncRequestV2 := *openapiclient.NewSyncRequestV2() // SyncRequestV2 |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SynchronizationApi.SyncPvcsV2(context.Background()).SyncRequestV2(syncRequestV2).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationApi.SyncPvcsV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncPvcsV2`: PVCsSyncResponse
    fmt.Fprintf(os.Stdout, "Response from `SynchronizationApi.SyncPvcsV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncPvcsV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncRequestV2** | [**SyncRequestV2**](SyncRequestV2.md) |  | 

### Return type

[**PVCsSyncResponse**](PVCsSyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

