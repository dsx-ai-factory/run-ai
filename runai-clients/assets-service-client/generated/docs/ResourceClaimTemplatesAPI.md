# \ResourceClaimTemplatesAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateRct**](ResourceClaimTemplatesAPI.md#CreateRct) | **Post** /api/v1/k8s-objects/resource-claim-templates | Create resource claim template.
[**DeleteRctById**](ResourceClaimTemplatesAPI.md#DeleteRctById) | **Delete** /api/v1/k8s-objects/resource-claim-templates/{AssetId} | Delete a resource claim template.
[**GetRctById**](ResourceClaimTemplatesAPI.md#GetRctById) | **Get** /api/v1/k8s-objects/resource-claim-templates/{AssetId} | Retrieve a resource claim template.
[**GetRctWorkloads**](ResourceClaimTemplatesAPI.md#GetRctWorkloads) | **Get** /api/v1/k8s-objects/resource-claim-templates/{AssetId}/workloads | List the workloads using a resource claim template.
[**ListRcts**](ResourceClaimTemplatesAPI.md#ListRcts) | **Get** /api/v1/k8s-objects/resource-claim-templates | List resource claim templates.
[**UpdateRctById**](ResourceClaimTemplatesAPI.md#UpdateRctById) | **Put** /api/v1/k8s-objects/resource-claim-templates/{AssetId} | Update a resource claim template.



## CreateRct

> Rct CreateRct(ctx).RctCreationRequest(rctCreationRequest).Execute()

Create resource claim template.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	rctCreationRequest := *openapiclient.NewRctCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("system")), *openapiclient.NewRctSpec(map[string]interface{}{"key": interface{}(123)})) // RctCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.CreateRct(context.Background()).RctCreationRequest(rctCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.CreateRct``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateRct`: Rct
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.CreateRct`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateRctRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **rctCreationRequest** | [**RctCreationRequest**](RctCreationRequest.md) |  | 

### Return type

[**Rct**](Rct.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRctById

> HttpResponse DeleteRctById(ctx, assetId).Execute()

Delete a resource claim template.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.DeleteRctById(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.DeleteRctById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteRctById`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.DeleteRctById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRctByIdRequest struct via the builder pattern


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


## GetRctById

> Rct GetRctById(ctx, assetId).StatusInfo(statusInfo).Execute()

Retrieve a resource claim template.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.GetRctById(context.Background(), assetId).StatusInfo(statusInfo).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.GetRctById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRctById`: Rct
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.GetRctById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRctByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 

### Return type

[**Rct**](Rct.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRctWorkloads

> ResourceClaimTemplateWorkloadsResponse GetRctWorkloads(ctx, assetId).Offset(offset).Limit(limit).Execute()

List the workloads using a resource claim template.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.GetRctWorkloads(context.Background(), assetId).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.GetRctWorkloads``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRctWorkloads`: ResourceClaimTemplateWorkloadsResponse
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.GetRctWorkloads`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRctWorkloadsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**ResourceClaimTemplateWorkloadsResponse**](ResourceClaimTemplateWorkloadsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListRcts

> RctListResponse ListRcts(ctx).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).IncludeAncestors(includeAncestors).IncludeDescendants(includeDescendants).StatusInfo(statusInfo).Execute()

List resource claim templates.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	sortBy := "sortBy_example" // string | Sort results by parameter. (optional)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. At most one scope anchor (clusterId, departmentId or projectId) may be provided. The anchor is combined with includeAncestors / includeDescendants to determine which scopes of the hierarchy are returned. (optional)
	includeAncestors := true // bool | When a scope anchor (clusterId, departmentId or projectId) is provided, set includeAncestors=true to also include resource claim templates from the anchor's ancestor scopes (for example, for a project: its parent department and cluster, as well as tenant and system scoped templates). Defaults to true to preserve the backward-compatible scope-usability behavior. (optional) (default to true)
	includeDescendants := true // bool | Indicates whether to include assets from all descendant scopes under the specified scope. When set to true, and a specific scope identifier (for example, clusterId) is provided, the query also returns assets from all nested scopes within it (for example, departments and projects).  Mutually exclusive with scope. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.ListRcts(context.Background()).Offset(offset).Limit(limit).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).IncludeAncestors(includeAncestors).IncludeDescendants(includeDescendants).StatusInfo(statusInfo).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.ListRcts``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListRcts`: RctListResponse
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.ListRcts`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListRctsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **sortBy** | **string** | Sort results by parameter. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contain, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. At most one scope anchor (clusterId, departmentId or projectId) may be provided. The anchor is combined with includeAncestors / includeDescendants to determine which scopes of the hierarchy are returned. | 
 **includeAncestors** | **bool** | When a scope anchor (clusterId, departmentId or projectId) is provided, set includeAncestors&#x3D;true to also include resource claim templates from the anchor&#39;s ancestor scopes (for example, for a project: its parent department and cluster, as well as tenant and system scoped templates). Defaults to true to preserve the backward-compatible scope-usability behavior. | [default to true]
 **includeDescendants** | **bool** | Indicates whether to include assets from all descendant scopes under the specified scope. When set to true, and a specific scope identifier (for example, clusterId) is provided, the query also returns assets from all nested scopes within it (for example, departments and projects).  Mutually exclusive with scope. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 

### Return type

[**RctListResponse**](RctListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateRctById

> Rct UpdateRctById(ctx, assetId).RctUpdateRequest(rctUpdateRequest).Execute()

Update a resource claim template.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	rctUpdateRequest := *openapiclient.NewRctUpdateRequest(*openapiclient.NewAssetUpdateRequest("my-asset"), *openapiclient.NewRctSpec(map[string]interface{}{"key": interface{}(123)})) // RctUpdateRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.UpdateRctById(context.Background(), assetId).RctUpdateRequest(rctUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.UpdateRctById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateRctById`: Rct
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.UpdateRctById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateRctByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **rctUpdateRequest** | [**RctUpdateRequest**](RctUpdateRequest.md) |  | 

### Return type

[**Rct**](Rct.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

