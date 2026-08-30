# \InternalAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetAncestors**](InternalAPI.md#GetAncestors) | **Get** /internal/{resourceType}/{id}/ancestors | Get scoped resource ancestors.
[**GetScopeName**](InternalAPI.md#GetScopeName) | **Get** /internal/{resourceType}/{id}/name | Get scoped resource name.
[**InternalTenantsTenantIdMetricStoreGet**](InternalAPI.md#InternalTenantsTenantIdMetricStoreGet) | **Get** /internal/tenants/{tenantId}/metric-store | Get the metric store for a tenant



## GetAncestors

> Ancestors GetAncestors(ctx, resourceType, id).Execute()

Get scoped resource ancestors.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	resourceType := "resourceType_example" // string | 
	id := "id_example" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.GetAncestors(context.Background(), resourceType, id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.GetAncestors``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAncestors`: Ancestors
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.GetAncestors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**resourceType** | **string** |  | 
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetAncestorsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Ancestors**](Ancestors.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetScopeName

> GetScopeName200Response GetScopeName(ctx, resourceType, id).ClusterId(clusterId).Execute()

Get scoped resource name.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	resourceType := "resourceType_example" // string | 
	id := "id_example" // string | 
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.GetScopeName(context.Background(), resourceType, id).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.GetScopeName``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetScopeName`: GetScopeName200Response
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.GetScopeName`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**resourceType** | **string** |  | 
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetScopeNameRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 

### Return type

[**GetScopeName200Response**](GetScopeName200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InternalTenantsTenantIdMetricStoreGet

> MetricStore InternalTenantsTenantIdMetricStoreGet(ctx, tenantId).Execute()

Get the metric store for a tenant

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/backend-api-client"
)

func main() {
	tenantId := "tenantId_example" // string | Unique identifier of the tenant.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.InternalTenantsTenantIdMetricStoreGet(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.InternalTenantsTenantIdMetricStoreGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `InternalTenantsTenantIdMetricStoreGet`: MetricStore
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.InternalTenantsTenantIdMetricStoreGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Unique identifier of the tenant. | 

### Other Parameters

Other parameters are passed through a pointer to a apiInternalTenantsTenantIdMetricStoreGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**MetricStore**](MetricStore.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

