# \WorkloadActionsAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetWorkloadActionsHistory**](WorkloadActionsAPI.md#GetWorkloadActionsHistory) | **Get** /api/v1/workloads/{workloadId}/actions | Get workload action history.
[**InternalCreateWorkloadActions**](WorkloadActionsAPI.md#InternalCreateWorkloadActions) | **Post** /internal/workloads/actions | Store workload actions.



## GetWorkloadActionsHistory

> GetWorkloadActionsResponse GetWorkloadActionsHistory(ctx, workloadId).Offset(offset).Limit(limit).Execute()

Get workload action history.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadActionsAPI.GetWorkloadActionsHistory(context.Background(), workloadId).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadActionsAPI.GetWorkloadActionsHistory``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadActionsHistory`: GetWorkloadActionsResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadActionsAPI.GetWorkloadActionsHistory`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadActionsHistoryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**GetWorkloadActionsResponse**](GetWorkloadActionsResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InternalCreateWorkloadActions

> InternalCreateWorkloadActions(ctx).CreateWorkloadActionsRequest(createWorkloadActionsRequest).Execute()

Store workload actions.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	createWorkloadActionsRequest := *openapiclient.NewCreateWorkloadActionsRequest([]openapiclient.WorkloadActionRecordInput{*openapiclient.NewWorkloadActionRecordInput("497f6eca-6276-4993-bfeb-53cbbbba6f08", "71f69d83-ba66-4822-adf5-55ce55efd210", "WorkloadId_example", "my-workload-name", openapiclient.WorkloadActionType("create"), time.Now())}) // CreateWorkloadActionsRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkloadActionsAPI.InternalCreateWorkloadActions(context.Background()).CreateWorkloadActionsRequest(createWorkloadActionsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadActionsAPI.InternalCreateWorkloadActions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInternalCreateWorkloadActionsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createWorkloadActionsRequest** | [**CreateWorkloadActionsRequest**](CreateWorkloadActionsRequest.md) |  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

