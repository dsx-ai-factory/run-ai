# \StatusAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**Status**](StatusAPI.md#Status) | **Get** /cluster-api/status | status check



## Status

> Status(ctx).Execute()

status check

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/cluster-api-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.StatusAPI.Status(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StatusAPI.Status``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiStatusRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

