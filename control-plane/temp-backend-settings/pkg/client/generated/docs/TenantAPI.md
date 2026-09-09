# \TenantAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetTenantSettings**](TenantAPI.md#GetTenantSettings) | **Get** /v1/k8s/setting | Get all tenant settings.
[**GetTenantSettingsByKey**](TenantAPI.md#GetTenantSettingsByKey) | **Get** /v1/k8s/setting/{settingKey} | Get a tenant setting by key.
[**UpdateTenantSetting**](TenantAPI.md#UpdateTenantSetting) | **Put** /v1/k8s/setting | Update a tenant setting.



## GetTenantSettings

> []Setting GetTenantSettings(ctx).ExcludeMigrated(excludeMigrated).TenantId(tenantId).Execute()

Get all tenant settings.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/client"
)

func main() {
	excludeMigrated := true // bool | When true, excludes settings that have been migrated to external services (e.g., Identity Manager). (optional) (default to false)
	tenantId := "tenantId_example" // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantAPI.GetTenantSettings(context.Background()).ExcludeMigrated(excludeMigrated).TenantId(tenantId).Execute()
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
 **tenantId** | **string** |  | 

### Return type

[**[]Setting**](Setting.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenantSettingsByKey

> bool GetTenantSettingsByKey(ctx, settingKey).TenantId(tenantId).Execute()

Get a tenant setting by key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/client"
)

func main() {
	settingKey := "settingKey_example" // string | The setting key to retrieve
	tenantId := "tenantId_example" // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantAPI.GetTenantSettingsByKey(context.Background(), settingKey).TenantId(tenantId).Execute()
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
**settingKey** | **string** | The setting key to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantSettingsByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **string** |  | 

### Return type

**bool**

### Authorization

No authorization required

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
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/client"
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

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

