# \TenantAPI

All URIs are relative to *http://localhost:8181*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ExtractTenantId**](TenantAPI.md#ExtractTenantId) | **Get** /internal/plugins/extract-tenant-id | Resolve tenant ID from the request hostname



## ExtractTenantId

> ExtractTenantId(ctx).XForwardedHost(xForwardedHost).Execute()

Resolve tenant ID from the request hostname



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/gateway-hub/pkg/client"
)

func main() {
	xForwardedHost := "xForwardedHost_example" // string | Hostname of the original client request, set by Traefik. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TenantAPI.ExtractTenantId(context.Background()).XForwardedHost(xForwardedHost).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantAPI.ExtractTenantId``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiExtractTenantIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xForwardedHost** | **string** | Hostname of the original client request, set by Traefik. | 

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

