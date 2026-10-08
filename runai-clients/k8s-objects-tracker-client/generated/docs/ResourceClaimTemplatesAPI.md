# \ResourceClaimTemplatesAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateResourceClaimTemplate**](ResourceClaimTemplatesAPI.md#CreateResourceClaimTemplate) | **Post** /api/v1/resource-claim-templates | create a Resource Claim Template
[**DeleteResourceClaimTemplate**](ResourceClaimTemplatesAPI.md#DeleteResourceClaimTemplate) | **Delete** /api/v1/resource-claim-templates | delete a Resource Claim Template
[**ListResourceClaimTemplates**](ResourceClaimTemplatesAPI.md#ListResourceClaimTemplates) | **Get** /api/v1/resource-claim-templates | list details of ResourceClaimTemplates
[**SyncResourceClaimTemplates**](ResourceClaimTemplatesAPI.md#SyncResourceClaimTemplates) | **Post** /api/v1/resource-claim-templates/sync | sync the full set of ResourceClaimTemplates observed in a cluster
[**UpdateResourceClaimTemplate**](ResourceClaimTemplatesAPI.md#UpdateResourceClaimTemplate) | **Put** /api/v1/resource-claim-templates | update a Resource Claim Template



## CreateResourceClaimTemplate

> ResourceClaimTemplate CreateResourceClaimTemplate(ctx).ClusterId(clusterId).ResourceClaimTemplate(resourceClaimTemplate).Execute()

create a Resource Claim Template

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
	resourceClaimTemplate := *openapiclient.NewResourceClaimTemplate("rct-gpu-1", "runai", openapiclient.ResourceScopeType("cluster"), "ScopeId_example") // ResourceClaimTemplate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.CreateResourceClaimTemplate(context.Background()).ClusterId(clusterId).ResourceClaimTemplate(resourceClaimTemplate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.CreateResourceClaimTemplate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateResourceClaimTemplate`: ResourceClaimTemplate
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.CreateResourceClaimTemplate`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateResourceClaimTemplateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **resourceClaimTemplate** | [**ResourceClaimTemplate**](ResourceClaimTemplate.md) |  | 

### Return type

[**ResourceClaimTemplate**](ResourceClaimTemplate.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteResourceClaimTemplate

> DeleteResourceClaimTemplate(ctx).ClusterId(clusterId).ScopeType(scopeType).ScopeId(scopeId).Name(name).Execute()

delete a Resource Claim Template

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
	scopeType := openapiclient.ResourceScopeType("cluster") // ResourceScopeType | The resource scope type to filter by.
	scopeId := "2" // string | The scope resource id that identifies the row.
	name := "name_example" // string | filter by name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ResourceClaimTemplatesAPI.DeleteResourceClaimTemplate(context.Background()).ClusterId(clusterId).ScopeType(scopeType).ScopeId(scopeId).Name(name).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.DeleteResourceClaimTemplate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteResourceClaimTemplateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **scopeType** | [**ResourceScopeType**](ResourceScopeType.md) | The resource scope type to filter by. | 
 **scopeId** | **string** | The scope resource id that identifies the row. | 
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


## ListResourceClaimTemplates

> ResourceClaimTemplates ListResourceClaimTemplates(ctx).ScopeType(scopeType).ScopeId(scopeId).Name(name).ClusterId(clusterId).Origin(origin).Execute()

list details of ResourceClaimTemplates

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
	scopeType := "project" // string | The type of resource we want to filter by. (optional)
	scopeId := "scopeId_example" // string | filter by scope id (optional)
	name := "name_example" // string | filter by name (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	origin := "origin_example" // string | filter by the origin of the resource. by default, we have to list only customer existing resources (to preserve backward compatibility) to list both existing pvc and assets related pvcs use 'all'. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.ListResourceClaimTemplates(context.Background()).ScopeType(scopeType).ScopeId(scopeId).Name(name).ClusterId(clusterId).Origin(origin).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.ListResourceClaimTemplates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListResourceClaimTemplates`: ResourceClaimTemplates
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.ListResourceClaimTemplates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListResourceClaimTemplatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scopeType** | **string** | The type of resource we want to filter by. | 
 **scopeId** | **string** | filter by scope id | 
 **name** | **string** | filter by name | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **origin** | **string** | filter by the origin of the resource. by default, we have to list only customer existing resources (to preserve backward compatibility) to list both existing pvc and assets related pvcs use &#39;all&#39;. | 

### Return type

[**ResourceClaimTemplates**](ResourceClaimTemplates.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncResourceClaimTemplates

> ResourceClaimTemplates SyncResourceClaimTemplates(ctx).ClusterId(clusterId).ResourceClaimTemplates(resourceClaimTemplates).Execute()

sync the full set of ResourceClaimTemplates observed in a cluster



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
	resourceClaimTemplates := *openapiclient.NewResourceClaimTemplates([]openapiclient.ResourceClaimTemplate{*openapiclient.NewResourceClaimTemplate("rct-gpu-1", "runai", openapiclient.ResourceScopeType("cluster"), "ScopeId_example")}) // ResourceClaimTemplates | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.SyncResourceClaimTemplates(context.Background()).ClusterId(clusterId).ResourceClaimTemplates(resourceClaimTemplates).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.SyncResourceClaimTemplates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncResourceClaimTemplates`: ResourceClaimTemplates
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.SyncResourceClaimTemplates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncResourceClaimTemplatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **resourceClaimTemplates** | [**ResourceClaimTemplates**](ResourceClaimTemplates.md) |  | 

### Return type

[**ResourceClaimTemplates**](ResourceClaimTemplates.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateResourceClaimTemplate

> ResourceClaimTemplate UpdateResourceClaimTemplate(ctx).ClusterId(clusterId).ResourceClaimTemplate(resourceClaimTemplate).Execute()

update a Resource Claim Template

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
	resourceClaimTemplate := *openapiclient.NewResourceClaimTemplate("rct-gpu-1", "runai", openapiclient.ResourceScopeType("cluster"), "ScopeId_example") // ResourceClaimTemplate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.UpdateResourceClaimTemplate(context.Background()).ClusterId(clusterId).ResourceClaimTemplate(resourceClaimTemplate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.UpdateResourceClaimTemplate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateResourceClaimTemplate`: ResourceClaimTemplate
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.UpdateResourceClaimTemplate`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateResourceClaimTemplateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **resourceClaimTemplate** | [**ResourceClaimTemplate**](ResourceClaimTemplate.md) |  | 

### Return type

[**ResourceClaimTemplate**](ResourceClaimTemplate.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

