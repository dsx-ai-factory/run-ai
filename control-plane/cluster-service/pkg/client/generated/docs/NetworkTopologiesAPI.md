# \NetworkTopologiesAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateNetworkTopology**](NetworkTopologiesAPI.md#CreateNetworkTopology) | **Post** /api/v1/network-topologies | Create network topologies
[**DeleteNetworkTopology**](NetworkTopologiesAPI.md#DeleteNetworkTopology) | **Delete** /api/v1/network-topologies/{networkTopologyId} | Delete network topology by ID.
[**GetNetworkTopologies**](NetworkTopologiesAPI.md#GetNetworkTopologies) | **Get** /api/v1/network-topologies | Get network topologies.
[**GetNetworkTopologiesCount**](NetworkTopologiesAPI.md#GetNetworkTopologiesCount) | **Get** /api/v1/network-topologies/count | Count the amount of network topologies in a tenant.
[**GetNetworkTopologyById**](NetworkTopologiesAPI.md#GetNetworkTopologyById) | **Get** /api/v1/network-topologies/{networkTopologyId} | Get network topology by id.
[**SyncNetworkTopologies**](NetworkTopologiesAPI.md#SyncNetworkTopologies) | **Post** /api/v1/clusters/{clusterUuid}/network-topologies/sync | Sync network topologies from cluster.



## CreateNetworkTopology

> NetworkTopology CreateNetworkTopology(ctx).NetworkTopologyCreateRequest(networkTopologyCreateRequest).Execute()

Create network topologies



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
	networkTopologyCreateRequest := *openapiclient.NewNetworkTopologyCreateRequest([]string{"Levels_example"}, "default-topology", "71f69d83-ba66-4822-adf5-55ce55efd210") // NetworkTopologyCreateRequest | The network topology to create.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NetworkTopologiesAPI.CreateNetworkTopology(context.Background()).NetworkTopologyCreateRequest(networkTopologyCreateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NetworkTopologiesAPI.CreateNetworkTopology``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateNetworkTopology`: NetworkTopology
	fmt.Fprintf(os.Stdout, "Response from `NetworkTopologiesAPI.CreateNetworkTopology`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateNetworkTopologyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **networkTopologyCreateRequest** | [**NetworkTopologyCreateRequest**](NetworkTopologyCreateRequest.md) | The network topology to create. | 

### Return type

[**NetworkTopology**](NetworkTopology.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteNetworkTopology

> DeleteNetworkTopology(ctx, networkTopologyId).Execute()

Delete network topology by ID.



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
	networkTopologyId := "123e4567-e89b-12d3-a456-426614174000" // string | The unique identifier for the network topology

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NetworkTopologiesAPI.DeleteNetworkTopology(context.Background(), networkTopologyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NetworkTopologiesAPI.DeleteNetworkTopology``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**networkTopologyId** | **string** | The unique identifier for the network topology | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteNetworkTopologyRequest struct via the builder pattern


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


## GetNetworkTopologies

> GetNetworkTopologies200Response GetNetworkTopologies(ctx).FilterBy(filterBy).Execute()

Get network topologies.



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
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Supported operators:  == Equals  != Not equals  <= Less than or equal  >= Greater than or equal  =@ Contains !@ Does not contain  =^ Starts with  =$ Ends with Date values must be in ISO 8601 timestamp format. Only the following operators support dates: ==, !=, <=, >=. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NetworkTopologiesAPI.GetNetworkTopologies(context.Background()).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NetworkTopologiesAPI.GetNetworkTopologies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNetworkTopologies`: GetNetworkTopologies200Response
	fmt.Fprintf(os.Stdout, "Response from `NetworkTopologiesAPI.GetNetworkTopologies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNetworkTopologiesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Supported operators:  &#x3D;&#x3D; Equals  !&#x3D; Not equals  &lt;&#x3D; Less than or equal  &gt;&#x3D; Greater than or equal  &#x3D;@ Contains !@ Does not contain  &#x3D;^ Starts with  &#x3D;$ Ends with Date values must be in ISO 8601 timestamp format. Only the following operators support dates: &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D;, &gt;&#x3D;. | 

### Return type

[**GetNetworkTopologies200Response**](GetNetworkTopologies200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNetworkTopologiesCount

> CountResponse GetNetworkTopologiesCount(ctx).FilterBy(filterBy).Execute()

Count the amount of network topologies in a tenant.



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
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Supported operators:  == Equals  != Not equals  <= Less than or equal  >= Greater than or equal  =@ Contains !@ Does not contain  =^ Starts with  =$ Ends with Date values must be in ISO 8601 timestamp format. Only the following operators support dates: ==, !=, <=, >=. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NetworkTopologiesAPI.GetNetworkTopologiesCount(context.Background()).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NetworkTopologiesAPI.GetNetworkTopologiesCount``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNetworkTopologiesCount`: CountResponse
	fmt.Fprintf(os.Stdout, "Response from `NetworkTopologiesAPI.GetNetworkTopologiesCount`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNetworkTopologiesCountRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Supported operators:  &#x3D;&#x3D; Equals  !&#x3D; Not equals  &lt;&#x3D; Less than or equal  &gt;&#x3D; Greater than or equal  &#x3D;@ Contains !@ Does not contain  &#x3D;^ Starts with  &#x3D;$ Ends with Date values must be in ISO 8601 timestamp format. Only the following operators support dates: &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D;, &gt;&#x3D;. | 

### Return type

[**CountResponse**](CountResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNetworkTopologyById

> NetworkTopology GetNetworkTopologyById(ctx, networkTopologyId).Execute()

Get network topology by id.



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
	networkTopologyId := "123e4567-e89b-12d3-a456-426614174000" // string | The unique identifier for the network topology

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NetworkTopologiesAPI.GetNetworkTopologyById(context.Background(), networkTopologyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NetworkTopologiesAPI.GetNetworkTopologyById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNetworkTopologyById`: NetworkTopology
	fmt.Fprintf(os.Stdout, "Response from `NetworkTopologiesAPI.GetNetworkTopologyById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**networkTopologyId** | **string** | The unique identifier for the network topology | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNetworkTopologyByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**NetworkTopology**](NetworkTopology.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncNetworkTopologies

> SyncNetworkTopologies(ctx, clusterUuid).NetworkTopologySyncRequest(networkTopologySyncRequest).Execute()

Sync network topologies from cluster.

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
	networkTopologySyncRequest := *openapiclient.NewNetworkTopologySyncRequest() // NetworkTopologySyncRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NetworkTopologiesAPI.SyncNetworkTopologies(context.Background(), clusterUuid).NetworkTopologySyncRequest(networkTopologySyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NetworkTopologiesAPI.SyncNetworkTopologies``: %v\n", err)
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

Other parameters are passed through a pointer to a apiSyncNetworkTopologiesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **networkTopologySyncRequest** | [**NetworkTopologySyncRequest**](NetworkTopologySyncRequest.md) |  | 

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

