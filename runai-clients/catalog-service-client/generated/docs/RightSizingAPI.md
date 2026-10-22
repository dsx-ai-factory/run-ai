# \RightSizingAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**RecommendInferenceResources**](RightSizingAPI.md#RecommendInferenceResources) | **Post** /api/v1/inference/right-sizing/recommendation | Recommend GPU allocation for a NIM test/dev workload.



## RecommendInferenceResources

> RightSizingResponse RecommendInferenceResources(ctx).RightSizingRequest(rightSizingRequest).Execute()

Recommend GPU allocation for a NIM test/dev workload.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client"
)

func main() {
	rightSizingRequest := *openapiclient.NewRightSizingRequest("ClusterId_example", "Image_example", "Tag_example") // RightSizingRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RightSizingAPI.RecommendInferenceResources(context.Background()).RightSizingRequest(rightSizingRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RightSizingAPI.RecommendInferenceResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RecommendInferenceResources`: RightSizingResponse
	fmt.Fprintf(os.Stdout, "Response from `RightSizingAPI.RecommendInferenceResources`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRecommendInferenceResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **rightSizingRequest** | [**RightSizingRequest**](RightSizingRequest.md) |  | 

### Return type

[**RightSizingResponse**](RightSizingResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

