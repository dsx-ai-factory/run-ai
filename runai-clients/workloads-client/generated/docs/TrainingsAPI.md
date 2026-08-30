# \TrainingsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateTraining**](TrainingsAPI.md#CreateTraining) | **Post** /api/v1/workloads/trainings | Create a training.
[**DeleteTraining**](TrainingsAPI.md#DeleteTraining) | **Delete** /api/v1/workloads/trainings/{workloadId} | Delete a training.
[**GetTraining**](TrainingsAPI.md#GetTraining) | **Get** /api/v1/workloads/trainings/{workloadId} | Get training data.
[**ListTrainingsUpdates**](TrainingsAPI.md#ListTrainingsUpdates) | **Get** /api/v1/workloads/trainings/updates | List training updates.
[**RequestTrainingsUpdates**](TrainingsAPI.md#RequestTrainingsUpdates) | **Post** /api/v1/workloads/trainings/updates | Request training updates
[**ResumeTraining**](TrainingsAPI.md#ResumeTraining) | **Post** /api/v1/workloads/trainings/{workloadId}/resume | Resume a training.
[**SuspendTraining**](TrainingsAPI.md#SuspendTraining) | **Post** /api/v1/workloads/trainings/{workloadId}/suspend | Suspend a training.



## CreateTraining

> Training CreateTraining(ctx).TrainingCreationRequest(trainingCreationRequest).Execute()

Create a training.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/workloads-client"
)

func main() {
	trainingCreationRequest := *openapiclient.NewTrainingCreationRequest("my-workload-name", "1", "71f69d83-ba66-4822-adf5-55ce55efd210") // TrainingCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TrainingsAPI.CreateTraining(context.Background()).TrainingCreationRequest(trainingCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.CreateTraining``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateTraining`: Training
	fmt.Fprintf(os.Stdout, "Response from `TrainingsAPI.CreateTraining`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateTrainingRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trainingCreationRequest** | [**TrainingCreationRequest**](TrainingCreationRequest.md) |  | 

### Return type

[**Training**](Training.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteTraining

> HttpResponse DeleteTraining(ctx, workloadId).Execute()

Delete a training.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/workloads-client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TrainingsAPI.DeleteTraining(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.DeleteTraining``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteTraining`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `TrainingsAPI.DeleteTraining`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteTrainingRequest struct via the builder pattern


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


## GetTraining

> Training GetTraining(ctx, workloadId).Execute()

Get training data.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/workloads-client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TrainingsAPI.GetTraining(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.GetTraining``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTraining`: Training
	fmt.Fprintf(os.Stdout, "Response from `TrainingsAPI.GetTraining`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTrainingRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Training**](Training.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListTrainingsUpdates

> WorkloadsSyncResponse ListTrainingsUpdates(ctx).UpdatesAfter(updatesAfter).ClusterId(clusterId).Execute()

List training updates.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/workloads-client"
)

func main() {
	updatesAfter := int64(12345) // int64 | list only data with update counter greater than the provided value. for example updatesAfter=5 should fetch data with updates counter 6 and above.
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TrainingsAPI.ListTrainingsUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.ListTrainingsUpdates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListTrainingsUpdates`: WorkloadsSyncResponse
	fmt.Fprintf(os.Stdout, "Response from `TrainingsAPI.ListTrainingsUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListTrainingsUpdatesRequest struct via the builder pattern


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


## RequestTrainingsUpdates

> WorkloadsSyncResponse RequestTrainingsUpdates(ctx).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()

Request training updates



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/workloads-client"
)

func main() {
	updatesAfter := int64(12345) // int64 | list only data with update counter greater than the provided value. for example updatesAfter=5 should fetch data with updates counter 6 and above.
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TrainingsAPI.RequestTrainingsUpdates(context.Background()).UpdatesAfter(updatesAfter).ClusterId(clusterId).ClusterSyncReport(clusterSyncReport).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.RequestTrainingsUpdates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RequestTrainingsUpdates`: WorkloadsSyncResponse
	fmt.Fprintf(os.Stdout, "Response from `TrainingsAPI.RequestTrainingsUpdates`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRequestTrainingsUpdatesRequest struct via the builder pattern


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


## ResumeTraining

> HttpResponse ResumeTraining(ctx, workloadId).Execute()

Resume a training.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/workloads-client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TrainingsAPI.ResumeTraining(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.ResumeTraining``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ResumeTraining`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `TrainingsAPI.ResumeTraining`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiResumeTrainingRequest struct via the builder pattern


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


## SuspendTraining

> HttpResponse SuspendTraining(ctx, workloadId).Execute()

Suspend a training.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/workloads-client"
)

func main() {
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TrainingsAPI.SuspendTraining(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TrainingsAPI.SuspendTraining``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SuspendTraining`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `TrainingsAPI.SuspendTraining`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSuspendTrainingRequest struct via the builder pattern


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

