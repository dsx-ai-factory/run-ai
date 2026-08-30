# \MetricStoresAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateMetricStore**](MetricStoresAPI.md#CreateMetricStore) | **Post** /internal/metric-stores | Create/Save metric store for tenant
[**DeleteMetricStores**](MetricStoresAPI.md#DeleteMetricStores) | **Delete** /internal/metric-stores | Delete tenant metric stores
[**GetMetricStores**](MetricStoresAPI.md#GetMetricStores) | **Get** /internal/metric-stores | Get tenant metric stores



## CreateMetricStore

> MetricStore CreateMetricStore(ctx).MetricStore(metricStore).Execute()

Create/Save metric store for tenant



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
)

func main() {
	metricStore := *openapiclient.NewMetricStore(int32(123)) // MetricStore | metrics store

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MetricStoresAPI.CreateMetricStore(context.Background()).MetricStore(metricStore).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MetricStoresAPI.CreateMetricStore``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateMetricStore`: MetricStore
	fmt.Fprintf(os.Stdout, "Response from `MetricStoresAPI.CreateMetricStore`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateMetricStoreRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **metricStore** | [**MetricStore**](MetricStore.md) | metrics store | 

### Return type

[**MetricStore**](MetricStore.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteMetricStores

> DeleteMetricStores(ctx).OverrideTenantId(overrideTenantId).Execute()

Delete tenant metric stores



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
)

func main() {
	overrideTenantId := int32(123) // int32 | The tenant Id of the current request, overrides the tenantId in the header. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.MetricStoresAPI.DeleteMetricStores(context.Background()).OverrideTenantId(overrideTenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MetricStoresAPI.DeleteMetricStores``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteMetricStoresRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **overrideTenantId** | **int32** | The tenant Id of the current request, overrides the tenantId in the header. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetMetricStores

> []MetricStore GetMetricStores(ctx).OverrideTenantId(overrideTenantId).Execute()

Get tenant metric stores



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
)

func main() {
	overrideTenantId := int32(123) // int32 | The tenant Id of the current request, overrides the tenantId in the header. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MetricStoresAPI.GetMetricStores(context.Background()).OverrideTenantId(overrideTenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MetricStoresAPI.GetMetricStores``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetMetricStores`: []MetricStore
	fmt.Fprintf(os.Stdout, "Response from `MetricStoresAPI.GetMetricStores`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetMetricStoresRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **overrideTenantId** | **int32** | The tenant Id of the current request, overrides the tenantId in the header. | 

### Return type

[**[]MetricStore**](MetricStore.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

