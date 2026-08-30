# \DatasourcesAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateDatasourceAsset**](DatasourcesAPI.md#CreateDatasourceAsset) | **Post** /api/v1/asset/datasource | Placeholder.
[**ListDatasourceAssets**](DatasourcesAPI.md#ListDatasourceAssets) | **Get** /api/v1/asset/datasource | List datasource assets.



## CreateDatasourceAsset

> CreateDatasourceAsset(ctx).AssetMeta(assetMeta).Execute()

Placeholder.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetMeta := *openapiclient.NewAssetMeta("my-asset", openapiclient.Scope("system"), "Id_example", openapiclient.AssetKind("compute"), "CreatedBy_example", time.Now(), "UpdatedBy_example", time.Now()) // AssetMeta | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DatasourcesAPI.CreateDatasourceAsset(context.Background()).AssetMeta(assetMeta).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DatasourcesAPI.CreateDatasourceAsset``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateDatasourceAssetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **assetMeta** | [**AssetMeta**](AssetMeta.md) |  | 

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


## ListDatasourceAssets

> DatasourceListResponse ListDatasourceAssets(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).StatusInfo(statusInfo).AssetIds(assetIds).ComplyToReplicaType(complyToReplicaType).DataVolumes(dataVolumes).IncludeDescendants(includeDescendants).Type_(type_).Execute()

List datasource assets.



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
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)
	assetIds := "dbf4767e-2fa1-43b0-97a2-7c0cecda180b,550e8400-e29b-41d4-a716-44665544000a" // string | Filter results by the ids of the assets. Provided value should be a comma separated string of UUIDs. (optional)
	complyToReplicaType := "complyToReplicaType_example" // string | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. (optional)
	dataVolumes := true // bool | Whether the query should include data volumes (optional) (default to false)
	includeDescendants := true // bool | Indicates whether to include assets from all descendant scopes under the specified scope. When set to true, and a specific scope identifier (for example, clusterId) is provided, the query also returns assets from all nested scopes within it (for example, departments and projects).  Mutually exclusive with scope. (optional)
	type_ := "type__example" // string | Filter datasources by type. If provided, only datasources of the specified type will be returned. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DatasourcesAPI.ListDatasourceAssets(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).ComplyToProject(complyToProject).ComplyToWorkloadType(complyToWorkloadType).StatusInfo(statusInfo).AssetIds(assetIds).ComplyToReplicaType(complyToReplicaType).DataVolumes(dataVolumes).IncludeDescendants(includeDescendants).Type_(type_).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DatasourcesAPI.ListDatasourceAssets``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListDatasourceAssets`: DatasourceListResponse
	fmt.Fprintf(os.Stdout, "Response from `DatasourcesAPI.ListDatasourceAssets`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListDatasourceAssetsRequest struct via the builder pattern


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
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 
 **assetIds** | **string** | Filter results by the ids of the assets. Provided value should be a comma separated string of UUIDs. | 
 **complyToReplicaType** | **string** | Include workload creation compliance information of an asset, for a given replica type, as part of the response. To check compliance, you need to provide both project id and workload type. For distributed, replica type should be provided as well. | 
 **dataVolumes** | **bool** | Whether the query should include data volumes | [default to false]
 **includeDescendants** | **bool** | Indicates whether to include assets from all descendant scopes under the specified scope. When set to true, and a specific scope identifier (for example, clusterId) is provided, the query also returns assets from all nested scopes within it (for example, departments and projects).  Mutually exclusive with scope. | 
 **type_** | **string** | Filter datasources by type. If provided, only datasources of the specified type will be returned. | 

### Return type

[**DatasourceListResponse**](DatasourceListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

