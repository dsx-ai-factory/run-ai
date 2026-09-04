# \InternalAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetAllLogos**](InternalAPI.md#GetAllLogos) | **Get** /internal/api/v1/logos | Get all tenant logos (paginated)
[**GetTenantByNameOrAlias**](InternalAPI.md#GetTenantByNameOrAlias) | **Get** /internal/api/v1/tenants/resolve | Get tenant by name or alias
[**GetTenantDomain**](InternalAPI.md#GetTenantDomain) | **Get** /internal/api/v1/tenants/{tenantId}/domain | Get tenant domain
[**GetTenantExtendedMetricStore**](InternalAPI.md#GetTenantExtendedMetricStore) | **Get** /internal/api/v1/tenants/{tenantId}/extended-metric-store | Get tenant Metric Store with type and datasouce id
[**GetTenantMetricStore**](InternalAPI.md#GetTenantMetricStore) | **Get** /internal/api/v1/tenants/{tenantId}/metric-store | Get tenant Metric Store
[**ManagementGetTenantsInternal**](InternalAPI.md#ManagementGetTenantsInternal) | **Get** /internal/management/api/v1/tenants | Get tenants list internal



## GetAllLogos

> TenantLogosResponse GetAllLogos(ctx).Offset(offset).Limit(limit).Execute()

Get all tenant logos (paginated)

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client"
)

func main() {
	offset := int32(56) // int32 | Number of logos to skip (default 0) (optional) (default to 0)
	limit := int32(56) // int32 | Maximum number of logos to return (default 50, max 100) (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.GetAllLogos(context.Background()).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.GetAllLogos``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAllLogos`: TenantLogosResponse
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.GetAllLogos`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetAllLogosRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **offset** | **int32** | Number of logos to skip (default 0) | [default to 0]
 **limit** | **int32** | Maximum number of logos to return (default 50, max 100) | [default to 50]

### Return type

[**TenantLogosResponse**](TenantLogosResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenantByNameOrAlias

> Tenant GetTenantByNameOrAlias(ctx).TenantName(tenantName).Execute()

Get tenant by name or alias

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client"
)

func main() {
	tenantName := "tenantName_example" // string | The name or alias of the tenant to resolve

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.GetTenantByNameOrAlias(context.Background()).TenantName(tenantName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.GetTenantByNameOrAlias``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTenantByNameOrAlias`: Tenant
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.GetTenantByNameOrAlias`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantByNameOrAliasRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantName** | **string** | The name or alias of the tenant to resolve | 

### Return type

[**Tenant**](Tenant.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenantDomain

> string GetTenantDomain(ctx, tenantId).Execute()

Get tenant domain

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client"
)

func main() {
	tenantId := int32(56) // int32 | The id of the tenant to retrieve

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.GetTenantDomain(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.GetTenantDomain``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTenantDomain`: string
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.GetTenantDomain`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **int32** | The id of the tenant to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantDomainRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenantExtendedMetricStore

> ExtendedMetricStore GetTenantExtendedMetricStore(ctx, tenantId).Execute()

Get tenant Metric Store with type and datasouce id

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client"
)

func main() {
	tenantId := int32(56) // int32 | The id of the tenant to retrieve

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.GetTenantExtendedMetricStore(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.GetTenantExtendedMetricStore``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTenantExtendedMetricStore`: ExtendedMetricStore
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.GetTenantExtendedMetricStore`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **int32** | The id of the tenant to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantExtendedMetricStoreRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ExtendedMetricStore**](ExtendedMetricStore.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenantMetricStore

> MetricStore GetTenantMetricStore(ctx, tenantId).Execute()

Get tenant Metric Store

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client"
)

func main() {
	tenantId := int32(56) // int32 | The id of the tenant to retrieve

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.GetTenantMetricStore(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.GetTenantMetricStore``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTenantMetricStore`: MetricStore
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.GetTenantMetricStore`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **int32** | The id of the tenant to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantMetricStoreRequest struct via the builder pattern


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


## ManagementGetTenantsInternal

> []Tenant ManagementGetTenantsInternal(ctx).Execute()

Get tenants list internal

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.ManagementGetTenantsInternal(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.ManagementGetTenantsInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ManagementGetTenantsInternal`: []Tenant
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.ManagementGetTenantsInternal`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiManagementGetTenantsInternalRequest struct via the builder pattern


### Return type

[**[]Tenant**](Tenant.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

