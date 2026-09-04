# \TenantAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetTenant**](TenantAPI.md#GetTenant) | **Get** /v1/k8s/tenant | Get tenant
[**GetTenantConfig**](TenantAPI.md#GetTenantConfig) | **Get** /v1/k8s/config | Get tenant auth config
[**GetTenantSettings**](TenantAPI.md#GetTenantSettings) | **Get** /v1/k8s/setting | Get all tenant settings.
[**GetTenantSettingsByKey**](TenantAPI.md#GetTenantSettingsByKey) | **Get** /v1/k8s/setting/{settingKey} | Get a tenant setting by key.
[**UpdateTenantSetting**](TenantAPI.md#UpdateTenantSetting) | **Put** /v1/k8s/setting | Update a tenant setting.



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
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/backend-api-client"
)

func main() {
	tenantName := "tenantName_example" // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantAPI.GetTenant(context.Background()).TenantName(tenantName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantAPI.GetTenant``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTenant`: Tenant
	fmt.Fprintf(os.Stdout, "Response from `TenantAPI.GetTenant`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantName** | **string** |  | 

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


## GetTenantConfig

> Config GetTenantConfig(ctx).TenantName(tenantName).Execute()

Get tenant auth config



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/backend-api-client"
)

func main() {
	tenantName := "tenantName_example" // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantAPI.GetTenantConfig(context.Background()).TenantName(tenantName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantAPI.GetTenantConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTenantConfig`: Config
	fmt.Fprintf(os.Stdout, "Response from `TenantAPI.GetTenantConfig`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantName** | **string** |  | 

### Return type

[**Config**](Config.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenantSettings

> []Setting GetTenantSettings(ctx).ExcludeMigrated(excludeMigrated).Execute()

Get all tenant settings.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/backend-api-client"
)

func main() {
	excludeMigrated := true // bool | When true, excludes settings that have been migrated to external services (e.g., Identity Manager). (optional) (default to false)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantAPI.GetTenantSettings(context.Background()).ExcludeMigrated(excludeMigrated).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantAPI.GetTenantSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTenantSettings`: []Setting
	fmt.Fprintf(os.Stdout, "Response from `TenantAPI.GetTenantSettings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **excludeMigrated** | **bool** | When true, excludes settings that have been migrated to external services (e.g., Identity Manager). | [default to false]

### Return type

[**[]Setting**](Setting.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenantSettingsByKey

> bool GetTenantSettingsByKey(ctx, settingKey).Execute()

Get a tenant setting by key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/backend-api-client"
)

func main() {
	settingKey := "settingKey_example" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantAPI.GetTenantSettingsByKey(context.Background(), settingKey).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantAPI.GetTenantSettingsByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTenantSettingsByKey`: bool
	fmt.Fprintf(os.Stdout, "Response from `TenantAPI.GetTenantSettingsByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**settingKey** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantSettingsByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**bool**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateTenantSetting

> TenantSettingCreationResponse UpdateTenantSetting(ctx).TenantId(tenantId).TenantSettingCreationRequest(tenantSettingCreationRequest).Execute()

Update a tenant setting.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/backend-api-client"
)

func main() {
	tenantId := "tenantId_example" // string |  (optional)
	tenantSettingCreationRequest := *openapiclient.NewTenantSettingCreationRequest() // TenantSettingCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantAPI.UpdateTenantSetting(context.Background()).TenantId(tenantId).TenantSettingCreationRequest(tenantSettingCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantAPI.UpdateTenantSetting``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateTenantSetting`: TenantSettingCreationResponse
	fmt.Fprintf(os.Stdout, "Response from `TenantAPI.UpdateTenantSetting`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateTenantSettingRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** |  | 
 **tenantSettingCreationRequest** | [**TenantSettingCreationRequest**](TenantSettingCreationRequest.md) |  | 

### Return type

[**TenantSettingCreationResponse**](TenantSettingCreationResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

