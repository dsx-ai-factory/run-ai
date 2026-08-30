# \MetricsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetMetrics**](MetricsAPI.md#GetMetrics) | **Post** /internal/metrics | Get metrics.



## GetMetrics

> GetMetrics200Response GetMetrics(ctx).MetricInformation(metricInformation).OverrideTenantId(overrideTenantId).Execute()

Get metrics.

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
	metricInformation := *openapiclient.NewMetricInformation("Name_example") // MetricInformation | Requested Metric Information
	overrideTenantId := int32(123) // int32 | The tenant Id of the current request, overrides the tenantId in the header. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MetricsAPI.GetMetrics(context.Background()).MetricInformation(metricInformation).OverrideTenantId(overrideTenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MetricsAPI.GetMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetMetrics`: GetMetrics200Response
	fmt.Fprintf(os.Stdout, "Response from `MetricsAPI.GetMetrics`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetMetricsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **metricInformation** | [**MetricInformation**](MetricInformation.md) | Requested Metric Information | 
 **overrideTenantId** | **int32** | The tenant Id of the current request, overrides the tenantId in the header. | 

### Return type

[**GetMetrics200Response**](GetMetrics200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

