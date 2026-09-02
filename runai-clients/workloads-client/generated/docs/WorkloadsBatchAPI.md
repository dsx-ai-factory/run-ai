# \WorkloadsBatchAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**BatchWorkloads**](WorkloadsBatchAPI.md#BatchWorkloads) | **Post** /api/v1/workloads/batch | Workload batch operations.



## BatchWorkloads

> BatchResponse BatchWorkloads(ctx).WorkloadBatch(workloadBatch).Execute()

Workload batch operations.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	workloadBatch := *openapiclient.NewWorkloadBatch([]string{"Ids_example"}, openapiclient.Action("delete")) // WorkloadBatch |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsBatchAPI.BatchWorkloads(context.Background()).WorkloadBatch(workloadBatch).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsBatchAPI.BatchWorkloads``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BatchWorkloads`: BatchResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsBatchAPI.BatchWorkloads`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiBatchWorkloadsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadBatch** | [**WorkloadBatch**](WorkloadBatch.md) |  | 

### Return type

[**BatchResponse**](BatchResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

