# \MetricsAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetWorkloadsScrapeableMetrics**](MetricsAPI.md#GetWorkloadsScrapeableMetrics) | **Get** /api/v1/workloads/scrapeable-metrics | Get workloads metrics in openmetrics format for scraping.



## GetWorkloadsScrapeableMetrics

> string GetWorkloadsScrapeableMetrics(ctx).Execute()

Get workloads metrics in openmetrics format for scraping.

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MetricsAPI.GetWorkloadsScrapeableMetrics(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MetricsAPI.GetWorkloadsScrapeableMetrics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadsScrapeableMetrics`: string
	fmt.Fprintf(os.Stdout, "Response from `MetricsAPI.GetWorkloadsScrapeableMetrics`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadsScrapeableMetricsRequest struct via the builder pattern


### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

