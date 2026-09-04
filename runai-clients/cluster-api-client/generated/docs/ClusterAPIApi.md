# \ClusterAPIAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetWorkloadStatus**](ClusterAPIAPI.md#GetWorkloadStatus) | **Get** /cluster-api/api/v1/{projectName}/workloads/{workloadType}/{workloadFramework}/{workloadName}/status | Get the status of a workload



## GetWorkloadStatus

> WorkloadStatus GetWorkloadStatus(ctx, projectName, workloadType, workloadName, workloadFramework).Execute()

Get the status of a workload



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/cluster-api-client"
)

func main() {
	projectName := "project1" // string | The project name
	workloadType := "workloadType_example" // string | The workload type
	workloadName := "jp" // string | The workload name
	workloadFramework := "workloadFramework_example" // string | The workload framework

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClusterAPIAPI.GetWorkloadStatus(context.Background(), projectName, workloadType, workloadName, workloadFramework).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterAPIAPI.GetWorkloadStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadStatus`: WorkloadStatus
	fmt.Fprintf(os.Stdout, "Response from `ClusterAPIAPI.GetWorkloadStatus`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**projectName** | **string** | The project name | 
**workloadType** | **string** | The workload type | 
**workloadName** | **string** | The workload name | 
**workloadFramework** | **string** | The workload framework | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------





### Return type

[**WorkloadStatus**](WorkloadStatus.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

