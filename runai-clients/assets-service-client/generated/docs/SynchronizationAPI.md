# \SynchronizationAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CredentialsSyncUpdates**](SynchronizationAPI.md#CredentialsSyncUpdates) | **Post** /api/v1/asset/credentials/updates | Allow cluster to request a list of updates that needs to be synced to the cluster and to send success/failure about the result of previous sync.
[**DatasourceSyncUpdates**](SynchronizationAPI.md#DatasourceSyncUpdates) | **Post** /api/v1/asset/datasource/updates | Allow cluster to request a list of updates that needs to be synced to the cluster and to send success/failure about the result of previous sync.
[**ListAssetsById**](SynchronizationAPI.md#ListAssetsById) | **Get** /api/v1/asset | List assets by their ids
[**SyncAssetsResourcesIds**](SynchronizationAPI.md#SyncAssetsResourcesIds) | **Post** /api/v1/asset/sync/asset-resources-ids | Sync asset resources ids.
[**SyncValidate**](SynchronizationAPI.md#SyncValidate) | **Post** /api/v1/asset/sync/validate | Validate assets sync.



## CredentialsSyncUpdates

> CredentialsUpdatesResp CredentialsSyncUpdates(ctx).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()

Allow cluster to request a list of updates that needs to be synced to the cluster and to send success/failure about the result of previous sync.

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
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	after := int64(123) // int64 | search for updates with update counter greater than the value passed for this parameter (optional)
	clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.CredentialsSyncUpdates(context.Background()).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.CredentialsSyncUpdates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CredentialsSyncUpdates`: CredentialsUpdatesResp
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.CredentialsSyncUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCredentialsSyncUpdatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **after** | **int64** | search for updates with update counter greater than the value passed for this parameter | 
 **clusterSyncReport** | [**ClusterSyncReport**](ClusterSyncReport.md) |  | 

### Return type

[**CredentialsUpdatesResp**](CredentialsUpdatesResp.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DatasourceSyncUpdates

> DatasourceUpdatesResp DatasourceSyncUpdates(ctx).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()

Allow cluster to request a list of updates that needs to be synced to the cluster and to send success/failure about the result of previous sync.

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
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	after := int64(123) // int64 | search for updates with update counter greater than the value passed for this parameter (optional)
	clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.DatasourceSyncUpdates(context.Background()).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.DatasourceSyncUpdates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DatasourceSyncUpdates`: DatasourceUpdatesResp
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.DatasourceSyncUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDatasourceSyncUpdatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **after** | **int64** | search for updates with update counter greater than the value passed for this parameter | 
 **clusterSyncReport** | [**ClusterSyncReport**](ClusterSyncReport.md) |  | 

### Return type

[**DatasourceUpdatesResp**](DatasourceUpdatesResp.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
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
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	id := "id_example" // string | comma seperated list of asset ids

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.ListAssetsById(context.Background()).Id(id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.ListAssetsById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListAssetsById`: AssetListByIdsResponse
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.ListAssetsById`: %v\n", resp)
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


## SyncAssetsResourcesIds

> SyncAssetsResourcesIds(ctx).AssetResourcesIdsSyncRequest(assetResourcesIdsSyncRequest).Execute()

Sync asset resources ids.



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
	assetResourcesIdsSyncRequest := *openapiclient.NewAssetResourcesIdsSyncRequest([]openapiclient.AssetResourcesIdsItem{*openapiclient.NewAssetResourcesIdsItem("ResourceId_example", "AssetId_example")}) // AssetResourcesIdsSyncRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SynchronizationAPI.SyncAssetsResourcesIds(context.Background()).AssetResourcesIdsSyncRequest(assetResourcesIdsSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.SyncAssetsResourcesIds``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncAssetsResourcesIdsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **assetResourcesIdsSyncRequest** | [**AssetResourcesIdsSyncRequest**](AssetResourcesIdsSyncRequest.md) |  | 

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


## SyncValidate

> ValidateSyncResponse SyncValidate(ctx).ValidateSyncRequest(validateSyncRequest).Execute()

Validate assets sync.



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
	validateSyncRequest := *openapiclient.NewValidateSyncRequest("ClusterId_example", "ProjectId_example", openapiclient.AssetKind("compute"), []string{"AssetNames_example"}) // ValidateSyncRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.SyncValidate(context.Background()).ValidateSyncRequest(validateSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.SyncValidate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncValidate`: ValidateSyncResponse
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.SyncValidate`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncValidateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateSyncRequest** | [**ValidateSyncRequest**](ValidateSyncRequest.md) |  | 

### Return type

[**ValidateSyncResponse**](ValidateSyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

