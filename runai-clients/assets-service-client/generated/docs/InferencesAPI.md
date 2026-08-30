# \InferencesAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateInference**](InferencesAPI.md#CreateInference) | **Post** /api/v1/inference | Create inference workload.
[**CreateInferenceV2**](InferencesAPI.md#CreateInferenceV2) | **Post** /api/v2/inference | Create an inference - assets based.
[**DeleteInferenceById**](InferencesAPI.md#DeleteInferenceById) | **Delete** /api/v1/inference/{workloadId} | Delete an inference workload.
[**GetHuggingFaceModelInfo**](InferencesAPI.md#GetHuggingFaceModelInfo) | **Post** /api/v2/huggingface-model/info-request | Get info of a Hugging Face model.
[**GetInferenceById**](InferencesAPI.md#GetInferenceById) | **Get** /api/v1/inference/{workloadId} | Get inference content.
[**GetInferenceByIdV2**](InferencesAPI.md#GetInferenceByIdV2) | **Get** /api/v2/inference/{workloadId} | Get inference data - assets based.
[**GetInferenceMerged**](InferencesAPI.md#GetInferenceMerged) | **Get** /api/v2/inference/merged | Get merged inference workload



## CreateInference

> Inference CreateInference(ctx).WorkloadCreationRequest(workloadCreationRequest).Execute()

Create inference workload.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	workloadCreationRequest := *openapiclient.NewWorkloadCreationRequest("Name_example", "71f69d83-ba66-4822-adf5-55ce55efd210", int32(123), "Namespace_example", *openapiclient.NewAssetsIds("Environment_example")) // WorkloadCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.CreateInference(context.Background()).WorkloadCreationRequest(workloadCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.CreateInference``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateInference`: Inference
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.CreateInference`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateInferenceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadCreationRequest** | [**WorkloadCreationRequest**](WorkloadCreationRequest.md) |  | 

### Return type

[**Inference**](Inference.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateInferenceV2

> InferenceV2 CreateInferenceV2(ctx).InferenceCreationRequestV2(inferenceCreationRequestV2).Execute()

Create an inference - assets based.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	inferenceCreationRequestV2 := *openapiclient.NewInferenceCreationRequestV2(*openapiclient.NewWorkloadCreationMeta("my-workload-name", "1", "71f69d83-ba66-4822-adf5-55ce55efd210"), *openapiclient.NewAssetsIds("Environment_example")) // InferenceCreationRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.CreateInferenceV2(context.Background()).InferenceCreationRequestV2(inferenceCreationRequestV2).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.CreateInferenceV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateInferenceV2`: InferenceV2
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.CreateInferenceV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateInferenceV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **inferenceCreationRequestV2** | [**InferenceCreationRequestV2**](InferenceCreationRequestV2.md) |  | 

### Return type

[**InferenceV2**](InferenceV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteInferenceById

> HttpResponse DeleteInferenceById(ctx, workloadId).Execute()

Delete an inference workload.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.DeleteInferenceById(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.DeleteInferenceById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteInferenceById`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.DeleteInferenceById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteInferenceByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HttpResponse**](HttpResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


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
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
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


## GetInferenceById

> Inference GetInferenceById(ctx, workloadId).Execute()

Get inference content.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.GetInferenceById(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.GetInferenceById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInferenceById`: Inference
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.GetInferenceById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInferenceByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Inference**](Inference.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInferenceByIdV2

> InferenceV2 GetInferenceByIdV2(ctx, workloadId).Execute()

Get inference data - assets based.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.GetInferenceByIdV2(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.GetInferenceByIdV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInferenceByIdV2`: InferenceV2
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.GetInferenceByIdV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInferenceByIdV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**InferenceV2**](InferenceV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInferenceMerged

> InferenceV2 GetInferenceMerged(ctx).ProjectId(projectId).WorkloadId(workloadId).TemplateId(templateId).Execute()

Get merged inference workload



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	projectId := "4500001" // string | The id of the project (optional)
	workloadId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the workload (optional)
	templateId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the template (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InferencesAPI.GetInferenceMerged(context.Background()).ProjectId(projectId).WorkloadId(workloadId).TemplateId(templateId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InferencesAPI.GetInferenceMerged``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInferenceMerged`: InferenceV2
	fmt.Fprintf(os.Stdout, "Response from `InferencesAPI.GetInferenceMerged`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetInferenceMergedRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **string** | The id of the project | 
 **workloadId** | **string** | The id of the workload | 
 **templateId** | **string** | The id of the template | 

### Return type

[**InferenceV2**](InferenceV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

