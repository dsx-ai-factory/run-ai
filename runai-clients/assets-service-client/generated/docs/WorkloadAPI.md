# \WorkloadAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AssetNameConflictValidation**](WorkloadAPI.md#AssetNameConflictValidation) | **Get** /api/v1/asset/name-availability | Asset name conflict validation



## AssetNameConflictValidation

> AssetNameConflictValidation(ctx).Kind(kind).AssetName(assetName).ScopeType(scopeType).ScopeId(scopeId).Execute()

Asset name conflict validation



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
	kind := "kind_example" // string | Filter results by asset kind.
	assetName := "assetName_example" // string | the name of the asset
	scopeType := "scopeType_example" // string | The resource scope type to filter by.
	scopeId := "scopeId_example" // string | The id of the cluster, department or project, depending on the scope type. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadAPI.AssetNameConflictValidation(context.Background()).Kind(kind).AssetName(assetName).ScopeType(scopeType).ScopeId(scopeId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadAPI.AssetNameConflictValidation``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAssetNameConflictValidationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kind** | **string** | Filter results by asset kind. | 
 **assetName** | **string** | the name of the asset | 
 **scopeType** | **string** | The resource scope type to filter by. | 
 **scopeId** | **string** | The id of the cluster, department or project, depending on the scope type. | 

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

