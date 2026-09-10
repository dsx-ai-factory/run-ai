# \K8sObjectAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetElementObject**](K8sObjectAPI.md#GetElementObject) | **Get** /cluster-api/api/v2/workloads/{workloadId}/elements/{elementId}/object | Get the k8s object of an element within a workload
[**GetPodObject**](K8sObjectAPI.md#GetPodObject) | **Get** /cluster-api/api/v2/workloads/{workloadId}/pods/{podId}/object | Get the k8s object of a pod within a workload
[**GetWorkloadObject**](K8sObjectAPI.md#GetWorkloadObject) | **Get** /cluster-api/api/v2/workloads/{workloadId}/object | Get the k8s object of a workload



## GetElementObject

> Object GetElementObject(ctx, workloadId, elementId).Execute()

Get the k8s object of an element within a workload



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
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier (UUID) of the workload.
	elementId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier (UUID) of the element within the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.K8sObjectAPI.GetElementObject(context.Background(), workloadId, elementId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `K8sObjectAPI.GetElementObject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetElementObject`: Object
	fmt.Fprintf(os.Stdout, "Response from `K8sObjectAPI.GetElementObject`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The unique identifier (UUID) of the workload. | 
**elementId** | **string** | The unique identifier (UUID) of the element within the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetElementObjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Object**](Object.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPodObject

> Object GetPodObject(ctx, workloadId, podId).Execute()

Get the k8s object of a pod within a workload



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
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier (UUID) of the workload.
	podId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier (UUID) of the pod within the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.K8sObjectAPI.GetPodObject(context.Background(), workloadId, podId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `K8sObjectAPI.GetPodObject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPodObject`: Object
	fmt.Fprintf(os.Stdout, "Response from `K8sObjectAPI.GetPodObject`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The unique identifier (UUID) of the workload. | 
**podId** | **string** | The unique identifier (UUID) of the pod within the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPodObjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Object**](Object.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadObject

> Object GetWorkloadObject(ctx, workloadId).Execute()

Get the k8s object of a workload



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
	workloadId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier (UUID) of the workload.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.K8sObjectAPI.GetWorkloadObject(context.Background(), workloadId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `K8sObjectAPI.GetWorkloadObject``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadObject`: Object
	fmt.Fprintf(os.Stdout, "Response from `K8sObjectAPI.GetWorkloadObject`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**workloadId** | **string** | The unique identifier (UUID) of the workload. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadObjectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Object**](Object.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

