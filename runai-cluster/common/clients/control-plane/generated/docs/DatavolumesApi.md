# \DatavolumesApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateDatavolume**](DatavolumesApi.md#CreateDatavolume) | **Post** /api/v1/datavolumes | Create a datavolume
[**DeleteDatavolume**](DatavolumesApi.md#DeleteDatavolume) | **Delete** /api/v1/datavolumes/{datavolumeId} | Delete datavolume
[**GetDatavolume**](DatavolumesApi.md#GetDatavolume) | **Get** /api/v1/datavolumes/{datavolumeId} | Get datavolume
[**GetDatavolumeAncestors**](DatavolumesApi.md#GetDatavolumeAncestors) | **Get** /internal/datavolumes/{datavolumeId}/ancestors | Get the datavolume ancestors
[**GetDatavolumeSharedScopes**](DatavolumesApi.md#GetDatavolumeSharedScopes) | **Get** /api/v1/datavolumes/{datavolumeId}/sharedScopes | Get the datavolume&#39;s shared scopes
[**GetDatavolumes**](DatavolumesApi.md#GetDatavolumes) | **Get** /api/v1/datavolumes | List datavolumes in permitted scopes
[**GetDatavolumesInternal**](DatavolumesApi.md#GetDatavolumesInternal) | **Get** /api/v1/datavolumes/sync | List datavolumes in cluster.
[**PatchDatavolume**](DatavolumesApi.md#PatchDatavolume) | **Patch** /api/v1/datavolumes/{datavolumeId} | Patch datavolume
[**PatchDatavolumeSharedScopes**](DatavolumesApi.md#PatchDatavolumeSharedScopes) | **Patch** /api/v1/datavolumes/{datavolumeId}/sharedScopes | Patch the datavolume&#39;s shared scopes
[**SyncDatavolumes**](DatavolumesApi.md#SyncDatavolumes) | **Post** /api/v1/datavolumes/sync | Sync the statuses of the datavolumes in the cluster
[**UpdateDatavolumeStatus**](DatavolumesApi.md#UpdateDatavolumeStatus) | **Put** /api/v1/datavolumes/{datavolumeId}/status | Update the status of the datavolume form the cluster



## CreateDatavolume

> Datavolume CreateDatavolume(ctx).DatavolumeCreationFields(datavolumeCreationFields).Execute()

Create a datavolume

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
    datavolumeCreationFields := *openapiclient.NewDatavolumeCreationFields("datavolume-a", "pvc-a", "5") // DatavolumeCreationFields | The datavolume to create.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DatavolumesApi.CreateDatavolume(context.Background()).DatavolumeCreationFields(datavolumeCreationFields).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DatavolumesApi.CreateDatavolume``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateDatavolume`: Datavolume
    fmt.Fprintf(os.Stdout, "Response from `DatavolumesApi.CreateDatavolume`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateDatavolumeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **datavolumeCreationFields** | [**DatavolumeCreationFields**](DatavolumeCreationFields.md) | The datavolume to create. | 

### Return type

[**Datavolume**](Datavolume.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteDatavolume

> HttpResponse3 DeleteDatavolume(ctx, datavolumeId).Execute()

Delete datavolume

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
    datavolumeId := "71f69d83-ba66-4822-adf5-55ce55efd210" // string | The id of the datavolume to retrieve

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DatavolumesApi.DeleteDatavolume(context.Background(), datavolumeId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DatavolumesApi.DeleteDatavolume``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteDatavolume`: HttpResponse3
    fmt.Fprintf(os.Stdout, "Response from `DatavolumesApi.DeleteDatavolume`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**datavolumeId** | **string** | The id of the datavolume to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDatavolumeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HttpResponse3**](HttpResponse3.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDatavolume

> Datavolume GetDatavolume(ctx, datavolumeId).Execute()

Get datavolume

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
    datavolumeId := "71f69d83-ba66-4822-adf5-55ce55efd210" // string | The id of the datavolume to retrieve

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DatavolumesApi.GetDatavolume(context.Background(), datavolumeId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DatavolumesApi.GetDatavolume``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDatavolume`: Datavolume
    fmt.Fprintf(os.Stdout, "Response from `DatavolumesApi.GetDatavolume`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**datavolumeId** | **string** | The id of the datavolume to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDatavolumeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Datavolume**](Datavolume.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDatavolumeAncestors

> Ancestors GetDatavolumeAncestors(ctx, datavolumeId).Execute()

Get the datavolume ancestors

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
    datavolumeId := "71f69d83-ba66-4822-adf5-55ce55efd210" // string | The id of the datavolume to retrieve

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DatavolumesApi.GetDatavolumeAncestors(context.Background(), datavolumeId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DatavolumesApi.GetDatavolumeAncestors``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDatavolumeAncestors`: Ancestors
    fmt.Fprintf(os.Stdout, "Response from `DatavolumesApi.GetDatavolumeAncestors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**datavolumeId** | **string** | The id of the datavolume to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDatavolumeAncestorsRequest struct via the builder pattern


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


## GetDatavolumeSharedScopes

> []SharedScope GetDatavolumeSharedScopes(ctx, datavolumeId).Execute()

Get the datavolume's shared scopes

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
    datavolumeId := "71f69d83-ba66-4822-adf5-55ce55efd210" // string | The id of the datavolume to retrieve

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DatavolumesApi.GetDatavolumeSharedScopes(context.Background(), datavolumeId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DatavolumesApi.GetDatavolumeSharedScopes``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDatavolumeSharedScopes`: []SharedScope
    fmt.Fprintf(os.Stdout, "Response from `DatavolumesApi.GetDatavolumeSharedScopes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**datavolumeId** | **string** | The id of the datavolume to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDatavolumeSharedScopesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]SharedScope**](SharedScope.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDatavolumes

> GetDatavolumes200Response GetDatavolumes(ctx).RequestType(requestType).UsableInProjectId(usableInProjectId).Offset(offset).Limit(limit).SortBy(sortBy).SortOrder(sortOrder).FilterBy(filterBy).Execute()

List datavolumes in permitted scopes



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
    requestType := openapiclient.DatavolumeRequestType("Originated") // DatavolumeRequestType | Which datavolumes would be returned in the response. Originated - datavolumes that are originated in the permitted scopes of the caller. UsableInProject - datavolumes that can be used in a specific project; if you use this value, you must also provide the project ID in the \"usableInProjectId\" query param.
    usableInProjectId := "5" // string | Only when using \"UsableInProject\" requestType; Filter results for only datavolumes that are shared with - or originated in - the project. (optional)
    offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
    limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
    sortBy := "sortBy_example" // string | Sort results by a parameters. (optional)
    sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
    filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DatavolumesApi.GetDatavolumes(context.Background()).RequestType(requestType).UsableInProjectId(usableInProjectId).Offset(offset).Limit(limit).SortBy(sortBy).SortOrder(sortOrder).FilterBy(filterBy).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DatavolumesApi.GetDatavolumes``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDatavolumes`: GetDatavolumes200Response
    fmt.Fprintf(os.Stdout, "Response from `DatavolumesApi.GetDatavolumes`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetDatavolumesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **requestType** | [**DatavolumeRequestType**](DatavolumeRequestType.md) | Which datavolumes would be returned in the response. Originated - datavolumes that are originated in the permitted scopes of the caller. UsableInProject - datavolumes that can be used in a specific project; if you use this value, you must also provide the project ID in the \&quot;usableInProjectId\&quot; query param. | 
 **usableInProjectId** | **string** | Only when using \&quot;UsableInProject\&quot; requestType; Filter results for only datavolumes that are shared with - or originated in - the project. | 
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **sortBy** | **string** | Sort results by a parameters. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 

### Return type

[**GetDatavolumes200Response**](GetDatavolumes200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDatavolumesInternal

> GetDatavolumesInternal200Response GetDatavolumesInternal(ctx).ClusterId(clusterId).Offset(offset).Limit(limit).Execute()

List datavolumes in cluster.



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
    clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
    offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
    limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DatavolumesApi.GetDatavolumesInternal(context.Background()).ClusterId(clusterId).Offset(offset).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DatavolumesApi.GetDatavolumesInternal``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDatavolumesInternal`: GetDatavolumesInternal200Response
    fmt.Fprintf(os.Stdout, "Response from `DatavolumesApi.GetDatavolumesInternal`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetDatavolumesInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**GetDatavolumesInternal200Response**](GetDatavolumesInternal200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PatchDatavolume

> DatavolumeNoSharedScopes PatchDatavolume(ctx, datavolumeId).DatavolumePatchFields(datavolumePatchFields).Execute()

Patch datavolume

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
    datavolumeId := "71f69d83-ba66-4822-adf5-55ce55efd210" // string | The id of the datavolume to retrieve
    datavolumePatchFields := *openapiclient.NewDatavolumePatchFields() // DatavolumePatchFields | Datavolume to update.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DatavolumesApi.PatchDatavolume(context.Background(), datavolumeId).DatavolumePatchFields(datavolumePatchFields).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DatavolumesApi.PatchDatavolume``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `PatchDatavolume`: DatavolumeNoSharedScopes
    fmt.Fprintf(os.Stdout, "Response from `DatavolumesApi.PatchDatavolume`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**datavolumeId** | **string** | The id of the datavolume to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchDatavolumeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **datavolumePatchFields** | [**DatavolumePatchFields**](DatavolumePatchFields.md) | Datavolume to update. | 

### Return type

[**DatavolumeNoSharedScopes**](DatavolumeNoSharedScopes.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PatchDatavolumeSharedScopes

> []SharedScope PatchDatavolumeSharedScopes(ctx, datavolumeId).SharedScopesPatchRequest(sharedScopesPatchRequest).Execute()

Patch the datavolume's shared scopes

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
    datavolumeId := "71f69d83-ba66-4822-adf5-55ce55efd210" // string | The id of the datavolume to retrieve
    sharedScopesPatchRequest := *openapiclient.NewSharedScopesPatchRequest() // SharedScopesPatchRequest | Requested SharedScopes of the datavolume to patch.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DatavolumesApi.PatchDatavolumeSharedScopes(context.Background(), datavolumeId).SharedScopesPatchRequest(sharedScopesPatchRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DatavolumesApi.PatchDatavolumeSharedScopes``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `PatchDatavolumeSharedScopes`: []SharedScope
    fmt.Fprintf(os.Stdout, "Response from `DatavolumesApi.PatchDatavolumeSharedScopes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**datavolumeId** | **string** | The id of the datavolume to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchDatavolumeSharedScopesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **sharedScopesPatchRequest** | [**SharedScopesPatchRequest**](SharedScopesPatchRequest.md) | Requested SharedScopes of the datavolume to patch. | 

### Return type

[**[]SharedScope**](SharedScope.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDatavolumes

> SyncDatavolumes(ctx).DatavolumesSyncRequest(datavolumesSyncRequest).Execute()

Sync the statuses of the datavolumes in the cluster



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
    datavolumesSyncRequest := *openapiclient.NewDatavolumesSyncRequest([]openapiclient.DatavolumeStatusForSync{*openapiclient.NewDatavolumeStatusForSync(openapiclient.DatavolumePhase("Ready"), "71f69d83-ba66-4822-adf5-55ce55efd210")}, "71f69d83-ba66-4822-adf5-55ce55efd210") // DatavolumesSyncRequest | The statuses to sync. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DatavolumesApi.SyncDatavolumes(context.Background()).DatavolumesSyncRequest(datavolumesSyncRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DatavolumesApi.SyncDatavolumes``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncDatavolumesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **datavolumesSyncRequest** | [**DatavolumesSyncRequest**](DatavolumesSyncRequest.md) | The statuses to sync. | 

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


## UpdateDatavolumeStatus

> UpdateDatavolumeStatus(ctx, datavolumeId).DatavolumeStatusInternal(datavolumeStatusInternal).Execute()

Update the status of the datavolume form the cluster

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
    datavolumeId := "71f69d83-ba66-4822-adf5-55ce55efd210" // string | The id of the datavolume to retrieve
    datavolumeStatusInternal := *openapiclient.NewDatavolumeStatusInternal(openapiclient.DatavolumePhase("Ready")) // DatavolumeStatusInternal | The status of the datavolume in the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DatavolumesApi.UpdateDatavolumeStatus(context.Background(), datavolumeId).DatavolumeStatusInternal(datavolumeStatusInternal).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DatavolumesApi.UpdateDatavolumeStatus``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**datavolumeId** | **string** | The id of the datavolume to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDatavolumeStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **datavolumeStatusInternal** | [**DatavolumeStatusInternal**](DatavolumeStatusInternal.md) | The status of the datavolume in the cluster. | 

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

