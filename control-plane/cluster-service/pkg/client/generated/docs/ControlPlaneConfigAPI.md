# \ControlPlaneConfigAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetClusterControlPlaneConfig**](ControlPlaneConfigAPI.md#GetClusterControlPlaneConfig) | **Get** /api/v1/clusters/{clusterUuid}/config/control-plane | Retrieve the control plane configuration of a Cluster.



## GetClusterControlPlaneConfig

> ClusterControlPlaneConfig GetClusterControlPlaneConfig(ctx, clusterUuid).Execute()

Retrieve the control plane configuration of a Cluster.



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
	resp, r, err := apiClient.ControlPlaneConfigAPI.GetClusterControlPlaneConfig(context.Background(), clusterUuid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ControlPlaneConfigAPI.GetClusterControlPlaneConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterControlPlaneConfig`: ClusterControlPlaneConfig
	fmt.Fprintf(os.Stdout, "Response from `ControlPlaneConfigAPI.GetClusterControlPlaneConfig`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterControlPlaneConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ClusterControlPlaneConfig**](ClusterControlPlaneConfig.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

