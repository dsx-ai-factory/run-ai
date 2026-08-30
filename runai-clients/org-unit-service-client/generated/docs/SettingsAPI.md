# \SettingsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetOrgUnitSettingByKey**](SettingsAPI.md#GetOrgUnitSettingByKey) | **Get** /api/v1/org-unit/settings/{key} | Get org unit setting by key
[**GetOrgUnitSettings**](SettingsAPI.md#GetOrgUnitSettings) | **Get** /api/v1/org-unit/settings | Get org unit settings
[**UpdateOrgUnitSettingByKey**](SettingsAPI.md#UpdateOrgUnitSettingByKey) | **Put** /api/v1/org-unit/settings/{key} | Update org unit setting by key



## GetOrgUnitSettingByKey

> OrgUnitSetting GetOrgUnitSettingByKey(ctx, key).TenantId(tenantId).Execute()

Get org unit setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/org-unit-service-client"
)

func main() {
	key := openapiclient.OrgUnitSettingsKeyEnum("limitQuotaOverSubscription") // OrgUnitSettingsKeyEnum | The settings key
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetOrgUnitSettingByKey(context.Background(), key).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetOrgUnitSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetOrgUnitSettingByKey`: OrgUnitSetting
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetOrgUnitSettingByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**OrgUnitSettingsKeyEnum**](.md) | The settings key | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetOrgUnitSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**OrgUnitSetting**](OrgUnitSetting.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetOrgUnitSettings

> OrgUnitSettingsGetResponse GetOrgUnitSettings(ctx).TenantId(tenantId).Execute()

Get org unit settings

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/org-unit-service-client"
)

func main() {
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetOrgUnitSettings(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetOrgUnitSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetOrgUnitSettings`: OrgUnitSettingsGetResponse
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetOrgUnitSettings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetOrgUnitSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**OrgUnitSettingsGetResponse**](OrgUnitSettingsGetResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateOrgUnitSettingByKey

> UpdateOrgUnitSettingByKey(ctx, key).UpdateOrgUnitSettingByKeyRequest(updateOrgUnitSettingByKeyRequest).TenantId(tenantId).Execute()

Update org unit setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/org-unit-service-client"
)

func main() {
	key := openapiclient.OrgUnitSettingsKeyEnum("limitQuotaOverSubscription") // OrgUnitSettingsKeyEnum | The settings key
	updateOrgUnitSettingByKeyRequest := *openapiclient.NewUpdateOrgUnitSettingByKeyRequest() // UpdateOrgUnitSettingByKeyRequest | 
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SettingsAPI.UpdateOrgUnitSettingByKey(context.Background(), key).UpdateOrgUnitSettingByKeyRequest(updateOrgUnitSettingByKeyRequest).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.UpdateOrgUnitSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**OrgUnitSettingsKeyEnum**](.md) | The settings key | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateOrgUnitSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateOrgUnitSettingByKeyRequest** | [**UpdateOrgUnitSettingByKeyRequest**](UpdateOrgUnitSettingByKeyRequest.md) |  | 
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

