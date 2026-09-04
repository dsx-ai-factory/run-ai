# \ModelInferenceApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateModelInference**](ModelInferenceApi.md#CreateModelInference) | **Post** /api/v1/model-inference | Create model inference.



## CreateModelInference

> ModelInference CreateModelInference(ctx).ModelInferenceCreationRequest(modelInferenceCreationRequest).Execute()

Create model inference.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    modelInferenceCreationRequest := *openapiclient.NewModelInferenceCreationRequest("Name_example", "ClusterId_example", int32(123), "Namespace_example", "ModelId_example") // ModelInferenceCreationRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ModelInferenceApi.CreateModelInference(context.Background()).ModelInferenceCreationRequest(modelInferenceCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ModelInferenceApi.CreateModelInference``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateModelInference`: ModelInference
    fmt.Fprintf(os.Stdout, "Response from `ModelInferenceApi.CreateModelInference`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateModelInferenceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **modelInferenceCreationRequest** | [**ModelInferenceCreationRequest**](ModelInferenceCreationRequest.md) |  | 

### Return type

[**ModelInference**](ModelInference.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

