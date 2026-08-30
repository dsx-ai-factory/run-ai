# \SettingsAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetAppearanceSettingByKey**](SettingsAPI.md#GetAppearanceSettingByKey) | **Get** /api/v1/ui-appearance/settings/{key} | Get appearance setting by key
[**GetAppearanceSettings**](SettingsAPI.md#GetAppearanceSettings) | **Get** /api/v1/ui-appearance/settings | Get appearance settings
[**GetBrandingSettingByKey**](SettingsAPI.md#GetBrandingSettingByKey) | **Get** /api/v1/branding/settings/{key} | Get branding setting by key
[**GetBrandingSettings**](SettingsAPI.md#GetBrandingSettings) | **Get** /api/v1/branding/settings | Get branding settings
[**GetNotificationsSettingByKey**](SettingsAPI.md#GetNotificationsSettingByKey) | **Get** /api/v1/notifications/settings/{key} | Get notifications setting by key
[**GetNotificationsSettings**](SettingsAPI.md#GetNotificationsSettings) | **Get** /api/v1/notifications/settings | Get notifications settings
[**GetWorkloadsSettingByKey**](SettingsAPI.md#GetWorkloadsSettingByKey) | **Get** /api/v1/ui-workloads/settings/{key} | Get workloads setting by key
[**GetWorkloadsSettings**](SettingsAPI.md#GetWorkloadsSettings) | **Get** /api/v1/ui-workloads/settings | Get workloads settings
[**UpdateAppearanceSettingByKey**](SettingsAPI.md#UpdateAppearanceSettingByKey) | **Put** /api/v1/ui-appearance/settings/{key} | Update appearance setting by key
[**UpdateBrandingSettingByKey**](SettingsAPI.md#UpdateBrandingSettingByKey) | **Put** /api/v1/branding/settings/{key} | Update branding setting by key
[**UpdateNotificationsSettingByKey**](SettingsAPI.md#UpdateNotificationsSettingByKey) | **Put** /api/v1/notifications/settings/{key} | Update notifications setting by key
[**UpdateWorkloadsSettingByKey**](SettingsAPI.md#UpdateWorkloadsSettingByKey) | **Put** /api/v1/ui-workloads/settings/{key} | Update workloads setting by key



## GetAppearanceSettingByKey

> map[string]interface{} GetAppearanceSettingByKey(ctx, key).TenantId(tenantId).Execute()

Get appearance setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/bff-service/pkg/client"
)

func main() {
	key := openapiclient.appearanceSettingsKeyEnum("bannerConfig") // AppearanceSettingsKeyEnum | The appearance setting key
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetAppearanceSettingByKey(context.Background(), key).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetAppearanceSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAppearanceSettingByKey`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetAppearanceSettingByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**AppearanceSettingsKeyEnum**](.md) | The appearance setting key | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetAppearanceSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

**map[string]interface{}**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAppearanceSettings

> AppearanceSettingsGetAllResponse GetAppearanceSettings(ctx).TenantId(tenantId).Execute()

Get appearance settings

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/bff-service/pkg/client"
)

func main() {
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetAppearanceSettings(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetAppearanceSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAppearanceSettings`: AppearanceSettingsGetAllResponse
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetAppearanceSettings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetAppearanceSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**AppearanceSettingsGetAllResponse**](AppearanceSettingsGetAllResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetBrandingSettingByKey

> map[string]interface{} GetBrandingSettingByKey(ctx, key).TenantId(tenantId).Execute()

Get branding setting by key



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/bff-service/pkg/client"
)

func main() {
	key := openapiclient.brandingSettingsKeyEnum("logo") // BrandingSettingsKeyEnum | Identifier of the branding setting to retrieve or update
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetBrandingSettingByKey(context.Background(), key).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetBrandingSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetBrandingSettingByKey`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetBrandingSettingByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**BrandingSettingsKeyEnum**](.md) | Identifier of the branding setting to retrieve or update | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetBrandingSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

**map[string]interface{}**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetBrandingSettings

> BrandingSettingsGetAllResponse GetBrandingSettings(ctx).TenantId(tenantId).Execute()

Get branding settings



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/bff-service/pkg/client"
)

func main() {
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetBrandingSettings(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetBrandingSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetBrandingSettings`: BrandingSettingsGetAllResponse
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetBrandingSettings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetBrandingSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**BrandingSettingsGetAllResponse**](BrandingSettingsGetAllResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNotificationsSettingByKey

> map[string]interface{} GetNotificationsSettingByKey(ctx, key).TenantId(tenantId).Execute()

Get notifications setting by key



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/bff-service/pkg/client"
)

func main() {
	key := openapiclient.notificationsSettingsKeyEnum("adminMessage") // NotificationsSettingsKeyEnum | Identifier of the notifications setting to retrieve or update.
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetNotificationsSettingByKey(context.Background(), key).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetNotificationsSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNotificationsSettingByKey`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetNotificationsSettingByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**NotificationsSettingsKeyEnum**](.md) | Identifier of the notifications setting to retrieve or update. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNotificationsSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

**map[string]interface{}**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNotificationsSettings

> NotificationsSettingsGetAllResponse GetNotificationsSettings(ctx).TenantId(tenantId).Execute()

Get notifications settings



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/bff-service/pkg/client"
)

func main() {
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetNotificationsSettings(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetNotificationsSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNotificationsSettings`: NotificationsSettingsGetAllResponse
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetNotificationsSettings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNotificationsSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**NotificationsSettingsGetAllResponse**](NotificationsSettingsGetAllResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadsSettingByKey

> map[string]interface{} GetWorkloadsSettingByKey(ctx, key).TenantId(tenantId).Execute()

Get workloads setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/bff-service/pkg/client"
)

func main() {
	key := openapiclient.workloadsSettingsKeyEnum("enableNIM") // WorkloadsSettingsKeyEnum | The workloads setting key
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetWorkloadsSettingByKey(context.Background(), key).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetWorkloadsSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadsSettingByKey`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetWorkloadsSettingByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**WorkloadsSettingsKeyEnum**](.md) | The workloads setting key | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadsSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

**map[string]interface{}**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadsSettings

> WorkloadsSettingsGetAllResponse GetWorkloadsSettings(ctx).TenantId(tenantId).Execute()

Get workloads settings

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/bff-service/pkg/client"
)

func main() {
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetWorkloadsSettings(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetWorkloadsSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadsSettings`: WorkloadsSettingsGetAllResponse
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetWorkloadsSettings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadsSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**WorkloadsSettingsGetAllResponse**](WorkloadsSettingsGetAllResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateAppearanceSettingByKey

> UpdateAppearanceSettingByKey(ctx, key).UpdateBannerConfig(updateBannerConfig).TenantId(tenantId).Execute()

Update appearance setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/bff-service/pkg/client"
)

func main() {
	key := openapiclient.appearanceSettingsKeyEnum("bannerConfig") // AppearanceSettingsKeyEnum | The appearance setting key
	updateBannerConfig := *openapiclient.NewUpdateBannerConfig() // UpdateBannerConfig | 
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SettingsAPI.UpdateAppearanceSettingByKey(context.Background(), key).UpdateBannerConfig(updateBannerConfig).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.UpdateAppearanceSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**AppearanceSettingsKeyEnum**](.md) | The appearance setting key | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateAppearanceSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateBannerConfig** | [**UpdateBannerConfig**](UpdateBannerConfig.md) |  | 
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


## UpdateBrandingSettingByKey

> UpdateBrandingSettingByKey(ctx, key).UpdateLogo(updateLogo).TenantId(tenantId).Execute()

Update branding setting by key



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/bff-service/pkg/client"
)

func main() {
	key := openapiclient.brandingSettingsKeyEnum("logo") // BrandingSettingsKeyEnum | Identifier of the branding setting to retrieve or update
	updateLogo := *openapiclient.NewUpdateLogo("Logo_example") // UpdateLogo | 
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SettingsAPI.UpdateBrandingSettingByKey(context.Background(), key).UpdateLogo(updateLogo).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.UpdateBrandingSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**BrandingSettingsKeyEnum**](.md) | Identifier of the branding setting to retrieve or update | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateBrandingSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateLogo** | [**UpdateLogo**](UpdateLogo.md) |  | 
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


## UpdateNotificationsSettingByKey

> UpdateNotificationsSettingByKey(ctx, key).UpdateAdminMessage(updateAdminMessage).TenantId(tenantId).Execute()

Update notifications setting by key



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/bff-service/pkg/client"
)

func main() {
	key := openapiclient.notificationsSettingsKeyEnum("adminMessage") // NotificationsSettingsKeyEnum | Identifier of the notifications setting to retrieve or update.
	updateAdminMessage := *openapiclient.NewUpdateAdminMessage() // UpdateAdminMessage | 
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SettingsAPI.UpdateNotificationsSettingByKey(context.Background(), key).UpdateAdminMessage(updateAdminMessage).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.UpdateNotificationsSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**NotificationsSettingsKeyEnum**](.md) | Identifier of the notifications setting to retrieve or update. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNotificationsSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateAdminMessage** | [**UpdateAdminMessage**](UpdateAdminMessage.md) |  | 
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


## UpdateWorkloadsSettingByKey

> UpdateWorkloadsSettingByKey(ctx, key).UpdateWorkloadsSettingByKeyRequest(updateWorkloadsSettingByKeyRequest).TenantId(tenantId).Execute()

Update workloads setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/bff-service/pkg/client"
)

func main() {
	key := openapiclient.workloadsSettingsKeyEnum("enableNIM") // WorkloadsSettingsKeyEnum | The workloads setting key
	updateWorkloadsSettingByKeyRequest := *openapiclient.NewUpdateWorkloadsSettingByKeyRequest() // UpdateWorkloadsSettingByKeyRequest | 
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SettingsAPI.UpdateWorkloadsSettingByKey(context.Background(), key).UpdateWorkloadsSettingByKeyRequest(updateWorkloadsSettingByKeyRequest).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.UpdateWorkloadsSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**WorkloadsSettingsKeyEnum**](.md) | The workloads setting key | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateWorkloadsSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateWorkloadsSettingByKeyRequest** | [**UpdateWorkloadsSettingByKeyRequest**](UpdateWorkloadsSettingByKeyRequest.md) |  | 
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

