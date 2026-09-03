# \NVIDIANIMAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateNimService**](NVIDIANIMAPI.md#CreateNimService) | **Post** /api/v2/workloads/nim-services | Create a NVIDIA NIM service. [Experimental]
[**GetNimServiceById**](NVIDIANIMAPI.md#GetNimServiceById) | **Get** /api/v2/workloads/nim-services/{WorkloadV2Id} | Get a NVIDIA NIM service. [Experimental]
[**UpdateNimServiceSpec**](NVIDIANIMAPI.md#UpdateNimServiceSpec) | **Patch** /api/v2/workloads/nim-services/{WorkloadV2Id} | Update NVIDIA NIM service spec. [Experimental]



## CreateNimService

> NimServiceResponse CreateNimService(ctx).NimServiceCreateRequest(nimServiceCreateRequest).Execute()

Create a NVIDIA NIM service. [Experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client"
)

func main() {
	nimServiceCreateRequest := *openapiclient.NewNimServiceCreateRequest(*openapiclient.NewNIMServiceMetadataCreateParams("my-workload-name", "1"), *openapiclient.NewNimServiceSpec()) // NimServiceCreateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NVIDIANIMAPI.CreateNimService(context.Background()).NimServiceCreateRequest(nimServiceCreateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NVIDIANIMAPI.CreateNimService``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateNimService`: NimServiceResponse
	fmt.Fprintf(os.Stdout, "Response from `NVIDIANIMAPI.CreateNimService`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateNimServiceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **nimServiceCreateRequest** | [**NimServiceCreateRequest**](NimServiceCreateRequest.md) |  | 

### Return type

[**NimServiceResponse**](NimServiceResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNimServiceById

> NimServiceResponse GetNimServiceById(ctx, workloadV2Id).Execute()

Get a NVIDIA NIM service. [Experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client"
)

func main() {
	workloadV2Id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NVIDIANIMAPI.GetNimServiceById(context.Background(), workloadV2Id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NVIDIANIMAPI.GetNimServiceById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNimServiceById`: NimServiceResponse
	fmt.Fprintf(os.Stdout, "Response from `NVIDIANIMAPI.GetNimServiceById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadV2Id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNimServiceByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**NimServiceResponse**](NimServiceResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateNimServiceSpec

> NimServiceResponse UpdateNimServiceSpec(ctx, workloadV2Id).NimServiceUpdateRequest(nimServiceUpdateRequest).Execute()

Update NVIDIA NIM service spec. [Experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client"
)

func main() {
	workloadV2Id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	nimServiceUpdateRequest := *openapiclient.NewNimServiceUpdateRequest(*openapiclient.NewNimServiceSpec()) // NimServiceUpdateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NVIDIANIMAPI.UpdateNimServiceSpec(context.Background(), workloadV2Id).NimServiceUpdateRequest(nimServiceUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NVIDIANIMAPI.UpdateNimServiceSpec``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateNimServiceSpec`: NimServiceResponse
	fmt.Fprintf(os.Stdout, "Response from `NVIDIANIMAPI.UpdateNimServiceSpec`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadV2Id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNimServiceSpecRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nimServiceUpdateRequest** | [**NimServiceUpdateRequest**](NimServiceUpdateRequest.md) |  | 

### Return type

[**NimServiceResponse**](NimServiceResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

