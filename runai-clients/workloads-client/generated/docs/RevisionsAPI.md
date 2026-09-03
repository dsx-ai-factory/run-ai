# \RevisionsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CountInferenceWorkloadRevisions**](RevisionsAPI.md#CountInferenceWorkloadRevisions) | **Get** /api/v1/workloads/inferences/{workloadId}/revisions/count | Get inference workload revisions count.
[**DeleteRevision**](RevisionsAPI.md#DeleteRevision) | **Delete** /api/v1/workloads/inferences/revisions/{revisionId} | Delete a revision.
[**GetInferenceWorkloadRevisions**](RevisionsAPI.md#GetInferenceWorkloadRevisions) | **Get** /api/v1/workloads/inferences/{workloadId}/revisions | Get inference workload revisions by id.
[**GetRevision**](RevisionsAPI.md#GetRevision) | **Get** /api/v1/workloads/inferences/revisions/{revisionId} | Get revision data.
[**SyncRevisions**](RevisionsAPI.md#SyncRevisions) | **Post** /api/v1/workloads/inferences/revisions/sync | Sync revisions.



## CountInferenceWorkloadRevisions

> CountWorkloads200Response CountInferenceWorkloadRevisions(ctx, workloadId).Deleted(deleted).Execute()

Get inference workload revisions count.



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
	deleted := true // bool | Return only deleted resources when `true`. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RevisionsAPI.CountInferenceWorkloadRevisions(context.Background(), workloadId).Deleted(deleted).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RevisionsAPI.CountInferenceWorkloadRevisions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountInferenceWorkloadRevisions`: CountWorkloads200Response
	fmt.Fprintf(os.Stdout, "Response from `RevisionsAPI.CountInferenceWorkloadRevisions`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCountInferenceWorkloadRevisionsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deleted** | **bool** | Return only deleted resources when &#x60;true&#x60;. | 

### Return type

[**CountWorkloads200Response**](CountWorkloads200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRevision

> DeleteRevision(ctx, revisionId).Execute()

Delete a revision.



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
	revisionId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the revision.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RevisionsAPI.DeleteRevision(context.Background(), revisionId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RevisionsAPI.DeleteRevision``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**revisionId** | **string** | The  Universally Unique Identifier (UUID) of the revision. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRevisionRequest struct via the builder pattern


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


## GetInferenceWorkloadRevisions

> GetInferenceWorkloadRevisions200Response GetInferenceWorkloadRevisions(ctx, workloadId).Deleted(deleted).Execute()

Get inference workload revisions by id.



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
	deleted := true // bool | Return only deleted resources when `true`. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RevisionsAPI.GetInferenceWorkloadRevisions(context.Background(), workloadId).Deleted(deleted).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RevisionsAPI.GetInferenceWorkloadRevisions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInferenceWorkloadRevisions`: GetInferenceWorkloadRevisions200Response
	fmt.Fprintf(os.Stdout, "Response from `RevisionsAPI.GetInferenceWorkloadRevisions`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The  Universally Unique Identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInferenceWorkloadRevisionsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deleted** | **bool** | Return only deleted resources when &#x60;true&#x60;. | 

### Return type

[**GetInferenceWorkloadRevisions200Response**](GetInferenceWorkloadRevisions200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRevision

> Revision GetRevision(ctx, revisionId).Execute()

Get revision data.



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
	revisionId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The  Universally Unique Identifier (UUID) of the revision.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RevisionsAPI.GetRevision(context.Background(), revisionId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RevisionsAPI.GetRevision``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRevision`: Revision
	fmt.Fprintf(os.Stdout, "Response from `RevisionsAPI.GetRevision`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**revisionId** | **string** | The  Universally Unique Identifier (UUID) of the revision. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRevisionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Revision**](Revision.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncRevisions

> SyncRevisions(ctx).RevisionsSyncRequest(revisionsSyncRequest).Execute()

Sync revisions.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	revisionsSyncRequest := *openapiclient.NewRevisionsSyncRequest("1", []openapiclient.RevisionSyncFields{*openapiclient.NewRevisionSyncFields("runai-revision", "my-workload-00001", "Id_example", "WorkloadId_example", "71f69d83-ba66-4822-adf5-55ce55efd210", "1", "1", time.Now(), openapiclient.Phase("Creating"))}) // RevisionsSyncRequest | The revisions to sync. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RevisionsAPI.SyncRevisions(context.Background()).RevisionsSyncRequest(revisionsSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RevisionsAPI.SyncRevisions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncRevisionsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **revisionsSyncRequest** | [**RevisionsSyncRequest**](RevisionsSyncRequest.md) | The revisions to sync. | 

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

