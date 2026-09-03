# \SettingsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetAuthorizationSettingByKey**](SettingsAPI.md#GetAuthorizationSettingByKey) | **Get** /api/v1/authorization/settings/{key} | Get authorization setting by key
[**GetAuthorizationSettings**](SettingsAPI.md#GetAuthorizationSettings) | **Get** /api/v1/authorization/settings | Get authorization settings
[**UpdateAuthorizationSettingByKey**](SettingsAPI.md#UpdateAuthorizationSettingByKey) | **Put** /api/v1/authorization/settings/{key} | Update authorization setting by key



## GetAuthorizationSettingByKey

> EnableScopedAuthorizationSetting GetAuthorizationSettingByKey(ctx, key).TenantId(tenantId).Execute()

Get authorization setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	key := openapiclient.AuthorizationSettingsKeyEnum("enableScopedAuthorization") // AuthorizationSettingsKeyEnum | The settings key
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetAuthorizationSettingByKey(context.Background(), key).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetAuthorizationSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAuthorizationSettingByKey`: EnableScopedAuthorizationSetting
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetAuthorizationSettingByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**AuthorizationSettingsKeyEnum**](.md) | The settings key | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetAuthorizationSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**EnableScopedAuthorizationSetting**](EnableScopedAuthorizationSetting.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAuthorizationSettings

> AuthorizationSettingsGetResponse GetAuthorizationSettings(ctx).TenantId(tenantId).Execute()

Get authorization settings

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetAuthorizationSettings(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetAuthorizationSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAuthorizationSettings`: AuthorizationSettingsGetResponse
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetAuthorizationSettings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetAuthorizationSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**AuthorizationSettingsGetResponse**](AuthorizationSettingsGetResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateAuthorizationSettingByKey

> UpdateAuthorizationSettingByKey(ctx, key).UpdateAuthorizationSettingByKeyRequest(updateAuthorizationSettingByKeyRequest).TenantId(tenantId).Execute()

Update authorization setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	key := openapiclient.AuthorizationSettingsKeyEnum("enableScopedAuthorization") // AuthorizationSettingsKeyEnum | The settings key
	updateAuthorizationSettingByKeyRequest := *openapiclient.NewUpdateAuthorizationSettingByKeyRequest() // UpdateAuthorizationSettingByKeyRequest | 
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SettingsAPI.UpdateAuthorizationSettingByKey(context.Background(), key).UpdateAuthorizationSettingByKeyRequest(updateAuthorizationSettingByKeyRequest).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.UpdateAuthorizationSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**AuthorizationSettingsKeyEnum**](.md) | The settings key | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateAuthorizationSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateAuthorizationSettingByKeyRequest** | [**UpdateAuthorizationSettingByKeyRequest**](UpdateAuthorizationSettingByKeyRequest.md) |  | 
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

