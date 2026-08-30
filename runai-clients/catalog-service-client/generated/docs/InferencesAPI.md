# \InferencesAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SearchHuggingFaceCatalog**](InferencesAPI.md#SearchHuggingFaceCatalog) | **Get** /api/v1/huggingface/catalog | Search Hugging Face models.



## SearchHuggingFaceCatalog

> HuggingFaceModelSearchResults SearchHuggingFaceCatalog(ctx).Query(query).Execute()

Search Hugging Face models.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/catalog-service-client"
)

func main() {
	query := "meta-llama/Llama-3.1-405B-Instruct or Llama-3.1-405B-Instruct or Llama-3.1-405B" // string | Search query used to look up Hugging Face models. Accepts a full repository ID (`org/model`), a standalone model name, or a partial match string. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.SearchHuggingFaceCatalog(context.Background()).Query(query).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.SearchHuggingFaceCatalog``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SearchHuggingFaceCatalog`: HuggingFaceModelSearchResults
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.SearchHuggingFaceCatalog`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSearchHuggingFaceCatalogRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **query** | **string** | Search query used to look up Hugging Face models. Accepts a full repository ID (&#x60;org/model&#x60;), a standalone model name, or a partial match string. | 

### Return type

[**HuggingFaceModelSearchResults**](HuggingFaceModelSearchResults.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

