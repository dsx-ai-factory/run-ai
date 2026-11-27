# \InferencesAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetHuggingFaceModelInfo**](InferencesAPI.md#GetHuggingFaceModelInfo) | **Post** /api/v2/huggingface-model/info-request | Get info of a Hugging Face model.



## GetHuggingFaceModelInfo

> HuggingFaceModelInfo GetHuggingFaceModelInfo(ctx).HuggingFaceModelInfoRequest(huggingFaceModelInfoRequest).Execute()

Get info of a Hugging Face model.



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
	huggingFaceModelInfoRequest := *openapiclient.NewHuggingFaceModelInfoRequest("meta-llama/Llama-2-7b-chat-hf") // HuggingFaceModelInfoRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.GetHuggingFaceModelInfo(context.Background()).HuggingFaceModelInfoRequest(huggingFaceModelInfoRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.GetHuggingFaceModelInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetHuggingFaceModelInfo`: HuggingFaceModelInfo
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.GetHuggingFaceModelInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetHuggingFaceModelInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **huggingFaceModelInfoRequest** | [**HuggingFaceModelInfoRequest**](HuggingFaceModelInfoRequest.md) |  | 

### Return type

[**HuggingFaceModelInfo**](HuggingFaceModelInfo.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

