# \TenantsAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateTenantInternal**](TenantsAPI.md#CreateTenantInternal) | **Post** /internal/api/v1/tenants | Create tenant
[**DeleteTenantByIdInternal**](TenantsAPI.md#DeleteTenantByIdInternal) | **Delete** /internal/api/v1/tenants/{tenantId} | soft/hard delete tenant
[**GetTenant**](TenantsAPI.md#GetTenant) | **Get** /internal/api/v1/tenants | Get tenant
[**GetTenantByIdInternal**](TenantsAPI.md#GetTenantByIdInternal) | **Get** /internal/api/v1/tenants/{tenantId} | Get tenant by id



## CreateTenantInternal

> TenantCreationResponse CreateTenantInternal(ctx).TenantCreationRequest(tenantCreationRequest).Execute()

Create tenant

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
)

func main() {
	tenantCreationRequest := *openapiclient.NewTenantCreationRequest("name", "admin@example.com") // TenantCreationRequest | The tenant to create.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantsAPI.CreateTenantInternal(context.Background()).TenantCreationRequest(tenantCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantsAPI.CreateTenantInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateTenantInternal`: TenantCreationResponse
	fmt.Fprintf(os.Stdout, "Response from `TenantsAPI.CreateTenantInternal`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateTenantInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantCreationRequest** | [**TenantCreationRequest**](TenantCreationRequest.md) | The tenant to create. | 

### Return type

[**TenantCreationResponse**](TenantCreationResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteTenantByIdInternal

> DeleteTenantById200Response DeleteTenantByIdInternal(ctx, tenantId).IsHardDelete(isHardDelete).Execute()

soft/hard delete tenant

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
)

func main() {
	tenantId := int32(56) // int32 | The id of the tenant to retrieve
	isHardDelete := true // bool | Boolean indicator for soft/hard delete (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantsAPI.DeleteTenantByIdInternal(context.Background(), tenantId).IsHardDelete(isHardDelete).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantsAPI.DeleteTenantByIdInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteTenantByIdInternal`: DeleteTenantById200Response
	fmt.Fprintf(os.Stdout, "Response from `TenantsAPI.DeleteTenantByIdInternal`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **int32** | The id of the tenant to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteTenantByIdInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **isHardDelete** | **bool** | Boolean indicator for soft/hard delete | 

### Return type

[**DeleteTenantById200Response**](DeleteTenantById200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenant

> Tenant GetTenant(ctx).TenantName(tenantName).Execute()

Get tenant

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
)

func main() {
	tenantName := "tenantName_example" // string | The name of the tenant to retrieve

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantsAPI.GetTenant(context.Background()).TenantName(tenantName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantsAPI.GetTenant``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTenant`: Tenant
	fmt.Fprintf(os.Stdout, "Response from `TenantsAPI.GetTenant`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantName** | **string** | The name of the tenant to retrieve | 

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


## GetTenantByIdInternal

> Tenant GetTenantByIdInternal(ctx, tenantId).Execute()

Get tenant by id

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
)

func main() {
	tenantId := int32(56) // int32 | The id of the tenant to retrieve

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantsAPI.GetTenantByIdInternal(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantsAPI.GetTenantByIdInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTenantByIdInternal`: Tenant
	fmt.Fprintf(os.Stdout, "Response from `TenantsAPI.GetTenantByIdInternal`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **int32** | The id of the tenant to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantByIdInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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

