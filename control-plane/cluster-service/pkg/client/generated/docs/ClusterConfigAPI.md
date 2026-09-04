# \ClusterConfigAPI

All URIs are relative to *https://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetClusterConfig**](ClusterConfigAPI.md#GetClusterConfig) | **Get** /api/v1/clusters/{clusterUuid}/config | Get the cluster configuration.
[**GetInternalClusterConfig**](ClusterConfigAPI.md#GetInternalClusterConfig) | **Get** /api/v1/clusters/{clusterUuid}/internal-config | Get the internal configuration of a cluster. This API is intended to be used by RunAI components on the cluster.
[**PatchClusterConfig**](ClusterConfigAPI.md#PatchClusterConfig) | **Patch** /api/v1/clusters/{clusterUuid}/config | Patch the cluster configuration.
[**PutClusterConfig**](ClusterConfigAPI.md#PutClusterConfig) | **Put** /api/v1/clusters/{clusterUuid}/config | Put the cluster configuration.
[**PutClusterConfigStatus**](ClusterConfigAPI.md#PutClusterConfigStatus) | **Put** /api/v1/clusters/{clusterUuid}/config/status | Put the status of the cluster configuration. This endpoint will be used by the cluster to sync its remote configuration status.



## GetClusterConfig

> GetPutPatchClusterConfigResponse GetClusterConfig(ctx, clusterUuid).Execute()

Get the cluster configuration.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClusterConfigAPI.GetClusterConfig(context.Background(), clusterUuid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterConfigAPI.GetClusterConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterConfig`: GetPutPatchClusterConfigResponse
	fmt.Fprintf(os.Stdout, "Response from `ClusterConfigAPI.GetClusterConfig`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GetPutPatchClusterConfigResponse**](GetPutPatchClusterConfigResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInternalClusterConfig

> GetInternalClusterConfigResponse GetInternalClusterConfig(ctx, clusterUuid).Execute()

Get the internal configuration of a cluster. This API is intended to be used by RunAI components on the cluster.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClusterConfigAPI.GetInternalClusterConfig(context.Background(), clusterUuid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterConfigAPI.GetInternalClusterConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInternalClusterConfig`: GetInternalClusterConfigResponse
	fmt.Fprintf(os.Stdout, "Response from `ClusterConfigAPI.GetInternalClusterConfig`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInternalClusterConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GetInternalClusterConfigResponse**](GetInternalClusterConfigResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PatchClusterConfig

> GetPutPatchClusterConfigResponse PatchClusterConfig(ctx, clusterUuid).PatchClusterConfigRequest(patchClusterConfigRequest).Execute()

Patch the cluster configuration.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	patchClusterConfigRequest := *openapiclient.NewPatchClusterConfigRequest() // PatchClusterConfigRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClusterConfigAPI.PatchClusterConfig(context.Background(), clusterUuid).PatchClusterConfigRequest(patchClusterConfigRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterConfigAPI.PatchClusterConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PatchClusterConfig`: GetPutPatchClusterConfigResponse
	fmt.Fprintf(os.Stdout, "Response from `ClusterConfigAPI.PatchClusterConfig`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchClusterConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **patchClusterConfigRequest** | [**PatchClusterConfigRequest**](PatchClusterConfigRequest.md) |  | 

### Return type

[**GetPutPatchClusterConfigResponse**](GetPutPatchClusterConfigResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PutClusterConfig

> GetPutPatchClusterConfigResponse PutClusterConfig(ctx, clusterUuid).PutClusterConfigRequest(putClusterConfigRequest).Execute()

Put the cluster configuration.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	putClusterConfigRequest := *openapiclient.NewPutClusterConfigRequest() // PutClusterConfigRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ClusterConfigAPI.PutClusterConfig(context.Background(), clusterUuid).PutClusterConfigRequest(putClusterConfigRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterConfigAPI.PutClusterConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PutClusterConfig`: GetPutPatchClusterConfigResponse
	fmt.Fprintf(os.Stdout, "Response from `ClusterConfigAPI.PutClusterConfig`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiPutClusterConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **putClusterConfigRequest** | [**PutClusterConfigRequest**](PutClusterConfigRequest.md) |  | 

### Return type

[**GetPutPatchClusterConfigResponse**](GetPutPatchClusterConfigResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PutClusterConfigStatus

> PutClusterConfigStatus(ctx, clusterUuid).PutClusterConfigStatusRequest(putClusterConfigStatusRequest).Execute()

Put the status of the cluster configuration. This endpoint will be used by the cluster to sync its remote configuration status.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	putClusterConfigStatusRequest := *openapiclient.NewPutClusterConfigStatusRequest(*openapiclient.NewPutClusterConfigStatusRequestStatus([]openapiclient.Condition{*openapiclient.NewCondition(time.Now(), "Message_example", "Reason_example", "Status_example", "Type_example")})) // PutClusterConfigStatusRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ClusterConfigAPI.PutClusterConfigStatus(context.Background(), clusterUuid).PutClusterConfigStatusRequest(putClusterConfigStatusRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterConfigAPI.PutClusterConfigStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiPutClusterConfigStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **putClusterConfigStatusRequest** | [**PutClusterConfigStatusRequest**](PutClusterConfigStatusRequest.md) |  | 

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

