# \SettingsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetCatalogSettingByKey**](SettingsAPI.md#GetCatalogSettingByKey) | **Get** /api/v1/catalog/settings/{key} | Get catalog setting by key
[**GetCatalogSettings**](SettingsAPI.md#GetCatalogSettings) | **Get** /api/v1/catalog/settings | Get catalog settings
[**UpdateCatalogSettingByKey**](SettingsAPI.md#UpdateCatalogSettingByKey) | **Put** /api/v1/catalog/settings/{key} | Update catalog setting by key



## GetCatalogSettingByKey

> GetEnableNgcIntegration GetCatalogSettingByKey(ctx, key).TenantId(tenantId).Execute()

Get catalog setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/catalog-service-client"
)

func main() {
	key := openapiclient.CatalogSettingsKeyEnum("enableNgcIntegration") // CatalogSettingsKeyEnum | The settings key
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetCatalogSettingByKey(context.Background(), key).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetCatalogSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetCatalogSettingByKey`: GetEnableNgcIntegration
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetCatalogSettingByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**CatalogSettingsKeyEnum**](.md) | The settings key | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetCatalogSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**GetEnableNgcIntegration**](GetEnableNgcIntegration.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetCatalogSettings

> CatalogSettingsGetResponse GetCatalogSettings(ctx).TenantId(tenantId).Execute()

Get catalog settings

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/catalog-service-client"
)

func main() {
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetCatalogSettings(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetCatalogSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetCatalogSettings`: CatalogSettingsGetResponse
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetCatalogSettings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetCatalogSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**CatalogSettingsGetResponse**](CatalogSettingsGetResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateCatalogSettingByKey

> UpdateCatalogSettingByKey(ctx, key).UpdateCatalogSettingByKeyRequest(updateCatalogSettingByKeyRequest).TenantId(tenantId).Execute()

Update catalog setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/catalog-service-client"
)

func main() {
	key := openapiclient.CatalogSettingsKeyEnum("enableNgcIntegration") // CatalogSettingsKeyEnum | The settings key
	updateCatalogSettingByKeyRequest := *openapiclient.NewUpdateCatalogSettingByKeyRequest() // UpdateCatalogSettingByKeyRequest | 
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SettingsAPI.UpdateCatalogSettingByKey(context.Background(), key).UpdateCatalogSettingByKeyRequest(updateCatalogSettingByKeyRequest).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.UpdateCatalogSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**CatalogSettingsKeyEnum**](.md) | The settings key | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateCatalogSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateCatalogSettingByKeyRequest** | [**UpdateCatalogSettingByKeyRequest**](UpdateCatalogSettingByKeyRequest.md) |  | 
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

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

