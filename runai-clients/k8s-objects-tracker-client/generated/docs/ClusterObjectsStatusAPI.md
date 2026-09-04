# \ClusterObjectsStatusAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ListClusterObjectsStatus**](ClusterObjectsStatusAPI.md#ListClusterObjectsStatus) | **Get** /internal/api/v1/cluster-objects/status | list of objects in the cluster and their found issues



## ListClusterObjectsStatus

> ClusterObjectsStatus ListClusterObjectsStatus(ctx).ResourceType(resourceType).SecretType(secretType).ScopeId(scopeId).ScopeType(scopeType).Execute()

list of objects in the cluster and their found issues

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
	resourceType := openapiclient.ClusterResourceType("pvc") // ClusterResourceType | filter by resource type (optional)
	secretType := "secretType_example" // string | filter by secret type. supported values - password, access-key, docker-registry. (optional)
	scopeId := "scopeId_example" // string | filter by scope id (optional)
	scopeType := "project" // string | The type of resource we want to filter by. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClusterObjectsStatusAPI.ListClusterObjectsStatus(context.Background()).ResourceType(resourceType).SecretType(secretType).ScopeId(scopeId).ScopeType(scopeType).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterObjectsStatusAPI.ListClusterObjectsStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListClusterObjectsStatus`: ClusterObjectsStatus
	fmt.Fprintf(os.Stdout, "Response from `ClusterObjectsStatusAPI.ListClusterObjectsStatus`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListClusterObjectsStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **resourceType** | [**ClusterResourceType**](ClusterResourceType.md) | filter by resource type | 
 **secretType** | **string** | filter by secret type. supported values - password, access-key, docker-registry. | 
 **scopeId** | **string** | filter by scope id | 
 **scopeType** | **string** | The type of resource we want to filter by. | 

### Return type

[**ClusterObjectsStatus**](ClusterObjectsStatus.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

