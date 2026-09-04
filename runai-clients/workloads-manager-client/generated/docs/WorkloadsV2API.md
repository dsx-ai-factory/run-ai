# \WorkloadsV2API

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**BatchWorkloads**](WorkloadsV2API.md#BatchWorkloads) | **Post** /internal/workloads/batch | Workload batch operations.
[**CreateWorkloadInternal**](WorkloadsV2API.md#CreateWorkloadInternal) | **Post** /internal/v2/workloads | Create a workload from internal v1-to-v2 conversion.
[**CreateWorkloadV2**](WorkloadsV2API.md#CreateWorkloadV2) | **Post** /api/v2/workloads | Create a workload.
[**DeleteWorkloadV2**](WorkloadsV2API.md#DeleteWorkloadV2) | **Delete** /api/v2/workloads/{WorkloadV2Id} | Delete a workload.
[**GetWorkloadV2ById**](WorkloadsV2API.md#GetWorkloadV2ById) | **Get** /api/v2/workloads/{WorkloadV2Id} | Get a specific workload.
[**GetWorkloadV2Updates**](WorkloadsV2API.md#GetWorkloadV2Updates) | **Post** /api/v2/workloads/updates | Get workload updates for agent polling + report issues with previous updates
[**SetWorkloadRunState**](WorkloadsV2API.md#SetWorkloadRunState) | **Patch** /api/v2/workloads/{WorkloadV2Id}/runState | Set the run state of a workload.
[**UpdateWorkloadInternal**](WorkloadsV2API.md#UpdateWorkloadInternal) | **Put** /internal/v2/workloads/{WorkloadV2Id} | Update a workload created through the internal v1-to-v2 conversion endpoint.
[**UpdateWorkloadV2**](WorkloadsV2API.md#UpdateWorkloadV2) | **Put** /api/v2/workloads/{WorkloadV2Id} | Update workload spec.



## BatchWorkloads

> BatchResponse BatchWorkloads(ctx).WorkloadsV2Batch(workloadsV2Batch).Execute()

Workload batch operations.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client"
)

func main() {
	workloadsV2Batch := *openapiclient.NewWorkloadsV2Batch([]string{"Ids_example"}, openapiclient.WorkloadsV2BatchAction("delete")) // WorkloadsV2Batch |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsV2API.BatchWorkloads(context.Background()).WorkloadsV2Batch(workloadsV2Batch).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsV2API.BatchWorkloads``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BatchWorkloads`: BatchResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsV2API.BatchWorkloads`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiBatchWorkloadsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadsV2Batch** | [**WorkloadsV2Batch**](WorkloadsV2Batch.md) |  | 

### Return type

[**BatchResponse**](BatchResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateWorkloadInternal

> WorkloadV2 CreateWorkloadInternal(ctx).WorkloadV2CreateRequest(workloadV2CreateRequest).DryRun(dryRun).Execute()

Create a workload from internal v1-to-v2 conversion.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client"
)

func main() {
	workloadV2CreateRequest := *openapiclient.NewWorkloadV2CreateRequest(*openapiclient.NewWorkloadV2MetadataCreateParams()) // WorkloadV2CreateRequest | Workload metadata and manifest
	dryRun := false // bool | True to perform validations only, without side effects. (optional) (default to false)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsV2API.CreateWorkloadInternal(context.Background()).WorkloadV2CreateRequest(workloadV2CreateRequest).DryRun(dryRun).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsV2API.CreateWorkloadInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateWorkloadInternal`: WorkloadV2
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsV2API.CreateWorkloadInternal`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateWorkloadInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadV2CreateRequest** | [**WorkloadV2CreateRequest**](WorkloadV2CreateRequest.md) | Workload metadata and manifest | 
 **dryRun** | **bool** | True to perform validations only, without side effects. | [default to false]

### Return type

[**WorkloadV2**](WorkloadV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateWorkloadV2

> WorkloadV2 CreateWorkloadV2(ctx).WorkloadV2CreateRequest(workloadV2CreateRequest).DryRun(dryRun).Execute()

Create a workload.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client"
)

func main() {
	workloadV2CreateRequest := *openapiclient.NewWorkloadV2CreateRequest(*openapiclient.NewWorkloadV2MetadataCreateParams()) // WorkloadV2CreateRequest | Workload metadata and manifest
	dryRun := false // bool | True to perform validations only, without side effects. (optional) (default to false)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsV2API.CreateWorkloadV2(context.Background()).WorkloadV2CreateRequest(workloadV2CreateRequest).DryRun(dryRun).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsV2API.CreateWorkloadV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateWorkloadV2`: WorkloadV2
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsV2API.CreateWorkloadV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateWorkloadV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadV2CreateRequest** | [**WorkloadV2CreateRequest**](WorkloadV2CreateRequest.md) | Workload metadata and manifest | 
 **dryRun** | **bool** | True to perform validations only, without side effects. | [default to false]

### Return type

[**WorkloadV2**](WorkloadV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteWorkloadV2

> WorkloadV2 DeleteWorkloadV2(ctx, workloadV2Id).Execute()

Delete a workload.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client"
)

func main() {
	workloadV2Id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsV2API.DeleteWorkloadV2(context.Background(), workloadV2Id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsV2API.DeleteWorkloadV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteWorkloadV2`: WorkloadV2
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsV2API.DeleteWorkloadV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadV2Id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteWorkloadV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**WorkloadV2**](WorkloadV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadV2ById

> WorkloadV2GetResponse GetWorkloadV2ById(ctx, workloadV2Id).Execute()

Get a specific workload.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client"
)

func main() {
	workloadV2Id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsV2API.GetWorkloadV2ById(context.Background(), workloadV2Id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsV2API.GetWorkloadV2ById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadV2ById`: WorkloadV2GetResponse
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsV2API.GetWorkloadV2ById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadV2Id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadV2ByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**WorkloadV2GetResponse**](WorkloadV2GetResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadV2Updates

> WorkloadsV2Updates GetWorkloadV2Updates(ctx).ClusterId(clusterId).UpdatesAfter(updatesAfter).Limit(limit).ClusterSyncReport(clusterSyncReport).Execute()

Get workload updates for agent polling + report issues with previous updates



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	updatesAfter := int64(12345) // int64 | list only data with update counter greater than the provided value. for example updatesAfter=5 should fetch data with updates counter 6 and above.
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsV2API.GetWorkloadV2Updates(context.Background()).ClusterId(clusterId).UpdatesAfter(updatesAfter).Limit(limit).ClusterSyncReport(clusterSyncReport).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsV2API.GetWorkloadV2Updates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadV2Updates`: WorkloadsV2Updates
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsV2API.GetWorkloadV2Updates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadV2UpdatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **updatesAfter** | **int64** | list only data with update counter greater than the provided value. for example updatesAfter&#x3D;5 should fetch data with updates counter 6 and above. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **clusterSyncReport** | [**ClusterSyncReport**](ClusterSyncReport.md) |  | 

### Return type

[**WorkloadsV2Updates**](WorkloadsV2Updates.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetWorkloadRunState

> WorkloadV2 SetWorkloadRunState(ctx, workloadV2Id).WorkloadRunStateRequest(workloadRunStateRequest).Execute()

Set the run state of a workload.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client"
)

func main() {
	workloadV2Id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	workloadRunStateRequest := *openapiclient.NewWorkloadRunStateRequest(openapiclient.RunState("suspend")) // WorkloadRunStateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsV2API.SetWorkloadRunState(context.Background(), workloadV2Id).WorkloadRunStateRequest(workloadRunStateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsV2API.SetWorkloadRunState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SetWorkloadRunState`: WorkloadV2
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsV2API.SetWorkloadRunState`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadV2Id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetWorkloadRunStateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadRunStateRequest** | [**WorkloadRunStateRequest**](WorkloadRunStateRequest.md) |  | 

### Return type

[**WorkloadV2**](WorkloadV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateWorkloadInternal

> WorkloadV2 UpdateWorkloadInternal(ctx, workloadV2Id).WorkloadV2UpdateRequest(workloadV2UpdateRequest).Execute()

Update a workload created through the internal v1-to-v2 conversion endpoint.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client"
)

func main() {
	workloadV2Id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	workloadV2UpdateRequest := *openapiclient.NewWorkloadV2UpdateRequest() // WorkloadV2UpdateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsV2API.UpdateWorkloadInternal(context.Background(), workloadV2Id).WorkloadV2UpdateRequest(workloadV2UpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsV2API.UpdateWorkloadInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateWorkloadInternal`: WorkloadV2
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsV2API.UpdateWorkloadInternal`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadV2Id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateWorkloadInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadV2UpdateRequest** | [**WorkloadV2UpdateRequest**](WorkloadV2UpdateRequest.md) |  | 

### Return type

[**WorkloadV2**](WorkloadV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateWorkloadV2

> WorkloadV2 UpdateWorkloadV2(ctx, workloadV2Id).WorkloadV2UpdateRequest(workloadV2UpdateRequest).Execute()

Update workload spec.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client"
)

func main() {
	workloadV2Id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	workloadV2UpdateRequest := *openapiclient.NewWorkloadV2UpdateRequest() // WorkloadV2UpdateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkloadsV2API.UpdateWorkloadV2(context.Background(), workloadV2Id).WorkloadV2UpdateRequest(workloadV2UpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkloadsV2API.UpdateWorkloadV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateWorkloadV2`: WorkloadV2
	fmt.Fprintf(os.Stdout, "Response from `WorkloadsV2API.UpdateWorkloadV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadV2Id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateWorkloadV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadV2UpdateRequest** | [**WorkloadV2UpdateRequest**](WorkloadV2UpdateRequest.md) |  | 

### Return type

[**WorkloadV2**](WorkloadV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

