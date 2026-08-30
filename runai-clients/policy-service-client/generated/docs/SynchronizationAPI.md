# \SynchronizationAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SyncDistributedInferencePolicies**](SynchronizationAPI.md#SyncDistributedInferencePolicies) | **Get** /api/v2/policy/distributed-inferences/sync | Get distributed inference policies that needs to be synced to the cluster
[**SyncDistributedPolicies**](SynchronizationAPI.md#SyncDistributedPolicies) | **Get** /api/v2/policy/distributed/sync | Get distributed policies that needs to be synced to the cluster
[**SyncInferencePolicies**](SynchronizationAPI.md#SyncInferencePolicies) | **Get** /api/v2/policy/inferences/sync | Get inference policies that needs to be synced to the cluster
[**SyncRequestDistributedInferencePolicies**](SynchronizationAPI.md#SyncRequestDistributedInferencePolicies) | **Post** /api/v2/policy/distributed-inferences/sync | Request distributed inference policies that needs to be synced to the cluster and allow the cluster to report success/failure of previous updates
[**SyncRequestDistributedPolicies**](SynchronizationAPI.md#SyncRequestDistributedPolicies) | **Post** /api/v2/policy/distributed/sync | Request distributed policies that needs to be synced to the cluster and allow the cluster to report success/failure of previous updates
[**SyncRequestInferencePolicies**](SynchronizationAPI.md#SyncRequestInferencePolicies) | **Post** /api/v2/policy/inferences/sync | Request inference policies that needs to be synced to the cluster and allow the cluster to report success/failure of previous updates
[**SyncRequestTrainingPolicies**](SynchronizationAPI.md#SyncRequestTrainingPolicies) | **Post** /api/v2/policy/trainings/sync | Request training policies that needs to be synced to the cluster and allow the cluster to report success/failure of previous updates
[**SyncRequestWorkspacePolicies**](SynchronizationAPI.md#SyncRequestWorkspacePolicies) | **Post** /api/v2/policy/workspaces/sync | Request workspace policies that needs to be synced to the cluster and allow the cluster to report success/failure of previous updates
[**SyncTrainingPolicies**](SynchronizationAPI.md#SyncTrainingPolicies) | **Get** /api/v2/policy/trainings/sync | Get training policies that needs to be synced to the cluster
[**SyncWorkspacePolicies**](SynchronizationAPI.md#SyncWorkspacePolicies) | **Get** /api/v2/policy/workspaces/sync | Get workspace policies that needs to be synced to the cluster
[**UpdateSyncStatusDistributed**](SynchronizationAPI.md#UpdateSyncStatusDistributed) | **Put** /api/v2/policy/distributed/sync/status | Update the sync status of a distributed policy
[**UpdateSyncStatusDistributedInference**](SynchronizationAPI.md#UpdateSyncStatusDistributedInference) | **Put** /api/v2/policy/distributed-inferences/sync/status | Update the sync status of a distributed inference policy
[**UpdateSyncStatusInference**](SynchronizationAPI.md#UpdateSyncStatusInference) | **Put** /api/v2/policy/inferences/sync/status | Update the sync status of an inference policy
[**UpdateSyncStatusTraining**](SynchronizationAPI.md#UpdateSyncStatusTraining) | **Put** /api/v2/policy/trainings/sync/status | Update the sync status of a training policy
[**UpdateSyncStatusWorkspace**](SynchronizationAPI.md#UpdateSyncStatusWorkspace) | **Put** /api/v2/policy/workspaces/sync/status | Update the sync status of a workspace policy



## SyncDistributedInferencePolicies

> PolicySyncResponse SyncDistributedInferencePolicies(ctx).ClusterId(clusterId).After(after).Execute()

Get distributed inference policies that needs to be synced to the cluster

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	after := int64(123) // int64 | search for updates with update counter greater than the value passed for this parameter (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.SyncDistributedInferencePolicies(context.Background()).ClusterId(clusterId).After(after).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.SyncDistributedInferencePolicies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncDistributedInferencePolicies`: PolicySyncResponse
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.SyncDistributedInferencePolicies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncDistributedInferencePoliciesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **after** | **int64** | search for updates with update counter greater than the value passed for this parameter | 

### Return type

[**PolicySyncResponse**](PolicySyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDistributedPolicies

> PolicySyncResponse SyncDistributedPolicies(ctx).ClusterId(clusterId).After(after).Execute()

Get distributed policies that needs to be synced to the cluster

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	after := int64(123) // int64 | search for updates with update counter greater than the value passed for this parameter (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.SyncDistributedPolicies(context.Background()).ClusterId(clusterId).After(after).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.SyncDistributedPolicies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncDistributedPolicies`: PolicySyncResponse
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.SyncDistributedPolicies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncDistributedPoliciesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **after** | **int64** | search for updates with update counter greater than the value passed for this parameter | 

### Return type

[**PolicySyncResponse**](PolicySyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncInferencePolicies

> PolicySyncResponse SyncInferencePolicies(ctx).ClusterId(clusterId).After(after).Execute()

Get inference policies that needs to be synced to the cluster

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	after := int64(123) // int64 | search for updates with update counter greater than the value passed for this parameter (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.SyncInferencePolicies(context.Background()).ClusterId(clusterId).After(after).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.SyncInferencePolicies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncInferencePolicies`: PolicySyncResponse
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.SyncInferencePolicies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncInferencePoliciesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **after** | **int64** | search for updates with update counter greater than the value passed for this parameter | 

### Return type

[**PolicySyncResponse**](PolicySyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncRequestDistributedInferencePolicies

> PolicySyncResponse SyncRequestDistributedInferencePolicies(ctx).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()

Request distributed inference policies that needs to be synced to the cluster and allow the cluster to report success/failure of previous updates

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	after := int64(123) // int64 | search for updates with update counter greater than the value passed for this parameter (optional)
	clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.SyncRequestDistributedInferencePolicies(context.Background()).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.SyncRequestDistributedInferencePolicies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncRequestDistributedInferencePolicies`: PolicySyncResponse
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.SyncRequestDistributedInferencePolicies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncRequestDistributedInferencePoliciesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **after** | **int64** | search for updates with update counter greater than the value passed for this parameter | 
 **clusterSyncReport** | [**ClusterSyncReport**](ClusterSyncReport.md) |  | 

### Return type

[**PolicySyncResponse**](PolicySyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncRequestDistributedPolicies

> PolicySyncResponse SyncRequestDistributedPolicies(ctx).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()

Request distributed policies that needs to be synced to the cluster and allow the cluster to report success/failure of previous updates

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	after := int64(123) // int64 | search for updates with update counter greater than the value passed for this parameter (optional)
	clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.SyncRequestDistributedPolicies(context.Background()).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.SyncRequestDistributedPolicies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncRequestDistributedPolicies`: PolicySyncResponse
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.SyncRequestDistributedPolicies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncRequestDistributedPoliciesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **after** | **int64** | search for updates with update counter greater than the value passed for this parameter | 
 **clusterSyncReport** | [**ClusterSyncReport**](ClusterSyncReport.md) |  | 

### Return type

[**PolicySyncResponse**](PolicySyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncRequestInferencePolicies

> PolicySyncResponse SyncRequestInferencePolicies(ctx).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()

Request inference policies that needs to be synced to the cluster and allow the cluster to report success/failure of previous updates

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	after := int64(123) // int64 | search for updates with update counter greater than the value passed for this parameter (optional)
	clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.SyncRequestInferencePolicies(context.Background()).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.SyncRequestInferencePolicies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncRequestInferencePolicies`: PolicySyncResponse
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.SyncRequestInferencePolicies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncRequestInferencePoliciesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **after** | **int64** | search for updates with update counter greater than the value passed for this parameter | 
 **clusterSyncReport** | [**ClusterSyncReport**](ClusterSyncReport.md) |  | 

### Return type

[**PolicySyncResponse**](PolicySyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncRequestTrainingPolicies

> PolicySyncResponse SyncRequestTrainingPolicies(ctx).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()

Request training policies that needs to be synced to the cluster and allow the cluster to report success/failure of previous updates

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	after := int64(123) // int64 | search for updates with update counter greater than the value passed for this parameter (optional)
	clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.SyncRequestTrainingPolicies(context.Background()).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.SyncRequestTrainingPolicies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncRequestTrainingPolicies`: PolicySyncResponse
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.SyncRequestTrainingPolicies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncRequestTrainingPoliciesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **after** | **int64** | search for updates with update counter greater than the value passed for this parameter | 
 **clusterSyncReport** | [**ClusterSyncReport**](ClusterSyncReport.md) |  | 

### Return type

[**PolicySyncResponse**](PolicySyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncRequestWorkspacePolicies

> PolicySyncResponse SyncRequestWorkspacePolicies(ctx).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()

Request workspace policies that needs to be synced to the cluster and allow the cluster to report success/failure of previous updates

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	after := int64(123) // int64 | search for updates with update counter greater than the value passed for this parameter (optional)
	clusterSyncReport := *openapiclient.NewClusterSyncReport() // ClusterSyncReport |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.SyncRequestWorkspacePolicies(context.Background()).ClusterId(clusterId).After(after).ClusterSyncReport(clusterSyncReport).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.SyncRequestWorkspacePolicies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncRequestWorkspacePolicies`: PolicySyncResponse
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.SyncRequestWorkspacePolicies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncRequestWorkspacePoliciesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **after** | **int64** | search for updates with update counter greater than the value passed for this parameter | 
 **clusterSyncReport** | [**ClusterSyncReport**](ClusterSyncReport.md) |  | 

### Return type

[**PolicySyncResponse**](PolicySyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncTrainingPolicies

> PolicySyncResponse SyncTrainingPolicies(ctx).ClusterId(clusterId).After(after).Execute()

Get training policies that needs to be synced to the cluster

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	after := int64(123) // int64 | search for updates with update counter greater than the value passed for this parameter (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.SyncTrainingPolicies(context.Background()).ClusterId(clusterId).After(after).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.SyncTrainingPolicies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncTrainingPolicies`: PolicySyncResponse
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.SyncTrainingPolicies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncTrainingPoliciesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **after** | **int64** | search for updates with update counter greater than the value passed for this parameter | 

### Return type

[**PolicySyncResponse**](PolicySyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncWorkspacePolicies

> PolicySyncResponse SyncWorkspacePolicies(ctx).ClusterId(clusterId).After(after).Execute()

Get workspace policies that needs to be synced to the cluster

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	after := int64(123) // int64 | search for updates with update counter greater than the value passed for this parameter (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SynchronizationAPI.SyncWorkspacePolicies(context.Background()).ClusterId(clusterId).After(after).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.SyncWorkspacePolicies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncWorkspacePolicies`: PolicySyncResponse
	fmt.Fprintf(os.Stdout, "Response from `SynchronizationAPI.SyncWorkspacePolicies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncWorkspacePoliciesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **after** | **int64** | search for updates with update counter greater than the value passed for this parameter | 

### Return type

[**PolicySyncResponse**](PolicySyncResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateSyncStatusDistributed

> UpdateSyncStatusDistributed(ctx).SyncStatusUpdate(syncStatusUpdate).Execute()

Update the sync status of a distributed policy



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	syncStatusUpdate := *openapiclient.NewSyncStatusUpdate() // SyncStatusUpdate |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SynchronizationAPI.UpdateSyncStatusDistributed(context.Background()).SyncStatusUpdate(syncStatusUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.UpdateSyncStatusDistributed``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateSyncStatusDistributedRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncStatusUpdate** | [**SyncStatusUpdate**](SyncStatusUpdate.md) |  | 

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


## UpdateSyncStatusDistributedInference

> UpdateSyncStatusDistributedInference(ctx).SyncStatusUpdate(syncStatusUpdate).Execute()

Update the sync status of a distributed inference policy



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	syncStatusUpdate := *openapiclient.NewSyncStatusUpdate() // SyncStatusUpdate |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SynchronizationAPI.UpdateSyncStatusDistributedInference(context.Background()).SyncStatusUpdate(syncStatusUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.UpdateSyncStatusDistributedInference``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateSyncStatusDistributedInferenceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncStatusUpdate** | [**SyncStatusUpdate**](SyncStatusUpdate.md) |  | 

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


## UpdateSyncStatusInference

> UpdateSyncStatusInference(ctx).SyncStatusUpdate(syncStatusUpdate).Execute()

Update the sync status of an inference policy



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	syncStatusUpdate := *openapiclient.NewSyncStatusUpdate() // SyncStatusUpdate |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SynchronizationAPI.UpdateSyncStatusInference(context.Background()).SyncStatusUpdate(syncStatusUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.UpdateSyncStatusInference``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateSyncStatusInferenceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncStatusUpdate** | [**SyncStatusUpdate**](SyncStatusUpdate.md) |  | 

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


## UpdateSyncStatusTraining

> UpdateSyncStatusTraining(ctx).SyncStatusUpdate(syncStatusUpdate).Execute()

Update the sync status of a training policy



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	syncStatusUpdate := *openapiclient.NewSyncStatusUpdate() // SyncStatusUpdate |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SynchronizationAPI.UpdateSyncStatusTraining(context.Background()).SyncStatusUpdate(syncStatusUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.UpdateSyncStatusTraining``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateSyncStatusTrainingRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncStatusUpdate** | [**SyncStatusUpdate**](SyncStatusUpdate.md) |  | 

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


## UpdateSyncStatusWorkspace

> UpdateSyncStatusWorkspace(ctx).SyncStatusUpdate(syncStatusUpdate).Execute()

Update the sync status of a workspace policy



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	syncStatusUpdate := *openapiclient.NewSyncStatusUpdate() // SyncStatusUpdate |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SynchronizationAPI.UpdateSyncStatusWorkspace(context.Background()).SyncStatusUpdate(syncStatusUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SynchronizationAPI.UpdateSyncStatusWorkspace``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateSyncStatusWorkspaceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncStatusUpdate** | [**SyncStatusUpdate**](SyncStatusUpdate.md) |  | 

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

