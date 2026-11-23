# \TrainingsAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteDistributedById**](TrainingsAPI.md#DeleteDistributedById) | **Delete** /v1/k8s/clusters/{clusterUuid}/distributed/{workloadId} | Delete a distributed, per cluster sync update [Obsolete]
[**DeleteTrainingById**](TrainingsAPI.md#DeleteTrainingById) | **Delete** /v1/k8s/clusters/{clusterUuid}/trainings/{workloadId} | Delete a training, per cluster sync update [Obsolete]
[**ListTrainings**](TrainingsAPI.md#ListTrainings) | **Get** /v1/k8s/clusters/{clusterUuid}/trainings | List training in the cluster
[**SyncDistributedStatus**](TrainingsAPI.md#SyncDistributedStatus) | **Post** /v1/k8s/clusters/{clusterUuid}/distributed/status | Sync distributed status [Obsolete]
[**SyncTrainingStatus**](TrainingsAPI.md#SyncTrainingStatus) | **Post** /v1/k8s/clusters/{clusterUuid}/trainings/status | Sync training status [Obsolete]
[**UpdateDistributedStatus**](TrainingsAPI.md#UpdateDistributedStatus) | **Put** /v1/k8s/clusters/{clusterUuid}/distributed/status | Update distributed status [Obsolete]
[**UpdateTrainingStatus**](TrainingsAPI.md#UpdateTrainingStatus) | **Put** /v1/k8s/clusters/{clusterUuid}/trainings/status | Update training status [Obsolete]



## DeleteDistributedById

> DeleteDistributedById(ctx, clusterUuid, workloadId).Execute()

Delete a distributed, per cluster sync update [Obsolete]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TrainingsAPI.DeleteDistributedById(context.Background(), clusterUuid, workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.DeleteDistributedById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDistributedByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



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


## DeleteTrainingById

> DeleteTrainingById(ctx, clusterUuid, workloadId).Execute()

Delete a training, per cluster sync update [Obsolete]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TrainingsAPI.DeleteTrainingById(context.Background(), clusterUuid, workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.DeleteTrainingById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteTrainingByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



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


## ListTrainings

> TrainingListResponse ListTrainings(ctx, clusterUuid).NodeId(nodeId).Filter(filter).SortBy(sortBy).SortDirection(sortDirection).From(from).Limit(limit).Execute()

List training in the cluster

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	nodeId := "nodeId_example" // string | Unique identifier of the node. (optional)
	filter := "filter_example" // string |  (optional)
	sortBy := "sortBy_example" // string | Order of the results. (optional)
	sortDirection := "sortDirection_example" // string |  (optional)
	from := int32(56) // int32 | Start the response from a given number of result. Used along with 'limit' to retrieve the results paginated. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TrainingsAPI.ListTrainings(context.Background(), clusterUuid).NodeId(nodeId).Filter(filter).SortBy(sortBy).SortDirection(sortDirection).From(from).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.ListTrainings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListTrainings`: TrainingListResponse
	fmt.Fprintf(os.Stdout, "Response from `TrainingsAPI.ListTrainings`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiListTrainingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodeId** | **string** | Unique identifier of the node. | 
 **filter** | **string** |  | 
 **sortBy** | **string** | Order of the results. | 
 **sortDirection** | **string** |  | 
 **from** | **int32** | Start the response from a given number of result. Used along with &#39;limit&#39; to retrieve the results paginated. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**TrainingListResponse**](TrainingListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDistributedStatus

> []WorkloadStatusUpdate SyncDistributedStatus(ctx, clusterUuid).WorkloadStatusUpdate(workloadStatusUpdate).Execute()

Sync distributed status [Obsolete]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	workloadStatusUpdate := []openapiclient.WorkloadStatusUpdate{*openapiclient.NewWorkloadStatusUpdate("Id_example", "Status_example")} // []WorkloadStatusUpdate |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TrainingsAPI.SyncDistributedStatus(context.Background(), clusterUuid).WorkloadStatusUpdate(workloadStatusUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.SyncDistributedStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncDistributedStatus`: []WorkloadStatusUpdate
	fmt.Fprintf(os.Stdout, "Response from `TrainingsAPI.SyncDistributedStatus`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncDistributedStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadStatusUpdate** | [**[]WorkloadStatusUpdate**](WorkloadStatusUpdate.md) |  | 

### Return type

[**[]WorkloadStatusUpdate**](WorkloadStatusUpdate.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncTrainingStatus

> []WorkloadStatusUpdate SyncTrainingStatus(ctx, clusterUuid).WorkloadStatusUpdate(workloadStatusUpdate).Execute()

Sync training status [Obsolete]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	workloadStatusUpdate := []openapiclient.WorkloadStatusUpdate{*openapiclient.NewWorkloadStatusUpdate("Id_example", "Status_example")} // []WorkloadStatusUpdate |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TrainingsAPI.SyncTrainingStatus(context.Background(), clusterUuid).WorkloadStatusUpdate(workloadStatusUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.SyncTrainingStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncTrainingStatus`: []WorkloadStatusUpdate
	fmt.Fprintf(os.Stdout, "Response from `TrainingsAPI.SyncTrainingStatus`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncTrainingStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadStatusUpdate** | [**[]WorkloadStatusUpdate**](WorkloadStatusUpdate.md) |  | 

### Return type

[**[]WorkloadStatusUpdate**](WorkloadStatusUpdate.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDistributedStatus

> WorkloadStatusUpdate UpdateDistributedStatus(ctx, clusterUuid).WorkloadStatusUpdate(workloadStatusUpdate).Execute()

Update distributed status [Obsolete]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	workloadStatusUpdate := *openapiclient.NewWorkloadStatusUpdate("Id_example", "Status_example") // WorkloadStatusUpdate |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TrainingsAPI.UpdateDistributedStatus(context.Background(), clusterUuid).WorkloadStatusUpdate(workloadStatusUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.UpdateDistributedStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateDistributedStatus`: WorkloadStatusUpdate
	fmt.Fprintf(os.Stdout, "Response from `TrainingsAPI.UpdateDistributedStatus`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDistributedStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadStatusUpdate** | [**WorkloadStatusUpdate**](WorkloadStatusUpdate.md) |  | 

### Return type

[**WorkloadStatusUpdate**](WorkloadStatusUpdate.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateTrainingStatus

> WorkloadStatusUpdate UpdateTrainingStatus(ctx, clusterUuid).WorkloadStatusUpdate(workloadStatusUpdate).Execute()

Update training status [Obsolete]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	workloadStatusUpdate := *openapiclient.NewWorkloadStatusUpdate("Id_example", "Status_example") // WorkloadStatusUpdate |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TrainingsAPI.UpdateTrainingStatus(context.Background(), clusterUuid).WorkloadStatusUpdate(workloadStatusUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.UpdateTrainingStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateTrainingStatus`: WorkloadStatusUpdate
	fmt.Fprintf(os.Stdout, "Response from `TrainingsAPI.UpdateTrainingStatus`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateTrainingStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workloadStatusUpdate** | [**WorkloadStatusUpdate**](WorkloadStatusUpdate.md) |  | 

### Return type

[**WorkloadStatusUpdate**](WorkloadStatusUpdate.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

