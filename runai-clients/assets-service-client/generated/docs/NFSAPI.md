# \NFSAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateNfsAsset**](NFSAPI.md#CreateNfsAsset) | **Post** /api/v1/asset/datasource/nfs | Create an NFS asset.
[**DeleteNfsAssetById**](NFSAPI.md#DeleteNfsAssetById) | **Delete** /api/v1/asset/datasource/nfs/{AssetId} | Delete an NFS asset.
[**GetNfsAssetById**](NFSAPI.md#GetNfsAssetById) | **Get** /api/v1/asset/datasource/nfs/{AssetId} | Get an NFS asset.
[**ListNfsAssets**](NFSAPI.md#ListNfsAssets) | **Get** /api/v1/asset/datasource/nfs | List NFS assets.
[**UpdateNfsAssetById**](NFSAPI.md#UpdateNfsAssetById) | **Put** /api/v1/asset/datasource/nfs/{AssetId} | Update an NFS asset.



## CreateNfsAsset

> S3Asset CreateNfsAsset(ctx).NFSCreationRequest(nFSCreationRequest).Execute()

Create an NFS asset.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	nFSCreationRequest := *openapiclient.NewNFSCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("system")), *openapiclient.NewNFSAssetSpec()) // NFSCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NFSAPI.CreateNfsAsset(context.Background()).NFSCreationRequest(nFSCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NFSAPI.CreateNfsAsset``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateNfsAsset`: S3Asset
	fmt.Fprintf(os.Stdout, "Response from `NFSAPI.CreateNfsAsset`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateNfsAssetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **nFSCreationRequest** | [**NFSCreationRequest**](NFSCreationRequest.md) |  | 

### Return type

[**S3Asset**](S3Asset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteNfsAssetById

> HttpResponse DeleteNfsAssetById(ctx, assetId).Execute()

Delete an NFS asset.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NFSAPI.DeleteNfsAssetById(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NFSAPI.DeleteNfsAssetById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteNfsAssetById`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `NFSAPI.DeleteNfsAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteNfsAssetByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## GetNfsAssetById

> NFSAsset GetNfsAssetById(ctx, assetId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).ComplyToReplicaType(complyToReplicaType).Execute()

Get an NFS asset.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	complyToProject := int32(56) // int32 | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
	complyToWorkloadType := "complyToWorkloadType_example" // string | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
	complyToReplicaType := "complyToReplicaType_example" // string | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NFSAPI.GetNfsAssetById(context.Background(), assetId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).ComplyToReplicaType(complyToReplicaType).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NFSAPI.GetNfsAssetById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNfsAssetById`: NFSAsset
	fmt.Fprintf(os.Stdout, "Response from `NFSAPI.GetNfsAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNfsAssetByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **complyToProject** | **int32** | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **complyToWorkloadType** | **string** | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **complyToReplicaType** | **string** | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. | 

### Return type

[**NFSAsset**](NFSAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListNfsAssets

> NFSListResponse ListNfsAssets(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).AssetIds(assetIds).ComplyToReplicaType(complyToReplicaType).Execute()

List NFS assets.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	name := "name_example" // string | Filter results by name. (optional)
	scope := "scope_example" // string | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. (optional)
	projectId := int32(56) // int32 | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. (optional)
	departmentId := "1" // string | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. (optional)
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	complyToProject := int32(56) // int32 | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
	complyToWorkloadType := "complyToWorkloadType_example" // string | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. (optional)
	assetIds := "dbf4767e-2fa1-43b0-97a2-7c0cecda180b,550e8400-e29b-41d4-a716-44665544000a" // string | Filter results by the ids of the assets. Provided value should be a comma separated string of UUIDs. (optional)
	complyToReplicaType := "complyToReplicaType_example" // string | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NFSAPI.ListNfsAssets(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).AssetIds(assetIds).ComplyToReplicaType(complyToReplicaType).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NFSAPI.ListNfsAssets``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListNfsAssets`: NFSListResponse
	fmt.Fprintf(os.Stdout, "Response from `NFSAPI.ListNfsAssets`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListNfsAssetsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. | 
 **projectId** | **int32** | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. | 
 **departmentId** | **string** | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. | 
 **clusterId** | **string** | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. | 
 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **complyToProject** | **int32** | Include workload creation compliance information of an asset, for a given project, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **complyToWorkloadType** | **string** | Include workload creation compliance information of an asset, for a given workload type, as part of the response. To check compliance, you need to provide both project id and workload type. | 
 **assetIds** | **string** | Filter results by the ids of the assets. Provided value should be a comma separated string of UUIDs. | 
 **complyToReplicaType** | **string** | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. | 

### Return type

[**NFSListResponse**](NFSListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateNfsAssetById

> NFSAsset UpdateNfsAssetById(ctx, assetId).NFSUpdateRequest(nFSUpdateRequest).Execute()

Update an NFS asset.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	nFSUpdateRequest := *openapiclient.NewNFSUpdateRequest(*openapiclient.NewAssetUpdateRequest("my-asset"), *openapiclient.NewNFSAssetSpec()) // NFSUpdateRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NFSAPI.UpdateNfsAssetById(context.Background(), assetId).NFSUpdateRequest(nFSUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NFSAPI.UpdateNfsAssetById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateNfsAssetById`: NFSAsset
	fmt.Fprintf(os.Stdout, "Response from `NFSAPI.UpdateNfsAssetById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNfsAssetByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nFSUpdateRequest** | [**NFSUpdateRequest**](NFSUpdateRequest.md) |  | 

### Return type

[**NFSAsset**](NFSAsset.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

