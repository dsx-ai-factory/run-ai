# \DistributedInferencesAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateDistributedInference**](DistributedInferencesAPI.md#CreateDistributedInference) | **Post** /api/v1/workloads/distributed-inferences | Create a distributed inference.
[**DeleteDistributedInference**](DistributedInferencesAPI.md#DeleteDistributedInference) | **Delete** /api/v1/workloads/distributed-inferences/{workloadId} | Delete a distributed inference.
[**GetDistributedInference**](DistributedInferencesAPI.md#GetDistributedInference) | **Get** /api/v1/workloads/distributed-inferences/{workloadId} | Get a distributed inference data.
[**ListDistributedInferencesUpdates**](DistributedInferencesAPI.md#ListDistributedInferencesUpdates) | **Get** /api/v1/workloads/distributed-inferences/updates | List distributed inference updates.
[**RequestDistributedInferencesUpdates**](DistributedInferencesAPI.md#RequestDistributedInferencesUpdates) | **Post** /api/v1/workloads/distributed-inferences/updates | Request distributed inference updates
[**UpdateDistributedInferenceSpec**](DistributedInferencesAPI.md#UpdateDistributedInferenceSpec) | **Patch** /api/v1/workloads/distributed-inferences/{workloadId} | Update distributed inference spec.



## CreateDistributedInference

> DistributedInference CreateDistributedInference(ctx).DistributedInferenceCreationRequest(distributedInferenceCreationRequest).Execute()

Create a distributed inference.



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
	distributedInferenceCreationRequest := *openapiclient.NewDistributedInferenceCreationRequest("my-workload-name", "1", "71f69d83-ba66-4822-adf5-55ce55efd210") // DistributedInferenceCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DistributedInferencesAPI.CreateDistributedInference(context.Background()).DistributedInferenceCreationRequest(distributedInferenceCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DistributedInferencesAPI.CreateDistributedInference``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateDistributedInference`: DistributedInference
	fmt.Fprintf(os.Stdout, "Response from `DistributedInferencesAPI.CreateDistributedInference`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateDistributedInferenceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **distributedInferenceCreationRequest** | [**DistributedInferenceCreationRequest**](DistributedInferenceCreationRequest.md) |  | 

### Return type

[**DistributedInference**](DistributedInference.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteDistributedInference

> HttpResponse DeleteDistributedInference(ctx, workloadId).Execute()

Delete a distributed inference.



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
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DistributedInferencesAPI.DeleteDistributedInference(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DistributedInferencesAPI.DeleteDistributedInference``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteDistributedInference`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `DistributedInferencesAPI.DeleteDistributedInference`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDistributedInferenceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HttpResponse**](HttpResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDistributedInference

> DistributedInference GetDistributedInference(ctx, workloadId).Execute()

Get a distributed inference data.



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
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DistributedInferencesAPI.GetDistributedInference(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DistributedInferencesAPI.GetDistributedInference``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDistributedInference`: DistributedInference
	fmt.Fprintf(os.Stdout, "Response from `DistributedInferencesAPI.GetDistributedInference`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDistributedInferenceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DistributedInference**](DistributedInference.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListDistributedInferencesUpdates

> WorkloadsSyncResponse ListDistributedInferencesUpdates(ctx).UpdatesAfter(updatesAfter).ClusterId(clusterId).Execute()

List distributed inference updates.



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
	updatesAfter := int64(12345) // int64 | list only data with update counter greater than the provided value. for example updatesAfter=5 should fetch data with updates counter 6 and above.
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DistributedInferencesAPI.ListDistributedInferencesUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DistributedInferencesAPI.ListDistributedInferencesUpdates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListDistributedInferencesUpdates`: WorkloadsSyncResponse
	fmt.Fprintf(os.Stdout, "Response from `DistributedInferencesAPI.ListDistributedInferencesUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListDistributedInferencesUpdatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updatesAfter** | **int64** | list only data with update counter greater than the provided value. for example updatesAfter&#x3D;5 should fetch data with updates counter 6 and above. | 
 **clusterId** | **string** | The id of the cluster | 

### Return type

[**WorkloadsSyncResponse**](WorkloadsSyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RequestDistributedInferencesUpdates

> WorkloadsSyncResponse RequestDistributedInferencesUpdates(ctx).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()

Request distributed inference updates



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
	updatesAfter := int64(12345) // int64 | list only data with update counter greater than the provided value. for example updatesAfter=5 should fetch data with updates counter 6 and above.
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DistributedInferencesAPI.RequestDistributedInferencesUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DistributedInferencesAPI.RequestDistributedInferencesUpdates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RequestDistributedInferencesUpdates`: WorkloadsSyncResponse
	fmt.Fprintf(os.Stdout, "Response from `DistributedInferencesAPI.RequestDistributedInferencesUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRequestDistributedInferencesUpdatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updatesAfter** | **int64** | list only data with update counter greater than the provided value. for example updatesAfter&#x3D;5 should fetch data with updates counter 6 and above. | 
 **clusterId** | **string** | The id of the cluster | 
 **clusterSyncReport** | [**ClusterSyncReport**](ClusterSyncReport.md) |  | 

### Return type

[**WorkloadsSyncResponse**](WorkloadsSyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDistributedInferenceSpec

> DistributedInference UpdateDistributedInferenceSpec(ctx, workloadId).UpdateRequest(updateRequest).Execute()

Update distributed inference spec.



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
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.
	updateRequest := *openapiclient.NewUpdateRequest() // UpdateRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DistributedInferencesAPI.UpdateDistributedInferenceSpec(context.Background(), workloadId).UpdateRequest(updateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DistributedInferencesAPI.UpdateDistributedInferenceSpec``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateDistributedInferenceSpec`: DistributedInference
	fmt.Fprintf(os.Stdout, "Response from `DistributedInferencesAPI.UpdateDistributedInferenceSpec`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDistributedInferenceSpecRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateRequest** | [**UpdateRequest**](UpdateRequest.md) |  | 

### Return type

[**DistributedInference**](DistributedInference.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

