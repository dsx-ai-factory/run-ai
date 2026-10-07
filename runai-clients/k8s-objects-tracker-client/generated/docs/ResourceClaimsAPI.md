# \ResourceClaimsAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ListResourceClaims**](ResourceClaimsAPI.md#ListResourceClaims) | **Get** /api/v1/resource-claims | list details of ResourceClaims
[**SyncProjectResourceClaims**](ResourceClaimsAPI.md#SyncProjectResourceClaims) | **Post** /api/v1/resource-claims/{clusterUuid}/namespaces/{namespace} | sync the details of project ResourceClaims to the control plane



## ListResourceClaims

> ResourceClaims ListResourceClaims(ctx).ScopeType(scopeType).ScopeId(scopeId).Name(name).ClusterId(clusterId).IncludeAncestors(includeAncestors).IncludeDescendants(includeDescendants).Execute()

list details of ResourceClaims

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
	includeAncestors := true // bool | include resources of both requested scope, and its ancestors. for example, if project scope is requested with includeAncestors=true, department and cluster scoped assets will be included in the response. (optional)
	includeDescendants := true // bool | Whether the query should include from all nested layers within the specified scope (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimsAPI.ListResourceClaims(context.Background()).ScopeType(scopeType).ScopeId(scopeId).Name(name).ClusterId(clusterId).IncludeAncestors(includeAncestors).IncludeDescendants(includeDescendants).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimsAPI.ListResourceClaims``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListResourceClaims`: ResourceClaims
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimsAPI.ListResourceClaims`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListResourceClaimsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scopeType** | **string** | The type of resource we want to filter by. | 
 **scopeId** | **string** | filter by scope id | 
 **name** | **string** | filter by name | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **includeAncestors** | **bool** | include resources of both requested scope, and its ancestors. for example, if project scope is requested with includeAncestors&#x3D;true, department and cluster scoped assets will be included in the response. | 
 **includeDescendants** | **bool** | Whether the query should include from all nested layers within the specified scope | 

### Return type

[**ResourceClaims**](ResourceClaims.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncProjectResourceClaims

> ResourceClaims SyncProjectResourceClaims(ctx, clusterUuid, namespace).ResourceClaims(resourceClaims).Execute()

sync the details of project ResourceClaims to the control plane



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
	resourceClaims := *openapiclient.NewResourceClaims([]openapiclient.ResourceClaim{*openapiclient.NewResourceClaim("pod-x-claim-0-abcde", "runai-project-foo")}) // ResourceClaims | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResourceClaimsAPI.SyncProjectResourceClaims(context.Background(), clusterUuid, namespace).ResourceClaims(resourceClaims).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResourceClaimsAPI.SyncProjectResourceClaims``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncProjectResourceClaims`: ResourceClaims
	fmt.Fprintf(os.Stdout, "Response from `ResourceClaimsAPI.SyncProjectResourceClaims`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncProjectResourceClaimsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **resourceClaims** | [**ResourceClaims**](ResourceClaims.md) |  | 

### Return type

[**ResourceClaims**](ResourceClaims.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

