# \NodeTypesAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateNodeType**](NodeTypesAPI.md#CreateNodeType) | **Post** /api/v1/org-unit/node-type | Create node type
[**DeleteNodeType**](NodeTypesAPI.md#DeleteNodeType) | **Delete** /api/v1/org-unit/node-type/{nodeTypeId} | Delete node type
[**GetNodeTypes**](NodeTypesAPI.md#GetNodeTypes) | **Get** /api/v1/org-unit/node-type | Get node types
[**UpdateNodeType**](NodeTypesAPI.md#UpdateNodeType) | **Put** /api/v1/org-unit/node-type/{nodeTypeId} | Update node type



## CreateNodeType

> NodeType CreateNodeType(ctx).NodeTypeCreateRequest(nodeTypeCreateRequest).Execute()

Create node type



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {
	nodeTypeCreateRequest := *openapiclient.NewNodeTypeCreateRequest() // NodeTypeCreateRequest | NodeType to create.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodeTypesAPI.CreateNodeType(context.Background()).NodeTypeCreateRequest(nodeTypeCreateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeTypesAPI.CreateNodeType``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateNodeType`: NodeType
	fmt.Fprintf(os.Stdout, "Response from `NodeTypesAPI.CreateNodeType`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateNodeTypeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **nodeTypeCreateRequest** | [**NodeTypeCreateRequest**](NodeTypeCreateRequest.md) | NodeType to create. | 

### Return type

[**NodeType**](NodeType.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteNodeType

> DeleteNodeType(ctx, nodeTypeId).Execute()

Delete node type



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {
	nodeTypeId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The node-type id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NodeTypesAPI.DeleteNodeType(context.Background(), nodeTypeId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeTypesAPI.DeleteNodeType``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodeTypeId** | **string** | The node-type id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteNodeTypeRequest struct via the builder pattern


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


## GetNodeTypes

> GetNodeTypes200Response GetNodeTypes(ctx).ClusterId(clusterId).Execute()

Get node types



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodeTypesAPI.GetNodeTypes(context.Background()).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeTypesAPI.GetNodeTypes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodeTypes`: GetNodeTypes200Response
	fmt.Fprintf(os.Stdout, "Response from `NodeTypesAPI.GetNodeTypes`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNodeTypesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 

### Return type

[**GetNodeTypes200Response**](GetNodeTypes200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateNodeType

> NodeType UpdateNodeType(ctx, nodeTypeId).NodeTypeUpdateRequest(nodeTypeUpdateRequest).Execute()

Update node type



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {
	nodeTypeId := "575c19e8-c7c3-45b0-8290-2f47397a8383" // string | The node-type id
	nodeTypeUpdateRequest := *openapiclient.NewNodeTypeUpdateRequest() // NodeTypeUpdateRequest | NodeType to update.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NodeTypesAPI.UpdateNodeType(context.Background(), nodeTypeId).NodeTypeUpdateRequest(nodeTypeUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodeTypesAPI.UpdateNodeType``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateNodeType`: NodeType
	fmt.Fprintf(os.Stdout, "Response from `NodeTypesAPI.UpdateNodeType`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodeTypeId** | **string** | The node-type id | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNodeTypeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodeTypeUpdateRequest** | [**NodeTypeUpdateRequest**](NodeTypeUpdateRequest.md) | NodeType to update. | 

### Return type

[**NodeType**](NodeType.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

