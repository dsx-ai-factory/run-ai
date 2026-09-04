# \InternalAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetSettingByKeyForAllTenants**](InternalAPI.md#GetSettingByKeyForAllTenants) | **Get** /internal/tenants/all-tenants/setting/{key} | Get setting by key for all tenants
[**InternalTenantsTenantIdMetricStoreGet**](InternalAPI.md#InternalTenantsTenantIdMetricStoreGet) | **Get** /internal/tenants/{tenantId}/metric-store | Get the metric store for a tenant



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
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/backend-api-client"
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

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InternalTenantsTenantIdMetricStoreGet

> MetricStore InternalTenantsTenantIdMetricStoreGet(ctx, tenantId).Execute()

Get the metric store for a tenant

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
	tenantId := "tenantId_example" // string | Unique identifier of the tenant.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.InternalTenantsTenantIdMetricStoreGet(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.InternalTenantsTenantIdMetricStoreGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `InternalTenantsTenantIdMetricStoreGet`: MetricStore
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.InternalTenantsTenantIdMetricStoreGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Unique identifier of the tenant. | 

### Other Parameters

Other parameters are passed through a pointer to a apiInternalTenantsTenantIdMetricStoreGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**MetricStore**](MetricStore.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

