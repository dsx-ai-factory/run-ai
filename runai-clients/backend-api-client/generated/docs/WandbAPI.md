# \WandbAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateSweep**](WandbAPI.md#CreateSweep) | **Post** /v1/k8s/wandb/sweeps | Create a sweep.



## CreateSweep

> WandbSweepsResponse CreateSweep(ctx).WandbSweepsRequest(wandbSweepsRequest).Execute()

Create a sweep.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/backend-api-client"
)

func main() {
	wandbSweepsRequest := *openapiclient.NewWandbSweepsRequest("703b507256379410b2dd15fd10500te065r307y0m", "myproject", map[string]interface{}({"name":"my-sweep","method":"random","metric":{"goal":"maximize","name":"val_acc"},"parameters":{"lr":{"max":0.1,"min":1.0E-4}}})) // WandbSweepsRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WandbAPI.CreateSweep(context.Background()).WandbSweepsRequest(wandbSweepsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WandbAPI.CreateSweep``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateSweep`: WandbSweepsResponse
	fmt.Fprintf(os.Stdout, "Response from `WandbAPI.CreateSweep`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateSweepRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **wandbSweepsRequest** | [**WandbSweepsRequest**](WandbSweepsRequest.md) |  | 

### Return type

[**WandbSweepsResponse**](WandbSweepsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

