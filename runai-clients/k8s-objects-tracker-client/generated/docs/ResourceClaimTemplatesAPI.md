# \ResourceClaimTemplatesAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ListResourceClaimTemplates**](ResourceClaimTemplatesAPI.md#ListResourceClaimTemplates) | **Get** /api/v1/resource-claim-templates | list details of ResourceClaimTemplates
[**SyncClusterWideResourceClaimTemplates**](ResourceClaimTemplatesAPI.md#SyncClusterWideResourceClaimTemplates) | **Post** /api/v1/resource-claim-templates/{clusterUuid} | sync the details of cluster-wide ResourceClaimTemplates to the control plane
[**SyncDepartmentWideResourceClaimTemplates**](ResourceClaimTemplatesAPI.md#SyncDepartmentWideResourceClaimTemplates) | **Post** /api/v1/resource-claim-templates/{clusterUuid}/departments/{departmentId} | sync the details of department-wide ResourceClaimTemplates to the control plane
[**SyncProjectResourceClaimTemplates**](ResourceClaimTemplatesAPI.md#SyncProjectResourceClaimTemplates) | **Post** /api/v1/resource-claim-templates/{clusterUuid}/namespaces/{namespace} | sync the details of project ResourceClaimTemplates to the control plane



## ListResourceClaimTemplates

> ResourceClaimTemplates ListResourceClaimTemplates(ctx).ScopeType(scopeType).ScopeId(scopeId).Name(name).ClusterId(clusterId).Origin(origin).IncludeAncestors(includeAncestors).IncludeDescendants(includeDescendants).Execute()

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
	includeAncestors := true // bool | include resources of both requested scope, and its ancestors. for example, if project scope is requested with includeAncestors=true, department and cluster scoped assets will be included in the response. (optional)
	includeDescendants := true // bool | Whether the query should include from all nested layers within the specified scope (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.ListResourceClaimTemplates(context.Background()).ScopeType(scopeType).ScopeId(scopeId).Name(name).ClusterId(clusterId).Origin(origin).IncludeAncestors(includeAncestors).IncludeDescendants(includeDescendants).Execute()
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
 **includeAncestors** | **bool** | include resources of both requested scope, and its ancestors. for example, if project scope is requested with includeAncestors&#x3D;true, department and cluster scoped assets will be included in the response. | 
 **includeDescendants** | **bool** | Whether the query should include from all nested layers within the specified scope | 

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


## SyncClusterWideResourceClaimTemplates

> ResourceClaimTemplates SyncClusterWideResourceClaimTemplates(ctx, clusterUuid).ResourceClaimTemplates(resourceClaimTemplates).Execute()

sync the details of cluster-wide ResourceClaimTemplates to the control plane

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
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	resourceClaimTemplates := *openapiclient.NewResourceClaimTemplates([]openapiclient.ResourceClaimTemplate{*openapiclient.NewResourceClaimTemplate("rct-gpu-1", "runai")}) // ResourceClaimTemplates | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.SyncClusterWideResourceClaimTemplates(context.Background(), clusterUuid).ResourceClaimTemplates(resourceClaimTemplates).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.SyncClusterWideResourceClaimTemplates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncClusterWideResourceClaimTemplates`: ResourceClaimTemplates
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.SyncClusterWideResourceClaimTemplates`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncClusterWideResourceClaimTemplatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

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


## SyncDepartmentWideResourceClaimTemplates

> ResourceClaimTemplates SyncDepartmentWideResourceClaimTemplates(ctx, clusterUuid, departmentId).ResourceClaimTemplates(resourceClaimTemplates).Execute()

sync the details of department-wide ResourceClaimTemplates to the control plane

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
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	departmentId := "1" // string | The id of the department.
	resourceClaimTemplates := *openapiclient.NewResourceClaimTemplates([]openapiclient.ResourceClaimTemplate{*openapiclient.NewResourceClaimTemplate("rct-gpu-1", "runai")}) // ResourceClaimTemplates | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.SyncDepartmentWideResourceClaimTemplates(context.Background(), clusterUuid, departmentId).ResourceClaimTemplates(resourceClaimTemplates).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.SyncDepartmentWideResourceClaimTemplates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncDepartmentWideResourceClaimTemplates`: ResourceClaimTemplates
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.SyncDepartmentWideResourceClaimTemplates`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncDepartmentWideResourceClaimTemplatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## SyncProjectResourceClaimTemplates

> ResourceClaimTemplates SyncProjectResourceClaimTemplates(ctx, clusterUuid, namespace).ResourceClaimTemplates(resourceClaimTemplates).Execute()

sync the details of project ResourceClaimTemplates to the control plane

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
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	resourceClaimTemplates := *openapiclient.NewResourceClaimTemplates([]openapiclient.ResourceClaimTemplate{*openapiclient.NewResourceClaimTemplate("rct-gpu-1", "runai")}) // ResourceClaimTemplates | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimTemplatesAPI.SyncProjectResourceClaimTemplates(context.Background(), clusterUuid, namespace).ResourceClaimTemplates(resourceClaimTemplates).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimTemplatesAPI.SyncProjectResourceClaimTemplates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncProjectResourceClaimTemplates`: ResourceClaimTemplates
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimTemplatesAPI.SyncProjectResourceClaimTemplates`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncProjectResourceClaimTemplatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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

