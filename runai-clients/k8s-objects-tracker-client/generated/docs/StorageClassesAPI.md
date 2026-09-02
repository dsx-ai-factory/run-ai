# \StorageClassesAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateStorageClass**](StorageClassesAPI.md#CreateStorageClass) | **Post** /api/v1/storage-classes | create a Storage Class
[**DeleteStorageClass**](StorageClassesAPI.md#DeleteStorageClass) | **Delete** /api/v1/storage-classes | delete a Storage Class
[**GetStorageClasses**](StorageClassesAPI.md#GetStorageClasses) | **Get** /api/v1/storage-classes | get a Storage Class/Classes for a given cluster
[**GetStorageClassesV2**](StorageClassesAPI.md#GetStorageClassesV2) | **Get** /api/v2/storage-classes | get a Storage Class/Classes for a given cluster
[**SyncStorageClass**](StorageClassesAPI.md#SyncStorageClass) | **Post** /api/v1/storage-classes/sync | Sync cluster Storage Classes
[**UpdateStorageClass**](StorageClassesAPI.md#UpdateStorageClass) | **Put** /api/v1/storage-classes | update a Storage Class
[**V1CreateStorageClass**](StorageClassesAPI.md#V1CreateStorageClass) | **Post** /v1/k8s/clusters/{uuid}/storage-classes | Create a storageClass.
[**V1DeleteStorageClass**](StorageClassesAPI.md#V1DeleteStorageClass) | **Delete** /v1/k8s/clusters/{uuid}/storage-classes/{storage_class_name} | Delete a storageClass.
[**V1GetStorageClasses**](StorageClassesAPI.md#V1GetStorageClasses) | **Get** /v1/k8s/clusters/{uuid}/storage-classes | Get all storageClasses from a cluster.
[**V1SyncStorageClass**](StorageClassesAPI.md#V1SyncStorageClass) | **Post** /v1/k8s/clusters/{uuid}/storage-classes/sync | Sync storageClass to a cluster.
[**V1UpdateStorageClass**](StorageClassesAPI.md#V1UpdateStorageClass) | **Put** /v1/k8s/clusters/{uuid}/storage-classes/{storage_class_name} | Update a storageClass.



## CreateStorageClass

> StorageClass CreateStorageClass(ctx).StorageClass(storageClass).Execute()

create a Storage Class

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	storageClass := *openapiclient.NewStorageClass("my-storage-class") // StorageClass | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassesAPI.CreateStorageClass(context.Background()).StorageClass(storageClass).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesAPI.CreateStorageClass``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateStorageClass`: StorageClass
	fmt.Fprintf(os.Stdout, "Response from `StorageClassesAPI.CreateStorageClass`: %v\n", resp)
}
```

### Path Parameters



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

> DeleteStorageClass(ctx).ClusterId(clusterId).Name(name).Execute()

delete a Storage Class

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	name := "name_example" // string | filter by name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.StorageClassesAPI.DeleteStorageClass(context.Background()).ClusterId(clusterId).Name(name).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesAPI.DeleteStorageClass``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteStorageClassRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **name** | **string** | filter by name | 

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

> StorageClasses GetStorageClasses(ctx).ClusterId(clusterId).Name(name).IncludeNone(includeNone).Execute()

get a Storage Class/Classes for a given cluster

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	name := "name_example" // string | filter by name (optional)
	includeNone := true // bool | Include runai-none storage class, which is used for static provisioning of PVC. In static provisioning storage is omitted. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassesAPI.GetStorageClasses(context.Background()).ClusterId(clusterId).Name(name).IncludeNone(includeNone).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesAPI.GetStorageClasses``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetStorageClasses`: StorageClasses
	fmt.Fprintf(os.Stdout, "Response from `StorageClassesAPI.GetStorageClasses`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetStorageClassesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **name** | **string** | filter by name | 
 **includeNone** | **bool** | Include runai-none storage class, which is used for static provisioning of PVC. In static provisioning storage is omitted. | 

### Return type

[**StorageClasses**](StorageClasses.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetStorageClassesV2

> StorageClassesV2 GetStorageClassesV2(ctx).ClusterId(clusterId).Name(name).IncludeNone(includeNone).Execute()

get a Storage Class/Classes for a given cluster

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	name := "name_example" // string | filter by name (optional)
	includeNone := true // bool | Include runai-none storage class, which is used for static provisioning of PVC. In static provisioning storage is omitted. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassesAPI.GetStorageClassesV2(context.Background()).ClusterId(clusterId).Name(name).IncludeNone(includeNone).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesAPI.GetStorageClassesV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetStorageClassesV2`: StorageClassesV2
	fmt.Fprintf(os.Stdout, "Response from `StorageClassesAPI.GetStorageClassesV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetStorageClassesV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **name** | **string** | filter by name | 
 **includeNone** | **bool** | Include runai-none storage class, which is used for static provisioning of PVC. In static provisioning storage is omitted. | 

### Return type

[**StorageClassesV2**](StorageClassesV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncStorageClass

> StorageClasses SyncStorageClass(ctx).ClusterId(clusterId).StorageClasses(storageClasses).Execute()

Sync cluster Storage Classes



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	storageClasses := *openapiclient.NewStorageClasses([]openapiclient.StorageClass{*openapiclient.NewStorageClass("my-storage-class")}) // StorageClasses | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassesAPI.SyncStorageClass(context.Background()).ClusterId(clusterId).StorageClasses(storageClasses).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesAPI.SyncStorageClass``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncStorageClass`: StorageClasses
	fmt.Fprintf(os.Stdout, "Response from `StorageClassesAPI.SyncStorageClass`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncStorageClassRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **storageClasses** | [**StorageClasses**](StorageClasses.md) |  | 

### Return type

[**StorageClasses**](StorageClasses.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateStorageClass

> StorageClass UpdateStorageClass(ctx).StorageClass(storageClass).Execute()

update a Storage Class

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	storageClass := *openapiclient.NewStorageClass("my-storage-class") // StorageClass | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassesAPI.UpdateStorageClass(context.Background()).StorageClass(storageClass).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesAPI.UpdateStorageClass``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateStorageClass`: StorageClass
	fmt.Fprintf(os.Stdout, "Response from `StorageClassesAPI.UpdateStorageClass`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateStorageClassRequest struct via the builder pattern


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


## V1CreateStorageClass

> V1StorageClass V1CreateStorageClass(ctx, uuid).V1StorageClass(v1StorageClass).Execute()

Create a storageClass.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
	v1StorageClass := *openapiclient.NewV1StorageClass("StorageClassName_example") // V1StorageClass | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassesAPI.V1CreateStorageClass(context.Background(), uuid).V1StorageClass(v1StorageClass).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesAPI.V1CreateStorageClass``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `V1CreateStorageClass`: V1StorageClass
	fmt.Fprintf(os.Stdout, "Response from `StorageClassesAPI.V1CreateStorageClass`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1CreateStorageClassRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **v1StorageClass** | [**V1StorageClass**](V1StorageClass.md) |  | 

### Return type

[**V1StorageClass**](V1StorageClass.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1DeleteStorageClass

> V1DeleteStorageClass(ctx, uuid, storageClassName).Execute()

Delete a storageClass.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
	storageClassName := "storageClassName_example" // string | Unique identifier of the storageClass

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.StorageClassesAPI.V1DeleteStorageClass(context.Background(), uuid, storageClassName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesAPI.V1DeleteStorageClass``: %v\n", err)
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

Other parameters are passed through a pointer to a apiV1DeleteStorageClassRequest struct via the builder pattern


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


## V1GetStorageClasses

> V1GetStorageClasses200Response V1GetStorageClasses(ctx, uuid).IncludeNone(includeNone).Execute()

Get all storageClasses from a cluster.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
	includeNone := true // bool | Include runai-none storage class to be able to create PVCs without a storage class (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassesAPI.V1GetStorageClasses(context.Background(), uuid).IncludeNone(includeNone).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesAPI.V1GetStorageClasses``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `V1GetStorageClasses`: V1GetStorageClasses200Response
	fmt.Fprintf(os.Stdout, "Response from `StorageClassesAPI.V1GetStorageClasses`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1GetStorageClassesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **includeNone** | **bool** | Include runai-none storage class to be able to create PVCs without a storage class | 

### Return type

[**V1GetStorageClasses200Response**](V1GetStorageClasses200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1SyncStorageClass

> V1StorageClass V1SyncStorageClass(ctx, uuid).V1StorageClass(v1StorageClass).Execute()

Sync storageClass to a cluster.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
	v1StorageClass := []openapiclient.V1StorageClass{*openapiclient.NewV1StorageClass("StorageClassName_example")} // []V1StorageClass | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassesAPI.V1SyncStorageClass(context.Background(), uuid).V1StorageClass(v1StorageClass).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesAPI.V1SyncStorageClass``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `V1SyncStorageClass`: V1StorageClass
	fmt.Fprintf(os.Stdout, "Response from `StorageClassesAPI.V1SyncStorageClass`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1SyncStorageClassRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **v1StorageClass** | [**[]V1StorageClass**](V1StorageClass.md) |  | 

### Return type

[**V1StorageClass**](V1StorageClass.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## V1UpdateStorageClass

> []V1StorageClass V1UpdateStorageClass(ctx, uuid, storageClassName).V1StorageClass(v1StorageClass).Execute()

Update a storageClass.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	uuid := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the cluster
	storageClassName := "storageClassName_example" // string | Unique identifier of the storageClass
	v1StorageClass := *openapiclient.NewV1StorageClass("StorageClassName_example") // V1StorageClass | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassesAPI.V1UpdateStorageClass(context.Background(), uuid, storageClassName).V1StorageClass(v1StorageClass).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassesAPI.V1UpdateStorageClass``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `V1UpdateStorageClass`: []V1StorageClass
	fmt.Fprintf(os.Stdout, "Response from `StorageClassesAPI.V1UpdateStorageClass`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**uuid** | **string** | Unique identifier of the cluster | 
**storageClassName** | **string** | Unique identifier of the storageClass | 

### Other Parameters

Other parameters are passed through a pointer to a apiV1UpdateStorageClassRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **v1StorageClass** | [**V1StorageClass**](V1StorageClass.md) |  | 

### Return type

[**[]V1StorageClass**](V1StorageClass.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

