# \ModelInferenceAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateModelInference**](ModelInferenceAPI.md#CreateModelInference) | **Post** /api/v1/model-inference | Create model-inference.
[**CreateModelInferenceV2**](ModelInferenceAPI.md#CreateModelInferenceV2) | **Post** /api/v2/model-inference | Create a model-inference



## CreateModelInference

> ModelInference CreateModelInference(ctx).ModelInferenceCreationRequest(modelInferenceCreationRequest).Execute()

Create model-inference.



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
	modelInferenceCreationRequest := *openapiclient.NewModelInferenceCreationRequest("Name_example", "71f69d83-ba66-4822-adf5-55ce55efd210", int32(123), "Namespace_example", "ModelId_example") // ModelInferenceCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelInferenceAPI.CreateModelInference(context.Background()).ModelInferenceCreationRequest(modelInferenceCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelInferenceAPI.CreateModelInference``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateModelInference`: ModelInference
	fmt.Fprintf(os.Stdout, "Response from `ModelInferenceAPI.CreateModelInference`: %v\n", resp)
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


## CreateModelInferenceV2

> ModelInferenceV2 CreateModelInferenceV2(ctx).ModelInferenceCreationRequestV2(modelInferenceCreationRequestV2).Execute()

Create a model-inference



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
	modelInferenceCreationRequestV2 := *openapiclient.NewModelInferenceCreationRequestV2(*openapiclient.NewWorkloadCreationMeta("my-workload-name", "1", "71f69d83-ba66-4822-adf5-55ce55efd210"), *openapiclient.NewModelInferenceCreationRequestV2Assets("Model_example", "Compute_example")) // ModelInferenceCreationRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelInferenceAPI.CreateModelInferenceV2(context.Background()).ModelInferenceCreationRequestV2(modelInferenceCreationRequestV2).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelInferenceAPI.CreateModelInferenceV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateModelInferenceV2`: ModelInferenceV2
	fmt.Fprintf(os.Stdout, "Response from `ModelInferenceAPI.CreateModelInferenceV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateModelInferenceV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **modelInferenceCreationRequestV2** | [**ModelInferenceCreationRequestV2**](ModelInferenceCreationRequestV2.md) |  | 

### Return type

[**ModelInferenceV2**](ModelInferenceV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

