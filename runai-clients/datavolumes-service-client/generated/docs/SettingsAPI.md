# \SettingsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetDatavolumesSettingByKey**](SettingsAPI.md#GetDatavolumesSettingByKey) | **Get** /api/v1/datavolumes/settings/{key} | Get datavolumes setting by key
[**GetDatavolumesSettings**](SettingsAPI.md#GetDatavolumesSettings) | **Get** /api/v1/datavolumes/settings | Get datavolumes settings
[**UpdateDatavolumesSettingByKey**](SettingsAPI.md#UpdateDatavolumesSettingByKey) | **Put** /api/v1/datavolumes/settings/{key} | Update datavolumes setting by key



## GetDatavolumesSettingByKey

> GetEnableDataVolumes GetDatavolumesSettingByKey(ctx, key).TenantId(tenantId).Execute()

Get datavolumes setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/datavolumes-service-client"
)

func main() {
	key := openapiclient.DatavolumesSettingsKeyEnum("enableDataVolumes") // DatavolumesSettingsKeyEnum | The settings key
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetDatavolumesSettingByKey(context.Background(), key).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetDatavolumesSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDatavolumesSettingByKey`: GetEnableDataVolumes
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetDatavolumesSettingByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**DatavolumesSettingsKeyEnum**](.md) | The settings key | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDatavolumesSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**GetEnableDataVolumes**](GetEnableDataVolumes.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDatavolumesSettings

> DatavolumesSettingsGetResponse GetDatavolumesSettings(ctx).TenantId(tenantId).Execute()

Get datavolumes settings

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/datavolumes-service-client"
)

func main() {
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetDatavolumesSettings(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetDatavolumesSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDatavolumesSettings`: DatavolumesSettingsGetResponse
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetDatavolumesSettings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetDatavolumesSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**DatavolumesSettingsGetResponse**](DatavolumesSettingsGetResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDatavolumesSettingByKey

> UpdateDatavolumesSettingByKey(ctx, key).UpdateDatavolumesSettingByKeyRequest(updateDatavolumesSettingByKeyRequest).TenantId(tenantId).Execute()

Update datavolumes setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/datavolumes-service-client"
)

func main() {
	key := openapiclient.DatavolumesSettingsKeyEnum("enableDataVolumes") // DatavolumesSettingsKeyEnum | The settings key
	updateDatavolumesSettingByKeyRequest := *openapiclient.NewUpdateDatavolumesSettingByKeyRequest() // UpdateDatavolumesSettingByKeyRequest | 
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SettingsAPI.UpdateDatavolumesSettingByKey(context.Background(), key).UpdateDatavolumesSettingByKeyRequest(updateDatavolumesSettingByKeyRequest).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.UpdateDatavolumesSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**DatavolumesSettingsKeyEnum**](.md) | The settings key | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDatavolumesSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateDatavolumesSettingByKeyRequest** | [**UpdateDatavolumesSettingByKeyRequest**](UpdateDatavolumesSettingByKeyRequest.md) |  | 
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

