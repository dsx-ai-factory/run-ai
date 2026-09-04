# \UnifiedAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetUnifiedWorkload**](UnifiedAPI.md#GetUnifiedWorkload) | **Post** /api/v1/workloads/unified | Request unified workload information (internal for UI purposes)



## GetUnifiedWorkload

> UnifiedWorkload GetUnifiedWorkload(ctx).UnifiedWorkloadRequest(unifiedWorkloadRequest).Execute()

Request unified workload information (internal for UI purposes)



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
	unifiedWorkloadRequest := *openapiclient.NewUnifiedWorkloadRequest() // UnifiedWorkloadRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UnifiedAPI.GetUnifiedWorkload(context.Background()).UnifiedWorkloadRequest(unifiedWorkloadRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UnifiedAPI.GetUnifiedWorkload``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetUnifiedWorkload`: UnifiedWorkload
	fmt.Fprintf(os.Stdout, "Response from `UnifiedAPI.GetUnifiedWorkload`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetUnifiedWorkloadRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **unifiedWorkloadRequest** | [**UnifiedWorkloadRequest**](UnifiedWorkloadRequest.md) |  | 

### Return type

[**UnifiedWorkload**](UnifiedWorkload.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

