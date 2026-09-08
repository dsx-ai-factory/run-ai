# \RedirectValidationAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ValidateRedirect**](RedirectValidationAPI.md#ValidateRedirect) | **Get** /internal/plugins/workload-redirect-validation | Validate the OAuth redirect_uri for workload auth flows



## ValidateRedirect

> ValidateRedirect(ctx).RedirectUri(redirectUri).ClientId(clientId).Execute()

Validate the OAuth redirect_uri for workload auth flows



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
	redirectUri := "redirectUri_example" // string | OAuth redirect URI to validate.
	clientId := "clientId_example" // string | OAuth client_id in the format runai-internal.<uuid>.workloads_auth.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RedirectValidationAPI.ValidateRedirect(context.Background()).RedirectUri(redirectUri).ClientId(clientId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RedirectValidationAPI.ValidateRedirect``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiValidateRedirectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **redirectUri** | **string** | OAuth redirect URI to validate. | 
 **clientId** | **string** | OAuth client_id in the format runai-internal.&lt;uuid&gt;.workloads_auth. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/html

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

