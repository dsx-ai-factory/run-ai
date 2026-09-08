# \InternalAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetSettingByKeyForAllTenants**](InternalAPI.md#GetSettingByKeyForAllTenants) | **Get** /internal/tenants/all-tenants/setting/{key} | Get setting by key for all tenants



## GetSettingByKeyForAllTenants

> map[string]interface{} GetSettingByKeyForAllTenants(ctx, key).Execute()

Get setting by key for all tenants



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/client"
)

func main() {
	key := "key_example" // string | The setting key to retrieve

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.GetSettingByKeyForAllTenants(context.Background(), key).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.GetSettingByKeyForAllTenants``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSettingByKeyForAllTenants`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.GetSettingByKeyForAllTenants`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | **string** | The setting key to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSettingByKeyForAllTenantsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

